package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
)

const MaxAgentTurns = 8
const MaxToolCalls = 3

// Agent is the main AI agent that orchestrates conversation and tool calls.
type Agent struct {
	ollama   *OllamaClient
	registry *tools.ToolRegistry
	verbose  bool
}

// NewAgent creates a new Agent.
func NewAgent(ollama *OllamaClient, registry *tools.ToolRegistry, verbose bool) *Agent {
	return &Agent{
		ollama:   ollama,
		registry: registry,
		verbose:  verbose,
	}
}

// extractToolCallFromContent finds a JSON object with "name" and "arguments".
func extractToolCallFromContent(content string) (name string, args map[string]interface{}, found bool) {
	start := strings.Index(content, "{")
	if start == -1 {
		return "", nil, false
	}
	depth := 0
	end := -1
outer:
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i + 1
				break outer
			}
		}
	}
	if end == -1 {
		return "", nil, false
	}
	jsonPart := content[start:end]
	var parsed struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal([]byte(jsonPart), &parsed); err == nil && parsed.Name != "" {
		return parsed.Name, parsed.Arguments, true
	}
	return "", nil, false
}

// stripToolCallFromContent removes the JSON tool call from the content.
func stripToolCallFromContent(content string) string {
	if _, _, found := extractToolCallFromContent(content); found {
		start := strings.Index(content, "{")
		end := strings.LastIndex(content, "}")
		if start != -1 && end != -1 && end > start {
			cleaned := strings.TrimSpace(content[:start] + content[end+1:])
			return cleaned
		}
	}
	return content
}

// processToolCall executes a single tool call and returns the result message and record.
func (a *Agent) processToolCall(toolName string, args map[string]interface{}, start time.Time) (Message, ToolCallRecord) {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		errMsg := fmt.Sprintf("failed to marshal arguments: %v", err)
		if a.verbose {
			a.log("Argument marshal failed: %s", errMsg)
		}
		return Message{
				Role:     "tool",
				ToolName: toolName,
				Content:  errMsg,
			}, ToolCallRecord{
				Name:      toolName,
				Arguments: argsJSON,
				Error:     errMsg,
			}
	}

	if a.verbose {
		a.log("Executing tool: %s with args: %s", toolName, string(argsJSON))
	}

	// Skip validation to avoid contract name mismatches.
	callables := a.registry.AllCallables()
	fn, ok := callables[toolName]
	if !ok {
		errMsg := fmt.Sprintf("tool '%s' not found", toolName)
		return Message{
				Role:     "tool",
				ToolName: toolName,
				Content:  errMsg,
			}, ToolCallRecord{
				Name:      toolName,
				Arguments: argsJSON,
				Error:     errMsg,
			}
	}

	resultBytes, err := fn(argsJSON)
	if err != nil {
		errMsg := fmt.Sprintf("execution error: %v", err)
		return Message{
				Role:     "tool",
				ToolName: toolName,
				Content:  errMsg,
			}, ToolCallRecord{
				Name:      toolName,
				Arguments: argsJSON,
				Error:     errMsg,
			}
	}

	var toolResult tools.Result
	var resultData interface{}
	if err := json.Unmarshal(resultBytes, &toolResult); err == nil {
		resultData = toolResult
	} else {
		resultData = string(resultBytes)
	}

	if a.verbose {
		a.log("Tool result: %s", string(resultBytes))
	}

	return Message{
			Role:     "tool",
			ToolName: toolName,
			Content:  string(resultBytes),
		}, ToolCallRecord{
			Name:       toolName,
			Arguments:  argsJSON,
			Result:     resultData,
			DurationMS: time.Since(start).Milliseconds(),
		}
}

