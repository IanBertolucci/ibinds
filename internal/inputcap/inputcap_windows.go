//go:build windows

package inputcap

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DirectInput8 bindings, hand-written COM vtables against dinput8.dll.
// See docs/FORMAT.md. HRESULT 0 (DI_OK) indicates success.

var (
	dinput8                = syscall.NewLazyDLL("dinput8.dll")
	procDirectInput8Create = dinput8.NewProc("DirectInput8Create")
	procCdfDIJoystick2     = dinput8.NewProc("c_dfDIJoystick2")
)

const (
	directInputVersion  = 0x0800
	di8DevClassGameCtrl = 4
	diEDFLAttachedOnly  = 0x00000001
	diSCLBackground     = 0x00000008
	diSCLNonExclusive   = 0x00000001
	diOK                = 0
	diEnumContinue      = 1
)

var (
	iidIDirectInput8 = windows.GUID{0xBF798030, 0x483A, 0x4DA2, [8]byte{0xAA, 0x99, 0x5D, 0x64, 0xED, 0x36, 0x97, 0x00}}
	guidSysJoystick  = windows.GUID{0x6F1D2B70, 0xD5A0, 0x11CF, [8]byte{0xBF, 0xC7, 0x44, 0x45, 0x53, 0x54, 0x00, 0x00}}
)

type iDirectInput8Vtbl struct {
	QueryInterface  uintptr
	AddRef          uintptr
	Release         uintptr
	CreateDevice    uintptr
	EnumDevices     uintptr
	GetDeviceStatus uintptr
	RunControlPanel uintptr
	Initialize      uintptr
}

type iDirectInput8 struct {
	vtbl *iDirectInput8Vtbl
}

func (d *iDirectInput8) CreateDevice(rguid *windows.GUID, out **iDirectInputDevice8) uintptr {
	r1, _, _ := syscall.SyscallN(
		d.vtbl.CreateDevice,
		uintptr(unsafe.Pointer(d)),
		uintptr(unsafe.Pointer(rguid)),
		uintptr(unsafe.Pointer(out)),
		0,
	)
	return r1
}

func (d *iDirectInput8) EnumDevices(devType uint32, callback uintptr, ref uintptr, flags uint32) uintptr {
	r1, _, _ := syscall.SyscallN(
		d.vtbl.EnumDevices,
		uintptr(unsafe.Pointer(d)),
		uintptr(devType),
		callback,
		ref,
		uintptr(flags),
	)
	return r1
}

type iDirectInputDevice8Vtbl struct {
	QueryInterface           uintptr
	AddRef                   uintptr
	Release                  uintptr
	GetCapabilities          uintptr
	EnumObjects              uintptr
	GetProperty              uintptr
	SetProperty              uintptr
	Acquire                  uintptr
	Unacquire                uintptr
	GetDeviceState           uintptr
	GetDeviceData            uintptr
	SetDataFormat            uintptr
	SetEventNotification     uintptr
	SetCooperativeLevel      uintptr
	GetObjectInfo            uintptr
	GetDeviceInfo            uintptr
	RunControlPanel          uintptr
	Initialize               uintptr
	CreateEffect             uintptr
	EnumEffects              uintptr
	GetEffectInfo            uintptr
	GetForceFeedbackState    uintptr
	SendForceFeedbackCommand uintptr
	EnumCreatedEffectObjects uintptr
	Escape                   uintptr
	Poll                     uintptr
	SendDeviceData           uintptr
}

type iDirectInputDevice8 struct {
	vtbl *iDirectInputDevice8Vtbl
}

func (d *iDirectInputDevice8) SetDataFormat(format uintptr) uintptr {
	r1, _, _ := syscall.SyscallN(
		d.vtbl.SetDataFormat,
		uintptr(unsafe.Pointer(d)),
		format,
	)
	return r1
}

