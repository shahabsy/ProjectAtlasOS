package query

import (
	"fmt"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// GetAssetByID retrieves an asset by its node ID.
func (e *Engine) GetAssetByID(id string) (*models.Asset, error) {
	node, err := db.GetNodeByID(e.dbPath, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset by ID %s: %w", id, err)
	}
	if node == nil {
		return nil, nil
	}
	if node.Type != "asset" {
		return nil, fmt.Errorf("node %s is not an asset (type: %s)", id, node.Type)
	}
	return &models.Asset{
		ID:   node.ID,
		GUID: node.GUID,
		Type: node.Type,
		Name: node.Name,
	}, nil
}

// GetAssetsByType retrieves all assets of a specific type (e.g., "Texture", "Material").
func (e *Engine) GetAssetsByType(assetType string) ([]*models.Asset, error) {
	nodes, err := db.GetNodesByType(e.dbPath, "asset")
	if err != nil {
		return nil, fmt.Errorf("failed to get assets by type: %w", err)
	}
	var result []*models.Asset
	for _, n := range nodes {
		// The asset subtype is stored in the node's type field, but since we query for type="asset",
		// we need to check the name or guid for hints.
		// For now, we'll filter by name if the asset type is known.
		if strings.Contains(strings.ToLower(n.Name), strings.ToLower(assetType)) {
			result = append(result, &models.Asset{
				ID:   n.ID,
				GUID: n.GUID,
				Type: n.Type,
				Name: n.Name,
			})
		}
	}
	return result, nil
}

// ListAssets returns all assets in the project.
func (e *Engine) ListAssets() ([]*models.Asset, error) {
	nodes, err := db.GetNodesByType(e.dbPath, "asset")
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}
	result := make([]*models.Asset, 0, len(nodes))
	for _, n := range nodes {
		result = append(result, &models.Asset{
			ID:   n.ID,
			GUID: n.GUID,
			Type: n.Type,
			Name: n.Name,
		})
	}
	return result, nil
}
