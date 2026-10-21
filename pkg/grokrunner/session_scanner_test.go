package grokrunner

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDiscoverGrokSessions_ExtractionAndFiltering(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := "/tmp/test-workspace-scanner"
	encodedWs := url.PathEscape(testWs)
	testDir := filepath.Join(home, ".grok", "sessions", encodedWs)
	defer os.RemoveAll(testDir)

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 1. Session with summary.json
	sess1 := filepath.Join(testDir, "sess-01")
	os.MkdirAll(sess1, 0755)
	summaryData, _ := json.Marshal(map[string]interface{}{
		"session_summary": "Architecture Review & Refactor",
		"num_messages":    5,
	})
	os.WriteFile(filepath.Join(sess1, "summary.json"), summaryData, 0644)

	// 2. Session with chat_history.jsonl (<user_query>)
	sess2 := filepath.Join(testDir, "sess-02")
	os.MkdirAll(sess2, 0755)
	chatLine, _ := json.Marshal(map[string]interface{}{
		"role": "user",
		"content": "<user_info>\nOS: macos\n</user_info>\n<user_query>\nBuatkan skrip testing otomatis untuk Go\n</user_query>",
	})
	os.WriteFile(filepath.Join(sess2, "chat_history.jsonl"), append(chatLine, '\n'), 0644)

	// 3. Ghost session (0 messages, no prompts)
	sess3 := filepath.Join(testDir, "sess-ghost")
	os.MkdirAll(sess3, 0755)
	ghostSummary, _ := json.Marshal(map[string]interface{}{
		"session_summary": "",
		"num_messages":    0,
	})
	os.WriteFile(filepath.Join(sess3, "summary.json"), ghostSummary, 0644)

	// 4. Subagent internal session (session_kind = "subagent")
	sess4 := filepath.Join(testDir, "sess-subagent")
	os.MkdirAll(sess4, 0755)
	subagentSummary, _ := json.Marshal(map[string]interface{}{
		"session_summary": "Subagent Background Review Task",
		"session_kind":    "subagent",
		"num_messages":    12,
	})
	os.WriteFile(filepath.Join(sess4, "summary.json"), subagentSummary, 0644)

	results, err := DiscoverGrokSessions(testWs)
	if err != nil {
		t.Fatalf("DiscoverGrokSessions failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 valid sessions (ghost and subagent filtered out), got %d", len(results))
	}

	foundSummary := false
	foundChat := false

	for _, s := range results {
		if s.ID == "sess-01" && s.Title == "Architecture Review & Refactor" {
			foundSummary = true
		}
		if s.ID == "sess-02" && s.Title == "Buatkan skrip testing otomatis untuk Go" {
			foundChat = true
		}
	}

	if !foundSummary {
		t.Errorf("Failed to resolve title from summary.json for sess-01")
	}
	if !foundChat {
		t.Errorf("Failed to resolve title from user_query in chat_history.jsonl for sess-02")
	}
}

