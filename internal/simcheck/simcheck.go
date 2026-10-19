package simcheck

import (
	"errors"
	"strings"
)

// ErrSimRunning is returned by callers when the simulator must be closed
// before a write.
var ErrSimRunning = errors.New("iRacing is running; close it before writing")

const processName = "iRacingSim64DX11.exe"

// Running reports whether the iRacing simulator process is running.
func Running() (bool, error) {
	return running()
}

// IsSimProcessLine reports whether a tasklist output line refers to the
// iRacing simulator process (case-insensitive).
func IsSimProcessLine(line string) bool {
	return strings.Contains(strings.ToLower(line), strings.ToLower(processName))
}
