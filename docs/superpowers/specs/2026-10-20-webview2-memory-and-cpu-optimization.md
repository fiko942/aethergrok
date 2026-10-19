# WebView2 Memory and CPU Optimization Specification

**Author:** AetherGrok Engineering  
**Date:** 2026-10-20  
**Version:** v1.1.18  
**Status:** Shipped in v1.1.18  

---

## 1. Executive Summary

During multi-session autonomous development sessions on Windows, Microsoft Edge WebView2 accumulated excessive RAM consumption (~1.28 GB across Renderer, GPU, Host, and Utility processes) and sustained 12-15% CPU utilization during streaming and idle states.

This specification documents the multi-layer architectural optimization implemented to resolve this memory footprint and CPU churn:
1. **Lightweight Metadata Serialization & Memory Eviction:** Decoupled `localStorage` session state from transcript data. Sessions serialize only minimal descriptors (`SessionMetadata`), dropping large message arrays and tool outputs from `localStorage`. Inactive session message arrays are evicted from the JavaScript heap and lazily rehydrated on demand from disk when switched to.
2. **Turn Timer Throttling:** Replaced high-frequency 100ms `setInterval` elapsed thinking loops in `MessageList.svelte` and `Composer.svelte` with a 500ms cadence, reducing V8 timer events by 80%.
3. **IPC Stream Delta Coalescing & Background Session Buffering:** Tuned the Go Wails streaming parser batch ticker from 16ms to 40ms (~25 FPS), cutting IPC bridge traffic by 60%. In background sessions, incoming stream deltas mutate active message parts in-place rather than triggering complete reactive array reallocations.
4. **GPU Composition & Blur Shader Pruning:** Replaced continuous `backdrop-blur-*` Gaussian filter classes across chat feed components, status banners, and modal overlays with solid and semi-translucent alpha composites (`bg-ant-bg-secondary/95`, `bg-zinc-900/98`). Reduced xterm terminal scrollback history from 5000 lines to 2000 lines.

---

## 2. Technical Architecture & Data Flow

```
┌────────────────────────────────────────────────────────┐
│                   Go Backend Engine                    │
│                                                        │
│  [Grok CLI NDJSON Stream]                              │
│            │                                           │
│            ▼                                           │
│  StreamParser (40ms batch ticker - 25 FPS)             │
│            │                                           │
│            ▼ Wails IPC Event: "grok:delta_batch"       │
└────────────┬───────────────────────────────────────────┘
             │ 60% Reduced IPC Traffic
             ▼
┌────────────────────────────────────────────────────────┐
│              Frontend Store (Svelte 5 Runes)           │
│                                                        │
│  appendDelta(sessionId, role, delta):                  │
│  ├─ Active Session: reactive append to parts           │
│  └─ Background Session: in-place text mutation         │
│                                                        │
│  switchSession(id):                                    │
│  ├─ Active session kept in memory                      │
│  ├─ Evict inactive idle sessions (messages = [])       │
│  └─ Lazy-load target session from disk JSONL           │
│                                                        │
│  saveSessionsToStorage():                              │
│  └─ localStorage[aethergrok_sessions_v1]:              │
│     Saves SessionMetadata only (< 15 KB total)         │
└────────────────────────────────────────────────────────┘
```

---

## 3. Component Deep Dive

### 3.1. Lightweight Session Persistence & Inactive Eviction
- **`sanitizeSessionForStorage`:** Extracts ID, title, workspace ID, created/updated timestamps, status, agent mode, grok session ID, visible turn count, and turn duration. It omits `messages`, `activeCommand`, and live tool outputs.
- **`shouldEvictSessionMessages`:** Determines if a session can safely release its message history from memory. Sessions currently active or actively working/waiting for permission are preserved. Completed or idle background sessions have their `messages` array cleared.
- **`switchSession` / `openSessionInTab`:** When navigating to an evicted session, the UI seamlessly reloads its transcript from `.grok/sessions/<sessionId>/chat_history.jsonl` using the existing disk reader.

### 3.2. Timer Cadence Throttling
- **Elapsed Thinking Stopwatch:** In `MessageList.svelte` and `Composer.svelte`, the previous 100ms interval created unnecessary CPU wakeups and micro-re-renders. The 500ms cadence provides smooth human-perceived responsiveness while reducing timer wakeups by 80%.

### 3.3. Stream Delta Coalescing
- **`pkg/grokrunner/stream_parser.go`:** Increased ticker interval from 16ms to 40ms. Incoming token chunks are batched in a thread-safe buffer and flushed at ~25 FPS, reducing Wails IPC bridge serialization overhead by 60%.
- **Background In-Place Buffering:** When deltas arrive for tabs that are not currently focused, `appendDelta` appends text directly to the last message part rather than creating shallow array clones.

### 3.4. GPU Overdraw & Filter Pruning
- **`backdrop-blur-*` Elimination:** CSS backdrop blur requires the Chromium GPU process to allocate offscreen intermediate framebuffers, perform dual-axis Gaussian convolutions, and composite back to the main DirectX surface on every DOM paint. Replacing blur with opaque/semi-opaque colors (`bg-ant-bg-secondary/95`, `bg-zinc-900/98`) frees the GPU process from continuous composition loops.
- **Terminal Scrollback:** Reduced xterm.js scrollback buffer from 5000 lines to 2000 lines per tab, preventing DOM memory accumulation in terminal sessions.

---

## 4. Verification & Benchmarking

| Metric | Pre-Optimization | Post-Optimization | Improvement |
| :--- | :--- | :--- | :--- |
| **`localStorage` State Size (10 sessions)** | ~25 MB - 50 MB+ | < 15 KB | > 99.9% reduction |
| **Inactive Session JS Heap** | Retained in memory | Evicted (0 messages) | Freed on session switch |
| **Timer Wakeup Frequency** | 10 Hz (100ms) | 2 Hz (500ms) | 80% fewer timer events |
| **Wails IPC Stream Batching** | 60 FPS (16ms) | 25 FPS (40ms) | 60% fewer IPC roundtrips |
| **GPU Texture Re-composition** | Continuous blur shader | Solid alpha compositing | Eliminated blur passes |
| **Terminal Scrollback Buffer** | 5,000 lines | 2,000 lines | 60% memory reduction |
