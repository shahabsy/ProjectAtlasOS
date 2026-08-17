package commands

import (
	"testing"

	"github.com/shahabsy/ProjectAtlasOS/internal/ai"
	"github.com/shahabsy/ProjectAtlasOS/internal/query"
	"github.com/shahabsy/ProjectAtlasOS/internal/statistics"
	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
)

func TestAgentQueries(t *testing.T) {
	// This test requires a real graph.db; we'll skip if not present.
	// For CI, you could use a test database.

	dbPath := "D:/PortfolioProject/XtreamShooter/.atlas/graph.db"
	q := query.NewEngine(dbPath)
	stats := statistics.NewEngine(q)
	ctx := tools.NewContext(q, stats)
	reg := tools.NewToolRegistry(ctx)
	ollama := ai.NewOllamaClient("", "qwen3-coder-next:latest")
	agent := ai.NewAgent(ollama, reg, true) // verbose for debugging

	tests := []struct {
		name           string
		query          string
		expectedTools  []string
		answerContains string
	}{
		{
			name:           "list scenes",
			query:          "List all scenes",
			expectedTools:  []string{"list_scenes"},
			answerContains: "Game",
		},
		{
			name:           "count gameobjects",
			query:          "How many GameObjects are in the scene?",
			expectedTools:  []string{"scene_stats"},
			answerContains: "72",
		},
		// Add more tests.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := agent.RunWithHistory(tt.query, nil)
			if err != nil {
				t.Fatalf("agent error: %v", err)
			}
			if !result.Success {
				t.Fatalf("agent failed: %s", result.Error)
			}
			// Check that expected tools were called.
			toolNames := make(map[string]bool)
			for _, call := range result.ToolCalls {
				toolNames[call.Name] = true
			}
			for _, expected := range tt.expectedTools {
				if !toolNames[expected] {
					t.Errorf("expected tool %s not called", expected)
				}
			}
			// Check answer contains expected substring.
			if tt.answerContains != "" && !contains(result.Answer, tt.answerContains) {
				t.Errorf("answer did not contain '%s': %s", tt.answerContains, result.Answer)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			(s[1:len(substr)+1] == substr)))
}
