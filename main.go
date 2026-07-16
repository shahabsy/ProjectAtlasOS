package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/core"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]
	switch command {
	case "init":
		dryRun := false
		for _, arg := range os.Args {
			if arg == "--dry-run" {
				dryRun = true
				break
			}
		}
		core.Init(dryRun)
	case "index":
		handleIndexCmd(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleIndexCmd(args []string) {
	if len(args) < 1 || args[0] != "node" {
		fmt.Printf("Use `atlas index node --type=<type> --guid=<guid> --name=<name>`")
		return
	}

	typ, guid, name := "", "", ""
	for i := 1; i < len(args); i++ {
		switch {
		case strings.HasPrefix(args[i], "--type="):
			typ = strings.TrimPrefix(args[i], "--type=")
		case strings.HasPrefix(args[i], "--guid="):
			guid = strings.TrimPrefix(args[i], "--guid=")
		case strings.HasPrefix(args[i], "--name="):
			name = strings.TrimPrefix(args[i], "--name=")
		}
	}

	if typ == "" || guid == "" || name == "" {
		fmt.Printf("Missing required arguments: --type, --guid, --name")
		return
	}

	if err := core.IndexNodeCmd(typ, guid, name); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: atlas [init|index]")
	fmt.Println("\nCommands:")
	fmt.Println("	init		Install Atlas Kernel")
	fmt.Println("	index		Index nodes into graph.db")
	fmt.Println("	node		Add a scene/prefab/shader node (requires --type, --guid, --name)")
}
