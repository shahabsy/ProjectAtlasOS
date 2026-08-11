package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const schemaContent = `
CREATE TABLE IF NOT EXISTS nodes (
	id TEXT PRIMARY KEY,
	type TEXT NOT NULL,
	guid TEXT,
	global_id TEXT UNIQUE,
	name TEXT,
	json TEXT,
	path TEXT,
	metadata TEXT,
	created_at INTEGER,
	updated_at INTEGER,
	last_indexed_at INTEGER
);
CREATE TABLE IF NOT EXISTS edges (
	source TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	target TEXT NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
	relationship TEXT NOT NULL,
	metadata TEXT,
	created_at INTEGER,
	PRIMARY KEY (source, target, relationship)
);
CREATE INDEX IF NOT EXISTS idx_nodes_type ON nodes(type);
CREATE INDEX IF NOT EXISTS idx_nodes_global_id ON nodes(global_id);
CREATE INDEX IF NOT EXISTS idx_edges_source ON edges(source);
CREATE INDEX IF NOT EXISTS idx_edges_target ON edges(target);
CREATE INDEX IF NOT EXISTS idx_edges_relationship ON edges(relationship);
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

// ---------- Path‑based (non‑transaction) operations ----------
func InsertNode(dbPath, id, typ, guid, globalId, name, path string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	now := time.Now().Unix()
	_, err = db.Exec(`
	INSERT OR REPLACE INTO nodes (id, type, guid, global_id, name, path, json, created_at, updated_at, last_indexed_at)
	VALUES (?, ?, ?, ?, ?, ?, '{}', ?, ?, ?)`, id, typ, guid, globalId, name, path, now, now, now)
	return err
}

func InsertEdge(dbPath, srcID, tgtID, rel, metadata string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	now := time.Now().Unix()
	_, err = db.Exec(`
	INSERT OR REPLACE INTO edges (source, target, relationship, metadata, created_at)
	VALUES (?, ?, ?, ?, ?)`, srcID, tgtID, rel, metadata, now)
	return err
}

// ---------- Transaction‑based operations (renamed to avoid collision) ----------
func OpenWithBusyTimeout(dbPath string, timeoutMs int) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	// Enable foreign key constraints
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}

	// Set busy timeout
	if _, err = db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", timeoutMs)); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func Begin(conn *sql.DB) (*sql.Tx, error) {
	return conn.Begin()
}

func Rollback(tx *sql.Tx) error {
	return tx.Rollback()
}

func CommitTx(tx *sql.Tx) error {
	return tx.Commit()
}

// InsertNodeTx inserts a node using an existing transaction.
func InsertNodeTx(tx *sql.Tx, id, typ, guid, globalId, name, path string) error {
	if globalId == "" {
		globalId = id
	}
	now := time.Now().Unix()
	_, err := tx.Exec(`
		INSERT OR REPLACE INTO nodes (id, type, guid, global_id, name, path, json, created_at, updated_at, last_indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, '{}', ?, ?, ?)`,
		id, typ, guid, globalId, name, path, now, now, now)
	return err
}

// InsertEdgeTx inserts an edge using an existing transaction.
func InsertEdgeTx(tx *sql.Tx, srcID, tgtID, rel, metadata string) error {
	now := time.Now().Unix()
	_, err := tx.Exec(`
		INSERT OR REPLACE INTO edges (source, target, relationship, metadata, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		srcID, tgtID, rel, metadata, now)
	return err
}

// ---------- Query functions (err declarations fixed) ----------
func GetDBPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".atlas", "graph.db")
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

