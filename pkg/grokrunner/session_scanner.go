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

// ResolveWorkspaceSessionsDir returns the directory path under ~/.grok/sessions for workspacePath.
// It checks both the provided path and the canonical symlink-evaluated path (e.g. macOS /var -> /private/var),
// because Grok CLI uses Rust's std::fs::canonicalize when storing session directories.
func ResolveWorkspaceSessionsDir(sessionsDir, workspacePath string) string {
	if workspacePath == "" {
		return sessionsDir
	}
	cleanWs := filepath.Clean(workspacePath)

	// Collect candidate workspace paths (original clean path, canonical symlink resolved path,
	// and macOS /private variations)
	candidates := []string{cleanWs}
	if evalPath, err := filepath.EvalSymlinks(cleanWs); err == nil && evalPath != "" && evalPath != cleanWs {
		candidates = append(candidates, evalPath)
	}

	// On macOS, /var, /tmp, /etc symlink to /private/var, /private/tmp, /private/etc.
	// Rust std::fs::canonicalize resolves these to /private/..., while some tools keep /var/...
	var macVariations []string
	for _, c := range candidates {
		if strings.HasPrefix(c, "/private/") {
			unpriv := strings.TrimPrefix(c, "/private")
			macVariations = append(macVariations, unpriv)
		} else if strings.HasPrefix(c, "/var/") || strings.HasPrefix(c, "/tmp/") || strings.HasPrefix(c, "/etc/") {
			priv := "/private" + c
			macVariations = append(macVariations, priv)
		}
	}
	for _, mv := range macVariations {
		found := false
		for _, c := range candidates {
			if c == mv {
				found = true
				break
			}
		}
		if !found {
			candidates = append(candidates, mv)
		}
	}

	// Check each candidate for existing sessions directories
	for _, ws := range candidates {
		// 1. Try exact encoded path (with native clean path)
		primary := filepath.Join(sessionsDir, EncodeGrokWorkspacePath(ws))
		if fi, err := os.Stat(primary); err == nil && fi.IsDir() {
			return primary
		}

		// 1b. On Windows or cross-platform, try with forward-slashed and backslashed variations
		wsSlash := filepath.ToSlash(ws)
		slashEncoded := filepath.Join(sessionsDir, EncodeGrokWorkspacePath(wsSlash))
		if fi, err := os.Stat(slashEncoded); err == nil && fi.IsDir() {
			return slashEncoded
		}

		wsBackslash := strings.ReplaceAll(wsSlash, "/", "\\")
		backslashEncoded := filepath.Join(sessionsDir, EncodeGrokWorkspacePath(wsBackslash))
		if fi, err := os.Stat(backslashEncoded); err == nil && fi.IsDir() {
			return backslashEncoded
		}

		// 2. Try legacy url.PathEscape
		legacy := filepath.Join(sessionsDir, url.PathEscape(ws))
		if fi, err := os.Stat(legacy); err == nil && fi.IsDir() {
			return legacy
		}
	}

	// 3. Scan all directory entries in sessionsDir and match decoded folder names against all candidates
	entries, err := os.ReadDir(sessionsDir)
	if err == nil {
		for _, ws := range candidates {
			normalizedTarget := strings.ToLower(strings.TrimRight(filepath.ToSlash(ws), "/"))
			// Also compare target without trailing drive backslash if Windows drive root e.g. "c:"
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
					normalizedDecoded := strings.ToLower(strings.TrimRight(filepath.ToSlash(filepath.Clean(decoded)), "/"))
					if normalizedDecoded == normalizedTarget {
						return filepath.Join(sessionsDir, name)
					}
					// Also handle Windows drive letter cases e.g. "c:/" vs "c:"
					if strings.TrimSuffix(normalizedDecoded, "/") == strings.TrimSuffix(normalizedTarget, "/") {
						return filepath.Join(sessionsDir, name)
					}
				}
			}
		}
	}

	// Fallback to canonical encoded path if evaluated path differs and already exists, else primary
	if len(candidates) > 1 {
		canonicalDir := filepath.Join(sessionsDir, EncodeGrokWorkspacePath(candidates[1]))
		if fi, err := os.Stat(canonicalDir); err == nil && fi.IsDir() {
			return canonicalDir
		}
	}

	return filepath.Join(sessionsDir, EncodeGrokWorkspacePath(cleanWs))
}