func TestLoadGrokSessionMessages_CompleteSequence(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := "/tmp/test-workspace-messages"
	encodedWs := url.PathEscape(testWs)
	testDir := filepath.Join(home, ".grok", "sessions", encodedWs, "session-test-seq")
	defer os.RemoveAll(filepath.Join(home, ".grok", "sessions", encodedWs))

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	lines := []string{
		`{"type":"system","content":"System prompt..."}`,
		`{"type":"user","content":"<user_info>OS: macos</user_info>\n<user_query>Hello, please write a test</user_query>"}`,
		`{"type":"assistant","content":"","tool_calls":[{"id":"tc-1","name":"read_file","arguments":"{\"file\":\"main.go\"}"}]}`,
		`{"type":"tool_result","tool_call_id":"tc-1","content":"package main\nfunc main() {}"}`,
		`{"type":"assistant","content":"Here is the test: func TestMain(t *testing.T) {}"}`,
		`{"type":"user","content":"<user_query>Can you add another test case?</user_query>"}`,
		`{"type":"assistant","content":"Sure, added TestSecond(t *testing.T) {}"}`,
	}

	f, err := os.Create(filepath.Join(testDir, "chat_history.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range lines {
		f.WriteString(l + "\n")
	}
	f.Close()

	msgs, err := LoadGrokSessionMessages(testWs, "session-test-seq")
	if err != nil {
		t.Fatalf("LoadGrokSessionMessages failed: %v", err)
	}

	if len(msgs) != 4 {
		t.Fatalf("Expected 4 messages (2 user, 2 assistant), got %d", len(msgs))
	}

	if msgs[0].Role != "user" || msgs[0].Content != "Hello, please write a test" {
		t.Errorf("Message 0 mismatch: %+v", msgs[0])
	}
	if msgs[1].Role != "assistant" || len(msgs[1].ToolCalls) != 1 || msgs[1].ToolCalls[0].Result != "package main\nfunc main() {}" {
		t.Errorf("Message 1 tool calls mismatch: %+v", msgs[1])
	}
	if msgs[2].Role != "user" || msgs[2].Content != "Can you add another test case?" {
		t.Errorf("Message 2 mismatch: %+v", msgs[2])
	}
	if msgs[3].Role != "assistant" || msgs[3].Content != "Sure, added TestSecond(t *testing.T) {}" {
		t.Errorf("Message 3 mismatch: %+v", msgs[3])
	}
}

func TestLoadGrokSessionMessages_InterleavedChronology(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := "/tmp/test-workspace-interleaved"
	encodedWs := url.PathEscape(testWs)
	testDir := filepath.Join(home, ".grok", "sessions", encodedWs, "session-test-interleave")
	defer os.RemoveAll(filepath.Join(home, ".grok", "sessions", encodedWs))

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	lines := []string{
		`{"type":"user","content":"<user_query>Investigate and fix the issue</user_query>"}`,
		`{"type":"assistant","content":"I will inspect the workspace files first.","tool_calls":[{"id":"tc-read","name":"read_file","arguments":"{\"target_file\":\"app.go\"}"}]}`,
		`{"type":"tool_result","tool_call_id":"tc-read","content":"func main() {}"}`,
		`{"type":"assistant","content":"I found the bug. Now updating the implementation.","tool_calls":[{"id":"tc-write","name":"write_file","arguments":"{\"path\":\"app.go\"}"}]}`,
		`{"type":"tool_result","tool_call_id":"tc-write","content":"Success"}`,
		`{"type":"assistant","content":"All changes completed and verified."}`,
	}

	f, err := os.Create(filepath.Join(testDir, "chat_history.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range lines {
		f.WriteString(l + "\n")
	}
	f.Close()

	msgs, err := LoadGrokSessionMessages(testWs, "session-test-interleave")
	if err != nil {
		t.Fatalf("LoadGrokSessionMessages failed: %v", err)
	}

	if len(msgs) != 2 {
		t.Fatalf("Expected 2 messages (1 user, 1 consolidated assistant), got %d", len(msgs))
	}

	asst := msgs[1]
	if asst.Role != "assistant" {
		t.Fatalf("Expected assistant message, got role %s", asst.Role)
	}

	if len(asst.Parts) != 5 {
		t.Fatalf("Expected 5 interleaved parts (text -> tool -> text -> tool -> text), got %d", len(asst.Parts))
	}

	if asst.Parts[0].Type != "text" || !strings.Contains(asst.Parts[0].Text, "I will inspect") {
		t.Errorf("Part 0 should be introductory text, got %+v", asst.Parts[0])
	}
	if asst.Parts[1].Type != "tool_calls" || len(asst.Parts[1].ToolCalls) != 1 || asst.Parts[1].ToolCalls[0].Tool != "read_file" {
		t.Errorf("Part 1 should be read_file tool call, got %+v", asst.Parts[1])
	}
	if asst.Parts[2].Type != "text" || !strings.Contains(asst.Parts[2].Text, "I found the bug") {
		t.Errorf("Part 2 should be middle text, got %+v", asst.Parts[2])
	}
	if asst.Parts[3].Type != "tool_calls" || len(asst.Parts[3].ToolCalls) != 1 || asst.Parts[3].ToolCalls[0].Tool != "write_file" {
		t.Errorf("Part 3 should be write_file tool call, got %+v", asst.Parts[3])
	}
	if asst.Parts[4].Type != "text" || !strings.Contains(asst.Parts[4].Text, "All changes completed") {
		t.Errorf("Part 4 should be final text, got %+v", asst.Parts[4])
	}
}

func TestEncodeGrokWorkspacePath(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    `C:\Users\Administrator\Desktop\project pribadi\aethergrok`,
			expected: `C%3A%5CUsers%5CAdministrator%5CDesktop%5Cproject%20pribadi%5Caethergrok`,
		},
		{
			input:    `/Users/fiko942/Desktop/affilia`,
			expected: `%2FUsers%2Ffiko942%2FDesktop%2Faffilia`,
		},
		{
			input:    `D:\UMM\Semester 7\Penjaminan Kualitas`,
			expected: `D%3A%5CUMM%5CSemester%207%5CPenjaminan%20Kualitas`,
		},
	}

	for _, c := range cases {
		// Note: On unix, filepath.Clean won't convert backslashes, but EncodeGrokWorkspacePath preserves characters
		var actual string
		clean := c.input
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
		actual = sb.String()

		if actual != c.expected {
			t.Errorf("Path encoding for %s:\nGot:  %s\nWant: %s", c.input, actual, c.expected)
		}
	}
}

func TestDiscoverGrokSessions_WindowsPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := `C:\TestWinWorkspace\project alpha`
	encodedWs := EncodeGrokWorkspacePath(testWs)
	testDir := filepath.Join(home, ".grok", "sessions", encodedWs)
	defer os.RemoveAll(testDir)

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	sess := filepath.Join(testDir, "win-sess-01")
	os.MkdirAll(sess, 0755)
	summaryData, _ := json.Marshal(map[string]interface{}{
		"session_summary": "Windows Session Testing",
		"num_messages":    3,
	})
	os.WriteFile(filepath.Join(sess, "summary.json"), summaryData, 0644)

	results, err := DiscoverGrokSessions(testWs)
	if err != nil {
		t.Fatalf("DiscoverGrokSessions failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 session for Windows path, got %d", len(results))
	}

	if results[0].Title != "Windows Session Testing" {
		t.Errorf("Expected title 'Windows Session Testing', got '%s'", results[0].Title)
	}
}