func (d *iDirectInputDevice8) SetCooperativeLevel(hwnd uintptr, flags uint32) uintptr {
	r1, _, _ := syscall.SyscallN(
		d.vtbl.SetCooperativeLevel,
		uintptr(unsafe.Pointer(d)),
		hwnd,
		uintptr(flags),
	)
	return r1
}

func (d *iDirectInputDevice8) Acquire() uintptr {
	r1, _, _ := syscall.SyscallN(d.vtbl.Acquire, uintptr(unsafe.Pointer(d)))
	return r1
}

func (d *iDirectInputDevice8) GetDeviceState(dataSize uint32, data uintptr) uintptr {
	r1, _, _ := syscall.SyscallN(
		d.vtbl.GetDeviceState,
		uintptr(unsafe.Pointer(d)),
		uintptr(dataSize),
		data,
	)
	return r1
}

func (d *iDirectInputDevice8) GetCapabilities(caps uintptr) uintptr {
	r1, _, _ := syscall.SyscallN(
		d.vtbl.GetCapabilities,
		uintptr(unsafe.Pointer(d)),
		caps,
	)
	return r1
}

// diDeviceInstance mirrors DIDEVICEINSTANCEW.
type diDeviceInstance struct {
	dwSize          uint32
	guidInstance    windows.GUID
	guidProduct     windows.GUID
	dwDevType       uint32
	tszInstanceName [260]uint16
	tszProductName  [260]uint16
	guidFFDriver    windows.GUID
	wUsagePage      uint16
	wUsage          uint16
}

// diDevCaps mirrors DIDEVCAPS.
type diDevCaps struct {
	dwSize                uint32
	dwFlags               uint32
	dwDevType             uint32
	dwAxes                uint32
	dwButtons             uint32
	dwPOVs                uint32
	dwFFSamplePeriod      uint32
	dwFFMinTimeResolution uint32
	dwFirmwareRevision    uint32
	dwHardwareRevision    uint32
	dwFFDriverVersion     uint32
}

// diJoyState2 mirrors DIJOYSTATE2.
type diJoyState2 struct {
	lX, lY, lZ, lRx, lRy, lRz       int32
	rglSlider                       [2]int32
	rgdwPOV                         [4]uint32
	rgbButtons                      [128]byte
	lVX, lVY, lVZ, lVRx, lVRy, lVRz int32
	rglVSlider                      [2]int32
	lAX, lAY, lAZ, lARx, lARy, lARz int32
	rglASlider                      [2]int32
	lFX, lFY, lFZ, lFRx, lFRy, lFRz int32
	rglFSlider                      [2]int32
}

// diDataFormat mirrors DIDATAFORMAT.
type diDataFormat struct {
	dwSize     uint32
	dwObjSize  uint32
	dwFlags    uint32
	dwDataSize uint32
	dwNumObjs  uint32
	rgodf      uintptr
}

var enumInstances []diDeviceInstance

//go:uintptrescapes
func enumDevicesCallback(inst *diDeviceInstance, ref uintptr) uintptr {
	enumInstances = append(enumInstances, *inst)
	return diEnumContinue
}

type directInputCapturer struct{}

// New returns the DirectInput capturer.
func New() Capturer {
	return directInputCapturer{}
}

func createInput() (*iDirectInput8, error) {
	var p uintptr
	r1, _, _ := procDirectInput8Create.Call(
		0,
		directInputVersion,
		uintptr(unsafe.Pointer(&iidIDirectInput8)),
		uintptr(unsafe.Pointer(&p)),
		0,
	)
	if r1 != 0 {
		return nil, fmt.Errorf("inputcap: DirectInput8Create failed: 0x%x", r1)
	}
	return (*iDirectInput8)(unsafe.Pointer(p)), nil
}

func enumerateJoysticks() ([]diDeviceInstance, error) {
	di, err := createInput()
	if err != nil {
		return nil, err
	}
	enumInstances = nil
	hr := di.EnumDevices(di8DevClassGameCtrl, windows.NewCallback(enumDevicesCallback), 0, diEDFLAttachedOnly)
	if hr != diOK {
		return nil, fmt.Errorf("inputcap: EnumDevices failed: 0x%x", hr)
	}
	out := append([]diDeviceInstance(nil), enumInstances...)
	return out, nil
}

