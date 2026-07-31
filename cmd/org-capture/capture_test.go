package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaptureCreatesInboxWithPlainNote(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	if err := capture(path, "Buy milk", nil); err != nil {
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

	if err := capture(path, "Buy milk\nneed 2%\nand oat milk too", nil); err != nil {
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

	if err := capture(path, "new note", nil); err != nil {
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
	if err := capture(path, note, nil); err != nil {
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

	if err := capture(path, "   \n\n", nil); err == nil {
		t.Fatal("expected error for empty note")
	}
}

func TestCaptureAppendsTagsToHeading(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	if err := capture(path, "Buy milk", []string{"home", "errand"}); err != nil {
		t.Fatalf("capture: %v", err)
	}

	got, _ := os.ReadFile(path)
	line := strings.TrimSuffix(string(got), "\n")
	if !strings.HasPrefix(line, "* Buy milk") || !strings.HasSuffix(line, ":home:errand:") {
		t.Fatalf("expected heading with trailing tags, got %q", got)
	}
}

func TestCaptureAppendsTagsOnlyToHeadingLineNotBody(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	if err := capture(path, "Buy milk\nneed 2%", []string{"home"}); err != nil {
		t.Fatalf("capture: %v", err)
	}

	got, _ := os.ReadFile(path)
	lines := strings.SplitN(strings.TrimSuffix(string(got), "\n"), "\n", 2)
	if len(lines) != 2 {
		t.Fatalf("expected heading + body line, got %q", got)
	}
	if !strings.HasPrefix(lines[0], "* Buy milk") || !strings.HasSuffix(lines[0], ":home:") {
		t.Fatalf("expected heading with trailing tag, got %q", lines[0])
	}
	if lines[1] != "need 2%" {
		t.Fatalf("expected body untouched, got %q", lines[1])
	}
}
