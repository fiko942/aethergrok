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

type TruncationPayload struct {
	Message     string `json:"message"`
	ErrorKind   string `json:"error_kind"`
	PromptUsage struct {
		InputTokens     int64 `json:"inputTokens"`
		OutputTokens    int64 `json:"outputTokens"`
		TotalTokens     int64 `json:"totalTokens"`
		ReasoningTokens int64 `json:"reasoningTokens"`
		ModelCalls      int64 `json:"modelCalls"`
		NumTurns        int   `json:"numTurns"`
	} `json:"promptUsage"`
}

func parseTruncationError(errMsg string) (isTruncation bool, normalizedMsg string, totalTokens int, usageMap map[string]interface{}) {
	if !strings.Contains(errMsg, "max_tokens") && !strings.Contains(errMsg, "response truncated") {
		return false, errMsg, 0, nil
	}

	idx := strings.Index(errMsg, "{")
	if idx != -1 {
		jsonPart := errMsg[idx:]
		var tp TruncationPayload
		if err := json.Unmarshal([]byte(jsonPart), &tp); err == nil && (tp.ErrorKind == "max_tokens_truncation" || strings.Contains(tp.Message, "max_tokens")) {
			totalTokens = int(tp.PromptUsage.TotalTokens)
			if totalTokens == 0 {
				totalTokens = int(tp.PromptUsage.InputTokens + tp.PromptUsage.OutputTokens)
			}
			usageMap = map[string]interface{}{
				"inputTokens":     tp.PromptUsage.InputTokens,
				"outputTokens":    tp.PromptUsage.OutputTokens,
				"totalTokens":     tp.PromptUsage.TotalTokens,
				"reasoningTokens": tp.PromptUsage.ReasoningTokens,
				"modelCalls":      tp.PromptUsage.ModelCalls,
				"numTurns":        tp.PromptUsage.NumTurns,
			}
			turnsStr := ""
			if tp.PromptUsage.NumTurns > 0 {
				turnsStr = fmt.Sprintf(" after %d tool iterations", tp.PromptUsage.NumTurns)
			}
			normalizedMsg = fmt.Sprintf("Context limit reached: Response truncated by max_tokens%s (%d accumulated tokens)", turnsStr, totalTokens)
			return true, normalizedMsg, totalTokens, usageMap
		}
	}

	return true, "Context limit reached: Response truncated by max_tokens", 0, nil
}

// extractTextContent recursively extracts text strings from polymorphic content blocks
func extractTextContent(content interface{}) string {
	if content == nil {
		return ""
	}
	if str, ok := content.(string); ok {
		return str
	}
	if m, ok := content.(map[string]interface{}); ok {
		if txt, ok := m["text"].(string); ok {
			return txt
		}
		if sub, ok := m["content"].(map[string]interface{}); ok {
			if txt, ok := sub["text"].(string); ok {
				return txt
			}
		}
		if sub, ok := m["content"].(string); ok {
			return sub
		}
	}
	if arr, ok := content.([]interface{}); ok {
		var sb strings.Builder
		for _, item := range arr {
			txt := extractTextContent(item)
			if txt != "" {
				sb.WriteString(txt)
			}
		}
		return sb.String()
	}
	return ""
}

