package ai

import (
	"encoding/json"
	"fmt"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
)

func BuildSystemContext(registry *tools.ToolRegistry) (string, error) {
	callables := registry.AllCallables()
	capabilityFn, ok := callables["get_graph_capabilities"]
	if !ok {
		return "", fmt.Errorf("get_graph_capabilities tool not found")
	}
	resultBytes, err := capabilityFn(json.RawMessage(`{}`))
	if err != nil {
		return "", err
	}
	var toolResult tools.Result
	if err := json.Unmarshal(resultBytes, &toolResult); err != nil {
		return "", fmt.Errorf("failed to unmarshal tool result: %w", err)
	}
	if !toolResult.Success {
		return "", fmt.Errorf("get_graph_capabilities failed: %s", toolResult.Error.Message)
	}
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

	prompt := fmt.Sprintf(`You are Atlas, an AI assistant for Unity projects.

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

	type capItem struct {
		name      string
		available bool
	}
	capabilities := []capItem{
		{"Scenes", result.Capabilities.Scenes},
		{"GameObjects", result.Capabilities.GameObjects},
		{"Components", result.Capabilities.Components},
		{"Scripts", result.Capabilities.Scripts},
		{"SerializedFields", result.Capabilities.SerializedFields},
		{"AssetReferences", result.Capabilities.AssetReferences},
		{"Prefabs", result.Capabilities.Prefabs},
	}
	for _, c := range capabilities {
		status := "✓"
		if !c.available {
			status = "✗"
		}
		prompt += fmt.Sprintf("- %s: %s\n", c.name, status)
	}

	prompt += `
You have access to a set of tools that can query the Unity project graph.
For common queries (listing assets, scenes, scripts, gameobjects, project stats), the system handles them directly.
You are only invoked for questions that require reasoning, multi-step planning, or ambiguous interpretation.

**Rules:**
- Never guess or fabricate data. Use the tools when needed.
- For complex questions, plan the tool calls and execute them sequentially.
- Do not ask follow‑up questions; provide the final answer using the tool results.

Current conversation:`
	return prompt, nil
}
