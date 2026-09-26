# Implementation Plan: Header Zoom and External URL Routing

## Summary
Documenting the implementation and verification of window zoom vs macOS Space fullscreen handling and external URL routing.

## Completed Tasks
- [x] Task 1: Update header double-click handler in `App.svelte` to use `WindowToggleMaximise()` instead of `WindowFullscreen()`.
- [x] Task 2: Configure custom link renderer in `markdownRenderer.ts` with `target="_blank"` and `rel="noopener noreferrer"`.
- [x] Task 3: Add message-level link click handler in `MessageItem.svelte`.
- [x] Task 4: Add global capture link click listener in `App.svelte` to catch all external links.
- [x] Task 5: Run Svelte check and frontend production build.
- [x] Task 6: Document rules and architecture in persistent Superpowers workspace memory.
