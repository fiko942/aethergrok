# Specification: Multi-Session Scalability & High-Performance Session Switching

## 1. Overview
When working across multiple projects with dozens or hundreds of sessions, switching between conversations must be instant (sub-50ms) and free of event-loop blockages or "Not Responding" freezes on Windows.

---

## 2. Root Cause & Architecture Overhaul

### 2.1 Git Root Resolution Memoization (`src/worktree.ts`)
- **Problem**: `sessionCwdsForRepo` and `localTrustedSessionCwds` traverse up to 64 parent directory levels checking for `.git` markers. Across 10 repos, this triggered hundreds of redundant `fs.existsSync` calls on every permission check.
- **Solution**: Implemented `gitRootCache` with a 15,000ms TTL. Repeated lookups for identical workspace directories return instantly in $O(1)$ time.

### 2.2 Stat Optimization & Modern Session Fast-Path (`src/sessions.ts`)
- **Problem**: `statSessionActivity` was statting `events.jsonl` first, and if found, statting `updates.jsonl`. For 99% of sessions, this meant 2 synchronous stats per session folder on every scan pass.
- **Solution**: Reordered lookup to check `updates.jsonl` first. If present, it immediately returns `{ mtimeMs, hasTranscript: true }` in **1 single stat**, cutting filesystem I/O in half.

### 2.3 Trusted CWD Memoization (`src/sidebar.ts`)
- **Problem**: `authorizedSessionCwds` and `isAuthorizedCwd` were computing directory arrays from scratch on every permission check, calling `localTrustedSessionCwds` dozens of times during a single session switch.
- **Solution**: Implemented `trustedCwdsCache` bound to `authEpoch` with a 3,000ms TTL, eliminating repetitive catalog and worktree loops.

### 2.4 Unified Index Session Caching (`src/sidebar.ts`)
- **Problem**: `buildPinnedSessions`, `resolveLocalRepoTarget`, `clearAllSessions`, and `sweepEmptySessions` were calling un-cached `indexSessions` directly, bypassing `cachedIndexSessions`.
- **Solution**: Routed all session index queries through `this.cachedIndexSessions` with 5,000ms TTL and proper cache invalidation on session creation, renaming, and deletion.

### 2.5 Projects Rail Render Deduplication (`media/projects-rail.js`)
- **Problem**: Receiving `sessionName` frame triggered full `render()` unconditionally even if `state.activeSessionId` had not changed.
- **Solution**: Added identity guard `if (state.activeSessionId !== msg.sessionId)` to prevent redundant DOM rebuilding.

---

## 3. Verification
- **Test Suite**: 251 test files and 6,086 tests passing 100% (`pnpm test`).
- **Build**: Successfully compiled and packaged `win-unpacked` and NSIS installer `Grok-Build-Desktop-4.5.2-win-x64.exe`.
