# Sinkronisasi & Verifikasi Menyeluruh Frontend-Backend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Memastikan seluruh sinkronisasi data frontend-backend (Settings, Workspaces, UI State, Runner Options, IPC Bridge) bekerja 100% konsisten, bebas race condition saat startup, dan tanpa regression runtime.

**Architecture:** Menerapkan audit interaksi end-to-end pada startup lifecycle, store hydration sequence, synchronous fallback, dynamic settings propagation ke Wails backend runner, dan validasi unit test serta packaging app.

**Tech Stack:** Go 1.27 + Wails v2 + Svelte 5 (Runes) + TypeScript + Vite + Tailwind CSS.

## Global Constraints
- Menggunakan pure-Go di backend (`pkg/storage`) tanpa dependensi CGo eksternal.
- Hydration stores (`SettingsStore`, `SessionStore`) harus aman saat Wails IPC bridge (`window.go.main.App`) belum siap (fallback gracefully).
- State updates pada settings & UI state harus debounced untuk efisiensi I/O disk.
- Semua unit test Go (`go test ./...`) dan build frontend (`vite build`) harus lulus 100%.

---

### Task 1: Audit & Harmonize Storage IPC Bridge and Startup Lifecycle

**Files:**
- Modify: `app.go:50-75`
- Modify: `frontend/src/lib/stores/settings.svelte.ts:75-105`
- Modify: `frontend/src/lib/stores/session.svelte.ts:220-295`
- Test: `pkg/storage/storage_test.go`

**Interfaces:**
- Consumes: `App.GetAppSettings()`, `App.SaveAppSettings()`, `App.GetWorkspaces()`, `App.SaveWorkspaces()`, `App.GetUIState()`, `App.SaveUIState()`
- Produces: Reliable startup hydration without race conditions or overwriting stored values with defaults.

- [ ] **Step 1: Write verification test for empty vs populated storage hydration**
- [ ] **Step 2: Run Go unit tests on storage package**
- [ ] **Step 3: Ensure Wails runtime bridge check handles retry if IPC is not yet bound on early load**
- [ ] **Step 4: Run frontend typecheck/build**

---

### Task 2: Synchronize Runner Execution Settings with Persistent Backend Store

**Files:**
- Modify: `pkg/grokrunner/runner.go:240-270`
- Modify: `app.go:40-60`
- Test: `pkg/grokrunner/stream_parser_test.go`

**Interfaces:**
- Consumes: `storage.StorageManager.GetSettings()` inside `grokrunner.Runner`
- Produces: Grok CLI args populated with default model, reasoning effort, and fallback configurations directly from persistent backend.

- [ ] **Step 1: Add unit test verifying Runner falls back to StorageManager settings when prompt options are omitted**
- [ ] **Step 2: Run grokrunner test suite**
- [ ] **Step 3: Verify all test cases pass**

---

### Task 3: Comprehensive Verification (Go Tests, Svelte Build, macOS App Bundle)

**Files:**
- Test: `go test -v ./...`
- Build: `npm run build --prefix frontend`
- Package: `bash build-macos.sh`

- [ ] **Step 1: Run all backend Go test suites**
- [ ] **Step 2: Compile frontend bundle with Vite**
- [ ] **Step 3: Build macOS production .app and .dmg bundle**
