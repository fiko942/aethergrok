package grokrunner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// StreamCallbacks handles events parsed from the grok NDJSON stream
type StreamCallbacks struct {
	OnDelta             func(event DeltaEvent)
	OnToolCall          func(event ToolCallEvent)
	OnPermissionRequest func(event PermissionRequestEvent)
	OnComplete          func(event TurnCompleteEvent)
	OnError             func(err error)
}

// StreamParser reads NDJSON lines from grok stdout and batches delta tokens with a 16ms ticker
type StreamParser struct {
	sessionID string
	callbacks StreamCallbacks
	flushChan chan struct{}
}

// NewStreamParser creates a new stream parser for a session
func NewStreamParser(sessionID string, callbacks StreamCallbacks) *StreamParser {
	return &StreamParser{
		sessionID: sessionID,
		callbacks: callbacks,
		flushChan: make(chan struct{}, 1),
	}
}

// Parse reads from the given reader until EOF or context cancellation
func (p *StreamParser) Parse(ctx context.Context, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	// Allocate larger buffer for long lines/diffs
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var mu sync.Mutex
	var deltaBuffer strings.Builder
	var lastRole string

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	flushDelta := func() {
		mu.Lock()
		defer mu.Unlock()
		if deltaBuffer.Len() > 0 {
			text := deltaBuffer.String()
			deltaBuffer.Reset()
			if p.callbacks.OnDelta != nil {
				p.callbacks.OnDelta(DeltaEvent{
					SessionID: p.sessionID,
					Delta:     text,
					Role:      lastRole,
				})
			}
		}
	}

	flushLoopDone := make(chan struct{})
	go func() {
		defer close(flushLoopDone)
		for {
			select {
			case <-ctx.Done():
				flushDelta()
				return
			case <-p.flushChan:
				flushDelta()
				return
			case <-ticker.C:
				flushDelta()
			}
		}
	}()

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			p.triggerFlush()
			<-flushLoopDone
			return ctx.Err()
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var raw RawNDJSONEvent
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			// Fallback: treat unrecognized raw line as plain delta text
			mu.Lock()
			if lastRole == "" {
				lastRole = "assistant"
			}
			deltaBuffer.WriteString(line)
			deltaBuffer.WriteString("\n")
			mu.Unlock()
			continue
		}

		switch raw.Type {
		case "delta", "content", "token", "message", "text":
			content := raw.Delta
			if content == "" {
				content = raw.Content
			}
			if content == "" {
				content = raw.Data
			}
			if content != "" {
				mu.Lock()
				if raw.Role != "" {
					lastRole = raw.Role
				} else if lastRole == "" {
					lastRole = "assistant"
				}
				deltaBuffer.WriteString(content)
				mu.Unlock()
			}

		case "tool_call", "tool_use", "tool":
			// Tool calls should flush any buffered text first
			flushDelta()
			if p.callbacks.OnToolCall != nil {
				status := raw.ToolStatus
				if status == "" {
					status = "running"
				}
				p.callbacks.OnToolCall(ToolCallEvent{
					SessionID: p.sessionID,
					ToolID:    raw.ToolID,
					ToolName:  raw.ToolName,
					Input:     raw.ToolInput,
					Output:    raw.ToolOutput,
					Status:    status,
					Error:     raw.Error,
				})
			}

		case "permission_request", "permission", "ask":
			flushDelta()
			if p.callbacks.OnPermissionRequest != nil {
				p.callbacks.OnPermissionRequest(PermissionRequestEvent{
					SessionID:   p.sessionID,
					RequestID:   raw.RequestID,
					ToolName:    raw.ToolName,
					Description: raw.Description,
					Details:     raw.Details,
					Options: []PermissionOption{
						{ID: "allow_once", Label: "Allow Once"},
						{ID: "allow_always", Label: "Always Allow"},
						{ID: "reject", Label: "Reject"},
					},
				})
			}

		case "complete", "turn_complete", "done", "end":
			flushDelta()
			if p.callbacks.OnComplete != nil {
				status := raw.Status
				if status == "" {
					status = "success"
				}
				p.callbacks.OnComplete(TurnCompleteEvent{
					SessionID:    p.sessionID,
					Status:       status,
					Error:        raw.Error,
					TotalTokens:  raw.Tokens,
					FinishReason: raw.Status,
				})
			}

		case "error":
			flushDelta()
			if p.callbacks.OnError != nil {
				errMsg := raw.Message
				if errMsg == "" {
					errMsg = raw.Error
				}
				if errMsg == "" {
					errMsg = "Grok execution error"
				}
				p.callbacks.OnError(fmt.Errorf("%s", errMsg))
			}
		}
	}

	p.triggerFlush()
	<-flushLoopDone

	return scanner.Err()
}

func (p *StreamParser) triggerFlush() {
	select {
	case p.flushChan <- struct{}{}:
	default:
	}
}
