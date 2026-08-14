package core

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	_ "modernc.org/sqlite"
)

// ---------- Scene JSON structures ----------
type SceneData struct {
	Name        string           `json:"name"`
	Guid        string           `json:"guid"`
	GameObjects []GameObjectData `json:"gameObjects"`
}

type GameObjectData struct {
	Id         string          `json:"id"`
	GlobalId   string          `json:"globalId"`
	Name       string          `json:"name"`
	ParentId   string          `json:"parent_id"`
	PrefabGuid string          `json:"prefab_guid"`
	Components []ComponentData `json:"components"`
}

type ComponentData struct {
	GlobalId         string                `json:"globalId"`
	Type             string                `json:"type"`
	ClassName        string                `json:"class_name"`
	ScriptGuid       string                `json:"script_guid"`
	SerializedFields []SerializedFieldData `json:"serialized_fields"`
	AssetReferences  []AssetReferenceData  `json:"asset_references"`
}

type SerializedFieldData struct {
	Name        string `json:"name"`
	ReferenceId string `json:"reference_id"`
}

type AssetReferenceData struct {
	Guid string `json:"guid"`
	Type string `json:"type"`
	Name string `json:"name"`
}

// findUnityProject is already defined in init.go – we do NOT redeclare it here.

// IndexFullScene indexes a Unity scene from its JSON representation.
func IndexFullScene(jsonData []byte) error {
	var scene SceneData
	if err := json.Unmarshal(jsonData, &scene); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	fmt.Printf("[Atlas DEBUG] Scene name: '%s'\n", scene.Name)
	fmt.Printf("[Atlas DEBUG] GUID: '%s'\n", scene.Guid)
	fmt.Printf("[Atlas DEBUG] Number of GameObjects: %d\n", len(scene.GameObjects))

	if scene.Guid == "" {
		return fmt.Errorf("scene GUID is empty. Make sure the scene is saved and the path is valid.")
	}
	if len(scene.Guid) < 8 {
		return fmt.Errorf("scene GUID is too short: %s", scene.Guid)
	}

	// Use the project root finder from init.go
	projectRoot, err := findUnityProject()
	if err != nil {
		return err
	}
	dbPath := db.GetDBPath(projectRoot)

	// Open connection with 5‑second busy timeout
	conn, err := db.OpenWithBusyTimeout(dbPath, 5000)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer conn.Close()

	// Create the database schema before running migrations
	if err := db.CreateDB(dbPath); err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
	}

	// Ensure schema is up‑to‑date
	if err := db.RunMigrations(dbPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Begin transaction
	tx, err := db.Begin(conn)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.Rollback(tx)

	// Insert Scene Node using transaction
	sceneNodeId := "scene_" + scene.Guid[:8]
	if err := db.InsertNodeTx(tx, sceneNodeId, "scene", "scene", scene.Guid, "", scene.Name, ""); err != nil {
		return fmt.Errorf("failed to insert scene node: %w", err)
	}

	// (Optional maps – used for debugging or later)
	scriptsMap := make(map[string]bool)
	prefabsMap := make(map[string]bool)
	assetsMap := make(map[string]bool)

	// Helper to check node existence
	nodeExists := func(tx *sql.Tx, id string) (bool, error) {
		var exists bool
		err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM nodes WHERE id = ?)", id).Scan(&exists)
		return exists, err
	}

	// Process GameObjects
	for idx, gobj := range scene.GameObjects {
		// Use full GlobalId as node ID for GameObject
		gobNodeId := gobj.GlobalId
		if gobNodeId == "" {
			gobNodeId = fmt.Sprintf("gobj_fallback_%d", idx)
			fmt.Printf("[Atlas WARN] No GlobalId for GameObject '%s', using fallback.\n", gobj.Name)
		}

		// Insert GameObject node
		if err := db.InsertNodeTx(tx, gobNodeId, "gameobject", "", "", gobj.GlobalId, gobj.Name, ""); err != nil {
			return fmt.Errorf("failed to insert GameObject '%s': %w", gobj.Name, err)
		}
		// Edge: scene → GameObject
		if err := db.InsertEdgeTx(tx, sceneNodeId, gobNodeId, "CONTAINS", ""); err != nil {
			return fmt.Errorf("failed to link scene to GameObject: %w", err)
		}
		if gobj.ParentId != "" {
			if err := db.InsertEdgeTx(tx, gobj.ParentId, gobNodeId, "CHILD_OF", ""); err != nil {
				return fmt.Errorf("failed to link parent-child: %w", err)
			}
		}
		if gobj.PrefabGuid != "" {
			prefabsMap[gobj.PrefabGuid] = true
			prefabNodeId := "prefab_" + gobj.PrefabGuid
			if err := db.InsertNodeTx(tx, prefabNodeId, "prefab", "prefab", gobj.PrefabGuid, "", gobj.Name, ""); err != nil {
				return fmt.Errorf("failed to insert prefab node: %w", err)
			}
			if err := db.InsertEdgeTx(tx, gobNodeId, prefabNodeId, "INSTANCE_OF", ""); err != nil {
				return fmt.Errorf("failed to link GameObject to prefab: %w", err)
			}
		}

		// Process Components
		for _, comp := range gobj.Components {
			if len(comp.GlobalId) < 8 {
				fmt.Printf("[Atlas WARN] Component GlobalId too short: '%s', skipping\n", comp.GlobalId)
				continue
			}
			// Use full GlobalId as node ID for Component
			compNodeId := comp.GlobalId
			if err := db.InsertNodeTx(tx, compNodeId, "component", comp.Type, "", comp.GlobalId, comp.Type, ""); err != nil {
				return fmt.Errorf("failed to insert component '%s': %w", comp.Type, err)
			}
			if err := db.InsertEdgeTx(tx, gobNodeId, compNodeId, "HAS_COMPONENT", ""); err != nil {
				return fmt.Errorf("failed to link component to GameObject: %w", err)
			}

			// Script Edge
			if comp.ScriptGuid != "" {
				scriptNodeId := "script_" + comp.ScriptGuid
				scriptsMap[comp.ScriptGuid] = true

				// Insert script node
				if err := db.InsertNodeTx(tx, scriptNodeId, "script", "script", comp.ScriptGuid, "", comp.ClassName, ""); err != nil {
					return fmt.Errorf("failed to insert script node (id: %s, guid: %s): %w", scriptNodeId, comp.ScriptGuid, err)
				}

				// Verify the script node exists before creating edge
				exists, err := nodeExists(tx, scriptNodeId)
				if err != nil {
					return fmt.Errorf("failed to verify script node existence: %w", err)
				}
				if !exists {
					return fmt.Errorf("script node '%s' (guid: %s) does not exist after insertion", scriptNodeId, comp.ScriptGuid)
				}

				// Verify component node exists (should, but double‑check)
				compExists, err := nodeExists(tx, compNodeId)
				if err != nil {
					return fmt.Errorf("failed to verify component node existence: %w", err)
				}
				if !compExists {
					return fmt.Errorf("component node '%s' does not exist before creating USES_SCRIPT edge", compNodeId)
				}

				fmt.Printf("[Atlas DEBUG] Creating USES_SCRIPT edge from %s to %s\n", compNodeId, scriptNodeId)
				if err := db.InsertEdgeTx(tx, compNodeId, scriptNodeId, "USES_SCRIPT", ""); err != nil {
					return fmt.Errorf("failed to link component to script (component: %s, script: %s): %w", compNodeId, scriptNodeId, err)
				}
				fmt.Printf("[Atlas DEBUG] Added USES_SCRIPT edge from %s to %s\n", compNodeId, scriptNodeId)
			}

			// Serialized Fields
			for _, field := range comp.SerializedFields {
				fieldNodeId := compNodeId + "_field_" + field.Name
				if err := db.InsertNodeTx(tx, fieldNodeId, "serialized_field", "", "", "", field.Name, ""); err != nil {
					return fmt.Errorf("failed to insert serialized field node: %w", err)
				}
				if err := db.InsertEdgeTx(tx, compNodeId, fieldNodeId, "HAS_FIELD", ""); err != nil {
					return fmt.Errorf("failed to link serialized field to component: %w", err)
				}
				if field.ReferenceId != "" {
					if len(field.ReferenceId) < 8 {
						fmt.Printf("[Atlas WARN] ReferenceId too short: '%s', skipping asset edge\n", field.ReferenceId)
						continue
					}
					assetsMap[field.ReferenceId] = true
					targetNodeId := "asset_" + field.ReferenceId
					// Insert asset node (may already exist)
					if err := db.InsertNodeTx(tx, targetNodeId, "asset", "asset", field.ReferenceId, "", field.Name, ""); err != nil {
						return fmt.Errorf("failed to insert asset node for reference: %w", err)
					}
					if err := db.InsertEdgeTx(tx, fieldNodeId, targetNodeId, "REFERENCES", ""); err != nil {
						return fmt.Errorf("failed to link serialized field to asset: %w", err)
					}
				}
			}

			// Asset References (may update existing asset nodes with correct type/name)
			for _, assetRef := range comp.AssetReferences {
				if len(assetRef.Guid) < 8 {
					fmt.Printf("[Atlas WARN] Asset GUID too short: '%s', skipping\n", assetRef.Guid)
					continue
				}
				assetsMap[assetRef.Guid] = true
				assetNodeId := "asset_" + assetRef.Guid
				if err := db.InsertNodeTx(tx, assetNodeId, "asset", assetRef.Type, assetRef.Guid, "", assetRef.Name, ""); err != nil {
					return fmt.Errorf("failed to insert asset node: %w", err)
				}
				rel := "USES_" + assetRef.Type
				if err := db.InsertEdgeTx(tx, compNodeId, assetNodeId, rel, ""); err != nil {
					return fmt.Errorf("failed to link asset: %w", err)
				}
			}
		}
	}

	// Commit
	if err := db.CommitTx(tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("✅ Indexed scene '%s' with %d GameObjects\n", scene.Name, len(scene.GameObjects))
	return nil
}
