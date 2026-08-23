package ai

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
)

const MaxAgentTurns = 8

// Agent is the main AI agent that orchestrates conversation and tool calls.
type Agent struct {
	ollama   *OllamaClient
	registry *tools.ToolRegistry
	verbose  bool
	cache    *Cache
}

// NewAgent creates a new Agent.
func NewAgent(ollama *OllamaClient, registry *tools.ToolRegistry, verbose bool) *Agent {
	return &Agent{
		ollama:   ollama,
		registry: registry,
		verbose:  verbose,
		cache:    NewCache(5 * time.Minute), // cache TTL 5 minutes
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

	// Prepare tools.
	toolDefs := a.buildToolDefinitions()
	callables := a.registry.AllCallables()

	// --- Step 1: Deterministic router ---
	decision := RouteQuery(question, a.registry)
	if a.verbose {
		a.log("[Router] Decision: %+v", decision)
	}
	if decision != nil && decision.Confidence >= 0.9 {
		if a.verbose {
			a.log("[Router] Direct tool: %s (confidence %.2f)", decision.Tool, decision.Confidence)
		}
		// Try cache first.
		cacheKey := decision.Tool + ":" + string(decision.Arguments)
		cached := a.cache.Get(cacheKey)
		if cached != nil {
			if a.verbose {
				a.log("[Cache] Hit for key: %s", cacheKey)
			}
			formatted, err := formatToolResult(decision.Tool, cached)
			if err != nil {
				if a.verbose {
					a.log("[Format] Error formatting cached result: %v", err)
				}
				formatted = string(cached)
			}
			return AgentResult{
				Answer: formatted,
				ToolCalls: []ToolCallRecord{{
					Name:      decision.Tool,
					Arguments: decision.Arguments,
					Result:    json.RawMessage(cached),
				}},
				Success:    true,
				Iterations: 1,
			}, nil
		}
		if a.verbose {
			a.log("[Cache] Miss for key: %s", cacheKey)
		}

		// Execute tool.
		fn, ok := callables[decision.Tool]
		if !ok {
			return AgentResult{Success: false, Error: fmt.Sprintf("router selected unknown tool: %s", decision.Tool)}, nil
		}
		resultBytes, err := fn(decision.Arguments)
		if err != nil {
			return AgentResult{Success: false, Error: err.Error()}, nil
		}
		// Store in cache.
		a.cache.Set(cacheKey, resultBytes)

		// Format result.
		formatted, err := formatToolResult(decision.Tool, resultBytes)
		if err != nil {
			if a.verbose {
				a.log("[Format] Error formatting result: %v", err)
			}
			formatted = string(resultBytes)
		}
		var toolResult tools.Result
		var resultData interface{}
		if err := json.Unmarshal(resultBytes, &toolResult); err == nil {
			resultData = toolResult
		} else {
			resultData = string(resultBytes)
		}
		return AgentResult{
			Answer: formatted,
			ToolCalls: []ToolCallRecord{{
				Name:      decision.Tool,
				Arguments: decision.Arguments,
				Result:    resultData,
			}},
			Success:    true,
			Iterations: 1,
		}, nil
	}

	// --- Step 2: Keyword fallback (safety net) ---
	lower := strings.ToLower(question)
	fallbackTool := ""
	switch {
	case strings.Contains(lower, "asset") || strings.Contains(lower, "assets"):
		fallbackTool = "list_assets"
	case strings.Contains(lower, "scene") || strings.Contains(lower, "scenes"):
		fallbackTool = "list_scenes"
	case strings.Contains(lower, "script") || strings.Contains(lower, "scripts"):
		fallbackTool = "list_all_scripts"
	case strings.Contains(lower, "gameobject") || strings.Contains(lower, "game objects") || strings.Contains(lower, "object") && !strings.Contains(lower, "component"):
		fallbackTool = "list_all_gameobjects"
	case strings.Contains(lower, "stat") || strings.Contains(lower, "stats") || strings.Contains(lower, "project summary"):
		fallbackTool = "project_stats"
	}
	if fallbackTool != "" {
		if a.verbose {
			a.log("[Fallback] Direct tool: %s (keyword match)", fallbackTool)
		}
		// Use cache.
		cacheKey := fallbackTool + ":{}"
		cached := a.cache.Get(cacheKey)
		if cached != nil {
			if a.verbose {
				a.log("[Cache] Hit for key: %s", cacheKey)
			}
			formatted, err := formatToolResult(fallbackTool, cached)
			if err != nil {
				if a.verbose {
					a.log("[Format] Error formatting cached result: %v", err)
				}
				formatted = string(cached)
			}
			return AgentResult{
				Answer: formatted,
				ToolCalls: []ToolCallRecord{{
					Name:      fallbackTool,
					Arguments: json.RawMessage(`{}`),
					Result:    json.RawMessage(cached),
				}},
				Success:    true,
				Iterations: 1,
			}, nil
		}
		if a.verbose {
			a.log("[Cache] Miss for key: %s", cacheKey)
		}

		fn, ok := callables[fallbackTool]
		if !ok {
			return AgentResult{Success: false, Error: fmt.Sprintf("fallback tool %s not found", fallbackTool)}, nil
		}
		resultBytes, err := fn(json.RawMessage(`{}`))
		if err != nil {
			return AgentResult{Success: false, Error: err.Error()}, nil
		}
		a.cache.Set(cacheKey, resultBytes)
		formatted, err := formatToolResult(fallbackTool, resultBytes)
		if err != nil {
			if a.verbose {
				a.log("[Format] Error formatting result: %v", err)
			}
			formatted = string(resultBytes)
		}
		var toolResult tools.Result
		var resultData interface{}
		if err := json.Unmarshal(resultBytes, &toolResult); err == nil {
			resultData = toolResult
		} else {
			resultData = string(resultBytes)
		}
		return AgentResult{
			Answer: formatted,
			ToolCalls: []ToolCallRecord{{
				Name:      fallbackTool,
				Arguments: json.RawMessage(`{}`),
				Result:    resultData,
			}},
			Success:    true,
			Iterations: 1,
		}, nil
	}

	// --- Step 3: AI Agent Loop (ambiguous/complex queries) ---
	if a.verbose {
		a.log("[Agent] No deterministic route; delegating to LLM.")
	}
	var toolCalls []ToolCallRecord
	iterations := 0

	for turn := 0; turn < MaxAgentTurns; turn++ {
		iterations++
		resp, err := a.ollama.Chat(history, toolDefs)
		if err != nil {
			return AgentResult{Success: false, Error: err.Error()}, err
		}

		if a.verbose {
			a.log("Received response from Ollama")
			if len(resp.ToolCalls) > 0 {
				a.log("Tool calls requested: %d", len(resp.ToolCalls))
			}
		}

		history = append(history, *resp)

		if len(resp.ToolCalls) == 0 {
			answer := strings.TrimSpace(resp.Content)
			if answer == "" {
				return AgentResult{
					Success:    false,
					Error:      "model returned empty response",
					Iterations: iterations,
				}, nil
			}
			return AgentResult{
				Answer:     answer,
				ToolCalls:  toolCalls,
				Success:    true,
				Iterations: iterations,
			}, nil
		}

		// Process each tool call.
		for _, call := range resp.ToolCalls {
			start := time.Now()
			toolName := call.Function.Name
			rawArgsMap := call.Function.Arguments

			rawArgsBytes, marshalErr := json.Marshal(rawArgsMap)
			if marshalErr != nil {
				errMsg := fmt.Sprintf("failed to marshal arguments: %v", marshalErr)
				if a.verbose {
					a.log("Marshaling failed: %s", errMsg)
				}
				history = append(history, Message{
					Role:     "tool",
					ToolName: toolName,
					Content:  errMsg,
				})
				toolCalls = append(toolCalls, ToolCallRecord{
					Name:      toolName,
					Arguments: json.RawMessage(rawArgsBytes),
					Error:     errMsg,
				})
				continue
			}

			if a.verbose {
				a.log("Executing tool: %s with args: %s", toolName, string(rawArgsBytes))
			}

			// Validate (pass raw JSON bytes).
			if err := a.validateToolCall(toolName, json.RawMessage(rawArgsBytes)); err != nil {
				errMsg := err.Error()
				if a.verbose {
					a.log("Validation failed: %s", errMsg)
				}
				history = append(history, Message{
					Role:     "tool",
					ToolName: toolName,
					Content:  errMsg,
				})
				toolCalls = append(toolCalls, ToolCallRecord{
					Name:      toolName,
					Arguments: json.RawMessage(rawArgsBytes),
					Error:     errMsg,
				})
				continue
			}

			fn, ok := callables[toolName]
			if !ok {
				errMsg := fmt.Sprintf("tool '%s' not found", toolName)
				history = append(history, Message{
					Role:     "tool",
					ToolName: toolName,
					Content:  errMsg,
				})
				toolCalls = append(toolCalls, ToolCallRecord{
					Name:      toolName,
					Arguments: json.RawMessage(rawArgsBytes),
					Error:     errMsg,
				})
				continue
			}

			// Check cache for LLM-selected tool calls (we can also cache them).
			cacheKey := toolName + ":" + string(rawArgsBytes)
			cached := a.cache.Get(cacheKey)
			var resultBytes []byte
			if cached != nil {
				if a.verbose {
					a.log("[Cache] Hit for key: %s", cacheKey)
				}
				resultBytes = cached
			} else {
				if a.verbose {
					a.log("[Cache] Miss for key: %s", cacheKey)
				}
				var err error
				resultBytes, err = fn(json.RawMessage(rawArgsBytes))
				if err != nil {
					errMsg := fmt.Sprintf("execution error: %v", err)
					history = append(history, Message{
						Role:     "tool",
						ToolName: toolName,
						Content:  errMsg,
					})
					toolCalls = append(toolCalls, ToolCallRecord{
						Name:      toolName,
						Arguments: json.RawMessage(rawArgsBytes),
						Error:     errMsg,
					})
					continue
				}
				a.cache.Set(cacheKey, resultBytes)
			}

			// Parse result for structured logging.
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

			history = append(history, Message{
				Role:     "tool",
				ToolName: toolName,
				Content:  string(resultBytes),
			})
			toolCalls = append(toolCalls, ToolCallRecord{
				Name:       toolName,
				Arguments:  json.RawMessage(rawArgsBytes),
				Result:     resultData,
				DurationMS: time.Since(start).Milliseconds(),
			})
		}
	}

	return AgentResult{
		Success:    false,
		Error:      fmt.Sprintf("agent reached max turns (%d) without final answer", MaxAgentTurns),
		ToolCalls:  toolCalls,
		Iterations: iterations,
	}, nil
}

// validateToolCall checks tool existence, JSON validity, and schema conformance.
func (a *Agent) validateToolCall(toolName string, args json.RawMessage) error {
	contract, ok := a.registry.GetContract(toolName)
	if !ok {
		return fmt.Errorf("unknown tool: %s", toolName)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(args, &parsed); err != nil {
		return fmt.Errorf("invalid JSON arguments: %w", err)
	}

	schema := contract.InputSchema
	for _, required := range schema.Required {
		if _, exists := parsed[required]; !exists {
			return fmt.Errorf("missing required argument: %s", required)
		}
	}
	return nil
}

// log prints verbose messages.
func (a *Agent) log(msg string, args ...interface{}) {
	if a.verbose {
		fmt.Printf("[Agent] %s\n", fmt.Sprintf(msg, args...))
	}
}

// buildToolDefinitions constructs the tool definitions for the Ollama API.
func (a *Agent) buildToolDefinitions() []Tool {
	var tools []Tool
	for name := range a.registry.AllCallables() {
		contract, ok := a.registry.GetContract(name)
		if !ok {
			continue
		}
		params := Parameters{
			Type:     "object",
			Required: contract.InputSchema.Required,
		}
		if contract.InputSchema.Properties != nil {
			params.Properties = make(map[string]Property)
			for key, prop := range contract.InputSchema.Properties {
				params.Properties[key] = Property{
					Type:        prop.Type,
					Description: prop.Description,
				}
			}
		}
		tools = append(tools, Tool{
			Type: "function",
			Function: FunctionDef{
				Name:        name,
				Description: contract.Description,
				Parameters:  params,
			},
		})
	}
	return tools
}
