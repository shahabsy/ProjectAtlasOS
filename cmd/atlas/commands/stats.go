package commands

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
	"github.com/spf13/cobra"
)

// statsCmd is the root of "stats" subcommands.
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show statistics",
}

var projectStatsCmd = &cobra.Command{
	Use:   "project",
	Short: "Show project statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := registry.Stats.ProjectStats(tools.ProjectStatsRequest{})
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(resp)
		} else {
			s := resp.Summary
			fmt.Println("Project Statistics:")
			fmt.Printf("  Scenes:             %d\n", s.TotalScenes)
			fmt.Printf("  GameObjects:        %d\n", s.TotalGameObjects)
			fmt.Printf("  Components:         %d\n", s.TotalComponents)
			fmt.Printf("  Scripts:            %d\n", s.TotalScripts)
			fmt.Printf("  Unique Scripts:     %d\n", s.UniqueScripts)
			fmt.Printf("  Assets:             %d\n", s.TotalAssets)
			fmt.Printf("  Serialized Fields:  %d\n", s.TotalSerializedFields)
		}
		return nil
	},
}

var sceneStatsCmd = &cobra.Command{
	Use:   "scene <scene-id>",
	Short: "Show statistics for a scene",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sceneID := args[0]
		resp, err := registry.Stats.SceneStats(tools.SceneStatsRequest{SceneID: sceneID})
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(resp)
		} else {
			s := resp.Statistics
			fmt.Printf("Scene Statistics (ID: %s):\n", sceneID)
			fmt.Printf("  GameObjects:  %d\n", s.GameObjectCount)
			fmt.Printf("  Components:   %d\n", s.ComponentCount)
			fmt.Printf("  Scripts:      %d\n", s.ScriptCount)
			fmt.Printf("  Prefabs:      %d\n", s.PrefabCount)
			fmt.Printf("  Assets:       %d\n", s.AssetCount)
			fmt.Printf("  Fields:       %d\n", s.FieldCount)
		}
		return nil
	},
}

func init() {
	statsCmd.AddCommand(projectStatsCmd)
	statsCmd.AddCommand(sceneStatsCmd)
}
