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
		agent := ai.NewAgent(ollama, registry)
		answer, err := agent.Run(question)
		if err != nil {
			return err
		}
		// Print answer, even if empty
		if answer == "" {
			fmt.Println("[Atlas] No answer received. The model may not have responded.")
		} else {
			fmt.Println(answer)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(askCmd)
}
