package ai

import (
	"encoding/json"
	"strings"

	"github.com/shahabsy/ProjectAtlasOS/internal/tools"
)

type RouteDecision struct {
	Tool       string          `json:"tool"`
	Arguments  json.RawMessage `json:"arguments"`
	Confidence float64         `json:"confidence"`
}

// Expanded action words to catch a wide range of query forms.
var actionWords = []string{
	"list", "show", "display", "all", "get", "give me",
	"what", "which", "find", "count", "how many", "number of",
	"in", "for",
}

func RouteQuery(query string, registry *tools.ToolRegistry) *RouteDecision {
	lower := strings.ToLower(query)

	makeDecision := func(tool string, args map[string]interface{}) *RouteDecision {
		data, _ := json.Marshal(args)
		return &RouteDecision{
			Tool:       tool,
			Arguments:  data,
			Confidence: 1.0,
		}
	}

	// Block describe/explain queries (they require reasoning)
	if matchesAny(lower, []string{"describe", "explain", "tell me about", "what is"}) {
		return nil
	}

	// Stats/count intents (highest priority – catch all count and stat queries)
	if matchesAny(lower, []string{"stat", "statistics", "stats", "project summary", "overview", "how many", "count", "number of"}) {
		return makeDecision("project_stats", nil)
	}

	// Asset intents
	if matchesAny(lower, []string{"asset", "assets"}) {
		if matchesAny(lower, actionWords) {
			return makeDecision("list_assets", nil)
		}
	}

	// Scene intents
	if matchesAny(lower, []string{"scene", "scenes"}) {
		// Avoid matching if it's a description (already blocked, but keep safety)
		if !matchesAny(lower, []string{"describe", "explain"}) {
			if matchesAny(lower, actionWords) {
				return makeDecision("list_scenes", nil)
			}
		}
	}

	// Script intents
	if matchesAny(lower, []string{"script", "scripts"}) {
		if matchesAny(lower, actionWords) {
			return makeDecision("list_all_scripts", nil)
		}
	}

	// GameObject intents
	if matchesAny(lower, []string{"gameobject", "game objects", "object", "objects"}) {
		if matchesAny(lower, actionWords) {
			return makeDecision("list_all_gameobjects", nil)
		}
	}

	return nil
}

func matchesAny(input string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(input, kw) {
			return true
		}
	}
	return false
}
