package ai

// Session maintains conversation state for interactive chat.
type Session struct {
	History []Message
	Agent   *Agent
}

// NewSession creates a new session with the given agent.
func NewSession(agent *Agent) *Session {
	return &Session{
		History: []Message{},
		Agent:   agent,
	}
}

// Ask processes a user question within the session context.
func (s *Session) Ask(question string) (AgentResult, error) {
	// Add user message to history.
	s.History = append(s.History, Message{Role: "user", Content: question})
	// Run agent with full history.
	return s.Agent.RunWithHistory(question, s.History)
}
