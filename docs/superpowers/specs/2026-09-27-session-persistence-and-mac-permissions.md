# Specification: Session Persistence, Continuation & Persistent macOS Permissions

## Overview
This superpower specification details the architecture, implementation, and verification for:
1. **Multi-Turn Session Continuation (`--resume`)**: Preserving context and disk space by continuing existing sessions rather than spawning fresh ones per prompt.
2. **Session History Loading**: Full reconstruction of chat turns, tool calls, and assistant messages without tail truncation.
3. **Persistent macOS Permissions Across Updates**: Ensuring macOS TCC permissions (Accessibility, Microphone, Speech Recognition, Automation) persist across version builds via Designated Requirement (`identifier "com.wails.aethergrok"`).

---

## 1. Multi-Turn Session Continuation

### Problem & Cause
By default, executing `grok -p "<prompt>"` in Grok CLI initializes a brand-new session UUID in `~/.grok/sessions/<workspace>/` every turn. Without passing `--resume <sessionId>`, multi-turn chats repeatedly spawn isolated directories on disk and fail to accumulate context.

### Solution Architecture
1. **Runner Parameterization (`pkg/grokrunner/runner.go`)**:
   - Inspect `req.Options.GrokSessionID` or `req.SessionID`.
   - If the session ID exists on disk in `~/.grok/sessions/<workspace>/<sessionId>`, invoke `grok --resume <sessionId> -p <prompt>`.
   - If starting an explicitly assigned new session, use `grok --session-id <uuid> -p <prompt>`.
2. **Stream Parser & Events (`pkg/grokrunner/stream_parser.go`)**:
   - When the `end` event streams from the Grok CLI process, emit `TurnCompleteEvent` containing the authoritative `GrokSessionID`.
3. **Frontend Integration (`frontend/src/App.svelte` & `session.svelte.ts`)**:
   - When submitting prompt turns via `RunPromptStream`, supply `options.grokSessionId` using `session.grokSessionId || session.id`.
   - Persist `grokSessionId` to the active session object upon turn completion.

---

## 2. Session History Loading & Rendering

### Parsing Mechanism (`pkg/grokrunner/session_scanner.go`)
- Scan and parse `chat_history.jsonl` (and fallback to `updates.jsonl`).
- Extract user queries from `<user_query>...</user_query>` envelopes, ignoring system reminder headers.
- Reconstruct assistant responses, tool calls (`name`, `arguments`, `result`), and status without omitting trailing responses.
- Use deterministic, indexed message IDs (`fmt.Sprintf("%s_u_%d", sessionID, lineIdx)` and `%s_a_%d`).

### Frontend State Synchronization
- In `loadSessionHistoryFromDisk` (`session.svelte.ts`), fetch history using `targetGrokId = session.grokSessionId || session.id`.
- Adjust `session.visibleTurnCount` to ensure all loaded messages from disk render cleanly without bottom truncation.

---

## 3. Persistent macOS Security & TCC Permissions Across Updates

### Problem & Cause
Standard ad-hoc code signing (`codesign --sign -`) binds authorization to the executable's current binary hash (`cdhash`). Every recompile changes the `cdhash`, causing macOS TCC (Transparency, Consent, and Control) to treat the app as untrusted and re-prompt for permissions on every update.

Additionally, invoking `osascript` with `tell application "System Settings"` triggers unnecessary AppleEvents/Automation permission prompts.

### Solution Architecture
1. **Designated Requirement Code Signing (`scripts/build-production.mjs`)**:
   - Apply explicit Designated Requirement tied to the Bundle ID during post-build re-signing:
     ```bash
     codesign --force --deep --sign - --requirements '= designated => identifier "com.wails.aethergrok"' build/bin/aethergrok.app
     ```
   - This ensures macOS TCC recognizes all subsequent builds with the same identifier, preserving user permissions across version updates.
2. **Direct Preference URL Launchers (`pkg/permissions/`)**:
   - Replace `osascript` scripts with direct URL calls (`open x-apple.systempreferences:com.apple.preference.security?...`) in `permissions_darwin.go` and `microphone_darwin.go`.
3. **Info.plist Permission Declarations (`resources/app-assets/darwin/Info.plist`)**:
   - Maintain explicit usage description strings for `NSMicrophoneUsageDescription`, `NSSpeechRecognitionUsageDescription`, and `NSAppleEventsUsageDescription`.

---

## 4. Verification & Testing
- **Go Unit Tests**: `pkg/grokrunner/session_scanner_test.go` verifies session discovery and chronological turn extraction.
- **Code Signing Verification**: `codesign -d -r- /Applications/aethergrok.app` outputs `designated => identifier "com.wails.aethergrok"`.
- **Frontend Build**: Verified clean Vite compilation and Svelte 5 component checks.