func GetNodeByGUID(dbPath, guid string) (*Node, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var n Node
	var path, metadata, jsonStr sql.NullString
	var globalID sql.NullString

	err = db.QueryRow(`
		SELECT id, type, guid, global_id, name, path, metadata, json,
		       created_at, updated_at, last_indexed_at
		FROM nodes WHERE guid = ?
	`, guid).Scan(
		&n.ID, &n.Type, &n.GUID, &globalID, &n.Name,
		&path, &metadata, &jsonStr,
		&n.CreatedAt, &n.UpdatedAt, &n.LastIndexed,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	n.GlobalID = globalID.String
	n.Path = path.String
	n.Metadata = metadata.String
	n.JSON = jsonStr.String
	return &n, nil
}

func GetNodesByType(dbPath, nodeType string) ([]Node, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, type, guid, global_id, name, path, metadata, json,
		       created_at, updated_at, last_indexed_at
		FROM nodes WHERE type = ? ORDER BY created_at DESC
	`, nodeType)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		var path, metadata, jsonStr sql.NullString
		var globalID sql.NullString

		if err := rows.Scan(
			&n.ID, &n.Type, &n.GUID, &globalID, &n.Name,
			&path, &metadata, &jsonStr,
			&n.CreatedAt, &n.UpdatedAt, &n.LastIndexed,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		n.GlobalID = globalID.String
		n.Path = path.String
		n.Metadata = metadata.String
		n.JSON = jsonStr.String
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
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
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

func GetComponentsByGameObject(dbPath, gameObjNodeID string) ([]Node, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT 
			n.id, n.type, n.guid, n.name,
			COALESCE(s.target, '') as script_id
		FROM nodes n
		JOIN edges e ON e.target = n.id
		LEFT JOIN edges s ON s.source = n.id AND s.relationship = 'USES_SCRIPT'
		WHERE e.source = ? AND e.relationship = 'HAS_COMPONENT'
		ORDER BY n.name
	`, gameObjNodeID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		var scriptID string
		if err := rows.Scan(&n.ID, &n.Type, &n.GUID, &n.Name, &scriptID); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		// Store the script ID in the node's metadata field (temporary hack)
		// Better: extend the Node struct with a ScriptID field.
		n.Metadata = scriptID
		nodes = append(nodes, n)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

type Node struct {
	ID          string
	Type        string
	GUID        string
	GlobalID    string
	Name        string
	Path        string
	Metadata    string
	JSON        string
	CreatedAt   int64
	UpdatedAt   int64
	LastIndexed int64
}

// ---------- Edge query functions ----------

// Edge represents a relationship between two nodes.
type Edge struct {
	SourceID     string
	TargetID     string
	Relationship string
	Metadata     string
	CreatedAt    int64
}

// GetNodeByID retrieves a node by its ID.
func GetNodeByID(dbPath, id string) (*Node, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var n Node
	var path, metadata, jsonStr sql.NullString
	var globalID sql.NullString

	err = db.QueryRow(`
		SELECT id, type, guid, global_id, name, path, metadata, json,
		       created_at, updated_at, last_indexed_at
		FROM nodes WHERE id = ?
	`, id).Scan(
		&n.ID, &n.Type, &n.GUID, &globalID, &n.Name,
		&path, &metadata, &jsonStr,
		&n.CreatedAt, &n.UpdatedAt, &n.LastIndexed,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	n.GlobalID = globalID.String
	n.Path = path.String
	n.Metadata = metadata.String
	n.JSON = jsonStr.String
	return &n, nil
}

// GetEdgesBySource returns all edges where the given node is the source.
func GetEdgesBySource(dbPath, sourceID string) ([]Edge, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT source, target, relationship, metadata, created_at
		 FROM edges WHERE source = ?`,
		sourceID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges by source: %w", err)
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.SourceID, &e.TargetID, &e.Relationship, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}
		edges = append(edges, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return edges, nil
}

// GetEdgesByTarget returns all edges where the given node is the target.
func GetEdgesByTarget(dbPath, targetID string) ([]Edge, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT source, target, relationship, metadata, created_at
		 FROM edges WHERE target = ?`,
		targetID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges by target: %w", err)
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.SourceID, &e.TargetID, &e.Relationship, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}
		edges = append(edges, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return edges, nil
}

// GetEdgesByRelationship returns all edges with a specific relationship.
func GetEdgesByRelationship(dbPath, relationship string) ([]Edge, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	rows, err := db.Query(
		`SELECT source, target, relationship, metadata, created_at
		 FROM edges WHERE relationship = ?`,
		relationship,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query edges by relationship: %w", err)
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.SourceID, &e.TargetID, &e.Relationship, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan edge: %w", err)
		}
		edges = append(edges, e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return edges, nil
}
