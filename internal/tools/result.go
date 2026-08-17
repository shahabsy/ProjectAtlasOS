package tools

// Result represents a structured response from a tool.
type Result struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *Error      `json:"error,omitempty"`
}

// Error provides structured error information.
type Error struct {
	Code    string `json:"code"`    // e.g., "SCENE_NOT_FOUND"
	Message string `json:"message"` // Human‑readable description
}

// NewSuccessResult creates a successful result.
func NewSuccessResult(data interface{}) Result {
	return Result{Success: true, Data: data}
}

// NewErrorResult creates an error result.
func NewErrorResult(code, message string) Result {
	return Result{
		Success: false,
		Error: &Error{
			Code:    code,
			Message: message,
		},
	}
}
