package core

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
)

// IndexNodeCmd adds a node to the graph database
func IndexNodeCmd(typ, guid, name string) error {
	projectRoot, err := findUnityProject()
	if err != nil {
		return fmt.Errorf("not a Unity project: %w", err)
	}

	dbPath := db.GetDBPath(projectRoot)

	node, err := db.GetNodeByGUID(dbPath, guid)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	var id string
	if node == nil {
		// New code - generate ID from GUID (first 8 chars for readability)
		id = "node_" + guid[:8]
	} else {
		id = node.ID
	}

	if err := db.InsertNode(dbPath, id, typ, guid, name); err != nil {
		return fmt.Errorf("insert failed: %w", err)
	}

	fmt.Printf("Indexed %s '%s' (guid: %s)\n", typ, name, guid)
	return nil
}
