# Implementation Plan: Fix Loss of Image/File Attachments on Prompt Edit & Rollback

## 1. Problem Diagnosis & Root Cause
When a user edits the last prompt via "Edit last turn" (rollback), or restores a turn, the attachments/images are lost from the prompt box preview even though the user prompt originally had images/attachments attached.

### Root Cause Analysis:
1. **`addMessage` omits `attachments` in `ChatMessage`**:
   - In `App.svelte` line 479 (`executeTurn`), `addMessage` is called with `{ role: 'user', content: payload.text, images: payload.images, ... }` but `attachments: payload.attachments` is **completely omitted**.
   - In `session.svelte.ts` line 1495 (`addMessage`), `ChatMessage` object construction explicitly checks `images` and `toolCalls`, but never saves `attachments: message.attachments`.
   - As a result, when a message is saved to session history in store or storage, any non-image or file attachments (Markdown, PDFs, code files, document attachments) are dropped.
2. **`restorePrompt` in `Composer.svelte` does not reconstruct `attachedFiles` from `images` when `attachments` is empty**:
   - When a user submits only images (e.g. via Snapshot or File Picker), `attachedImages` contains the `VisionImage[]`.
   - When rolling back via `rollbackLastUserTurn`, `rollback.images` contains `VisionImage[]` (which has `filePath`, `dataUrl`, `id`, etc.), but `userMsg.attachments` was empty/undefined.
   - `restorePrompt({ text, images, attachments })` sets `attachedImages = payload.images` and `attachedFiles = payload.attachments || []`.
   - In `Composer.svelte`, the UI chip strip directly renders `{#if attachedFiles.length > 0}` (lines 1219-1235)!
   - Because `attachedFiles` was empty (or missing the image entries that were in `attachedImages`), the chip strip rendered **0 chips**, making the prompt box appear empty of any attached files or images.
   - However, when the user presses Enter, `executeTurn` received `images: [...attachedImages]`, which still had the images from `attachedImages`, sending them behind the scenes while visually invisible to the user in the prompt box!
3. **Queue edit synchronization (`handleEditQueuedPrompt`)**:
   - When editing a queued prompt that might only have `images` populated, `attachedFiles` should also be ensured to contain corresponding file chips so the prompt box always displays the image and file preview chips consistently.

## 2. Proposed Solution & Fixes

### Task 1: Preserve `attachments` in `sessionStore.addMessage` and `App.svelte.executeTurn`
- In `frontend/src/App.svelte` (`executeTurn`):
  Pass `attachments: payload.attachments && payload.attachments.length > 0 ? payload.attachments : undefined` to `sessionStore.addMessage`.
- In `frontend/src/lib/stores/session.svelte.ts` (`addMessage`):
  Assign `attachments: message.attachments ? [...message.attachments] : undefined` to the created `ChatMessage` object.

### Task 2: Reconstruct and Synchronize `attachedFiles` & `attachedImages` in `Composer.svelte`
- In `frontend/src/lib/components/chat/Composer.svelte` (`restorePrompt` and `handleEditQueuedPrompt`):
  - When `payload.images` are provided, ensure that for every image in `payload.images`, a corresponding entry exists in `attachedFiles` (with `isImage: true`, `dataUrl`, `filePath`, `name`, `id`) if not already present.
  - When `payload.attachments` are provided, ensure any image attachments (`isImage: true` or image mime-type) are synchronized into `attachedImages` if not already present.
  - This ensures two-way consistency: both the visual prompt box chips (`attachedFiles`) and the vision execution context (`attachedImages`) are fully populated and visible.

### Task 3: Unit Tests & Verification
- Add Vitest tests verifying:
  1. `sessionStore.addMessage` stores both `images` and `attachments`.
  2. `sessionStore.rollbackLastUserTurn` returns both `images` and `attachments`.
  3. `Composer` prompt restoration reconstructs attachment chips and vision images for visual display and submission.
- Run `go test -v ./pkg/grokrunner/...`, `npx vitest run`, and `npm --prefix frontend run build`.
