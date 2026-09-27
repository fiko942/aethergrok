# Enterprise Logging Architecture & Standards Specification

> **Version:** 1.0.0  
> **Status:** Active Standard  
> **Scope:** Entire AetherGrok Desktop Studio (Frontend Svelte 5, Go Backend IPC, State Stores, Components)

---

## 1. Core Principles

1. **Zero Resource Bloat (RAM, CPU, Storage)**:
   - In-memory logs are stored in a fixed circular ring buffer (`maxEntries = 1000`). Old entries are shifted off to prevent memory leaks (< 500 KB RAM footprint).
   - Logging calls (`logger.info`, `logger.debug`, `logger.warn`, `logger.error`) are non-blocking and fire-and-forget.
   - Heavy data objects must not be serialized synchronously unless an error or debug inspection is active.

2. **Structured Domains (Categories)**:
   All log entries must declare one of the standard categories:
   - `UI`: User clicks, modal transitions, navigation, tab operations, selection changes.
   - `SESSION`: Session hydration, lifecycle, message additions, turns, compaction.
   - `TERMINAL`: PTY creation, resize events, close, rename, tab switching.
   - `VOICE`: Audio input devices, recording lifecycle, dictation transcription.
   - `BACKEND`: Go Wails IPC calls, file discovery, system settings queries.
   - `SETTINGS`: Configuration modifications, theme changes, shortcut recordings.
   - `SYSTEM`: Unhandled exceptions, promise rejections, runtime initialization.

3. **Log Levels**:
   - `DEBUG`: Verbose diagnostics (e.g. detailed payload dumps, token measurements).
   - `INFO`: Normal state transitions and user operations (e.g. tab opened, session created).
   - `WARN`: Recoverable anomalies or fallback conditions (e.g. missing workspace path fallback).
   - `ERROR`: Unhandled exceptions, failed backend commands, corrupted data states.

---

## 2. Mandatory Rules for Adding New Features

When creating or modifying any new feature in AetherGrok, engineers and AI subagents **must** adhere to the following checklist:

### A. Store & Service Layer
- Import `logger` from `$lib/stores/logger.svelte`.
- Log the start and success of significant asynchronous operations (`logger.info`).
- Wrap IPC / I/O operations in `try / catch` blocks and record errors with `logger.error('CATEGORY', 'Contextual error message', err)`.

```typescript
import { logger } from '$lib/stores/logger.svelte';

async function performFeatureAction(id: string) {
  logger.info('UI', `Starting feature action for ${id}`);
  try {
    const result = await window.go.main.App.SomeBackendCall(id);
    logger.debug('BACKEND', `Backend call succeeded for ${id}`, result);
  } catch (err) {
    logger.error('BACKEND', `Failed to execute backend call for ${id}`, err);
    throw err;
  }
}
```

### B. UI Component Layer
- Log key user interactions that change global state (modal open/close, drag operations, delete confirmations).
- Avoid logging high-frequency mousemove or scroll events unless aggregated or debounced.

### C. Error Boundary Integration
- Global uncaught errors and unhandled promise rejections are automatically intercepted by `LoggerStore.setupGlobalErrorHandlers()`.
- Component-level error handlers should capture local context and pass it to `logger.error(...)`.

---

## 3. Log Exporting & Troubleshooting

Users and QA engineers can inspect, filter, and export application logs directly inside:
- **Settings Modal → System Logs Tab (`Logs`)**
- Exports available in **JSON format** and formatted **`.log` plain text**.
- One-click clipboard copy for inclusion in GitHub issues and bug reports.
