package models

// DataAvailability indicates which graph domains are populated.
type DataAvailability struct {
	Scenes           bool `json:"scenes"`
	GameObjects      bool `json:"game_objects"`
	Components       bool `json:"components"`
	Scripts          bool `json:"scripts"`
	SerializedFields bool `json:"serialized_fields"`
	AssetReferences  bool `json:"asset_references"`
	Prefabs          bool `json:"prefabs"`
	// Add more as indexing expands
}
