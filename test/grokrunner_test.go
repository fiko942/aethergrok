package test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"aethergrok/pkg/grokrunner"
)

func TestStreamParser_TokenBatching16ms(t *testing.T) {
	sessionID := "test-session-1"

	var mu sync.Mutex
	var deltas []grokrunner.DeltaEvent

	callbacks := grokrunner.StreamCallbacks{
		OnDelta: func(event grokrunner.DeltaEvent) {
			mu.Lock()
			deltas = append(deltas, event)
			mu.Unlock()
		},
	}

	parser := grokrunner.NewStreamParser(sessionID, callbacks)

	// Simulate rapid token delivery
	inputData := `{"type":"delta","delta":"Hello "}
{"type":"delta","delta":"world, "}
{"type":"delta","delta":"this "}
{"type":"delta","delta":"is "}
{"type":"delta","delta":"batched!"}
{"type":"complete","status":"success"}
`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := parser.Parse(ctx, strings.NewReader(inputData))
	if err != nil {
		t.Fatalf("unexpected error parsing stream: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(deltas) == 0 {
		t.Fatalf("expected delta events, got 0")
	}

	var combined strings.Builder
	for _, d := range deltas {
		combined.WriteString(d.Delta)
	}

	expected := "Hello world, this is batched!"
	if combined.String() != expected {
		t.Errorf("expected combined text %q, got %q", expected, combined.String())
	}
}

func TestStreamParser_ToolCallAndCompleteEvents(t *testing.T) {
	sessionID := "test-session-2"

	var toolCalls []grokrunner.ToolCallEvent
	var turnComplete *grokrunner.TurnCompleteEvent
	var permissionReq *grokrunner.PermissionRequestEvent
	var mu sync.Mutex

	callbacks := grokrunner.StreamCallbacks{
		OnToolCall: func(event grokrunner.ToolCallEvent) {
			mu.Lock()
			toolCalls = append(toolCalls, event)
			mu.Unlock()
		},
		OnPermissionRequest: func(event grokrunner.PermissionRequestEvent) {
			mu.Lock()
			permissionReq = &event
			mu.Unlock()
		},
		OnComplete: func(event grokrunner.TurnCompleteEvent) {
			mu.Lock()
			turnComplete = &event
			mu.Unlock()
		},
	}

	parser := grokrunner.NewStreamParser(sessionID, callbacks)

	inputData := `{"type":"tool_call","tool_id":"call_1","tool_name":"run_terminal_command","tool_input":{"command":"ls -la"}}
{"type":"permission_request","request_id":"req_1","tool_name":"run_terminal_command","description":"Run shell command"}
{"type":"complete","status":"success","tokens":42}
`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := parser.Parse(ctx, strings.NewReader(inputData))
	if err != nil {
		t.Fatalf("unexpected error parsing stream: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(toolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(toolCalls))
	}
	if toolCalls[0].ToolName != "run_terminal_command" {
		t.Errorf("expected tool name run_terminal_command, got %s", toolCalls[0].ToolName)
	}
	if permissionReq == nil {
		t.Fatalf("expected permission request event, got nil")
	}
	if permissionReq.RequestID != "req_1" {
		t.Errorf("expected requestId req_1, got %s", permissionReq.RequestID)
	}
	if turnComplete == nil {
		t.Fatalf("expected turn complete event, got nil")
	}
	if turnComplete.TotalTokens != 42 {
		t.Errorf("expected 42 tokens, got %d", turnComplete.TotalTokens)
	}
}

func TestRunner_LifecycleAndCancel(t *testing.T) {
	runner := grokrunner.NewRunner()
	// Point to bash or sh to test execution & cancellation
	runner.SetBinaryPath("sleep")

	sessionID := "test-session-cancel"
	req := grokrunner.PromptRequest{
		SessionID: sessionID,
		Prompt:    "10", // sleep 10
	}

	ctx := context.Background()
	callbacks := grokrunner.StreamCallbacks{}

	err := runner.StartSession(ctx, req, callbacks)
	if err != nil {
		t.Fatalf("failed to start session: %v", err)
	}

	// Cancel the session
	err = runner.Cancel(sessionID)
	if err != nil {
		t.Fatalf("failed to cancel session: %v", err)
	}
}
