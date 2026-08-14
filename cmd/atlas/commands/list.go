package commands

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
	"github.com/spf13/cobra"
)

// listCmd is the root of "list" subcommands.
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List resources (scenes, assets, etc.)",
}

var listScenesCmd = &cobra.Command{
	Use:   "scenes",
	Short: "List all scenes",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := registry.Scene.ListScenes()
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(resp)
		} else {
			fmt.Printf("Scenes (%d):\n", len(resp.Scenes))
			for _, s := range resp.Scenes {
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
		resp, err := registry.Asset.ListAssets(tools.ListAssetsRequest{})
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(resp)
		} else {
			fmt.Printf("Assets (%d):\n", len(resp.Assets))
			for _, a := range resp.Assets {
				fmt.Printf("  - %s (ID: %s, Type: %s)\n", a.Name, a.ID, a.Type)
			}
		}
		return nil
	},
}

func init() {
	listCmd.AddCommand(listScenesCmd)
	listCmd.AddCommand(listAssetsCmd)
}
