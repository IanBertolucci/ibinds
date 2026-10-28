package main

import "testing"

func TestVersionString(t *testing.T) {
	oldVersion, oldCommit, oldBuildDate := version, commit, buildDate
	t.Cleanup(func() {
		version, commit, buildDate = oldVersion, oldCommit, oldBuildDate
	})

	version = "1.2.3"
	commit = "abc1234"
	buildDate = "2026-10-27"

	want := "ibinds 1.2.3 (abc1234, 2026-10-27)"
	if got := versionString(); got != want {
		t.Fatalf("versionString() = %q, want %q", got, want)
	}
}
