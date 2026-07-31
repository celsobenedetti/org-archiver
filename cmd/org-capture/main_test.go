package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNoInboxSet(t *testing.T) {
	var errBuf strings.Builder
	code := run([]string{"note"}, "", nil, &errBuf)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errBuf.String(), "ORG_INBOX") {
		t.Fatalf("expected error mentioning ORG_INBOX, got %q", errBuf.String())
	}
}

func TestRunNoteFromArg(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	var errBuf strings.Builder
	code := run([]string{"from arg"}, path, nil, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0, stderr = %q", code, errBuf.String())
	}

	got, _ := os.ReadFile(path)
	if string(got) != "* from arg\n" {
		t.Fatalf("content mismatch: %q", got)
	}
}

func TestRunNoteFromStdin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	var errBuf strings.Builder
	code := run(nil, path, strings.NewReader("from stdin\nsecond line"), &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0, stderr = %q", code, errBuf.String())
	}

	got, _ := os.ReadFile(path)
	want := "* from stdin\nsecond line\n"
	if string(got) != want {
		t.Fatalf("content mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestRunCaptureErrorReported(t *testing.T) {
	dir := t.TempDir()
	// A directory path can't be read/written as a file, forcing capture to fail.
	var errBuf strings.Builder
	code := run([]string{"note"}, dir, nil, &errBuf)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if errBuf.Len() == 0 {
		t.Fatal("expected error message on stderr")
	}
}
