package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Getenv("ORG_INBOX"), os.Getenv("EDITOR"), os.Stdin, os.Stderr))
}

func run(args []string, inbox, editor string, stdin io.Reader, stderr io.Writer) int {
	fs := flag.NewFlagSet("org-capture", flag.ContinueOnError)
	fs.SetOutput(stderr)
	tagsFlag := fs.String("tags", "", "comma-separated tags to attach to the headline, e.g. work,urgent")
	editFlag := fs.Bool("e", false, "open $EDITOR to compose the note when none is given as an argument")
	fs.Usage = func() { printUsage(stderr) }
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if inbox == "" {
		fmt.Fprintln(stderr, "org-capture: $ORG_INBOX is not set")
		return 2
	}

	positional := fs.Args()

	var note string
	switch {
	case len(positional) > 0:
		note = positional[0]
	case !isTerminal(stdin):
		b, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "org-capture: %v\n", err)
			return 1
		}
		note = string(b)
	case *editFlag:
		n, err := openEditor(editor)
		if err != nil {
			fmt.Fprintf(stderr, "org-capture: %v\n", err)
			return 1
		}
		note = n
	default:
		printUsage(stderr)
		return 2
	}

	if err := capture(inbox, note, parseTags(*tagsFlag)); err != nil {
		fmt.Fprintf(stderr, "org-capture: %v\n", err)
		return 1
	}
	return 0
}

// isTerminal reports whether stdin is an interactive terminal rather than a
// pipe or redirected file, so a no-args, no-input invocation can fall back to
// -e or usage instead of hanging on io.ReadAll waiting for input that will
// never come.
func isTerminal(stdin io.Reader) bool {
	f, ok := stdin.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// openEditor launches $EDITOR on a scratch file and returns its contents once
// the editor exits. It talks to the real terminal (os.Stdin/Stdout/Stderr)
// rather than run's injected stdin/stderr, since an interactive editor needs
// the controlling terminal, not a redirected stream.
func openEditor(editor string) (string, error) {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return "", fmt.Errorf("$EDITOR is not set")
	}

	tmp, err := os.CreateTemp("", "org-capture-*.org")
	if err != nil {
		return "", err
	}
	path := tmp.Name()
	tmp.Close()
	defer os.Remove(path)

	cmd := exec.Command(parts[0], append(parts[1:], path)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running $EDITOR: %w", err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// parseTags splits a comma-separated tag list, trimming whitespace and
// dropping empty entries.
func parseTags(csv string) []string {
	if csv == "" {
		return nil
	}
	var tags []string
	for _, p := range strings.Split(csv, ",") {
		if p = strings.TrimSpace(p); p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

// printUsage prints a usage message describing how a note is supplied.
func printUsage(w io.Writer) {
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(w, "usage: %s [--tags csv] [-e] <note> | %s [--tags csv] (note read from stdin)\n", name, name)
	fmt.Fprintf(w, "  captures note to $ORG_INBOX as a new headline; note's first line\n")
	fmt.Fprintf(w, "  becomes the heading, remaining lines become its body\n")
	fmt.Fprintf(w, "  --tags csv   comma-separated tags appended to the heading, e.g. --tags work,urgent\n")
	fmt.Fprintf(w, "  -e           open $EDITOR to compose the note when none is given and stdin isn't piped\n")
}
