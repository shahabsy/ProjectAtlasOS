package tools

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// SceneTools provides tools for working with scenes.
type SceneTools struct {
	ctx *Context
}

// NewSceneTools creates a new SceneTools instance.
func NewSceneTools(ctx *Context) *SceneTools {
	return &SceneTools{ctx: ctx}
}

// Meta implements ToolMetaProvider.
func (t *SceneTools) Meta() ToolMetadata {
	return ToolMetadata{
		Name:        "list_scenes",
		Description: "Returns all scenes in the Unity project.",
		Parameters:  []Parameter{},
	}
}

// ---------- ListScenes ----------

// ListScenesResponse is the output for ListScenes.
type ListScenesResponse struct {
	Scenes []*models.Scene `json:"scenes"`
}

// ListScenes returns all scenes in the project.
func (t *SceneTools) ListScenes() (ListScenesResponse, error) {
	scenes, err := t.ctx.Query.ListScenes()
	if err != nil {
		return ListScenesResponse{}, fmt.Errorf("list_scenes: %w", err)
	}
	return ListScenesResponse{Scenes: scenes}, nil
}

// ---------- DescribeScene ----------

// DescribeSceneResponse is the output for DescribeScene.
type DescribeSceneResponse struct {
	Scene      *models.Scene           `json:"scene"`
	Statistics *models.SceneStatistics `json:"statistics"`
	Warnings   []string                `json:"warnings,omitempty"`
}

// DescribeScene returns detailed information about a scene.
func (t *SceneTools) DescribeScene(sceneID string) (DescribeSceneResponse, error) {
	if sceneID == "" {
		return DescribeSceneResponse{}, fmt.Errorf("describe_scene: scene_id is required")
	}

	scene, err := t.ctx.Query.GetSceneByID(sceneID)
	if err != nil {
		return DescribeSceneResponse{}, fmt.Errorf("describe_scene: failed to get scene: %w", err)
	}
	if scene == nil {
		return DescribeSceneResponse{}, fmt.Errorf("describe_scene: scene not found: %s", sceneID)
	}

	stats, err := t.ctx.Statistics.SceneSummary(sceneID)
	var warnings []string
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("statistics unavailable: %v", err))
		stats = &models.SceneStatistics{} // safe zero-value fallback
	}
	if stats.ScriptCount == 0 {
		warnings = append(warnings, "ScriptCount=0 is a known indexing limitation; use project_stats for accurate script totals")
	}

	return DescribeSceneResponse{
		Scene:      scene,
		Statistics: stats,
		Warnings:   warnings,
	}, nil
}
