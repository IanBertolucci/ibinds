package cfgparse

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return data
}

func parseFixture(t *testing.T, name string) *File {
	t.Helper()
	f, err := Parse(loadFixture(t, name))
	if err != nil {
		t.Fatalf("Parse(%s): %v", name, err)
	}
	return f
}

func findRecord(t *testing.T, f *File, name string) *Record {
	t.Helper()
	for i := range f.Records {
		if f.Records[i].Name == name {
			return &f.Records[i]
		}
	}
	t.Fatalf("record %q not found", name)
	return nil
}

func TestParseRejectsBadMagic(t *testing.T) {
	data := make([]byte, 64)
	copy(data, "XXXX")
	if _, err := Parse(data); !errors.Is(err, ErrBadMagic) {
		t.Fatalf("err = %v, want ErrBadMagic", err)
	}
}

func TestParseRejectsTruncated(t *testing.T) {
	full := loadFixture(t, "f4_key_u.cfg")
	cut := full[:len(full)/2]
	if _, err := Parse(cut); !errors.Is(err, ErrTruncated) {
		t.Fatalf("err = %v, want ErrTruncated", err)
	}
}

func TestParseRealBtn130(t *testing.T) {
	f := parseFixture(t, "f4_btn130.cfg")
	r := findRecord(t, f, "TractionControlArm")
	if r.Kind != KindButton {
		t.Fatalf("Kind = %v, want KindButton", r.Kind)
	}
	if r.Button != 130 {
		t.Fatalf("Button = %d, want 130", r.Button)
	}
	if r.Device == nil || r.Device.InstanceGUID != "{991805D0-2EB6-11F1-800A-444553540000}" {
		t.Fatalf("Device = %+v, want MOZA InstanceGUID", r.Device)
	}
}

func TestParseRealBtn128(t *testing.T) {
	f := parseFixture(t, "f4_btn128.cfg")
	r := findRecord(t, f, "TractionControlArm")
	if r.Kind != KindButton || r.Button != 128 {
		t.Fatalf("record = %+v, want KindButton Button=128", r)
	}
}

func TestParseRealKeyU(t *testing.T) {
	f := parseFixture(t, "f4_key_u.cfg")
	r := findRecord(t, f, "TractionControlArm")
	if r.Kind != KindKey || r.Key != 'U' {
		t.Fatalf("record = %+v, want KindKey Key='U'", r)
	}
}

func TestParseRealCurrent(t *testing.T) {
	f := parseFixture(t, "f4_key_u.cfg")

	tca := findRecord(t, f, "TractionControlArm")
	if tca.Kind != KindKey || tca.Key != 'U' {
		t.Fatalf("TractionControlArm = %+v, want KindKey Key='U'", tca)
	}

	lfa := findRecord(t, f, "LowFuelAccept")
	if lfa.Kind != KindKey || lfa.Key != 'N' {
		t.Fatalf("LowFuelAccept = %+v, want KindKey Key='N'", lfa)
	}

	ign := findRecord(t, f, "Ignition")
	if ign.Kind != KindUnbound {
		t.Fatalf("Ignition = %+v, want KindUnbound", ign)
	}
}

func TestParseAxisRecordsPreserved(t *testing.T) {
	f := parseFixture(t, "f4_key_u.cfg")

	sl := findRecord(t, f, "SteerLeft")
	if sl.Kind != KindAxis {
		t.Fatalf("SteerLeft Kind = %v, want KindAxis", sl.Kind)
	}

	thr := findRecord(t, f, "Throttle")
	if thr.Kind != KindAxis {
		t.Fatalf("Throttle Kind = %v, want KindAxis", thr.Kind)
	}

	for _, name := range []string{"SteerLeft", "Throttle"} {
		r := findRecord(t, f, name)
		got := r.Bytes()
		want := f.Raw[r.Offset : r.Offset+len(got)]
		if len(got) == 0 || string(got) != string(want) {
			t.Fatalf("%s Bytes() does not round-trip", name)
		}
	}
}

func TestRoundTripIdentity(t *testing.T) {
	for _, name := range []string{"f4_btn130.cfg", "f4_btn128.cfg", "f4_key_u.cfg"} {
		raw := loadFixture(t, name)
		f, err := Parse(raw)
		if err != nil {
			t.Fatalf("Parse(%s): %v", name, err)
		}
		var b []byte
		b = append(b, raw[:f.Records[0].Offset]...)
		for _, r := range f.Records {
			b = append(b, r.Bytes()...)
		}
		if string(b) != string(raw) {
			t.Fatalf("%s round-trip mismatch (len %d vs %d)", name, len(b), len(raw))
		}
	}
}
