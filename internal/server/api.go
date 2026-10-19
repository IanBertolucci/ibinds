package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ibinds/internal/calib"
	"ibinds/internal/cfgparse"
	"ibinds/internal/config"
	"ibinds/internal/editor"
	"ibinds/internal/fileops"
	"ibinds/internal/inputcap"
	"ibinds/internal/labels"
	"ibinds/internal/ops"
	"ibinds/internal/profiles"
	"ibinds/internal/simcheck"
	"ibinds/web"
)

func (s *Server) routes() {
	s.mux.HandleFunc("/", s.handleIndex)
	s.mux.HandleFunc("/api/health", s.api(s.handleHealth))
	s.mux.HandleFunc("/api/profiles", s.api(s.handleProfiles))
	s.mux.HandleFunc("/api/profiles/", s.api(s.handleProfileControls))
	s.mux.HandleFunc("/api/settings", s.api(s.handleSettings))
	s.mux.HandleFunc("/api/plan/binding", s.api(s.handlePlanBinding))
	s.mux.HandleFunc("/api/apply/binding", s.api(s.handleApplyBinding))
	s.mux.HandleFunc("/api/plan/calib", s.api(s.handlePlanCalib))
	s.mux.HandleFunc("/api/apply/calib", s.api(s.handleApplyCalib))
	s.mux.HandleFunc("/api/input/capture", s.api(s.handleInputCapture))
}

// api wraps a handler with the token check.
func (s *Server) api(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("token") != s.token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	dist, err := fs.Sub(web.Assets, "dist")
	if err != nil {
		http.Error(w, "assets unavailable", http.StatusInternalServerError)
		return
	}
	p := strings.TrimPrefix(r.URL.Path, "/")
	if p == "" {
		p = "index.html"
	}
	if _, err := fs.Stat(dist, p); err != nil {
		p = "index.html"
	}
	data, err := fs.ReadFile(dist, p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, p, time.Time{}, bytes.NewReader(data))
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type profileDTO struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	IsBaseline  bool   `json:"isBaseline"`
	HasControls bool   `json:"hasControls"`
	HasCalib    bool   `json:"hasCalib"`
}

func (s *Server) handleProfiles(w http.ResponseWriter, r *http.Request) {
	cfg := s.cfg.Load()
	ps, err := profiles.Scan(cfg.ProfilesDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	out := make([]profileDTO, 0, len(ps))
	for _, p := range ps {
		out = append(out, profileDTO{
			ID:          p.ID,
			DisplayName: labels.Car(p.ID),
			IsBaseline:  p.IsBaseline,
			HasControls: p.HasControls,
			HasCalib:    p.HasCalib,
		})
	}
	writeJSON(w, http.StatusOK, out)
}

type controlDTO struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Kind        string `json:"kind"`
	Key         string `json:"key"`
	Button      int    `json:"button"`
	DeviceName  string `json:"deviceName"`
	AxisName    string `json:"axisName"`
}

