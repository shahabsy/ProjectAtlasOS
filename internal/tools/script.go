package tools

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// ScriptTools provides tools for scripts.
type ScriptTools struct {
	ctx *Context
}

// NewScriptTools creates a new ScriptTools instance.
func NewScriptTools(ctx *Context) *ScriptTools {
	return &ScriptTools{ctx: ctx}
}

// Meta implements ToolMetaProvider.
func (t *ScriptTools) Meta() ToolMetadata {
	return ToolMetadata{
		Name:        "find_scripts",
		Description: "Finds scripts by name (case‑insensitive partial match).",
		Parameters: []Parameter{
			{Name: "name", Type: "string", Description: "Script name to search for", Required: true},
		},
	}
}

// ---------- FindScripts ----------

// FindScriptsRequest is the input for FindScripts.
type FindScriptsRequest struct {
	Name string `json:"name"`
}

// FindScriptsResponse is the output for FindScripts.
type FindScriptsResponse struct {
	Scripts []*models.Script `json:"scripts"`
}

// FindScripts finds scripts by name (substring match).
func (t *ScriptTools) FindScripts(req FindScriptsRequest) (FindScriptsResponse, error) {
	if req.Name == "" {
		return FindScriptsResponse{}, fmt.Errorf("find_scripts: name is required")
	}

	scripts, err := t.ctx.Query.GetScriptsByName(req.Name)
	if err != nil {
		return FindScriptsResponse{}, fmt.Errorf("find_scripts: %w", err)
	}
	return FindScriptsResponse{Scripts: scripts}, nil
}

// ---------- GetGameObjectsUsingScript ----------

// GetGameObjectsUsingScriptRequest is the input for GetGameObjectsUsingScript.
type GetGameObjectsUsingScriptRequest struct {
	ScriptID string `json:"script_id"`
}

// GetGameObjectsUsingScriptResponse is the output for GetGameObjectsUsingScript.
type GetGameObjectsUsingScriptResponse struct {
	GameObjects []*models.GameObject `json:"game_objects"`
}

// GetGameObjectsUsingScript finds all GameObjects that have a component using the given script.
func (t *ScriptTools) GetGameObjectsUsingScript(req GetGameObjectsUsingScriptRequest) (GetGameObjectsUsingScriptResponse, error) {
	if req.ScriptID == "" {
		return GetGameObjectsUsingScriptResponse{}, fmt.Errorf("get_gameobjects_using_script: script_id is required")
	}

	gobjs, err := t.ctx.Query.GetGameObjectsUsingScript(req.ScriptID)
	if err != nil {
		return GetGameObjectsUsingScriptResponse{}, fmt.Errorf("get_gameobjects_using_script: %w", err)
	}
	return GetGameObjectsUsingScriptResponse{GameObjects: gobjs}, nil
}
