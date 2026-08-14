package commands

import (
	"fmt"
	"os"

	"github.com/shahabsy/ProjectAtlasOS/internal/query"
	"github.com/shahabsy/ProjectAtlasOS/internal/statistics"
	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
	"github.com/spf13/cobra"
)

var (
	dbPath     string
	jsonOutput bool
	registry   *tools.ToolRegistry
)

var rootCmd = &cobra.Command{
	Use:   "atlas",
	Short: "Atlas – Unity knowledge graph CLI",
	Long: `Atlas provides deterministic tools for exploring Unity projects.
It builds a knowledge graph from your Unity scenes and exposes it via a CLI.

Commands:
  atlas list scenes       List all scenes
  atlas describe scene    Show scene details
  atlas find gameobject   Find GameObjects by name
  atlas stats project     Show project statistics
  atlas stats scene       Show scene statistics
  atlas list assets       List all assets`,
	Version: "0.5.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == "help" || cmd.Name() == "version" {
			return nil
		}
		db, err := getDBPathSafe()
		if err != nil {
			return err
		}
		q := query.NewEngine(db)
		stats := statistics.NewEngine(q)
		ctx := tools.NewContext(q, stats)
		registry = tools.NewToolRegistry(ctx)
		return nil
	},
}

// findCmd is the root of "find" subcommands.
var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find resources (GameObjects, components, scripts)",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "Path to the graph.db file (default: auto‑detect)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	// Add subcommands
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(describeCmd)
	rootCmd.AddCommand(findCmd)
	rootCmd.AddCommand(statsCmd)
}
