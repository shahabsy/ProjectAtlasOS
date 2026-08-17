package ai

import (
	"encoding/json"
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
)

// BuildSystemContext creates a concise system prompt with project summary and capabilities.
// It calls the "get_graph_capabilities" tool via the registry to get real data.
func BuildSystemContext(registry *tools.ToolRegistry) (string, error) {
	// Call get_graph_capabilities using the registry.
	callables := registry.AllCallables()
	capabilityFn, ok := callables["get_graph_capabilities"]
	if !ok {
		return "", fmt.Errorf("get_graph_capabilities tool not found")
	}
	// Execute with empty args.
	resultBytes, err := capabilityFn(json.RawMessage(`{}`))
	if err != nil {
		return "", err
	}
	// Unmarshal result (expecting tools.Result).
	var toolResult tools.Result
	if err := json.Unmarshal(resultBytes, &toolResult); err != nil {
		return "", fmt.Errorf("failed to unmarshal tool result: %w", err)
	}
	if !toolResult.Success {
		return "", fmt.Errorf("get_graph_capabilities failed: %s", toolResult.Error.Message)
	}
	// Extract data.
	dataBytes, err := json.Marshal(toolResult.Data)
	if err != nil {
		return "", err
	}
	var result struct {
		ProjectOverview struct {
			Scenes      int `json:"scenes"`
			GameObjects int `json:"game_objects"`
			Components  int `json:"components"`
			Scripts     int `json:"scripts"`
			Assets      int `json:"assets"`
		} `json:"project_overview"`
		Capabilities struct {
			Scenes           bool `json:"scenes"`
			GameObjects      bool `json:"game_objects"`
			Components       bool `json:"components"`
			Scripts          bool `json:"scripts"`
			SerializedFields bool `json:"serialized_fields"`
			AssetReferences  bool `json:"asset_references"`
			Prefabs          bool `json:"prefabs"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(dataBytes, &result); err != nil {
		return "", err
	}

	// Build the prompt.
	prompt := fmt.Sprintf(`You are Atlas, an AI assistant for Unity projects. You help developers understand their Unity project structure and relationships.

Project Summary:
- Scenes: %d
- GameObjects: %d
- Components: %d
- Scripts: %d
- Assets: %d

Capabilities:
`, result.ProjectOverview.Scenes, result.ProjectOverview.GameObjects,
		result.ProjectOverview.Components, result.ProjectOverview.Scripts,
		result.ProjectOverview.Assets)

	capabilities := map[string]bool{
		"Scenes":           result.Capabilities.Scenes,
		"GameObjects":      result.Capabilities.GameObjects,
		"Components":       result.Capabilities.Components,
		"Scripts":          result.Capabilities.Scripts,
		"SerializedFields": result.Capabilities.SerializedFields,
		"AssetReferences":  result.Capabilities.AssetReferences,
		"Prefabs":          result.Capabilities.Prefabs,
	}
	for name, available := range capabilities {
		status := "✓"
		if !available {
			status = "✗"
		}
		prompt += fmt.Sprintf("- %s: %s\n", name, status)
	}

	prompt += `
You have access to the following tools to retrieve detailed information. Always use the tools to answer questions; do not guess or fabricate data.

When you need to answer a question:
1. Decide which tool(s) are needed.
2. Call the tool(s) with appropriate arguments.
3. Use the returned data to answer the user's question.

If you are unsure or the required information is not available, say so clearly.

Current conversation:`

	return prompt, nil
}
