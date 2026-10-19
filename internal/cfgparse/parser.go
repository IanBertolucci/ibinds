package cfgparse

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math/bits"
)

var (
	ErrBadMagic  = errors.New("cfgparse: bad magic")
	ErrTruncated = errors.New("cfgparse: truncated")
)

var (
	magic      = []byte("GFCC")
	chunkMagic = []byte("LRTC")
)

// File is a parsed controls.cfg.
type File struct {
	Raw     []byte
	Records []Record
}

// Parse validates magic and the size field, locates the record region, and
// walks the records (name + NUL + fixed 68-byte region) until EOF.
func Parse(data []byte) (*File, error) {
	if len(data) < 12 {
		return nil, ErrTruncated
	}
	if !bytes.Equal(data[0:4], magic) {
		return nil, ErrBadMagic
	}
	if int(binary.LittleEndian.Uint32(data[8:12])) != len(data)-12 {
		return nil, ErrTruncated
	}

	idx := bytes.Index(data, chunkMagic)
	if idx < 0 {
		return nil, ErrTruncated
	}
	start := idx + len(chunkMagic) + 8

	f := &File{Raw: data}
	pos := start
	for pos < len(data) {
		nul := bytes.IndexByte(data[pos:], 0)
		if nul < 0 {
			return nil, ErrTruncated
		}
		name := string(data[pos : pos+nul])
		payloadStart := pos + nul + 1
		end := payloadStart + FixedRegionSize
		if end > len(data) {
			return nil, ErrTruncated
		}
		rec := Record{
			Name:         name,
			Offset:       pos,
			PayloadStart: payloadStart,
			Opaque:       data[pos:end],
		}
		classify(&rec, data[payloadStart:end])
		f.Records = append(f.Records, rec)
		pos = end
	}
	return f, nil
}

func classify(r *Record, fixed []byte) {
	switch binary.LittleEndian.Uint32(fixed[TypeOffset : TypeOffset+4]) {
	case 0:
		r.Kind = KindUnbound
	case 1:
		r.Kind = KindAxis
	case 2:
		r.Kind = KindButton
		r.Button = buttonFromMask(fixed[MaskWordsOffset : MaskWordsOffset+MaskWordCount*4])
		r.Device = &Device{
			InstanceGUID: formatGUID(fixed[GUIDOffset : GUIDOffset+16]),
			ProductGUID:  formatProductGUID(fixed[ProductOffset : ProductOffset+4]),
		}
	case 4:
		r.Kind = KindKey
		r.Key = fixed[KeyOffset]
	default:
		r.Kind = KindUnknown
	}
}

// buttonFromMask finds the lowest set button across the five mask words
// (buttons 0..159). Button N is bit N%32 of word N/32.
func buttonFromMask(words []byte) int {
	for w := 0; w < MaskWordCount; w++ {
		word := binary.LittleEndian.Uint32(words[w*4:])
		if word != 0 {
			return w*32 + bits.TrailingZeros32(word)
		}
	}
	return 0
}

// formatGUID renders a 16-byte Windows GUID (Data1/Data2/Data3 little-endian)
// as "{8-4-4-4-12}" uppercase hex.
func formatGUID(b []byte) string {
	d1 := binary.LittleEndian.Uint32(b[0:4])
	d2 := binary.LittleEndian.Uint16(b[4:6])
	d3 := binary.LittleEndian.Uint16(b[6:8])
	d4 := b[8:16]
	return fmt.Sprintf("{%08X-%04X-%04X-%02X%02X-%02X%02X%02X%02X%02X%02X}",
		d1, d2, d3, d4[0], d4[1], d4[2], d4[3], d4[4], d4[5], d4[6], d4[7])
}

// formatProductGUID renders a 4-byte ProductID as the derived ProductGUID
// string "{<pid:08X>-0000-0000-0000-504944564944}".
func formatProductGUID(pid []byte) string {
	return fmt.Sprintf("{%08X-0000-0000-0000-504944564944}", binary.LittleEndian.Uint32(pid))
}
