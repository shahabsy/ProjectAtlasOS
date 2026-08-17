package ai

import (
	"encoding/json"
)

// ToolCallRecord logs a single tool invocation.
type ToolCallRecord struct {
	ID         string          `json:"id,omitempty"`
	Name       string          `json:"name"`
	Arguments  json.RawMessage `json:"arguments,omitempty"`
	Result     interface{}     `json:"result,omitempty"`
	Error      string          `json:"error,omitempty"`
	DurationMS int64           `json:"duration_ms,omitempty"`
}

// AgentResult is the complete output of an agent run.
type AgentResult struct {
	Answer     string           `json:"answer"`
	ToolCalls  []ToolCallRecord `json:"tool_calls"`
	Success    bool             `json:"success"`
	Iterations int              `json:"iterations"`
	Error      string           `json:"error,omitempty"`
}
