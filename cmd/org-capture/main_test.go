package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunNoInboxSet(t *testing.T) {
	var errBuf strings.Builder
	code := run([]string{"note"}, "", "", nil, &errBuf)
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
	code := run([]string{"from arg"}, path, "", nil, &errBuf)
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
	code := run(nil, path, "", strings.NewReader("from stdin\nsecond line"), &errBuf)
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
	code := run([]string{"note"}, dir, "", nil, &errBuf)
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if errBuf.Len() == 0 {
		t.Fatal("expected error message on stderr")
	}
}

func TestRunPipedEmptyStdinIsNotTreatedAsTerminal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	w.Close() // EOF immediately, simulating `: | org-capture`

	var errBuf strings.Builder
	code := run(nil, path, "", r, &errBuf)
	if code != 1 {
		t.Fatalf("exit = %d, want 1 (empty note from closed pipe), stderr = %q", code, errBuf.String())
	}
	if strings.Contains(errBuf.String(), "usage") {
		t.Fatalf("piped stdin should not trigger usage message: %q", errBuf.String())
	}
}

func TestIsTerminalFalseForNonFileReader(t *testing.T) {
	if isTerminal(strings.NewReader("x")) {
		t.Fatal("expected false for a non-*os.File reader")
	}
}

func TestPrintUsageMentionsInboxAndInputMethods(t *testing.T) {
	var buf strings.Builder
	printUsage(&buf)
	if !strings.Contains(buf.String(), "usage:") {
		t.Fatalf("expected usage message, got %q", buf.String())
	}
	if !strings.Contains(buf.String(), "ORG_INBOX") {
		t.Fatalf("expected usage message to mention ORG_INBOX, got %q", buf.String())
	}
}

func TestRunTagsFlagAppendsToHeading(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	var errBuf strings.Builder
	code := run([]string{"--tags", "work, urgent", "call the client"}, path, "", nil, &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0, stderr = %q", code, errBuf.String())
	}

	got, _ := os.ReadFile(path)
	if !strings.Contains(string(got), ":work:urgent:") {
		t.Fatalf("expected tags in heading, got %q", got)
	}
}

func TestRunPipedStdinTakesPriorityOverEditFlag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "inbox.org")

	var errBuf strings.Builder
	code := run([]string{"-e"}, path, "", strings.NewReader("piped note"), &errBuf)
	if code != 0 {
		t.Fatalf("exit = %d, want 0, stderr = %q", code, errBuf.String())
	}

	got, _ := os.ReadFile(path)
	if string(got) != "* piped note\n" {
		t.Fatalf("expected piped stdin to win over -e, got %q", got)
	}
}

func TestParseTagsSplitsTrimsAndDropsEmpty(t *testing.T) {
	got := parseTags(" work ,, urgent ,")
	want := []string{"work", "urgent"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestParseTagsEmptyReturnsNil(t *testing.T) {
	if got := parseTags(""); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestOpenEditorRunsEditorAndReadsResult(t *testing.T) {
	dir := t.TempDir()
	editorScript := filepath.Join(dir, "fake-editor.sh")
	script := "#!/bin/sh\nprintf 'edited note\\nsecond line\\n' > \"$1\"\n"
	if err := os.WriteFile(editorScript, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := openEditor(editorScript)
	if err != nil {
		t.Fatalf("openEditor: %v", err)
	}
	want := "edited note\nsecond line\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestOpenEditorEmptyEditorErrors(t *testing.T) {
	if _, err := openEditor(""); err == nil {
		t.Fatal("expected error when $EDITOR is unset")
	}
}
