package graph

import (
	"fmt"
)

type ScriptQueries struct {
	repo *Repository
}

func NewScriptQueries(repo *Repository) *ScriptQueries {
	return &ScriptQueries{repo: repo}
}

// ListByScene returns all scripts used in a scene.
func (q *ScriptQueries) ListByScene(sceneID string) ([]ScriptInfo, error) {
	rows, err := q.repo.Query(`
		SELECT DISTINCT s.id, s.guid, s.name, s.path
		FROM nodes s
		JOIN edges e1 ON e1.target = s.id AND e1.relationship = 'USES_SCRIPT'
		JOIN nodes c ON e1.source = c.id AND c.type = 'component'
		JOIN edges e2 ON e2.target = c.id AND e2.relationship = 'HAS_COMPONENT'
		JOIN nodes g ON e2.source = g.id AND g.type = 'gameobject'
		JOIN edges e3 ON e3.target = g.id AND e3.relationship = 'CONTAINS'
		WHERE e3.source = ? AND s.type = 'script'
		ORDER BY s.name
	`, sceneID)
	if err != nil {
		return nil, fmt.Errorf("failed to list scripts for scene %s: %w", sceneID, err)
	}
	defer rows.Close()

	var scripts []ScriptInfo
	for rows.Next() {
		var s ScriptInfo
		if err := rows.Scan(&s.ID, &s.GUID, &s.Name, &s.Path); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		// Class name is stored in the 'name' column
		s.ClassName = s.Name
		s.Namespace = "" // Not stored in current schema
		scripts = append(scripts, s)
	}
	return scripts, nil
}

// ByGUID returns a script by its asset GUID.
func (q *ScriptQueries) ByGUID(guid string) (*ScriptInfo, error) {
	var s ScriptInfo
	err := q.repo.QueryRow(`
		SELECT id, name, path
		FROM nodes WHERE guid = ? AND type = 'script'
	`, guid).Scan(&s.ID, &s.Name, &s.Path)
	if err != nil {
		return nil, fmt.Errorf("script with GUID '%s' not found: %w", guid, err)
	}
	s.GUID = guid
	s.ClassName = s.Name
	s.Namespace = ""
	return &s, nil
}

// Usages returns all GameObjects that use a given script.
func (q *ScriptQueries) Usages(scriptID string) ([]GameObjectInfo, error) {
	rows, err := q.repo.Query(`
		SELECT g.id, g.global_id, g.name, e3.source
		FROM nodes g
		JOIN edges e3 ON e3.target = g.id AND e3.relationship = 'CONTAINS'
		JOIN nodes c ON e3.source = c.id AND c.type = 'component'
		JOIN edges e2 ON e2.target = c.id AND e2.relationship = 'HAS_COMPONENT'
		JOIN nodes s ON e2.source = s.id AND s.type = 'script'
		WHERE s.id = ? AND g.type = 'gameobject'
		GROUP BY g.id
		ORDER BY g.name
	`, scriptID)
	if err != nil {
		return nil, fmt.Errorf("failed to find usages of script %s: %w", scriptID, err)
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
