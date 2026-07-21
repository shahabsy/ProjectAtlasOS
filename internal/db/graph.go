package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

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
	created_at INTEGER,
	updated_at INTEGER
);
CREATE TABLE IF NOT EXISTS edges (
	source TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	target TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	relationship TEXT NOT NULL,
	properties JSON,
	created_at INTEGER,
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

func GetDBPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".atlas", "graph.db")
}

// Insert Node adds or updates a node in the graph
func InsertNode(dbPath, id, typ, guid, name string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	now := time.Now().Unix()

	_, err = db.Exec(`
	INSERT OR REPLACE INTO nodes (id, type, guid, name, json, created_at, updated_at)
	VALUES (?, ?, ?, ?, '{}', ?, ?)`, id, typ, guid, name, now, now)
	return err
}

// InsertEdge adds a relationship edge
func InsertEdge(dbPath, srcID, tgtID, rel string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	now := time.Now().Unix()

	_, err = db.Exec(`
	INSERT OR REPLACE INTO edges (source, target, relationship, properties, created_at)
	VALUES (?, ?, ?, '{}', ?)`, srcID, tgtID, rel, now)
	return err
}

// GetNodeByGUID retrieves a node by its Unity GUID
func GetNodeByGUID(dbPath, guid string) (*Node, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var id, typ, name string
	err = db.QueryRow(`SELECT id, type, name FROM nodes WHERE guid = ?`, guid).
		Scan(&id, &typ, &name)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &Node{ID: id, Type: typ, GUID: guid, Name: name}, nil
}

func CountNodes(dbPath string) (int, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, err
	}
	defer db.Close()
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count)
	return count, err
}

func GetNodesByType(dbPath, nodeType string) ([]Node, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`SELECT id, type, guid, name, created_at FROM nodes WHERE type = ? ORDER BY created_at DESC`, nodeType)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		var createdAt int64
		if err := rows.Scan(&n.ID, &n.Type, &n.GUID, &n.Name, &createdAt); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		n.CreatedAt = createdAt
		nodes = append(nodes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

func GetSceneCount(dbPath string) (int, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM nodes WHERE type = 'scene'`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func GetGameObjectsByScene(dbPath, sceneNodeID string) ([]Node, error) {
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	rows, err := database.Query(`
		SELECT n.id, n.type, n.guid, n.name
		FROM nodes n
		JOIN edges e ON e.target = n.id WHERE e.source = ? AND e.relationship = 'CONTAINS' AND n.type = 'gameobject'
		ORDER BY n.name`, sceneNodeID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Type, &n.GUID, &n.Name); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		nodes = append(nodes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

// GetComponentsByGameObject retrieves all component nodes for a GameObject by node ID
func GetComponentsByGameObject(dbPath, gameObjNodeID string) ([]Node, error) {
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	rows, err := database.Query(`
		SELECT n.id, n.type, n.guid, n.name
		FROM nodes n
		JOIN edges e ON e.target = n.id
		WHERE e.source = ? AND e.relationship = 'HAS_COMPONENT'
		ORDER BY n.name`, gameObjNodeID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Type, &n.GUID, &n.Name); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		nodes = append(nodes, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

type Node struct {
	ID        string
	Type      string // "scene", "prefab", etc.
	GUID      string
	Name      string
	CreatedAt int64
}
