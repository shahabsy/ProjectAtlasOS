package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
)

const maxAgentTurns = 8

// Agent is the main AI agent that orchestrates conversation and tool calls.
type Agent struct {
	ollama   *OllamaClient
	registry *tools.ToolRegistry
	history  []Message
}

// NewAgent creates a new Agent.
func NewAgent(ollama *OllamaClient, registry *tools.ToolRegistry) *Agent {
	return &Agent{
		ollama:   ollama,
		registry: registry,
		history:  []Message{},
	}
}

// Run processes a user query and returns the final answer.
func (a *Agent) Run(userQuery string) (string, error) {
	// Build system context (static, deterministic).
	systemPrompt, err := BuildSystemContext(a.registry)
	if err != nil {
		return "", fmt.Errorf("failed to build system context: %w", err)
	}

	// Initialise history.
	a.history = []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userQuery},
	}

	// Prepare tool definitions once.
	toolDefs := a.buildToolDefinitions()
	callables := a.registry.AllCallables()

	// Agent loop.
	for turn := 0; turn < maxAgentTurns; turn++ {
		resp, err := a.ollama.Chat(a.history, toolDefs)
		if err != nil {
			return "", fmt.Errorf("ollama chat error: %w", err)
		}

		// Always preserve the assistant response.
		a.history = append(a.history, *resp)

		// If the model didn't request any tool, this is the final answer.
		if len(resp.ToolCalls) == 0 {
			if strings.TrimSpace(resp.Content) == "" {
				return "", fmt.Errorf("model returned empty response after tool execution")
			}
			return resp.Content, nil
		}

		// Execute all requested tools.
		for _, call := range resp.ToolCalls {
			toolName := call.Function.Name
			args := call.Function.Arguments

			// Validate that the tool exists.
			fn, ok := callables[toolName]
			if !ok {
				errMsg := fmt.Sprintf("Tool '%s' not found", toolName)
				a.history = append(a.history, Message{
					Role:     "tool",
					ToolName: toolName,
					Content:  errMsg,
				})
				continue
			}

			// Convert arguments to JSON for the tool.
			argBytes, err := json.Marshal(args)
			if err != nil {
				errMsg := fmt.Sprintf("Failed to marshal arguments: %v", err)
				a.history = append(a.history, Message{
					Role:     "tool",
					ToolName: toolName,
					Content:  errMsg,
				})
				continue
			}

			// Execute the tool.
			resultBytes, err := fn(argBytes)
			if err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				a.history = append(a.history, Message{
					Role:     "tool",
					ToolName: toolName,
					Content:  errMsg,
				})
				continue
			}

			// Append the tool result as a "tool" role message.
			a.history = append(a.history, Message{
				Role:     "tool",
				ToolName: toolName,
				Content:  string(resultBytes),
			})
		}
		// Loop continues – the model will process the tool results in the next turn.
	}
	return "", fmt.Errorf("agent reached max turns (%d) without final answer", maxAgentTurns)
}

// buildToolDefinitions converts the registry tools to Ollama's native Tool format.
func (a *Agent) buildToolDefinitions() []Tool {
	callables := a.registry.AllCallables()
	toolDefs := map[string]FunctionDef{
		"list_scenes": {
			Name:        "list_scenes",
			Description: "Returns all scenes in the Unity project. Use this before any tool that requires a scene_id, unless you already have the exact ID.",
			Parameters: Parameters{
				Type:       "object",
				Properties: map[string]Property{},
				Required:   []string{},
			},
		},
		"describe_scene": {
			Name: "describe_scene",
			Description: `Returns detailed information about a scene, including statistics.
IMPORTANT: scene_id must be an actual ID returned by list_scenes. Never invent an ID.
If the user refers to a scene by name, call list_scenes first to obtain the correct ID.`,
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"scene_id": {Type: "string", Description: "The ID of the scene to describe"},
				},
				Required: []string{"scene_id"},
			},
		},
		"find_gameobjects": {
			Name: "find_gameobjects",
			Description: `Finds GameObjects by name (partial match) within a scene.
IMPORTANT: scene_id must be an actual ID returned by list_scenes. Do not guess.`,
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"scene_id": {Type: "string", Description: "The ID of the scene"},
					"name":     {Type: "string", Description: "The name or partial name to search for"},
				},
				Required: []string{"scene_id", "name"},
			},
		},
		"get_gameobject_hierarchy": {
			Name: "get_gameobject_hierarchy",
			Description: `Returns the GameObject's parent/children hierarchy.
IMPORTANT: gameobject_id must be a real ID from find_gameobjects or the graph. Do not invent.`,
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"gameobject_id": {Type: "string", Description: "The ID of the GameObject"},
				},
				Required: []string{"gameobject_id"},
			},
		},
		"get_components": {
			Name: "get_components",
			Description: `Returns all components attached to a GameObject.
IMPORTANT: gameobject_id must be a real ID from find_gameobjects or the graph.`,
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"gameobject_id": {Type: "string", Description: "The ID of the GameObject"},
				},
				Required: []string{"gameobject_id"},
			},
		},
		"find_components_by_type": {
			Name:        "find_components_by_type",
			Description: "Finds all components of a specific Unity type (e.g., 'Rigidbody', 'MeshRenderer').",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"type": {Type: "string", Description: "The Unity component class name (e.g., 'Transform')"},
				},
				Required: []string{"type"},
			},
		},
		"find_scripts": {
			Name:        "find_scripts",
			Description: "Finds scripts by name (partial match).",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"name": {Type: "string", Description: "Script name to search for"},
				},
				Required: []string{"name"},
			},
		},
		"get_gameobjects_using_script": {
			Name:        "get_gameobjects_using_script",
			Description: "Returns all GameObjects that have a component using the given script.",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"script_id": {Type: "string", Description: "The ID of the script"},
				},
				Required: []string{"script_id"},
			},
		},
		"list_assets": {
			Name:        "list_assets",
			Description: "Returns all assets in the project.",
			Parameters: Parameters{
				Type:       "object",
				Properties: map[string]Property{},
				Required:   []string{},
			},
		},
		"project_stats": {
			Name:        "project_stats",
			Description: "Returns aggregated project statistics.",
			Parameters: Parameters{
				Type:       "object",
				Properties: map[string]Property{},
				Required:   []string{},
			},
		},
		"scene_stats": {
			Name:        "scene_stats",
			Description: "Returns statistics for a specific scene. Requires a valid scene_id.",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"scene_id": {Type: "string", Description: "The ID of the scene"},
				},
				Required: []string{"scene_id"},
			},
		},
		"get_graph_capabilities": {
			Name:        "get_graph_capabilities",
			Description: "Returns a summary of project and available graph data domains.",
			Parameters: Parameters{
				Type:       "object",
				Properties: map[string]Property{},
				Required:   []string{},
			},
		},
	}

	var tools []Tool
	for name, fn := range toolDefs {
		if _, ok := callables[name]; ok {
			tools = append(tools, Tool{
				Type:     "function",
				Function: fn,
			})
		}
	}
	return tools
}
