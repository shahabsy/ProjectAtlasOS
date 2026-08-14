package tools

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// StatsTools provides tools for statistics.
type StatsTools struct {
	ctx *Context
}

// NewStatsTools creates a new StatsTools instance.
func NewStatsTools(ctx *Context) *StatsTools {
	return &StatsTools{ctx: ctx}
}

// Meta implements ToolMetaProvider.
func (t *StatsTools) Meta() ToolMetadata {
	return ToolMetadata{
		Name:        "project_stats",
		Description: "Returns aggregated project statistics (scenes, GameObjects, components, scripts, assets).",
		Parameters:  []Parameter{},
	}
}

// ---------- ProjectStats ----------

// ProjectStatsRequest is the input for ProjectStats.
type ProjectStatsRequest struct{}

// ProjectStatsResponse is the output for ProjectStats.
type ProjectStatsResponse struct {
	Summary *models.ProjectSummary `json:"summary"`
}

// ProjectStats returns aggregated project statistics.
func (t *StatsTools) ProjectStats(req ProjectStatsRequest) (ProjectStatsResponse, error) {
	summary, err := t.ctx.Statistics.ProjectSummary()
	if err != nil {
		return ProjectStatsResponse{}, fmt.Errorf("project_stats: %w", err)
	}
	return ProjectStatsResponse{Summary: summary}, nil
}

// ---------- SceneStats ----------

// SceneStatsRequest is the input for SceneStats.
type SceneStatsRequest struct {
	SceneID string `json:"scene_id"`
}

// SceneStatsResponse is the output for SceneStats.
type SceneStatsResponse struct {
	Statistics *models.SceneStatistics `json:"statistics"`
}

// SceneStats returns statistics for a specific scene.
func (t *StatsTools) SceneStats(req SceneStatsRequest) (SceneStatsResponse, error) {
	if req.SceneID == "" {
		return SceneStatsResponse{}, fmt.Errorf("scene_stats: scene_id is required")
	}

	stats, err := t.ctx.Statistics.SceneSummary(req.SceneID)
	if err != nil {
		return SceneStatsResponse{}, fmt.Errorf("scene_stats: %w", err)
	}
	return SceneStatsResponse{Statistics: stats}, nil
}
