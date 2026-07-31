package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCaptureCreatesInboxWithPlainNote(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	if err := capture(path, "Buy milk"); err != nil {
		t.Fatalf("capture: %v", err)
	}

	got, _ := os.ReadFile(path)
	want := "* Buy milk\n"
	if string(got) != want {
		t.Fatalf("content mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestCaptureMultilinePlainNoteBecomesBody(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	if err := capture(path, "Buy milk\nneed 2%\nand oat milk too"); err != nil {
		t.Fatalf("capture: %v", err)
	}

	got, _ := os.ReadFile(path)
	want := "* Buy milk\nneed 2%\nand oat milk too\n"
	if string(got) != want {
		t.Fatalf("content mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestCaptureAppendsToExistingInbox(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")
	if err := os.WriteFile(path, []byte("* existing\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := capture(path, "new note"); err != nil {
		t.Fatalf("capture: %v", err)
	}

	got, _ := os.ReadFile(path)
	want := "* existing\n* new note\n"
	if string(got) != want {
		t.Fatalf("content mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestCaptureUsesExplicitHeadlineAsIs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	note := "** already a headline\nwith a body line"
	if err := capture(path, note); err != nil {
		t.Fatalf("capture: %v", err)
	}

	got, _ := os.ReadFile(path)
	want := "** already a headline\nwith a body line\n"
	if string(got) != want {
		t.Fatalf("content mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestCaptureRejectsEmptyNote(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	if err := capture(path, "   \n\n"); err == nil {
		t.Fatal("expected error for empty note")
	}
}
