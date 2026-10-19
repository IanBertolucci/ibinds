package cfgparse

// BindingKind classifies a control record's binding.
type BindingKind int

const (
	KindUnbound BindingKind = iota
	KindKey
	KindButton
	KindAxis
	KindUnknown
)

// Device identifies a DirectInput device by its GUID strings (the same
// "{...}" format used by joyCalib.yaml).
type Device struct {
	InstanceGUID string
	ProductGUID  string
}

// Record is one control binding record. Opaque holds the record's full byte
// range (name + NUL + fixed region) so that reserialization is byte-exact.
type Record struct {
	Name         string
	Offset       int // absolute offset of the record's first byte (the name)
	PayloadStart int // absolute offset of the 68-byte fixed region
	Kind         BindingKind
	Key          byte  // ASCII key, for KindKey
	Button       int   // lowest set button, for KindButton
	Buttons      []int // all set buttons, for KindButton
	Device       *Device
	Opaque       []byte
}

// Bytes returns the record's original bytes, identical to the source.
func (r *Record) Bytes() []byte {
	return r.Opaque
}

// Offsets (relative to PayloadStart) of the fixed 68-byte region, exported
// for the editor package. See docs/FORMAT.md.
const (
	FixedRegionSize = 68
	TypeOffset      = 8
	KeyOffset       = 12
	MaskWordsOffset = 12
	MaskWordCount   = 5
	GUIDOffset      = 36
	ProductOffset   = 52
	PIDVIDOffset    = 62
)
