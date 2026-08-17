package tools

import (
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

type ComponentTools struct{ ctx *Context }

func NewComponentTools(ctx *Context) *ComponentTools { return &ComponentTools{ctx: ctx} }

func (t *ComponentTools) Contract() Contract {
	return Contract{
		Name:        "component_tools",
		Description: "Tools for querying Components.",
		InputSchema: Schema{
			Type: "object",
			Properties: map[string]Property{
				"game_object_id": {
					Type:        "string",
					Description: "ID of the GameObject whose components to list",
				},
			},
			Required: []string{"game_object_id"},
		},
		ReadOnly: true,
	}
}

// ListComponentsResponse is the output for ListComponents.
type ListComponentsResponse struct {
	Components []*models.Component `json:"components"`
}

// ListComponents returns all components attached to a given GameObject.
func (t *ComponentTools) ListComponents(gameObjectID string) Result {
	if gameObjectID == "" {
		return NewErrorResult("MISSING_ARGUMENT", "game_object_id is required")
	}
	comps, err := t.ctx.Query.GetComponentsByGameObject(gameObjectID)
	if err != nil {
		return NewErrorResult("QUERY_FAILED", err.Error())
	}
	return NewSuccessResult(ListComponentsResponse{Components: comps})
}
