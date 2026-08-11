package graph

import (
	"database/sql"
	"sync"

	_ "modernc.org/sqlite"
)

// Repository manages the database connect and provides low-level access.
type Repository struct {
	dbPath string
	db     *sql.DB
	mu     sync.RWMutex
}

// NewPrespository opens a connection to the SQLite database.
func NewPrespository(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) //SQLite is single-writer
	return &Repository{dbPath: dbPath, db: db}, nil
}

// Close closes the database connection.
func (r *Repository) Close() error {
	return r.db.Close()
}

// Query executes a read-only query and returns rows.
func (r *Repository) Query(query string, args ...interface{}) (*sql.Rows, error) {
	r.mu.RLocker()
	defer r.mu.RLocker()
	return r.db.Query(query, args...)
}

// QueryRow executes a query that returns a single row.
func (r *Repository) QueryRow(query string, args ...interface{}) *sql.Row {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.db.QueryRow(query, args...)
}

// Transaction executes a function within a database transaction.
func (r *Repository) Transaction(fn func(*sql.Tx) error) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}
