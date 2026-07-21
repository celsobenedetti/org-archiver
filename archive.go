package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/niklasfasching/go-org/org"
)

// doneStates are the TODO keywords whose headlines get archived.
var doneStates = map[string]bool{
	"DONE":      true,
	"CANCELLED": true,
}

func isDone(status string) bool { return doneStates[status] }

// forceDoneStatus reclassifies headlines whose leading title text is one of our
// done keywords but that go-org left unrecognized because the file's own
// #+TODO: line doesn't declare it. This keeps {DONE, CANCELLED} archivable
// regardless of a file's TODO configuration (the "hardcoded set" requirement).
func forceDoneStatus(nodes []org.Node) {
	for i, n := range nodes {
		h, ok := n.(org.Headline)
		if !ok {
			continue
		}
		if h.Status == "" && len(h.Title) > 0 {
			if t, ok := h.Title[0].(org.Text); ok {
				for kw := range doneStates {
					if strings.HasPrefix(t.Content, kw+" ") {
						h.Status = kw
						t.Content = strings.TrimPrefix(t.Content, kw+" ")
						h.Title[0] = t
						break
					}
				}
			}
		}
		forceDoneStatus(h.Children)
		nodes[i] = h
	}
}

// partition splits nodes into the tree to keep (done subtrees removed) and the
// done headlines to archive, in document (pre-order) order. When a headline is
// done its whole subtree is archived and not descended into.
func partition(nodes []org.Node) (kept []org.Node, archived []org.Headline) {
	for _, n := range nodes {
		h, ok := n.(org.Headline)
		if !ok {
			kept = append(kept, n)
			continue
		}
		if isDone(h.Status) {
			archived = append(archived, h)
			continue
		}
		keptChildren, childArchived := partition(h.Children)
		h.Children = keptChildren
		kept = append(kept, h)
		archived = append(archived, childArchived...)
	}
	return kept, archived
}

// shiftHeadline re-levels a headline and all its descendant headlines by delta.
func shiftHeadline(h org.Headline, delta int) org.Headline {
	h.Lvl += delta
	children := make([]org.Node, len(h.Children))
	for i, c := range h.Children {
		if ch, ok := c.(org.Headline); ok {
			children[i] = shiftHeadline(ch, delta)
		} else {
			children[i] = c
		}
	}
	h.Children = children
	return h
}

// buildArchiveSection wraps the archived headlines in a level-1
// "archive <timestamp>" heading, re-leveling each subtree root to level 2.
func buildArchiveSection(now time.Time, archived []org.Headline) org.Headline {
	children := make([]org.Node, len(archived))
	for i, h := range archived {
		children[i] = shiftHeadline(h, 2-h.Lvl)
	}
	return org.Headline{
		Lvl:      1,
		Title:    []org.Node{org.Text{Content: "archive " + now.Format("2006-01-02 15:04")}},
		Children: children,
	}
}

// render pretty-prints nodes as an org-mode string.
func render(nodes []org.Node) string {
	w := org.NewOrgWriter()
	org.WriteNodes(w, nodes...)
	return w.String()
}

// parseOrg parses org source, returning its top-level nodes. Our done states
// are registered as TODO keywords so go-org populates Headline.Status for them
// (files with their own #+TODO: line still override this default).
func parseOrg(source, path string) ([]org.Node, error) {
	c := org.New().Silent()
	c.DefaultSettings["TODO"] = "TODO | DONE CANCELLED"
	d := c.Parse(strings.NewReader(source), path)
	if d.Error != nil {
		return nil, d.Error
	}
	forceDoneStatus(d.Nodes)
	return d.Nodes, nil
}

// archivePath returns the .org_archive path for a source file.
func archivePath(source string) string {
	return source + "_archive"
}

// writeAtomic writes data to path via a temp file in the same directory then
// renames over the target, preserving mode when the file already exists.
func writeAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	mode := os.FileMode(0o644)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode()
	}
	tmp, err := os.CreateTemp(dir, ".org-archiver-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, mode); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

// processFile archives done items from a single source file into its
// .org_archive, rewriting the source. It returns the number of items archived.
// A file with nothing to archive is left untouched.
func processFile(path string, now time.Time) (int, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	nodes, err := parseOrg(string(source), path)
	if err != nil {
		return 0, err
	}

	kept, archived := partition(nodes)
	if len(archived) == 0 {
		return 0, nil
	}

	section := buildArchiveSection(now, archived)

	archPath := archivePath(path)
	var archNodes []org.Node
	if existing, err := os.ReadFile(archPath); err == nil {
		archNodes, err = parseOrg(string(existing), archPath)
		if err != nil {
			return 0, fmt.Errorf("parsing existing archive %s: %w", archPath, err)
		}
	} else if !os.IsNotExist(err) {
		return 0, err
	}
	archNodes = append(archNodes, section)

	if err := writeAtomic(path, []byte(render(kept))); err != nil {
		return 0, err
	}
	if err := writeAtomic(archPath, []byte(render(archNodes))); err != nil {
		return 0, err
	}
	return len(archived), nil
}
