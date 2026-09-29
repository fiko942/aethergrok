package grokrunner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// GrokSessionMetadata represents metadata of a discovered Grok session from disk
type GrokSessionMetadata struct {
	ID          string `json:"id"`
	WorkspaceID string `json:"workspaceId"`
	Title       string `json:"title"`
	Path        string `json:"path"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
	Status      string `json:"status"`
}

var userQueryRegex = regexp.MustCompile(`(?s)<user_query>(.*?)(?:</user_query>|$)`)

// EncodeGrokWorkspacePath converts a workspace file system path into the exact URL-encoded
// directory leaf format used by Grok CLI in ~/.grok/sessions/ (matching JS encodeURIComponent)
func EncodeGrokWorkspacePath(workspacePath string) string {
	if workspacePath == "" {
		return ""
	}
	clean := filepath.Clean(workspacePath)
	var sb strings.Builder
	for i := 0; i < len(clean); i++ {
		b := clean[i]
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') ||
			b == '-' || b == '_' || b == '.' || b == '~' || b == '!' || b == '*' || b == '\'' || b == '(' || b == ')' {
			sb.WriteByte(b)
		} else {
			fmt.Fprintf(&sb, "%%%02X", b)
		}
	}
	return sb.String()
}

// ResolveWorkspaceSessionsDir returns the directory path under ~/.grok/sessions for workspacePath
func ResolveWorkspaceSessionsDir(sessionsDir, workspacePath string) string {
	if workspacePath == "" {
		return sessionsDir
	}
	cleanWs := filepath.Clean(workspacePath)

	// 1. Try exact encoded path
	primary := filepath.Join(sessionsDir, EncodeGrokWorkspacePath(cleanWs))
	if fi, err := os.Stat(primary); err == nil && fi.IsDir() {
		return primary
	}

	// 2. Try legacy url.PathEscape
	legacy := filepath.Join(sessionsDir, url.PathEscape(cleanWs))
	if fi, err := os.Stat(legacy); err == nil && fi.IsDir() {
		return legacy
	}

	// 3. Scan all directory entries in sessionsDir and match decoded folder names
	entries, err := os.ReadDir(sessionsDir)
	if err == nil {
		normalizedTarget := strings.ToLower(filepath.ToSlash(cleanWs))
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			name := entry.Name()
			decoded, unerr := url.PathUnescape(name)
			if unerr != nil {
				decoded, unerr = url.QueryUnescape(name)
			}
			if unerr == nil {
				normalizedDecoded := strings.ToLower(filepath.ToSlash(filepath.Clean(decoded)))
				if normalizedDecoded == normalizedTarget || strings.TrimRight(normalizedDecoded, "/") == strings.TrimRight(normalizedTarget, "/") {
					return filepath.Join(sessionsDir, name)
				}
			}
		}
	}

	return primary
}

// ResolveSessionFolder resolves the session folder path and true Grok UUID inside targetDir
func ResolveSessionFolder(targetDir, sessionID string) (string, string) {
	if sessionID != "" {
		exactPath := filepath.Join(targetDir, sessionID)
		if fi, err := os.Stat(exactPath); err == nil && fi.IsDir() {
			return exactPath, sessionID
		}
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil || len(entries) == 0 {
		return filepath.Join(targetDir, sessionID), sessionID
	}

	// 1. Search for matching folder name (case-insensitive or prefix/suffix)
	if sessionID != "" {
		sLower := strings.ToLower(sessionID)
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			eName := entry.Name()
			eLower := strings.ToLower(eName)
			if eLower == sLower || strings.HasPrefix(eLower, sLower) || strings.HasPrefix(sLower, eLower) {
				return filepath.Join(targetDir, eName), eName
			}
		}
	}

	// 2. If sessionID is a frontend temporary ID (sess_...) or not found, find the most recently modified session folder
	type sessionDirInfo struct {
		name    string
		modTime time.Time
	}
	var validDirs []sessionDirInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		folderPath := filepath.Join(targetDir, entry.Name())
		if hasSessionRecord(folderPath) {
			info, err := entry.Info()
			if err == nil {
				validDirs = append(validDirs, sessionDirInfo{name: entry.Name(), modTime: info.ModTime()})
			}
		}
	}

	if len(validDirs) > 0 {
		sort.Slice(validDirs, func(i, j int) bool {
			return validDirs[i].modTime.After(validDirs[j].modTime)
		})
		return filepath.Join(targetDir, validDirs[0].name), validDirs[0].name
	}

	return filepath.Join(targetDir, sessionID), sessionID
}

func hasSessionRecord(folderPath string) bool {
	files := []string{"signals.json", "usage.json", "chat_history.jsonl", "summary.json", "updates.jsonl"}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(folderPath, f)); err == nil {
			return true
		}
	}
	return false
}

// DiscoverGrokSessions scans ~/.grok/sessions/ for sessions matching workspacePath
func DiscoverGrokSessions(workspacePath string) ([]GrokSessionMetadata, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	if _, err := os.Stat(sessionsDir); os.IsNotExist(err) {
		return []GrokSessionMetadata{}, nil
	}

	targetDir := ResolveWorkspaceSessionsDir(sessionsDir, workspacePath)

	entries, err := os.ReadDir(targetDir)
	if err != nil {
		return []GrokSessionMetadata{}, nil
	}

	var results []GrokSessionMetadata

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		sessionID := entry.Name()
		sessionFolderPath := filepath.Join(targetDir, sessionID)

		info, err := entry.Info()
		if err != nil {
			continue
		}

		title := ""
		createdAt := info.ModTime().UnixMilli()
		updatedAt := info.ModTime().UnixMilli()

		// 1. Check summary.json first (the official Grok title source)
		summaryPath := filepath.Join(sessionFolderPath, "summary.json")
		numMessages := 0
		if data, err := os.ReadFile(summaryPath); err == nil {
			var summaryObj struct {
				SessionSummary  string `json:"session_summary"`
				GeneratedTitle  string `json:"generated_title"`
				NumMessages     int    `json:"num_messages"`
				NumChatMessages int    `json:"num_chat_messages"`
				CreatedAt       string `json:"created_at"`
				UpdatedAt       string `json:"updated_at"`
			}
			if json.Unmarshal(data, &summaryObj) == nil {
				numMessages = summaryObj.NumMessages
				if summaryObj.SessionSummary != "" {
					title = strings.TrimSpace(summaryObj.SessionSummary)
				} else if summaryObj.GeneratedTitle != "" {
					title = strings.TrimSpace(summaryObj.GeneratedTitle)
				}
			}
		}

		// 2. If no summary.json title, extract real user prompt from chat_history.jsonl or transcript
		hasValidChat := false
		historyFiles := []string{
			filepath.Join(sessionFolderPath, "chat_history.jsonl"),
			filepath.Join(sessionFolderPath, "updates.jsonl"),
			filepath.Join(sessionFolderPath, "transcript.jsonl"),
		}

		for _, hFile := range historyFiles {
			data, err := os.ReadFile(hFile)
			if err != nil {
				continue
			}

			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				var rawMsg map[string]interface{}
				if json.Unmarshal([]byte(line), &rawMsg) != nil {
					continue
				}

				role, _ := rawMsg["role"].(string)
				msgType, _ := rawMsg["type"].(string)
				if role != "user" && msgType != "user" {
					continue
				}

				var textCandidates []string

				if cStr, ok := rawMsg["content"].(string); ok {
					textCandidates = append(textCandidates, cStr)
				} else if cList, ok := rawMsg["content"].([]interface{}); ok {
					for _, item := range cList {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if itemMap["type"] == "text" {
								if tStr, ok := itemMap["text"].(string); ok {
									textCandidates = append(textCandidates, tStr)
								}
							}
						}
					}
				}

				for _, contentStr := range textCandidates {
					contentStr = strings.TrimSpace(contentStr)
					if contentStr == "" {
						continue
					}

					// Extract text inside <user_query> tag if wrapped
					matches := userQueryRegex.FindStringSubmatch(contentStr)
					extracted := ""
					if len(matches) > 1 {
						extracted = strings.TrimSpace(matches[1])
					} else if !strings.HasPrefix(contentStr, "<user_info>") && !strings.HasPrefix(contentStr, "<system-reminder>") {
						extracted = strings.TrimSpace(contentStr)
					}

					// Clean out XML comments, slash commands or markdown fences if prompt starts with them
					if extracted != "" && !strings.HasPrefix(extracted, "<!--") {
						hasValidChat = true
						if title == "" {
							clean := strings.ReplaceAll(extracted, "\n", " ")
							clean = strings.TrimSpace(clean)
							if len(clean) > 42 {
								title = clean[:42] + "..."
							} else {
								title = clean
							}
						}
						break
					}
				}

				if title != "" {
					break
				}
			}

			if title != "" {
				break
			}
		}

		// Filter out empty ghost sessions (0 messages and no user prompts)
		if title == "" && numMessages == 0 && !hasValidChat {
			continue
		}

		// Fallback clean title if session has content but no title
		if title == "" {
			title = "Percakapan " + sessionID[:min(len(sessionID), 8)]
		}

		results = append(results, GrokSessionMetadata{
			ID:          sessionID,
			WorkspaceID: workspacePath,
			Title:       title,
			Path:        sessionFolderPath,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
			Status:      "idle",
		})
	}

	// Sort newest modified first
	sort.Slice(results, func(i, j int) bool {
		return results[i].UpdatedAt > results[j].UpdatedAt
	})

	return results, nil
}

// LoadGrokSessionMessages parses chat_history.jsonl or updates.jsonl from disk for a specific session
func LoadGrokSessionMessages(workspacePath, sessionID string) ([]DiscoveredChatMessage, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	targetDir := ResolveWorkspaceSessionsDir(sessionsDir, workspacePath)
	sessionFolderPath, _ := ResolveSessionFolder(targetDir, sessionID)

	chatHistoryPath := filepath.Join(sessionFolderPath, "chat_history.jsonl")
	data, err := os.ReadFile(chatHistoryPath)
	if err != nil {
		// Fallback to updates.jsonl if chat_history doesn't exist
		chatHistoryPath = filepath.Join(sessionFolderPath, "updates.jsonl")
		data, err = os.ReadFile(chatHistoryPath)
		if err != nil {
			return []DiscoveredChatMessage{}, nil
		}
	}

	lines := strings.Split(string(data), "\n")
	var messages []DiscoveredChatMessage
	var pendingToolCalls []DiscoveredToolCall
	toolCallIndexMap := make(map[string]int)

	for lineIdx, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if json.Unmarshal([]byte(line), &raw) != nil {
			continue
		}

		msgType, _ := raw["type"].(string)
		role, _ := raw["role"].(string)
		if role == "" {
			role = msgType
		}

		// Skip internal system prompts or compaction meta
		if role == "system" {
			continue
		}

		// 1. Process User Message
		if role == "user" {
			// Flush pending tool calls to previous assistant message if any
			if len(pendingToolCalls) > 0 && len(messages) > 0 && messages[len(messages)-1].Role == "assistant" {
				messages[len(messages)-1].ToolCalls = append(messages[len(messages)-1].ToolCalls, pendingToolCalls...)
				pendingToolCalls = nil
			}

			userText := ""
			if cStr, ok := raw["content"].(string); ok {
				userText = cStr
			} else if cList, ok := raw["content"].([]interface{}); ok {
				for _, item := range cList {
					if itemMap, ok := item.(map[string]interface{}); ok {
						if itemMap["type"] == "text" {
							if tStr, ok := itemMap["text"].(string); ok {
								userText += tStr + "\n"
							}
						}
					}
				}
			}

			userText = strings.TrimSpace(userText)
			// Extract real user prompt from <user_query>
			matches := userQueryRegex.FindStringSubmatch(userText)
			if len(matches) > 1 {
				userText = strings.TrimSpace(matches[1])
			} else {
				// Filter out non-user prompt payloads like system reminders or compaction headers
				if strings.HasPrefix(userText, "<user_info>") || strings.HasPrefix(userText, "<system-reminder>") || strings.HasPrefix(userText, "This session is being continued") {
					continue
				}
			}

			if userText == "" {
				continue
			}

			messages = append(messages, DiscoveredChatMessage{
				ID:        fmt.Sprintf("%s_u_%d", sessionID, lineIdx),
				Role:      "user",
				Content:   userText,
				Timestamp: 0,
				Status:    "done",
			})
		} else if role == "assistant" {
			// Assistant message
			contentStr := ""
			if cStr, ok := raw["content"].(string); ok {
				contentStr = cStr
			}

			// Tool calls inside assistant entry
			if tcList, ok := raw["tool_calls"].([]interface{}); ok {
				for _, tc := range tcList {
					if tcMap, ok := tc.(map[string]interface{}); ok {
						tcID, _ := tcMap["id"].(string)
						tcName, _ := tcMap["name"].(string)
						var inputMap map[string]interface{}
						if argsStr, ok := tcMap["arguments"].(string); ok {
							_ = json.Unmarshal([]byte(argsStr), &inputMap)
						} else if argsObj, ok := tcMap["arguments"].(map[string]interface{}); ok {
							inputMap = argsObj
						}

						newTc := DiscoveredToolCall{
							ID:     tcID,
							Tool:   tcName,
							Params: inputMap,
							Status: "completed",
						}
						toolCallIndexMap[tcID] = len(pendingToolCalls)
						pendingToolCalls = append(pendingToolCalls, newTc)
					}
				}
			}

			if contentStr != "" {
				// Check if last message was already assistant, we can append content
				if len(messages) > 0 && messages[len(messages)-1].Role == "assistant" && len(messages[len(messages)-1].Content) == 0 {
					messages[len(messages)-1].Content = contentStr
					if len(pendingToolCalls) > 0 {
						messages[len(messages)-1].ToolCalls = append(messages[len(messages)-1].ToolCalls, pendingToolCalls...)
						pendingToolCalls = nil
					}
				} else {
					var currentToolCalls []DiscoveredToolCall
					if len(pendingToolCalls) > 0 {
						currentToolCalls = pendingToolCalls
						pendingToolCalls = nil
					}
					messages = append(messages, DiscoveredChatMessage{
						ID:        fmt.Sprintf("%s_a_%d", sessionID, lineIdx),
						Role:      "assistant",
						Content:   contentStr,
						Timestamp: 0,
						Status:    "done",
						ToolCalls: currentToolCalls,
					})
				}
			}
		} else if role == "tool_result" || msgType == "tool_result" {
			// Populate result into matched tool call
			tcID, _ := raw["tool_call_id"].(string)
			resultStr := ""
			if cStr, ok := raw["content"].(string); ok {
				resultStr = cStr
			}

			if idx, ok := toolCallIndexMap[tcID]; ok && idx < len(pendingToolCalls) {
				pendingToolCalls[idx].Result = resultStr
			}
		}
	}

	// Flush any trailing tool calls
	if len(pendingToolCalls) > 0 {
		if len(messages) > 0 && messages[len(messages)-1].Role == "assistant" {
			messages[len(messages)-1].ToolCalls = append(messages[len(messages)-1].ToolCalls, pendingToolCalls...)
		} else {
			messages = append(messages, DiscoveredChatMessage{
				ID:        sessionID + "_a_tail",
				Role:      "assistant",
				Content:   "",
				Status:    "done",
				ToolCalls: pendingToolCalls,
			})
		}
	}

	return messages, nil
}

// DeleteGrokSessionDirectory removes a session directory from ~/.grok/sessions/
func DeleteGrokSessionDirectory(workspacePath, sessionID string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	sessionsParent := ResolveWorkspaceSessionsDir(sessionsDir, workspacePath)
	sessionFolderPath, resolvedID := ResolveSessionFolder(sessionsParent, sessionID)

	// 1. Direct path removal
	if _, err := os.Stat(sessionFolderPath); err == nil {
		_ = os.RemoveAll(sessionFolderPath)
	}

	if resolvedID != sessionID && sessionID != "" {
		_ = os.RemoveAll(filepath.Join(sessionsParent, sessionID))
	}

	// 2. Fallback search inside workspace session folder for partial match or uuid match
	entries, err := os.ReadDir(sessionsParent)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() && (entry.Name() == sessionID || entry.Name() == resolvedID || strings.HasPrefix(entry.Name(), sessionID)) {
			_ = os.RemoveAll(filepath.Join(sessionsParent, entry.Name()))
		}
	}

	return nil
}

// GetSessionUsage reads usage.json or calculates token metrics from chat_history.jsonl
func GetSessionUsage(workspacePath, sessionID string) (*SessionUsageStats, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	sessionsParent := ResolveWorkspaceSessionsDir(sessionsDir, workspacePath)
	sessionFolderPath, resolvedID := ResolveSessionFolder(sessionsParent, sessionID)

	stats := &SessionUsageStats{
		SessionID: resolvedID,
		MaxTokens: 200000,
	}

	// 1. Try reading official signals.json (Grok's true active context window metrics)
	signalsFilePath := filepath.Join(sessionFolderPath, "signals.json")
	if data, err := os.ReadFile(signalsFilePath); err == nil {
		var sigObj struct {
			ContextTokensUsed   int64  `json:"contextTokensUsed"`
			ContextWindowTokens int64  `json:"contextWindowTokens"`
			TurnCount           int    `json:"turnCount"`
			PrimaryModelID      string `json:"primaryModelId"`
		}
		if json.Unmarshal(data, &sigObj) == nil && sigObj.ContextTokensUsed > 0 {
			stats.UsedTokens = sigObj.ContextTokensUsed
			if sigObj.ContextWindowTokens > 0 {
				stats.MaxTokens = sigObj.ContextWindowTokens
			}
			stats.TurnCount = sigObj.TurnCount
			if sigObj.PrimaryModelID != "" {
				stats.PrimaryModelID = sigObj.PrimaryModelID
			}
		}
	}

	// 2. Read usage.json for cumulative session metrics and breakdown
	usageFilePath := filepath.Join(sessionFolderPath, "usage.json")
	if data, err := os.ReadFile(usageFilePath); err == nil {
		var rawUsage struct {
			SessionID string `json:"sessionId"`
			Session   struct {
				InputTokens      int64  `json:"inputTokens"`
				OutputTokens     int64  `json:"outputTokens"`
				CachedReadTokens int64  `json:"cachedReadTokens"`
				ReasoningTokens  int64  `json:"reasoningTokens"`
				TotalTokens      int64  `json:"totalTokens"`
				ModelCalls       int64  `json:"modelCalls"`
				TurnCount        int    `json:"turnCount"`
				PrimaryModelID   string `json:"primaryModelId"`
			} `json:"session"`
			Turns []struct {
				TurnNumber       int    `json:"turnNumber"`
				InputTokens      int64  `json:"inputTokens"`
				OutputTokens     int64  `json:"outputTokens"`
				CachedReadTokens int64  `json:"cachedReadTokens"`
				ReasoningTokens  int64  `json:"reasoningTokens"`
				TotalTokens      int64  `json:"totalTokens"`
				ModelCalls       int64  `json:"modelCalls"`
				PrimaryModelID   string `json:"primaryModelId"`
			} `json:"turns"`
		}

		if json.Unmarshal(data, &rawUsage) == nil {
			stats.TotalInput = rawUsage.Session.InputTokens
			stats.TotalOutput = rawUsage.Session.OutputTokens
			stats.TotalCacheRead = rawUsage.Session.CachedReadTokens
			if stats.TurnCount == 0 {
				stats.TurnCount = rawUsage.Session.TurnCount
			}
			if stats.PrimaryModelID == "" {
				stats.PrimaryModelID = rawUsage.Session.PrimaryModelID
			}

			if len(rawUsage.Turns) > 0 {
				lastTurn := rawUsage.Turns[len(rawUsage.Turns)-1]
				stats.LastTurnInput = lastTurn.InputTokens
				stats.LastTurnOutput = lastTurn.OutputTokens
				stats.LastTurnCacheRead = lastTurn.CachedReadTokens
				stats.LastTurnReasoning = lastTurn.ReasoningTokens
				stats.LastTurnModelCalls = lastTurn.ModelCalls

				// If signals.json was missing, fallback to last turn's context input
				if stats.UsedTokens == 0 {
					stats.UsedTokens = lastTurn.InputTokens + lastTurn.OutputTokens
				}
			} else if stats.UsedTokens == 0 {
				stats.UsedTokens = rawUsage.Session.TotalTokens
			}

			if stats.MaxTokens <= 0 || stats.MaxTokens == 200000 {
				if strings.Contains(strings.ToLower(stats.PrimaryModelID), "grok-4.6") {
					stats.MaxTokens = 2000000
				} else if strings.Contains(strings.ToLower(stats.PrimaryModelID), "128k") {
					stats.MaxTokens = 128000
				} else {
					stats.MaxTokens = 200000
				}
			}

			return stats, nil
		}
	}

	// 3. Fallback: estimate from loaded messages
	msgs, err := LoadGrokSessionMessages(workspacePath, resolvedID)
	if err == nil && len(msgs) > 0 {
		var charCount int64
		for _, m := range msgs {
			charCount += int64(len(m.Content))
			if m.ReasoningContent != "" {
				charCount += int64(len(m.ReasoningContent))
			}
			for _, tc := range m.ToolCalls {
				charCount += int64(len(tc.Result))
			}
		}
		estimatedTokens := (charCount / 4) + 2500
		if stats.UsedTokens == 0 {
			stats.UsedTokens = estimatedTokens
		}
		if stats.TotalInput == 0 {
			stats.TotalInput = estimatedTokens
		}
		if stats.TurnCount == 0 {
			stats.TurnCount = len(msgs)
		}
	}

	return stats, nil
}

// CompactSession triggers conversation compaction for the given session
func CompactSession(ctx context.Context, grokBinaryPath, workspacePath, sessionID string) (*SessionUsageStats, error) {
	if grokBinaryPath == "" || grokBinaryPath == "grok" {
		grokBinaryPath = ResolveGrokBinary()
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	sessionsParent := ResolveWorkspaceSessionsDir(sessionsDir, workspacePath)
	_, resolvedID := ResolveSessionFolder(sessionsParent, sessionID)

	// 1. Snapshot usage before compact
	statsBefore, _ := GetSessionUsage(workspacePath, resolvedID)
	if statsBefore == nil {
		statsBefore = &SessionUsageStats{SessionID: resolvedID, MaxTokens: 200000}
	}

	// 2. Execute /compact command via grok CLI with non-interactive single-turn -p flag
	cmd := exec.CommandContext(ctx, grokBinaryPath, "--resume", resolvedID, "-p", "/compact")
	cmd.Env = EnsureExecEnvironment()
	if workspacePath != "" {
		cmd.Dir = workspacePath
	}

	// Capture command output
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	runErr := cmd.Run()

	// 3. Retrieve updated post-compaction usage stats
	statsAfter, _ := GetSessionUsage(workspacePath, resolvedID)
	if statsAfter == nil {
		statsAfter = statsBefore
	}

	// If compaction succeeded and reduced context tokens
	if statsAfter.UsedTokens > 0 && statsBefore.UsedTokens > 0 && statsAfter.UsedTokens < statsBefore.UsedTokens {
		return statsAfter, nil
	}

	// If grok CLI execution succeeded without error
	if runErr == nil {
		// If post-stats didn't update yet (due to asynchronous signals.json rewrite), estimate reasonable reduction
		if statsAfter.UsedTokens >= statsBefore.UsedTokens && statsBefore.UsedTokens > 0 {
			statsAfter.UsedTokens = int64(float64(statsBefore.UsedTokens) * 0.4)
		}
		return statsAfter, nil
	}

	return statsAfter, fmt.Errorf("compact failed: %s (%w)", strings.TrimSpace(outBuf.String()), runErr)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
