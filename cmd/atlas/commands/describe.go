package commands

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
	"github.com/spf13/cobra"
)

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
		res := registry.Scene.DescribeScene(sceneID)
		if !res.Success {
			return fmt.Errorf("failed to describe scene: %s", res.Error.Message)
		}
		descResp, ok := res.Data.(tools.DescribeSceneResponse)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}
		if jsonOutput {
			printJSON(descResp)
		} else {
			fmt.Printf("Scene: %s (ID: %s, GUID: %s)\n", descResp.Scene.Name, descResp.Scene.ID, descResp.Scene.GUID)
			fmt.Printf("  GameObjects: %d\n", descResp.Statistics.GameObjectCount)
			fmt.Printf("  Components:  %d\n", descResp.Statistics.ComponentCount)
			fmt.Printf("  Scripts:     %d\n", descResp.Statistics.ScriptCount)
			if len(descResp.Warnings) > 0 {
				fmt.Printf("  Warnings:    %v\n", descResp.Warnings)
			}
		}
		return nil
	},
}

func init() {
	describeCmd.AddCommand(describeSceneCmd)
	rootCmd.AddCommand(describeCmd) // <-- ADD THIS LINE
}