// RunWithHistory processes a user question with an explicit conversation history.
func (a *Agent) RunWithHistory(question string, history []Message) (AgentResult, error) {
	if a.verbose {
		a.log("User query: %s", question)
	}

	// Build system context if not already in history.
	var hasSystem bool
	for _, msg := range history {
		if msg.Role == "system" {
			hasSystem = true
			break
		}
	}
	if !hasSystem {
		systemPrompt, err := BuildSystemContext(a.registry)
		if err != nil {
			return AgentResult{}, fmt.Errorf("failed to build system context: %w", err)
		}
		history = append([]Message{{Role: "system", Content: systemPrompt}}, history...)
	}

	toolDefs := a.buildToolDefinitions()
	var toolCalls []ToolCallRecord
	iterations := 0
	toolCallCount := 0
	var lastAnswer string

	for turn := 0; turn < MaxAgentTurns; turn++ {
		iterations++
		resp, err := a.ollama.Chat(history, toolDefs)
		if err != nil {
			return AgentResult{Success: false, Error: err.Error()}, err
		}

		if a.verbose {
			a.log("Raw response: content=%q, tool_calls=%d", resp.Content, len(resp.ToolCalls))
		}

		// Clean content: remove any tool call JSON.
		cleanContent := stripToolCallFromContent(resp.Content)
		if cleanContent != "" {
			lastAnswer = cleanContent
		}

		history = append(history, *resp)

		// If there are explicit tool_calls, process them.
		if len(resp.ToolCalls) > 0 {
			for _, call := range resp.ToolCalls {
				start := time.Now()
				toolName := call.Function.Name
				argsMap := call.Function.Arguments
				toolMsg, record := a.processToolCall(toolName, argsMap, start)
				toolCalls = append(toolCalls, record)
				toolCallCount++
				history = append(history, toolMsg)
			}
			if toolCallCount >= MaxToolCalls {
				if a.verbose {
					a.log("Reached max tool calls (%d), breaking loop", MaxToolCalls)
				}
				if lastAnswer != "" {
					return AgentResult{
						Answer:     lastAnswer,
						ToolCalls:  toolCalls,
						Success:    true,
						Iterations: iterations,
					}, nil
				}
				fallbackResult := a.handleFallback(question)
				return fallbackResult, nil
			}
			continue
		}

		// No explicit tool_calls – try to extract a JSON tool call from content.
		if resp.Content != "" {
			if name, args, found := extractToolCallFromContent(resp.Content); found {
				if a.verbose {
					a.log("Detected tool call in content: %s", name)
				}
				start := time.Now()
				toolMsg, record := a.processToolCall(name, args, start)
				toolCalls = append(toolCalls, record)
				toolCallCount++
				history = append(history, toolMsg)
				if toolCallCount >= MaxToolCalls {
					if a.verbose {
						a.log("Reached max tool calls (%d), breaking loop", MaxToolCalls)
					}
					if lastAnswer != "" {
						return AgentResult{
							Answer:     lastAnswer,
							ToolCalls:  toolCalls,
							Success:    true,
							Iterations: iterations,
						}, nil
					}
					fallbackResult := a.handleFallback(question)
					return fallbackResult, nil
				}
				continue
			}
		}

		// No tool calls at all – final answer.
		answer := strings.TrimSpace(resp.Content)
		if answer == "" {
			if lastAnswer != "" {
				return AgentResult{
					Answer:     lastAnswer,
					ToolCalls:  toolCalls,
					Success:    true,
					Iterations: iterations,
				}, nil
			}
			fallbackResult := a.handleFallback(question)
			return fallbackResult, nil
		}
		// We have a final answer.
		return AgentResult{
			Answer:     answer,
			ToolCalls:  toolCalls,
			Success:    true,
			Iterations: iterations,
		}, nil
	}

	// If we exit the loop without returning, use the last answer if any.
	if lastAnswer != "" {
		return AgentResult{
			Answer:     lastAnswer,
			ToolCalls:  toolCalls,
			Success:    true,
			Iterations: iterations,
		}, nil
	}
	return AgentResult{
		Success:    false,
		Error:      fmt.Sprintf("agent reached max turns (%d) without final answer", MaxAgentTurns),
		ToolCalls:  toolCalls,
		Iterations: iterations,
	}, nil
}

// log prints verbose messages.
func (a *Agent) log(msg string, args ...interface{}) {
	if a.verbose {
		fmt.Printf("[Agent] %s\n", fmt.Sprintf(msg, args...))
	}
}

