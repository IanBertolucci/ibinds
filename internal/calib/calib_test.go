package calib

import (
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

func TestParseRealF4Fixture(t *testing.T) {
	c, err := Parse(loadFixture(t, "f4_joycalib.yaml"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(c.Devices) != 2 {
		t.Fatalf("len(Devices) = %d, want 2", len(c.Devices))
	}

	moza := c.Devices[1]
	if moza.Name != "MOZA R9 Base" {
		t.Fatalf("Devices[1].Name = %q, want MOZA R9 Base", moza.Name)
	}
	if moza.InstanceGUID != "{991805D0-2EB6-11F1-800A-444553540000}" {
		t.Fatalf("InstanceGUID = %q", moza.InstanceGUID)
	}
	wantAxes := []CalibAxis{
		{Axis: 0, Name: "X Axis", Min: 0, Center: 32766, Max: 65535},
		{Axis: 4, Name: "Y Rotation", Min: -32767, Center: -32767, Max: 32767},
	}
	if len(moza.Axes) != len(wantAxes) {
		t.Fatalf("len(Axes) = %d, want %d", len(moza.Axes), len(wantAxes))
	}
	for i := range wantAxes {
		if moza.Axes[i] != wantAxes[i] {
			t.Fatalf("Axes[%d] = %+v, want %+v", i, moza.Axes[i], wantAxes[i])
		}
	}
}

func TestParseDevicesSynthetic(t *testing.T) {
	data := []byte(
		"CalibrationInfo:\r\n" +
			" DeviceList:\r\n" +
			" - DeviceName: 'Wheel A'\r\n" +
			"   InstanceGUID: '{11111111-2222-3333-4444-555555555555}'\r\n" +
			"   ProductGUID: '{AAAAAAAA-0000-0000-0000-504944564944}'\r\n" +
			"   AxisList:\r\n" +
			"   - Axis: 0\r\n" +
			"     AxisName: 'X Axis'\r\n" +
			"     CalibMin: 1\r\n" +
			"     CalibCenter: 2\r\n" +
			"     CalibMax: 3\r\n" +
			" - DeviceName: 'Wheel B'\r\n" +
			"   InstanceGUID: '{99999999-8888-7777-6666-555555555555}'\r\n" +
			"   ProductGUID: '{BBBBBBBB-0000-0000-0000-504944564944}'\r\n" +
			"   AxisList:\r\n" +
			"   - Axis: 1\r\n" +
			"     AxisName: 'Brake'\r\n" +
			"     CalibMin: -10\r\n" +
			"     CalibCenter: 0\r\n" +
			"     CalibMax: 10\r\n",
	)
	c, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(c.Devices) != 2 {
		t.Fatalf("len(Devices) = %d, want 2", len(c.Devices))
	}
	if c.Devices[0].Name != "Wheel A" || c.Devices[1].Name != "Wheel B" {
		t.Fatalf("names = %q, %q", c.Devices[0].Name, c.Devices[1].Name)
	}
	if c.Devices[0].InstanceGUID != "{11111111-2222-3333-4444-555555555555}" {
		t.Fatalf("Devices[0].InstanceGUID = %q", c.Devices[0].InstanceGUID)
	}
	if c.Devices[0].ProductGUID != "{AAAAAAAA-0000-0000-0000-504944564944}" {
		t.Fatalf("Devices[0].ProductGUID = %q", c.Devices[0].ProductGUID)
	}
	if len(c.Devices[0].Axes) != 1 {
		t.Fatalf("len(Devices[0].Axes) = %d, want 1", len(c.Devices[0].Axes))
	}
	wantAxis := CalibAxis{Axis: 0, Name: "X Axis", Min: 1, Center: 2, Max: 3}
	if c.Devices[0].Axes[0] != wantAxis {
		t.Fatalf("Devices[0].Axes[0] = %+v, want %+v", c.Devices[0].Axes[0], wantAxis)
	}
	wantBrake := CalibAxis{Axis: 1, Name: "Brake", Min: -10, Center: 0, Max: 10}
	if c.Devices[1].Axes[0] != wantBrake {
		t.Fatalf("Devices[1].Axes[0] = %+v, want %+v", c.Devices[1].Axes[0], wantBrake)
	}
}

func TestValidateAcceptsValidYAML(t *testing.T) {
	if err := Validate(loadFixture(t, "f4_joycalib.yaml")); err != nil {
		t.Fatalf("Validate(f4) = %v, want nil", err)
	}
}

func TestValidateAcceptsEmptyDeviceList(t *testing.T) {
	data := []byte("CalibrationInfo:\n")
	if err := Validate(data); err != nil {
		t.Fatalf("Validate(empty) = %v, want nil", err)
	}
}

func TestValidateRejectsGarbage(t *testing.T) {
	if err := Validate([]byte("not: [valid")); err == nil {
		t.Fatal("Validate(garbage) = nil, want error")
	}
}