// ResolveSessionFolder strictly resolves the session folder path and true Grok UUID inside targetDir.
// It only matches exact or case-insensitive UUID folder names, and never falls back to
// unrelated session folders to avoid cross-session contamination.
func ResolveSessionFolder(targetDir, sessionID string) (string, string) {
	if sessionID == "" {
		return "", ""
	}

	exactPath := filepath.Join(targetDir, sessionID)
	if fi, err := os.Stat(exactPath); err == nil && fi.IsDir() {
		return exactPath, sessionID
	}

	entries, err := os.ReadDir(targetDir)
	if err != nil || len(entries) == 0 {
		return filepath.Join(targetDir, sessionID), sessionID
	}

	// Search for matching folder name (case-insensitive or prefix match for truncated UUIDs)
	sLower := strings.ToLower(sessionID)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		eName := entry.Name()
		eLower := strings.ToLower(eName)
		if eLower == sLower {
			return filepath.Join(targetDir, eName), eName
		}
	}

	// Secondary check: if sessionID is at least an 8-character UUID prefix, match it
	if len(sessionID) >= 8 && isUUIDPrefix(sessionID) {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			eName := entry.Name()
			if strings.HasPrefix(strings.ToLower(eName), sLower) {
				return filepath.Join(targetDir, eName), eName
			}
		}
	}

	return filepath.Join(targetDir, sessionID), sessionID
}

func isUUIDPrefix(str string) bool {
	for _, ch := range str {
		if ch == '-' {
			continue
		}
		if !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			return false
		}
	}
	return true
}

func hasSessionRecord(folderPath string) bool {
	files := []string{
		"signals.json",
		"usage.json",
		"chat_history.jsonl",
		"summary.json",
		"updates.jsonl",
		"events.jsonl",
		"prompt_context.json",
	}
	for _, f := range files {
		if _, err := os.Stat(filepath.Join(folderPath, f)); err == nil {
			return true
		}
	}
	return false
}

