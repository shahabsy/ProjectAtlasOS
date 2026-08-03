package core

import (
	"fmt"
	"os"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
)

// IndexCmd handles the `index` subcommand.
// It parses flags and delegates to the appropriate indexing function.
func IndexCmd(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: atlas index [node|scene] ...")
		os.Exit(1)
	}

	switch args[0] {
	case "node":
		IndexNodeCmd(args[1:])
	case "scene":
		IndexSceneCmd(args[1:])
	default:
		fmt.Printf("Unknown index subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

// IndexSceneCmd handles `index scene --full --file <path>`.
func IndexSceneCmd(args []string) {
	var full bool
	var filePath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--full":
			full = true
		case "--file":
			if i+1 < len(args) {
				filePath = args[i+1]
				i++
			}
		}
	}

	if !full || filePath == "" {
		fmt.Println("Usage: atlas index scene --full --file <path>")
		os.Exit(1)
	}

	// Read JSON file
	jsonData, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
		os.Exit(1)
	}

	// Delegate to the core indexing function
	if err := IndexFullScene(jsonData); err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
}

// IndexNodeCmd handles `index node --type=... --guid=... --name=...`.
func IndexNodeCmd(args []string) {
	var typ, guid, name string
	for i := 0; i < len(args); i++ {
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
		fmt.Println("Missing required arguments: --type, --guid, --name")
		os.Exit(1)
	}
	if err := IndexNode(typ, guid, name); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		os.Exit(1)
	}
}

// IndexNode inserts a single node into the graph.
// This is the entry point for `index node`.
func IndexNode(typ, guid, name string) error {
	projectRoot, err := findUnityProject()
	if err != nil {
		return fmt.Errorf("not a Unity project: %w", err)
	}
	dbPath := db.GetDBPath(projectRoot)

	// Check if node already exists by GUID
	node, err := db.GetNodeByGUID(dbPath, guid)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	var id string
	if node == nil {
		// Generate a new ID
		id = typ + "_" + guid[:8]
	} else {
		id = node.ID
	}

	if err := db.InsertNode(dbPath, id, typ, guid, "", name, ""); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}
	fmt.Printf("✅ Indexed %s '%s' (guid: %s)\n", typ, name, guid)
	return nil
}
