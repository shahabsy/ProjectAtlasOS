package graph

import "fmt"

type ProjectStats struct {
	SceneCount      int `json:"scene_count"`
	GameObjectCount int `json:"game_object_count"`
	ComponentCount  int `json:"component_count"`
	ScriptCount     int `json:"script_count"`
	PrefabCount     int `json:"prefab_count"`
	AssetCount      int `json:"asset_count"`
}

// ProjectStats returns aggregated project-wide counts.
func (q *SceneQueries) ProjectStats() (*ProjectStats, error) {
	var stats ProjectStats
	err := q.repo.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM nodes WHERE type = 'scene') AS scene_count,
			(SELECT COUNT(*) FROM nodes WHERE type = 'gameobject') AS gameobject_count,
			(SELECT COUNT(*) FROM nodes WHERE type = 'component') AS component_count,
			(SELECT COUNT(*) FROM nodes WHERE type = 'script') AS script_count,
			(SELECT COUNT(*) FROM nodes WHERE type = 'prefab') AS prefab_count,
			(SELECT COUNT(*) FROM nodes WHERE type IN ('material', 'shader', 'texture')) AS asset_count
	`).Scan(
		&stats.SceneCount,
		&stats.GameObjectCount,
		&stats.ComponentCount,
		&stats.ScriptCount,
		&stats.PrefabCount,
		&stats.AssetCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get project statistics: %w", err)
	}
	return &stats, nil
}
