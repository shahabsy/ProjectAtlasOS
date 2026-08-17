package query

import (
	"sync"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

type Engine struct {
	dbPath string
	mu     sync.RWMutex
}

func NewEngine(dbPath string) *Engine {
	return &Engine{
		dbPath: dbPath,
	}
}

func (e *Engine) DBPath() string {
	return e.dbPath
}

// GetScriptsByScene returns all unique Scripts used in components within the given scene.
func (e *Engine) GetScriptsByScene(sceneID string) ([]*models.Script, error) {
	nodes, err := db.GetScriptsByScene(e.dbPath, sceneID)
	if err != nil {
		return nil, err
	}
	scripts := make([]*models.Script, 0, len(nodes))
	for _, n := range nodes {
		scripts = append(scripts, &models.Script{
			ID:   n.ID,
			Name: n.Name,
			GUID: n.GUID,
			// If your Script model has a Path or other fields, set them accordingly.
			// Example: Path: n.Path, etc.
		})
	}
	return scripts, nil
}
