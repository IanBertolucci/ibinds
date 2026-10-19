//go:build !windows

package inputcap

type stubCapturer struct{}

func (stubCapturer) Capture(opts CaptureOpts) (*CaptureResult, error) {
	return nil, ErrUnsupported
}

// New returns a capturer. On non-Windows platforms capture is unsupported.
func New() Capturer {
	return stubCapturer{}
}

// Enumerate returns connected devices. On non-Windows platforms it is
// unsupported.
func Enumerate() ([]DeviceInfo, error) {
	return nil, ErrUnsupported
}
