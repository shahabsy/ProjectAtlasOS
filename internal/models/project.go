package models

type ProjectSummary struct {
	TotalScenes           int   `json:"total_scenes"`
	TotalGameObjects      int   `json:"total_game_objects"`
	TotalComponents       int   `json:"total_components"`
	TotalScripts          int   `json:"total_scripts"`
	TotalAssets           int   `json:"total_assets"`
	UniqueScripts         int   `json:"unique_scripts"`
	TotalSerializedFields int   `json:"total_serialized_fields"`
	LastIndexed           int64 `json:"last_indexed"` // unix timestamp in
}
