package grokrunner

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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

	// In Grok, workspace directory path is URL path encoded, e.g. %2FUsers%2Ffiko942%2FDesktop%2Faffilia
	encodedPath := url.PathEscape(workspacePath)
	targetDir := filepath.Join(sessionsDir, encodedPath)

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

// DeleteGrokSessionDirectory removes a session directory from ~/.grok/sessions/
func DeleteGrokSessionDirectory(workspacePath, sessionID string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	encodedPath := url.PathEscape(workspacePath)
	targetDir := filepath.Join(home, ".grok", "sessions", encodedPath, sessionID)

	if _, err := os.Stat(targetDir); err == nil {
		return os.RemoveAll(targetDir)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