// SessionFolderExists checks whether an initialized session directory actually exists on disk in targetWsDir.
// Cross-workspace fallback is strictly prohibited to guarantee session isolation between different workspaces.
func SessionFolderExists(sessionsDir, targetWsDir, sessionID string) (bool, string, string) {
	if sessionID == "" || targetWsDir == "" {
		return false, "", ""
	}

	folderPath, resolvedID := ResolveSessionFolder(targetWsDir, sessionID)
	if fi, err := os.Stat(folderPath); err == nil && fi.IsDir() {
		if hasSessionRecord(folderPath) {
			return true, folderPath, resolvedID
		}
	}

	return false, "", ""
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
				SessionKind     string `json:"session_kind"`
				NumMessages     int    `json:"num_messages"`
				NumChatMessages int    `json:"num_chat_messages"`
				CreatedAt       string `json:"created_at"`
				UpdatedAt       string `json:"updated_at"`
			}
			if json.Unmarshal(data, &summaryObj) == nil {
				// Exclude internal subagent sessions from user-visible session lists
				if strings.HasPrefix(strings.ToLower(summaryObj.SessionKind), "subagent") {
					continue
				}

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

func dirExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

func fileExists(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && !fi.IsDir()
}

// LoadSubagentExecutionDetails loads and parses the complete execution state,
// status, final output, duration, and child tool calls for a subagent from its session directory.
func LoadSubagentExecutionDetails(workspacePath, subagentID string) (*SubagentExecutionDetails, error) {
	subagentID = strings.TrimSpace(subagentID)
	if subagentID == "" {
		return &SubagentExecutionDetails{SubagentID: subagentID, Status: "running", Calls: []DiscoveredToolCall{}}, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	var sessionFolderPath string

	hasSessionFiles := func(dir string) bool {
		return fileExists(filepath.Join(dir, "updates.jsonl")) ||
			fileExists(filepath.Join(dir, "chat_history.jsonl")) ||
			fileExists(filepath.Join(dir, "summary.json"))
	}

	// 1. Try finding inside specified workspace
	if workspacePath != "" {
		targetDir := ResolveWorkspaceSessionsDir(sessionsDir, workspacePath)
		folder, _ := ResolveSessionFolder(targetDir, subagentID)
		if dirExists(folder) && hasSessionFiles(folder) {
			sessionFolderPath = folder
		} else if dirExists(targetDir) {
			// Check <parent>/subagents/<subagentID>
			if entries, err := os.ReadDir(targetDir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						subDir := filepath.Join(targetDir, entry.Name(), "subagents", subagentID)
						if dirExists(subDir) && hasSessionFiles(subDir) {
							sessionFolderPath = subDir
							break
						}
					}
				}
			}
		}
	}

	// 2. Fallback: Search all workspace folders under ~/.grok/sessions/ if not found yet
	if sessionFolderPath == "" {
		if wsEntries, err := os.ReadDir(sessionsDir); err == nil {
			sLower := strings.ToLower(subagentID)
			for _, wsEntry := range wsEntries {
				if !wsEntry.IsDir() {
					continue
				}
				candidateDir := filepath.Join(sessionsDir, wsEntry.Name())
				candidateFolder, candidateID := ResolveSessionFolder(candidateDir, subagentID)
				if dirExists(candidateFolder) && hasSessionFiles(candidateFolder) &&
					(strings.EqualFold(candidateID, subagentID) || strings.HasPrefix(strings.ToLower(candidateID), sLower)) {
					sessionFolderPath = candidateFolder
					break
				}
				// Also check subagents/ within this workspace
				if pEntries, err := os.ReadDir(candidateDir); err == nil {
					for _, pEntry := range pEntries {
						if pEntry.IsDir() {
							subDir := filepath.Join(candidateDir, pEntry.Name(), "subagents", subagentID)
							if dirExists(subDir) && hasSessionFiles(subDir) {
								sessionFolderPath = subDir
								break
							}
						}
					}
				}
				if sessionFolderPath != "" {
					break
				}
			}
		}
	}

	if sessionFolderPath == "" || !dirExists(sessionFolderPath) {
		return &SubagentExecutionDetails{
			SubagentID: subagentID,
			Status:     "running",
			Calls:      []DiscoveredToolCall{},
		}, nil
	}

	details := &SubagentExecutionDetails{
		SubagentID: subagentID,
		Status:     "running",
		Calls:      []DiscoveredToolCall{},
	}

	// Parse tool calls
	updatesPath := filepath.Join(sessionFolderPath, "updates.jsonl")
	chatPath := filepath.Join(sessionFolderPath, "chat_history.jsonl")
	summaryPath := filepath.Join(sessionFolderPath, "summary.json")

	if fileExists(updatesPath) {
		if calls, err := ParseSubagentUpdatesJsonl(updatesPath); err == nil && len(calls) > 0 {
			details.Calls = calls
		}
	}
	if len(details.Calls) == 0 && fileExists(chatPath) {
		if calls, err := ParseSubagentChatHistoryJsonl(chatPath); err == nil && len(calls) > 0 {
			details.Calls = calls
		}
	}

	// Read updates.jsonl for turn_completed status, elapsed duration, and message chunks
	var lastAssistantChunks []string
	if fileExists(updatesPath) {
		if data, err := os.ReadFile(updatesPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				var raw map[string]interface{}
				if err := json.Unmarshal([]byte(line), &raw); err != nil {
					continue
				}
				params, _ := raw["params"].(map[string]interface{})
				if params == nil {
					continue
				}
				updateMap, _ := params["update"].(map[string]interface{})
				if updateMap == nil {
					continue
				}
				uType, _ := updateMap["sessionUpdate"].(string)
				if uType == "" {
					uType, _ = updateMap["type"].(string)
				}
				uType = strings.ToLower(strings.TrimSpace(uType))

				if uType == "turn_completed" {
					stopReason, _ := updateMap["stop_reason"].(string)
					stopReason = strings.ToLower(strings.TrimSpace(stopReason))
					if stopReason == "end_turn" || stopReason == "stop" || stopReason == "complete" || stopReason == "" {
						details.Status = "completed"
					} else if stopReason == "error" || stopReason == "failed" {
						details.Status = "failed"
					}
					if elapsed, ok := updateMap["elapsed_ms"].(float64); ok && elapsed > 0 {
						details.DurationMs = int64(elapsed)
					}
				} else if uType == "agent_message_chunk" || uType == "agent_message" {
					if contentMap, ok := updateMap["content"].(map[string]interface{}); ok {
						if text, ok := contentMap["text"].(string); ok && text != "" {
							lastAssistantChunks = append(lastAssistantChunks, text)
						}
					}
				} else if uType == "tool_call" {
					lastAssistantChunks = nil
				}
			}
		}
	}

	// Read chat_history.jsonl for complete final assistant output
	if fileExists(chatPath) {
		if data, err := os.ReadFile(chatPath); err == nil {
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				var raw map[string]interface{}
				if err := json.Unmarshal([]byte(line), &raw); err != nil {
					continue
				}
				role, _ := raw["role"].(string)
				if role == "" {
					role, _ = raw["type"].(string)
				}
				if role == "assistant" {
					if text, ok := raw["content"].(string); ok && strings.TrimSpace(text) != "" {
						details.Output = strings.TrimSpace(text)
						if details.Status != "failed" {
							details.Status = "completed"
						}
					}
				}
			}
		}
	}

	// Fallback output from updates.jsonl chunks if chat_history didn't have it
	if details.Output == "" && len(lastAssistantChunks) > 0 {
		details.Output = strings.TrimSpace(strings.Join(lastAssistantChunks, ""))
	}

	// Read summary.json for duration fallback
	if fileExists(summaryPath) {
		if data, err := os.ReadFile(summaryPath); err == nil {
			var summary map[string]interface{}
			if err := json.Unmarshal(data, &summary); err == nil {
				createdAtStr, _ := summary["created_at"].(string)
				updatedAtStr, _ := summary["updated_at"].(string)
				if createdAtStr != "" && updatedAtStr != "" {
					tCreated, err1 := time.Parse(time.RFC3339Nano, createdAtStr)
					if err1 != nil {
						tCreated, _ = time.Parse(time.RFC3339, createdAtStr)
					}
					tUpdated, err2 := time.Parse(time.RFC3339Nano, updatedAtStr)
					if err2 != nil {
						tUpdated, _ = time.Parse(time.RFC3339, updatedAtStr)
					}
					if !tCreated.IsZero() && !tUpdated.IsZero() && tUpdated.After(tCreated) {
						diffMs := tUpdated.Sub(tCreated).Milliseconds()
						if details.DurationMs == 0 {
							details.DurationMs = diffMs
						}
					}
				}
			}
		}
	}

	// If all child tool calls are completed, and we have output, ensure status is completed
	if details.Status == "running" && details.Output != "" && len(details.Calls) > 0 {
		allCompleted := true
		for _, c := range details.Calls {
			if c.Status == "running" {
				allCompleted = false
				break
			}
		}
		if allCompleted {
			details.Status = "completed"
		}
	}

	return details, nil
}

