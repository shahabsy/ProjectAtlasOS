// Package tools provides deterministic, typed capabilities for interacting with the Unity knowledge graph.
// Tools are the foundational layer used by CLI, HTTP API, and AI components.
package tools

// Parameter describes a tool input for AI function-calling.
type Parameter struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "string", "integer", "boolean", "object", "array"
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Enum        []string `json:"enum,omitempty"` // optional allowed values
}

// ToolMetadata contains metadata for LLM function-calling registration.
type ToolMetadata struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  []Parameter `json:"parameters"`
}

// ToolMetaProvider is implemented by all tool structs.
type ToolMetaProvider interface {
	Meta() ToolMetadata
}
