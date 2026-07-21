package core

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	db "github.com/shahabsy/ProjectAtlasOS/internal/db" // ✅ Use alias to avoid conflict
)

func QueryCmd(args []string) error {
	if len(args) == 0 {
		return showQueryUsage()
	}

	subCmd := args[0]
	switch subCmd {
	case "scenes":
		return listScenes()
	case "gameobjects":
		return listGameObjects(args[1:])
	case "components":
		return listComponents(args[1:])
	default:
		return fmt.Errorf("unknown query subcommand: %s. Use 'atlas query --help'", subCmd)
	}
}

func showQueryUsage() error {
	fmt.Println(`atlas query - Explore the knowledge graph

Usage:
  atlas query scenes               List all indexed scenes
  atlas query gameobjects <options> List GameObjects in a scene
  atlas query components <options>  List components on a GameObject
  atlas query --help                Show this help message

Examples:
  	atlas query scenes
  	atlas query gameobjects --scene=d20ebab...
  	atlas query components --gameobject=gobj_12345
	 atlas query components --gameobject=gobj_1ABCDEF`)
	return nil
}

// listScenes shows all indexed scenes with timestamps
func listScenes() error {
	projectRoot, err := findUnityProject()
	if err != nil {
		return fmt.Errorf("not a Unity project: %w", err)
	}

	dbPath := db.GetDBPath(projectRoot)

	scenes, err := db.GetNodesByType(dbPath, "scene")
	if err != nil {
		return fmt.Errorf("failed to query scenes: %w", err)
	}

	count, countErr := db.GetSceneCount(dbPath)
	if countErr != nil {
		count = -1 // Indicate error in count
	}

	fmt.Printf("📁 Indexed Scenes (%d total):\n", count)

	if len(scenes) == 0 {
		fmt.Println("\n💡 No scenes indexed yet.")
		fmt.Println("   Run `atlas index scene --full` on a saved scene first.")
		return nil
	}

	for _, s := range scenes {
		t := time.Unix(s.CreatedAt, 0).Format("2006-01-02 15:04:05")
		fmt.Printf("  - %s\n", s.Name)
		fmt.Printf("    GUID: %s\n", s.GUID)
		fmt.Printf("    Indexed at: %s\n", t)
		if strings.HasPrefix(s.ID, "scene_auto") {
			fmt.Printf("    ⚠️  Auto-generated (was empty GUID)\n")
		}
	}

	return nil
}

// listGameObjects shows all GameObjects in a scene
func listGameObjects(args []string) error {
	var sceneGUID string

	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--scene=") {
			sceneGUID = strings.TrimPrefix(args[i], "--scene=")
		}
	}

	if sceneGUID == "" {
		return fmt.Errorf("missing --scene=<guid> argument. Use 'atlas query scenes' to find scenes")
	}

	projectRoot, err := findUnityProject()
	if err != nil {
		return err
	}
	dbPath := db.GetDBPath(projectRoot)

	// Find the scene node by GUID
	sceneNode, err := db.GetNodeByGUID(dbPath, sceneGUID)
	if err != nil {
		return fmt.Errorf("failed to find scene: %w", err)
	}
	if sceneNode == nil {
		return fmt.Errorf("scene with GUID '%s' not found. Run 'atlas query scenes' to list all indexed scenes", sceneGUID)
	}

	// Get GameObjects in this scene
	gobjs, err := getGameObjectsByScene(dbPath, sceneNode.ID)
	if err != nil {
		return fmt.Errorf("failed to query GameObjects: %w", err)
	}

	fmt.Printf("\nGameObjects in '%s' (%d total):\n", sceneNode.Name, len(gobjs))

	for _, gobj := range gobjs {
		fmt.Printf("  - %s (ID: %s)\n", gobj.Name, gobj.ID)
	}

	return nil
}

// listComponents shows all components on a GameObject
func listComponents(args []string) error {
	var gameObjID string

	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--gameobject=") {
			gameObjID = strings.TrimPrefix(args[i], "--gameobject=")
		}
	}

	if gameObjID == "" {
		return fmt.Errorf("missing --gameobject=<id> argument. Use 'atlas query gameobjects' to find GameObjects")
	}

	projectRoot, err := findUnityProject()
	if err != nil {
		return err
	}
	dbPath := db.GetDBPath(projectRoot)

	// Get components for this GameObject
	comps, err := db.GetComponentsByGameObject(dbPath, gameObjID)
	if err != nil {
		return fmt.Errorf("failed to query components: %w", err)
	}

	fmt.Printf("\nComponents on '%s' (%d total):\n", gameObjID, len(comps))

	for i, comp := range comps {
		fmt.Printf("%2d. %s\n", i+1, comp.Name)
	}

	return nil
}

// getGameObjectsByScene retrieves all GameObject nodes in a scene
func getGameObjectsByScene(dbPath, sceneNodeID string) ([]db.Node, error) {
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	rows, err := database.Query(`
		SELECT n.id, n.type, n.guid, n.name
		FROM nodes n
		JOIN edges e ON e.target = n.id
		WHERE e.source = ? AND e.relationship = 'CONTAINS' AND n.type = 'gameobject'
		ORDER BY n.name`, sceneNodeID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var nodes []db.Node // ✅ Now uses db alias correctly
	for rows.Next() {
		var n db.Node
		if err := rows.Scan(&n.ID, &n.Type, &n.GUID, &n.Name); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		nodes = append(nodes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}
