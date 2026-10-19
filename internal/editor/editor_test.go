package editor

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"ibinds/internal/cfgparse"
)

const (
	mozaGUID     = "{991805D0-2EB6-11F1-800A-444553540000}"
	mozaProdGUID = "{0002346E-0000-0000-0000-504944564944}"
)

func loadFile(t *testing.T, name string) *cfgparse.File {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "cfgparse", "testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	f, err := cfgparse.Parse(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	return f
}

func loadBytes(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "cfgparse", "testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return data
}

func findRecord(f *cfgparse.File, name string) *cfgparse.Record {
	for i := range f.Records {
		if f.Records[i].Name == name {
			return &f.Records[i]
		}
	}
	return nil
}

func TestApplyKeyReproducesRealFixture(t *testing.T) {
	f := loadFile(t, "f4_btn128.cfg")
	out, rec, err := ApplyBinding(f, "TractionControlArm", Binding{Key: 'U', HasKey: true})
	if err != nil {
		t.Fatalf("ApplyBinding: %v", err)
	}
	if rec.Kind != cfgparse.KindKey || rec.Key != 'U' {
		t.Fatalf("record = %+v, want KindKey Key='U'", rec)
	}
	if want := loadBytes(t, "f4_key_u.cfg"); !bytes.Equal(out, want) {
		t.Fatalf("output not byte-identical to f4_key_u.cfg")
	}
}

func TestApplyButtonReplacementReproducesRealFixture(t *testing.T) {
	f := loadFile(t, "f4_btn130.cfg")
	out, rec, err := ApplyBinding(f, "TractionControlArm", Binding{
		Button: 128, HasButton: true,
		DeviceInstanceGUID: mozaGUID,
		ProductGUID:        mozaProdGUID,
	})
	if err != nil {
		t.Fatalf("ApplyBinding: %v", err)
	}
	if rec.Kind != cfgparse.KindButton || rec.Button != 128 {
		t.Fatalf("record = %+v, want KindButton Button=128", rec)
	}
	if want := loadBytes(t, "f4_btn128.cfg"); !bytes.Equal(out, want) {
		t.Fatalf("output not byte-identical to f4_btn128.cfg")
	}
}

func TestApplyKeyToButtonReproducesRealFixture(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	out, rec, err := ApplyBinding(f, "TractionControlArm", Binding{
		Button: 130, HasButton: true,
		DeviceInstanceGUID: mozaGUID,
		ProductGUID:        mozaProdGUID,
	})
	if err != nil {
		t.Fatalf("ApplyBinding: %v", err)
	}
	if rec.Kind != cfgparse.KindButton || rec.Button != 130 {
		t.Fatalf("record = %+v, want KindButton Button=130", rec)
	}
	if want := loadBytes(t, "f4_btn130.cfg"); !bytes.Equal(out, want) {
		t.Fatalf("output not byte-identical to f4_btn130.cfg")
	}
}

func TestApplyToUnbound(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	out, rec, err := ApplyBinding(f, "Ignition", Binding{Key: 'I', HasKey: true})
	if err != nil {
		t.Fatalf("ApplyBinding: %v", err)
	}
	if rec.Kind != cfgparse.KindKey || rec.Key != 'I' {
		t.Fatalf("record = %+v, want KindKey Key='I'", rec)
	}
	if len(out) != len(f.Raw) {
		t.Fatalf("len(out) = %d, want %d", len(out), len(f.Raw))
	}
}

func TestApplyPreservesUnknownBytes(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	target := findRecord(f, "TractionControlArm")
	if target == nil {
		t.Fatal("TractionControlArm not found")
	}
	out, _, err := ApplyBinding(f, "TractionControlArm", Binding{Key: 'I', HasKey: true})
	if err != nil {
		t.Fatalf("ApplyBinding: %v", err)
	}

	// Edited field ranges (relative to PayloadStart): type +8..12, key/mask +12..32, GUID+PID+PIDVID +36..68.
	inEdited := func(off int) bool {
		rel := off - target.PayloadStart
		return (rel >= cfgparse.TypeOffset && rel < cfgparse.TypeOffset+4) ||
			(rel >= cfgparse.KeyOffset && rel < cfgparse.KeyOffset+20) ||
			(rel >= cfgparse.GUIDOffset && rel < cfgparse.FixedRegionSize)
	}
	for i := range f.Raw {
		if f.Raw[i] != out[i] && !inEdited(i) {
			t.Fatalf("byte %d (0x%x) changed outside edited fields: %02x -> %02x", i, i, f.Raw[i], out[i])
		}
	}
}

func TestApplyUnknownControl(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	_, _, err := ApplyBinding(f, "NoSuchControl", Binding{Key: 'U', HasKey: true})
	if !errors.Is(err, ErrUnknownControl) {
		t.Fatalf("err = %v, want ErrUnknownControl", err)
	}
}

func TestApplyRefusesAxisKind(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	_, _, err := ApplyBinding(f, "SteerLeft", Binding{Key: 'U', HasKey: true})
	if !errors.Is(err, ErrUnsupportedKind) {
		t.Fatalf("err = %v, want ErrUnsupportedKind", err)
	}
}

func TestWriteRefusesUnvalidatedInput(t *testing.T) {
	f := loadFile(t, "f4_key_u.cfg")
	corrupted := append([]byte(nil), f.Raw...)
	corrupted[0] = 'X'
	f.Raw = corrupted

	_, _, err := ApplyBinding(f, "TractionControlArm", Binding{Key: 'U', HasKey: true})
	if !errors.Is(err, cfgparse.ErrBadMagic) {
		t.Fatalf("err = %v, want cfgparse.ErrBadMagic", err)
	}
}
