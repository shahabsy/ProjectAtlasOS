package models

type Component struct {
	ID           string `json:"id"`
	GlobalID     string `json:"global_id"`
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
	GameObjectID string `json:"game_object_id"`
	ScriptID     string `json:"script_id,omitempty"` // Already present
	// Related data (populated when requested)
	Script *Script            `json:"script,omitempty"`
	Fields []*SerializedField `json:"fields,omitempty"`
	Assets []*Asset           `json:"assets,omitempty"`
}

type SerializedField struct {
	ID          string `json:"id"`
	ComponentID string `json:"component_id"`
	Name        string `json:"name"`
	Value       string `json:"value,omitempty"`
	ReferenceID string `json:"reference_id,omitempty"`
}