// LoadSubagentToolCalls loads and parses all tool calls executed by a child subagent from its session directory on disk.
func LoadSubagentToolCalls(workspacePath, subagentID string) ([]DiscoveredToolCall, error) {
	details, err := LoadSubagentExecutionDetails(workspacePath, subagentID)
	if err != nil {
		return nil, err
	}
	return details.Calls, nil
}

// ParseSubagentUpdatesJsonl parses tool calls from an updates.jsonl file
func ParseSubagentUpdatesJsonl(updatesPath string) ([]DiscoveredToolCall, error) {
	data, err := os.ReadFile(updatesPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	toolCallMap := make(map[string]*DiscoveredToolCall)
	var order []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		params, _ := raw["params"].(map[string]interface{})
		if params == nil {
			continue
		}

		updateMap, _ := params["update"].(map[string]interface{})
		if updateMap == nil {
			continue
		}

		sessionUpdate, _ := updateMap["sessionUpdate"].(string)
		if sessionUpdate == "" {
			sessionUpdate, _ = updateMap["type"].(string)
		}
		sessionUpdate = strings.ToLower(strings.TrimSpace(sessionUpdate))

		toolCallID, _ := updateMap["toolCallId"].(string)
		if toolCallID == "" {
			toolCallID, _ = updateMap["tool_call_id"].(string)
		}
		if toolCallID == "" {
			continue
		}

		var ts int64
		if tNum, ok := raw["timestamp"].(float64); ok && tNum > 0 {
			ts = int64(tNum)
			if ts < 10000000000 {
				ts *= 1000
			}
		} else if meta, ok := raw["_meta"].(map[string]interface{}); ok {
			if agMs, ok := meta["agentTimestampMs"].(float64); ok && agMs > 0 {
				ts = int64(agMs)
			}
		}

		if sessionUpdate == "tool_call" {
			toolName := ""
			if meta, ok := updateMap["_meta"].(map[string]interface{}); ok {
				if xTool, ok := meta["x.ai/tool"].(map[string]interface{}); ok {
					if n, ok := xTool["name"].(string); ok && n != "" {
						toolName = n
					}
				}
			}
			if toolName == "" {
				toolName, _ = updateMap["title"].(string)
			}
			if toolName == "" {
				toolName, _ = updateMap["name"].(string)
			}

			inputMap := make(map[string]interface{})
			if rawInput, ok := updateMap["rawInput"].(map[string]interface{}); ok {
				inputMap = rawInput
			} else if toolInput, ok := updateMap["toolInput"].(map[string]interface{}); ok {
				inputMap = toolInput
			}

			tc, exists := toolCallMap[toolCallID]
			if !exists {
				tc = &DiscoveredToolCall{
					ID:        toolCallID,
					Tool:      toolName,
					Params:    inputMap,
					Status:    "running",
					StartTime: ts,
				}
				toolCallMap[toolCallID] = tc
				order = append(order, toolCallID)
			} else {
				if tc.Tool == "" {
					tc.Tool = toolName
				}
				if len(tc.Params) == 0 && len(inputMap) > 0 {
					tc.Params = inputMap
				}
				if tc.StartTime == 0 {
					tc.StartTime = ts
				}
			}
		} else if sessionUpdate == "tool_call_update" {
			tc, exists := toolCallMap[toolCallID]
			if !exists {
				toolName, _ := updateMap["title"].(string)
				if meta, ok := updateMap["_meta"].(map[string]interface{}); ok {
					if xTool, ok := meta["x.ai/tool"].(map[string]interface{}); ok {
						if n, ok := xTool["name"].(string); ok && n != "" {
							toolName = n
						}
					}
				}
				inputMap := make(map[string]interface{})
				if rawInput, ok := updateMap["rawInput"].(map[string]interface{}); ok {
					inputMap = rawInput
				}
				tc = &DiscoveredToolCall{
					ID:        toolCallID,
					Tool:      toolName,
					Params:    inputMap,
					Status:    "running",
					StartTime: ts,
				}
				toolCallMap[toolCallID] = tc
				order = append(order, toolCallID)
			}

			rawStatus, _ := updateMap["status"].(string)
			statusLower := strings.ToLower(strings.TrimSpace(rawStatus))
			if statusLower == "completed" || statusLower == "success" || statusLower == "done" {
				tc.Status = "completed"
			} else if statusLower == "failed" || statusLower == "error" || statusLower == "rejected" {
				tc.Status = "failed"
			}

			out := extractToolOutput(updateMap["rawOutput"], updateMap["content"])
			if out != "" {
				tc.Result = out
				if tc.Status == "running" && (statusLower == "" || statusLower == "completed") {
					tc.Status = "completed"
				}
			}

			if ts > 0 {
				tc.EndTime = ts
			}

			if tc.Tool == "" {
				if title, ok := updateMap["title"].(string); ok && title != "" {
					tc.Tool = title
				}
			}
			if len(tc.Params) == 0 {
				if rawInput, ok := updateMap["rawInput"].(map[string]interface{}); ok {
					tc.Params = rawInput
				}
			}
		}
	}

	result := make([]DiscoveredToolCall, 0, len(order))
	for _, id := range order {
		if tc, ok := toolCallMap[id]; ok {
			result = append(result, *tc)
		}
	}
	return result, nil
}

