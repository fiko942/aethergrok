package grokrunner

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
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
var subagentIDRegex = regexp.MustCompile(`(?i)(?:subagent_id|task_id|session_id)["']?\s*[:=]\s*[`+"`"+`"']?([a-zA-Z0-9_-]{4,})[`+"`"+`"']?`)
var uuidFallbackRegex = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

type sessionCacheEntry struct {
	modTime  time.Time
	fileSize int64
	messages []DiscoveredChatMessage
}

var (
	sessionCacheMu sync.RWMutex
	sessionCache   = make(map[string]*sessionCacheEntry)
)

// InvalidateSessionCache clears the in-memory parsed history cache for a session folder
func InvalidateSessionCache(sessionFolderPath string) {
	if sessionFolderPath == "" {
		return
	}
	sessionCacheMu.Lock()
	defer sessionCacheMu.Unlock()
	for k := range sessionCache {
		if strings.HasPrefix(k, sessionFolderPath) {
			delete(sessionCache, k)
		}
	}
}

// ClearAllSessionCache clears the entire in-memory session history cache
func ClearAllSessionCache() {
	sessionCacheMu.Lock()
	defer sessionCacheMu.Unlock()
	sessionCache = make(map[string]*sessionCacheEntry)
}

// forEachJSONLLine reads a JSONL file line-by-line using a streaming bufio.Reader
// without reading the entire file into memory at once, avoiding high RAM spikes on multi-megabyte files.
func forEachJSONLLine(filePath string, maxLines int, fn func(lineIdx int, lineBytes []byte) bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 64*1024)
	lineIdx := 0

	for {
		if maxLines > 0 && lineIdx >= maxLines {
			break
		}

		line, err := reader.ReadBytes('\n')
		if len(line) > 0 {
			trimmed := bytes.TrimSpace(line)
			if len(trimmed) > 0 {
				if !fn(lineIdx, trimmed) {
					break
				}
				lineIdx++
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}
	return nil
}

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
			if title != "" {
				break
			}
			_ = forEachJSONLLine(hFile, 100, func(lineIdx int, lineBytes []byte) bool {
				var rawMsg map[string]interface{}
				if json.Unmarshal(lineBytes, &rawMsg) != nil {
					return true
				}

				role, _ := rawMsg["role"].(string)
				msgType, _ := rawMsg["type"].(string)
				if role != "user" && msgType != "user" {
					return true
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
					} else {
						if strings.Contains(contentStr, "<system-reminder>") {
							continue
						}
						if strings.HasPrefix(contentStr, "<user_info>") {
							lastTagEnd := strings.LastIndex(contentStr, ">")
							if lastTagEnd != -1 && lastTagEnd < len(contentStr)-1 {
								trailing := strings.TrimSpace(contentStr[lastTagEnd+1:])
								if trailing != "" {
									extracted = trailing
								} else {
									continue
								}
							} else {
								continue
							}
						} else {
							extracted = strings.TrimSpace(contentStr)
						}
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
						return false // Found title, early exit stream!
					}
				}
				return true
			})
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
	var fi os.FileInfo
	if info, err := os.Stat(chatHistoryPath); err == nil {
		fi = info
	} else {
		// Fallback to updates.jsonl if chat_history doesn't exist
		chatHistoryPath = filepath.Join(sessionFolderPath, "updates.jsonl")
		if info, err := os.Stat(chatHistoryPath); err == nil {
			fi = info
		} else {
			return []DiscoveredChatMessage{}, nil
		}
	}

	// 1. Check in-memory ModTime and FileSize cache
	if fi != nil {
		sessionCacheMu.RLock()
		cached, exists := sessionCache[chatHistoryPath]
		sessionCacheMu.RUnlock()
		if exists && cached.modTime.Equal(fi.ModTime()) && cached.fileSize == fi.Size() {
			return cached.messages, nil
		}
	}

	var messages []DiscoveredChatMessage

	_ = forEachJSONLLine(chatHistoryPath, 0, func(lineIdx int, lineBytes []byte) bool {
		var raw map[string]interface{}
		if json.Unmarshal(lineBytes, &raw) != nil {
			return true
		}

		msgType, _ := raw["type"].(string)
		role, _ := raw["role"].(string)
		if role == "" {
			role = msgType
		}

		// Skip internal system prompts, reasoning, or compaction meta
		if role == "system" || role == "reasoning" || msgType == "reasoning" {
			return true
		}

		// 1. Process User Message
		if role == "user" {
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
			matches := userQueryRegex.FindStringSubmatch(userText)
			if len(matches) > 1 {
				userText = strings.TrimSpace(matches[1])
			} else {
				// Pure system reminders (injected by CLI harness without user query) must NOT count as user turns
				if strings.Contains(userText, "<system-reminder>") {
					return true
				}
				if strings.HasPrefix(userText, "<user_info>") {
					lastTagEnd := strings.LastIndex(userText, ">")
					if lastTagEnd != -1 && lastTagEnd < len(userText)-1 {
						trailing := strings.TrimSpace(userText[lastTagEnd+1:])
						if trailing != "" {
							userText = trailing
						} else {
							return true
						}
					} else {
						return true
					}
				}
			}

			if userText == "" {
				return true
			}

			messages = append(messages, DiscoveredChatMessage{
				ID:        fmt.Sprintf("%s_u_%d", sessionID, lineIdx),
				Role:      "user",
				Content:   userText,
				Timestamp: 0,
				Status:    "done",
				Parts: []DiscoveredMessagePart{
					{
						Type: "text",
						Text: userText,
					},
				},
			})
		} else if role == "assistant" {
			contentStr := ""
			if cStr, ok := raw["content"].(string); ok {
				contentStr = strings.TrimSpace(cStr)
			}
			tcList, _ := raw["tool_calls"].([]interface{})

			// Prevent orphan assistant messages before the first user turn if blank and no tool calls
			if len(messages) == 0 && contentStr == "" && len(tcList) == 0 {
				return true
			}

			// Ensure an assistant message exists for this turn
			if len(messages) == 0 || messages[len(messages)-1].Role != "assistant" {
				messages = append(messages, DiscoveredChatMessage{
					ID:        fmt.Sprintf("%s_a_%d", sessionID, lineIdx),
					Role:      "assistant",
					Status:    "done",
				})
			}
			asst := &messages[len(messages)-1]

			// In assistant turns, any introductory text was generated before tool calls
			if contentStr != "" {
				if asst.Content != "" {
					asst.Content += "\n\n" + contentStr
				} else {
					asst.Content = contentStr
				}

				if len(asst.Parts) > 0 && asst.Parts[len(asst.Parts)-1].Type == "text" {
					asst.Parts[len(asst.Parts)-1].Text += "\n\n" + contentStr
				} else {
					asst.Parts = append(asst.Parts, DiscoveredMessagePart{
						Type: "text",
						Text: contentStr,
					})
				}
			}

			// Tool calls in this assistant entry follow the preceding text
			if tcList, ok := raw["tool_calls"].([]interface{}); ok && len(tcList) > 0 {
				var lineToolCalls []DiscoveredToolCall
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
						lineToolCalls = append(lineToolCalls, newTc)
					}
				}

				if len(lineToolCalls) > 0 {
					asst.ToolCalls = append(asst.ToolCalls, lineToolCalls...)
					if len(asst.Parts) > 0 && asst.Parts[len(asst.Parts)-1].Type == "tool_calls" {
						asst.Parts[len(asst.Parts)-1].ToolCalls = append(asst.Parts[len(asst.Parts)-1].ToolCalls, lineToolCalls...)
					} else {
						asst.Parts = append(asst.Parts, DiscoveredMessagePart{
							Type:      "tool_calls",
							ToolCalls: lineToolCalls,
						})
					}
				}
			}
		} else if role == "tool_result" || msgType == "tool_result" {
			// Populate result into matched tool call
			tcID, _ := raw["tool_call_id"].(string)
			resultStr := ""
			if cStr, ok := raw["content"].(string); ok {
				resultStr = cStr
			}

			if len(messages) > 0 && messages[len(messages)-1].Role == "assistant" {
				asst := &messages[len(messages)-1]
				for i := range asst.ToolCalls {
					if asst.ToolCalls[i].ID == tcID {
						asst.ToolCalls[i].Result = resultStr
						break
					}
				}
				for i := range asst.Parts {
					if asst.Parts[i].Type == "tool_calls" {
						for j := range asst.Parts[i].ToolCalls {
							if asst.Parts[i].ToolCalls[j].ID == tcID {
								asst.Parts[i].ToolCalls[j].Result = resultStr
								break
							}
						}
					}
				}
			}
		}
		return true
	})

	// Post-process assistant messages: populate subagent details and child tool calls from disk
	for msgIdx := range messages {
		if messages[msgIdx].Role != "assistant" {
			continue
		}
		asst := &messages[msgIdx]

		for tcIdx := range asst.ToolCalls {
			tc := &asst.ToolCalls[tcIdx]
			toolLower := strings.ToLower(tc.Tool)
			isSub := strings.Contains(toolLower, "subagent") || toolLower == "delegatetask" || toolLower == "runagent"
			if !isSub && tc.Params != nil {
				if _, ok := tc.Params["subagent_type"]; ok {
					isSub = true
				} else if isBg, _ := tc.Params["is_background"].(bool); isBg {
					if desc, _ := tc.Params["description"].(string); strings.Contains(strings.ToLower(desc), "subagent") {
						isSub = true
					}
				}
			}

			if isSub {
				subID := ""
				if tc.Params != nil {
					if sid, ok := tc.Params["subagent_id"].(string); ok && sid != "" {
						subID = sid
					} else if tid, ok := tc.Params["task_id"].(string); ok && tid != "" {
						subID = tid
					} else if tid, ok := tc.Params["id"].(string); ok && tid != "" {
						subID = tid
					}
				}
				if subID == "" && tc.Result != "" {
					matches := subagentIDRegex.FindStringSubmatch(tc.Result)
					if len(matches) > 1 {
						subID = matches[1]
					} else {
						uMatches := uuidFallbackRegex.FindStringSubmatch(tc.Result)
						if len(uMatches) > 0 {
							subID = uMatches[0]
						}
					}
				}

				// Disk folder fallback: check <sessionFolderPath>/subagents/ if subID is still not found in text
				if subID == "" && sessionFolderPath != "" {
					subagentsParent := filepath.Join(sessionFolderPath, "subagents")
					if dirExists(subagentsParent) {
						if subEntries, err := os.ReadDir(subagentsParent); err == nil {
							var candidateIDs []string
							for _, se := range subEntries {
								if se.IsDir() {
									candidateIDs = append(candidateIDs, se.Name())
								}
							}
							if len(candidateIDs) == 1 {
								subID = candidateIDs[0]
							} else if len(candidateIDs) > 1 && tc.Params != nil {
								desc, _ := tc.Params["description"].(string)
								prompt, _ := tc.Params["prompt"].(string)
								for _, cID := range candidateIDs {
									metaFile := filepath.Join(subagentsParent, cID, "meta.json")
									if data, err := os.ReadFile(metaFile); err == nil {
										var m struct {
											Description string `json:"description"`
											Prompt      string `json:"prompt"`
										}
										if json.Unmarshal(data, &m) == nil {
											if (desc != "" && strings.Contains(strings.ToLower(m.Description), strings.ToLower(desc))) ||
												(prompt != "" && strings.Contains(strings.ToLower(m.Prompt), strings.ToLower(prompt))) {
												subID = cID
												break
											}
										}
									}
								}
							}
						}
					}
				}

				if subID != "" {
					tc.ChildSessionID = subID
					if subDetails, err := LoadSubagentExecutionDetails(workspacePath, subID); err == nil && subDetails != nil {
						if len(subDetails.Calls) > 0 {
							tc.ChildToolCalls = subDetails.Calls
						}
						if subDetails.Output != "" && (tc.Result == "" || strings.Contains(strings.ToLower(tc.Result), "subagent started in background")) {
							tc.Result = subDetails.Output
						}
						if subDetails.Status == "completed" || subDetails.Status == "failed" {
							tc.Status = subDetails.Status
						}
						if subDetails.StartTime > 0 && tc.StartTime == 0 {
							tc.StartTime = subDetails.StartTime
						}
						if subDetails.EndTime > 0 && tc.EndTime == 0 {
							tc.EndTime = subDetails.EndTime
						}
						if subDetails.DurationMs > 0 && tc.StartTime > 0 && tc.EndTime == 0 {
							tc.EndTime = tc.StartTime + subDetails.DurationMs
						}
						if tc.Params == nil {
							tc.Params = make(map[string]interface{})
						}
						if _, ok := tc.Params["description"]; !ok && subDetails.Description != "" {
							tc.Params["description"] = subDetails.Description
						}
						if _, ok := tc.Params["prompt"]; !ok && subDetails.Prompt != "" {
							tc.Params["prompt"] = subDetails.Prompt
						}
						if _, ok := tc.Params["subagent_type"]; !ok && subDetails.SubagentType != "" {
							tc.Params["subagent_type"] = subDetails.SubagentType
						}
						if _, ok := tc.Params["model"]; !ok && subDetails.Model != "" {
							tc.Params["model"] = subDetails.Model
						}
					}
				}
			}
		}

		// Synchronize updated tool calls into asst.Parts
		for partIdx := range asst.Parts {
			if asst.Parts[partIdx].Type == "tool_calls" {
				for ptcIdx := range asst.Parts[partIdx].ToolCalls {
					ptc := &asst.Parts[partIdx].ToolCalls[ptcIdx]
					for _, tc := range asst.ToolCalls {
						if tc.ID == ptc.ID {
							ptc.ChildToolCalls = tc.ChildToolCalls
							ptc.ChildSessionID = tc.ChildSessionID
							ptc.Result = tc.Result
							ptc.Status = tc.Status
							ptc.StartTime = tc.StartTime
							ptc.EndTime = tc.EndTime
							ptc.Params = tc.Params
							break
						}
					}
				}
			}
		}
	}

	// Cache parsed messages with ModTime and FileSize to make subsequent reads 0.01ms
	if fi != nil {
		sessionCacheMu.Lock()
		sessionCache[chatHistoryPath] = &sessionCacheEntry{
			modTime:  fi.ModTime(),
			fileSize: fi.Size(),
			messages: messages,
		}
		sessionCacheMu.Unlock()
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

	// Invalidate in-memory session cache for this folder
	InvalidateSessionCache(sessionFolderPath)
	if resolvedID != sessionID && sessionID != "" {
		InvalidateSessionCache(filepath.Join(sessionsParent, sessionID))
	}

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

	hasSubagentFiles := func(dir string) bool {
		return fileExists(filepath.Join(dir, "meta.json")) ||
			fileExists(filepath.Join(dir, "output.json")) ||
			fileExists(filepath.Join(dir, "updates.jsonl")) ||
			fileExists(filepath.Join(dir, "chat_history.jsonl")) ||
			fileExists(filepath.Join(dir, "summary.json"))
	}

	var candidateDirs []string
	addCandidate := func(dir string) {
		if dir == "" || !dirExists(dir) || !hasSubagentFiles(dir) {
			return
		}
		for _, c := range candidateDirs {
			if c == dir {
				return
			}
		}
		candidateDirs = append(candidateDirs, dir)
	}

	sLower := strings.ToLower(subagentID)

	// 1. Try finding inside specified workspace
	if workspacePath != "" {
		targetDir := ResolveWorkspaceSessionsDir(sessionsDir, workspacePath)
		if dirExists(targetDir) {
			folder, candidateID := ResolveSessionFolder(targetDir, subagentID)
			if strings.EqualFold(candidateID, subagentID) || strings.HasPrefix(strings.ToLower(candidateID), sLower) {
				addCandidate(folder)
			}
			// Check <parent>/subagents/<subagentID>
			if entries, err := os.ReadDir(targetDir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						subagentsParent := filepath.Join(targetDir, entry.Name(), "subagents")
						if dirExists(subagentsParent) {
							subFolder, sID := ResolveSessionFolder(subagentsParent, subagentID)
							if strings.EqualFold(sID, subagentID) || strings.HasPrefix(strings.ToLower(sID), sLower) {
								addCandidate(subFolder)
							}
							addCandidate(filepath.Join(subagentsParent, subagentID))
						}
					}
				}
			}
		}
	}

	// 2. Search all workspace folders under ~/.grok/sessions/
	if wsEntries, err := os.ReadDir(sessionsDir); err == nil {
		for _, wsEntry := range wsEntries {
			if !wsEntry.IsDir() {
				continue
			}
			candidateDir := filepath.Join(sessionsDir, wsEntry.Name())
			candidateFolder, candidateID := ResolveSessionFolder(candidateDir, subagentID)
			if strings.EqualFold(candidateID, subagentID) || strings.HasPrefix(strings.ToLower(candidateID), sLower) {
				addCandidate(candidateFolder)
			}
			// Also check subagents/ within this workspace
			if pEntries, err := os.ReadDir(candidateDir); err == nil {
				for _, pEntry := range pEntries {
					if pEntry.IsDir() {
						subagentsParent := filepath.Join(candidateDir, pEntry.Name(), "subagents")
						if dirExists(subagentsParent) {
							subFolder, sID := ResolveSessionFolder(subagentsParent, subagentID)
							if strings.EqualFold(sID, subagentID) || strings.HasPrefix(strings.ToLower(sID), sLower) {
								addCandidate(subFolder)
							}
							addCandidate(filepath.Join(subagentsParent, subagentID))
						}
					}
				}
			}
		}
	}

	if len(candidateDirs) == 0 {
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

	var lastAssistantChunks []string

	for _, dir := range candidateDirs {
		// A. Parse meta.json if present
		metaPath := filepath.Join(dir, "meta.json")
		if fileExists(metaPath) {
			if data, err := os.ReadFile(metaPath); err == nil {
				var meta struct {
					Status           string `json:"status"`
					DurationMs       int64  `json:"duration_ms"`
					StartedAt        string `json:"started_at"`
					CompletedAt      string `json:"completed_at"`
					ToolCalls        int    `json:"tool_calls"`
					Description      string `json:"description"`
					Prompt           string `json:"prompt"`
					SubagentType     string `json:"subagent_type"`
					EffectiveModelID string `json:"effective_model_id"`
				}
				if err := json.Unmarshal(data, &meta); err == nil {
					mStatus := strings.ToLower(strings.TrimSpace(meta.Status))
					if mStatus == "completed" || mStatus == "success" || mStatus == "done" {
						details.Status = "completed"
					} else if mStatus == "failed" || mStatus == "error" {
						details.Status = "failed"
					}
					var tStart, tEnd time.Time
					if meta.StartedAt != "" {
						var err1 error
						tStart, err1 = time.Parse(time.RFC3339Nano, meta.StartedAt)
						if err1 != nil {
							tStart, _ = time.Parse(time.RFC3339, meta.StartedAt)
						}
						if !tStart.IsZero() && details.StartTime == 0 {
							details.StartTime = tStart.UnixMilli()
						}
					}
					if meta.CompletedAt != "" {
						var err2 error
						tEnd, err2 = time.Parse(time.RFC3339Nano, meta.CompletedAt)
						if err2 != nil {
							tEnd, _ = time.Parse(time.RFC3339, meta.CompletedAt)
						}
						if !tEnd.IsZero() && details.EndTime == 0 {
							details.EndTime = tEnd.UnixMilli()
						}
					}
					if meta.DurationMs > 0 && details.DurationMs == 0 {
						details.DurationMs = meta.DurationMs
					} else if details.DurationMs == 0 && !tStart.IsZero() && !tEnd.IsZero() && tEnd.After(tStart) {
						details.DurationMs = tEnd.Sub(tStart).Milliseconds()
					}
					if details.Description == "" && meta.Description != "" {
						details.Description = meta.Description
					}
					if details.Prompt == "" && meta.Prompt != "" {
						details.Prompt = meta.Prompt
					}
					if details.SubagentType == "" && meta.SubagentType != "" {
						details.SubagentType = meta.SubagentType
					}
					if details.Model == "" && meta.EffectiveModelID != "" {
						details.Model = meta.EffectiveModelID
					}
				}
			}
		}

		// B. Parse output.json if present
		outputPath := filepath.Join(dir, "output.json")
		if fileExists(outputPath) {
			if data, err := os.ReadFile(outputPath); err == nil {
				var outObj struct {
					Output string `json:"output"`
				}
				if err := json.Unmarshal(data, &outObj); err == nil && strings.TrimSpace(outObj.Output) != "" {
					if details.Output == "" {
						details.Output = strings.TrimSpace(outObj.Output)
					}
					if details.Status != "failed" {
						details.Status = "completed"
					}
				}
			}
		}

		// C. Parse child tool calls from updates.jsonl or chat_history.jsonl
		updatesPath := filepath.Join(dir, "updates.jsonl")
		chatPath := filepath.Join(dir, "chat_history.jsonl")
		summaryPath := filepath.Join(dir, "summary.json")

		if fileExists(updatesPath) {
			if len(details.Calls) == 0 {
				if calls, err := ParseSubagentUpdatesJsonl(updatesPath); err == nil && len(calls) > 0 {
					details.Calls = calls
				}
			}

			// Read turn_completed and chunks using streaming reader
			_ = forEachJSONLLine(updatesPath, 0, func(lineIdx int, lineBytes []byte) bool {
				var raw map[string]interface{}
				if err := json.Unmarshal(lineBytes, &raw); err != nil {
					return true
				}
				params, _ := raw["params"].(map[string]interface{})
				if params == nil {
					return true
				}
				updateMap, _ := params["update"].(map[string]interface{})
				if updateMap == nil {
					return true
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
						if details.Status != "failed" {
							details.Status = "completed"
						}
					} else if stopReason == "error" || stopReason == "failed" {
						details.Status = "failed"
					}
					if elapsed, ok := updateMap["elapsed_ms"].(float64); ok && elapsed > 0 && details.DurationMs == 0 {
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
				return true
			})
		}

		if len(details.Calls) == 0 && fileExists(chatPath) {
			if calls, err := ParseSubagentChatHistoryJsonl(chatPath); err == nil && len(calls) > 0 {
				details.Calls = calls
			}
		}

		// Read chat_history.jsonl for complete final assistant output using streaming reader
		if details.Output == "" && fileExists(chatPath) {
			_ = forEachJSONLLine(chatPath, 0, func(lineIdx int, lineBytes []byte) bool {
				var raw map[string]interface{}
				if err := json.Unmarshal(lineBytes, &raw); err != nil {
					return true
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
				return true
			})
		}

		// Read summary.json for duration fallback
		if details.DurationMs == 0 && fileExists(summaryPath) {
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
	}

	// Fallback output from updates.jsonl chunks if chat_history or output.json didn't have it
	if details.Output == "" && len(lastAssistantChunks) > 0 {
		details.Output = strings.TrimSpace(strings.Join(lastAssistantChunks, ""))
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

	if details.StartTime == 0 && len(details.Calls) > 0 && details.Calls[0].StartTime > 0 {
		details.StartTime = details.Calls[0].StartTime
	}
	if details.EndTime == 0 && len(details.Calls) > 0 && details.Calls[len(details.Calls)-1].EndTime > 0 {
		details.EndTime = details.Calls[len(details.Calls)-1].EndTime
	}
	if details.DurationMs == 0 && details.StartTime > 0 && details.EndTime > details.StartTime {
		details.DurationMs = details.EndTime - details.StartTime
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
	toolCallMap := make(map[string]*DiscoveredToolCall)
	var order []string

	err := forEachJSONLLine(updatesPath, 0, func(lineIdx int, lineBytes []byte) bool {
		var raw map[string]interface{}
		if err := json.Unmarshal(lineBytes, &raw); err != nil {
			return true
		}

		params, _ := raw["params"].(map[string]interface{})
		if params == nil {
			return true
		}

		updateMap, _ := params["update"].(map[string]interface{})
		if updateMap == nil {
			return true
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
			return true
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
			if toolName == "" {
				toolName, _ = updateMap["tool_name"].(string)
			}
			if toolName == "" {
				toolName, _ = updateMap["tool"].(string)
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
		} else if sessionUpdate == "tool_call_update" || sessionUpdate == "tool_result" {
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
			if out == "" {
				if oStr, ok := updateMap["output"].(string); ok {
					out = oStr
				} else if rStr, ok := updateMap["result"].(string); ok {
					out = rStr
				}
			}
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
		return true
	})
	if err != nil {
		return nil, err
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
	toolCallMap := make(map[string]*DiscoveredToolCall)
	var order []string

	err := forEachJSONLLine(chatPath, 0, func(lineIdx int, lineBytes []byte) bool {
		var raw map[string]interface{}
		if err := json.Unmarshal(lineBytes, &raw); err != nil {
			return true
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
		return true
	})
	if err != nil {
		return nil, err
	}

	result := make([]DiscoveredToolCall, 0, len(order))
	for _, id := range order {
		if tc, ok := toolCallMap[id]; ok {
			result = append(result, *tc)
		}
	}
	return result, nil
}
