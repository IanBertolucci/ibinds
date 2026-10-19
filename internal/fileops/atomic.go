package fileops

import (
	"os"
	"path/filepath"
)

// AtomicWrite writes data to path atomically: it writes a temp file in the
// same directory, preserves the existing file's mode, then renames it over
// path. A crash mid-write can never leave a truncated file.
func AtomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := func() { _ = os.Remove(tmpName) }

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		cleanup()
		return err
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return err
	}
	if info, err := os.Stat(path); err == nil {
		if err := os.Chmod(tmpName, info.Mode()); err != nil {
			cleanup()
			return err
		}
	}
	if err := os.Rename(tmpName, path); err != nil {
		cleanup()
		return err
	}
	return nil
}
