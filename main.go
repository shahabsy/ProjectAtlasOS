package main

import (
	"fmt"
	"os"

	"github.com/shahabsy/ProjectAtlasOS/internal/core"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "init" {
		fmt.Println("Usage: atlas init [--dry-run]")
		os.Exit(0)
	}

	dryRun := false
	for _, arg := range os.Args {
		if arg == "--dry-run" {
			dryRun = true
			break
		}
	}

	if err := core.Init(dryRun); err != nil {
		fmt.Fprint(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
