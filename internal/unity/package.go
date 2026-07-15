package unity

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/db"
)

//go:embed packages/com.atlas.kernel/*
var upmContent embed.FS

func Install(projectRoot string, dryRun bool) error {
	pkgPath := filepath.Join(projectRoot, "Packages", "com.atlas.kernel")

	if dryRun {
		return printFiles()
	}

	// 1. Create target directory
	if err := os.MkdirAll(filepath.Dir(pkgPath), 0755); err != nil {
		return err
	}

	// 2. Copy embedded files
	if err := copyFS(upmContent, "packages/com.atlas.kernel", pkgPath); err != nil {
		return err
	}

	// 3. Initialize graph.db
	atlasDir := filepath.Join(projectRoot, ".atlas")
	graphDB := filepath.Join(atlasDir, "graph.db")

	if _, err := os.Stat(graphDB); os.IsNotExist(err) {
		if err := db.CreateDB(graphDB); err != nil {
			return err
		}
	}

	return nil
}

func printFiles() error {
	fmt.Println("Files to be installed (dry run):")
	return fs.WalkDir(upmContent, "packages/com.atlas.kernel", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		fmt.Println("  " + strings.TrimPrefix(path, "packages/"))
		return nil
	})
}

func copyFS(fsys embed.FS, src, dst string) error {
	return fs.WalkDir(fsys, src, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		destPath := filepath.Join(dst, strings.TrimPrefix(path, src))
		if destPath == dst {
			return nil // Skip copying the root directory itself
		}

		dir := filepath.Dir(destPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		content, readErr := fsys.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		return os.WriteFile(destPath, content, 0644)
	})
}
