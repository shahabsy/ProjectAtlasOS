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

// NewToolRegistry creates a registry with all tools.
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
	return r
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
