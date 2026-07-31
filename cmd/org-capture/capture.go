package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/celsobenedetti/org-archiver/internal/orgfile"
	"github.com/niklasfasching/go-org/org"
)

// headlineRe matches a line that is already an org headline marker.
var headlineRe = regexp.MustCompile(`^\*+\s`)

// capture appends note to the inbox file at path, creating the file if it
// doesn't exist yet. A note whose first line is already a headline (e.g.
// "* Buy milk") is used as-is, so a full subtree can be captured verbatim;
// otherwise the first line becomes a new level-1 headline's title and any
// remaining lines become its body.
func capture(path, note string) error {
	note = strings.TrimRight(note, "\n")
	if strings.TrimSpace(note) == "" {
		return fmt.Errorf("empty note")
	}
	if !headlineRe.MatchString(note) {
		note = "* " + note
	}
	note += "\n"

	newNodes, err := orgfile.Parse(note, "<capture>")
	if err != nil {
		return err
	}

	var nodes []org.Node
	if existing, err := os.ReadFile(path); err == nil {
		nodes, err = orgfile.Parse(string(existing), path)
		if err != nil {
			return fmt.Errorf("parsing existing inbox %s: %w", path, err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	nodes = append(nodes, newNodes...)
	return orgfile.WriteAtomic(path, []byte(orgfile.Render(nodes)))
}