func TestGetSessionUsage_SignalsAndUsage(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := "/tmp/test-workspace-usage-stats"
	encodedWs := EncodeGrokWorkspacePath(testWs)
	testDir := filepath.Join(home, ".grok", "sessions", encodedWs, "01a0-real-uuid-1234")
	defer os.RemoveAll(filepath.Join(home, ".grok", "sessions", encodedWs))

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 1. Write signals.json
	signalsData, _ := json.Marshal(map[string]interface{}{
		"contextTokensUsed":   168180,
		"contextWindowTokens": 200000,
		"turnCount":           2,
		"primaryModelId":      "geminigacor",
	})
	os.WriteFile(filepath.Join(testDir, "signals.json"), signalsData, 0644)

	// 2. Write usage.json
	usageData, _ := json.Marshal(map[string]interface{}{
		"sessionId": "01a0-real-uuid-1234",
		"session": map[string]interface{}{
			"inputTokens":      8585469,
			"outputTokens":     43381,
			"cachedReadTokens": 2412653,
			"totalTokens":      8628850,
			"modelCalls":       85,
			"turnCount":        2,
			"primaryModelId":   "geminigacor",
		},
		"turns": []map[string]interface{}{
			{
				"turnNumber":       1,
				"inputTokens":      4000000,
				"outputTokens":     20000,
				"cachedReadTokens": 1000000,
			},
			{
				"turnNumber":       2,
				"inputTokens":      4585469,
				"outputTokens":     23381,
				"cachedReadTokens": 1412653,
			},
		},
	})
	os.WriteFile(filepath.Join(testDir, "usage.json"), usageData, 0644)

	// Test 1: Direct UUID lookup
	stats1, err := GetSessionUsage(testWs, "01a0-real-uuid-1234")
	if err != nil {
		t.Fatalf("GetSessionUsage failed: %v", err)
	}
	if stats1.UsedTokens != 168180 {
		t.Errorf("Expected 168180 UsedTokens, got %d", stats1.UsedTokens)
	}
	if stats1.MaxTokens != 200000 {
		t.Errorf("Expected 200000 MaxTokens, got %d", stats1.MaxTokens)
	}
	if stats1.TotalInput != 8585469 {
		t.Errorf("Expected 8585469 TotalInput, got %d", stats1.TotalInput)
	}

	// Test 2: Unmatched or temporary placeholder sess_ lookup must return empty stats and NOT borrow unrelated session
	stats2, err := GetSessionUsage(testWs, "sess_abc123456_random")
	if err != nil {
		t.Fatalf("GetSessionUsage with placeholder failed: %v", err)
	}
	if stats2.UsedTokens != 0 {
		t.Errorf("Expected placeholder to resolve to 0 UsedTokens (isolated), got %d", stats2.UsedTokens)
	}
	if stats2.SessionID != "sess_abc123456_random" {
		t.Errorf("Expected placeholder to retain its own ID 'sess_abc123456_random', got '%s'", stats2.SessionID)
	}
}

func TestResolveSessionFolder_Isolation(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := "/tmp/test-workspace-isolation-unique"
	encodedWs := EncodeGrokWorkspacePath(testWs)
	testDir := filepath.Join(home, ".grok", "sessions", encodedWs)
	defer os.RemoveAll(testDir)

	sess1Dir := filepath.Join(testDir, "11111111-1111-1111-1111-111111111111")
	sess2Dir := filepath.Join(testDir, "22222222-2222-2222-2222-222222222222")
	_ = os.MkdirAll(sess1Dir, 0755)
	_ = os.MkdirAll(sess2Dir, 0755)

	_ = os.WriteFile(filepath.Join(sess1Dir, "chat_history.jsonl"), []byte("{}\n"), 0644)
	_ = os.WriteFile(filepath.Join(sess2Dir, "chat_history.jsonl"), []byte("{}\n"), 0644)

	// Exact match
	_, id1 := ResolveSessionFolder(testDir, "11111111-1111-1111-1111-111111111111")
	if id1 != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("Expected exact match id1, got %s", id1)
	}

	// Unrelated session ID must NOT resolve to sess1 or sess2
	unrelated := "33333333-3333-3333-3333-333333333333"
	_, idUnrelated := ResolveSessionFolder(testDir, unrelated)
	if idUnrelated == "11111111-1111-1111-1111-111111111111" || idUnrelated == "22222222-2222-2222-2222-222222222222" {
		t.Errorf("ResolveSessionFolder leaked an unrelated session folder: got %s", idUnrelated)
	}

	// Frontend placeholder must NOT resolve to sess1 or sess2
	placeholder := "sess_abc123"
	_, idPlaceholder := ResolveSessionFolder(testDir, placeholder)
	if idPlaceholder == "11111111-1111-1111-1111-111111111111" || idPlaceholder == "22222222-2222-2222-2222-222222222222" {
		t.Errorf("ResolveSessionFolder leaked an existing session to placeholder: got %s", idPlaceholder)
	}
}

