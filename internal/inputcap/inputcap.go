package inputcap

import (
	"errors"
	"time"
)

// ErrUnsupported is returned by capture on platforms without DirectInput
// support (or when no capturer is available).
var ErrUnsupported = errors.New("input capture only supported on Windows")

// CaptureOpts configures a capture request.
type CaptureOpts struct {
	Timeout time.Duration
}

// CaptureResult is the button/device captured.
type CaptureResult struct {
	Button             int
	DeviceInstanceGUID string
	ProductGUID        string
}

// Capturer captures a button press from a connected device.
type Capturer interface {
	Capture(opts CaptureOpts) (*CaptureResult, error)
}
