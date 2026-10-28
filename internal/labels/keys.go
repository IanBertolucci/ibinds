package labels

import "fmt"

// iRacing stores keyboard keys in its own Key enum, not ASCII or Windows
// virtual-key codes. Printable letters/digits mostly match ASCII, but several
// keys do not. These values come from real controls.cfg files and iRacing's
// documented default shortcuts; see docs/FORMAT.md.
var keyLabels = map[byte]string{
	0x08: "Backspace",
	0x09: "Tab",
	0x0c: "Numpad 5",
	0x0d: "Enter",
	0x13: "Pause",
	0x1b: "Esc",
	0x20: "Space",
	0x21: "Page Up",
	0x22: "Page Down",
	0x23: "End",
	0x24: "Home",
	0x25: "Left",
	0x26: "Up",
	0x27: "Right",
	0x28: "Down",
	0x2c: "Print Screen",
	0x2d: "Insert",
	0x2e: "Delete",
	0x6b: ",",
	0x6c: "-",
	0x6d: ".",
	0x70: ";",
	0x72: "=",
	0x76: "]",
	0x78: "[",
	0xb1: "Numpad 3",
	0xb2: "Numpad 1",
	0xb3: "Numpad 7",
	0xb4: "Numpad 4",
	0xb5: "Numpad 8",
	0xb6: "Numpad 6",
	0xb9: "Numpad .",
}

// Key returns a display label for an iRacing key code. Unknown non-printable
// codes are shown as hex so the UI never renders a replacement character.
func Key(code byte) string {
	if code >= 0xc5 && code <= 0xd0 {
		return fmt.Sprintf("F%d", int(code-0xc5)+1)
	}
	if s, ok := keyLabels[code]; ok {
		return s
	}
	if code >= 0x20 && code <= 0x7e {
		return string(rune(code))
	}
	return fmt.Sprintf("0x%02X", code)
}
