//go:build windows

package config

import "golang.org/x/sys/windows"

// DocumentsDir resolves the user's Documents folder via the Windows Known
// Folder API, which handles OneDrive-redirected Documents correctly.
func DocumentsDir() (string, error) {
	return windows.KnownFolderPath(windows.FOLDERID_Documents, windows.KF_FLAG_DEFAULT)
}
