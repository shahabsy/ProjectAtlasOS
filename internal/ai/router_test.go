package ai

import (
	"testing"
)

func TestRouteQuery(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"List all assets", "list_assets"},
		{"Show me assets", "list_assets"},
		{"Give me a summary of assets", "list_assets"},
		{"What assets are in the project?", "list_assets"},
		{"List scenes", "list_scenes"},
		{"Show all scenes", "list_scenes"},
		{"List scripts", "list_all_scripts"},
		{"Show me scripts", "list_all_scripts"},
		{"Project stats", "project_stats"},
		{"Statistics", "project_stats"},
		{"How many GameObjects are there?", "project_stats"},
		{"Count components", "project_stats"},
		{"Number of scripts", "project_stats"},
		{"List gameobjects", "list_all_gameobjects"},
		{"Show GameObjects", "list_all_gameobjects"},
		{"Objects in the project", "list_all_gameobjects"},
		{"Why is my player not moving?", ""},
		{"Find components with missing references", ""},
		{"Describe the main scene", ""},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			decision := RouteQuery(tt.query, nil)
			if tt.expected == "" {
				if decision != nil {
					t.Errorf("expected nil, got %+v", decision)
				}
			} else {
				if decision == nil {
					t.Errorf("expected tool %s, got nil", tt.expected)
				} else if decision.Tool != tt.expected {
					t.Errorf("expected tool %s, got %s", tt.expected, decision.Tool)
				}
				if decision != nil && decision.Confidence != 1.0 {
					t.Errorf("expected confidence 1.0, got %f", decision.Confidence)
				}
			}
		})
	}
}
