# Bottom Dock Collapsed Bar Redesign Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign the collapsed bottom terminal bar to match the premium Ant Design Studio aesthetic, featuring individual clickable tab pills with active highlight, quick-add button, clear visual contrast, and sleek action controls.

**UX & Design Deficiencies in Current State:**
1. **Plain, flat appearance:** The current bar is a generic 28px strip with a single static text title and plain pill badge that looks like an unfinished placeholder.
2. **Missing clickable tabs:** Unlike the right dock collapsed bar which allows clicking individual tabs, the bottom dock only displays one title instead of clickable tabs for all open terminal instances (`Terminal 1`, `backend`, `frontend`, etc.).
3. **Imbalanced typography & borders:** Font weights and icon contrasts are too subtle, lacking clear separation from the prompt composer directly beneath it.

**Target Design Architecture:**
- **Height & Spacing:** `h-8` (32px), clean top divider `border-t border-ant-border`, background `bg-ant-bg-secondary/80 backdrop-blur-md`.
- **Left Tab Row:**
  - Horizontal scrollable list of tab chips matching the right dock's clickable behavior.
  - Active tab chip: `bg-ant-bg text-ant-primary border-ant-primary/40 font-medium shadow-xs`.
  - Inactive tab chips: `text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border-transparent`.
  - Clicking any tab chip immediately expands the terminal and switches focus to that tab.
  - Plus (`+`) button to create and immediately open a new terminal.
- **Right Action Cluster:**
  - Dock to right toggle button with custom tooltip.
  - Expand button styled as a refined icon button or micro-pill with keyboard hint or smooth hover state.

---

## Detailed Task Breakdown

### Task 1: Redesign Bottom Dock Collapsed Bar in TerminalPanel.svelte
- Modify: `frontend/src/lib/components/terminal/TerminalPanel.svelte`
- Replace single title in `dockPosition === 'bottom'` collapsed view with horizontal tab pills loop:
  ```svelte
  {#each activeTerminals as tab (tab.id)}
    {@const isActive = tab.id === activeTermId}
    <Tooltip title={`Open ${tab.title}`} placement="top">
      <button
        type="button"
        onclick={() => {
          terminalStore.switchTerminal(sessionId, tab.id);
          terminalStore.toggleSessionCollapse(sessionId, false);
          refitActiveTerminal();
        }}
        class="flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-mono transition border {isActive ? 'bg-ant-bg text-ant-primary border-ant-primary/40 shadow-xs' : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border-transparent'}"
      >
        <TerminalIcon size={11} class="{isActive ? 'text-ant-primary' : 'text-ant-text-secondary'} flex-shrink-0" />
        <span class="truncate max-w-[120px]">{tab.title}</span>
      </button>
    </Tooltip>
  {/each}
  ```
- Add inline `+` button to add new terminal from collapsed bar.
- Refine right actions (dock toggle, expand button).

### Task 2: Build & Native Package Verification
- Run `cd frontend && npm run build`.
- Rebuild native app with `./build-macos.sh`.
- Launch and verify appearance in both light and dark themes.
