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
		for _, arg := range os.Args[2:] {
			if arg == "--dry-run" {
				dryRun = true
				break
			}
		}
		if err := core.Init(dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "index":
		handleIndexCmd(os.Args[2:])
	case "verify":
		if err := core.VerifyCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "query":
		if err := core.QueryCmd(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "analyze":
		handleAnalyzeCmd(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleIndexCmd(args []string) {
	if len(args) == 0 {
		fmt.Printf("Use atlas index [node|scene] ...")
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "node":
		handleIndexNode(args[1:])
	case "scene":
		handleIndexScene(args[1:])
	default:
		fmt.Printf("Unknow index subcommand: %s\n", subCmd)
		os.Exit(1)
	}
}

func handleIndexNode(args []string) {
	typ, guid, name := "", "", ""
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
	core.IndexNodeCmd([]string{"--type=" + typ, "--guid=" + guid, "--name=" + name})
}
func handleIndexScene(args []string) {
	var full bool
	var jsonPayload string
	var filePath string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--full":
			full = true
		case "--json":
			if i+1 < len(args) {
				jsonPayload = args[i+1]
				i++
			}
		case "--file":
			if i+1 < len(args) {
				filePath = args[i+1]
				i++
			}
		}
	}
	if !full {
		fmt.Println("Usage: atlas index scene --full [--json '...' | -- file data.json]")
		os.Exit(1)
	}

	var sceneData []byte
	var err error
	if filePath != "" {
		sceneData, err = os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read file: %v\n", err)
			os.Exit(1)
		}
	} else if jsonPayload != "" {
		sceneData = []byte(jsonPayload)
	} else {
		fmt.Println("Missing --json or --file")
		os.Exit(1)
	}

	if err := core.IndexFullScene(sceneData); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func handleAnalyzeCmd(args []string) {
	if len(args) == 0 {
		fmt.Println("Use `atlas analyze lighting`")
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "lighting":
		if err := core.AnalyzeLightingCmd(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown analyze command: %s. Use 'atlas analyze lighting'\n", subCmd)
	}
}

func printUsage() {
	fmt.Println("Atlas OS – AI Game Development Operating System")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  atlas init [--dry-run]          Install Atlas Kernel in Unity project")
	fmt.Println("  atlas index node ...            Index a single node (scene, prefab, shader)")
	fmt.Println("  atlas index scene --full ...    Index full scene hierarchy from JSON file")
	fmt.Println("  atlas verify                    Show indexed nodes in database")
	fmt.Println("  atlas query scenes              List all indexed scenes")
	fmt.Println("  atlas query gameobjects ...     List GameObjects in a scene")
	fmt.Println("  atlas query components ...      List components on a GameObject")
	fmt.Println("  atlas analyze lighting          Analyze scene lighting setup")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  atlas init --dry-run")
	fmt.Println("  atlas index node --type=scene --guid=d20ebab... --name=\"MainMenu\"")
	fmt.Println("  atlas index scene --full --file scene_data.json")
	fmt.Println("  atlas query scenes")
	fmt.Println("  atlas query gameobjects --scene=d20ebab...")
	fmt.Println("  atlas analyze lighting")
}
