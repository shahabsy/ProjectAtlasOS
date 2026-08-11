package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
	"github.com/shahabsy/ProjectAtlasOS/internal/unity"
	"github.com/shahabsy/ProjectAtlasOS/internal/utils"
)

func Init(dryRun bool) error {
	projectRoot, err := findUnityProject()
	if err != nil {
		return fmt.Errorf("not a Unity project: %w", err)
	}

	fmt.Printf("[Atlas] Found Unity project at: %s\n", projectRoot)

	if err := unity.Install(projectRoot, dryRun); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	if dryRun {
		fmt.Println("\n Dry run complete -- no files written.")
	} else {
		fmt.Println("\n Atlas kernel installed successfully.")
	}
	// Ensure database schema exista and run migration
	dbPath := db.GetDBPath(projectRoot)

	// Craete database schema if not present
	if err := db.CreateDB(dbPath); err != nil {
		return fmt.Errorf("database creation failed: %w", err)
	}
	// Migrate database
	if err := db.RunMigrations(dbPath); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}
	return nil
}

func findUnityProject() (string, error) {
	cwd, _ := os.Getwd()
	for {
		if utils.Exists(filepath.Join(cwd, "Assets")) &&
			utils.Exists(filepath.Join(cwd, "ProjectSettings", "ProjectVersion.txt")) {
			return cwd, nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return "", fmt.Errorf("no Unity project found.")
}
