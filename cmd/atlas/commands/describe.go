package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// describeCmd is the root of "describe" subcommands.
var describeCmd = &cobra.Command{
	Use:   "describe",
	Short: "Describe a resource (scene, GameObject, etc.)",
}

var describeSceneCmd = &cobra.Command{
	Use:   "scene <scene-id>",
	Short: "Describe a scene",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sceneID := args[0]
		resp, err := registry.Scene.DescribeScene(sceneID)
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(resp)
		} else {
			fmt.Printf("Scene: %s (ID: %s, GUID: %s)\n", resp.Scene.Name, resp.Scene.ID, resp.Scene.GUID)
			fmt.Printf("  GameObjects: %d\n", resp.Statistics.GameObjectCount)
			fmt.Printf("  Components:  %d\n", resp.Statistics.ComponentCount)
			fmt.Printf("  Scripts:     %d\n", resp.Statistics.ScriptCount)
			if len(resp.Warnings) > 0 {
				fmt.Printf("  Warnings:    %v\n", resp.Warnings)
			}
		}
		return nil
	},
}

func init() {
	describeCmd.AddCommand(describeSceneCmd)
}
