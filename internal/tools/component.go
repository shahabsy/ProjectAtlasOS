package tools

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// ComponentTools provides tools for components.
type ComponentTools struct {
	ctx *Context
}

// NewComponentTools creates a new ComponentTools instance.
func NewComponentTools(ctx *Context) *ComponentTools {
	return &ComponentTools{ctx: ctx}
}

// Meta implements ToolMetaProvider.
func (t *ComponentTools) Meta() ToolMetadata {
	return ToolMetadata{
		Name:        "get_components",
		Description: "Returns all components attached to a GameObject.",
		Parameters: []Parameter{
			{Name: "gameobject_id", Type: "string", Description: "GameObject ID", Required: true},
		},
	}
}

// ---------- GetComponents ----------

// GetComponentsRequest is the input for GetComponents.
type GetComponentsRequest struct {
	GameObjectID string `json:"gameobject_id"`
}

// GetComponentsResponse is the output for GetComponents.
type GetComponentsResponse struct {
	Components []*models.Component `json:"components"`
}

// GetComponents returns all components of a GameObject.
func (t *ComponentTools) GetComponents(req GetComponentsRequest) (GetComponentsResponse, error) {
	if req.GameObjectID == "" {
		return GetComponentsResponse{}, fmt.Errorf("get_components: gameobject_id is required")
	}

	comps, err := t.ctx.Query.GetComponentsByGameObject(req.GameObjectID)
	if err != nil {
		return GetComponentsResponse{}, fmt.Errorf("get_components: %w", err)
	}
	return GetComponentsResponse{Components: comps}, nil
}

// ---------- GetFields (stub) ----------

// GetFieldsRequest is the input for GetFields.
type GetFieldsRequest struct {
	ComponentID string `json:"component_id"`
}

// GetFieldsResponse is the output for GetFields.
type GetFieldsResponse struct {
	Fields []*models.SerializedField `json:"fields"`
}

// GetFields returns the serialized fields of a component.
// Currently a stub – will be implemented when GetFieldsByComponent is added to Query Engine.
func (t *ComponentTools) GetFields(req GetFieldsRequest) (GetFieldsResponse, error) {
	if req.ComponentID == "" {
		return GetFieldsResponse{}, fmt.Errorf("get_fields: component_id is required")
	}
	// TODO: Implement when query.GetFieldsByComponent is available.
	return GetFieldsResponse{Fields: []*models.SerializedField{}}, nil
}
