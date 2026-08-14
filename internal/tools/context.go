package tools

import (
	"github.com/shahabsy/ProjectAtlasOS/internal/query"
	"github.com/shahabsy/ProjectAtlasOS/internal/statistics"
)

// Context holds shared engine references for all tools.
// Engines are injected, not created, to allow connection reuse.
type Context struct {
	Query      *query.Engine
	Statistics *statistics.Engine
}

// NewContext creates a tool context from existing engines.
func NewContext(q *query.Engine, s *statistics.Engine) *Context {
	return &Context{Query: q, Statistics: s}
}
