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
	Model           string   `json:"model,omitempty"`
	ReasoningEffort string   `json:"reasoningEffort,omitempty"` // low, medium, high
	WorkingDir      string   `json:"workingDir,omitempty"`
	SkillDirs       []string `json:"skillDirs,omitempty"`
	Temperature     *float64 `json:"temperature,omitempty"`
	DisableTools    bool     `json:"disableTools,omitempty"`
	SystemPrompt    string   `json:"systemPrompt,omitempty"`
	CustomFlags     []string `json:"customFlags,omitempty"`
	GrokSessionID   string   `json:"grokSessionId,omitempty"` // Underlying Grok UUID on disk to resume
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
	SessionID     string `json:"sessionId"`
	GrokSessionID string `json:"grokSessionId,omitempty"`
	Delta         string `json:"delta"`
	Role          string `json:"role,omitempty"`
	Tokens        int    `json:"tokens,omitempty"`
}

// ToolCallEvent represents a tool invocation or completion
type ToolCallEvent struct {
	SessionID     string                 `json:"sessionId"`
	GrokSessionID string                 `json:"grokSessionId,omitempty"`
	ToolID        string                 `json:"toolId"`
	ToolName      string                 `json:"toolName"`
	Input         map[string]interface{} `json:"input,omitempty"`
	Output        string                 `json:"output,omitempty"`
	Status        string                 `json:"status"` // "running", "completed", "failed"
	Error         string                 `json:"error,omitempty"`
}

// PermissionOption represents an approval option
type PermissionOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// PermissionRequestEvent represents a required user confirmation
type PermissionRequestEvent struct {
	SessionID   string                 `json:"sessionId"`
	RequestID   string                 `json:"requestId"`
	ToolName    string                 `json:"toolName"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Options     []PermissionOption     `json:"options,omitempty"`
}

// PermissionResponse is the user's decision on a permission request
type PermissionResponse struct {
	SessionID string `json:"sessionId"`
	RequestID string `json:"requestId"`
	Decision  string `json:"decision"` // "allow_once", "allow_always", "reject"
}

// DiscoveredChatMessage represents a reconstructed message for the frontend
type DiscoveredChatMessage struct {
	ID               string                 `json:"id"`
	Role             string                 `json:"role"` // "user", "assistant", "system"
	Content          string                 `json:"content"`
	Timestamp        int64                  `json:"timestamp"`
	ReasoningContent string                 `json:"reasoningContent,omitempty"`
	ToolCalls        []DiscoveredToolCall   `json:"toolCalls,omitempty"`
	Tokens           map[string]interface{} `json:"tokens,omitempty"`
	Status           string                 `json:"status,omitempty"`
}

// DiscoveredToolCall represents an executed tool reconstructed from history
type DiscoveredToolCall struct {
	ID        string                 `json:"id"`
	Tool      string                 `json:"tool"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Result    string                 `json:"result,omitempty"`
	Status    string                 `json:"status"` // "completed", "failed", "running"
	StartTime int64                  `json:"startTime,omitempty"`
	EndTime   int64                  `json:"endTime,omitempty"`
}

// SessionUsageStats holds comprehensive token breakdown for a session
type SessionUsageStats struct {
	SessionID          string `json:"sessionId"`
	UsedTokens         int64  `json:"usedTokens"`
	MaxTokens          int64  `json:"maxTokens"`
	LastTurnInput      int64  `json:"lastTurnInput"`
	LastTurnOutput     int64  `json:"lastTurnOutput"`
	LastTurnCacheRead  int64  `json:"lastTurnCacheRead"`
	LastTurnReasoning  int64  `json:"lastTurnReasoning"`
	LastTurnModelCalls int64  `json:"lastTurnModelCalls"`
	TotalInput         int64  `json:"totalInput"`
	TotalOutput        int64  `json:"totalOutput"`
	TotalCacheRead     int64  `json:"totalCacheRead"`
	TurnCount          int    `json:"turnCount"`
	PrimaryModelID     string `json:"primaryModelId"`
}

