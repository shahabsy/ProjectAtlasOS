package graph

import "fmt"

// PrefabQueries provides queries for prefabs.
type PrefabQueries struct {
	repo *Repository
}

func NewPrefabQueries(repo *Repository) *PrefabQueries {
	return &PrefabQueries{repo: repo}
}

// List returns all prefabs in the project.
func (q *PrefabQueries) List() ([]PrefabInfo, error) {
	rows, err := q.repo.Query(`
		SELECT id, guid, name, path
		FROM nodes WHERE type = 'prefab'
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list prefabs: %w", err)
	}
	defer rows.Close()

	var prefabs []PrefabInfo
	for rows.Next() {
		var p PrefabInfo
		if err := rows.Scan(&p.ID, &p.GUID, &p.Name, &p.Path); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		prefabs = append(prefabs, p)
	}
	return prefabs, nil
}

// ByGUID returns a prefab by its asset GUID.
func (q *PrefabQueries) ByGUID(guid string) (*PrefabInfo, error) {
	var p PrefabInfo
	err := q.repo.QueryRow(`
		SELECT id, name, path
		FROM nodes WHERE guid = ? AND type = 'prefab'
	`, guid).Scan(&p.ID, &p.Name, &p.Path)
	if err != nil {
		return nil, fmt.Errorf("prefab with GUID '%s' not found: %w", guid, err)
	}
	p.GUID = guid
	return &p, nil
}

// Instances returns all GameObjects that are instances of a prefab.
func (q *PrefabQueries) Instances(prefabID string) ([]GameObjectInfo, error) {
	rows, err := q.repo.Query(`
		SELECT g.id, g.global_id, g.name, e2.source
		FROM nodes g
		JOIN edges e1 ON e1.source = g.id AND e1.relationship = 'INSTANCE_OF'
		JOIN nodes p ON e1.target = p.id AND p.type = 'prefab'
		JOIN edges e2 ON e2.target = g.id AND e2.relationship = 'CONTAINS'
		WHERE p.id = ? AND g.type = 'gameobject'
	`, prefabID)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances of prefab %s: %w", prefabID, err)
	}
	defer rows.Close()

	var gos []GameObjectInfo
	for rows.Next() {
		var gobj GameObjectInfo
		if err := rows.Scan(&gobj.ID, &gobj.GlobalId, &gobj.Name, &gobj.SceneID); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		gos = append(gos, gobj)
	}
	return gos, nil
}
