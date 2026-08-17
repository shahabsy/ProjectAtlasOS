package tools

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

type StatsTools struct{ ctx *Context }

func NewStatsTools(ctx *Context) *StatsTools { return &StatsTools{ctx: ctx} }

func (t *StatsTools) Contract() Contract {
	return Contract{
		Name:        "stats_tools",
		Description: "Tools for project statistics.",
		InputSchema: Schema{Type: "object", Properties: map[string]Property{}, Required: []string{}},
		ReadOnly:    true,
	}
}

type ProjectStatsResponse struct {
	Summary *models.ProjectSummary `json:"summary"`
}

func (t *StatsTools) ProjectStats() Result {
	summary, err := t.ctx.Statistics.ProjectSummary()
	if err != nil {
		return NewErrorResult("STATS_FAILED", fmt.Sprintf("project_stats: %v", err))
	}
	return NewSuccessResult(ProjectStatsResponse{Summary: summary})
}

type SceneStatsResponse struct {
	Statistics *models.SceneStatistics `json:"statistics"`
}

func (t *StatsTools) SceneStats(sceneID string) Result {
	if sceneID == "" {
		return NewErrorResult("MISSING_ARGUMENT", "scene_id is required")
	}
	stats, err := t.ctx.Statistics.SceneSummary(sceneID)
	if err != nil {
		return NewErrorResult("STATS_FAILED", fmt.Sprintf("scene_stats: %v", err))
	}
	return NewSuccessResult(SceneStatsResponse{Statistics: stats})
}

type GetGraphCapabilitiesResponse struct {
	ProjectOverview struct {
		Scenes      int `json:"scenes"`
		GameObjects int `json:"game_objects"`
		Components  int `json:"components"`
		Scripts     int `json:"scripts"`
		Assets      int `json:"assets"`
	} `json:"project_overview"`
	Capabilities models.DataAvailability `json:"capabilities"`
}

func (t *StatsTools) GetGraphCapabilities() Result {
	summary, err := t.ctx.Statistics.ProjectSummary()
	if err != nil {
		return NewErrorResult("STATS_FAILED", fmt.Sprintf("get_graph_capabilities: %v", err))
	}
	caps := models.DataAvailability{
		Scenes:           true,
		GameObjects:      true,
		Components:       true,
		Scripts:          true,
		SerializedFields: false,
		AssetReferences:  false,
		Prefabs:          false,
	}
	resp := GetGraphCapabilitiesResponse{
		ProjectOverview: struct {
			Scenes      int `json:"scenes"`
			GameObjects int `json:"game_objects"`
			Components  int `json:"components"`
			Scripts     int `json:"scripts"`
			Assets      int `json:"assets"`
		}{
			Scenes:      summary.TotalScenes,
			GameObjects: summary.TotalGameObjects,
			Components:  summary.TotalComponents,
			Scripts:     summary.UniqueScripts,
			Assets:      summary.TotalAssets,
		},
		Capabilities: caps,
	}
	return NewSuccessResult(resp)
}
