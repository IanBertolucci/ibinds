package labels

import "testing"

func TestCarFallback(t *testing.T) {
	if got := Car("f4"); got != "f4" {
		t.Fatalf("Car(f4) = %q, want f4", got)
	}
	if got := Car("totallyunknowncar"); got != "totallyunknowncar" {
		t.Fatalf("Car(unknown) = %q, want the id itself", got)
	}
}

func TestControlFallback(t *testing.T) {
	if got := Control("LowFuelAccept"); got != "LowFuelAccept" {
		t.Fatalf("Control(LowFuelAccept) = %q, want LowFuelAccept", got)
	}
	if got := Control("NotARealControl"); got != "NotARealControl" {
		t.Fatalf("Control(unknown) = %q, want the id itself", got)
	}
}
