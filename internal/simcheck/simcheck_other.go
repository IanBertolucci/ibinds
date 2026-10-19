//go:build !windows

package simcheck

func running() (bool, error) {
	return false, nil
}
