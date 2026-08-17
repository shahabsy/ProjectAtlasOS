package tools

// Contract defines the formal specification of a tool.
type Contract struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema Schema `json:"input_schema"`
	ReadOnly    bool   `json:"read_only"`
}

// Schema defines the expected input structure.
type Schema struct {
	Type       string              `json:"type"` // "object"
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property defines a single parameter.
type Property struct {
	Type        string   `json:"type"` // "string", "integer", "boolean", "object", "array"
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}
