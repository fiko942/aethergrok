# Superpowers Architecture & Implementation Spec: External URL & Header Zoom Handling

## Date: 2026-09-26

## Overview
This specification details two key window and webview behaviors in AetherGrok:
1. **Window Maximize/Zoom on Header Double-Click**: Filling the screen on the active desktop without creating an isolated macOS Space fullscreen.
2. **External URL Routing**: Intercepting link clicks across chat markdown, messages, and the entire webview so they open in the user's default OS browser instead of navigating inside the Wails webview.

---

## 1. Window Maximize / Zoom Behavior
- **Issue**: Triggering `WindowFullscreen()` creates a new virtual desktop (macOS Space) in Mission Control, which isolates the app from the user's workflow.
- **Solution**:
  - Bound `ondblclick={handleHeaderDoubleClick}` to `<header>` in `frontend/src/App.svelte`.
  - Use `window.runtime.WindowToggleMaximise()` to invoke native macOS `[window zoom:nil]`.
  - Check `WindowIsFullscreen()` to cleanly exit if fullscreen was accidentally entered (`WindowUnfullscreen()`).
  - Guard interactive UI elements (`button`, `input`, `select`, `textarea`, `a`, `[role="button"]`, `[role="tab"]`) with `target?.closest(...)` so standard clicks do not trigger window zoom.

---

## 2. External URL Routing to Default Browser
- **Issue**: Standard HTML anchor (`<a href="...">`) clicks in a webview cause the webview window to navigate directly to the target URL, replacing the entire application interface.
- **Solution**:
  - **Markdown Renderer (`frontend/src/lib/utils/markdownRenderer.ts`)**:
    - Configured marked renderer with custom `link` method:
      ```typescript
      customRenderer.link = function ({ href, title, text }) {
        const titleAttr = title ? ` title="${title}"` : '';
        return `<a href="${href}" target="_blank" rel="noopener noreferrer"${titleAttr} class="text-ant-primary hover:text-blue-400 underline underline-offset-2 transition">${text}</a>`;
      };
      ```
  - **Message Click Interceptor (`frontend/src/lib/components/chat/MessageItem.svelte`)**:
    - Catches any click on an `<a>` element in rendered messages and dispatches it through `App.OpenExternalURL(href)` or `window.runtime.BrowserOpenURL(href)`.
  - **Global Document Click Capture (`frontend/src/App.svelte`)**:
    - Added capture-phase listener `window.addEventListener('click', handleGlobalDocumentClick, true)`.
    - Detects links matching `^https?://` or `^mailto:` and prevents default navigation, opening the URL in the operating system's default browser (Safari, Chrome, etc.).

---

## Verification
- `npm --prefix frontend run check`: 0 errors.
- `npm --prefix frontend run build`: Clean production build output.
