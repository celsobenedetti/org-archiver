package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/celsobenedetti/org-archiver/internal/orgfile"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout))
}

func run(files []string, stdin io.Reader, stdout io.Writer) int {
	// No files: act as a filter, formatting stdin to stdout.
	if len(files) == 0 {
		source, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		out, err := format(string(source), "<stdin>")
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return 1
		}
		io.WriteString(stdout, out)
		return 0
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

// format standardizes org source: go-org's round-trip plus collapsing runs of
// blank lines.
func format(source, path string) (string, error) {
	nodes, err := orgfile.Parse(source, path)
	if err != nil {
		return "", err
	}
	return collapseBlankLines(orgfile.Render(nodes)), nil
}

// formatFile rewrites path with standardized formatting in place, reporting
// whether the file changed. An already-formatted file is left untouched.
func formatFile(path string) (bool, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	out, err := format(string(source), path)
	if err != nil {
		return false, err
	}
	if out == string(source) {
		return false, nil
	}
	if err := orgfile.WriteAtomic(path, []byte(out)); err != nil {
		return false, err
	}
	return true, nil
}

// collapseBlankLines collapses runs of consecutive blank lines into a single
// blank line. Lines inside #+BEGIN_…/#+END_… blocks are left verbatim, since
// blank lines there (e.g. in source blocks) are significant.
func collapseBlankLines(s string) string {
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	inBlock, prevBlank := false, false
	for _, line := range lines {
		delim := strings.ToUpper(strings.TrimSpace(line))
		if strings.HasPrefix(delim, "#+BEGIN_") {
			inBlock = true
		}
		blank := !inBlock && strings.TrimSpace(line) == ""
		if blank && prevBlank {
			continue
		}
		out = append(out, line)
		prevBlank = blank
		if strings.HasPrefix(delim, "#+END_") {
			inBlock = false
		}
	}
	return strings.Join(out, "\n")
}
