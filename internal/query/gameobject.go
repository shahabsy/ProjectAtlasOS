package query

import (
	"fmt"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// GetGameObjectsByScene retrieves all GameObjects in a given scene.
func (e *Engine) GetGameObjectsByScene(sceneID string) ([]*models.GameObject, error) {
	nodes, err := db.GetGameObjectsByScene(e.dbPath, sceneID)
	if err != nil {
		return nil, fmt.Errorf("failed to get GameObjects for scene %s: %w", sceneID, err)
	}
	result := make([]*models.GameObject, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, &models.GameObject{
			ID:       n.ID,
			GlobalID: n.GUID,
			Name:     n.Name,
			SceneID:  sceneID,
		})
	}
	return result, nil
}

// FindGameObjectsByName finds GameObjects by name (case‑insensitive substring match).
func (e *Engine) FindGameObjectsByName(sceneID, name string) ([]*models.GameObject, error) {
	all, err := e.GetGameObjectsByScene(sceneID)
	if err != nil {
		return nil, err
	}
	var result []*models.GameObject
	lowerName := strings.ToLower(name)
	for _, gobj := range all {
		if strings.Contains(strings.ToLower(gobj.Name), lowerName) {
			result = append(result, gobj)
		}
	}
	return result, nil
}

// GetGameObjectByID retrieves a single GameObject by its node ID.
func (e *Engine) GetGameObjectByID(id string) (*models.GameObject, error) {
	node, err := db.GetNodeByID(e.dbPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get GameObject by ID %s: %w", id, err)
	}
	if node == nil {
		return nil, nil
	}
	if node.Type != "gameobject" {
		return nil, fmt.Errorf("node %s is not a GameObject (type: %s)", id, node.Type)
	}
	return &models.GameObject{
		ID:       node.ID,
		GlobalID: node.GUID,
		Name:     node.Name,
	}, nil
}
