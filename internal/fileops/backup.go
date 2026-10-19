package fileops

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Backup copies path to "<path>.bak-YYYYMMDD-HHMMSS" in the same directory.
// On a timestamp collision it advances by one second and retries.
func Backup(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	base := time.Now()
	for i := 0; i < 1000; i++ {
		ts := base.Add(time.Duration(i) * time.Second).Format("20060102-150405")
		backupPath := fmt.Sprintf("%s.bak-%s", path, ts)
		f, err := os.OpenFile(backupPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", err
		}
		if _, werr := f.Write(data); werr != nil {
			f.Close()
			return "", werr
		}
		if cerr := f.Close(); cerr != nil {
			return "", cerr
		}
		return backupPath, nil
	}
	return "", errors.New("fileops: could not create a unique backup")
}
