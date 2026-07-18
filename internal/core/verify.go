package core

import (
	"database/sql"
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	_ "modernc.org/sqlite"
)

// VerifyCmd checks the database and prints all indexed nodes
func VerifyCmd() error {
	projectRoot, err := findUnityProject()
	if err != nil {
		return fmt.Errorf("not a Unity project: %w", err)
	}

	dbPath := db.GetDBPath(projectRoot)

	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	rows, err := database.Query("SELECT id, type, name FROM nodes ORDER BY created_at DESC")
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	fmt.Println("Atlas Graph Nodes:")
	count := 0
	for rows.Next() {
		var id, typ, name string
		if err := rows.Scan(&id, &typ, &name); err != nil {
			return fmt.Errorf("scan failed: %w", err)
		}
		fmt.Printf("  - %s (%s) [id: %s]\n", name, typ, id)
		count++
	}

	if count == 0 {
		fmt.Println(" No nodes found. Did you run `atlas index`?")
	} else {
		fmt.Printf("\n Verified %d node(s) in database\n", count)
	}

	return rows.Err()
}
