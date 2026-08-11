package graph

import (
	"database/sql"
	"fmt"
)

// AssetQueries provides queries for assets (materials, shaders, textures).
type AssetQueries struct {
	repo *Repository
}

func NewAssetQueries(repo *Repository) *AssetQueries {
	return &AssetQueries{repo: repo}
}

// List returns all assets of a given type (or all if type is empty).
func (q *AssetQueries) List(assetType string) ([]AssetInfo, error) {
	var rows *sql.Rows
	var err error
	if assetType == "" {
		rows, err = q.repo.Query(`
			SELECT id, guid, type, name, path
			FROM nodes WHERE type IN ('material', 'shader', 'texture')
			ORDER BY type, name
		`)
	} else {
		rows, err = q.repo.Query(`
			SELECT id, guid, type, name, path
			FROM nodes WHERE type = ?
			ORDER BY name
		`, assetType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list assets: %w", err)
	}
	defer rows.Close()

	var assets []AssetInfo
	for rows.Next() {
		var a AssetInfo
		if err := rows.Scan(&a.ID, &a.GUID, &a.Type, &a.Name, &a.Path); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		assets = append(assets, a)
	}
	return assets, nil
}

// ByGUID returns an asset by its GUID.
func (q *AssetQueries) ByGUID(guid string) (*AssetInfo, error) {
	var a AssetInfo
	err := q.repo.QueryRow(`
		SELECT id, type, name, path
		FROM nodes WHERE guid = ?
	`, guid).Scan(&a.ID, &a.Type, &a.Name, &a.Path)
	if err != nil {
		return nil, fmt.Errorf("asset with GUID '%s' not found: %w", guid, err)
	}
	a.GUID = guid
	return &a, nil
}

// Usages returns all components that reference a given asset.
func (q *AssetQueries) Usages(assetID string) ([]ComponentInfo, error) {
	rows, err := q.repo.Query(`
		SELECT c.id, c.global_id, c.name, c.parent_id
		FROM nodes c
		JOIN edges e ON e.source = c.id AND e.relationship LIKE 'USES_%'
		WHERE e.target = ? AND c.type = 'component'
	`, assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to find usages of asset %s: %w", assetID, err)
	}
	defer rows.Close()

	var comps []ComponentInfo
	for rows.Next() {
		var comp ComponentInfo
		if err := rows.Scan(&comp.ID, &comp.GlobalId, &comp.Type, &comp.GameObjectID); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		comps = append(comps, comp)
	}
	return comps, nil
}
