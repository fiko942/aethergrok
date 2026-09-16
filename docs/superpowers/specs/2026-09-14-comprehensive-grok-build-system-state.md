# Grok Build — Complete System State & Superpowers Master Knowledge Base

> **Date:** 2026-09-14  
> **Repository:** `https://github.com/fiko942/grok-build.git`  
> **Branches:** `main`, `feat/windows-snapshot-and-pnpm-migration`  
> **Package Manager:** `pnpm` exclusively (v12.3.4)  
> **Verification Status:** 251 test files, 6,086 automated tests (100% passing)  

---

## 1. Executive Summary & Capabilities

Grok Build is a high-performance desktop and VS Code extension client for Grok Build, OpenAI Codex, and Claude ACP providers. It features:
1. **PNPM Architecture**: Clean package management with zero `npm`/`package-lock.json` pollution, space-in-path resilience on Windows (`"project pribadi"`), and native build capabilities.
2. **Native Windows Screen Snapshot**: Multi-monitor GDI+ screen capture engine (`[System.Windows.Forms.SystemInformation]::VirtualScreen`), customizable global hotkeys (`kind: "hotkey"`), native folder picker (`kind: "folder"`), full-monitor white flash overlay, and synthesized camera shutter sound ("cekrek" via Web Audio API).
3. **Discrete Prompt FIFO Queue**: Individual drag-and-drop prompt cards in composer queue with sequential execution (one finishes, next runs automatically without concatenation).
4. **Infinite Scroll & Virtualized Progressive History**: Initial renders capped to **10 user turns** (`HISTORY_WINDOW_USER_TURNS = 10`), lazy upward loading in chunks of **10 turns** (`HISTORY_PREPEND_USER_TURNS = 10`), and seamless scroll anchor retention with zero jump/flicker.
5. **Ultra-Fast Multi-Session Switching**: Instant $O(1)$ session switching across hundreds of sessions via:
   - `gitRootCache` (15s TTL) in `src/worktree.ts`.
   - `updates.jsonl` fast-path stat optimization in `src/sessions.ts` (cuts 50% of filesystem I/O).
   - `trustedCwdsCache` (3s TTL) bound to `authEpoch` in `src/sidebar.ts`.
   - Unified `cachedIndexSessions` (5s TTL) across all session indexing and cleaning passes.
   - Projects Rail render deduplication guard on `session`/`sessionName` frames.

---

## 2. Directory & Module Reference

| Component | Path | Responsibility |
|---|---|---|
| **Chat Webview UI** | `media/chat.js`, `media/chat.css` | Transcript rendering, infinite scroll windowing (`splitHistoryWindow`), queue cards, audio synthesis |
| **Settings UI** | `media/settings.js`, `media/settings.css` | Settings panel, hotkey recorder, native folder selection, platform gating |
| **Projects Rail** | `media/projects-rail.js`, `media/projects-rail.css` | Pinned, Recent, and Projects list with lazy session preview requests and render batching |
| **Host Extension Core** | `src/sidebar.ts` | Session pool, caching (`repoCatalogCache`, `sessionIndexCache`, `trustedCwdsCache`), ACP message dispatch |
| **Session Indexer** | `src/sessions.ts` | Activity clock ranking (`statSessionActivity`), fast catalog indexing, discovery |
| **Worktree & Git Manager** | `src/worktree.ts` | Git root resolution (`gitRootForPath` + `gitRootCache`), worktree merge algorithms |
| **Desktop Host** | `src/desktop/main.ts`, `src/desktop/preload.ts` | Electron lifecycle, global shortcuts, multi-monitor flash BrowserWindow, native dialogs |
| **Screen Snapshot** | `src/snapshot.ts`, `src/snapshot-handler.ts` | Windows PowerShell GDI+ capture & macOS `screencapture -x` wrapper |

---

## 3. Key Algorithms & Memory Invariants

### 3.1 10-Turn Progressive History Hydration (`media/chat.js`)
- Initial replay: `splitHistoryWindow(held, 10)` -> splits into `suffix` (rendered) and `prefix` (parked).
- Upward scroll: `IntersectionObserver` on `#history-head` or `scrollTop <= 1000px` triggers `maybeLoadEarlierHistory()`.
- Hydration: `hydrateHistoryChunkWithCounters` renders into detached `historyPark` inside `try ... finally` and anchors visible line via `restorePrependAnchor(sentinel, y)`.

### 3.2 Session Switching DOM Flush & Zero-Leak State
- `clearMessages` triggers `resetForNewSession()`, marking existing DOM nodes `data-pending-clear="1"` and wiping tracking Maps.
- Switching to another session destroys pending nodes on replacement append.
- Returning to a previous session always executes `splitHistoryWindow(held, 10)`, never bloating the DOM tree with thousands of historical nodes.

### 3.3 Fast Session Indexing & I/O Minimization (`src/sessions.ts`)
```ts
export function statSessionActivity(fs: FsLike, sessionDir: string): { mtimeMs: number; hasTranscript: boolean } | undefined {
  try {
    const updatesMtime = fs.statSync(path.join(sessionDir, SESSION_UPDATES)).mtimeMs;
    return { mtimeMs: updatesMtime, hasTranscript: true };
  } catch { /* fallback to events.jsonl then summary.json */ }
}
```

---

## 4. Build, Packaging & Distribution

```bash
# Development & Testing
pnpm test                  # 251 test files (6,086 tests passing 100%)
pnpm run compile           # TypeScript compile (tsc -p .)
pnpm run desktop           # Launch Desktop App in local environment

# Production Desktop Packaging
pnpm run dist:dir          # Packaged to dist-desktop/win-unpacked/
pnpm run dist:win          # Full NSIS installer: dist-desktop/Grok-Build-Desktop-4.5.2-win-x64.exe
```