func (s *Server) handleProfileControls(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/api/profiles/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 || parts[1] != "controls" {
		http.NotFound(w, r)
		return
	}
	id := parts[0]

	cfg := s.cfg.Load()
	devices, err := deviceNames(filepath.Join(cfg.ProfilesDir, id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := os.ReadFile(filepath.Join(cfg.ProfilesDir, id, "controls.cfg"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	f, err := cfgparse.Parse(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	out := make([]controlDTO, 0, len(f.Records))
	for _, rec := range f.Records {
		dto := controlDTO{
			Name:        rec.Name,
			DisplayName: labels.Control(rec.Name),
			Kind:        kindString(rec.Kind),
			Key:         keyString(rec),
			Button:      rec.Button,
		}
		if rec.Device != nil {
			dto.DeviceName = devices[rec.Device.InstanceGUID]
		}
		out = append(out, dto)
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *Server) handleSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ProfilesDir string `json:"profilesDir"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newCfg := &config.Config{ProfilesDir: req.ProfilesDir}
	if err := config.Save(newCfg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.cfg.Store(newCfg)
	writeJSON(w, http.StatusOK, newCfg)
}

type bindingJSON struct {
	Key                string `json:"key"`
	Button             int    `json:"button"`
	DeviceInstanceGUID string `json:"deviceInstanceGUID"`
	ProductGUID        string `json:"productGUID"`
}

func (b bindingJSON) toBinding() (editor.Binding, error) {
	if b.Key != "" {
		if len(b.Key) != 1 {
			return editor.Binding{}, errors.New("key must be a single character")
		}
		return editor.Binding{Key: b.Key[0], HasKey: true}, nil
	}
	if b.DeviceInstanceGUID != "" {
		return editor.Binding{
			Button:             b.Button,
			HasButton:          true,
			DeviceInstanceGUID: b.DeviceInstanceGUID,
			ProductGUID:        b.ProductGUID,
		}, nil
	}
	return editor.Binding{}, errors.New("binding requires a key or a device button")
}

type profileOverride struct {
	ProfileID string `json:"profileId"`
	Force     bool   `json:"force"`
}

type bindingReq struct {
	Control          string            `json:"control"`
	Binding          bindingJSON       `json:"binding"`
	ProfileIDs       []string          `json:"profileIds"`
	Force            bool              `json:"force"`
	ProfileOverrides []profileOverride `json:"profileOverrides"`
}

func (r bindingReq) forceFor(profileID string) bool {
	if r.Force {
		return true
	}
	for _, o := range r.ProfileOverrides {
		if o.ProfileID == profileID && o.Force {
			return true
		}
	}
	return false
}

type planItem struct {
	ProfileID   string            `json:"profileId"`
	Conflicts   []editor.Conflict `json:"conflicts"`
	Warning     string            `json:"warning"`
	DiffSummary string            `json:"diffSummary"`
	Force       bool              `json:"force"`
}

func (s *Server) handlePlanBinding(w http.ResponseWriter, r *http.Request) {
	var req bindingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := req.Binding.toBinding()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := s.cfg.Load()

	plans := make([]planItem, 0, len(req.ProfileIDs))
	for _, id := range req.ProfileIDs {
		force := req.forceFor(id)
		res, err := ops.PlanBinding(cfg.ProfilesDir, id, req.Control, b, force)
		if err != nil {
			http.Error(w, fmt.Sprintf("%s: %v", id, err), http.StatusInternalServerError)
			return
		}
		plans = append(plans, planItem{
			ProfileID:   id,
			Conflicts:   res.Conflicts,
			Warning:     res.Warning,
			DiffSummary: res.Diff,
			Force:       force,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"plans": plans})
}

type appliedItem struct {
	ProfileID   string `json:"profileId"`
	BackupPath  string `json:"backupPath"`
	DiffSummary string `json:"diffSummary"`
}

type skippedItem struct {
	ProfileID string `json:"profileId"`
	Reason    string `json:"reason"`
}

func (s *Server) handleApplyBinding(w http.ResponseWriter, r *http.Request) {
	var req bindingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := req.Binding.toBinding()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.simCheck != nil {
		running, err := s.simCheck()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if running {
			http.Error(w, simcheck.ErrSimRunning.Error(), http.StatusConflict)
			return
		}
	}

	cfg := s.cfg.Load()
	applied := []appliedItem{}
	skipped := []skippedItem{}

	for _, id := range req.ProfileIDs {
		force := req.forceFor(id)
		res, err := ops.PlanBinding(cfg.ProfilesDir, id, req.Control, b, force)
		if err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		if len(res.Conflicts) > 0 && !force {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: "binding conflict"})
			continue
		}

		cfgPath := filepath.Join(cfg.ProfilesDir, id, "controls.cfg")
		backupPath, err := fileops.Backup(cfgPath)
		if err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		data, err := os.ReadFile(cfgPath)
		if err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		f, err := cfgparse.Parse(data)
		if err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		newData, _, err := editor.ApplyBinding(f, req.Control, b)
		if err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		if err := os.WriteFile(cfgPath, newData, 0o644); err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		diff := fileops.DiffSummary(fileops.Diff(data, newData))
		applied = append(applied, appliedItem{ProfileID: id, BackupPath: backupPath, DiffSummary: diff})
	}

	writeJSON(w, http.StatusOK, map[string]any{"applied": applied, "skipped": skipped})
}

type calibReq struct {
	SourceProfileID string   `json:"sourceProfileId"`
	ProfileIDs      []string `json:"profileIds"`
}

func (s *Server) handlePlanCalib(w http.ResponseWriter, r *http.Request) {
	var req calibReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cfg := s.cfg.Load()
	src, err := os.ReadFile(filepath.Join(cfg.ProfilesDir, req.SourceProfileID, "joyCalib.yaml"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	plans := []planItem{}
	for _, id := range req.ProfileIDs {
		dst, err := os.ReadFile(filepath.Join(cfg.ProfilesDir, id, "joyCalib.yaml"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		plans = append(plans, planItem{
			ProfileID:   id,
			DiffSummary: fileops.DiffSummary(fileops.Diff(dst, src)),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"plans": plans})
}

func (s *Server) handleApplyCalib(w http.ResponseWriter, r *http.Request) {
	var req calibReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.simCheck != nil {
		running, err := s.simCheck()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if running {
			http.Error(w, simcheck.ErrSimRunning.Error(), http.StatusConflict)
			return
		}
	}
	cfg := s.cfg.Load()
	src, err := os.ReadFile(filepath.Join(cfg.ProfilesDir, req.SourceProfileID, "joyCalib.yaml"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := calib.Validate(src); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	applied := []appliedItem{}
	skipped := []skippedItem{}
	for _, id := range req.ProfileIDs {
		dstPath := filepath.Join(cfg.ProfilesDir, id, "joyCalib.yaml")
		old, err := os.ReadFile(dstPath)
		if err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		if err := calib.Validate(old); err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		backupPath, err := fileops.Backup(dstPath)
		if err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		if err := os.WriteFile(dstPath, src, 0o644); err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		if err := calib.Validate(src); err != nil {
			skipped = append(skipped, skippedItem{ProfileID: id, Reason: err.Error()})
			continue
		}
		diff := fileops.DiffSummary(fileops.Diff(old, src))
		applied = append(applied, appliedItem{ProfileID: id, BackupPath: backupPath, DiffSummary: diff})
	}
	writeJSON(w, http.StatusOK, map[string]any{"applied": applied, "skipped": skipped})
}

func (s *Server) handleInputCapture(w http.ResponseWriter, r *http.Request) {
	if s.inputCap == nil {
		http.Error(w, "input capture not available", http.StatusNotImplemented)
		return
	}
	var req struct {
		TimeoutMs int `json:"timeoutMs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := s.inputCap.Capture(inputcap.CaptureOpts{Timeout: time.Duration(req.TimeoutMs) * time.Millisecond})
	if err != nil {
		if errors.Is(err, inputcap.ErrUnsupported) {
			http.Error(w, err.Error(), http.StatusNotImplemented)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if result == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func kindString(k cfgparse.BindingKind) string {
	switch k {
	case cfgparse.KindUnbound:
		return "unbound"
	case cfgparse.KindKey:
		return "key"
	case cfgparse.KindButton:
		return "button"
	case cfgparse.KindAxis:
		return "axis"
	default:
		return "unknown"
	}
}

func keyString(rec cfgparse.Record) string {
	if rec.Kind != cfgparse.KindKey {
		return ""
	}
	return string([]byte{rec.Key})
}

// deviceNames maps InstanceGUID -> DeviceName from a profile's joyCalib.yaml.
func deviceNames(profDir string) (map[string]string, error) {
	out := map[string]string{}
	data, err := os.ReadFile(filepath.Join(profDir, "joyCalib.yaml"))
	if err != nil {
		if os.IsNotExist(err) {
			return out, nil
		}
		return nil, err
	}
	c, err := calib.Parse(data)
	if err != nil {
		return nil, err
	}
	for _, d := range c.Devices {
		out[d.InstanceGUID] = d.Name
	}
	return out, nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
