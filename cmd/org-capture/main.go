package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:], os.Getenv("ORG_INBOX"), os.Stdin, os.Stderr))
}

func run(args []string, inbox string, stdin io.Reader, stderr io.Writer) int {
	if inbox == "" {
		fmt.Fprintln(stderr, "org-capture: $ORG_INBOX is not set")
		return 2
	}

	var note string
	if len(args) > 0 {
		note = args[0]
	} else {
		b, err := io.ReadAll(stdin)
		if err != nil {
			fmt.Fprintf(stderr, "org-capture: %v\n", err)
			return 1
		}
		note = string(b)
	}

	if err := capture(inbox, note); err != nil {
		fmt.Fprintf(stderr, "org-capture: %v\n", err)
		return 1
	}
	return 0
}
