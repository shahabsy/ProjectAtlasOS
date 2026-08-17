package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/ai"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "Start an interactive chat session with the Atlas AI Agent",
	RunE: func(cmd *cobra.Command, args []string) error {
		ollama := ai.NewOllamaClient("", "qwen2.5-coder:32b")
		agent := ai.NewAgent(ollama, registry, verbose)
		session := ai.NewSession(agent)

		fmt.Println("Atlas Chat (type 'exit' to quit)")
		scanner := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("> ")
			if !scanner.Scan() {
				break
			}
			input := strings.TrimSpace(scanner.Text())
			if input == "exit" || input == "quit" {
				break
			}
			result, err := session.Ask(input)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			if !result.Success {
				fmt.Printf("Agent error: %s\n", result.Error)
				continue
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
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("input scanner error: %w", err)
		}
		return nil
	},
}

func init() {
	chatCmd.Flags().BoolVar(&verbose, "verbose", false, "Show detailed agent execution trace")
	rootCmd.AddCommand(chatCmd)
}
