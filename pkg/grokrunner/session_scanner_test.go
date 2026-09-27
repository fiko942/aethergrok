package grokrunner

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"
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

	results, err := DiscoverGrokSessions(testWs)
	if err != nil {
		t.Fatalf("DiscoverGrokSessions failed: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("Expected 2 valid sessions (ghost filtered out), got %d", len(results))
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
