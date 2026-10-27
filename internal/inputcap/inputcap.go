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
	// RawPOV holds the four POV values at capture time for diagnostics; it is
	// not part of the API response.
	RawPOV [4]uint32 `json:"-"`
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
	POVCount     int    `json:"povCount"`
	AxisCount    int    `json:"axisCount"`
}

// POV hats are not DirectInput buttons, but iRacing exposes their directions as
// virtual buttons. Physical buttons occupy 0..127; the four POV hats use
// 128..159, eight directions per hat (up, up-right, right, ... up-left).
const (
	povButtonBase  = 128
	povDirections  = 8
	povCentered    = 0xFFFFFFFF
	povMaxAngle    = 35999
	povDegreesStep = 36000 / povDirections
)

// povButton maps a DirectInput POV value (hundredths of a degree) for hat to
// iRacing's virtual button number. The second result is false when the hat is
// centered or the value is invalid.
func povButton(hat int, value uint32) (int, bool) {
	if value == povCentered || value > povMaxAngle {
		return 0, false
	}
	half := uint32(povDegreesStep / 2)
	dir := int((value+half)/uint32(povDegreesStep)) % povDirections
	return povButtonBase + hat*povDirections + dir, true
}
