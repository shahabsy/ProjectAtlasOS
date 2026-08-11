package query

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// GetComponentsByGameObject retrieves all components of a GameObject.
func (e *Engine) GetComponentsByGameObject(gobjID string) ([]*models.Component, error) {
	nodes, err := db.GetComponentsByGameObject(e.dbPath, gobjID)
	if err != nil {
		return nil, fmt.Errorf("failed to get components for GameObject %s: %w", gobjID, err)
	}
	result := make([]*models.Component, 0, len(nodes))
	for _, n := range nodes {
		comp := &models.Component{
			ID:           n.ID,
			GlobalID:     n.GUID,
			Type:         n.Type,
			GameObjectID: gobjID,
			ScriptID:     n.Metadata, // because we stored the script ID in Metadata
		}
		// If Metadata is empty, no script
		if comp.ScriptID == "" {
			comp.ScriptID = ""
		}
		result = append(result, comp)
	}
	return result, nil
}

// GetComponentByID retrieves a single component by its node ID.
func (e *Engine) GetComponentByID(id string) (*models.Component, error) {
	node, err := db.GetNodeByID(e.dbPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get component by ID %s: %w", id, err)
	}
	if node == nil {
		return nil, nil
	}
	if node.Type != "component" {
		return nil, fmt.Errorf("node %s is not a component (type: %s)", id, node.Type)
	}
	// Fetch script ID for this component
	edges, err := db.GetEdgesBySource(e.dbPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get edges for component: %w", err)
	}
	var scriptID string
	for _, edge := range edges {
		if edge.Relationship == models.EdgeUsesScript {
			scriptID = edge.TargetID
			break
		}
	}
	return &models.Component{
		ID:       node.ID,
		GlobalID: node.GUID,
		Type:     node.Type,
		ScriptID: scriptID,
	}, nil
}

func (e *Engine) GetFieldsByComponent(compID string) ([]*models.SerializedField, error) {
	// We need to query edges: HAS_FIELD from compID to field nodes.
	// For now, we'll return an empty slice.
	return []*models.SerializedField{}, nil
}
