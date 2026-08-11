package query

import (
	"sync"
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
