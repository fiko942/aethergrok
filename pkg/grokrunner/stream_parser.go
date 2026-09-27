package grokrunner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"
)

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func cleanANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// extractToolOutput parses output from polymorphic rawOutput or content blocks
func extractToolOutput(rawOutput interface{}, contentBlocks []interface{}) string {
	// 1. Check content blocks first (rendered text format)
	for _, block := range contentBlocks {
		if bMap, ok := block.(map[string]interface{}); ok {
			if cnt, ok := bMap["content"].(map[string]interface{}); ok {
				if txt, ok := cnt["text"].(string); ok && strings.TrimSpace(txt) != "" {
					return cleanANSI(txt)
				}
			} else if txt, ok := bMap["text"].(string); ok && strings.TrimSpace(txt) != "" {
				return cleanANSI(txt)
			}
		}
	}

	if rawOutput == nil {
		return ""
	}

	if str, ok := rawOutput.(string); ok {
		return cleanANSI(str)
	}

	if rawMap, ok := rawOutput.(map[string]interface{}); ok {
		// Unwrap TaskOutput / Result structure (e.g. from get_command_or_subagent_output or terminal runners)
		if resMap, ok := rawMap["Result"].(map[string]interface{}); ok {
			if out, ok := resMap["output"].(string); ok && strings.TrimSpace(out) != "" {
				return cleanANSI(out)
			}
		}
		if resMap, ok := rawMap["result"].(map[string]interface{}); ok {
			if out, ok := resMap["output"].(string); ok && strings.TrimSpace(out) != "" {
				return cleanANSI(out)
			}
		}

		// Output for prompt if available
		if outputForPrompt, ok := rawMap["output_for_prompt"].(string); ok && strings.TrimSpace(outputForPrompt) != "" {
			return cleanANSI(outputForPrompt)
		}

		// Byte array output in Bash { "type": "Bash", "output": [116, 111, 116, ...] }
		if arr, ok := rawMap["output"].([]interface{}); ok && len(arr) > 0 {
			bytes := make([]byte, 0, len(arr))
			allBytes := true
			for _, v := range arr {
				if num, ok := v.(float64); ok {
					bytes = append(bytes, byte(num))
				} else {
					allBytes = false
					break
				}
			}
			if allBytes && len(bytes) > 0 {
				return cleanANSI(string(bytes))
			}
		}

		// ReadFile: { "FileContent": { "content": "..." } }
		if fc, ok := rawMap["FileContent"].(map[string]interface{}); ok {
			if c, ok := fc["content"].(string); ok {
				return cleanANSI(c)
			}
		}

		// ListDir: { "Content": { "content": "..." } }
		if cMap, ok := rawMap["Content"].(map[string]interface{}); ok {
			if c, ok := cMap["content"].(string); ok {
				return cleanANSI(c)
			}
		}

		if out, ok := rawMap["output"].(string); ok {
			return cleanANSI(out)
		}
		if res, ok := rawMap["result"].(string); ok {
			return cleanANSI(res)
		}
	}

	if b, err := json.MarshalIndent(rawOutput, "", "  "); err == nil {
		return cleanANSI(string(b))
	}

	return fmt.Sprintf("%v", rawOutput)
}

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
	sessionID     string
	grokSessionID string
	callbacks     StreamCallbacks
	flushChan     chan struct{}
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
			continue
		}

		// Capture underlying real Grok Session ID if present in stream
		if raw.SessionID != "" {
			p.grokSessionID = raw.SessionID
		} else if raw.Params != nil {
			if sID, ok := raw.Params["sessionId"].(string); ok && sID != "" {
				p.grokSessionID = sID
			}
		}

		// Filter out internal metadata events like available_commands, thought
		if raw.Type == "available_commands" || raw.Type == "thought" {
			continue
		}

		switch raw.Type {
		case "delta", "content", "token", "message", "text":
			content := raw.Delta
			if content == "" {
				content = raw.Data
			}
			if content == "" {
				content = raw.Text
			}
			if content == "" {
				content = raw.Message
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

		case "tool_call", "tool_call_update", "tool_use", "tool":
			flushDelta()
			if p.callbacks.OnToolCall != nil {
				toolID := raw.ToolCallID
				if toolID == "" {
					toolID = raw.ToolID
				}
				toolName := raw.ToolName
				if toolName == "" {
					toolName = raw.ToolNameAlt
				}
				if toolName == "" {
					toolName = raw.Title
				}

				input := raw.RawInput
				if input == nil {
					input = raw.ToolInput
				}

				output := raw.ToolOutput
				if output == "" {
					output = extractToolOutput(raw.RawOutput, raw.Content)
				}

				status := raw.Status
				if status == "" {
					status = raw.ToolStatus
				}
				if status == "" || status == "pending" {
					status = "running"
				}

				p.callbacks.OnToolCall(ToolCallEvent{
					SessionID: p.sessionID,
					ToolID:    toolID,
					ToolName:  toolName,
					Input:     input,
					Output:    output,
					Status:    status,
					Error:     raw.Error,
				})
			}

		case "permission_request", "permission", "ask":
			flushDelta()
			if p.callbacks.OnPermissionRequest != nil {
				toolName := raw.ToolName
				if toolName == "" {
					toolName = raw.ToolNameAlt
				}
				p.callbacks.OnPermissionRequest(PermissionRequestEvent{
					SessionID:   p.sessionID,
					RequestID:   raw.RequestID,
					ToolName:    toolName,
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
				title := raw.Title
				p.callbacks.OnComplete(TurnCompleteEvent{
					SessionID:     p.sessionID,
					GrokSessionID: p.grokSessionID,
					Title:         title,
					Status:        status,
					Error:         raw.Error,
					TotalTokens:   raw.Tokens,
					FinishReason:  raw.Status,
				})
			}

		case "usage":
			// Usage metadata
			if raw.Usage != nil {
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
