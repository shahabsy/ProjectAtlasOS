package graph

import (
	"database/sql"
	"fmt"
)

// Edge represents a relationship between two nodes in the graph.
type Edge struct {
	SourceID     string
	TargetID     string
	Relationship string
	Metadata     string
	CreatedAt    int64
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
		 FROM edges
		 WHERE source = ?`,
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
		return nil, fmt.Errorf("rows iteration error: %w", err)
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
		 FROM edges
		 WHERE target = ?`,
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
		return nil, fmt.Errorf("rows iteration error: %w", err)
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
		 FROM edges
		 WHERE relationship = ?`,
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
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return edges, nil
}
