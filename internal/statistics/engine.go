package statistics

import (
	"github.com/shahabsy/ProjectAtlasOS/internal/query"
)

// Engine provides aggregated statistics about the project and scenes.
type Engine struct {
	query *query.Engine
}

// NewEngine creates a new statistics engine using the given query engine.
func NewEngine(q *query.Engine) *Engine {
	return &Engine{query: q}
}
