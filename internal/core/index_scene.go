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
	GlobalId   string          `json:"globalId"`
	Name       string          `json:"name"`
	Tags       []string        `json:"tags"`
	Layer      int             `json:"layer"`
	ParentId   string          `json:"parent_id"`
	PrefabGuid string          `json:"prefab_guid"`
	Components []ComponentData `json:"components"`
}

type ComponentData struct {
	GlobalId         string            `json:"globalId"`
	Type             string            `json:"type"`
	Enabled          bool              `json:"enabled"`
	ParentId         string            `json:"parent_id"`
	ScriptGuid       string            `json:"script_guid"`
	ClassName        string            `json:"class_name"`
	NamespaceName    string            `json:"namespace_name"`
	SerializedFields []SerializedField `json:"serialized_fields"`
	AssetReferences  []AssetReference  `json:"asset_references"`
}

type SerializedField struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Value         string `json:"value"`
	ReferenceType string `json:"reference_type"`
	ReferenceId   string `json:"reference_id"`
	ReferencePath string `json:"reference_path"`
}

type AssetReference struct {
	Type     string `json:"type"`
	Guid     string `json:"guid"`
	Name     string `json:"name"`
	SlotName string `json:"slot_name"`
}

func IndexFullScene(jsonData []byte) error {
	var scene SceneData
	if err := json.Unmarshal(jsonData, &scene); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	fmt.Printf("[Atlas DEBUG] Scene name: '%s'\n", scene.Name)
	fmt.Printf("[Atlas DEBUG] GUID: '%s'\n", scene.Guid)
	fmt.Printf("[Atlas DEBUG] Number of GameObjects: %d\n", len(scene.GameObjects))
	projectRoot, err := findUnityProject()
	if err != nil {
		return err
	}
	dbPath := db.GetDBPath(projectRoot)

	// 1. Insert Scene Node
	sceneNodeId := "scene_" + scene.Guid[:8]
	if err := db.InsertNode(dbPath, sceneNodeId, "scene", scene.Guid, scene.Guid, scene.Name, ""); err != nil {
		return fmt.Errorf("failed to insert scene node: %w", err)
	}
	// 2. Track scripts, prefabs, assets for deduplication
	scriptsMap := make(map[string]bool)
	prefabsMap := make(map[string]bool)
	assetsMap := make(map[string]bool)

	// 3. Process GameObjects
	for i, gobj := range scene.GameObjects {
		fmt.Printf("  [%d] ID: %s, Name: '%s'\n", i, gobj.Id, gobj.Name)
	}

	if scene.Guid == "" {
		return fmt.Errorf("scene GUID is empty. Make sure the scene is saved and the path is valid.")
	}
	if len(scene.Guid) < 8 {
		return fmt.Errorf("scene GUID is too short: %s", scene.Guid)
	}

	// 2. Insert all GameObjects and components
	for i, gobj := range scene.GameObjects {
		fmt.Printf("[%d] Processing GameObject: %s (GlobalId: %s)\n", i, gobj.Name, gobj.GlobalId)

		gobNodeId := gobj.Id
		if gobNodeId == "" {
			gobNodeId = fmt.Sprintf("gobj_fallback_%d", i)
			fmt.Printf("[Atlas Warn] No stable ID for game objects '%s'", gobj.Name)
		}
		fmt.Printf("Inserting GameObject with ID: %s\n", gobNodeId)
		// Insert GameObject node – use gobNodeId as unique guid
		if err := db.InsertNode(dbPath, gobNodeId, "gameobject", "", gobj.GlobalId, gobj.Name, ""); err != nil {
			return fmt.Errorf("failed to insert GameObject '%s': %w", gobj.Name, err)
		}

		// Edge: scene CONTAINS gameobject
		if err := db.InsertEdge(dbPath, sceneNodeId, gobNodeId, "CONTAINS", ""); err != nil {
			return fmt.Errorf("failed to link scene to GameObject: %w", err)
		}

		// Edge: parent-child relationship
		if gobj.ParentId != "" {
			if err := db.InsertEdge(dbPath, gobj.ParentId, gobNodeId, "CHILD_OF", ""); err != nil {
				return fmt.Errorf("failed to link parent-child: %w", err)
			}
		}
		if gobj.PrefabGuid != "" {
			prefabsMap[gobj.PrefabGuid] = true
			prefabNodeId := "prefab_" + gobj.PrefabGuid[:8]
			if err := db.InsertNode(dbPath, prefabNodeId, "prefab", gobj.PrefabGuid, "", gobj.Name, ""); err != nil {
				return fmt.Errorf("failed to insert prefab node:%w", err)
			}
			if err := db.InsertEdge(dbPath, gobNodeId, prefabNodeId, "INSTANCE_OF", ""); err != nil {
				return fmt.Errorf("failed to link GameObject to prefab: %w", err)
			}
		}
		// 4. Process Components
		for i, comp := range gobj.Components {
			compNodeId := comp.GlobalId
			if compNodeId == "" {
				compNodeId = fmt.Sprintf("comp_fallback_%s", i)
				fmt.Printf("[Atlas Warn] No GlobalId for component '%s', using fallback.\n", comp.Type)
			}
			if err := db.InsertNode(dbPath, compNodeId, "component", "", comp.GlobalId, comp.Type, ""); err != nil {
				return fmt.Errorf("failed to insert component '%s': %w", comp.Type, err)
			}
			if err := db.InsertEdge(dbPath, gobNodeId, compNodeId, "HAS_COMPONENT", ""); err != nil {
				return fmt.Errorf("failed to link component to GameObject: %w", err)
			}

			// Scripts
			if comp.ScriptGuid != "" {
				scriptsMap[comp.ScriptGuid] = true
				scriptNodeId := "script_" + comp.ScriptGuid[:8]
				if err := db.InsertNode(dbPath, scriptNodeId, "script", comp.ScriptGuid, "", comp.ClassName, ""); err != nil {
					return fmt.Errorf("failed to insert script node: %w", err)
				}
				if err := db.InsertEdge(dbPath, compNodeId, scriptNodeId, "USES_SCRIPT", ""); err != nil {
					return fmt.Errorf("failed to link component to script: %w", err)
				}
			}

			// Serialized Fields
			for _, field := range comp.SerializedFields {
				fieldNodeId := compNodeId + "_field_" + field.Name
				if err := db.InsertNode(dbPath, fieldNodeId, "serialized_field", "", "", field.Name, ""); err != nil {
					return fmt.Errorf("failed to insert serialized field node: %w", err)
				}
				if err := db.InsertEdge(dbPath, compNodeId, fieldNodeId, "HAS_FIELD", ""); err != nil {
					return fmt.Errorf("failed to link serialized field to component: %w", err)
				}
				if field.ReferenceId != "" {
					targetNodeId := field.ReferenceId
					if len(targetNodeId) > 32 {
						targetNodeId = "asset_" + targetNodeId[:8]
					} else {
						targetNodeId = "node_" + targetNodeId
					}
					assetsMap[field.ReferenceId] = true
					if err := db.InsertEdge(dbPath, fieldNodeId, targetNodeId, "REFERENCES", ""); err != nil {
						return fmt.Errorf("failed to link serialized field to asset: %w", err)
					}
				}
			}

			// Asset References (materials, textures, shaders, etc.)
			for _, assetRef := range comp.AssetReferences {
				assetsMap[assetRef.Guid] = true
				assetNodeId := "asset_" + assetRef.Guid[:8]
				if err := db.InsertNode(dbPath, assetNodeId, assetRef.Type, assetRef.Guid, "", assetRef.Name, ""); err != nil {
					return fmt.Errorf("failed to link serialized field to asset: %w", err)
				}
				rel := "USES_" + assetRef.Type
				if err := db.InsertEdge(dbPath, compNodeId, assetNodeId, rel, ""); err != nil {
					return fmt.Errorf("failed to link asset: %w", err)
				}
			}
		}
	}
	// 5. Optionally, you can log the counts of unique scripts, prefabs, and assets
	count, err := db.CountNodes(dbPath)
	if err != nil {
		fmt.Printf("[Atlas] Total nodes in DB after indexing %d\n", count)
	} else {
		fmt.Printf("[Atlas] Failed to count nodes: %v\n", err)
	}

	fmt.Printf("✅ Indexed scene '%s' with %d GameObjects\n", scene.Name, len(scene.GameObjects))
	return nil
}