func TestSessionFolderExists_StrictWorkspaceIsolation(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	ws1 := filepath.Join(t.TempDir(), "ws1")
	ws2 := filepath.Join(t.TempDir(), "ws2")
	_ = os.MkdirAll(ws1, 0755)
	_ = os.MkdirAll(ws2, 0755)

	ws1Dir := ResolveWorkspaceSessionsDir(sessionsDir, ws1)
	ws2Dir := ResolveWorkspaceSessionsDir(sessionsDir, ws2)
	_ = os.MkdirAll(ws1Dir, 0755)
	_ = os.MkdirAll(ws2Dir, 0755)
	defer os.RemoveAll(ws1Dir)
	defer os.RemoveAll(ws2Dir)

	sessID := "99999999-1111-2222-3333-444444444444"
	sessFolder1 := filepath.Join(ws1Dir, sessID)
	_ = os.MkdirAll(sessFolder1, 0755)
	_ = os.WriteFile(filepath.Join(sessFolder1, "chat_history.jsonl"), []byte("{}\n"), 0644)

	// ws1 should find sessID
	exists1, _, _ := SessionFolderExists(sessionsDir, ws1Dir, sessID)
	if !exists1 {
		t.Fatalf("Expected session to exist in ws1")
	}

	// ws2 MUST NOT find sessID under any circumstances (no cross-workspace fallback)
	exists2, _, _ := SessionFolderExists(sessionsDir, ws2Dir, sessID)
	if exists2 {
		t.Fatalf("Session from ws1 MUST NOT be found when querying ws2; cross-workspace fallback must be disabled")
	}
}

func TestResolveWorkspaceSessionsDir_SymlinkCanonicalization(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	realDir := t.TempDir()
	symlinkBase := t.TempDir()
	symlinkPath := filepath.Join(symlinkBase, "linked_ws")
	if err := os.Symlink(realDir, symlinkPath); err != nil {
		t.Skipf("Symlink creation not supported: %v", err)
	}

	// Canonical path as stored by Grok
	canonicalTargetDir := filepath.Join(sessionsDir, EncodeGrokWorkspacePath(realDir))
	_ = os.MkdirAll(canonicalTargetDir, 0755)
	defer os.RemoveAll(canonicalTargetDir)

	resolved := ResolveWorkspaceSessionsDir(sessionsDir, symlinkPath)
	if resolved != canonicalTargetDir {
		t.Errorf("Expected ResolveWorkspaceSessionsDir to resolve canonical symlink directory %s, got %s", canonicalTargetDir, resolved)
	}
}


func TestResolveWorkspaceSessionsDir_CrossPlatformDecoded(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := "/tmp/test-workspace-crossplatform"
	encodedWs := EncodeGrokWorkspacePath(testWs)
	testDir := filepath.Join(home, ".grok", "sessions", encodedWs)
	defer os.RemoveAll(testDir)

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	sessionsDir := filepath.Join(home, ".grok", "sessions")
	resolved := ResolveWorkspaceSessionsDir(sessionsDir, testWs)
	if resolved != testDir {
		t.Errorf("Expected %s, got %s", testDir, resolved)
	}
}

func TestParseSubagentUpdatesJsonl_RealLifecycle(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "updates.jsonl")
	content := strings.Join([]string{
		`{"timestamp":1792379656,"method":"session/update","params":{"sessionId":"sub-1","update":{"sessionUpdate":"tool_call","toolCallId":"call_1","title":"read_file","rawInput":{"target_file":"README.md"},"_meta":{"x.ai/tool":{"name":"read_file"}}}}}`,
		`{"timestamp":1792379657,"method":"session/update","params":{"sessionId":"sub-1","update":{"sessionUpdate":"tool_call_update","toolCallId":"call_1","status":"completed","content":[{"type":"content","content":{"type":"text","text":"# AsistenQ"}}]}}}`,
		`{"timestamp":1792379658,"method":"session/update","params":{"sessionId":"sub-1","update":{"sessionUpdate":"tool_call","toolCallId":"call_2","title":"run_terminal_command","rawInput":{"command":"ls -la"}}}}`,
		`{"timestamp":1792379659,"method":"session/update","params":{"sessionId":"sub-1","update":{"sessionUpdate":"tool_call_update","toolCallId":"call_2","status":"completed","rawOutput":{"output_for_prompt":"file1\nfile2"}}}}`,
	}, "\n")

	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	calls, err := ParseSubagentUpdatesJsonl(tempFile)
	if err != nil {
		t.Fatalf("ParseSubagentUpdatesJsonl failed: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("Expected 2 tool calls, got %d", len(calls))
	}

	if calls[0].ID != "call_1" || calls[0].Tool != "read_file" || calls[0].Status != "completed" || calls[0].Result != "# AsistenQ" {
		t.Errorf("Unexpected calls[0]: %+v", calls[0])
	}
	if calls[1].ID != "call_2" || calls[1].Tool != "run_terminal_command" || calls[1].Status != "completed" || calls[1].Result != "file1\nfile2" {
		t.Errorf("Unexpected calls[1]: %+v", calls[1])
	}
}