// ParseSubagentChatHistoryJsonl parses tool calls from a subagent chat_history.jsonl file
func ParseSubagentChatHistoryJsonl(chatPath string) ([]DiscoveredToolCall, error) {
	data, err := os.ReadFile(chatPath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	toolCallMap := make(map[string]*DiscoveredToolCall)
	var order []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			continue
		}

		role, _ := raw["role"].(string)
		if role == "" {
			role, _ = raw["type"].(string)
		}

		if role == "assistant" {
			if tcList, ok := raw["tool_calls"].([]interface{}); ok {
				for _, tcItem := range tcList {
					if tcMap, ok := tcItem.(map[string]interface{}); ok {
						id, _ := tcMap["id"].(string)
						name, _ := tcMap["name"].(string)
						var params map[string]interface{}
						if argsStr, ok := tcMap["arguments"].(string); ok {
							_ = json.Unmarshal([]byte(argsStr), &params)
						} else if argsObj, ok := tcMap["arguments"].(map[string]interface{}); ok {
							params = argsObj
						}
						if id != "" {
							if _, exists := toolCallMap[id]; !exists {
								newTc := &DiscoveredToolCall{
									ID:     id,
									Tool:   name,
									Params: params,
									Status: "completed",
								}
								toolCallMap[id] = newTc
								order = append(order, id)
							}
						}
					}
				}
			}
		} else if role == "tool_result" {
			id, _ := raw["tool_call_id"].(string)
			resultStr := ""
			if cStr, ok := raw["content"].(string); ok {
				resultStr = cStr
			}
			if tc, exists := toolCallMap[id]; exists {
				tc.Result = resultStr
				tc.Status = "completed"
			}
		}
	}

	result := make([]DiscoveredToolCall, 0, len(order))
	for _, id := range order {
		if tc, ok := toolCallMap[id]; ok {
			result = append(result, *tc)
		}
	}
	return result, nil
}
