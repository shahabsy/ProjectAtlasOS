package ai

// AgentConfig holds configuration for the AI agent.
type AgentConfig struct {
	// MaxTurns is the maximum number of conversation turns (rounds) before giving up.
	MaxTurns int `json:"max_turns"`
	// MaxToolCallsPerTurn is the maximum number of tool calls allowed in a single turn.
	MaxToolCallsPerTurn int `json:"max_tool_calls_per_turn"`
	// Verbose enables detailed execution logging.
	Verbose bool `json:"verbose"`
	// SafetyCeiling is the absolute maximum tool calls allowed (hard limit).
	SafetyCeiling int `json:"safety_ceiling"`
}

// DefaultConfig returns a safe default configuration.
func DefaultConfig() AgentConfig {
	return AgentConfig{
		MaxTurns:            8,
		MaxToolCallsPerTurn: 5,
		Verbose:             false,
		SafetyCeiling:       32,
	}
}
