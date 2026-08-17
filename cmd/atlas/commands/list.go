package commands

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources (scenes, assets, etc.)",
}

var listScenesCmd = &cobra.Command{
	Use:   "scenes",
	Short: "List all scenes",
	RunE: func(cmd *cobra.Command, args []string) error {
		res := registry.Scene.ListScenes()
		if !res.Success {
			return fmt.Errorf("failed to list scenes: %s", res.Error.Message)
		}
		listResp, ok := res.Data.(tools.ListScenesResponse)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}
		if jsonOutput {
			printJSON(listResp)
		} else {
			fmt.Printf("Scenes (%d):\n", len(listResp.Scenes))
			for _, s := range listResp.Scenes {
				fmt.Printf("  - %s (ID: %s, GUID: %s)\n", s.Name, s.ID, s.GUID)
			}
		}
		return nil
	},
}

var listAssetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "List all assets",
	RunE: func(cmd *cobra.Command, args []string) error {
		res := registry.Asset.ListAssets()
		if !res.Success {
			return fmt.Errorf("failed to list assets: %s", res.Error.Message)
		}
		listResp, ok := res.Data.(tools.ListAssetsResponse)
		if !ok {
			return fmt.Errorf("unexpected response type")
		}
		if jsonOutput {
			printJSON(listResp)
		} else {
			fmt.Printf("Assets (%d):\n", len(listResp.Assets))
			for _, a := range listResp.Assets {
				fmt.Printf("  - %s (ID: %s, Type: %s)\n", a.Name, a.ID, a.Type)
			}
		}
		return nil
	},
}

func init() {
	listCmd.AddCommand(listScenesCmd)
	listCmd.AddCommand(listAssetsCmd)
	rootCmd.AddCommand(listCmd) // <-- ADD THIS LINE
}
