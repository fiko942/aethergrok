package grokrunner

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestACPStreamParser_TurnCompleted(t *testing.T) {
	input := `{"timestamp":1790697528,"method":"session/update","params":{"sessionId":"test-session-123","update":{"sessionUpdate":"agent_message_chunk","content":{"type":"text","text":"Task complete!"}}}}
{"timestamp":1790697539,"method":"_x.ai/session/update","params":{"sessionId":"test-session-123","update":{"sessionUpdate":"turn_completed","prompt_id":"fa1b5d12-67cc-47c7-b942-93a0304502e0","stop_reason":"end_turn","usage":{"totalTokens":45000},"elapsed_ms":1234}}}`

	var mu sync.Mutex
	var deltas []string
	var completeEvent *TurnCompleteEvent

	callbacks := StreamCallbacks{
		OnDelta: func(e DeltaEvent) {
			mu.Lock()
			defer mu.Unlock()
			deltas = append(deltas, e.Delta)
		},
		OnComplete: func(e TurnCompleteEvent) {
			mu.Lock()
			defer mu.Unlock()
			completeEvent = &e
		},
	}

	parser := NewStreamParser("test-session-123", callbacks)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := parser.Parse(ctx, strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if !parser.HasCompleted() {
		t.Errorf("expected parser.HasCompleted() to be true")
	}

	if completeEvent == nil {
		t.Fatalf("expected OnComplete to be called, got nil")
	}

	if completeEvent.Status != "success" {
		t.Errorf("expected complete status 'success', got '%s'", completeEvent.Status)
	}

	if completeEvent.TotalTokens != 45000 {
		t.Errorf("expected total tokens 45000, got %d", completeEvent.TotalTokens)
	}

	if completeEvent.FinishReason != "end_turn" {
		t.Errorf("expected finishReason 'end_turn', got '%s'", completeEvent.FinishReason)
	}

	if len(deltas) == 0 || !strings.Contains(strings.Join(deltas, ""), "Task complete!") {
		t.Errorf("expected delta containing 'Task complete!', got %v", deltas)
	}
}

func TestACPStreamParser_ToolCallAndUpdate(t *testing.T) {
	input := `{"timestamp":1790697207,"method":"session/update","params":{"sessionId":"test-sess-2","update":{"sessionUpdate":"tool_call","toolCallId":"call_123","title":"Executing shell command","rawInput":{"command":"ls -la"},"_meta":{"x.ai/tool":{"name":"run_terminal_command"}}}}}
{"timestamp":1790697210,"method":"session/update","params":{"sessionId":"test-sess-2","update":{"sessionUpdate":"tool_call_update","toolCallId":"call_123","status":"Completed","rawOutput":{"type":"Bash","output_for_prompt":"file1.txt\nfile2.txt"}}}}
{"timestamp":1790697215,"method":"session/update","params":{"sessionId":"test-sess-2","update":{"sessionUpdate":"turn_completed"}}}`

	var mu sync.Mutex
	var toolCalls []ToolCallEvent
	var completed bool

	callbacks := StreamCallbacks{
		OnToolCall: func(e ToolCallEvent) {
			mu.Lock()
			defer mu.Unlock()
			toolCalls = append(toolCalls, e)
		},
		OnComplete: func(e TurnCompleteEvent) {
			mu.Lock()
			defer mu.Unlock()
			completed = true
		},
	}

	parser := NewStreamParser("test-sess-2", callbacks)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := parser.Parse(ctx, strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(toolCalls) < 2 {
		t.Fatalf("expected at least 2 tool call events, got %d", len(toolCalls))
	}

	firstCall := toolCalls[0]
	if firstCall.ToolID != "call_123" {
		t.Errorf("expected tool ID 'call_123', got '%s'", firstCall.ToolID)
	}
	if firstCall.ToolName != "run_terminal_command" {
		t.Errorf("expected tool name 'run_terminal_command', got '%s'", firstCall.ToolName)
	}
	if firstCall.Status != "running" {
		t.Errorf("expected first tool status 'running', got '%s'", firstCall.Status)
	}

	secondCall := toolCalls[1]
	if secondCall.Status != "completed" {
		t.Errorf("expected second tool status 'completed', got '%s'", secondCall.Status)
	}
	if !strings.Contains(secondCall.Output, "file1.txt") {
		t.Errorf("expected output to contain 'file1.txt', got '%s'", secondCall.Output)
	}

	if !completed {
		t.Errorf("expected turn to complete")
	}
}

func TestStreamParser_FallbackEmitComplete(t *testing.T) {
	input := `{"type":"delta","delta":"Hello world"}`

	var completeEvent *TurnCompleteEvent
	callbacks := StreamCallbacks{
		OnComplete: func(e TurnCompleteEvent) {
			completeEvent = &e
		},
	}

	parser := NewStreamParser("test-sess-fallback", callbacks)
	ctx := context.Background()

	_ = parser.Parse(ctx, strings.NewReader(input))

	if parser.HasCompleted() {
		t.Errorf("expected HasCompleted to be false before EmitComplete")
	}

	// Simulate runner fallback when process terminates
	parser.EmitComplete("success", "")

	if !parser.HasCompleted() {
		t.Errorf("expected HasCompleted to be true after EmitComplete")
	}

	if completeEvent == nil {
		t.Fatalf("expected fallback OnComplete to be called")
	}

	if completeEvent.Status != "success" {
		t.Errorf("expected status 'success', got '%s'", completeEvent.Status)
	}
}

func TestACPStreamParser_PermissionRequest(t *testing.T) {
	input := `{"timestamp":1790697207,"method":"session/update","params":{"sessionId":"test-sess-perm","update":{"sessionUpdate":"permission_request","requestId":"req_999","title":"write_file","description":"Write to app.go","details":{"path":"app.go"}}}}`

	var permReq *PermissionRequestEvent
	callbacks := StreamCallbacks{
		OnPermissionRequest: func(e PermissionRequestEvent) {
			permReq = &e
		},
	}

	parser := NewStreamParser("test-sess-perm", callbacks)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_ = parser.Parse(ctx, strings.NewReader(input))

	if permReq == nil {
		t.Fatalf("expected OnPermissionRequest to be called")
	}

	if permReq.RequestID != "req_999" {
		t.Errorf("expected request ID 'req_999', got '%s'", permReq.RequestID)
	}

	if permReq.Description != "Write to app.go" {
		t.Errorf("expected description 'Write to app.go', got '%s'", permReq.Description)
	}
}

func TestStreamParser_FirstLineChan(t *testing.T) {
	parser := NewStreamParser("test-sess-firstline", StreamCallbacks{})
	select {
	case <-parser.FirstLineChan():
		t.Fatalf("firstLineChan should not be closed initially")
	default:
	}

	input := "{\"type\":\"available_commands\"}\n{\"type\":\"delta\",\"delta\":\"hi\"}\n"
	ctx := context.Background()
	_ = parser.Parse(ctx, strings.NewReader(input))

	select {
	case <-parser.FirstLineChan():
		// Success!
	default:
		t.Fatalf("firstLineChan should be closed after parsing lines")
	}
}

func TestStreamParser_MaxTokensTruncationError(t *testing.T) {
	rawJSON := `{"type":"error","error":"Internal error: {\n  \"message\": \"response truncated by max_tokens\",\n  \"error_kind\": \"max_tokens_truncation\",\n  \"promptUsage\": {\n    \"inputTokens\": 5777245,\n    \"outputTokens\": 23083,\n    \"totalTokens\": 5800328,\n    \"reasoningTokens\": 18377,\n    \"modelCalls\": 54,\n    \"numTurns\": 54\n  }\n}"}`

	var completedEvent *TurnCompleteEvent
	var errReceived error

	callbacks := StreamCallbacks{
		OnError: func(err error) {
			errReceived = err
		},
		OnComplete: func(evt TurnCompleteEvent) {
			completedEvent = &evt
		},
	}

	parser := NewStreamParser("sess-1", callbacks)
	err := parser.Parse(context.Background(), strings.NewReader(rawJSON+"\n"))
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if completedEvent == nil {
		t.Fatal("expected TurnCompleteEvent to be emitted")
	}
	if completedEvent.ErrorKind != "max_tokens_truncation" {
		t.Fatalf("expected ErrorKind 'max_tokens_truncation', got '%s'", completedEvent.ErrorKind)
	}
	if completedEvent.TotalTokens != 5800328 {
		t.Fatalf("expected TotalTokens 5800328, got %d", completedEvent.TotalTokens)
	}
	if errReceived == nil {
		t.Fatal("expected OnError callback to be invoked")
	}
	if !strings.Contains(errReceived.Error(), "Context limit reached: Response truncated by max_tokens") {
		t.Fatalf("expected normalized error message, got: %v", errReceived)
	}
}
