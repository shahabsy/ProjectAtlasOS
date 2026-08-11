package models

type Script struct {
	ID        string `json:"id"`
	GUID      string `json:"guid"`
	Name      string `json:"name"`
	ClassName string `json:"class_name"`
	Namespace string `json:"namespace"`
	Path      string `json:"path"`
}
