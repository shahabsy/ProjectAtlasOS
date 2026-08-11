package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shahabsy/ProjectAtlasOS/internal/core"
	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	"github.com/shahabsy/ProjectAtlasOS/internal/graph"
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
	case "graph":
		handleGraphCmd(os.Args[2:])
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
		fmt.Printf("Unknown index subcommand: %s\n", subCmd)
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
		fmt.Println("Usage: atlas index scene --full [--json '...' | --file data.json]")
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

// ─── Graph Commands ───────────────────────────────────────────────────────

func handleGraphCmd(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: atlas graph <subcommand>")
		fmt.Println("Subcommands: scenes, scene, gameobjects, scripts, prefabs, assets, stats")
		os.Exit(1)
	}

	// Find project root and database path
	projectRoot, err := findUnityProject()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Not a Unity project: %v\n", err)
		os.Exit(1)
	}
	dbPath := db.GetDBPath(projectRoot)

	// ✅ Correct: pass dbPath and handle error
	repo, err := graph.NewPrespository(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()

	switch args[0] {
	case "scenes":
		listScenes(repo)
	case "scene":
		if len(args) < 2 {
			fmt.Println("Usage: atlas graph scene <name>")
			os.Exit(1)
		}
		showScene(repo, args[1])
	case "gameobjects":
		if len(args) < 2 {
			fmt.Println("Usage: atlas graph gameobjects <scene-name>")
			os.Exit(1)
		}
		listGameObjects(repo, args[1])
	case "scripts":
		if len(args) < 2 {
			fmt.Println("Usage: atlas graph scripts <scene-name>")
			os.Exit(1)
		}
		listScripts(repo, args[1])
	case "prefabs":
		listPrefabs(repo)
	case "assets":
		listAssets(repo)
	case "stats":
		showProjectStats(repo)
	default:
		fmt.Printf("Unknown graph subcommand: %s\n", args[0])
		os.Exit(1)
	}
}

func listScenes(repo *graph.Repository) {
	queries := graph.NewSceneQueries(repo)
	list, err := queries.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📁 Found %d scenes:\n", len(list))
	for _, s := range list {
		fmt.Printf("  - %s (%s)\n", s.Name, s.GUID)
	}
}

func showScene(repo *graph.Repository, name string) {
	queries := graph.NewSceneQueries(repo)
	s, err := queries.ByName(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	stats, err := queries.Statistics(s.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📁 Scene: %s\n", s.Name)
	fmt.Printf("   GUID: %s\n", s.GUID)
	fmt.Printf("   Last indexed: %s\n", time.Unix(s.LastIndexed, 0).Format(time.RFC3339))
	fmt.Printf("   GameObjects: %d\n", stats.GameObjectCount)
	fmt.Printf("   Components: %d\n", stats.ComponentCount)
	fmt.Printf("   Scripts: %d\n", stats.ScriptCount)
	fmt.Printf("   Prefabs: %d\n", stats.PrefabCount)
	fmt.Printf("   Assets: %d\n", stats.AssetCount)
}

func listGameObjects(repo *graph.Repository, sceneName string) {
	sceneQueries := graph.NewSceneQueries(repo)
	s, err := sceneQueries.ByName(sceneName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Scene '%s' not found: %v\n", sceneName, err)
		os.Exit(1)
	}
	queries := graph.NewGameObjectQueries(repo)
	objects, err := queries.ListByScene(s.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📦 GameObjects in scene '%s':\n", sceneName)
	for _, gobj := range objects {
		fmt.Printf("  - %s (ID: %s)\n", gobj.Name, gobj.ID)
	}
}

func listScripts(repo *graph.Repository, sceneName string) {
	sceneQueries := graph.NewSceneQueries(repo)
	s, err := sceneQueries.ByName(sceneName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Scene '%s' not found: %v\n", sceneName, err)
		os.Exit(1)
	}
	queries := graph.NewScriptQueries(repo)
	scripts, err := queries.ListByScene(s.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📄 Scripts in scene '%s':\n", sceneName)
	for _, scr := range scripts {
		fmt.Printf("  - %s (%s)\n", scr.Name, scr.ClassName)
	}
}

func listPrefabs(repo *graph.Repository) {
	queries := graph.NewPrefabQueries(repo)
	prefabs, err := queries.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📦 Found %d prefabs:\n", len(prefabs))
	for _, p := range prefabs {
		fmt.Printf("  - %s (%s)\n", p.Name, p.GUID)
	}
}

func listAssets(repo *graph.Repository) {
	queries := graph.NewAssetQueries(repo)
	assets, err := queries.List("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("🎨 Found %d assets:\n", len(assets))
	for _, a := range assets {
		fmt.Printf("  - %s (%s) [%s]\n", a.Name, a.GUID, a.Type)
	}
}

func showProjectStats(repo *graph.Repository) {
	queries := graph.NewSceneQueries(repo)
	stats, err := queries.ProjectStats()
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("📊 Project Statistics:\n")
	fmt.Printf("  Scenes: %d\n", stats.SceneCount)
	fmt.Printf("  GameObjects: %d\n", stats.GameObjectCount)
	fmt.Printf("  Components: %d\n", stats.ComponentCount)
	fmt.Printf("  Scripts: %d\n", stats.ScriptCount)
	fmt.Printf("  Prefabs: %d\n", stats.PrefabCount)
	fmt.Printf("  Assets: %d\n", stats.AssetCount)
}

// ─── Helper: Find Unity Project Root ────────────────────────────────────

func findUnityProject() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for {
		assetsPath := filepath.Join(dir, "Assets")
		projSettingsPath := filepath.Join(dir, "ProjectSettings", "ProjectVersion.txt")
		if _, err := os.Stat(assetsPath); err == nil {
			if _, err := os.Stat(projSettingsPath); err == nil {
				return dir, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("no Unity project found in current or parent directories")
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
	fmt.Println("  atlas graph scenes              List all indexed scenes (graph API)")
	fmt.Println("  atlas graph scene <name>        Show details of a scene")
	fmt.Println("  atlas graph gameobjects <scene> List GameObjects in a scene")
	fmt.Println("  atlas graph scripts <scene>     List scripts used in a scene")
	fmt.Println("  atlas graph prefabs             List all prefabs")
	fmt.Println("  atlas graph assets              List all assets")
	fmt.Println("  atlas graph stats               Show project statistics")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  atlas init --dry-run")
	fmt.Println("  atlas index node --type=scene --guid=d20ebab... --name=\"MainMenu\"")
	fmt.Println("  atlas index scene --full --file scene_data.json")
	fmt.Println("  atlas query scenes")
	fmt.Println("  atlas query gameobjects --scene=d20ebab...")
	fmt.Println("  atlas analyze lighting")
	fmt.Println("  atlas graph stats")
}
