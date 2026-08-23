package tools

import (
	"encoding/json"
)

// ToolRegistry holds all available tools by domain.
type ToolRegistry struct {
	Scene      *SceneTools
	GameObject *GameObjectTools
	Component  *ComponentTools
	Script     *ScriptTools
	Asset      *AssetTools
	Stats      *StatsTools
	contracts  map[string]Contract
}

// NewToolRegistry creates a registry with all tools and enriched descriptions.
func NewToolRegistry(ctx *Context) *ToolRegistry {
	r := &ToolRegistry{
		Scene:      NewSceneTools(ctx),
		GameObject: NewGameObjectTools(ctx),
		Component:  NewComponentTools(ctx),
		Script:     NewScriptTools(ctx),
		Asset:      NewAssetTools(ctx),
		Stats:      NewStatsTools(ctx),
		contracts:  make(map[string]Contract),
	}
	registerContract(r, r.Scene.Contract())
	registerContract(r, r.GameObject.Contract())
	registerContract(r, r.Component.Contract())
	registerContract(r, r.Script.Contract())
	registerContract(r, r.Asset.Contract())
	registerContract(r, r.Stats.Contract())

	r.overrideDescriptions()
	return r
}

func (r *ToolRegistry) overrideDescriptions() {
	overrides := map[string]string{
		"list_scenes":            "List all scenes in the Unity project. Returns scene IDs, names, and GUIDs.",
		"describe_scene":         "Describe a specific scene by its ID. Returns scene metadata and statistics (GameObject count, component count, script count). Note: asset_count may be 0 if assets are not directly referenced; use 'list_assets' for project-wide asset information.",
		"list_gameobjects":       "List all GameObjects in a specific scene (requires scene_id).",
		"list_all_gameobjects":   "List all GameObjects across the entire project (no scene_id needed). Internally fetches the first scene and lists its GameObjects.",
		"list_components":        "List all components attached to a specific GameObject. Use this to inspect what components a GameObject has, including script references.",
		"list_scripts":           "List all scripts used in a specific scene (requires scene_id). Returns script names, IDs, GUIDs, class names (where available), and paths.",
		"list_assets":            "List all assets in the entire project, regardless of scene association. Use this to get a complete inventory of all assets (scripts, prefabs, fonts, sprites, materials, etc.).",
		"list_all_scripts":       "List all scripts across the entire project (no scene_id needed). This is a convenience tool that internally fetches the first scene and lists its scripts.",
		"project_stats":          "Get overall project statistics: total scenes, GameObjects, components, scripts, unique scripts, assets, and serialized fields.",
		"scene_stats":            "Get detailed statistics for a specific scene: GameObjects, components, scripts, prefabs, assets, fields, and availability flags.",
		"get_graph_capabilities": "Return the capabilities of the graph database, such as which data types are available (scenes, GameObjects, components, scripts, serialized fields, etc.).",
	}
	for name, desc := range overrides {
		if c, ok := r.contracts[name]; ok {
			c.Description = desc
			r.contracts[name] = c
		}
	}
}

func registerContract(r *ToolRegistry, c Contract) {
	r.contracts[c.Name] = c
}

// GetContract returns the contract for a tool, or false if not found.
func (r *ToolRegistry) GetContract(name string) (Contract, bool) {
	c, ok := r.contracts[name]
	return c, ok
}

// AllTools returns a map of all tool names to their contracts.
func (r *ToolRegistry) AllTools() map[string]Contract {
	copy := make(map[string]Contract, len(r.contracts))
	for k, v := range r.contracts {
		copy[k] = v
	}
	return copy
}

// AllCallables returns a map of tool names to their executable functions.
func (r *ToolRegistry) AllCallables() map[string]func(json.RawMessage) ([]byte, error) {
	callables := make(map[string]func(json.RawMessage) ([]byte, error))

	// Scene
	callables["list_scenes"] = func(args json.RawMessage) ([]byte, error) {
		res := r.Scene.ListScenes()
		return json.Marshal(res)
	}
	callables["describe_scene"] = func(args json.RawMessage) ([]byte, error) {
		var params struct {
			SceneID string `json:"scene_id"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, err
		}
		res := r.Scene.DescribeScene(params.SceneID)
		return json.Marshal(res)
	}

	// GameObject
	callables["list_gameobjects"] = func(args json.RawMessage) ([]byte, error) {
		var params struct {
			SceneID string `json:"scene_id"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, err
		}
		res := r.GameObject.ListGameObjects(params.SceneID)
		return json.Marshal(res)
	}
	callables["list_all_gameobjects"] = func(args json.RawMessage) ([]byte, error) {
		sceneRes := r.Scene.ListScenes()
		if !sceneRes.Success {
			return json.Marshal(sceneRes)
		}
		var sceneData struct {
			Scenes []struct {
				ID string `json:"id"`
			} `json:"scenes"`
		}
		sceneBytes, _ := json.Marshal(sceneRes.Data)
		if err := json.Unmarshal(sceneBytes, &sceneData); err != nil {
			return nil, err
		}
		if len(sceneData.Scenes) == 0 {
			return json.Marshal(Result{Success: false, Error: &Error{Message: "no scenes found"}})
		}
		sceneID := sceneData.Scenes[0].ID
		goRes := r.GameObject.ListGameObjects(sceneID)
		return json.Marshal(goRes)
	}

	// Component
	callables["list_components"] = func(args json.RawMessage) ([]byte, error) {
		var params struct {
			GameObjectID string `json:"game_object_id"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, err
		}
		res := r.Component.ListComponents(params.GameObjectID)
		return json.Marshal(res)
	}

	// Script
	callables["list_scripts"] = func(args json.RawMessage) ([]byte, error) {
		var params struct {
			SceneID string `json:"scene_id"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, err
		}
		res := r.Script.ListScripts(params.SceneID)
		return json.Marshal(res)
	}
	callables["list_all_scripts"] = func(args json.RawMessage) ([]byte, error) {
		sceneRes := r.Scene.ListScenes()
		if !sceneRes.Success {
			return json.Marshal(sceneRes)
		}
		var sceneData struct {
			Scenes []struct {
				ID string `json:"id"`
			} `json:"scenes"`
		}
		sceneBytes, _ := json.Marshal(sceneRes.Data)
		if err := json.Unmarshal(sceneBytes, &sceneData); err != nil {
			return nil, err
		}
		if len(sceneData.Scenes) == 0 {
			return json.Marshal(Result{Success: false, Error: &Error{Message: "no scenes found"}})
		}
		sceneID := sceneData.Scenes[0].ID
		scriptRes := r.Script.ListScripts(sceneID)
		return json.Marshal(scriptRes)
	}

	// Asset
	callables["list_assets"] = func(args json.RawMessage) ([]byte, error) {
		res := r.Asset.ListAssets()
		return json.Marshal(res)
	}

	// Stats
	callables["project_stats"] = func(args json.RawMessage) ([]byte, error) {
		res := r.Stats.ProjectStats()
		return json.Marshal(res)
	}
	callables["scene_stats"] = func(args json.RawMessage) ([]byte, error) {
		var params struct {
			SceneID string `json:"scene_id"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return nil, err
		}
		res := r.Stats.SceneStats(params.SceneID)
		return json.Marshal(res)
	}
	callables["get_graph_capabilities"] = func(args json.RawMessage) ([]byte, error) {
		res := r.Stats.GetGraphCapabilities()
		return json.Marshal(res)
	}

	return callables
}
