# Subagent Chronology, Activity Persistence, Plan Filtering & Fail-Closed Lifecycle Guard Specification

**Author:** AetherGrok Engineering
**Date:** 2026-10-19
**Version:** v1.1.17
**Status:** Shipped in v1.1.17

---

## 1. Executive Summary

AetherGrok v1.1.17 addresses critical usability, safety, and persistence challenges encountered during long-running autonomous workflows:
1. **Fail-Closed Lifecycle Guard & Input Shield:** Eliminates accidental application termination while agent tasks run. A reactive bridge transmits active session task counts to the Go Wails backend (`SetFrontendActiveTasksCount`), while `beforeClose` halts unexpected quits and presents a confirmation dialog. An input shield suppresses conflicting shortcuts (`Ctrl+W`, `ShiftRight`, `ShiftLeft`) while editing.
2. **Chronological Message Interleaving:** Message content blocks and tool actions render in true temporal sequence, avoiding visual disorientation caused by premature grouping of subsequent tool calls above preceding narrative text.
3. **Subagent History & Detail Persistence:** Deep-scanner indexing in `pkg/grokrunner/session_scanner.go` reconstructs subagent transcripts, tool calls, and outputs directly from disk, preserving full subagent state across application restarts and session reloads.
4. **Subagent UX Polish & Todo Formatting:** Live stopwatch timer displays real-time execution duration, subagent activity lists auto-scroll down, `todo_write` arguments format into human-readable task status items, and plan trackers default to primary "main" goals.

---

## 2. Technical Architecture & Data Flow

```
┌────────────────────────────────────────────────────────┐
│                   Frontend (Svelte 5)                  │
│                                                        │
│  [Prompt Execution / Tool Calls / Subagents]          │
│                       │                                │
│                       ▼                                │
│  $effect: Compute total active running tasks           │
│                       │                                │
│                       ▼                                │
│  App.SetFrontendActiveTasksCount(count)                │
└───────────────────────┬────────────────────────────────┘
                        │ Wails IPC Bridge
                        ▼
┌────────────────────────────────────────────────────────┐
│                   Backend (Go / Wails)                 │
│                                                        │
│  App.beforeClose(ctx context.Context) (prevent bool)   │
│  ├─ if count > 0 || hasRunningProcesses:               │
│  │    ├─ Prompt: "Tasks in progress. Quit anyway?"    │
│  │    └─ Fail-closed fallback if ctx is nil            │
│  └─ else: allow clean quit                             │
└────────────────────────────────────────────────────────┘
```

---

## 3. Component Deep Dive

### 3.1. Accidental Quit Prevention & Lifecycle Guard
- **Frontend Input Shield:** In `frontend/src/App.svelte`, window-level keydown listeners inspect `isEditing` (active focus in `<input>` or `<textarea>`). Conflicting shortcuts that trigger window closures or background snapshots (`Ctrl+W`, `ShiftRight`, `ShiftLeft`) are intercepted and prevented.
- **Wails `beforeClose` Integration:** In `app.go`, the lifecycle hook evaluates `a.frontendActiveTasksCount`. If tasks are active, a confirmation dialog is presented via `wailsRuntime.MessageDialog`. If the context is headless or the dialog returns an error, it fails closed (`return true`) to preserve running processes.

### 3.2. Subagent Activity Persistence & Scanning
- **Deep Scanner:** `pkg/grokrunner/session_scanner.go` traverses subdirectories within each session's storage root to detect spawned subagent logs and transcripts.
- **Hydration:** Rebuilds `SubagentInfo` structures containing `prompt`, `instructions`, `toolCalls`, `output`, and `status`, restoring full modal inspection capability upon application restart.

### 3.3. Chronological Interleaved Message Rendering
- **Sequential Parts:** In `MessageItem.svelte`, turns containing multiple text blocks and tool calls preserve original stream index ordering.
- **Natural Reading Flow:** Narrative explanations preceding tool executions appear above tool cards, matching standard conversational expectations.

### 3.4. Subagent Timer & Todo Formatting
- **Stopwatch:** Tracks `Date.now() - startTime` reactively in `ToolCallCard.svelte` while `status === 'running'`, stopping precisely when the subagent completes.
- **Auto-Scroll:** Attaches auto-scroll listeners to the subagent activity log container in `ToolCallCard.svelte` and `SubagentDetailModal.svelte`.
- **Todo Parser:** In `frontend/src/lib/utils/toolUtils.ts`, detects `todo_write` arguments and converts `{ merge: true, todos: [...] }` payloads into styled badges representing `completed`, `in_progress`, and `pending` tasks.