// buildToolDefinitions returns the list of tools the agent can use.
func (a *Agent) buildToolDefinitions() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "list_scenes",
				Description: "Lists all scenes in the Unity project.",
				Parameters:  Parameters{Type: "object", Properties: map[string]Property{}, Required: []string{}},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "describe_scene",
				Description: "Returns detailed information about a specific scene.",
				Parameters: Parameters{
					Type: "object",
					Properties: map[string]Property{
						"scene_id": {Type: "string", Description: "The unique ID of the scene to describe."},
					},
					Required: []string{"scene_id"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "list_gameobjects",
				Description: "Lists all GameObjects in a given scene.",
				Parameters: Parameters{
					Type: "object",
					Properties: map[string]Property{
						"scene_id": {Type: "string", Description: "The ID of the scene."},
					},
					Required: []string{"scene_id"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "list_components",
				Description: "Lists all components attached to a specific GameObject.",
				Parameters: Parameters{
					Type: "object",
					Properties: map[string]Property{
						"game_object_id": {Type: "string", Description: "The ID of the GameObject."},
					},
					Required: []string{"game_object_id"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "list_scripts",
				Description: "Lists all scripts used in components within a scene.",
				Parameters: Parameters{
					Type: "object",
					Properties: map[string]Property{
						"scene_id": {Type: "string", Description: "The ID of the scene."},
					},
					Required: []string{"scene_id"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "list_assets",
				Description: "Lists all assets (textures, models, audio, etc.) in the project.",
				Parameters:  Parameters{Type: "object", Properties: map[string]Property{}, Required: []string{}},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "project_stats",
				Description: "Returns aggregated statistics for the whole project.",
				Parameters:  Parameters{Type: "object", Properties: map[string]Property{}, Required: []string{}},
			},
		},
		{
			Type: "function",
			Function: FunctionDef{
				Name:        "scene_stats",
				Description: "Returns statistics for a specific scene.",
				Parameters: Parameters{
					Type: "object",
					Properties: map[string]Property{
						"scene_id": {Type: "string", Description: "The ID of the scene."},
					},
					Required: []string{"scene_id"},
				},
			},
		},
	}
}

// handleFallback provides direct answers without AI for common queries.
// It handles scripts, components, assets, scene stats, project stats, and the original fallbacks.
func (a *Agent) handleFallback(question string) AgentResult {
	q := strings.ToLower(question)
	var toolCalls []ToolCallRecord
	var answer string

	// ---- Specific queries first ----

	// 1. Scripts
	if strings.Contains(q, "script") && (strings.Contains(q, "list") || strings.Contains(q, "used") || strings.Contains(q, "what")) {
		listRes := a.registry.Scene.ListScenes()
		if listRes.Success {
			data, ok := listRes.Data.(tools.ListScenesResponse)
			if ok && len(data.Scenes) > 0 {
				sceneID := data.Scenes[0].ID
				scriptRes := a.registry.Script.ListScripts(sceneID)
				if scriptRes.Success {
					scriptData, ok := scriptRes.Data.(tools.ListScriptsResponse)
					if ok {
						var sb strings.Builder
						sb.WriteString(fmt.Sprintf("Found %d scripts in scene '%s':\n", len(scriptData.Scripts), data.Scenes[0].Name))
						for _, s := range scriptData.Scripts {
							sb.WriteString(fmt.Sprintf("  - %s (ID: %s)\n", s.Name, s.ID))
						}
						answer = sb.String()
						toolCalls = append(toolCalls, ToolCallRecord{
							Name:      "list_scripts",
							Arguments: json.RawMessage(`{"scene_id":"` + sceneID + `"}`),
							Result:    scriptData,
						})
					}
				} else {
					answer = "Failed to list scripts: " + scriptRes.Error.Message
				}
			} else {
				answer = "No scenes found to list scripts."
			}
		} else {
			answer = "Failed to list scenes for scripts."
		}
		return AgentResult{Answer: answer, ToolCalls: toolCalls, Success: true, Iterations: 1}
	}

	// 2. Components
	if strings.Contains(q, "component") && (strings.Contains(q, "list") || strings.Contains(q, "what")) {
		listRes := a.registry.Scene.ListScenes()
		if listRes.Success {
			data, ok := listRes.Data.(tools.ListScenesResponse)
			if ok && len(data.Scenes) > 0 {
				sceneID := data.Scenes[0].ID
				goRes := a.registry.GameObject.ListGameObjects(sceneID)
				if goRes.Success {
					goData, ok := goRes.Data.(tools.ListGameObjectsResponse)
					if ok && len(goData.GameObjects) > 0 {
						var targetID string
						var targetName string
						for _, g := range goData.GameObjects {
							if strings.Contains(q, strings.ToLower(g.Name)) {
								targetID = g.ID
								targetName = g.Name
								break
							}
						}
						if targetID == "" {
							targetID = goData.GameObjects[0].ID
							targetName = goData.GameObjects[0].Name
						}
						compRes := a.registry.Component.ListComponents(targetID)
						if compRes.Success {
							compData, ok := compRes.Data.(tools.ListComponentsResponse)
							if ok {
								var sb strings.Builder
								sb.WriteString(fmt.Sprintf("Found %d components on GameObject '%s':\n", len(compData.Components), targetName))
								for _, c := range compData.Components {
									sb.WriteString(fmt.Sprintf("  - %s (ID: %s)\n", c.Type, c.ID))
								}
								answer = sb.String()
								toolCalls = append(toolCalls, ToolCallRecord{
									Name:      "list_components",
									Arguments: json.RawMessage(`{"game_object_id":"` + targetID + `"}`),
									Result:    compData,
								})
							}
						} else {
							answer = "Failed to list components: " + compRes.Error.Message
						}
					} else {
						answer = "No GameObjects found in the scene."
					}
				} else {
					answer = "Failed to list GameObjects: " + goRes.Error.Message
				}
			} else {
				answer = "No scenes found."
			}
		} else {
			answer = "Failed to list scenes."
		}
		return AgentResult{Answer: answer, ToolCalls: toolCalls, Success: true, Iterations: 1}
	}

	// 3. Assets
	if strings.Contains(q, "asset") && (strings.Contains(q, "list") || strings.Contains(q, "available") || strings.Contains(q, "what")) {
		res := a.registry.Asset.ListAssets()
		if res.Success {
			data, ok := res.Data.(tools.ListAssetsResponse)
			if ok {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("Found %d assets:\n", len(data.Assets)))
				for _, a := range data.Assets {
					sb.WriteString(fmt.Sprintf("  - %s (ID: %s, Type: %s)\n", a.Name, a.ID, a.Type))
				}
				answer = sb.String()
				toolCalls = append(toolCalls, ToolCallRecord{
					Name:      "list_assets",
					Arguments: json.RawMessage(`{}`),
					Result:    data,
				})
			}
		} else {
			answer = "Failed to list assets: " + res.Error.Message
		}
		return AgentResult{Answer: answer, ToolCalls: toolCalls, Success: true, Iterations: 1}
	}

	// 4. Scene stats (specific scene)
	if (strings.Contains(q, "scene stats") || strings.Contains(q, "statistics") && strings.Contains(q, "scene")) && !strings.Contains(q, "project") {
		listRes := a.registry.Scene.ListScenes()
		if listRes.Success {
			data, ok := listRes.Data.(tools.ListScenesResponse)
			if ok && len(data.Scenes) > 0 {
				sceneID := data.Scenes[0].ID
				statsRes := a.registry.Stats.SceneStats(sceneID)
				if statsRes.Success {
					statsData, ok := statsRes.Data.(tools.SceneStatsResponse)
					if ok {
						s := statsData.Statistics
						answer = fmt.Sprintf("Scene Statistics (ID: %s):\n  GameObjects: %d\n  Components: %d\n  Scripts: %d\n  Prefabs: %d\n  Assets: %d\n  Fields: %d",
							sceneID, s.GameObjectCount, s.ComponentCount, s.ScriptCount, s.PrefabCount, s.AssetCount, s.FieldCount)
						toolCalls = append(toolCalls, ToolCallRecord{
							Name:      "scene_stats",
							Arguments: json.RawMessage(`{"scene_id":"` + sceneID + `"}`),
							Result:    statsData,
						})
					}
				} else {
					answer = "Failed to get scene stats: " + statsRes.Error.Message
				}
			} else {
				answer = "No scenes found."
			}
		} else {
			answer = "Failed to list scenes."
		}
		return AgentResult{Answer: answer, ToolCalls: toolCalls, Success: true, Iterations: 1}
	}

	// 5. Project statistics (moved before "list scenes" to catch "show project stats")
	if (strings.Contains(q, "stats") || strings.Contains(q, "statistics")) && strings.Contains(q, "project") {
		res := a.registry.Stats.ProjectStats()
		if res.Success {
			data, ok := res.Data.(tools.ProjectStatsResponse)
			if ok {
				s := data.Summary
				answer = fmt.Sprintf(
					"Project Statistics:\n  Scenes: %d\n  GameObjects: %d\n  Components: %d\n  Unique Scripts: %d\n  Assets: %d\n  Serialized Fields: %d",
					s.TotalScenes, s.TotalGameObjects, s.TotalComponents, s.UniqueScripts, s.TotalAssets, s.TotalSerializedFields)
				toolCalls = append(toolCalls, ToolCallRecord{
					Name:      "project_stats",
					Arguments: json.RawMessage(`{}`),
					Result:    data,
				})
			}
		} else {
			answer = "Failed to get project statistics: " + res.Error.Message
		}
		return AgentResult{Answer: answer, ToolCalls: toolCalls, Success: true, Iterations: 1}
	}

	// ---- Broad fallbacks ----

	// List scenes
	if strings.Contains(q, "list") && (strings.Contains(q, "scene") || strings.Contains(q, "scenes")) {
		res := a.registry.Scene.ListScenes()
		if res.Success {
			data, ok := res.Data.(tools.ListScenesResponse)
			if ok && len(data.Scenes) > 0 {
				var sb strings.Builder
				sb.WriteString(fmt.Sprintf("Found %d scenes:\n", len(data.Scenes)))
				for _, s := range data.Scenes {
					sb.WriteString(fmt.Sprintf("  - %s (ID: %s)\n", s.Name, s.ID))
				}
				answer = sb.String()
				toolCalls = append(toolCalls, ToolCallRecord{
					Name:      "list_scenes",
					Arguments: json.RawMessage(`{}`),
					Result:    data,
				})
			} else {
				answer = "No scenes found."
			}
		} else {
			answer = "Failed to list scenes: " + res.Error.Message
		}
		return AgentResult{Answer: answer, ToolCalls: toolCalls, Success: true, Iterations: 1}
	}

	// Describe a scene (uses first scene)
	if strings.Contains(q, "describe") && strings.Contains(q, "scene") {
		listRes := a.registry.Scene.ListScenes()
		if listRes.Success {
			data, ok := listRes.Data.(tools.ListScenesResponse)
			if ok && len(data.Scenes) > 0 {
				sceneID := data.Scenes[0].ID
				descRes := a.registry.Scene.DescribeScene(sceneID)
				if descRes.Success {
					descData, ok := descRes.Data.(tools.DescribeSceneResponse)
					if ok {
						answer = fmt.Sprintf("Scene: %s (ID: %s)\n  GameObjects: %d\n  Components: %d\n  Scripts: %d",
							descData.Scene.Name, descData.Scene.ID,
							descData.Statistics.GameObjectCount,
							descData.Statistics.ComponentCount,
							descData.Statistics.ScriptCount)
						toolCalls = append(toolCalls, ToolCallRecord{
							Name:      "describe_scene",
							Arguments: json.RawMessage(`{"scene_id":"` + sceneID + `"}`),
							Result:    descData,
						})
					}
				} else {
					answer = "Failed to describe scene: " + descRes.Error.Message
				}
			} else {
				answer = "No scenes found to describe."
			}
		} else {
			answer = "Failed to list scenes for description."
		}
		return AgentResult{Answer: answer, ToolCalls: toolCalls, Success: true, Iterations: 1}
	}

	// List GameObjects (uses first scene)
	if strings.Contains(q, "gameobject") && (strings.Contains(q, "list") || strings.Contains(q, "gameobjects")) {
		listRes := a.registry.Scene.ListScenes()
		if listRes.Success {
			data, ok := listRes.Data.(tools.ListScenesResponse)
			if ok && len(data.Scenes) > 0 {
				sceneID := data.Scenes[0].ID
				goRes := a.registry.GameObject.ListGameObjects(sceneID)
				if goRes.Success {
					goData, ok := goRes.Data.(tools.ListGameObjectsResponse)
					if ok {
						var sb strings.Builder
						sb.WriteString(fmt.Sprintf("Found %d GameObjects in scene '%s':\n", len(goData.GameObjects), data.Scenes[0].Name))
						for _, g := range goData.GameObjects {
							sb.WriteString(fmt.Sprintf("  - %s (ID: %s)\n", g.Name, g.ID))
						}
						answer = sb.String()
						toolCalls = append(toolCalls, ToolCallRecord{
							Name:      "list_gameobjects",
							Arguments: json.RawMessage(`{"scene_id":"` + sceneID + `"}`),
							Result:    goData,
						})
					}
				} else {
					answer = "Failed to list GameObjects: " + goRes.Error.Message
				}
			} else {
				answer = "No scenes found."
			}
		} else {
			answer = "Failed to list scenes."
		}
		return AgentResult{Answer: answer, ToolCalls: toolCalls, Success: true, Iterations: 1}
	}

	// If nothing matched
	return AgentResult{
		Success:    false,
		Error:      "I don't understand that question. Try asking about scenes, GameObjects, components, scripts, assets, or statistics.",
		Iterations: 1,
	}
}
