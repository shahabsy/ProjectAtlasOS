package tools

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

type SceneTools struct{ ctx *Context }

func NewSceneTools(ctx *Context) *SceneTools { return &SceneTools{ctx: ctx} }

// Contract returns the formal specification of SceneTools.
func (t *SceneTools) Contract() Contract {
	return Contract{
		Name:        "scene_tools",
		Description: "Tools for working with scenes.",
		InputSchema: Schema{Type: "object", Properties: map[string]Property{}, Required: []string{}},
		ReadOnly:    true,
	}
}

// ---------- ListScenes ----------

type ListScenesResponse struct {
	Scenes []*models.Scene `json:"scenes"`
}

func (t *SceneTools) ListScenes() Result {
	scenes, err := t.ctx.Query.ListScenes()
	if err != nil {
		return NewErrorResult("QUERY_FAILED", err.Error())
	}
	return NewSuccessResult(ListScenesResponse{Scenes: scenes})
}

// ---------- DescribeScene ----------

type DescribeSceneResponse struct {
	Scene      *models.Scene           `json:"scene"`
	Statistics *models.SceneStatistics `json:"statistics"`
	Warnings   []string                `json:"warnings,omitempty"`
}

func (t *SceneTools) DescribeScene(sceneID string) Result {
	if sceneID == "" {
		return NewErrorResult("MISSING_ARGUMENT", "scene_id is required")
	}
	scene, err := t.ctx.Query.GetSceneByID(sceneID)
	if err != nil {
		return NewErrorResult("QUERY_FAILED", err.Error())
	}
	if scene == nil {
		return NewErrorResult("SCENE_NOT_FOUND", fmt.Sprintf("Scene with ID '%s' not found", sceneID))
	}
	stats, err := t.ctx.Statistics.SceneSummary(sceneID)
	var warnings []string
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("statistics unavailable: %v", err))
		stats = &models.SceneStatistics{}
	}
	if stats.ScriptCount == 0 {
		warnings = append(warnings, "ScriptCount=0 is a known indexing limitation; use project_stats for accurate totals")
	}
	return NewSuccessResult(DescribeSceneResponse{
		Scene:      scene,
		Statistics: stats,
		Warnings:   warnings,
	})
}
