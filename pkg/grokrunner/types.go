package grokrunner

// EventType represents the type of stream event
type EventType string

const (
	EventDelta             EventType = "grok:delta_batch"
	EventToolCall          EventType = "grok:tool_call"
	EventPermissionRequest EventType = "grok:permission_request"
	EventTurnComplete      EventType = "grok:complete"
	EventError             EventType = "grok:error"
)

// SessionOptions configures a Grok session run
type SessionOptions struct {
	Model            string   `json:"model,omitempty"`
	ReasoningEffort  string   `json:"reasoningEffort,omitempty"` // low, medium, high
	WorkingDir       string   `json:"workingDir,omitempty"`
	SkillDirs        []string `json:"skillDirs,omitempty"`
	Temperature      *float64 `json:"temperature,omitempty"`
	DisableTools     bool     `json:"disableTools,omitempty"`
	SystemPrompt     string   `json:"systemPrompt,omitempty"`
	CustomFlags      []string `json:"customFlags,omitempty"`
}

// PromptRequest is the request payload to start or continue a session
type PromptRequest struct {
	SessionID string         `json:"sessionId"`
	Prompt    string         `json:"prompt"`
	Images    []string       `json:"images,omitempty"`
	Options   SessionOptions `json:"options,omitempty"`
}

// DeltaEvent represents batched text tokens emitted to frontend
type DeltaEvent struct {
	SessionID string `json:"sessionId"`
	Delta     string `json:"delta"`
	Role      string `json:"role,omitempty"`
}

// ToolCallEvent represents a tool invocation or completion
type ToolCallEvent struct {
	SessionID string                 `json:"sessionId"`
	ToolID    string                 `json:"toolId"`
	ToolName  string                 `json:"toolName"`
	Input     map[string]interface{} `json:"input,omitempty"`
	Output    string                 `json:"output,omitempty"`
	Status    string                 `json:"status"` // "running", "completed", "failed"
	Error     string                 `json:"error,omitempty"`
}

// PermissionOption represents an approval option
type PermissionOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// PermissionRequestEvent represents a required user confirmation
type PermissionRequestEvent struct {
	SessionID   string             `json:"sessionId"`
	RequestID   string             `json:"requestId"`
	ToolName    string             `json:"toolName"`
	Description string             `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Options     []PermissionOption `json:"options,omitempty"`
}

// PermissionResponse is the user's decision on a permission request
type PermissionResponse struct {
	SessionID string `json:"sessionId"`
	RequestID string `json:"requestId"`
	Decision  string `json:"decision"` // "allow_once", "allow_always", "reject"
}

// TurnCompleteEvent signals that the turn has completed
type TurnCompleteEvent struct {
	SessionID    string `json:"sessionId"`
	Status       string `json:"status"` // "success", "interrupted", "error"
	Error        string `json:"error,omitempty"`
	TotalTokens  int    `json:"totalTokens,omitempty"`
	FinishReason string `json:"finishReason,omitempty"`
}

// RawNDJSONEvent models incoming JSON messages from `grok --output-format streaming-json`
type RawNDJSONEvent struct {
	Type        string                 `json:"type"`
	Role        string                 `json:"role,omitempty"`
	Content     string                 `json:"content,omitempty"`
	Delta       string                 `json:"delta,omitempty"`
	Data        string                 `json:"data,omitempty"`
	Message     string                 `json:"message,omitempty"`
	ToolID      string                 `json:"tool_id,omitempty"`
	ToolName    string                 `json:"tool_name,omitempty"`
	ToolInput   map[string]interface{} `json:"tool_input,omitempty"`
	ToolOutput  string                 `json:"tool_output,omitempty"`
	ToolStatus  string                 `json:"tool_status,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Tokens      int                    `json:"tokens,omitempty"`
}
