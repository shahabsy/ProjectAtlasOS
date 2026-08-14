package commands

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func getDBPathSafe() (string, error) {
	if dbPath != "" {
		return dbPath, nil
	}
	dir, err := os.Getwd()
	if err == nil {
		for {
			assetsPath := filepath.Join(dir, "Assets")
			if info, err := os.Stat(assetsPath); err == nil && info.IsDir() {
				return filepath.Join(dir, ".atlas", "graph.db"), nil
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return "", fmt.Errorf("could not find Unity project root (no 'Assets' folder found). Please specify --db flag")
}

func printJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
