package commands

import (
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
	"github.com/spf13/cobra"
)

var findComponentCmd = &cobra.Command{
	Use:   "component --type <type>",
	Short: "Find components by Unity type",
	RunE: func(cmd *cobra.Command, args []string) error {
		compType, _ := cmd.Flags().GetString("type")
		if compType == "" {
			return fmt.Errorf("--type is required")
		}
		resp, err := registry.Component.FindComponentsByType(tools.FindComponentsByTypeRequest{Type: compType})
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(resp)
		} else {
			fmt.Printf("Components of type '%s' (%d):\n", compType, len(resp.Components))
			for _, c := range resp.Components {
				fmt.Printf("  - %s (ID: %s)\n", c.Type, c.ID)
			}
		}
		return nil
	},
}

var findGameObjectWithComponentCmd = &cobra.Command{
	Use:   "gameobject --component <type>",
	Short: "Find GameObjects that have a component of a specific type",
	RunE: func(cmd *cobra.Command, args []string) error {
		compType, _ := cmd.Flags().GetString("component")
		if compType == "" {
			return fmt.Errorf("--component is required")
		}
		resp, err := registry.GameObject.FindGameObjectsWithComponent(tools.FindGameObjectsWithComponentRequest{ComponentType: compType})
		if err != nil {
			return err
		}
		if jsonOutput {
			printJSON(resp)
		} else {
			fmt.Printf("GameObjects with component '%s' (%d):\n", compType, len(resp.GameObjects))
			for _, g := range resp.GameObjects {
				fmt.Printf("  - %s (ID: %s)\n", g.Name, g.ID)
			}
		}
		return nil
	},
}

func init() {
	findComponentCmd.Flags().String("type", "", "Component type to find (e.g., Transform)")
	findGameObjectWithComponentCmd.Flags().String("component", "", "Component type to search for")

	findCmd.AddCommand(findComponentCmd)
	findCmd.AddCommand(findGameObjectWithComponentCmd)
}
