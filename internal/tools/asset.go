package tools

import (
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

type AssetTools struct{ ctx *Context }

func NewAssetTools(ctx *Context) *AssetTools { return &AssetTools{ctx: ctx} }

func (t *AssetTools) Contract() Contract {
	return Contract{
		Name:        "asset_tools",
		Description: "Tools for querying Assets.",
		InputSchema: Schema{
			Type:       "object",
			Properties: map[string]Property{},
			Required:   []string{},
		},
		ReadOnly: true,
	}
}

// ListAssetsResponse is the output for ListAssets.
type ListAssetsResponse struct {
	Assets []*models.Asset `json:"assets"`
}

// ListAssets returns all assets in the project.
func (t *AssetTools) ListAssets() Result {
	assets, err := t.ctx.Query.ListAssets()
	if err != nil {
		return NewErrorResult("QUERY_FAILED", err.Error())
	}
	return NewSuccessResult(ListAssetsResponse{Assets: assets})
}
