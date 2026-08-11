package graph

import (
	"fmt"
)

type GameObjectQueries struct {
	repo *Repository
}

func NewGameObjectQueries(repo *Repository) *GameObjectQueries {
	return &GameObjectQueries{repo: repo}
}

// ListByScene returns all GameObjects belonging to a scene.
// Parent ID is fetched via a LEFT JOIN on the CHILD_OF edge.
func (q *GameObjectQueries) ListByScene(sceneID string) ([]GameObjectInfo, error) {
	rows, err := q.repo.Query(`
		SELECT n.id, n.global_id, n.name, COALESCE(e2.source, '') as parent_id
		FROM nodes n
		JOIN edges e1 ON e1.target = n.id AND e1.relationship = 'CONTAINS'
		LEFT JOIN edges e2 ON e2.target = n.id AND e2.relationship = 'CHILD_OF'
		WHERE e1.source = ? AND n.type = 'gameobject'
		ORDER BY n.name
	`, sceneID)
	if err != nil {
		return nil, fmt.Errorf("failed to list GameObjects for scene %s: %w", sceneID, err)
	}
	defer rows.Close()

	var gos []GameObjectInfo
	for rows.Next() {
		var gobj GameObjectInfo
		if err := rows.Scan(&gobj.ID, &gobj.GlobalId, &gobj.Name, &gobj.ParentID); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		gobj.SceneID = sceneID
		gos = append(gos, gobj)
	}
	return gos, nil
}

// ByNameInScene returns a GameObject by name within a scene.
func (q *GameObjectQueries) ByNameInScene(sceneID, name string) (*GameObjectInfo, error) {
	var gobj GameObjectInfo
	err := q.repo.QueryRow(`
		SELECT n.id, n.global_id, n.name, COALESCE(e2.source, '') as parent_id
		FROM nodes n
		JOIN edges e1 ON e1.target = n.id AND e1.relationship = 'CONTAINS'
		LEFT JOIN edges e2 ON e2.target = n.id AND e2.relationship = 'CHILD_OF'
		WHERE e1.source = ? AND n.type = 'gameobject' AND n.name = ?
	`, sceneID, name).Scan(&gobj.ID, &gobj.GlobalId, &gobj.Name, &gobj.ParentID)
	if err != nil {
		return nil, fmt.Errorf("GameObject '%s' not found in scene %s: %w", name, sceneID, err)
	}
	gobj.SceneID = sceneID
	return &gobj, nil
}

// ByGlobalID returns a GameObject by its GlobalObjectId.
func (q *GameObjectQueries) ByGlobalID(globalID string) (*GameObjectInfo, error) {
	var gobj GameObjectInfo
	err := q.repo.QueryRow(`
		SELECT n.id, n.name, COALESCE(e2.source, '') as parent_id, e1.source as scene_id
		FROM nodes n
		JOIN edges e1 ON e1.target = n.id AND e1.relationship = 'CONTAINS'
		LEFT JOIN edges e2 ON e2.target = n.id AND e2.relationship = 'CHILD_OF'
		WHERE n.global_id = ? AND n.type = 'gameobject'
	`, globalID).Scan(&gobj.ID, &gobj.Name, &gobj.ParentID, &gobj.SceneID)
	if err != nil {
		return nil, fmt.Errorf("GameObject with GlobalID '%s' not found: %w", globalID, err)
	}
	gobj.GlobalId = globalID
	return &gobj, nil
}
