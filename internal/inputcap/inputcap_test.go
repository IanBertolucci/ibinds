package inputcap

import (
	"errors"
	"testing"
	"time"
)

func TestCaptureOptsDefaults(t *testing.T) {
	zero := CaptureOpts{}
	if got := zero.timeout(); got != DefaultTimeout {
		t.Fatalf("zero timeout = %v, want %v", got, DefaultTimeout)
	}
	explicit := CaptureOpts{Timeout: 5 * time.Second}
	if got := explicit.timeout(); got != 5*time.Second {
		t.Fatalf("explicit timeout = %v, want 5s", got)
	}
}

func TestUnsupportedOnNonWindows(t *testing.T) {
	c := New()
	_, err := c.Capture(CaptureOpts{})
	if !errors.Is(err, ErrUnsupported) {
		t.Fatalf("err = %v, want ErrUnsupported", err)
	}
}
