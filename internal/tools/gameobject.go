package tools

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// GameObjectTools provides tools for GameObjects.
type GameObjectTools struct {
	ctx *Context
}

// NewGameObjectTools creates a new GameObjectTools instance.
func NewGameObjectTools(ctx *Context) *GameObjectTools {
	return &GameObjectTools{ctx: ctx}
}

// Meta implements ToolMetaProvider.
func (t *GameObjectTools) Meta() ToolMetadata {
	return ToolMetadata{
		Name:        "find_gameobjects",
		Description: "Finds GameObjects by name (case‑insensitive partial match) within a scene.",
		Parameters: []Parameter{
			{Name: "scene_id", Type: "string", Description: "Scene ID or GUID", Required: true},
			{Name: "name", Type: "string", Description: "Name to search for (substring match)", Required: true},
		},
	}
}

// ---------- FindGameObjects ----------

// FindGameObjectsRequest is the input for FindGameObjects.
type FindGameObjectsRequest struct {
	SceneID string `json:"scene_id"`
	Name    string `json:"name"`
}

// FindGameObjectsResponse is the output for FindGameObjects.
type FindGameObjectsResponse struct {
	GameObjects []*models.GameObject `json:"game_objects"`
}

// FindGameObjects finds GameObjects by name (substring match) within a scene.
func (t *GameObjectTools) FindGameObjects(req FindGameObjectsRequest) (FindGameObjectsResponse, error) {
	if req.SceneID == "" {
		return FindGameObjectsResponse{}, fmt.Errorf("find_gameobjects: scene_id is required")
	}
	if req.Name == "" {
		return FindGameObjectsResponse{}, fmt.Errorf("find_gameobjects: name is required")
	}

	gobjs, err := t.ctx.Query.FindGameObjectsByName(req.SceneID, req.Name)
	if err != nil {
		return FindGameObjectsResponse{}, fmt.Errorf("find_gameobjects: %w", err)
	}
	return FindGameObjectsResponse{GameObjects: gobjs}, nil
}

// ---------- GetGameObjectHierarchy ----------

// GetGameObjectHierarchyResponse is the output for GetGameObjectHierarchy.
type GetGameObjectHierarchyResponse struct {
	GameObject *models.GameObject   `json:"gameobject"`
	Children   []*models.GameObject `json:"children"`
	Parents    []*models.GameObject `json:"parents"`
	Warnings   []string             `json:"warnings,omitempty"`
}

// GetGameObjectHierarchy correctly handles both CONTAINS (root) and CHILD_OF (nested) edges.
// Per Atlas graph semantics: Scene→GO uses CONTAINS, GO→GO uses CHILD_OF.
func (t *GameObjectTools) GetGameObjectHierarchy(gameObjectID string) (GetGameObjectHierarchyResponse, error) {
	if gameObjectID == "" {
		return GetGameObjectHierarchyResponse{}, fmt.Errorf("get_hierarchy: gameobject_id is required")
	}

	gobj, err := t.ctx.Query.GetGameObjectByID(gameObjectID)
	if err != nil {
		return GetGameObjectHierarchyResponse{}, fmt.Errorf("get_hierarchy: failed to get GameObject: %w", err)
	}
	if gobj == nil {
		return GetGameObjectHierarchyResponse{}, fmt.Errorf("get_hierarchy: GameObject not found: %s", gameObjectID)
	}

	var warnings []string

	// Try CHILD_OF first (nested GameObjects)
	childIDs, err := t.ctx.Query.WalkChildren(gameObjectID)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to get CHILD_OF children: %v", err))
		childIDs = []string{}
	}

	// If no CHILD_OF children found, try CONTAINS (root-level GameObjects under Scene)
	if len(childIDs) == 0 {
		neighbors, nErr := t.ctx.Query.GetNeighbors(gameObjectID, models.EdgeContains)
		if nErr == nil && len(neighbors) > 0 {
			childIDs = neighbors
		}
	}

	children := make([]*models.GameObject, 0, len(childIDs))
	for _, cid := range childIDs {
		c, cErr := t.ctx.Query.GetGameObjectByID(cid)
		if cErr == nil && c != nil {
			children = append(children, c)
		}
	}

	parentIDs, err := t.ctx.Query.WalkParents(gameObjectID)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("failed to get parents: %v", err))
		parentIDs = []string{}
	}
	parents := make([]*models.GameObject, 0, len(parentIDs))
	for _, pid := range parentIDs {
		p, pErr := t.ctx.Query.GetGameObjectByID(pid)
		if pErr == nil && p != nil {
			parents = append(parents, p)
		}
	}

	return GetGameObjectHierarchyResponse{
		GameObject: gobj,
		Children:   children,
		Parents:    parents,
		Warnings:   warnings,
	}, nil
}
