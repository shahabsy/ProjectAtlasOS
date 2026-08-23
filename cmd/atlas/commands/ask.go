package commands

import (
	"fmt"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/ai"
	"github.com/spf13/cobra"
)

var askCmd = &cobra.Command{
	Use:   "ask <question>",
	Short: "Ask a natural language question about the Unity project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		question := strings.Join(args, " ")
		ollama := ai.NewOllamaClient("", "qwen3-coder-next:latest")
		agent := ai.NewAgent(ollama, registry, verbose)
		result, err := agent.RunWithHistory(question, nil)
		if err != nil {
			return err
		}
		if !result.Success {
			return fmt.Errorf("agent failed: %s", result.Error)
		}
		if verbose {
			fmt.Println("\n[Trace]")
			for _, call := range result.ToolCalls {
				fmt.Printf("  Tool: %s\n", call.Name)
				if call.Error != "" {
					fmt.Printf("    Error: %s\n", call.Error)
				} else {
					fmt.Printf("    Result: %v\n", call.Result)
				}
			}
			fmt.Println()
		}
		fmt.Println(result.Answer)
		return nil
	},
}

func init() {
	// Bind the verbose flag (global) to this command.
	askCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed agent execution trace")
	rootCmd.AddCommand(askCmd)
}
