package models

// GameObject represents a Unity GameObject node.
type GameObject struct {
	ID       string `json:"id"`
	GlobalID string `json:"global_id"`
	Name     string `json:"name"`
	SceneID  string `json:"scene_id"`
	ParentID string `json:"parent_id"`
	// These can be populated on request (avoid cirular references)
	Children   []*GameObject `json:"children,omitempty"`
	Components []*Component  `json:"components,omitempty"`
}

// Hierarchy is a lightweight representation used in tree views.
type HierarchyNode struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Type     string          `json:"type"`
	Children []HierarchyNode `json:"children,omitempty"`
}
