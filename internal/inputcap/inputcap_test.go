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

func TestPOVButtonMapping(t *testing.T) {
	// iRacing reserves buttons 128..159 for four POV hats, eight directions
	// each, starting at up and going clockwise.
	cases := []struct {
		hat   int
		value uint32
		want  int
	}{
		{0, 0, 128},     // up
		{0, 4500, 129},  // up-right
		{0, 9000, 130},  // right
		{0, 13500, 131}, // down-right
		{0, 18000, 132}, // down
		{0, 22500, 133}, // down-left
		{0, 27000, 134}, // left
		{0, 31500, 135}, // up-left
		{0, 35999, 128}, // wrapped to up
		{1, 0, 136},     // second hat starts at 136
		{3, 31500, 159}, // last hat, up-left
	}
	for _, c := range cases {
		got, ok := povButton(c.hat, c.value)
		if !ok || got != c.want {
			t.Errorf("povButton(%d, %d) = %d, %v; want %d, true", c.hat, c.value, got, ok, c.want)
		}
	}

	if _, ok := povButton(0, povCentered); ok {
		t.Error("centered POV should not produce a button")
	}
	if _, ok := povButton(0, 40000); ok {
		t.Error("invalid POV angle should not produce a button")
	}
}
