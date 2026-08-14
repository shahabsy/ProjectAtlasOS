package models

// Asset represents any Unity asset (texture, material, prefab, etc.).
type Asset struct {
	ID       string `json:"id"`
	GUID     string `json:"guid"`
	Category string `json:"category"` // "asset"
	Type     string `json:"type"`     // "Texture2D", "Material", "AudioClip"
	Name     string `json:"name"`
	Path     string `json:"path"`
}

// Prefab is a specialised asset (kept separate for clarity).
type Prefab struct {
	ID   string `json:"id"`
	GUID string `json:"guid"`
	Name string `json:"name"`
	Path string `json:"path"`
}
