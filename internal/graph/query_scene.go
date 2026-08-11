package graph

import (
	"fmt"
)

// SceneQueries provides typed queries for scenes.
type SceneQueries struct {
	repo *Repository
}

func NewSceneQueries(repo *Repository) *SceneQueries {
	return &SceneQueries{repo: repo}
}

// List returns all scenes.
func (q *SceneQueries) List() ([]SceneInfo, error) {
	rows, err := q.repo.Query(`
		SELECT id, name, guid, created_at, updated_at, last_indexed_at
		FROM nodes WHERE type = 'scene' ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list scenes: %w", err)
	}
	defer rows.Close()

	var scenes []SceneInfo
	for rows.Next() {
		var s SceneInfo
		if err := rows.Scan(&s.ID, &s.Name, &s.GUID, &s.CreatedAt, &s.UpdatedAt, &s.LastIndexed); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		scenes = append(scenes, s)
	}
	return scenes, nil
}

// ByName returns a scene by exact name.
func (q *SceneQueries) ByName(name string) (*SceneInfo, error) {
	var s SceneInfo
	err := q.repo.QueryRow(`
		SELECT id, name, guid, created_at, updated_at, last_indexed_at
		FROM nodes WHERE type = 'scene' AND name = ?
	`, name).Scan(&s.ID, &s.Name, &s.GUID, &s.CreatedAt, &s.UpdatedAt, &s.LastIndexed)
	if err != nil {
		return nil, fmt.Errorf("scene '%s' not found: %w", name, err)
	}
	return &s, nil
}

// ByGUID returns a scene by asset GUID.
func (q *SceneQueries) ByGUID(guid string) (*SceneInfo, error) {
	var s SceneInfo
	err := q.repo.QueryRow(`
		SELECT id, name, guid, created_at, updated_at, last_indexed_at
		FROM nodes WHERE type = 'scene' AND guid = ?
	`, guid).Scan(&s.ID, &s.Name, &s.GUID, &s.CreatedAt, &s.UpdatedAt, &s.LastIndexed)
	if err != nil {
		return nil, fmt.Errorf("scene with GUID '%s' not found: %w", guid, err)
	}
	return &s, nil
}

// Statistics returns aggregated counts for a scene.
func (q *SceneQueries) Statistics(sceneID string) (*SceneStats, error) {
	var stats SceneStats
	err := q.repo.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM edges WHERE source = ? AND relationship = 'CONTAINS') AS gameobject_count,
			(SELECT COUNT(*) FROM nodes n
			 JOIN edges e ON e.target = n.id
			 WHERE e.source = ? AND n.type = 'component') AS component_count,
			(SELECT COUNT(DISTINCT e2.target) FROM nodes n
			 JOIN edges e ON e.target = n.id
			 JOIN edges e2 ON e2.source = n.id AND e2.relationship = 'USES_SCRIPT'
			 WHERE e.source = ?) AS script_count,
			(SELECT COUNT(DISTINCT e2.target) FROM nodes n
			 JOIN edges e ON e.target = n.id
			 JOIN edges e2 ON e2.source = n.id AND e2.relationship = 'INSTANCE_OF'
			 WHERE e.source = ?) AS prefab_count,
			(SELECT COUNT(DISTINCT e2.target) FROM nodes n
			 JOIN edges e ON e.target = n.id
			 JOIN edges e2 ON e2.source = n.id AND e2.relationship LIKE 'USES_%'
			 WHERE e.source = ?) AS asset_count
	`, sceneID, sceneID, sceneID, sceneID, sceneID).Scan(
		&stats.GameObjectCount,
		&stats.ComponentCount,
		&stats.ScriptCount,
		&stats.PrefabCount,
		&stats.AssetCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics for scene %s: %w", sceneID, err)
	}
	return &stats, nil
}
