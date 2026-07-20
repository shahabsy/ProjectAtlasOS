package core

import (
	"encoding/json"
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
)

// SceneData matches the JSON sent from Unity
type SceneData struct {
	Name        string           `json:"name"`
	Guid        string           `json:"guid"`
	GameObjects []GameObjectData `json:"gameObjects"`
}

type GameObjectData struct {
	Id         string          `json:"id"`
	Name       string          `json:"name"`
	Tags       []string        `json:"tags"`
	Layer      int             `json:"layer"`
	ParentId   string          `json:"parent_id"`
	Components []ComponentData `json:"components"`
}

type ComponentData struct {
	Type       string                 `json:"type"`
	Enabled    bool                   `json:"enabled"`
	Properties map[string]interface{} `json:"properties"`
}

func IndexFullScene(jsonData []byte) error {
	var scene SceneData
	if err := json.Unmarshal(jsonData, &scene); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	fmt.Printf("[Atlas DEBUG] Scene name: '%s'\n", scene.Name)
	fmt.Printf("[Atlas DEBUG] GUID: '%s'\n", scene.Guid)
	fmt.Printf("[Atlas DEBUG] Number of GameObjects: %d\n", len(scene.GameObjects))
	for i, gobj := range scene.GameObjects {
		fmt.Printf("  [%d] ID: %s, Name: '%s'\n", i, gobj.Id, gobj.Name)
	}

	if scene.Guid == "" {
		return fmt.Errorf("scene GUID is empty. Make sure the scene is saved and the path is valid.")
	}
	if len(scene.Guid) < 8 {
		return fmt.Errorf("scene GUID is too short: %s", scene.Guid)
	}

	projectRoot, err := findUnityProject()
	if err != nil {
		return err
	}
	dbPath := db.GetDBPath(projectRoot)

	// 1. Insert scene node (idempotent)
	sceneNodeId := "scene_" + scene.Guid[:8]
	if err := db.InsertNode(dbPath, sceneNodeId, "scene", scene.Guid, scene.Name); err != nil {
		return fmt.Errorf("failed to insert scene node: %w", err)
	}

	// 2. Insert all GameObjects and components
	for _, gobj := range scene.GameObjects {
		gobNodeId := gobj.Id

		// Insert GameObject node – use gobNodeId as unique guid
		if err := db.InsertNode(dbPath, gobNodeId, "gameobject", gobNodeId, gobj.Name); err != nil {
			return fmt.Errorf("failed to insert GameObject '%s': %w", gobj.Name, err)
		}

		// Edge: scene CONTAINS gameobject
		if err := db.InsertEdge(dbPath, sceneNodeId, gobNodeId, "CONTAINS"); err != nil {
			return fmt.Errorf("failed to link scene to GameObject: %w", err)
		}

		// Edge: parent-child relationship
		if gobj.ParentId != "" {
			if err := db.InsertEdge(dbPath, gobj.ParentId, gobNodeId, "PARENT_OF"); err != nil {
				return fmt.Errorf("failed to link parent-child: %w", err)
			}
		}

		// Index components
		for _, comp := range gobj.Components {
			compNodeId := gobNodeId + "_" + comp.Type
			compName := comp.Type
			if len(comp.Properties) > 0 {
				if name, ok := comp.Properties["name"]; ok {
					compName = fmt.Sprintf("%s (%v)", comp.Type, name)
				}
			}
			// Insert component node – use compNodeId as unique guid
			if err := db.InsertNode(dbPath, compNodeId, "component", compNodeId, compName); err != nil {
				return fmt.Errorf("failed to insert component: %w", err)
			}
			if err := db.InsertEdge(dbPath, gobNodeId, compNodeId, "HAS_COMPONENT"); err != nil {
				return fmt.Errorf("failed to link component: %w", err)
			}
		}
	}

	count, err := db.CountNodes(dbPath)
	if err != nil {
		fmt.Printf("[Atlas] Total nodes in DB after indexing %d\n", count)
	} else {
		fmt.Printf("[Atlas] Failed to count nodes: %v\n", err)
	}

	fmt.Printf("✅ Indexed scene '%s' with %d GameObjects\n", scene.Name, len(scene.GameObjects))
	return nil
}
