package query

import (
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// GetSceneGameObjects returns all GameObjects directly in a scene (via CONTAINS).
func (e *Engine) GetSceneGameObjects(sceneID string) ([]*models.GameObject, error) {
	return e.GetGameObjectsByScene(sceneID)
}

// GetGameObjectChildren returns the immediate children of a GameObject (via CHILD_OF).
func (e *Engine) GetGameObjectChildren(gameObjectID string) ([]*models.GameObject, error) {
	childIDs, err := e.WalkChildren(gameObjectID)
	if err != nil {
		return nil, err
	}
	var children []*models.GameObject
	for _, id := range childIDs {
		child, err := e.GetGameObjectByID(id)
		if err != nil {
			return nil, err
		}
		if child != nil {
			children = append(children, child)
		}
	}
	return children, nil
}

// GetComponentFields returns the serialized fields of a component.
// (We'll implement this properly later.)
func (e *Engine) GetComponentFields(componentID string) ([]*models.SerializedField, error) {
	// Placeholder for future implementation.
	return nil, nil
}
