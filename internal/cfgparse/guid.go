package cfgparse

import (
	"encoding/hex"
	"fmt"
)

// ParseGUID converts a "{8-4-4-4-12}" GUID string into its 16-byte binary
// form (Data1/Data2/Data3 little-endian, Data4 as-is).
func ParseGUID(s string) ([]byte, error) {
	h, err := guidHex(s)
	if err != nil {
		return nil, err
	}
	raw, err := hex.DecodeString(h)
	if err != nil || len(raw) != 16 {
		return nil, fmt.Errorf("cfgparse: invalid GUID %q", s)
	}
	out := make([]byte, 16)
	reverse(out[0:4], raw[0:4])
	reverse(out[4:6], raw[4:6])
	reverse(out[6:8], raw[6:8])
	copy(out[8:16], raw[8:16])
	return out, nil
}

// ParseProductGUID extracts the 4-byte ProductID from a ProductGUID string.
func ParseProductGUID(s string) ([]byte, error) {
	h, err := guidHex(s)
	if err != nil {
		return nil, err
	}
	raw, err := hex.DecodeString(h)
	if err != nil || len(raw) != 16 {
		return nil, fmt.Errorf("cfgparse: invalid GUID %q", s)
	}
	out := make([]byte, 4)
	reverse(out, raw[0:4])
	return out, nil
}

func guidHex(s string) (string, error) {
	b := make([]byte, 0, 32)
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '{', '}', '-':
			continue
		}
		b = append(b, s[i])
	}
	return string(b), nil
}

func reverse(dst, src []byte) {
	for i := range src {
		dst[i] = src[len(src)-1-i]
	}
}
