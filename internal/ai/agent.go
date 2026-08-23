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
}

// NewAgent creates a new Agent.
func NewAgent(ollama *OllamaClient, registry *tools.ToolRegistry, verbose bool) *Agent {
	return &Agent{
		ollama:   ollama,
		registry: registry,
		verbose:  verbose,
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
			rawArgsMap := call.Function.Arguments // map[string]any

			// Marshal to JSON bytes for passing to tool functions.
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
					Arguments: json.RawMessage(rawArgsBytes), // rawArgsBytes is nil if error, but we can still record
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

			// Execute tool with the JSON bytes.
			resultBytes, err := fn(json.RawMessage(rawArgsBytes))
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

	// Unmarshal into temporary map for validation.
	var parsed map[string]interface{}
	if err := json.Unmarshal(args, &parsed); err != nil {
		return fmt.Errorf("invalid JSON arguments: %w", err)
	}

	// Validate required fields from schema.
	schema := contract.InputSchema
	for _, required := range schema.Required {
		if _, exists := parsed[required]; !exists {
			return fmt.Errorf("missing required argument: %s", required)
		}
	}
	// Additional type checks can be added here.
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
		// Convert tools.Schema to ai.Parameters.
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
