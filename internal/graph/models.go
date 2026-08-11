package graph

type NodeType string

const (
	NodeTypeScene      NodeType = "scene"
	NodeTypeGameObject NodeType = "gameobject"
	NodeTypeComponent  NodeType = "component"
	NodeTypeScript     NodeType = "script"
	NodeTypePrefab     NodeType = "prefab"
	NodeTypeMaterial   NodeType = "material"
	NodeTypeShader     NodeType = "shader"
	NodeTypeTexture    NodeType = "texture"
	NodeTypeSerialized NodeType = "serialized_field"
)

// SceneInfo represents a Unity scene.
type SceneInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	GUID        string `json:"guid"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	LastIndexed int64  `json:"last_indexed"`
}

// GameObjectInfo represents a GameObject node.
type GameObjectInfo struct {
	ID       string `json:"id"`
	GlobalId string `json:"global_id"`
	Name     string `json:"name"`
	SceneID  string `json:"scene_id"`
	ParentID string `json:"parent_id,omitempty"`
}

// ComponentInfo represents a component node.
type ComponentInfo struct {
	ID           string `json:"id"`
	GlobalId     string `json:"global_id"`
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
	GameObjectID string `json:"game_object_id"`
	ScriptID     string `json:"script_id,omitempty"`
}

// ScriptInfo represents a script node.
type ScriptInfo struct {
	ID        string `json:"id"`
	GUID      string `json:"guid"`
	Name      string `json:"name"`
	ClassName string `json:"class_name"`
	Namespace string `json:"namespace"`
	Path      string `json:"path"`
}

// PrefabInfo represents a prefab node.
type PrefabInfo struct {
	ID   string `json:"id"`
	GUID string `json:"guid"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// AssetInfo represents an asset node.
type AssetInfo struct {
	ID   string `json:"id"`
	GUID string `json:"guid"`
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// EdgeInfo represents a relationship between two nodes.
type EdgeInfo struct {
	SourceID     string `json:"source_id"`
	TargetID     string `json:"target_id"`
	Relationship string `json:"relationship"`
	Metadata     string `json:"metadata,omitempty"`
}

// SceneStats provides aggregated statistics for a scene.
type SceneStats struct {
	GameObjectCount int `json:"game_object_count"`
	ComponentCount  int `json:"component_count"`
	ScriptCount     int `json:"script_count"`
	PrefabCount     int `json:"prefab_count"`
	AssetCount      int `json:"asset_count"`
}
