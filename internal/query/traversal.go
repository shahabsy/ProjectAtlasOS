package query

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	"github.com/shahabsy/ProjectAtlasOS/internal/models"
)

// WalkParents returns the IDs of all parent nodes (including ancestors) of a given node.
func (e *Engine) WalkParents(nodeID string) ([]string, error) {
	var parents []string
	current := nodeID
	for {
		edges, err := db.GetEdgesByTarget(e.dbPath, current)
		if err != nil {
			return nil, fmt.Errorf("failed to walk parents for %s: %w", nodeID, err)
		}
		var parentID string
		for _, edge := range edges {
			if edge.Relationship == models.EdgeChildOf {
				parentID = edge.SourceID
				break
			}
		}
		if parentID == "" {
			break
		}
		parents = append(parents, parentID)
		current = parentID
	}
	return parents, nil
}

// WalkChildren returns the immediate children of a given node (one level down).
func (e *Engine) WalkChildren(nodeID string) ([]string, error) {
	edges, err := db.GetEdgesBySource(e.dbPath, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to walk children for %s: %w", nodeID, err)
	}
	var children []string
	for _, edge := range edges {
		if edge.Relationship == models.EdgeChildOf {
			children = append(children, edge.TargetID)
		}
	}
	return children, nil
}

// GetNeighbors returns all nodes connected to the given node via a specific relationship.
func (e *Engine) GetNeighbors(nodeID, relationship string) ([]string, error) {
	outEdges, err := db.GetEdgesBySource(e.dbPath, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get outgoing edges: %w", err)
	}
	inEdges, err := db.GetEdgesByTarget(e.dbPath, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incoming edges: %w", err)
	}

	var neighbors []string
	for _, edge := range outEdges {
		if relationship == "" || edge.Relationship == relationship {
			neighbors = append(neighbors, edge.TargetID)
		}
	}
	for _, edge := range inEdges {
		if relationship == "" || edge.Relationship == relationship {
			neighbors = append(neighbors, edge.SourceID)
		}
	}
	return neighbors, nil
}

// WalkDependencies returns all nodes that depend on the given node (incoming edges).
func (e *Engine) WalkDependencies(nodeID string) ([]string, error) {
	edges, err := db.GetEdgesByTarget(e.dbPath, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get incoming edges for dependencies: %w", err)
	}
	var deps []string
	for _, edge := range edges {
		if edge.Relationship != "CHILD_OF" {
			deps = append(deps, edge.SourceID)
		}
	}
	return deps, nil
}

// WalkReverseDependencies returns all nodes that the given node depends on (outgoing edges).
func (e *Engine) WalkReverseDependencies(nodeID string) ([]string, error) {
	edges, err := db.GetEdgesBySource(e.dbPath, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get outgoing edges for reverse dependencies: %w", err)
	}
	var deps []string
	for _, edge := range edges {
		if edge.Relationship != "CHILD_OF" {
			deps = append(deps, edge.TargetID)
		}
	}
	return deps, nil
}
