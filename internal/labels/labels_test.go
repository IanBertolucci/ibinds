package labels

import "testing"

func TestCarFallback(t *testing.T) {
	if got := Car("audigto"); got != "audigto" {
		t.Fatalf("Car(audigto) = %q, want audigto", got)
	}
	if got := Car("totallyunknowncar"); got != "totallyunknowncar" {
		t.Fatalf("Car(unknown) = %q, want the id itself", got)
	}
}

func TestControlFallback(t *testing.T) {
	if got := Control("TractionControlArm"); got != "TractionControlArm" {
		t.Fatalf("Control(TractionControlArm) = %q, want TractionControlArm", got)
	}
	if got := Control("NotARealControl"); got != "NotARealControl" {
		t.Fatalf("Control(unknown) = %q, want the id itself", got)
	}
}

func TestKnownMappings(t *testing.T) {
	if got := Car("f4"); got != "F4" {
		t.Fatalf("Car(f4) = %q, want F4", got)
	}
	if got := Control("LowFuelAccept"); got != "Low Fuel Accept" {
		t.Fatalf("Control(LowFuelAccept) = %q, want \"Low Fuel Accept\"", got)
	}
}
