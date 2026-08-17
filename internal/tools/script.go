package tools

import (
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

type ScriptTools struct{ ctx *Context }

func NewScriptTools(ctx *Context) *ScriptTools { return &ScriptTools{ctx: ctx} }

func (t *ScriptTools) Contract() Contract {
	return Contract{
		Name:        "script_tools",
		Description: "Tools for querying Scripts.",
		InputSchema: Schema{
			Type: "object",
			Properties: map[string]Property{
				"scene_id": {
					Type:        "string",
					Description: "ID of the scene whose scripts to list",
				},
			},
			Required: []string{"scene_id"},
		},
		ReadOnly: true,
	}
}

// ListScriptsResponse is the output for ListScripts.
type ListScriptsResponse struct {
	Scripts []*models.Script `json:"scripts"`
}

// ListScripts returns all scripts in a given scene.
func (t *ScriptTools) ListScripts(sceneID string) Result {
	if sceneID == "" {
		return NewErrorResult("MISSING_ARGUMENT", "scene_id is required")
	}
	scripts, err := t.ctx.Query.GetScriptsByScene(sceneID)
	if err != nil {
		return NewErrorResult("QUERY_FAILED", err.Error())
	}
	return NewSuccessResult(ListScriptsResponse{Scripts: scripts})
}
