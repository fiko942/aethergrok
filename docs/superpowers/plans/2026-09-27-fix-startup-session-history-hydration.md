# Fix Startup Session History Hydration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ensure that when AetherGrok launches with a previously opened active session tab, its chat message history and token usage are immediately hydrated from disk without requiring the user to close and reopen the session from the sidebar.

**Architecture:** Update `SessionStore` lifecycle in `frontend/src/lib/stores/session.svelte.ts` so that:
1. On app startup (inside the initial async verification and sync lifecycle), the active restored session is automatically hydrated via `loadSessionHistoryFromDisk` and `loadSessionUsage`.
2. `switchSession(id)` checks if the session currently has empty messages (`messages.length === 0`), and if so, triggers hydration even if `activeSessionId === id`.
3. `reconcileSessionState()` or session recovery safeguards ensure any active tab without messages gets hydrated from disk on demand.

**Tech Stack:** Svelte 5 Runes (`$state`, `$derived`), TypeScript, Wails v2 Go backend IPC (`window.go.main.App.LoadGrokSessionHistory`, `window.go.main.App.GetSessionUsage`).

## Global Constraints
- State points directly in affirmative language. Avoid unnecessary contrastive negation.
- Maintain existing session metadata in `localStorage` without bloating it with full transcript arrays.
- Preserve light and dark theme UI styling and responsive behavior.
- Validate build with `./build-macos.sh` upon completion.

---

### Task 1: Auto-hydrate Active Session & Tab Histories on App Startup

**Files:**
- Modify: `frontend/src/lib/stores/session.svelte.ts:230-245`
- Modify: `frontend/src/lib/stores/session.svelte.ts:910-925`

**Interfaces:**
- Consumes: `loadSessionHistoryFromDisk(session: Session): Promise<void>`, `loadSessionUsage(session: Session): Promise<void>`
- Produces: Hydrated `session.messages` and `session.usage` immediately upon app launch and when clicking active tab

- [ ] **Step 1: Update `constructor` initialization sequence in `SessionStore`**

In `frontend/src/lib/stores/session.svelte.ts`, enhance the post-storage async init block to verify workspaces, sync discovered sessions on disk, and immediately hydrate the active session:

```ts
    // Trigger verification of workspaces and active session hydration on disk
    if (typeof window !== 'undefined') {
      setTimeout(async () => {
        await this.verifyAllWorkspaces();
        this.reconcileSessionState();
        await this.verifySessionsOnDisk();
        
        // Auto-hydrate the active session transcript and usage stats
        if (this.activeSessionId) {
          const activeSess = this.sessions.find((s) => s.id === this.activeSessionId);
          if (activeSess) {
            if (activeSess.messages.length === 0 || activeSess.grokSessionId) {
              await this.loadSessionHistoryFromDisk(activeSess);
            }
            await this.loadSessionUsage(activeSess);
          }
        }
      }, 50);
    }
```

- [ ] **Step 2: Update `switchSession` to hydrate if session messages are missing**

In `frontend/src/lib/stores/session.svelte.ts`, update `switchSession`:

```ts
  async switchSession(id: string): Promise<void> {
    if (this.activeSessionId === id) {
      const target = this.sessions.find((s) => s.id === id);
      if (target && target.messages.length === 0) {
        await this.loadSessionHistoryFromDisk(target);
        await this.loadSessionUsage(target);
      }
      return;
    }
    await this.openSessionInTab(id);
  }
```

- [ ] **Step 3: Verify TypeScript compilation and build**

Run: `cd frontend && npm run check` or `npm run build`
Expected: 0 errors, build succeeds cleanly.

---

### Task 2: Build Native macOS App and Verify

**Files:**
- Run build script: `./build-macos.sh`

- [ ] **Step 1: Execute macOS build script**

Run: `./build-macos.sh`
Expected: Build passes, signs `build/bin/aethergrok.app` with ad-hoc entitlements.

- [ ] **Step 2: Verify functionality**

Verify app launch lifecycle and ensure active tab chat transcripts load smoothly.
