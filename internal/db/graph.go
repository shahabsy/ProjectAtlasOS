package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Schema content is raed via //go:embed in a separate file to avoid issues.
// For simplicity here, we define it inline.
const schemaContent = `
CREATE TABLE IF NOT EXISTS nodes (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	guid TEXT UNIQUE,
	name TEXT,
	json TEXT,
	created_at INTEGER DEFAULT (strftime('%s', 'now')),
	updated_at INTEGER DEFAULT (strftime('%s', 'now'))
);
CREATE TABLE IF NOT EXISTS edges (
	source TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	target TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	relationship TEXT NOT NULL,
	properties JSON,
	created_at INTEGER DEFAULT (strftime('%s', 'now')),
	PRIMARY KEY (source, target, relationship)
);
CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);
CREATE INDEX IF NOT EXISTS idx_nodes_guid ON nodes(guid);
CREATE INDEX IF NOT EXISTS idx_edges_source ON edges(source);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges(target);
`

func CreateDB(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(schemaContent)
	return err
}
