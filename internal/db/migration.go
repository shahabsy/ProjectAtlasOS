package db

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	SchemaVersion = 3
)

// GetSchemaVersion returns the current user_version from PRAGMA.
func GetSchemaVersion(dbPath string) int {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0
	}
	defer db.Close()
	var version int
	db.QueryRow("PRAGMA user_version").Scan(&version)
	return version
}

// SetSchemaVersion updates the user_version.
func SetSchemaVersion(dbPath string, version int) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	return err
}

// RunMigrations ensures the database schema is at the latest version.
func RunMigrations(dbPath string) error {
	current := GetSchemaVersion(dbPath)
	if current >= SchemaVersion {
		return nil
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database for migration: %w", err)
	}
	defer db.Close()

	// Migration to version 2 (if not already applied)
	if current < 2 {
		// Add columns if they don't exist (idempotent)
		_, err = db.Exec(`ALTER TABLE nodes ADD COLUMN global_id TEXT`)
		if err != nil && !isDuplicateColumnError(err) {
			return fmt.Errorf("failed to add global_id column: %w", err)
		}
		_, err = db.Exec(`ALTER TABLE nodes ADD COLUMN path TEXT`)
		if err != nil && !isDuplicateColumnError(err) {
			return fmt.Errorf("failed to add path column: %w", err)
		}
		_, err = db.Exec(`ALTER TABLE nodes ADD COLUMN metadata TEXT`)
		if err != nil && !isDuplicateColumnError(err) {
			return fmt.Errorf("failed to add metadata column: %w", err)
		}
		_, err = db.Exec(`ALTER TABLE edges ADD COLUMN metadata TEXT`)
		if err != nil && !isDuplicateColumnError(err) {
			return fmt.Errorf("failed to add edges.metadata column: %w", err)
		}
		// Indexes
		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_nodes_global_id ON nodes(global_id)`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_edges_relationship ON edges(relationship)`)
		if err != nil {
			return err
		}
		current = 2
	}

	// Migration to version 3: add sub_type and composite indexes
	if current < 3 {
		_, err = db.Exec(`ALTER TABLE nodes ADD COLUMN sub_type TEXT`)
		if err != nil && !isDuplicateColumnError(err) {
			return fmt.Errorf("failed to add sub_type column: %w", err)
		}
		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_nodes_sub_type ON nodes(sub_type)`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_edges_source_rel ON edges(source, relationship)`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_edges_target_rel ON edges(target, relationship)`)
		if err != nil {
			return err
		}
		current = 3
	}

	// Set version
	if err := SetSchemaVersion(dbPath, current); err != nil {
		return fmt.Errorf("failed to set schema version: %w", err)
	}
	return nil
}

// isDuplicateColumnError checks if the error indicates a duplicate column.
func isDuplicateColumnError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate column name") ||
		strings.Contains(msg, "already exists")
}
