package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Message represents a single message in the Ollama chat history.
// This matches Ollama's native format exactly.
type Message struct {
	Role      string     `json:"role"`                 // "system", "user", "assistant", "tool"
	Content   string     `json:"content,omitempty"`    // text content
	ToolCalls []ToolCall `json:"tool_calls,omitempty"` // for assistant messages
	ToolName  string     `json:"tool_name,omitempty"`  // for tool messages
}

// ToolCall represents a function call requested by the model.
type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction contains the details of the function call.
type ToolCallFunction struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// Tool defines a tool that the model can call.
type Tool struct {
	Type     string      `json:"type"` // "function"
	Function FunctionDef `json:"function"`
}

// FunctionDef describes a callable function.
type FunctionDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

// Parameters describes the expected arguments of a function.
type Parameters struct {
	Type       string              `json:"type"` // "object"
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

// Property defines a single parameter.
type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ChatRequest is the payload for Ollama's /api/chat endpoint.
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
	Stream   bool      `json:"stream"`
}

// ChatResponse is the response from Ollama's /api/chat endpoint.
type ChatResponse struct {
	Message Message `json:"message"`
}

// OllamaClient handles communication with Ollama.
type OllamaClient struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

// NewOllamaClient creates a new client.
func NewOllamaClient(baseURL, model string) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3.2"
	}
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		Client:  &http.Client{},
	}
}

// Chat sends a chat request and returns the assistant's message.
func (c *OllamaClient) Chat(messages []Message, tools []Tool) (*Message, error) {
	req := ChatRequest{
		Model:    c.Model,
		Messages: messages,
		Tools:    tools,
		Stream:   false,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.Client.Post(c.BaseURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned status %d", resp.StatusCode)
	}
	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}
	return &chatResp.Message, nil
}
