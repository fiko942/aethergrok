# Superpowers Spec: Resizable, Collapsible, and Responsive Left Sidebar

**Date**: 2026-09-25  
**Version**: 1.0.0  
**Scope**: Frontend Layout, State Management, Responsive Drawer  

---

## Technical Specifications

### 1. State Model (`settingsStore`)
```typescript
interface AppSettings {
  // ... existing fields
  sidebarWidth: number; // 220 to 480 px, default: 288
  sidebarCollapsed: boolean; // default: false
}
```

### 2. Interaction Matrix

| Trigger | Condition | Result |
|---|---|---|
| Mouse Drag on Divider | Desktop (`>= 840px`), Not Collapsed | Clamps width between `220px` and `480px`. Dragging below `160px` sets `sidebarCollapsed = true`. |
| Double-Click Divider | Desktop (`>= 840px`) | Sets `sidebarWidth = 288` and `sidebarCollapsed = false`. |
| Header Toggle Button | Any viewport | Toggles `sidebarCollapsed = !sidebarCollapsed`. |
| Floating Edge Button | Collapsed state | Sets `sidebarCollapsed = false`. |
| `⌘ + B` / `Ctrl + B` | Global (outside text inputs) | Toggles `sidebarCollapsed = !sidebarCollapsed`. |
| Window Resize (`< 840px`) | Transitioning from desktop to compact | Automatically sets `sidebarCollapsed = true`. |
| Session Selected | Viewport `< 840px` | Opens session and sets `sidebarCollapsed = true` (auto-close drawer). |
| Backdrop Click / Esc | Viewport `< 840px` & drawer open | Sets `sidebarCollapsed = true`. |

---

## File Changes
1. `frontend/src/lib/stores/settings.svelte.ts`: Persistent sidebar width & collapse configuration.
2. `frontend/src/App.svelte`: Drag resize logic, header controls, edge toggle button, backdrop overlay.
3. `frontend/src/lib/components/layout/WorkspaceSidebar.svelte`: Auto-close session click handler for compact mode.
4. `frontend/src/lib/components/layout/SettingsModal.svelte`: Keyboard shortcut documentation.
