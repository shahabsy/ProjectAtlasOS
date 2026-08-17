package tools

import (
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

type GameObjectTools struct{ ctx *Context }

func NewGameObjectTools(ctx *Context) *GameObjectTools { return &GameObjectTools{ctx: ctx} }

func (t *GameObjectTools) Contract() Contract {
	return Contract{
		Name:        "gameobject_tools",
		Description: "Tools for querying GameObjects.",
		InputSchema: Schema{
			Type: "object",
			Properties: map[string]Property{
				"scene_id": {
					Type:        "string",
					Description: "ID of the scene whose GameObjects to list",
				},
			},
			Required: []string{"scene_id"},
		},
		ReadOnly: true,
	}
}

// ListGameObjectsResponse is the output for ListGameObjects.
type ListGameObjectsResponse struct {
	GameObjects []*models.GameObject `json:"game_objects"`
}

// ListGameObjects returns all GameObjects in a given scene.
func (t *GameObjectTools) ListGameObjects(sceneID string) Result {
	if sceneID == "" {
		return NewErrorResult("MISSING_ARGUMENT", "scene_id is required")
	}
	objs, err := t.ctx.Query.GetGameObjectsByScene(sceneID)
	if err != nil {
		return NewErrorResult("QUERY_FAILED", err.Error())
	}
	return NewSuccessResult(ListGameObjectsResponse{GameObjects: objs})
}
