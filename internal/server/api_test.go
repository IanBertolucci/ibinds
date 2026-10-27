package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"ibinds/internal/cfgparse"
	"ibinds/internal/config"
	"ibinds/internal/editor"
	"ibinds/internal/inputcap"
)

const testToken = "testtoken"

func copyFixture(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// setupProfilesDir builds a temp profiles dir; each entry maps a profile id to
// a controls.cfg fixture name under ../cfgparse/testdata. Every profile gets a
// joyCalib.yaml copy too.
func setupProfilesDir(t *testing.T, profiles map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for id, cfgName := range profiles {
		pd := filepath.Join(dir, id)
		if err := os.MkdirAll(pd, 0o755); err != nil {
			t.Fatal(err)
		}
		copyFixture(t, filepath.Join("..", "cfgparse", "testdata", cfgName), filepath.Join(pd, "controls.cfg"))
		copyFixture(t, filepath.Join("..", "calib", "testdata", "f4_joycalib.yaml"), filepath.Join(pd, "joyCalib.yaml"))
	}
	return dir
}

func newServer(t *testing.T, profilesDir string, opts ServerOpts) *httptest.Server {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(t.TempDir(), ".config"))
	t.Setenv("IBINDS_PROFILES_DIR", "")
	config.FlagDir = ""
	if opts.Token == "" {
		opts.Token = testToken
	}
	s := New(&config.Config{ProfilesDir: profilesDir}, opts)
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func get(t *testing.T, ts *httptest.Server, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func postJSON(t *testing.T, ts *httptest.Server, path string, body any) *http.Response {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.Post(ts.URL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func readBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestHealthRequiresToken(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	if resp := get(t, ts, "/api/health"); resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token: status = %d, want 401", resp.StatusCode)
	}
	if resp := get(t, ts, "/api/health?token="+testToken); resp.StatusCode != http.StatusOK {
		t.Fatalf("with token: status = %d, want 200", resp.StatusCode)
	}
}

func TestServesIndex(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	resp := get(t, ts, "/?token="+testToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	if body := string(readBody(t, resp)); !contains(body, "iBinds") {
		t.Fatalf("index body %q missing placeholder marker", body)
	}
}

func TestProfilesList(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg", "Baseline": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	resp := get(t, ts, "/api/profiles?token="+testToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var profiles []map[string]any
	if err := json.Unmarshal(readBody(t, resp), &profiles); err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 2 {
		t.Fatalf("len = %d, want 2", len(profiles))
	}
	if profiles[0]["id"] != "Baseline" || profiles[0]["isBaseline"] != true {
		t.Fatalf("first profile = %+v, want Baseline first", profiles[0])
	}
}

func TestControlsListResolvesDeviceName(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_btn128.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	resp := get(t, ts, "/api/profiles/f4/controls?token="+testToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var controls []map[string]any
	if err := json.Unmarshal(readBody(t, resp), &controls); err != nil {
		t.Fatal(err)
	}
	tca := findControl(controls, "TractionControlArm")
	if tca == nil {
		t.Fatal("TractionControlArm not found")
	}
	if tca["kind"] != "button" || int(tca["button"].(float64)) != 128 {
		t.Fatalf("TractionControlArm = %+v, want kind=button button=128", tca)
	}
	if tca["deviceName"] != "MOZA R9 Base" {
		t.Fatalf("deviceName = %v, want MOZA R9 Base", tca["deviceName"])
	}
}

func TestSettingsPersistsProfilesDir(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	newDir := filepath.Join(dir, "newdir")
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if resp := postJSON(t, ts, "/api/settings?token="+testToken, map[string]any{"profilesDir": newDir}); resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	if cfg.ProfilesDir != newDir {
		t.Fatalf("ProfilesDir = %q, want %q", cfg.ProfilesDir, newDir)
	}
}

func TestPlanBindingReportsConflict(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"control":    "Ignition",
		"binding":    map[string]any{"key": "U"},
		"profileIds": []string{"f4"},
		"force":      false,
	}
	resp := postJSON(t, ts, "/api/plan/binding?token="+testToken, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var plan struct {
		Plans []struct {
			Conflicts []map[string]any `json:"conflicts"`
		} `json:"plans"`
	}
	if err := json.Unmarshal(readBody(t, resp), &plan); err != nil {
		t.Fatal(err)
	}
	if len(plan.Plans) != 1 || len(plan.Plans[0].Conflicts) == 0 {
		t.Fatalf("plan = %+v, want non-empty conflicts", plan)
	}
}

func TestApplyPerProfileForce(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{
		"f4": "f4_key_u.cfg",
		"f5": "f4_key_u.cfg",
	})
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"control":    "Ignition",
		"binding":    map[string]any{"key": "U"},
		"profileIds": []string{"f4", "f5"},
		"force":      false,
		"profileOverrides": []map[string]any{
			{"profileId": "f4", "force": true},
		},
	}
	resp := postJSON(t, ts, "/api/apply/binding?token="+testToken, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", resp.StatusCode, readBody(t, resp))
	}
	var out struct {
		Applied []map[string]any `json:"applied"`
		Skipped []map[string]any `json:"skipped"`
	}
	if err := json.Unmarshal(readBody(t, resp), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Applied) != 1 || out.Applied[0]["profileId"] != "f4" {
		t.Fatalf("applied = %+v, want exactly f4", out.Applied)
	}
	if len(out.Skipped) != 1 || out.Skipped[0]["profileId"] != "f5" {
		t.Fatalf("skipped = %+v, want exactly f5", out.Skipped)
	}
}

func TestApplyBlockedWhenSimRunning(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	before, err := os.ReadFile(filepath.Join(dir, "f4", "controls.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	ts := newServer(t, dir, ServerOpts{SimCheck: func() (bool, error) { return true, nil }})

	req := map[string]any{
		"control":    "Ignition",
		"binding":    map[string]any{"key": "I"},
		"profileIds": []string{"f4"},
		"force":      false,
	}
	resp := postJSON(t, ts, "/api/apply/binding?token="+testToken, req)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
	after, err := os.ReadFile(filepath.Join(dir, "f4", "controls.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("controls.cfg changed despite sim running")
	}
}

func TestApplyCreatesBackup(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"control":    "Ignition",
		"binding":    map[string]any{"key": "q"},
		"profileIds": []string{"f4"},
		"force":      false,
	}
	resp := postJSON(t, ts, "/api/apply/binding?token="+testToken, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", resp.StatusCode, readBody(t, resp))
	}
	entries, err := os.ReadDir(filepath.Join(dir, "f4"))
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		if matched, _ := filepath.Match("controls.cfg.bak-*", e.Name()); matched {
			found = true
		}
	}
	if !found {
		t.Fatal("no timestamped backup found in profile dir")
	}
}

func TestPlanWarnsDeviceMissing(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"control": "Ignition",
		"binding": map[string]any{
			"button":             130,
			"deviceInstanceGUID": "{11111111-2222-3333-4444-555555555555}",
			"productGUID":        "{0002346E-0000-0000-0000-504944564944}",
		},
		"profileIds": []string{"f4"},
		"force":      false,
	}
	resp := postJSON(t, ts, "/api/plan/binding?token="+testToken, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var plan struct {
		Plans []struct {
			Warning string `json:"warning"`
		} `json:"plans"`
	}
	if err := json.Unmarshal(readBody(t, resp), &plan); err != nil {
		t.Fatal(err)
	}
	if len(plan.Plans) != 1 || plan.Plans[0].Warning == "" {
		t.Fatalf("plan = %+v, want non-empty warning", plan)
	}
}

func TestSettingsTakesEffectWithoutRestart(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	// Build a second profiles dir with a different car.
	dir2 := t.TempDir()
	pd2 := filepath.Join(dir2, "mustanggt4")
	os.MkdirAll(pd2, 0o755)
	copyFixture(t, filepath.Join("..", "cfgparse", "testdata", "f4_key_u.cfg"), filepath.Join(pd2, "controls.cfg"))

	if resp := postJSON(t, ts, "/api/settings?token="+testToken, map[string]any{"profilesDir": dir2}); resp.StatusCode != http.StatusOK {
		t.Fatalf("settings status = %d", resp.StatusCode)
	}

	resp := get(t, ts, "/api/profiles?token="+testToken)
	var profiles []map[string]any
	if err := json.Unmarshal(readBody(t, resp), &profiles); err != nil {
		t.Fatal(err)
	}
	if len(profiles) != 1 || profiles[0]["id"] != "mustanggt4" {
		t.Fatalf("profiles = %+v, want just mustanggt4 (from new dir)", profiles)
	}
}

func TestCalibCopyFlow(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	dst := filepath.Join(dir, "dst")
	os.MkdirAll(src, 0o755)
	os.MkdirAll(dst, 0o755)
	copyFixture(t, filepath.Join("..", "calib", "testdata", "f4_joycalib.yaml"), filepath.Join(src, "joyCalib.yaml"))
	// dst starts with a different calib file
	copyFixture(t, filepath.Join("..", "calib", "testdata", "baseline_joycalib.yaml"), filepath.Join(dst, "joyCalib.yaml"))

	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"sourceProfileId": "src",
		"profileIds":      []string{"dst"},
	}
	resp := postJSON(t, ts, "/api/apply/calib?token="+testToken, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", resp.StatusCode, readBody(t, resp))
	}

	srcData, _ := os.ReadFile(filepath.Join(src, "joyCalib.yaml"))
	dstData, err := os.ReadFile(filepath.Join(dst, "joyCalib.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(srcData, dstData) {
		t.Fatal("destination joyCalib.yaml not byte-identical to source")
	}
	// backup created in dst
	entries, _ := os.ReadDir(dst)
	found := false
	for _, e := range entries {
		if matched, _ := filepath.Match("joyCalib.yaml.bak-*", e.Name()); matched {
			found = true
		}
	}
	if !found {
		t.Fatal("no joyCalib.yaml backup found in dst")
	}
}

func findControl(controls []map[string]any, name string) map[string]any {
	for _, c := range controls {
		if c["name"] == name {
			return c
		}
	}
	return nil
}

func writeProfile(t *testing.T, dir, id, cfgName string, edits map[string]editor.Binding) {
	t.Helper()
	pd := filepath.Join(dir, id)
	if err := os.MkdirAll(pd, 0o755); err != nil {
		t.Fatal(err)
	}
	copyFixture(t, filepath.Join("..", "cfgparse", "testdata", cfgName), filepath.Join(pd, "controls.cfg"))
	copyFixture(t, filepath.Join("..", "calib", "testdata", "f4_joycalib.yaml"), filepath.Join(pd, "joyCalib.yaml"))
	if len(edits) == 0 {
		return
	}
	data, err := os.ReadFile(filepath.Join(pd, "controls.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	cur := data
	for control, b := range edits {
		f, err := cfgparse.Parse(cur)
		if err != nil {
			t.Fatal(err)
		}
		cur, _, err = editor.ApplyBinding(f, control, b)
		if err != nil {
			t.Fatalf("ApplyBinding %s: %v", control, err)
		}
	}
	if err := os.WriteFile(filepath.Join(pd, "controls.cfg"), cur, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestControlsListExposesButtons(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_btn130.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	resp := get(t, ts, "/api/profiles/f4/controls?token="+testToken)
	var controls []map[string]any
	if err := json.Unmarshal(readBody(t, resp), &controls); err != nil {
		t.Fatal(err)
	}
	tca := findControl(controls, "TractionControlArm")
	if tca == nil {
		t.Fatal("TractionControlArm not found")
	}
	buttons, ok := tca["buttons"].([]any)
	if !ok || len(buttons) != 1 || int(buttons[0].(float64)) != 130 {
		t.Fatalf("buttons = %v, want [130]", tca["buttons"])
	}
}

func TestControlsListButtonsEmptyForKey(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	resp := get(t, ts, "/api/profiles/f4/controls?token="+testToken)
	var controls []map[string]any
	if err := json.Unmarshal(readBody(t, resp), &controls); err != nil {
		t.Fatal(err)
	}
	tca := findControl(controls, "TractionControlArm")
	if tca == nil {
		t.Fatal("TractionControlArm not found")
	}
	buttons, ok := tca["buttons"].([]any)
	if !ok || len(buttons) != 0 {
		t.Fatalf("buttons = %v, want []", tca["buttons"])
	}
}

func TestPlanCopyReturnsPlans(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "src", "f4_key_u.cfg", nil)
	writeProfile(t, dir, "dst", "f4_btn128.cfg", nil)
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"sourceProfileId": "src",
		"controls":        []string{"TractionControlArm"},
		"profileIds":      []string{"dst"},
	}
	resp := postJSON(t, ts, "/api/plan/copy?token="+testToken, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", resp.StatusCode, readBody(t, resp))
	}
	var plan struct {
		Plans []map[string]any `json:"plans"`
	}
	if err := json.Unmarshal(readBody(t, resp), &plan); err != nil {
		t.Fatal(err)
	}
	if len(plan.Plans) != 1 {
		t.Fatalf("plans = %+v, want 1", plan.Plans)
	}
	if conflicts, ok := plan.Plans[0]["conflicts"].([]any); !ok || len(conflicts) != 0 {
		t.Fatalf("conflicts = %v, want empty", plan.Plans[0]["conflicts"])
	}
}

func TestApplyCopyWritesTarget(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "src", "f4_key_u.cfg", nil)
	writeProfile(t, dir, "dst", "f4_btn128.cfg", nil)
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"sourceProfileId": "src",
		"controls":        []string{"TractionControlArm"},
		"profileIds":      []string{"dst"},
	}
	resp := postJSON(t, ts, "/api/apply/copy?token="+testToken, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", resp.StatusCode, readBody(t, resp))
	}
	var out struct {
		Applied []map[string]any `json:"applied"`
	}
	if err := json.Unmarshal(readBody(t, resp), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Applied) != 1 || out.Applied[0]["profileId"] != "dst" {
		t.Fatalf("applied = %+v, want exactly dst", out.Applied)
	}
	want, err := os.ReadFile(filepath.Join("..", "cfgparse", "testdata", "f4_key_u.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "dst", "controls.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(want, got) {
		t.Fatal("dst/controls.cfg not byte-identical to f4_key_u.cfg")
	}
}

func TestApplyCopySkipsConflict(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "src", "f4_key_u.cfg", map[string]editor.Binding{"Ignition": {Key: 'U', HasKey: true}})
	writeProfile(t, dir, "dst", "f4_key_u.cfg", nil)
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"sourceProfileId": "src",
		"controls":        []string{"Ignition"},
		"profileIds":      []string{"dst"},
		"force":           false,
	}
	resp := postJSON(t, ts, "/api/apply/copy?token="+testToken, req)
	var out struct {
		Applied []map[string]any `json:"applied"`
		Skipped []map[string]any `json:"skipped"`
	}
	if err := json.Unmarshal(readBody(t, resp), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Applied) != 0 || len(out.Skipped) != 1 || out.Skipped[0]["profileId"] != "dst" {
		t.Fatalf("applied=%+v skipped=%+v, want dst skipped", out.Applied, out.Skipped)
	}

	req["force"] = true
	resp = postJSON(t, ts, "/api/apply/copy?token="+testToken, req)
	out = struct {
		Applied []map[string]any `json:"applied"`
		Skipped []map[string]any `json:"skipped"`
	}{}
	if err := json.Unmarshal(readBody(t, resp), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Applied) != 1 || out.Applied[0]["profileId"] != "dst" {
		t.Fatalf("applied = %+v, want dst applied when forced", out.Applied)
	}
}

func TestApplyCopyBlockedWhenSimRunning(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "src", "f4_key_u.cfg", nil)
	writeProfile(t, dir, "dst", "f4_btn128.cfg", nil)
	before, err := os.ReadFile(filepath.Join(dir, "dst", "controls.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	ts := newServer(t, dir, ServerOpts{SimCheck: func() (bool, error) { return true, nil }})

	req := map[string]any{
		"sourceProfileId": "src",
		"controls":        []string{"TractionControlArm"},
		"profileIds":      []string{"dst"},
	}
	resp := postJSON(t, ts, "/api/apply/copy?token="+testToken, req)
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", resp.StatusCode)
	}
	after, err := os.ReadFile(filepath.Join(dir, "dst", "controls.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("dst/controls.cfg changed despite sim running")
	}
}

func TestApplyCopyRejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "src", "f4_key_u.cfg", nil)
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"sourceProfileId": "src",
		"controls":        []string{"TractionControlArm"},
		"profileIds":      []string{".."},
	}
	resp := postJSON(t, ts, "/api/apply/copy?token="+testToken, req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", resp.StatusCode)
	}
}

func TestApplyCopySkipsNoChanges(t *testing.T) {
	dir := t.TempDir()
	// Ignition is unbound in f4_key_u.cfg, so copying it is a no-op.
	writeProfile(t, dir, "src", "f4_key_u.cfg", nil)
	writeProfile(t, dir, "dst", "f4_key_u.cfg", nil)
	before, err := os.ReadFile(filepath.Join(dir, "dst", "controls.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"sourceProfileId": "src",
		"controls":        []string{"Ignition"},
		"profileIds":      []string{"dst"},
	}
	resp := postJSON(t, ts, "/api/apply/copy?token="+testToken, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", resp.StatusCode, readBody(t, resp))
	}
	var out struct {
		Applied []map[string]any `json:"applied"`
		Skipped []map[string]any `json:"skipped"`
	}
	if err := json.Unmarshal(readBody(t, resp), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Applied) != 0 {
		t.Fatalf("applied = %+v, want 0 (no-op must not be applied)", out.Applied)
	}
	if len(out.Skipped) != 1 || out.Skipped[0]["profileId"] != "dst" {
		t.Fatalf("skipped = %+v, want exactly dst", out.Skipped)
	}
	after, err := os.ReadFile(filepath.Join(dir, "dst", "controls.cfg"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("dst/controls.cfg changed for a no-op copy")
	}
	entries, err := os.ReadDir(filepath.Join(dir, "dst"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if matched, _ := filepath.Match("controls.cfg.bak-*", e.Name()); matched {
			t.Fatalf("unexpected backup %s for a no-op copy", e.Name())
		}
	}
}

func TestApplyCopySkipsSourceProfile(t *testing.T) {
	dir := t.TempDir()
	writeProfile(t, dir, "src", "f4_key_u.cfg", nil)
	writeProfile(t, dir, "dst", "f4_btn128.cfg", nil)
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"sourceProfileId": "src",
		"controls":        []string{"TractionControlArm"},
		"profileIds":      []string{"src", "dst"},
	}
	resp := postJSON(t, ts, "/api/apply/copy?token="+testToken, req)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 (%s)", resp.StatusCode, readBody(t, resp))
	}
	var out struct {
		Applied []map[string]any `json:"applied"`
		Skipped []map[string]any `json:"skipped"`
	}
	if err := json.Unmarshal(readBody(t, resp), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Applied) != 1 || out.Applied[0]["profileId"] != "dst" {
		t.Fatalf("applied = %+v, want exactly dst", out.Applied)
	}
	if len(out.Skipped) != 1 || out.Skipped[0]["profileId"] != "src" {
		t.Fatalf("skipped = %+v, want exactly src (source excluded)", out.Skipped)
	}
}

type fakeCapturer struct {
	result *inputcap.CaptureResult
	err    error
}

func (f fakeCapturer) Capture(opts inputcap.CaptureOpts) (*inputcap.CaptureResult, error) {
	return f.result, f.err
}

func TestInputCaptureUsesInjectedCapturer(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{
		InputCapturer: fakeCapturer{result: &inputcap.CaptureResult{
			Button:             32,
			DeviceInstanceGUID: "{991805D0-2EB6-11F1-800A-444553540000}",
			ProductGUID:        "{0002346E-0000-0000-0000-504944564944}",
		}},
	})

	resp := postJSON(t, ts, "/api/input/capture?token="+testToken, map[string]any{"timeoutMs": 1000})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var result map[string]any
	if err := json.Unmarshal(readBody(t, resp), &result); err != nil {
		t.Fatal(err)
	}
	if int(result["button"].(float64)) != 32 {
		t.Fatalf("button = %v, want 32", result["button"])
	}
}

func TestInputDevicesFallsBackToCalib(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	resp := get(t, ts, "/api/input/devices?token="+testToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	var devices []map[string]any
	if err := json.Unmarshal(readBody(t, resp), &devices); err != nil {
		t.Fatal(err)
	}
	if len(devices) != 2 {
		t.Fatalf("len = %d, want 2 (Pedal PRS + MOZA R9 Base)", len(devices))
	}
	found := false
	for _, d := range devices {
		if d["name"] == "MOZA R9 Base" {
			found = true
		}
	}
	if !found {
		t.Fatalf("devices = %+v, want MOZA R9 Base present", devices)
	}
}

func TestApplyBindingRejectsTraversal(t *testing.T) {
	dir := setupProfilesDir(t, map[string]string{"f4": "f4_key_u.cfg"})
	ts := newServer(t, dir, ServerOpts{})

	req := map[string]any{
		"control":    "Ignition",
		"binding":    map[string]any{"key": "I"},
		"profileIds": []string{".."},
		"force":      false,
	}
	resp := postJSON(t, ts, "/api/apply/binding?token="+testToken, req)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for traversal profile id", resp.StatusCode)
	}
}

func contains(s, sub string) bool {
	return bytes.Contains([]byte(s), []byte(sub))
}
