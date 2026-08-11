package models

type Asset struct {
	ID   string `json:"id"`
	GUID string `json:"guid"`
	Type string `json:"type"` // Texture, Material, Prefab, shader, etc.
	Name string `json:"name"`
	Path string `json:"path"`
}

type Prefab struct {
	ID   string `json:"id"`
	GUID string `json:"guid"`
	Name string `json:"name"`
	Path string `json:"path"`
}