// TurnCompleteEvent signals that the turn has completed
type TurnCompleteEvent struct {
	SessionID     string                 `json:"sessionId"`
	GrokSessionID string                 `json:"grokSessionId,omitempty"`
	Title         string                 `json:"title,omitempty"`
	Status        string                 `json:"status"` // "success", "interrupted", "error"
	Error         string                 `json:"error,omitempty"`
	ErrorKind     string                 `json:"errorKind,omitempty"`
	TotalTokens   int                    `json:"totalTokens,omitempty"`
	FinishReason  string                 `json:"finishReason,omitempty"`
	Usage         map[string]interface{} `json:"usage,omitempty"`
}

// RawNDJSONEvent models incoming JSON messages from `grok --output-format streaming-json`
type RawNDJSONEvent struct {
	Method       string                 `json:"method,omitempty"`
	Type         string                 `json:"type,omitempty"`
	Role         string                 `json:"role,omitempty"`
	Delta        string                 `json:"delta,omitempty"`
	Data         string                 `json:"data,omitempty"`
	Message      string                 `json:"message,omitempty"`
	Text         string                 `json:"text,omitempty"`
	Title        string                 `json:"title,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	SessionIDAlt string                 `json:"sessionId,omitempty"`
	Params       map[string]interface{} `json:"params,omitempty"`
	Kind         string                 `json:"kind,omitempty"`
	ToolID       string                 `json:"tool_id,omitempty"`
	ToolCallID   string                 `json:"toolCallId,omitempty"`
	ToolName     string                 `json:"toolName,omitempty"`
	ToolNameAlt  string                 `json:"tool_name,omitempty"`
	RawInput     map[string]interface{} `json:"rawInput,omitempty"`
	ToolInput    map[string]interface{} `json:"tool_input,omitempty"`
	RawOutput    interface{}            `json:"rawOutput,omitempty"`
	Content      interface{}            `json:"content,omitempty"`
	ToolOutput   string                 `json:"tool_output,omitempty"`
	ToolStatus   string                 `json:"tool_status,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	Description  string                 `json:"description,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Status       string                 `json:"status,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Tokens       int                    `json:"tokens,omitempty"`
	Usage        map[string]interface{} `json:"usage,omitempty"`
	Meta         map[string]interface{} `json:"_meta,omitempty"`
}

// ACPUpdatePayload models update payloads nested within ACP JSON-RPC params.update
type ACPUpdatePayload struct {
	SessionUpdate string                 `json:"sessionUpdate,omitempty"`
	ToolCallID    string                 `json:"toolCallId,omitempty"`
	ToolID        string                 `json:"toolId,omitempty"`
	Kind          string                 `json:"kind,omitempty"`
	Title         string                 `json:"title,omitempty"`
	Status        string                 `json:"status,omitempty"`
	Content       interface{}            `json:"content,omitempty"`
	RawInput      map[string]interface{} `json:"rawInput,omitempty"`
	ToolInput     map[string]interface{} `json:"tool_input,omitempty"`
	RawOutput     interface{}            `json:"rawOutput,omitempty"`
	ToolOutput    string                 `json:"tool_output,omitempty"`
	Delta         string                 `json:"delta,omitempty"`
	Text          string                 `json:"text,omitempty"`
	Message       string                 `json:"message,omitempty"`
	PromptID      string                 `json:"prompt_id,omitempty"`
	StopReason    string                 `json:"stop_reason,omitempty"`
	Usage         map[string]interface{} `json:"usage,omitempty"`
	Tokens        int                    `json:"tokens,omitempty"`
	TokensAfter   int                    `json:"tokens_after,omitempty"`
	ElapsedMs     int64                  `json:"elapsed_ms,omitempty"`
	RequestID     string                 `json:"requestId,omitempty"`
	RequestIDAlt  string                 `json:"request_id,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Details       map[string]interface{} `json:"details,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Meta          map[string]interface{} `json:"_meta,omitempty"`
}
