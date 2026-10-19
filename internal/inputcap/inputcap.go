package inputcap

import (
	"errors"
	"time"
)

// ErrUnsupported is returned by capture/enumeration on platforms without
// DirectInput support (or when no device is available).
var ErrUnsupported = errors.New("input capture only supported on Windows")

// DefaultTimeout is used when CaptureOpts.Timeout is zero.
const DefaultTimeout = 10 * time.Second

// CaptureOpts configures a capture request.
type CaptureOpts struct {
	Timeout time.Duration
}

func (o CaptureOpts) timeout() time.Duration {
	if o.Timeout <= 0 {
		return DefaultTimeout
	}
	return o.Timeout
}

// CaptureResult is the button/device captured.
type CaptureResult struct {
	Button             int    `json:"button"`
	DeviceInstanceGUID string `json:"deviceInstanceGUID"`
	ProductGUID        string `json:"productGUID"`
}

// Capturer captures a button press from a connected device.
type Capturer interface {
	Capture(opts CaptureOpts) (*CaptureResult, error)
}

// DeviceInfo describes a connected DirectInput device.
type DeviceInfo struct {
	Name         string `json:"name"`
	InstanceGUID string `json:"instanceGUID"`
	ProductGUID  string `json:"productGUID"`
	ButtonCount  int    `json:"buttonCount"`
}
