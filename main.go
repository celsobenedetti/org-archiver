package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:], time.Now()))
}

func run(files []string, now time.Time) int {
	if len(files) == 0 {
		fmt.Fprintf(os.Stderr, "usage: %s <file.org> [file.org ...]\n", filepath.Base(os.Args[0]))
		return 2
	}

	exit := 0
	for _, f := range files {
		n, err := processFile(f, now)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", f, err)
			exit = 1
			continue
		}
		if n > 0 {
			fmt.Printf("%s: archived %d item(s) → %s\n", f, n, archivePath(f))
		}
	}
	return exit
}
