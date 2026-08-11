package statistics

import (
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

func (e *Engine) SceneSummary(sceneID string) (*models.SceneStatistics, error) {
	gobjs, err := e.query.GetGameObjectsByScene(sceneID)
	if err != nil {
		return nil, err
	}
	totalGameObjects := len(gobjs)

	totalComponents := 0
	scriptsSet := make(map[string]bool)

	for _, gobj := range gobjs {
		comps, err := e.query.GetComponentsByGameObject(gobj.ID)
		if err != nil {
			return nil, err
		}
		totalComponents += len(comps)

		// Now comp.ScriptID is correctly populated
		for _, comp := range comps {
			if comp.ScriptID != "" {
				scriptsSet[comp.ScriptID] = true
			}
		}
	}

	// (Prefab and asset counts will be added later)

	return &models.SceneStatistics{
		GameObjectCount: totalGameObjects,
		ComponentCount:  totalComponents,
		ScriptCount:     len(scriptsSet),
		PrefabCount:     0, // placeholder
		AssetCount:      0,
		FieldCount:      0,
	}, nil
}
