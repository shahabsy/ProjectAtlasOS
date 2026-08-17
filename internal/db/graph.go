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
	sub_type TEXT,                  -- Added: semantic type (e.g., "Transform", "Rigidbody")
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
-- Additional indexes for performance
CREATE INDEX IF NOT EXISTS idx_nodes_sub_type ON nodes(sub_type);
CREATE INDEX IF NOT EXISTS idx_edges_source_rel ON edges(source, relationship);
CREATE INDEX IF NOT EXISTS idx_edges_target_rel ON edges(target, relationship);
`

type Node struct {
	ID          string
	Type        string // scene, gameobject, component, script, assets, prefab etc.
	SubType     string // Transform, Camera, Texture, etc.
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

// Edge represents a relationship between two nodes.
type Edge struct {
	SourceID     string
	TargetID     string
	Relationship string
	Metadata     string
	CreatedAt    int64
}

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

// InsertNode (non‑transaction) with sub_type.
func InsertNode(dbPath, id, typ, subType, guid, globalId, name, path string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	now := time.Now().Unix()
	_, err = db.Exec(`
		INSERT OR REPLACE INTO nodes (id, type, sub_type, guid, global_id, name, path, json, created_at, updated_at, last_indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, '{}', ?, ?, ?)`,
		id, typ, subType, guid, globalId, name, path, now, now, now)
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

// ---------- Transaction‑based operations ----------
func OpenWithBusyTimeout(dbPath string, timeoutMs int) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, err
	}
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

func InsertNodeTx(tx *sql.Tx, id, typ, subType, guid, globalId, name, path string) error {
	if globalId == "" {
		globalId = id
	}
	now := time.Now().Unix()
	_, err := tx.Exec(`
		INSERT OR REPLACE INTO nodes (id, type, sub_type, guid, global_id, name, path, json, created_at, updated_at, last_indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, '{}', ?, ?, ?)`,
		id, typ, subType, guid, globalId, name, path, now, now, now)
	return err
}

func InsertEdgeTx(tx *sql.Tx, srcID, tgtID, rel, metadata string) error {
	now := time.Now().Unix()
	_, err := tx.Exec(`
		INSERT OR REPLACE INTO edges (source, target, relationship, metadata, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		srcID, tgtID, rel, metadata, now)
	return err
}

// ---------- Query functions ----------
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
		SELECT id, type, sub_type, guid, global_id, name, path, metadata, json,
		       created_at, updated_at, last_indexed_at
		FROM nodes WHERE id = ?
	`, id).Scan(
		&n.ID, &n.Type, &n.SubType, &n.GUID, &globalID, &n.Name,
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
		SELECT id, type, sub_type, guid, global_id, name, path, metadata, json,
		       created_at, updated_at, last_indexed_at
		FROM nodes WHERE guid = ?
	`, guid).Scan(
		&n.ID, &n.Type, &n.SubType, &n.GUID, &globalID, &n.Name,
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
		SELECT id, type, sub_type, guid, global_id, name, path, metadata, json,
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
			&n.ID, &n.Type, &n.SubType, &n.GUID, &globalID, &n.Name,
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
		JOIN edges e ON e.target = n.id
		WHERE e.source = ? AND e.relationship = 'CONTAINS' AND n.type = 'gameobject'
		ORDER BY n.name
	`, sceneNodeID)
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
		// Store script ID in Metadata for now
		n.Metadata = scriptID
		nodes = append(nodes, n)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

// ---------- New semantic query functions ----------

func GetComponentsByType(dbPath, componentType string) ([]Node, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT id, type, sub_type, guid, name, created_at, updated_at, last_indexed_at
		FROM nodes
		WHERE type = 'component' AND COALESCE(sub_type, type) = ?
		ORDER BY name
	`, componentType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Type, &n.SubType, &n.GUID, &n.Name,
			&n.CreatedAt, &n.UpdatedAt, &n.LastIndexed); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

func GetGameObjectsWithComponentType(dbPath, componentType string) ([]Node, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT gobj.id, gobj.type, gobj.sub_type, gobj.guid, gobj.name
		FROM nodes gobj
		JOIN edges e ON e.source = gobj.id
		JOIN nodes comp ON e.target = comp.id
		WHERE comp.type = 'component' 
		  AND COALESCE(comp.sub_type, comp.type) = ?
		  AND e.relationship = 'HAS_COMPONENT'
		  AND gobj.type = 'gameobject'
	`, componentType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Type, &n.SubType, &n.GUID, &n.Name); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return nodes, nil
}

// ---------- Edge query functions ----------

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

// GetScriptsByScene returns all script nodes used by GameObjects in a scene.
func GetScriptsByScene(dbPath, sceneID string) ([]Node, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(`
		SELECT DISTINCT s.id, s.type, s.sub_type, s.guid, s.name
		FROM nodes s
		JOIN edges e_script ON e_script.target = s.id AND e_script.relationship = 'USES_SCRIPT'
		JOIN nodes comp ON comp.id = e_script.source AND comp.type = 'component'
		JOIN edges e_comp ON e_comp.target = comp.id AND e_comp.relationship = 'HAS_COMPONENT'
		JOIN nodes gobj ON gobj.id = e_comp.source AND gobj.type = 'gameobject'
		JOIN edges e_scene ON e_scene.target = gobj.id AND e_scene.relationship = 'CONTAINS'
		WHERE e_scene.source = ?
	`, sceneID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scripts []Node
	for rows.Next() {
		var n Node
		// Only scan the columns we selected.
		if err := rows.Scan(&n.ID, &n.Type, &n.SubType, &n.GUID, &n.Name); err != nil {
			return nil, err
		}
		scripts = append(scripts, n)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return scripts, nil
}
