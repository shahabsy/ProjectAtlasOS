package db

import (
	"database/sql"
	"fmt"
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

// RunMigrations ensures the database is at the latest version.
func RunMigrations(dbPath string) error {
	current := GetSchemaVersion(dbPath)
	if current < 2 {
		// Migration to version 2: add new columns and tables
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			return err
		}
		defer db.Close()

		// Add columns to nodes if not exist (idempotent)
		_, err = db.Exec(`ALTER TABLE nodes ADD COLUMN global_id TEXT`)
		if err != nil {
			// ignore if already exists
		}
		_, err = db.Exec(`ALTER TABLE nodes ADD COLUMN path TEXT`)
		if err != nil {
			// ignore
		}
		_, err = db.Exec(`ALTER TABLE nodes ADD COLUMN metadata TEXT`)
		if err != nil {
			// ignore
		}
		_, err = db.Exec(`ALTER TABLE edges ADD COLUMN metadata TEXT`)
		if err != nil {
			// ignore
		}

		// New indexes
		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_nodes_global_id ON nodes(global_id)`)
		if err != nil {
			return err
		}
		_, err = db.Exec(`CREATE INDEX IF NOT EXISTS idx_edges_relationship ON edges(relationship)`)
		if err != nil {
			return err
		}

		// Set version to 2
		if err := SetSchemaVersion(dbPath, 2); err != nil {
			return err
		}
	}
	return nil
}
