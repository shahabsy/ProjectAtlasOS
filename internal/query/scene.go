package query

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

func (e *Engine) GetSceneByID(id string) (*models.Scene, error) {
	node, err := db.GetNodeByID(e.DBPath(), id)
	if err != nil {
		return nil, fmt.Errorf("failed to get scene by ID %s: %w", id, err)
	}

	if node == nil {
		return nil, nil
	}
	if node.Type != "scene" {
		return nil, fmt.Errorf("node %s is not a scene (type: %s)", id, node.Type)
	}
	return &models.Scene{
		ID:          node.ID,
		GUID:        node.GUID,
		Name:        node.Name,
		CreatedAt:   node.CreatedAt,
		UpdatedAt:   node.UpdatedAt,
		LastIndexed: node.LastIndexed,
	}, nil
}

func (e *Engine) GetSceneByGUID(guid string) (*models.Scene, error) {
	node, err := db.GetNodeByGUID(e.dbPath, guid)
	if err != nil {
		return nil, fmt.Errorf("failed to get scene by GUID %s: %s", guid, err)
	}
	if node == nil {
		return nil, nil
	}

	if node.Type != "scene" {
		return nil, fmt.Errorf("node with GUID %s is not a scene (type:%s)", guid, node.Type)
	}
	return &models.Scene{
		ID:          node.ID,
		GUID:        node.GUID,
		Name:        node.Name,
		CreatedAt:   node.CreatedAt,
		UpdatedAt:   node.UpdatedAt,
		LastIndexed: node.LastIndexed,
	}, nil
}

func (e *Engine) ListScenes() ([]*models.Scene, error) {
	nodes, err := db.GetNodesByType(e.dbPath, "scene")
	if err != nil {
		return nil, fmt.Errorf("failed to list scenes: %w", err)
	}
	scenes := make([]*models.Scene, 0, len(nodes))
	for _, node := range nodes {
		scenes = append(scenes, &models.Scene{
			ID:          node.ID,
			GUID:        node.GUID,
			Name:        node.Name,
			CreatedAt:   node.CreatedAt,
			UpdatedAt:   node.UpdatedAt,
			LastIndexed: node.LastIndexed,
		})
	}
	return scenes, nil
}