func TestParseSubagentChatHistoryJsonl_Fallback(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "chat_history.jsonl")
	content := strings.Join([]string{
		`{"type":"system","content":"System prompt"}`,
		`{"type":"user","content":[{"type":"text","text":"Analyze files"}]}`,
		`{"type":"assistant","content":"","tool_calls":[{"id":"call_c1","name":"read_file","arguments":"{\"target_file\":\"test.txt\"}"}]}`,
		`{"type":"tool_result","tool_call_id":"call_c1","content":"file contents"}`,
	}, "\n")

	if err := os.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	calls, err := ParseSubagentChatHistoryJsonl(tempFile)
	if err != nil {
		t.Fatalf("ParseSubagentChatHistoryJsonl failed: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("Expected 1 tool call, got %d", len(calls))
	}

	if calls[0].ID != "call_c1" || calls[0].Tool != "read_file" || calls[0].Result != "file contents" || calls[0].Status != "completed" {
		t.Errorf("Unexpected call: %+v", calls[0])
	}
}

func TestLoadSubagentExecutionDetails_RealSession(t *testing.T) {
	subagentID := "01a15062-c047-7d91-a7e1-10e862e35c64"
	wsPath := "/Users/fiko942/Desktop/Rekayasa Interaksi"
	details, err := LoadSubagentExecutionDetails(wsPath, subagentID)
	if err != nil {
		t.Fatalf("LoadSubagentExecutionDetails returned error: %v", err)
	}
	if details == nil || len(details.Calls) == 0 {
		t.Skip("Real subagent directory not present on this machine, skipping live check")
	}
	if details.Status != "completed" {
		t.Errorf("Expected status='completed', got %s", details.Status)
	}
	if len(details.Calls) != 13 {
		t.Errorf("Expected 13 tool calls, got %d", len(details.Calls))
	}
	if details.Output == "" {
		t.Errorf("Expected non-empty output for completed subagent")
	}
	if details.DurationMs <= 0 {
		t.Errorf("Expected positive durationMs, got %d", details.DurationMs)
	}
}

func TestLoadSubagentToolCalls_RealSession(t *testing.T) {
	// Query the verified real subagent session
	subagentID := "01a15226-fcb4-7973-a534-2dc31fbb1a8b"
	wsPath := "/Users/fiko942/Desktop/asistenq-youtube"
	calls, err := LoadSubagentToolCalls(wsPath, subagentID)
	if err != nil {
		t.Fatalf("LoadSubagentToolCalls returned error: %v", err)
	}
	if len(calls) == 0 {
		t.Skip("Real subagent directory not present on this machine, skipping live check")
	}
	if len(calls) != 34 {
		t.Errorf("Expected 34 tool calls for verified subagent, got %d", len(calls))
	}
	// Verify first call is read_file
	if calls[0].Tool != "read_file" {
		t.Errorf("Expected first tool to be read_file, got %s", calls[0].Tool)
	}
	// Verify last call is completed
	if calls[len(calls)-1].Status != "completed" {
		t.Errorf("Expected last tool to be completed, got %s", calls[len(calls)-1].Status)
	}
}

func TestLoadGrokSessionMessages_RealSession(t *testing.T) {
	wsPath := "/Users/fiko942/Desktop/Rekayasa Interaksi"
	sessionID := "01a0f8ba-2c45-7572-b164-abae6d172e3f"
	msgs, err := LoadGrokSessionMessages(wsPath, sessionID)
	if err != nil {
		t.Fatalf("LoadGrokSessionMessages returned error: %v", err)
	}
	if len(msgs) == 0 {
		t.Skip("Session not found")
	}

	foundSubagent := false
	for _, m := range msgs {
		for _, tc := range m.ToolCalls {
			if tc.Tool == "spawn_subagent" {
				foundSubagent = true
				if tc.ChildSessionID == "" {
					t.Errorf("Expected ChildSessionID to be populated for spawn_subagent, got empty")
				}
				if len(tc.ChildToolCalls) == 0 {
					t.Errorf("Expected ChildToolCalls to be populated from disk for subagent %s, got 0", tc.ChildSessionID)
				}
				if tc.Status != "completed" {
					t.Errorf("Expected subagent Status to be 'completed', got %s", tc.Status)
				}
				if tc.Result == "" || strings.Contains(strings.ToLower(tc.Result), "subagent started in background") {
					t.Errorf("Expected subagent Result to be replaced with final output, got: %s", tc.Result)
				}
			}
		}
	}
	if !foundSubagent {
		t.Errorf("Expected to find spawn_subagent tool call in messages")
	}
}

