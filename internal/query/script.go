package query

import (
	"fmt"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// GetScriptByID retrieves a script by its node ID.
func (e *Engine) GetScriptByID(id string) (*models.Script, error) {
	node, err := db.GetNodeByID(e.dbPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get script by ID %s: %w", id, err)
	}
	if node == nil {
		return nil, nil
	}
	if node.Type != "script" {
		return nil, fmt.Errorf("node %s is not a script (type: %s)", id, node.Type)
	}
	return &models.Script{
		ID:   node.ID,
		GUID: node.GUID,
		Name: node.Name,
	}, nil
}

// GetScriptsByName finds scripts by name (case‑insensitive substring match).
func (e *Engine) GetScriptsByName(name string) ([]*models.Script, error) {
	nodes, err := db.GetNodesByType(e.dbPath, "script")
	if err != nil {
		return nil, fmt.Errorf("failed to search scripts: %w", err)
	}
	var result []*models.Script
	lowerName := strings.ToLower(name)
	for _, n := range nodes {
		if strings.Contains(strings.ToLower(n.Name), lowerName) {
			result = append(result, &models.Script{
				ID:   n.ID,
				GUID: n.GUID,
				Name: n.Name,
			})
		}
	}
	return result, nil
}

// ListScripts returns all scripts in the project.
func (e *Engine) ListScripts() ([]*models.Script, error) {
	nodes, err := db.GetNodesByType(e.dbPath, "script")
	if err != nil {
		return nil, fmt.Errorf("failed to list scripts: %w", err)
	}
	result := make([]*models.Script, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, &models.Script{
			ID:   n.ID,
			GUID: n.GUID,
			Name: n.Name,
		})
	}
	return result, nil
}

// GetGameObjectsUsingScript finds all GameObjects that have a component using the given script.
func (e *Engine) GetGameObjectsUsingScript(scriptID string) ([]*models.GameObject, error) {
	// Step 1: Find all components that have a USES_SCRIPT edge to this script.
	edges, err := db.GetEdgesByTarget(e.dbPath, scriptID)
	if err != nil {
		return nil, fmt.Errorf("failed to find USES_SCRIPT edges: %w", err)
	}
	var componentIDs []string
	for _, edge := range edges {
		if edge.Relationship == models.EdgeUsesScript {
			componentIDs = append(componentIDs, edge.SourceID)
		}
	}
	if len(componentIDs) == 0 {
		return []*models.GameObject{}, nil
	}

	// Step 2: For each component, find the GameObject that owns it.
	var gameObjectIDs []string
	for _, compID := range componentIDs {
		edges, err := db.GetEdgesByTarget(e.dbPath, compID)
		if err != nil {
			return nil, fmt.Errorf("failed to find HAS_COMPONENT edges for component %s: %w", compID, err)
		}
		for _, edge := range edges {
			if edge.Relationship == models.EdgeHasComponent {
				gameObjectIDs = append(gameObjectIDs, edge.SourceID)
				break
			}
		}
	}

	// Step 3: Fetch the GameObject nodes.
	var result []*models.GameObject
	for _, gobjID := range gameObjectIDs {
		gobj, err := e.GetGameObjectByID(gobjID)
		if err != nil {
			return nil, fmt.Errorf("failed to get GameObject %s: %w", gobjID, err)
		}
		if gobj != nil {
			result = append(result, gobj)
		}
	}
	return result, nil
}
