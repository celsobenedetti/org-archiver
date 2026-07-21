package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatFileStandardizes(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "f.org")
	if err := os.WriteFile(src, []byte("#+title: hi\n* head\n:properties:\n:id: x\n:end:\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := formatFile(src)
	if err != nil {
		t.Fatalf("formatFile: %v", err)
	}
	if !changed {
		t.Fatal("expected changed = true")
	}

	got, _ := os.ReadFile(src)
	want := "#+TITLE: hi\n* head\n:PROPERTIES:\n:ID: x\n:END:\n"
	if string(got) != want {
		t.Fatalf("content mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestFormatFileNoOpLeavesUntouched(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "f.org")
	clean := "#+TITLE: hi\n* head\n"
	if err := os.WriteFile(src, []byte(clean), 0o644); err != nil {
		t.Fatal(err)
	}

	changed, err := formatFile(src)
	if err != nil {
		t.Fatalf("formatFile: %v", err)
	}
	if changed {
		t.Fatal("expected changed = false for already-formatted file")
	}
	got, _ := os.ReadFile(src)
	if string(got) != clean {
		t.Fatalf("no-op modified file:\n%s", got)
	}
}

func TestFormatCollapsesBlankLines(t *testing.T) {
	got, err := format("* one\n\n\n\ntext\n\n\n* two\n", "t.org")
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	want := "* one\n\ntext\n\n* two\n"
	if got != want {
		t.Fatalf("collapse mismatch:\n--- got ---\n%q\n--- want ---\n%q", got, want)
	}
}

func TestFormatPreservesBlankLinesInBlocks(t *testing.T) {
	src := "#+BEGIN_SRC python\nx=1\n\n\n\ny=2\n#+END_SRC\n"
	got, err := format(src, "t.org")
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	if got != src {
		t.Fatalf("block blank lines were altered:\n--- got ---\n%q\n--- want ---\n%q", got, src)
	}
}

func TestRunStdinToStdout(t *testing.T) {
	var out bytes.Buffer
	code := run(nil, strings.NewReader("#+title: hi\n\n\n* h\n"), &out)
	if code != 0 {
		t.Fatalf("exit = %d, want 0", code)
	}
	want := "#+TITLE: hi\n\n* h\n"
	if out.String() != want {
		t.Fatalf("stdout mismatch:\n--- got ---\n%q\n--- want ---\n%q", out.String(), want)
	}
}

func TestRunContinuesOnErrorAndReportsFailure(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.org")
	if err := os.WriteFile(good, []byte("#+title: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "nope.org")

	if code := run([]string{missing, good}, nil, nil); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	// the good file was still formatted despite the missing one
	got, _ := os.ReadFile(good)
	if string(got) != "#+TITLE: x\n" {
		t.Fatalf("good file not formatted: %q", got)
	}
}
