package fileops

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestBackupCreatesTimestampedCopy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "controls.cfg")
	content := []byte("hello world")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	backupPath, err := Backup(path)
	if err != nil {
		t.Fatalf("Backup: %v", err)
	}
	if !regexp.MustCompile(`\.bak-\d{8}-\d{6}$`).MatchString(backupPath) {
		t.Fatalf("backup name %q does not match pattern", backupPath)
	}
	got, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("backup content mismatch: %q", got)
	}
}

func TestBackupOverwritesNothing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "controls.cfg")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	b1, err := Backup(path)
	if err != nil {
		t.Fatalf("Backup 1: %v", err)
	}
	b2, err := Backup(path)
	if err != nil {
		t.Fatalf("Backup 2: %v", err)
	}
	if b1 == b2 {
		t.Fatalf("two backups share the same path %q", b1)
	}
}
