package models

// Component represents a Unity component node.
type Component struct {
	ID           string `json:"id"`
	GlobalID     string `json:"global_id"`
	Category     string `json:"category"` // always "component"
	Type         string `json:"type"`     // "Transform", "Rigidbody", etc.
	Enabled      bool   `json:"enabled"`
	GameObjectID string `json:"game_object_id"`
	ScriptID     string `json:"script_id,omitempty"`
	// Related data (populated when requested)
	Script *Script            `json:"script,omitempty"`
	Fields []*SerializedField `json:"fields,omitempty"`
	Assets []*Asset           `json:"assets,omitempty"`
}

// SerializedField represents a serialized field of a component.
type SerializedField struct {
	ID          string `json:"id"`
	ComponentID string `json:"component_id"`
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	ReferenceID string `json:"reference_id,omitempty"`
}
