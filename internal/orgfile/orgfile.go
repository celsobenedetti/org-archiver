// Package orgfile provides generic org-mode file primitives shared by the
// commands: plain parsing, pretty-printing via go-org's OrgWriter, and atomic
// file replacement.
package orgfile

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/niklasfasching/go-org/org"
)

// Parse parses org source with go-org's default settings and returns its
// top-level nodes.
func Parse(source, path string) ([]org.Node, error) {
	d := org.New().Silent().Parse(strings.NewReader(source), path)
	if d.Error != nil {
		return nil, d.Error
	}
	return d.Nodes, nil
}

// TagsColumn is the column headline tags are right-aligned to (go-org defaults
// to 77).
const TagsColumn = 80

// Render pretty-prints nodes as a standardized org-mode string using go-org's
// OrgWriter defaults, except tags are aligned to TagsColumn.
func Render(nodes []org.Node) string {
	w := org.NewOrgWriter()
	w.TagsColumn = TagsColumn
	org.WriteNodes(w, nodes...)
	return w.String()
}

// WriteAtomic writes data to path via a temp file in the same directory then
// renames over the target, preserving mode when the file already exists.
func WriteAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	mode := os.FileMode(0o644)
	if fi, err := os.Stat(path); err == nil {
		mode = fi.Mode()
	}
	tmp, err := os.CreateTemp(dir, ".orgfile-*")
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
