package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shahabsy/ProjectAtlasOS/internal/query"
	"github.com/shahabsy/ProjectAtlasOS/internal/statistics"
	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
	"github.com/spf13/cobra"
)

var (
	registry   *tools.ToolRegistry
	dbPath     string
	jsonOutput bool
	verbose    bool // global verbose flag
)

func getDefaultDBPath() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		dbPath := filepath.Join(dir, ".atlas", "graph.db")
		if _, err := os.Stat(dbPath); err == nil {
			return dbPath
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

var rootCmd = &cobra.Command{
	Use:   "atlas",
	Short: "Atlas - Unity project analysis tool",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if dbPath == "" {
			dbPath = getDefaultDBPath()
			if dbPath == "" {
				return fmt.Errorf("database not found. Please specify --db or run from a Unity project root")
			}
		}
		q := query.NewEngine(dbPath)
		stats := statistics.NewEngine(q)
		ctx := tools.NewContext(q, stats)
		registry = tools.NewToolRegistry(ctx)
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "Path to graph.db (auto‑detected if omitted)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	// Add a global verbose flag; subcommands can also set their own.
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Show detailed agent execution trace")
}
