package graph

type Traversal struct {
	repo *Repository
}

func NewTraversal(repo *Repository) *Traversal {
	return &Traversal{repo: repo}
}

// Children returns all child node IDs (incomming edges).
func (t *Traversal) Children(nodeID string) ([]string, error) {
	rows, err := t.repo.Query(`SELECT target FROM edges WHERE source = ?`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var children []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		children = append(children, id)
	}
	return children, nil
}

// Parents returns all parent node IDs (outgoing edges).
func (t *Traversal) Parents(nodeId string) ([]string, error) {
	rows, err := t.repo.Query(`SELECT source FROM edges WHERE target = ?`, nodeId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parents []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		parents = append(parents, id)
	}
	return parents, nil
}

// Neighors returns all connected node IDs (both directions).
func (t *Traversal) Neighbors(nodeID string) ([]string, error) {
	rows, err := t.repo.Query(`SELECT target FROM edges WHERE source = ?
	UNION
	SELECT source FROM edges WHERE target = ?`, nodeID, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var neighbors []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		neighbors = append(neighbors, id)
	}
	return neighbors, nil
}

// Walk traverses the graph from a start node, following a specific relationship.
func (t *Traversal) Walk(startID, relationship string) ([]string, error) {
	rows, err := t.repo.Query(`
	WITH RECURSIVE walk(id) AS(
		SELECT ? UNION
		SELECT target FROM edges
		JOIN walk ON edges.source = walk.id
		WHERE edges.relationship = ?
	)
	SELECT id FROM walk WHERE id != ?`, startID, relationship, startID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

// DEfendencyChain returns the full dependcency chain from a node to its leaves.
func (t *Traversal) DependencyChain(startID string) ([]EdgeInfo, error) {
	rows, err := t.repo.Query(`
		WITH RECURSIVE chain(src, tgt, rel) AS (
			SELECT source, target, relationship FROM edges WHERE source = ?
			UNION
			SELECT e.source, e.target, e.relationship
			FROM edges e
			JOIN chain c ON c.source = c.tgt
		)
		SELECT src, tgt, rel FROM chain`, startID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []EdgeInfo
	for rows.Next() {
		var e EdgeInfo
		if err := rows.Scan(&e.SourceID, &e.TargetID, &e.Relationship); err != nil {
			return nil, err
		}
		edges = append(edges, e)
	}
	return edges, nil
}
