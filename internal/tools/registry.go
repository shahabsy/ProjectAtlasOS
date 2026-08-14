package tools

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
