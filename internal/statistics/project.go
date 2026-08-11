package statistics

import (
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// ProjectSummary returns high‑level statistics for the entire Unity project.
func (e *Engine) ProjectSummary() (*models.ProjectSummary, error) {
	scenes, err := e.query.ListScenes()
	if err != nil {
		return nil, err
	}

	totalScenes := len(scenes)
	totalGameObjects := 0
	totalComponents := 0
	// totalFields will be added later when we implement field counting.

	for _, scene := range scenes {
		gobjs, err := e.query.GetGameObjectsByScene(scene.ID)
		if err != nil {
			return nil, err
		}
		totalGameObjects += len(gobjs)

		for _, gobj := range gobjs {
			comps, err := e.query.GetComponentsByGameObject(gobj.ID)
			if err != nil {
				return nil, err
			}
			totalComponents += len(comps)
			// Field counting will be added later.
		}
	}

	scripts, err := e.query.ListScripts()
	if err != nil {
		return nil, err
	}
	totalScripts := len(scripts)

	assets, err := e.query.ListAssets()
	if err != nil {
		return nil, err
	}
	totalAssets := len(assets)

	// TODO: Add serialized field count and last indexed timestamp.

	return &models.ProjectSummary{
		TotalScenes:           totalScenes,
		TotalGameObjects:      totalGameObjects,
		TotalComponents:       totalComponents,
		TotalScripts:          totalScripts,
		TotalAssets:           totalAssets,
		UniqueScripts:         totalScripts,
		TotalSerializedFields: 0, // placeholder
		LastIndexed:           0, // placeholder
	}, nil
}
