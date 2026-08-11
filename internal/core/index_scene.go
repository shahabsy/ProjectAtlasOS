package core

import (
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

// findUnityProject is defined in init.go – do not redeclare here.

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

	projectRoot, err := findUnityProject()
	if err != nil {
		return err
	}
	dbPath := db.GetDBPath(projectRoot)

	conn, err := db.OpenWithBusyTimeout(dbPath, 5000)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer conn.Close()

	if err := db.CreateDB(dbPath); err != nil {
		return fmt.Errorf("failed to create database schema: %w", err)
	}
	if err := db.RunMigrations(dbPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	tx, err := db.Begin(conn)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer db.Rollback(tx)

	// Scene node (using first 8 chars of GUID – it's short and unique)
	sceneNodeId := "scene_" + scene.Guid[:8]
	if err := db.InsertNodeTx(tx, sceneNodeId, "scene", scene.Guid, "", scene.Name, ""); err != nil {
		return fmt.Errorf("failed to insert scene node: %w", err)
	}

	scriptsMap := make(map[string]bool)
	prefabsMap := make(map[string]bool)
	assetsMap := make(map[string]bool)

	// Process GameObjects
	for idx, gobj := range scene.GameObjects {
		// Use the full GlobalId as the node ID – guaranteed unique
		gobNodeId := gobj.GlobalId
		if gobNodeId == "" {
			gobNodeId = fmt.Sprintf("gobj_fallback_%d", idx)
			fmt.Printf("[Atlas WARN] No GlobalId for GameObject '%s', using fallback.\n", gobj.Name)
		}

		if err := db.InsertNodeTx(tx, gobNodeId, "gameobject", "", gobj.GlobalId, gobj.Name, ""); err != nil {
			return fmt.Errorf("failed to insert GameObject '%s': %w", gobj.Name, err)
		}
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
			prefabNodeId := "prefab_" + gobj.PrefabGuid // full GUID
			if err := db.InsertNodeTx(tx, prefabNodeId, "prefab", gobj.PrefabGuid, "", gobj.Name, ""); err != nil {
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
			// Use the full GlobalId as the component ID
			compNodeId := comp.GlobalId
			if err := db.InsertNodeTx(tx, compNodeId, "component", "", comp.GlobalId, comp.Type, ""); err != nil {
				return fmt.Errorf("failed to insert component '%s': %w", comp.Type, err)
			}
			if err := db.InsertEdgeTx(tx, gobNodeId, compNodeId, "HAS_COMPONENT", ""); err != nil {
				return fmt.Errorf("failed to link component to GameObject: %w", err)
			}

			// Script Edge
			if comp.ScriptGuid != "" {
				scriptsMap[comp.ScriptGuid] = true
				scriptNodeId := "script_" + comp.ScriptGuid
				if err := db.InsertNodeTx(tx, scriptNodeId, "script", comp.ScriptGuid, "", comp.ClassName, ""); err != nil {
					return fmt.Errorf("failed to insert script node: %w", err)
				}
				if err := db.InsertEdgeTx(tx, compNodeId, scriptNodeId, "USES_SCRIPT", ""); err != nil {
					return fmt.Errorf("failed to link component to script: %w", err)
				}
				fmt.Printf("[Atlas DEBUG] Added USES_SCRIPT edge from %s to %s\n", compNodeId, scriptNodeId)
			}

			// Serialized Fields
			for _, field := range comp.SerializedFields {
				fieldNodeId := compNodeId + "_field_" + field.Name
				if err := db.InsertNodeTx(tx, fieldNodeId, "serialized_field", "", "", field.Name, ""); err != nil {
					return fmt.Errorf("failed to insert serialized field node: %w", err)
				}
				if err := db.InsertEdgeTx(tx, compNodeId, fieldNodeId, "HAS_FIELD", ""); err != nil {
					return fmt.Errorf("failed to link serialized field to component: %w", err)
				}
				if field.ReferenceId != "" {
					// Insert asset node with prefix "asset_"
					assetNodeId := "asset_" + field.ReferenceId
					// Use INSERT OR IGNORE to avoid duplicate conflicts (if the asset already exists)
					// But we want to keep the data consistent, so we use INSERT OR REPLACE.
					if err := db.InsertNodeTx(tx, assetNodeId, "asset", field.ReferenceId, "", field.Name, ""); err != nil {
						return fmt.Errorf("failed to insert asset node for reference '%s': %w", field.ReferenceId, err)
					}
					assetsMap[field.ReferenceId] = true
					if err := db.InsertEdgeTx(tx, fieldNodeId, assetNodeId, "REFERENCES", ""); err != nil {
						return fmt.Errorf("failed to link serialized field to asset (target: %s, ref: %s): %w", assetNodeId, field.ReferenceId, err)
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
				if err := db.InsertNodeTx(tx, assetNodeId, assetRef.Type, assetRef.Guid, "", assetRef.Name, ""); err != nil {
					return fmt.Errorf("failed to insert asset node: %w", err)
				}
				rel := "USES_" + assetRef.Type
				if err := db.InsertEdgeTx(tx, compNodeId, assetNodeId, rel, ""); err != nil {
					return fmt.Errorf("failed to link asset: %w", err)
				}
			}
		}
	}

	if err := db.CommitTx(tx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("✅ Indexed scene '%s' with %d GameObjects\n", scene.Name, len(scene.GameObjects))
	return nil
}