// extractToolOutput parses output from polymorphic rawOutput or content blocks
func extractToolOutput(rawOutput interface{}, contentBlocks interface{}) string {
	// 1. Check content blocks first (rendered text format)
	if contentBlocks != nil {
		if txt := extractTextContent(contentBlocks); strings.TrimSpace(txt) != "" {
			return cleanANSI(txt)
		}
	}

	if rawOutput == nil {
		return ""
	}

	if str, ok := rawOutput.(string); ok {
		return cleanANSI(str)
	}

	if rawMap, ok := rawOutput.(map[string]interface{}); ok {
		// Output for prompt if available (clean human readable text)
		if outputForPrompt, ok := rawMap["output_for_prompt"].(string); ok && strings.TrimSpace(outputForPrompt) != "" {
			return cleanANSI(outputForPrompt)
		}

		// Result structure (e.g. from get_command_or_subagent_output or terminal runners)
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

		if out, ok := rawMap["output"].(string); ok && strings.TrimSpace(out) != "" {
			return cleanANSI(out)
		}
		if res, ok := rawMap["result"].(string); ok && strings.TrimSpace(res) != "" {
			return cleanANSI(res)
		}
		if std, ok := rawMap["stdout"].(string); ok && strings.TrimSpace(std) != "" {
			return cleanANSI(std)
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
	firstLineChan chan struct{}
	firstLineOnce sync.Once
	mu            sync.Mutex
	hasCompleted  bool
}

// NewStreamParser creates a new stream parser for a session
func NewStreamParser(sessionID string, callbacks StreamCallbacks) *StreamParser {
	return &StreamParser{
		sessionID:     sessionID,
		callbacks:     callbacks,
		flushChan:     make(chan struct{}, 1),
		firstLineChan: make(chan struct{}),
	}
}

// SetInitialGrokSessionID sets the underlying Grok session UUID if not already set
func (p *StreamParser) SetInitialGrokSessionID(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.grokSessionID == "" && id != "" {
		p.grokSessionID = id
	}
}

// FirstLineChan returns a channel that is closed upon receiving the first non-empty stdout line
func (p *StreamParser) FirstLineChan() <-chan struct{} {
	return p.firstLineChan
}

// HasCompleted returns true if a turn completion or fatal error was emitted
func (p *StreamParser) HasCompleted() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.hasCompleted
}

// EmitComplete delivers a guaranteed turn completion event
func (p *StreamParser) EmitComplete(status string, errMsg string) {
	p.mu.Lock()
	if p.hasCompleted {
		p.mu.Unlock()
		return
	}
	p.hasCompleted = true
	p.mu.Unlock()

	p.triggerFlush()
	time.Sleep(20 * time.Millisecond)

	if p.callbacks.OnComplete != nil {
		p.callbacks.OnComplete(TurnCompleteEvent{
			SessionID:     p.sessionID,
			GrokSessionID: p.grokSessionID,
			Status:        status,
			Error:         errMsg,
			FinishReason:  status,
		})
	}
}

// Parse reads from the given reader until EOF or context cancellation
func (p *StreamParser) Parse(ctx context.Context, r io.Reader) error {
	defer p.firstLineOnce.Do(func() {
		close(p.firstLineChan)
	})

	scanner := bufio.NewScanner(r)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	var bufMu sync.Mutex
	var deltaBuffer strings.Builder
	var lastRole string
	var lastLiveTokens int

	ticker := time.NewTicker(16 * time.Millisecond)
	defer ticker.Stop()

	flushDelta := func() {
		bufMu.Lock()
		defer bufMu.Unlock()
		if deltaBuffer.Len() > 0 {
			text := deltaBuffer.String()
			deltaBuffer.Reset()
			if p.callbacks.OnDelta != nil {
				p.callbacks.OnDelta(DeltaEvent{
					SessionID:     p.sessionID,
					GrokSessionID: p.grokSessionID,
					Delta:         text,
					Role:          lastRole,
					Tokens:        lastLiveTokens,
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

		p.firstLineOnce.Do(func() {
			close(p.firstLineChan)
		})

		var raw RawNDJSONEvent
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		// Extract ACP update payload if present
		var payload ACPUpdatePayload
		if raw.Params != nil {
			if updateRaw, exists := raw.Params["update"]; exists {
				if updateBytes, err := json.Marshal(updateRaw); err == nil {
					_ = json.Unmarshal(updateBytes, &payload)
				}
			}
		}

		// Capture underlying real Grok Session ID if present
		if raw.SessionID != "" {
			p.grokSessionID = raw.SessionID
		} else if raw.SessionIDAlt != "" {
			p.grokSessionID = raw.SessionIDAlt
		} else if raw.Params != nil {
			if sID, ok := raw.Params["sessionId"].(string); ok && sID != "" {
				p.grokSessionID = sID
			}
		}

		// Determine normalized action type across ACP and flat NDJSON
		actionType := payload.SessionUpdate
		if actionType == "" {
			actionType = raw.Type
		}
		if actionType == "" && raw.Method != "" {
			methodLower := strings.ToLower(raw.Method)
			if strings.Contains(methodLower, "turn_completed") || strings.Contains(methodLower, "turn_complete") {
				actionType = "turn_completed"
			}
		}
		actionType = strings.ToLower(strings.TrimSpace(actionType))

		// Filter out internal metadata/thought events
		if actionType == "available_commands" || actionType == "thought" || actionType == "agent_thought_chunk" || actionType == "thought_chunk" {
			continue
		}

		// Capture live token metrics if present in envelope or metadata
		if raw.Params != nil {
			if meta, ok := raw.Params["_meta"].(map[string]interface{}); ok {
				if tok, ok := meta["totalTokens"].(float64); ok && tok > 0 {
					lastLiveTokens = int(tok)
				} else if tok, ok := meta["contextTokensUsed"].(float64); ok && tok > 0 {
					lastLiveTokens = int(tok)
				}
			}
		}
		if raw.Tokens > 0 {
			lastLiveTokens = raw.Tokens
		}
		if payload.TokensAfter > 0 {
			lastLiveTokens = payload.TokensAfter
		}

		switch actionType {
		case "agent_message_chunk", "user_message_chunk", "message_chunk", "text_chunk", "delta", "content", "token", "message", "text":
			content := payload.Delta
			if content == "" {
				content = payload.Text
			}
			if content == "" {
				content = payload.Message
			}
			if content == "" && payload.Content != nil {
				content = extractTextContent(payload.Content)
			}
			if content == "" {
				content = raw.Delta
			}
			if content == "" {
				content = raw.Data
			}
			if content == "" {
				content = raw.Text
			}
			if content == "" {
				content = raw.Message
			}
			if content == "" && raw.Content != nil {
				content = extractTextContent(raw.Content)
			}

			if content != "" {
				bufMu.Lock()
				if actionType == "user_message_chunk" {
					lastRole = "user"
				} else if raw.Role != "" {
					lastRole = raw.Role
				} else if lastRole == "" {
					lastRole = "assistant"
				}
				deltaBuffer.WriteString(content)
				bufMu.Unlock()
			}

		case "tool_call", "tool_call_update", "tool_use", "tool":
			flushDelta()
			if p.callbacks.OnToolCall != nil {
				toolID := payload.ToolCallID
				if toolID == "" {
					toolID = payload.ToolID
				}
				if toolID == "" {
					toolID = raw.ToolCallID
				}
				if toolID == "" {
					toolID = raw.ToolID
				}

				// Extract tool name from ACP metadata, title, or flat name
				var toolName string
				if payload.Meta != nil {
					if toolMeta, ok := payload.Meta["x.ai/tool"].(map[string]interface{}); ok {
						if name, ok := toolMeta["name"].(string); ok && name != "" {
							toolName = name
						}
					}
				}
				if toolName == "" && raw.Meta != nil {
					if toolMeta, ok := raw.Meta["x.ai/tool"].(map[string]interface{}); ok {
						if name, ok := toolMeta["name"].(string); ok && name != "" {
							toolName = name
						}
					}
				}
				if toolName == "" {
					toolName = raw.ToolName
				}
				if toolName == "" {
					toolName = raw.ToolNameAlt
				}
				if toolName == "" {
					toolName = payload.Title
				}
				if toolName == "" {
					toolName = raw.Title
				}

				input := payload.RawInput
				if input == nil {
					input = payload.ToolInput
				}
				if input == nil {
					input = raw.RawInput
				}
				if input == nil {
					input = raw.ToolInput
				}

				output := payload.ToolOutput
				if output == "" {
					output = extractToolOutput(payload.RawOutput, payload.Content)
				}
				if output == "" {
					output = raw.ToolOutput
				}
				if output == "" {
					output = extractToolOutput(raw.RawOutput, raw.Content)
				}

				rawStatus := payload.Status
				if rawStatus == "" {
					rawStatus = raw.Status
				}
				if rawStatus == "" {
					rawStatus = raw.ToolStatus
				}
				if rawStatus == "" && raw.Params != nil {
					if m, ok := raw.Params["_meta"].(map[string]interface{}); ok {
						if up, ok := m["updateParams"].(map[string]interface{}); ok {
							if st, ok := up["status"].(string); ok {
								rawStatus = st
							}
						}
					}
				}

				stLower := strings.ToLower(strings.TrimSpace(rawStatus))
				var status string
				if stLower == "completed" || stLower == "success" || stLower == "done" {
					status = "completed"
				} else if stLower == "failed" || stLower == "error" || stLower == "rejected" {
					status = "failed"
				} else if actionType == "tool_call_update" && output != "" && rawStatus == "" {
					status = "completed"
				} else {
					status = "running"
				}

				p.callbacks.OnToolCall(ToolCallEvent{
					SessionID:     p.sessionID,
					GrokSessionID: p.grokSessionID,
					ToolID:        toolID,
					ToolName:      toolName,
					Input:         input,
					Output:        output,
					Status:        status,
					Error:         payload.Error,
				})
			}

		case "permission_request", "permission", "ask", "request_permission":
			flushDelta()
			if p.callbacks.OnPermissionRequest != nil {
				reqID := payload.RequestID
				if reqID == "" {
					reqID = payload.RequestIDAlt
				}
				if reqID == "" {
					reqID = raw.RequestID
				}
				if reqID == "" {
					reqID = payload.ToolCallID
				}

				toolName := raw.ToolName
				if toolName == "" {
					toolName = raw.ToolNameAlt
				}
				if toolName == "" {
					toolName = payload.Title
				}

				desc := payload.Description
				if desc == "" {
					desc = raw.Description
				}

				details := payload.Details
				if details == nil {
					details = raw.Details
				}

				p.callbacks.OnPermissionRequest(PermissionRequestEvent{
					SessionID:   p.sessionID,
					RequestID:   reqID,
					ToolName:    toolName,
					Description: desc,
					Details:     details,
					Options: []PermissionOption{
						{ID: "allow_once", Label: "Allow Once"},
						{ID: "allow_always", Label: "Always Allow"},
						{ID: "reject", Label: "Reject"},
					},
				})
			}

		case "turn_completed", "turn_complete", "task_completed", "complete", "completed", "done", "end", "finish", "finished", "message_stop", "turn_finish":
			flushDelta()
			p.mu.Lock()
			p.hasCompleted = true
			p.mu.Unlock()

			if p.callbacks.OnComplete != nil {
				rawStatus := payload.Status
				if rawStatus == "" {
					rawStatus = raw.Status
				}
				stLower := strings.ToLower(strings.TrimSpace(rawStatus))
				status := "success"
				if stLower == "interrupted" || stLower == "cancelled" {
					status = "interrupted"
				} else if stLower == "error" || stLower == "failed" {
					status = "error"
				}

				title := payload.Title
				if title == "" {
					title = raw.Title
				}

				totalTokens := raw.Tokens
				if totalTokens == 0 && payload.Usage != nil {
					if tok, ok := payload.Usage["totalTokens"].(float64); ok {
						totalTokens = int(tok)
					} else if tok, ok := payload.Usage["total_tokens"].(float64); ok {
						totalTokens = int(tok)
					}
				}
				if totalTokens == 0 && raw.Usage != nil {
					if tok, ok := raw.Usage["totalTokens"].(float64); ok {
						totalTokens = int(tok)
					}
				}
				if totalTokens == 0 && raw.Params != nil {
					if m, ok := raw.Params["_meta"].(map[string]interface{}); ok {
						if tok, ok := m["totalTokens"].(float64); ok {
							totalTokens = int(tok)
						}
					}
				}

				finishReason := payload.StopReason
				if finishReason == "" {
					finishReason = raw.Status
				}

				errStr := payload.Error
				if errStr == "" {
					errStr = raw.Error
				}

				errorKind := ""
				var usageMap map[string]interface{}
				if isTrunc, normMsg, tokens, uMap := parseTruncationError(errStr); isTrunc {
					errorKind = "max_tokens_truncation"
					errStr = normMsg
					if totalTokens == 0 {
						totalTokens = tokens
					}
					usageMap = uMap
				}

				p.callbacks.OnComplete(TurnCompleteEvent{
					SessionID:     p.sessionID,
					GrokSessionID: p.grokSessionID,
					Title:         title,
					Status:        status,
					Error:         errStr,
					ErrorKind:     errorKind,
					TotalTokens:   totalTokens,
					Usage:         usageMap,
					FinishReason:  finishReason,
				})
			}

		case "error", "task_failed", "turn_failed":
			flushDelta()
			p.mu.Lock()
			p.hasCompleted = true
			p.mu.Unlock()

			errMsg := payload.Message
			if errMsg == "" {
				errMsg = payload.Error
			}
			if errMsg == "" {
				errMsg = raw.Message
			}
			if errMsg == "" {
				errMsg = raw.Error
			}
			if errMsg == "" {
				errMsg = "Grok execution error"
			}

			errorKind := ""
			totalTokens := 0
			var usageMap map[string]interface{}
			if isTrunc, normMsg, tokens, uMap := parseTruncationError(errMsg); isTrunc {
				errorKind = "max_tokens_truncation"
				errMsg = normMsg
				totalTokens = tokens
				usageMap = uMap
			}

			if p.callbacks.OnError != nil {
				p.callbacks.OnError(fmt.Errorf("%s", errMsg))
			}
			if p.callbacks.OnComplete != nil {
				p.callbacks.OnComplete(TurnCompleteEvent{
					SessionID:     p.sessionID,
					GrokSessionID: p.grokSessionID,
					Status:        "error",
					Error:         errMsg,
					ErrorKind:     errorKind,
					TotalTokens:   totalTokens,
					Usage:         usageMap,
					FinishReason:  "error",
				})
			}

		case "auto_compact_started", "compaction_checkpoint", "auto_compact_completed", "task_backgrounded", "background_tasks", "usage":
			// Informational ACP notifications - keep stream running smoothly
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
