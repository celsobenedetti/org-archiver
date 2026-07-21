package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/celsobenedetti/org-archiver/internal/orgfile"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(files []string) int {
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s <file.org> [file.org ...]\n", filepath.Base(os.Args[0]))
		return 2
	}

	exit := 0
	for _, f := range files {
		changed, err := formatFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", f, err)
			exit = 1
			continue
		}
		if changed {
			fmt.Printf("%s: formatted\n", f)
		}
	}
	return exit
}

// formatFile rewrites path with go-org's standardized formatting, reporting
// whether the file changed. An already-formatted file is left untouched.
func formatFile(path string) (bool, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	nodes, err := orgfile.Parse(string(source), path)
	if err != nil {
		return false, err
	}
	out := orgfile.Render(nodes)
	if out == string(source) {
		return false, nil
	}
	if err := orgfile.WriteAtomic(path, []byte(out)); err != nil {
		return false, err
	}
	return true, nil
}
