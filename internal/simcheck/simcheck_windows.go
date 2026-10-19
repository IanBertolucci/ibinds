//go:build windows

package simcheck

import (
	"os/exec"
	"strings"
)

func running() (bool, error) {
	out, err := exec.Command("tasklist", "/FI", "IMAGENAME eq iRacingSim64DX11.exe", "/NH").Output()
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if IsSimProcessLine(line) {
			return true, nil
		}
	}
	return false, nil
}