func TestSessionFolderExists_Validation(t *testing.T) {
	tempBase, err := os.MkdirTemp("", "grok_sessions_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempBase)

	ws1Dir := filepath.Join(tempBase, "workspace1")
	ws2Dir := filepath.Join(tempBase, "workspace2")
	_ = os.MkdirAll(ws1Dir, 0755)
	_ = os.MkdirAll(ws2Dir, 0755)

	// 1. Non-existent session
	exists, _, _ := SessionFolderExists(tempBase, ws1Dir, "non-existent-uuid")
	if exists {
		t.Errorf("Expected exists=false for non-existent session, got true")
	}

	// 2. Empty directory (no session records)
	emptySessionDir := filepath.Join(ws1Dir, "empty-session-uuid")
	_ = os.MkdirAll(emptySessionDir, 0755)
	existsEmpty, _, _ := SessionFolderExists(tempBase, ws1Dir, "empty-session-uuid")
	if existsEmpty {
		t.Errorf("Expected exists=false for empty session directory, got true")
	}

	// 3. Initialized session in ws1
	validSessionDir := filepath.Join(ws1Dir, "valid-session-uuid")
	_ = os.MkdirAll(validSessionDir, 0755)
	_ = os.WriteFile(filepath.Join(validSessionDir, "chat_history.jsonl"), []byte("{}\n"), 0644)
	existsValid, fPath, resolvedID := SessionFolderExists(tempBase, ws1Dir, "valid-session-uuid")
	if !existsValid {
		t.Errorf("Expected exists=true for initialized session, got false")
	}
	if resolvedID != "valid-session-uuid" {
		t.Errorf("Expected resolvedID 'valid-session-uuid', got '%s'", resolvedID)
	}
	if fPath != validSessionDir {
		t.Errorf("Expected fPath '%s', got '%s'", validSessionDir, fPath)
	}

	// 4. Session exists in ws2 (MUST NOT leak into ws1 lookup)
	ws2SessionDir := filepath.Join(ws2Dir, "ws2-session-uuid")
	_ = os.MkdirAll(ws2SessionDir, 0755)
	_ = os.WriteFile(filepath.Join(ws2SessionDir, "summary.json"), []byte("{}\n"), 0644)
	existsFallback, _, _ := SessionFolderExists(tempBase, ws1Dir, "ws2-session-uuid")
	if existsFallback {
		t.Errorf("Expected exists=false when session belongs to ws2, but got true (cross-workspace leak)")
	}

	// 5. Querying ws2 directly finds ws2 session
	existsDirect, fPath2, resolvedID2 := SessionFolderExists(tempBase, ws2Dir, "ws2-session-uuid")
	if !existsDirect {
		t.Errorf("Expected exists=true when querying ws2 directly, got false")
	}
	if resolvedID2 != "ws2-session-uuid" {
		t.Errorf("Expected resolvedID 'ws2-session-uuid', got '%s'", resolvedID2)
	}
	if fPath2 != ws2SessionDir {
		t.Errorf("Expected fPath '%s', got '%s'", ws2SessionDir, fPath2)
	}
}

func TestLoadGrokSessionMessages_SubagentHydration(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := "/tmp/test-workspace-subagent-hydrate"
	encodedWs := EncodeGrokWorkspacePath(testWs)
	testDir := filepath.Join(home, ".grok", "sessions", encodedWs)
	defer os.RemoveAll(testDir)

	parentSessDir := filepath.Join(testDir, "parent-sess-sub1")
	subID := "sub-uuid-4321-abcd"
	subInParentDir := filepath.Join(parentSessDir, "subagents", subID)
	subTopDir := filepath.Join(testDir, subID)

	if err := os.MkdirAll(subInParentDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subTopDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 1. Write parent chat_history.jsonl
	parentLines := []string{
		`{"type":"user","content":"<user_query>Run a subagent to inspect repository</user_query>"}`,
		`{"type":"assistant","content":"Launching subagent...","tool_calls":[{"id":"tc-sub","name":"spawn_subagent","arguments":"{\"description\":\"Inspect repo\",\"prompt\":\"Inspect files\"}"}]}`,
		`{"type":"tool_result","tool_call_id":"tc-sub","content":"Subagent started in background.\nsubagent_id: ` + subID + `\ndescription: Inspect repo"}`,
	}
	_ = os.WriteFile(filepath.Join(parentSessDir, "chat_history.jsonl"), []byte(strings.Join(parentLines, "\n")+"\n"), 0644)

	// 2. Write subagent meta.json and output.json in parent subagents/
	metaContent := `{"status":"completed","duration_ms":7500,"tool_calls":2,"started_at":"2026-10-19T10:00:00Z","completed_at":"2026-10-19T10:00:07.500Z","description":"Inspect repo","prompt":"Inspect files","subagent_type":"explore"}`
	_ = os.WriteFile(filepath.Join(subInParentDir, "meta.json"), []byte(metaContent), 0644)
	outContent := `{"output":"All repository files have been analyzed successfully."}`
	_ = os.WriteFile(filepath.Join(subInParentDir, "output.json"), []byte(outContent), 0644)

	// 3. Write child updates.jsonl in top-level subagent dir
	subUpdates := []string{
		`{"params":{"update":{"sessionUpdate":"tool_call","tool_call_id":"child_1","tool_name":"read_file","args":"{\"path\":\"README.md\"}"}}}`,
		`{"params":{"update":{"sessionUpdate":"tool_result","tool_call_id":"child_1","content":"# Project Readme"}}}`,
		`{"params":{"update":{"sessionUpdate":"turn_completed","stop_reason":"end_turn"}}}`,
	}
	_ = os.WriteFile(filepath.Join(subTopDir, "updates.jsonl"), []byte(strings.Join(subUpdates, "\n")+"\n"), 0644)

	msgs, err := LoadGrokSessionMessages(testWs, "parent-sess-sub1")
	if err != nil {
		t.Fatalf("LoadGrokSessionMessages failed: %v", err)
	}

	if len(msgs) != 2 {
		t.Fatalf("Expected 2 messages (user, assistant), got %d", len(msgs))
	}

	asst := msgs[1]
	if len(asst.ToolCalls) != 1 {
		t.Fatalf("Expected 1 tool call in assistant message, got %d", len(asst.ToolCalls))
	}

	tc := asst.ToolCalls[0]
	if tc.ChildSessionID != subID {
		t.Errorf("Expected ChildSessionID=%s, got %s", subID, tc.ChildSessionID)
	}
	if tc.Status != "completed" {
		t.Errorf("Expected Status='completed', got %s", tc.Status)
	}
	if tc.Result != "All repository files have been analyzed successfully." {
		t.Errorf("Expected Result='All repository files have been analyzed successfully.', got '%s'", tc.Result)
	}
	if len(tc.ChildToolCalls) != 1 {
		t.Fatalf("Expected 1 child tool call, got %d", len(tc.ChildToolCalls))
	}
	if tc.ChildToolCalls[0].Tool != "read_file" || tc.ChildToolCalls[0].Result != "# Project Readme" {
		t.Errorf("Unexpected child tool call: %+v", tc.ChildToolCalls[0])
	}

	if tc.StartTime <= 0 {
		t.Errorf("Expected positive StartTime, got %d", tc.StartTime)
	}
	if tc.EndTime <= 0 {
		t.Errorf("Expected positive EndTime, got %d", tc.EndTime)
	}
	if tc.Params["description"] != "Inspect repo" {
		t.Errorf("Expected description='Inspect repo', got %v", tc.Params["description"])
	}
	if tc.Params["prompt"] != "Inspect files" {
		t.Errorf("Expected prompt='Inspect files', got %v", tc.Params["prompt"])
	}
	if tc.Params["subagent_type"] != "explore" {
		t.Errorf("Expected subagent_type='explore', got %v", tc.Params["subagent_type"])
	}

	// Verify synchronization into asst.Parts
	if len(asst.Parts) == 0 {
		t.Fatalf("Expected asst.Parts to be populated")
	}
	var foundPartTool bool
	for _, p := range asst.Parts {
		if p.Type == "tool_calls" {
			for _, ptc := range p.ToolCalls {
				if ptc.ID == "tc-sub" {
					foundPartTool = true
					if len(ptc.ChildToolCalls) != 1 {
						t.Errorf("Part tool call ChildToolCalls count mismatch: got %d", len(ptc.ChildToolCalls))
					}
					if ptc.Result != "All repository files have been analyzed successfully." {
						t.Errorf("Part tool call Result mismatch: '%s'", ptc.Result)
					}
					if ptc.Status != "completed" {
						t.Errorf("Part tool call Status mismatch: '%s'", ptc.Status)
					}
				}
			}
		}
	}
	if !foundPartTool {
		t.Errorf("Did not find tc-sub in asst.Parts")
	}
}

func TestLoadGrokSessionMessages_PromptTagVariations(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := "/tmp/test-workspace-prompt-variants"
	encodedWs := EncodeGrokWorkspacePath(testWs)
	testDir := filepath.Join(home, ".grok", "sessions", encodedWs, "sess-variants")
	defer os.RemoveAll(filepath.Join(home, ".grok", "sessions", encodedWs))

	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	lines := []string{
		// Initial empty assistant message or reasoning should be skipped and not create orphan assistant message
		`{"type":"assistant","content":"","tool_calls":[]}`,
		`{"type":"reasoning","content":"Thinking about user intent..."}`,
		// Turn 1: user prompt with <user_info> prefix and prompt after tags
		`{"type":"user","content":"<user_info>\nOS: windows\nShell: powershell\n</user_info>\nInvestigate the database connection issue"}`,
		`{"type":"assistant","content":"I will inspect the database config.","tool_calls":[]}`,
		// Turn 2: plain text prompt without XML tags
		`{"type":"user","content":"Please apply the proposed fix to db.go"}`,
		`{"type":"assistant","content":"The fix has been applied.","tool_calls":[]}`,
	}

	f, err := os.Create(filepath.Join(testDir, "chat_history.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range lines {
		f.WriteString(l + "\n")
	}
	f.Close()

	msgs, err := LoadGrokSessionMessages(testWs, "sess-variants")
	if err != nil {
		t.Fatalf("LoadGrokSessionMessages failed: %v", err)
	}

	if len(msgs) != 4 {
		t.Fatalf("Expected 4 messages (2 user, 2 assistant) properly sequenced without orphan initial messages, got %d", len(msgs))
	}

	if msgs[0].Role != "user" || msgs[0].Content != "Investigate the database connection issue" {
		t.Errorf("Message 0 mismatch: %+v", msgs[0])
	}
	if msgs[1].Role != "assistant" || msgs[1].Content != "I will inspect the database config." {
		t.Errorf("Message 1 mismatch: %+v", msgs[1])
	}
	if msgs[2].Role != "user" || msgs[2].Content != "Please apply the proposed fix to db.go" {
		t.Errorf("Message 2 mismatch: %+v", msgs[2])
	}
	if msgs[3].Role != "assistant" || msgs[3].Content != "The fix has been applied." {
		t.Errorf("Message 3 mismatch: %+v", msgs[3])
	}
}

func TestForEachJSONLLine_StreamingAndEarlyExit(t *testing.T) {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "stream_test.jsonl")

	var contentBuilder strings.Builder
	for i := 0; i < 20; i++ {
		contentBuilder.WriteString(fmt.Sprintf(`{"index":%d,"data":"sample line %d"}`+"\n", i, i))
	}
	if err := os.WriteFile(filePath, []byte(contentBuilder.String()), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 1. Early exit when returning false at index 3
	var linesRead []int
	err := forEachJSONLLine(filePath, 0, func(idx int, lineBytes []byte) bool {
		linesRead = append(linesRead, idx)
		return idx < 3 // stop when idx reaches 3
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(linesRead) != 4 {
		t.Errorf("expected 4 lines read before early exit, got %d: %v", len(linesRead), linesRead)
	}

	// 2. Bound by maxLines
	var maxLinesRead []int
	err = forEachJSONLLine(filePath, 5, func(idx int, lineBytes []byte) bool {
		maxLinesRead = append(maxLinesRead, idx)
		return true
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(maxLinesRead) != 5 {
		t.Errorf("expected exactly 5 lines read with maxLines=5, got %d", len(maxLinesRead))
	}
}

func TestLoadGrokSessionMessages_ModTimeCaching(t *testing.T) {
	ClearAllSessionCache()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}

	testWs := "/tmp/test-workspace-caching"
	encodedWs := EncodeGrokWorkspacePath(testWs)
	sessDir := filepath.Join(home, ".grok", "sessions", encodedWs, "sess-cache-test")
	defer os.RemoveAll(filepath.Join(home, ".grok", "sessions", encodedWs))

	if err := os.MkdirAll(sessDir, 0755); err != nil {
		t.Fatal(err)
	}

	chatFile := filepath.Join(sessDir, "chat_history.jsonl")
	initialContent := `{"type":"user","content":"First prompt"}
{"type":"assistant","content":"First answer"}` + "\n"
	if err := os.WriteFile(chatFile, []byte(initialContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 1. First read parses from disk and populates cache
	msgs1, err := LoadGrokSessionMessages(testWs, "sess-cache-test")
	if err != nil {
		t.Fatalf("LoadGrokSessionMessages failed: %v", err)
	}
	if len(msgs1) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs1))
	}

	// 2. Second read should hit in-memory cache directly
	msgs2, err := LoadGrokSessionMessages(testWs, "sess-cache-test")
	if err != nil {
		t.Fatalf("LoadGrokSessionMessages failed on cache hit: %v", err)
	}
	if len(msgs2) != 2 || msgs2[0].Content != "First prompt" {
		t.Errorf("cached messages mismatch: %+v", msgs2)
	}

	// 3. Update file with new turn (change ModTime and size)
	time.Sleep(20 * time.Millisecond) // ensure timestamp ticks
	updatedContent := initialContent + `{"type":"user","content":"Second prompt"}
{"type":"assistant","content":"Second answer"}` + "\n"
	if err := os.WriteFile(chatFile, []byte(updatedContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 4. Third read detects ModTime/size change and re-parses
	msgs3, err := LoadGrokSessionMessages(testWs, "sess-cache-test")
	if err != nil {
		t.Fatalf("LoadGrokSessionMessages failed on updated file: %v", err)
	}
	if len(msgs3) != 4 {
		t.Fatalf("expected 4 messages after file update, got %d", len(msgs3))
	}

	// 5. Invalidate cache
	InvalidateSessionCache(sessDir)
	sessionCacheMu.RLock()
	cached := sessionCache[chatFile]
	sessionCacheMu.RUnlock()
	if cached != nil {
		t.Errorf("expected cache to be invalidated for %s, but found entry", chatFile)
	}
}



