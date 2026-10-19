package simcheck

import "testing"

func TestIsSimProcessLineMatches(t *testing.T) {
	line := "iRacingSim64DX11.exe  1234 Console  1  1,234 K"
	if !IsSimProcessLine(line) {
		t.Fatalf("IsSimProcessLine(%q) = false, want true", line)
	}
}

func TestIsSimProcessLineNoMatch(t *testing.T) {
	for _, line := range []string{
		"iRacingUI.exe  1234 Console  1  1,234 K",
		"notepad.exe  5678 Console  1  1,234 K",
	} {
		if IsSimProcessLine(line) {
			t.Fatalf("IsSimProcessLine(%q) = true, want false", line)
		}
	}
}

func TestIsSimProcessLineCaseInsensitive(t *testing.T) {
	if !IsSimProcessLine("iracingsim64dx11.exe  1234 Console  1  1,234 K") {
		t.Fatal("case-insensitive match failed")
	}
}
