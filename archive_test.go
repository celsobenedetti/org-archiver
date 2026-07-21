package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/niklasfasching/go-org/org"
)

func mustParse(t *testing.T, source string) []org.Node {
	t.Helper()
	nodes, err := parseOrg(source, "test.org")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return nodes
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func headlineTitles(archived []org.Headline) []string {
	titles := make([]string, len(archived))
	for i, h := range archived {
		titles[i] = strings.TrimSpace(org.String(h.Title...))
	}
	return titles
}

func TestPartitionTopLevel(t *testing.T) {
	nodes := mustParse(t, `* TODO keep me
* DONE archive me
* CANCELLED also archive
`)
	kept, archived := partition(nodes)

	if got := headlineTitles(archived); len(got) != 2 || got[0] != "archive me" || got[1] != "also archive" {
		t.Fatalf("archived = %v, want [archive me, also archive]", got)
	}
	if len(kept) != 1 {
		t.Fatalf("kept = %d nodes, want 1", len(kept))
	}
}

func TestPartitionArchivesWholeSubtree(t *testing.T) {
	nodes := mustParse(t, `* DONE parent
** TODO child still open
*** DONE grandchild
`)
	kept, archived := partition(nodes)

	if len(kept) != 0 {
		t.Fatalf("kept = %d, want 0 (whole subtree archived)", len(kept))
	}
	if len(archived) != 1 {
		t.Fatalf("archived = %d, want 1 (only the parent, subtree not descended)", len(archived))
	}
	if len(archived[0].Children) != 1 {
		t.Fatalf("parent should retain its child subtree, got %d children", len(archived[0].Children))
	}
}

func TestPartitionDoneNestedUnderOpenParent(t *testing.T) {
	nodes := mustParse(t, `* TODO open parent
** DONE done child
** TODO open child
`)
	kept, archived := partition(nodes)

	if len(archived) != 1 || headlineTitles(archived)[0] != "done child" {
		t.Fatalf("archived = %v, want [done child]", headlineTitles(archived))
	}
	parent, ok := kept[0].(org.Headline)
	if !ok {
		t.Fatalf("kept[0] is not a headline")
	}
	if len(parent.Children) != 1 {
		t.Fatalf("open parent should keep its one open child, got %d", len(parent.Children))
	}
}

func TestShiftHeadlineRelevels(t *testing.T) {
	nodes := mustParse(t, `*** DONE deep
**** sub
`)
	h := nodes[0].(org.Headline)
	shifted := shiftHeadline(h, 2-h.Lvl)

	if shifted.Lvl != 2 {
		t.Fatalf("root Lvl = %d, want 2", shifted.Lvl)
	}
	child := shifted.Children[0].(org.Headline)
	if child.Lvl != 3 {
		t.Fatalf("child Lvl = %d, want 3", child.Lvl)
	}
}

func TestBuildArchiveSection(t *testing.T) {
	nodes := mustParse(t, `*** DONE deep item
**** sub item
`)
	_, archived := partition(nodes)
	now := time.Date(2026, 8, 8, 12, 24, 0, 0, time.UTC)
	section := buildArchiveSection(now, archived)

	out := render([]org.Node{section})
	want := `* archive 2026-08-08 12:24
** DONE deep item
*** sub item
`
	if out != want {
		t.Fatalf("render mismatch:\n--- got ---\n%s\n--- want ---\n%s", out, want)
	}
}

func TestProcessFileEndToEnd(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.org")
	writeFile(t, src, `* TODO keep
* DONE finish it
** notes under done
* CANCELLED drop it
`)

	now := time.Date(2026, 8, 8, 12, 24, 0, 0, time.UTC)
	n, err := processFile(src, now)
	if err != nil {
		t.Fatalf("processFile: %v", err)
	}
	if n != 2 {
		t.Fatalf("archived count = %d, want 2", n)
	}

	gotSrc, _ := os.ReadFile(src)
	if strings.Contains(string(gotSrc), "DONE") || strings.Contains(string(gotSrc), "CANCELLED") {
		t.Fatalf("source still has done items:\n%s", gotSrc)
	}
	if !strings.Contains(string(gotSrc), "keep") {
		t.Fatalf("source lost the kept item:\n%s", gotSrc)
	}

	gotArch, err := os.ReadFile(archivePath(src))
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	wantArch := `* archive 2026-08-08 12:24
** DONE finish it
*** notes under done
** CANCELLED drop it
`
	if string(gotArch) != wantArch {
		t.Fatalf("archive mismatch:\n--- got ---\n%s\n--- want ---\n%s", gotArch, wantArch)
	}
}

func TestProcessFileNoOp(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.org")
	original := `* TODO nothing done here
* TODO still open
`
	writeFile(t, src, original)

	n, err := processFile(src, time.Now())
	if err != nil {
		t.Fatalf("processFile: %v", err)
	}
	if n != 0 {
		t.Fatalf("archived count = %d, want 0", n)
	}

	got, _ := os.ReadFile(src)
	if string(got) != original {
		t.Fatalf("source was modified on no-op:\n%s", got)
	}
	if _, err := os.Stat(archivePath(src)); !os.IsNotExist(err) {
		t.Fatalf("archive file should not have been created")
	}
}

func TestProcessFileAppendsToExistingArchive(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.org")
	writeFile(t, src, "* DONE second run\n")
	writeFile(t, archivePath(src), "* archive 2026-01-01 09:00\n** DONE first run\n")

	now := time.Date(2026, 8, 8, 12, 24, 0, 0, time.UTC)
	if _, err := processFile(src, now); err != nil {
		t.Fatalf("processFile: %v", err)
	}

	got, _ := os.ReadFile(archivePath(src))
	want := `* archive 2026-01-01 09:00
** DONE first run
* archive 2026-08-08 12:24
** DONE second run
`
	if string(got) != want {
		t.Fatalf("archive mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

// CANCELLED must be archived even when the file declares its own #+TODO: line
// that doesn't list it — the done set is hardcoded, not derived from #+TODO.
func TestPartitionDoneRegardlessOfFileTodoConfig(t *testing.T) {
	nodes := mustParse(t, `#+TODO: TODO NEXT | DONE
* NEXT keep open
* CANCELLED scrapped
* DONE completed
`)
	_, archived := partition(nodes)

	got := headlineTitles(archived)
	if len(got) != 2 || got[0] != "scrapped" || got[1] != "completed" {
		t.Fatalf("archived = %v, want [scrapped, completed]", got)
	}
}

// A pre-existing PROPERTIES drawer on an archived item is carried through.
func TestPropertiesDrawerPreserved(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "main.org")
	writeFile(t, src, `* DONE with props
:PROPERTIES:
:ID: abc-123
:END:
`)
	now := time.Date(2026, 8, 8, 12, 24, 0, 0, time.UTC)
	if _, err := processFile(src, now); err != nil {
		t.Fatalf("processFile: %v", err)
	}

	got, _ := os.ReadFile(archivePath(src))
	if !strings.Contains(string(got), ":ID: abc-123") {
		t.Fatalf("PROPERTIES drawer lost:\n%s", got)
	}
}
