# Superpower Specification: Floating Screenshot Drag-and-Drop Vision & Temporary Cache

## 1. Problem Definition & Root Cause

### Phenomenon
When taking a screenshot on macOS (Cmd+Shift+4 or Cmd+Shift+5), macOS displays a floating thumbnail at the bottom-right corner before writing the file to disk. If the user drags that thumbnail directly into the AetherGrok composer before macOS writes it to the Desktop:
1. The AI agent enters an indefinite hanging state (~5 minutes without progress).
2. The user experience degrades, giving the impression that the agent crashed or lost connection.

### Root Causes
1. **Invalid Grok CLI Vision Flag**:
   - `pkg/grokrunner/runner.go` previously appended `--image <path>` to the `grok` CLI command.
   - Grok CLI does not accept `--image` and fails with an argument error.
   - Grok CLI requires vision inputs formatted as ACP JSON content blocks passed via `--prompt-json <JSON>`.
2. **Missing Filesystem Path for Floating macOS Thumbnails**:
   - Floating screenshot thumbnails dragged before saving do not have a persistent filesystem path; they exist purely as in-memory data blobs inside the webview.
   - The frontend was passing a virtual file name (e.g. `image.png`) instead of an absolute disk path.

---

## 2. Risk Analysis & Mitigations (Analisis Risiko & Efek Samping)

1. **Risk of Disk Bloat from Dropped Screenshots**:
   - *Mitigation*: Dropped screenshot blobs are saved to `os.TempDir()` with the prefix `grok-snapshot-drop-*.png`.
   - They are automatically counted in **Settings > Cache** (`GetSnapshotCacheStats`) and purged when the user clicks **Clear Cache** (`ClearSnapshotCache`), as well as cleaned up when a session is closed.
2. **Risk of Broken Image Previews if Cache is Cleared**:
   - *Mitigation*: The frontend keeps the in-memory Base64 `dataUrl` in the message state (`VisionImage.dataUrl`), so chat history previews remain visible even if the temporary file on disk is deleted.
3. **Risk of Latency from High-Resolution Retina Screenshots**:
   - *Mitigation*: `SaveTemporaryImage` checks image dimensions and compresses any payload exceeding 650KB to high-quality JPEG, ensuring sub-second vision transfer to Grok CLI.

---

## 3. Implementation Details

1. **Backend Temporary Image Cache (`pkg/screen/cache.go` & `app.go`)**:
   - `SaveTemporaryImage(base64Data string, mimeType string) (*SnapshotResult, error)` decodes base64 data, compresses if necessary, and writes to `os.TempDir()` with prefix `grok-snapshot-drop-`.
   - Exposed to Wails frontend runtime via `app.go`.
2. **Vision Content Block Construction (`pkg/grokrunner/runner.go`)**:
   - When `len(req.Images) > 0`, `runner.go` encodes image bytes into ACP blocks and passes `--prompt-json <JSON>`:
     ```json
     [
       { "type": "image", "data": "<base64>", "mimeType": "image/png" },
       { "type": "text", "text": "<prompt>" }
     ]
     ```
   - Standard text-only turns continue using fast `-p <prompt>`.
3. **Frontend Integration (`Composer.svelte`)**:
   - In `processFiles`, when an image blob or screenshot proxy is dragged in, `Composer` automatically calls `window.go.main.App.SaveTemporaryImage` to obtain a real disk path, link it to `attachedImages`, and register it in temporary file tracking.

---

## 4. Verification

- `TestSnapshotCacheLifecycle`, `TestFormatBytes`, and `TestSaveTemporaryImage` in `pkg/screen` pass cleanly.
- `TestDiscoverGrokSessions_ExtractionAndFiltering` and `TestLoadGrokSessionMessages_CompleteSequence` pass cleanly.
- `wails build` succeeds and packages the application bundle.
