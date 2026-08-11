package graph

import (
	"database/sql"
	"fmt"
)

// ComponentQueries provides queries for components.
type ComponentQueries struct {
	repo *Repository
}

func NewComponentQueries(repo *Repository) *ComponentQueries {
	return &ComponentQueries{repo: repo}
}

// ListByGameObject returns all components attached to a GameObject.
func (q *ComponentQueries) ListByGameObject(gameObjectID string) ([]ComponentInfo, error) {
	rows, err := q.repo.Query(`
		SELECT n.id, n.global_id, n.name, n.enabled, s.id
		FROM nodes n
		LEFT JOIN edges e ON e.source = n.id AND e.relationship = 'USES_SCRIPT'
		LEFT JOIN nodes s ON e.target = s.id AND s.type = 'script'
		WHERE n.parent_id = ? AND n.type = 'component'
		ORDER BY n.name
	`, gameObjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list components for GameObject %s: %w", gameObjectID, err)
	}
	defer rows.Close()

	var comps []ComponentInfo
	for rows.Next() {
		var comp ComponentInfo
		var scriptID sql.NullString
		if err := rows.Scan(&comp.ID, &comp.GlobalId, &comp.Type, &comp.Enabled, &scriptID); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		if scriptID.Valid {
			comp.ScriptID = scriptID.String
		}
		comp.GameObjectID = gameObjectID
		comps = append(comps, comp)
	}
	return comps, nil
}

// ByGlobalID returns a component by its GlobalObjectId.
func (q *ComponentQueries) ByGlobalID(globalID string) (*ComponentInfo, error) {
	var comp ComponentInfo
	var scriptID sql.NullString
	err := q.repo.QueryRow(`
		SELECT n.id, n.name, n.enabled, n.parent_id, s.id
		FROM nodes n
		LEFT JOIN edges e ON e.source = n.id AND e.relationship = 'USES_SCRIPT'
		LEFT JOIN nodes s ON e.target = s.id AND s.type = 'script'
		WHERE n.global_id = ? AND n.type = 'component'
	`, globalID).Scan(&comp.ID, &comp.Type, &comp.Enabled, &comp.GameObjectID, &scriptID)
	if err != nil {
		return nil, fmt.Errorf("component with GlobalID '%s' not found: %w", globalID, err)
	}
	if scriptID.Valid {
		comp.ScriptID = scriptID.String
	}
	comp.GlobalId = globalID
	return &comp, nil
}
