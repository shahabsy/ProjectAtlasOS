package models

// Scene respresents a unity scene in the knowledge graph
type Scene struct {
	ID          string `json:"id"`
	GUID        string `json:"guid"`
	Name        string `json:"name"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
	LastIndexed int64  `json:"last_indexed"`
}

// SceneStatistics holds aggregated counts for a scene.
type SceneStatistics struct {
	GameObjectCount int `json:"game_object_count"`
	ComponentCount  int `json:"component_count"`
	ScriptCount     int `json:"script_count"`
	PrefabCount     int `json:"prefab_count"`
	AssetCount      int `json:"asset_count"`
	FieldCount      int `json:"field_count"`
	// Later we can Materials, Shaders, Textures etc. if need be
}
