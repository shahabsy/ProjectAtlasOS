package tools

import (
	"encoding/json"
)

// Callable is a function that executes a tool with JSON arguments and returns JSON.
type Callable func(args json.RawMessage) ([]byte, error)

// ToolRegistry holds all available tools by domain.
type ToolRegistry struct {
	Scene      *SceneTools
	GameObject *GameObjectTools
	Component  *ComponentTools
	Script     *ScriptTools
	Asset      *AssetTools
	Stats      *StatsTools
}

// NewToolRegistry creates a registry with all tools.
func NewToolRegistry(ctx *Context) *ToolRegistry {
	return &ToolRegistry{
		Scene:      NewSceneTools(ctx),
		GameObject: NewGameObjectTools(ctx),
		Component:  NewComponentTools(ctx),
		Script:     NewScriptTools(ctx),
		Asset:      NewAssetTools(ctx),
		Stats:      NewStatsTools(ctx),
	}
}

// AllTools returns all tools as a map of name → ToolMetaProvider.
// Used for both CLI command generation and AI function-calling registration.
func (r *ToolRegistry) AllTools() map[string]ToolMetaProvider {
	return map[string]ToolMetaProvider{
		"scene":      r.Scene,
		"gameobject": r.GameObject,
		"component":  r.Component,
		"script":     r.Script,
		"asset":      r.Asset,
		"stats":      r.Stats,
	}
}

// ToolNames returns a slice of all tool names.
func (r *ToolRegistry) ToolNames() []string {
	names := make([]string, 0, len(r.AllTools()))
	for name := range r.AllTools() {
		names = append(names, name)
	}
	return names
}

// AllCallables returns a map of tool name → Callable.
func (r *ToolRegistry) AllCallables() map[string]Callable {
	reg := make(map[string]Callable)

	// Scene tools
	reg["list_scenes"] = func(args json.RawMessage) ([]byte, error) {
		resp, err := r.Scene.ListScenes()
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}
	reg["describe_scene"] = func(args json.RawMessage) ([]byte, error) {
		var req struct{ SceneID string }
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.Scene.DescribeScene(req.SceneID)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}

	// GameObject tools
	reg["find_gameobjects"] = func(args json.RawMessage) ([]byte, error) {
		var req FindGameObjectsRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.GameObject.FindGameObjects(req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}
	reg["get_gameobject_hierarchy"] = func(args json.RawMessage) ([]byte, error) {
		var req struct{ GameObjectID string }
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.GameObject.GetGameObjectHierarchy(req.GameObjectID)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}
	reg["find_gameobjects_with_component"] = func(args json.RawMessage) ([]byte, error) {
		var req FindGameObjectsWithComponentRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.GameObject.FindGameObjectsWithComponent(req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}

	// Component tools
	reg["get_components"] = func(args json.RawMessage) ([]byte, error) {
		var req GetComponentsRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.Component.GetComponents(req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}
	reg["find_components_by_type"] = func(args json.RawMessage) ([]byte, error) {
		var req FindComponentsByTypeRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.Component.FindComponentsByType(req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}

	// Script tools
	reg["find_scripts"] = func(args json.RawMessage) ([]byte, error) {
		var req FindScriptsRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.Script.FindScripts(req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}
	reg["get_gameobjects_using_script"] = func(args json.RawMessage) ([]byte, error) {
		var req GetGameObjectsUsingScriptRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.Script.GetGameObjectsUsingScript(req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}

	// Asset tools
	reg["list_assets"] = func(args json.RawMessage) ([]byte, error) {
		var req ListAssetsRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.Asset.ListAssets(req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}

	// Stats tools
	reg["project_stats"] = func(args json.RawMessage) ([]byte, error) {
		var req ProjectStatsRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.Stats.ProjectStats(req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}
	reg["scene_stats"] = func(args json.RawMessage) ([]byte, error) {
		var req SceneStatsRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		resp, err := r.Stats.SceneStats(req)
		if err != nil {
			return nil, err
		}
		return json.Marshal(resp)
	}

	return reg
}