// Enumerate returns connected DirectInput game controllers.
func Enumerate() ([]DeviceInfo, error) {
	instances, err := enumerateJoysticks()
	if err != nil {
		return nil, err
	}
	out := make([]DeviceInfo, 0, len(instances))
	for i := range instances {
		inst := &instances[i]
		info := DeviceInfo{
			Name:         windows.UTF16ToString(inst.tszInstanceName[:]),
			InstanceGUID: formatGUID(inst.guidInstance),
			ProductGUID:  formatProductGUID(uint32(inst.guidProduct.Data1)),
		}
		if dev, err := openDevice(inst); err == nil {
			info.ButtonCount = deviceButtonCount(dev)
		}
		out = append(out, info)
	}
	return out, nil
}

func openDevice(inst *diDeviceInstance) (*iDirectInputDevice8, error) {
	di, err := createInput()
	if err != nil {
		return nil, err
	}
	var dev *iDirectInputDevice8
	hr := di.CreateDevice(&inst.guidInstance, &dev)
	if hr != diOK {
		return nil, fmt.Errorf("inputcap: CreateDevice failed: 0x%x", hr)
	}
	return dev, nil
}

func deviceButtonCount(dev *iDirectInputDevice8) int {
	var caps diDevCaps
	caps.dwSize = uint32(unsafe.Sizeof(caps))
	if hr := dev.GetCapabilities(uintptr(unsafe.Pointer(&caps))); hr != diOK {
		return 0
	}
	return int(caps.dwButtons)
}

// Capture polls all joysticks until a button is pressed or the timeout
// elapses. A nil result with nil error indicates a timeout.
func (directInputCapturer) Capture(opts CaptureOpts) (*CaptureResult, error) {
	instances, err := enumerateJoysticks()
	if err != nil {
		return nil, err
	}
	if len(instances) == 0 {
		return nil, ErrUnsupported
	}

	type open struct {
		inst *diDeviceInstance
		dev  *iDirectInputDevice8
	}
	devs := make([]open, 0, len(instances))
	for i := range instances {
		inst := &instances[i]
		dev, err := openDevice(inst)
		if err != nil {
			continue
		}
		if hr := dev.SetDataFormat(procCdfDIJoystick2.Addr()); hr != diOK {
			continue
		}
		if hr := dev.SetCooperativeLevel(0, diSCLBackground|diSCLNonExclusive); hr != diOK {
			continue
		}
		devs = append(devs, open{inst: inst, dev: dev})
	}
	if len(devs) == 0 {
		return nil, ErrUnsupported
	}

	deadline := time.Now().Add(opts.timeout())
	for {
		for _, d := range devs {
			var state diJoyState2
			if d.dev.Acquire() != diOK {
				continue
			}
			if hr := d.dev.GetDeviceState(uint32(unsafe.Sizeof(state)), uintptr(unsafe.Pointer(&state))); hr != diOK {
				continue
			}
			for i, b := range state.rgbButtons {
				if b&0x80 != 0 {
					return &CaptureResult{
						Button:             i,
						DeviceInstanceGUID: formatGUID(d.inst.guidInstance),
						ProductGUID:        formatProductGUID(uint32(d.inst.guidProduct.Data1)),
					}, nil
				}
			}
		}
		if time.Now().After(deadline) {
			return nil, nil
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func formatGUID(g windows.GUID) string {
	return fmt.Sprintf("{%08X-%04X-%04X-%02X%02X-%02X%02X%02X%02X%02X%02X}",
		g.Data1, g.Data2, g.Data3,
		g.Data4[0], g.Data4[1], g.Data4[2], g.Data4[3], g.Data4[4], g.Data4[5], g.Data4[6], g.Data4[7])
}

func formatProductGUID(pid uint32) string {
	return fmt.Sprintf("{%08X-0000-0000-0000-504944564944}", pid)
}
