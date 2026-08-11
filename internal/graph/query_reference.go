package graph

import "fmt"

// ReferenceQueries provides queries for references (edges).
type ReferenceQueries struct {
	repo *Repository
}

func NewReferenceQueries(repo *Repository) *ReferenceQueries {
	return &ReferenceQueries{repo: repo}
}

// Incoming returns all edges where the given node is the target.
func (q *ReferenceQueries) Incoming(nodeID string) ([]EdgeInfo, error) {
	rows, err := q.repo.Query(`
		SELECT source, relationship, metadata
		FROM edges WHERE target = ?
	`, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incoming references for %s: %w", nodeID, err)
	}
	defer rows.Close()

	var edges []EdgeInfo
	for rows.Next() {
		var e EdgeInfo
		if err := rows.Scan(&e.SourceID, &e.Relationship, &e.Metadata); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		e.TargetID = nodeID
		edges = append(edges, e)
	}
	return edges, nil
}

// Outgoing returns all edges where the given node is the source.
func (q *ReferenceQueries) Outgoing(nodeID string) ([]EdgeInfo, error) {
	rows, err := q.repo.Query(`
		SELECT target, relationship, metadata
		FROM edges WHERE source = ?
	`, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get outgoing references from %s: %w", nodeID, err)
	}
	defer rows.Close()

	var edges []EdgeInfo
	for rows.Next() {
		var e EdgeInfo
		if err := rows.Scan(&e.TargetID, &e.Relationship, &e.Metadata); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		e.SourceID = nodeID
		edges = append(edges, e)
	}
	return edges, nil
}
