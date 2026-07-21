package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunNoArgs(t *testing.T) {
	if code := run(nil, time.Now()); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

func TestRunContinuesOnErrorAndReportsFailure(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.org")
	if err := os.WriteFile(good, []byte("* DONE done\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "nope.org")

	now := time.Date(2026, 8, 8, 12, 24, 0, 0, time.UTC)
	code := run([]string{missing, good}, now)

	if code != 1 {
		t.Fatalf("exit = %d, want 1 (a file failed)", code)
	}
	// the good file was still processed despite the missing one
	if _, err := os.Stat(archivePath(good)); err != nil {
		t.Fatalf("good file was not processed: %v", err)
	}
}
