package models

type Edge struct {
	SourceID     string `json:"source_id"`
	TargetID     string `json:"target_id"`
	Relationship string `json:"relationship"` // CONTAINS, HAS_COMPONENT, etc.
	Metadata     string `json:"metadata,omitempty"`
}
