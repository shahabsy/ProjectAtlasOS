package tools

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// AssetTools provides tools for assets.
type AssetTools struct {
	ctx *Context
}

// NewAssetTools creates a new AssetTools instance.
func NewAssetTools(ctx *Context) *AssetTools {
	return &AssetTools{ctx: ctx}
}

// Meta implements ToolMetaProvider.
func (t *AssetTools) Meta() ToolMetadata {
	return ToolMetadata{
		Name:        "list_assets",
		Description: "Returns all assets in the project.",
		Parameters:  []Parameter{},
	}
}

// ---------- ListAssets ----------

// ListAssetsRequest is the input for ListAssets.
type ListAssetsRequest struct{}

// ListAssetsResponse is the output for ListAssets.
type ListAssetsResponse struct {
	Assets []*models.Asset `json:"assets"`
}

// ListAssets returns all assets in the project.
func (t *AssetTools) ListAssets(req ListAssetsRequest) (ListAssetsResponse, error) {
	assets, err := t.ctx.Query.ListAssets()
	if err != nil {
		return ListAssetsResponse{}, fmt.Errorf("list_assets: %w", err)
	}
	return ListAssetsResponse{Assets: assets}, nil
}

// ---------- GetAssetDependencies (stub) ----------

// GetAssetDependenciesRequest is the input for GetAssetDependencies.
type GetAssetDependenciesRequest struct {
	AssetID string `json:"asset_id"`
}

// GetAssetDependenciesResponse is the output for GetAssetDependencies.
type GetAssetDependenciesResponse struct {
	Dependencies []string `json:"dependencies"`
}

// GetAssetDependencies returns all nodes that depend on the given asset.
// Currently a stub – will be implemented when dependency traversal is complete.
func (t *AssetTools) GetAssetDependencies(req GetAssetDependenciesRequest) (GetAssetDependenciesResponse, error) {
	if req.AssetID == "" {
		return GetAssetDependenciesResponse{}, fmt.Errorf("get_asset_dependencies: asset_id is required")
	}
	// TODO: Implement using traversal to find incoming edges.
	return GetAssetDependenciesResponse{Dependencies: []string{}}, nil
}
