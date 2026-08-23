package ai

import (
	"encoding/json"
	"fmt"
	"strings"
)

// formatToolResult formats the raw JSON result from a tool into human‑readable text.
func formatToolResult(toolName string, resultBytes []byte) (string, error) {
	switch toolName {
	case "list_assets":
		return formatAssets(resultBytes)
	case "list_scenes":
		return formatScenes(resultBytes)
	case "list_all_gameobjects", "list_gameobjects":
		return formatGameObjects(resultBytes)
	case "list_all_scripts", "list_scripts":
		return formatScripts(resultBytes)
	case "project_stats":
		return formatProjectStats(resultBytes)
	case "describe_scene", "scene_stats":
		return formatSceneStats(resultBytes)
	default:
		// fallback: return raw JSON
		return string(resultBytes), nil
	}
}

func formatAssets(data []byte) (string, error) {
	var result struct {
		Data struct {
			Assets []struct {
				Name string `json:"name"`
				Type string `json:"type"`
				ID   string `json:"id"`
				GUID string `json:"guid"`
			} `json:"assets"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	if len(result.Data.Assets) == 0 {
		return "No assets found.", nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Assets (%d):\n", len(result.Data.Assets)))
	for _, a := range result.Data.Assets {
		// Use name if available, else fallback to GUID
		name := a.Name
		if name == "" {
			name = a.GUID
		}
		sb.WriteString(fmt.Sprintf("  - %s (type: %s, ID: %s)\n", name, a.Type, a.ID))
	}
	return sb.String(), nil
}

func formatScenes(data []byte) (string, error) {
	var result struct {
		Data struct {
			Scenes []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
				GUID string `json:"guid"`
			} `json:"scenes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	if len(result.Data.Scenes) == 0 {
		return "No scenes found.", nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Scenes (%d):\n", len(result.Data.Scenes)))
	for _, s := range result.Data.Scenes {
		sb.WriteString(fmt.Sprintf("  - %s (ID: %s, GUID: %s)\n", s.Name, s.ID, s.GUID))
	}
	return sb.String(), nil
}

func formatGameObjects(data []byte) (string, error) {
	var result struct {
		Data struct {
			GameObjects []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"game_objects"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	if len(result.Data.GameObjects) == 0 {
		return "No GameObjects found.", nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("GameObjects (%d):\n", len(result.Data.GameObjects)))
	for _, gobj := range result.Data.GameObjects {
		sb.WriteString(fmt.Sprintf("  - %s (ID: %s)\n", gobj.Name, gobj.ID))
	}
	return sb.String(), nil
}

func formatScripts(data []byte) (string, error) {
	var result struct {
		Data struct {
			Scripts []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
				GUID string `json:"guid"`
			} `json:"scripts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	if len(result.Data.Scripts) == 0 {
		return "No scripts found.", nil
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Scripts (%d):\n", len(result.Data.Scripts)))
	for _, scr := range result.Data.Scripts {
		sb.WriteString(fmt.Sprintf("  - %s (ID: %s, GUID: %s)\n", scr.Name, scr.ID, scr.GUID))
	}
	return sb.String(), nil
}

func formatProjectStats(data []byte) (string, error) {
	var result struct {
		Data struct {
			Summary struct {
				TotalScenes           int `json:"total_scenes"`
				TotalGameObjects      int `json:"total_game_objects"`
				TotalComponents       int `json:"total_components"`
				TotalScripts          int `json:"total_scripts"`
				UniqueScripts         int `json:"unique_scripts"`
				TotalAssets           int `json:"total_assets"`
				TotalSerializedFields int `json:"total_serialized_fields"`
			} `json:"summary"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	s := result.Data.Summary
	return fmt.Sprintf(`Project Statistics:
- Scenes:              %d
- GameObjects:         %d
- Components:          %d
- Scripts:             %d
- Unique Scripts:      %d
- Assets:              %d
- Serialized Fields:   %d`,
		s.TotalScenes, s.TotalGameObjects, s.TotalComponents,
		s.TotalScripts, s.UniqueScripts, s.TotalAssets,
		s.TotalSerializedFields), nil
}

func formatSceneStats(data []byte) (string, error) {
	var result struct {
		Data struct {
			Scene struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"scene"`
			Statistics struct {
				GameObjectCount int `json:"game_object_count"`
				ComponentCount  int `json:"component_count"`
				ScriptCount     int `json:"script_count"`
				PrefabCount     int `json:"prefab_count"`
				AssetCount      int `json:"asset_count"`
				FieldCount      int `json:"field_count"`
			} `json:"statistics"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return "", err
	}
	s := result.Data.Statistics
	return fmt.Sprintf(`Scene: %s (ID: %s)
Statistics:
- GameObjects:  %d
- Components:   %d
- Scripts:      %d
- Prefabs:      %d
- Assets:       %d
- Fields:       %d`,
		result.Data.Scene.Name, result.Data.Scene.ID,
		s.GameObjectCount, s.ComponentCount, s.ScriptCount,
		s.PrefabCount, s.AssetCount, s.FieldCount), nil
}
