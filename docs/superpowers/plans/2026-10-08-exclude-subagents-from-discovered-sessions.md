# Implementation Plan: Exclude Subagent Internal Sessions from Discovered Session List

## 1. Problem Diagnosis & Root Cause
In AetherGrok Desktop, when viewing workspace sessions in the left sidebar or Session dropdown (e.g. `smm.ziqva`), the user sees an overwhelming number of sessions (e.g. 57 sessions).

### Investigation Findings:
- Of the 57 session folders in `~/.grok/sessions/C%3A%5C...%5Csmm.ziqva`:
  - **47 sessions** are internal **subagent runs** spawned by parent sessions during plan execution, multi-agent dispatch, or tool reviews (`session_kind: "subagent"` in `summary.json`).
  - **10 sessions** are genuine user root sessions (`session_kind: "headless"` or standard user sessions).
- Grok CLI explicitly tags subagents in `summary.json`:
  ```json
  {
    "id": "01a115ea-7c98-7163-b50e-f4a7c4f8bfa9",
    "session_kind": "subagent",
    "agent_name": "general-purpose",
    "generated_title": "Add USD Exchange Rate Fields to Prisma Schema"
  }
  ```
- In `pkg/grokrunner/session_scanner.go` -> `DiscoverGrokSessions(workspacePath string)`:
  - The scanner loops over every directory in the workspace sessions folder.
  - It reads `summary.json`, but currently does NOT inspect `summaryObj.SessionKind`.
  - It unconditionally treats subagent folders as top-level user sessions and returns them in `results`.
  - The frontend `sessionStore.syncDiscoveredGrokSessions` imports all of them into the left session list.

## 2. Proposed Solution

### 1. Filter out `subagent` sessions in `DiscoverGrokSessions`:
In `pkg/grokrunner/session_scanner.go`:
- Extend the `summaryObj` struct in `DiscoverGrokSessions`:
  ```go
  var summaryObj struct {
      SessionSummary  string `json:"session_summary"`
      GeneratedTitle  string `json:"generated_title"`
      SessionKind     string `json:"session_kind"`
      NumMessages     int    `json:"num_messages"`
      NumChatMessages int    `json:"num_chat_messages"`
      CreatedAt       string `json:"created_at"`
      UpdatedAt       string `json:"updated_at"`
  }
  ```
- If `summaryObj.SessionKind == "subagent"`, skip this folder entirely (`continue`).
- Also check if `summaryObj.SessionKind == "subagent_review"` or similar future subagent variants, or if `summaryObj.SessionKind` starts with `"subagent"`.

### 2. Add Unit Test Coverage:
In `pkg/grokrunner/session_scanner_test.go`:
- Add a test verifying that `DiscoverGrokSessions` properly filters out session folders where `summary.json` has `session_kind: "subagent"`, while keeping user sessions (`headless`, `fork`, or standard sessions without `subagent` tag).

### 3. Verification:
- Run Go backend tests: `go test -v ./pkg/grokrunner/...`.
- Run frontend Vitest tests and production build.
- Commit and push to GitHub.
