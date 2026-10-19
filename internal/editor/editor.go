package editor

import (
	"encoding/binary"
	"errors"

	"ibinds/internal/cfgparse"
)

var (
	ErrUnknownControl  = errors.New("editor: unknown control")
	ErrUnsupportedKind = errors.New("editor: unsupported record kind")
	errInvalidBinding  = errors.New("editor: exactly one of HasKey/HasButton must be set")
)

// Binding describes a target binding for a control. Exactly one of HasKey and
// HasButton must be true.
type Binding struct {
	Key                byte
	HasKey             bool
	Button             int
	HasButton          bool
	DeviceInstanceGUID string
	ProductGUID        string
}

const (
	typeButton = 2
	typeKey    = 4
)

// ApplyBinding edits a single control record in-place and returns the new
// file bytes plus the reparsed record. Only unbound/key/button records are
// editable; axis and unknown records are refused.
func ApplyBinding(f *cfgparse.File, control string, b Binding) ([]byte, cfgparse.Record, error) {
	if !cfgparse.ValidMagic(f.Raw) {
		return nil, cfgparse.Record{}, cfgparse.ErrBadMagic
	}

	var rec *cfgparse.Record
	for i := range f.Records {
		if f.Records[i].Name == control {
			rec = &f.Records[i]
			break
		}
	}
	if rec == nil {
		return nil, cfgparse.Record{}, ErrUnknownControl
	}
	if rec.Kind == cfgparse.KindAxis || rec.Kind == cfgparse.KindUnknown {
		return nil, cfgparse.Record{}, ErrUnsupportedKind
	}
	if b.HasKey == b.HasButton {
		return nil, cfgparse.Record{}, errInvalidBinding
	}

	out := make([]byte, len(f.Raw))
	copy(out, f.Raw)
	fixed := rec.PayloadStart

	switch {
	case b.HasKey:
		binary.LittleEndian.PutUint32(out[fixed+cfgparse.TypeOffset:], typeKey)
		clearBytes(out[fixed+cfgparse.KeyOffset : fixed+cfgparse.KeyOffset+20])
		out[fixed+cfgparse.KeyOffset] = b.Key
		clearBytes(out[fixed+cfgparse.GUIDOffset : fixed+cfgparse.FixedRegionSize])
	case b.HasButton:
		if b.Button < 0 || b.Button >= cfgparse.MaskWordCount*32 {
			return nil, cfgparse.Record{}, errors.New("editor: button out of range")
		}
		binary.LittleEndian.PutUint32(out[fixed+cfgparse.TypeOffset:], typeButton)
		clearBytes(out[fixed+cfgparse.MaskWordsOffset : fixed+cfgparse.MaskWordsOffset+cfgparse.MaskWordCount*4])
		word := b.Button / 32
		binary.LittleEndian.PutUint32(out[fixed+cfgparse.MaskWordsOffset+word*4:], 1<<uint(b.Button%32))

		guid, err := cfgparse.ParseGUID(b.DeviceInstanceGUID)
		if err != nil {
			return nil, cfgparse.Record{}, err
		}
		copy(out[fixed+cfgparse.GUIDOffset:fixed+cfgparse.GUIDOffset+16], guid)

		pid, err := cfgparse.ParseProductGUID(b.ProductGUID)
		if err != nil {
			return nil, cfgparse.Record{}, err
		}
		copy(out[fixed+cfgparse.ProductOffset:fixed+cfgparse.ProductOffset+4], pid)
		clearBytes(out[fixed+cfgparse.ProductOffset+4 : fixed+cfgparse.PIDVIDOffset])
		copy(out[fixed+cfgparse.PIDVIDOffset:fixed+cfgparse.PIDVIDOffset+6], []byte("PIDVID"))
	}

	newFile, err := cfgparse.Parse(out)
	if err != nil {
		return nil, cfgparse.Record{}, err
	}
	for i := range newFile.Records {
		if newFile.Records[i].Name == control {
			return out, newFile.Records[i], nil
		}
	}
	return nil, cfgparse.Record{}, ErrUnknownControl
}

func clearBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
