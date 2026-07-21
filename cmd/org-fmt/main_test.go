package main

import (
	"os"
	"path/filepath"
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

func TestRunNoArgs(t *testing.T) {
	if code := run(nil); code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}

func TestRunContinuesOnErrorAndReportsFailure(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.org")
	if err := os.WriteFile(good, []byte("#+title: x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "nope.org")

	if code := run([]string{missing, good}); code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	// the good file was still formatted despite the missing one
	got, _ := os.ReadFile(good)
	if string(got) != "#+TITLE: x\n" {
		t.Fatalf("good file not formatted: %q", got)
	}
}
