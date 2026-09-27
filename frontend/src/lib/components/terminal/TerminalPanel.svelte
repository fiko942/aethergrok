<script lang="ts">
  import { onMount, onDestroy, tick } from 'svelte';
  import { Terminal as Xterm } from '@xterm/xterm';
  import { FitAddon } from '@xterm/addon-fit';
  import {
    Plus,
    X,
    Terminal as TerminalIcon,
    ChevronDown,
    ChevronUp,
    Send,
    Trash2,
    Maximize2,
    Minimize2,
    PanelBottom,
    PanelRight,
    Sparkles
  } from 'lucide-svelte';
  import Tooltip from '$lib/antd/Tooltip.svelte';
  import { terminalStore, type TerminalTab } from '$lib/stores/terminal.svelte';
  import { sessionStore } from '$lib/stores/session.svelte';
  import { settingsStore } from '$lib/stores/settings.svelte';

  const DARK_TERMINAL_THEME = {
    background: '#0e0e11',
    foreground: '#e4e4e7',
    cursor: '#38bdf8',
    selectionBackground: '#38bdf840',
    black: '#18181b',
    red: '#f43f5e',
    green: '#10b981',
    yellow: '#f59e0b',
    blue: '#3b82f6',
    magenta: '#d946ef',
    cyan: '#06b6d4',
    white: '#f4f4f5',
    brightBlack: '#71717a',
    brightRed: '#fb7185',
    brightGreen: '#34d399',
    brightYellow: '#fbbf24',
    brightBlue: '#60a5fa',
    brightMagenta: '#e879f9',
    brightCyan: '#22d3ee',
    brightWhite: '#ffffff'
  };

  const LIGHT_TERMINAL_THEME = {
    background: '#ffffff',
    foreground: '#0f172a',
    cursor: '#0284c7',
    selectionBackground: '#bae6fd80',
    black: '#0f172a',
    red: '#e11d48',
    green: '#059669',
    yellow: '#d97706',
    blue: '#2563eb',
    magenta: '#c026d3',
    cyan: '#0891b2',
    white: '#f8fafc',
    brightBlack: '#64748b',
    brightRed: '#f43f5e',
    brightGreen: '#10b981',
    brightYellow: '#f59e0b',
    brightBlue: '#3b82f6',
    brightMagenta: '#d946ef',
    brightCyan: '#06b6d4',
    brightWhite: '#0f172a'
  };

  function getXtermTheme(theme: string) {
    return theme === 'light-antd' ? LIGHT_TERMINAL_THEME : DARK_TERMINAL_THEME;
  }

  interface Props {
    sessionId: string;
    workspacePath: string;
    onAttachLogToComposer?: (text: string) => void;
  }

  let { sessionId, workspacePath, onAttachLogToComposer }: Props = $props();

  let activeTerminals = $derived(terminalStore.getTerminalTabs(sessionId));
  let activeTermId = $derived(terminalStore.getActiveTerminalId(sessionId));
  let isMaximized = $state(false);
  let isDragging = $state(false);
  let startX = 0;
  let startY = 0;
  let startHeight = 0;
  let startWidth = 0;

  let dockPosition = $derived(terminalStore.getDockPosition(sessionId));
  let isCollapsed = $derived(terminalStore.isSessionCollapsed(sessionId));

  let editingTermId = $state<string | null>(null);
  let editingTitle = $state<string>('');

  function selectOnFocus(node: HTMLInputElement) {
    // Immediate select and focus
    node.focus();
    node.select();
  }

  function startRename(tab: TerminalTab, e: MouseEvent) {
    e.stopPropagation();
    editingTermId = tab.id;
    editingTitle = tab.title;
  }

  function saveRename(tab: TerminalTab) {
    if (editingTermId === tab.id && editingTitle.trim()) {
      terminalStore.renameTerminal(sessionId, tab.id, editingTitle.trim());
    }
    editingTermId = null;
    editingTitle = '';
  }

  function handleRenameKeyDown(e: KeyboardEvent, tab: TerminalTab) {
    if (e.key === 'Enter') {
      e.preventDefault();
      saveRename(tab);
    } else if (e.key === 'Escape') {
      e.preventDefault();
      editingTermId = null;
      editingTitle = '';
    }
  }
  const terminalInstances = new Map<string, { term: Xterm; fitAddon: FitAddon; unsub?: () => void }>();
  let termContainerMap = new Map<string, HTMLElement>();

  // Auto create initial terminal if none exists and panel is opened
  $effect(() => {
    if (terminalStore.isOpen && activeTerminals.length === 0 && sessionId) {
      terminalStore.createTerminal(sessionId, workspacePath);
    }
  });

  // Dynamically update xterm themes when settingsStore.theme changes
  $effect(() => {
    const currentTheme = getXtermTheme(settingsStore.theme);
    for (const inst of terminalInstances.values()) {
      inst.term.options.theme = currentTheme;
    }
  });

  function terminalContainerAction(node: HTMLElement, termId: string) {
    termContainerMap.set(termId, node);
    initXterm(termId, node);

    return {
      destroy() {
        termContainerMap.delete(termId);
        cleanupXterm(termId);
      }
    };
  }

  function initXterm(termId: string, container: HTMLElement) {
    if (terminalInstances.has(termId)) {
      const existing = terminalInstances.get(termId)!;
      try {
        existing.fitAddon.fit();
      } catch (e) {}
      return;
    }

    const term = new Xterm({
      cursorBlink: true,
      cursorStyle: 'bar',
      fontSize: 12.5,
      fontFamily: 'SF Mono, Menlo, Monaco, Consolas, "Liberation Mono", monospace',
      theme: getXtermTheme(settingsStore.theme),
      allowTransparency: true,
      scrollback: 5000
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(container);

    setTimeout(() => {
      try {
        fitAddon.fit();
        if (window.go?.main?.App?.ResizeTerminal) {
          window.go.main.App.ResizeTerminal(termId, term.cols, term.rows);
        }
      } catch (e) {}
    }, 50);

    // User typing in terminal -> send to Go backend PTY
    term.onData((data) => {
      if (window.go?.main?.App?.WriteTerminal) {
        window.go.main.App.WriteTerminal(termId, data);
      }
    });

    // Handle incoming data from backend PTY
    let unsub: (() => void) | undefined;
    if (window.runtime?.EventsOn) {
      unsub = window.runtime.EventsOn(`terminal:data:${termId}`, (data: string) => {
        term.write(data);
      });
    }

    terminalInstances.set(termId, { term, fitAddon, unsub });
  }

  function cleanupXterm(termId: string) {
    const inst = terminalInstances.get(termId);
    if (inst) {
      if (inst.unsub) inst.unsub();
      inst.term.dispose();
      terminalInstances.delete(termId);
    }
  }

  let rafId: number | null = null;

  function handleResizeStart(e: MouseEvent) {
    e.preventDefault();
    isDragging = true;
    startX = e.clientX;
    startY = e.clientY;
    startHeight = terminalStore.panelHeight;
    startWidth = terminalStore.panelWidth;

    // Prevent text selection across the whole window while resizing
    document.body.style.userSelect = 'none';
    document.body.style.cursor = dockPosition === 'bottom' ? 'row-resize' : 'col-resize';
    window.getSelection()?.removeAllRanges();

    window.addEventListener('mousemove', handleResizeMove);
    window.addEventListener('mouseup', handleResizeEnd);
  }

  function handleResizeMove(e: MouseEvent) {
    if (!isDragging) return;
    if (rafId) return;

    rafId = requestAnimationFrame(() => {
      rafId = null;
      if (!isDragging) return;
      if (dockPosition === 'bottom') {
        const delta = startY - e.clientY;
        terminalStore.setPanelHeight(startHeight + delta);
      } else {
        const delta = startX - e.clientX;
        terminalStore.setPanelWidth(startWidth + delta);
      }
    });
  }

  function handleResizeEnd() {
    isDragging = false;
    if (rafId) {
      cancelAnimationFrame(rafId);
      rafId = null;
    }

    document.body.style.userSelect = '';
    document.body.style.cursor = '';

    window.removeEventListener('mousemove', handleResizeMove);
    window.removeEventListener('mouseup', handleResizeEnd);
    refitActiveTerminal();
  }

  function refitActiveTerminal() {
    tick().then(() => {
      if (activeTermId && terminalInstances.has(activeTermId)) {
        const inst = terminalInstances.get(activeTermId)!;
        try {
          inst.fitAddon.fit();
          if (window.go?.main?.App?.ResizeTerminal) {
            window.go.main.App.ResizeTerminal(activeTermId, inst.term.cols, inst.term.rows);
          }
        } catch (e) {}
      }
    });
  }

  function handleAddTerminal() {
    terminalStore.createTerminal(sessionId, workspacePath);
    refitActiveTerminal();
  }

  function toggleDock() {
    terminalStore.toggleDockPosition(sessionId);
    refitActiveTerminal();
  }

  function handleClearTerminal() {
    if (activeTermId && terminalInstances.has(activeTermId)) {
      const inst = terminalInstances.get(activeTermId)!;
      inst.term.clear();
    }
  }

  function handleSendToAgent() {
    if (!activeTermId || !terminalInstances.has(activeTermId)) return;
    const inst = terminalInstances.get(activeTermId)!;
    
    // Get terminal buffer content or selection
    const term = inst.term;
    const selection = term.getSelection();
    let textToSend = selection;

    if (!textToSend) {
      // Extract last 100 lines from buffer
      const buffer = term.buffer.active;
      const lines: string[] = [];
      const startLine = Math.max(0, buffer.length - 80);
      for (let i = startLine; i < buffer.length; i++) {
        const line = buffer.getLine(i);
        if (line) {
          lines.push(line.translateToString(true));
        }
      }
      textToSend = lines.join('\n').trim();
    }

    if (textToSend && onAttachLogToComposer) {
      const activeTabObj = activeTerminals.find((t) => t.id === activeTermId);
      const title = activeTabObj ? activeTabObj.title : 'Terminal Output';
      const formatted = `Here is the recent log output from \`${title}\`:\n\n\`\`\`bash\n${textToSend}\n\`\`\``;
      onAttachLogToComposer(formatted);
    }
  }

  $effect(() => {
    // When switching active tab, refit the newly focused terminal
    if (activeTermId) {
      refitActiveTerminal();
    }
  });

  onDestroy(() => {
    // Clean up all frontend terminal listeners
    for (const [id, inst] of terminalInstances.entries()) {
      if (inst.unsub) inst.unsub();
      inst.term.dispose();
    }
    terminalInstances.clear();
  });
</script>

{#if activeTerminals.length > 0}
  {#if isCollapsed}
    <!-- Slim Collapsed Bar (Session Isolated) -->
    {#if dockPosition === 'bottom'}
      <div class="h-7 w-full bg-ant-bg-secondary border-t border-ant-border flex items-center justify-between px-3 text-xs text-ant-text-secondary select-none flex-shrink-0 z-20">
        <div class="flex items-center space-x-2">
          <TerminalIcon size={12} class="text-ant-primary" />
          <span class="font-mono text-[11px] font-medium text-ant-text truncate max-w-[180px]">
            {activeTerminals.find((t) => t.id === activeTermId)?.title || 'Terminal'}
          </span>
          <span class="px-1.5 py-0.2 text-[10px] rounded-full bg-ant-bg-tertiary text-ant-text-secondary border border-ant-border">
            {activeTerminals.length} tab{activeTerminals.length > 1 ? 's' : ''}
          </span>
        </div>
        <div class="flex items-center space-x-1">
          <Tooltip title="Dock to right" placement="top">
            <button
              type="button"
              onclick={toggleDock}
              class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
            >
              <PanelRight size={12} />
            </button>
          </Tooltip>
          <Tooltip title="Expand terminal panel" placement="top">
            <button
              type="button"
              onclick={() => {
                terminalStore.toggleSessionCollapse(sessionId, false);
                refitActiveTerminal();
              }}
              class="flex items-center space-x-1 px-2 py-0.5 text-[11px] rounded bg-ant-primary/10 text-ant-primary hover:bg-ant-primary/20 transition font-medium"
            >
              <ChevronUp size={13} />
              <span>Expand</span>
            </button>
          </Tooltip>
        </div>
      </div>
    {:else}
      <!-- Right Dock Slim Collapsed Bar with Individual Clickable Tab Pills -->
      <div class="w-9 h-full bg-ant-bg-secondary border-l border-ant-border flex flex-col items-center justify-between py-2 text-xs text-ant-text-secondary select-none flex-shrink-0 z-20">
        <!-- Top Tab Stack: Each terminal tab is clickable and focuses that terminal -->
        <div class="flex flex-col items-center space-y-2 overflow-y-auto no-scrollbar w-full px-1">
          {#each activeTerminals as tab (tab.id)}
            {@const isActive = tab.id === activeTermId}
            <Tooltip title={`Open ${tab.title}`} placement="left">
              <button
                type="button"
                onclick={() => {
                  terminalStore.switchTerminal(sessionId, tab.id);
                  terminalStore.toggleSessionCollapse(sessionId, false);
                  refitActiveTerminal();
                }}
                class="flex flex-col items-center py-2 px-1 rounded-md transition w-full {isActive ? 'bg-ant-primary/15 text-ant-primary font-medium border border-ant-primary/30 shadow-sm' : 'text-ant-text-secondary hover:text-ant-text hover:bg-white/5 border border-transparent'}"
              >
                <TerminalIcon size={12} class="{isActive ? 'text-ant-primary' : 'text-ant-text-secondary'} mb-1.5 flex-shrink-0" />
                <span class="text-[10px] font-mono [writing-mode:vertical-rl] tracking-wide truncate max-h-[100px] leading-tight">
                  {tab.title}
                </span>
              </button>
            </Tooltip>
          {/each}

          <Tooltip title="New Terminal" placement="left">
            <button
              type="button"
              onclick={() => {
                terminalStore.createTerminal(sessionId, workspacePath);
                terminalStore.toggleSessionCollapse(sessionId, false);
                refitActiveTerminal();
              }}
              class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
            >
              <Plus size={13} />
            </button>
          </Tooltip>
        </div>

        <!-- Bottom Controls for Collapsed Strip -->
        <div class="flex flex-col items-center space-y-1.5 pt-2 border-t border-ant-border/40 w-full">
          <Tooltip title="Dock to bottom" placement="left">
            <button
              type="button"
              onclick={toggleDock}
              class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
            >
              <PanelBottom size={12} />
            </button>
          </Tooltip>
          <Tooltip title="Expand terminal panel" placement="left">
            <button
              type="button"
              onclick={() => {
                terminalStore.toggleSessionCollapse(sessionId, false);
                refitActiveTerminal();
              }}
              class="p-1 text-ant-primary hover:bg-ant-primary/10 rounded transition"
            >
              <ChevronDown size={14} class="rotate-90" />
            </button>
          </Tooltip>
        </div>
      </div>
    {/if}
  {:else}
    <!-- Expanded Terminal Viewport (Bottom or Right Dock) -->
    {#if dockPosition === 'bottom'}
      <div
        class="flex flex-col w-full bg-ant-bg border-t border-ant-border z-20 select-none font-serif transition-all duration-75 flex-shrink-0"
        style="height: {isMaximized ? 'calc(100vh - 120px)' : `${terminalStore.panelHeight}px`}; min-height: 140px;"
      >
        <!-- Drag Resize Handle (Top border for Bottom dock) -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <div
          role="separator"
          class="h-1.5 w-full cursor-row-resize bg-transparent hover:bg-ant-primary/40 active:bg-ant-primary transition-colors flex items-center justify-center -mt-0.5 z-30"
          onmousedown={handleResizeStart}
        >
          <div class="w-10 h-0.5 bg-ant-border rounded-full"></div>
        </div>

        <!-- Terminal Header & Multi-Tab Bar -->
        <div class="flex items-center justify-between px-3 h-8 bg-ant-bg-secondary border-b border-ant-border flex-shrink-0 text-xs text-ant-text">
          <!-- Left: Terminal Sub-Tabs -->
          <div class="flex items-center space-x-1 overflow-x-auto no-scrollbar flex-1 min-w-0 pr-2">
            {#each activeTerminals as tab (tab.id)}
              {@const isActive = tab.id === activeTermId}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div
                class="flex items-center space-x-1.5 px-2.5 py-1 rounded-t-md text-xs transition-colors cursor-pointer border-t-2 {isActive ? 'bg-ant-bg text-ant-text border-ant-primary font-medium shadow-sm' : 'border-transparent text-ant-text-secondary hover:text-ant-text hover:bg-white/5'}"
                onclick={() => terminalStore.switchTerminal(sessionId, tab.id)}
                ondblclick={(e) => startRename(tab, e)}
              >
                <TerminalIcon size={12} class={isActive ? 'text-ant-primary' : 'text-ant-text-secondary'} />
                {#if editingTermId === tab.id}
                  <!-- svelte-ignore a11y_autofocus -->
                  <input
                    type="text"
                    use:selectOnFocus
                    bind:value={editingTitle}
                    class="px-1 py-0.5 text-[11px] font-mono rounded bg-ant-bg-tertiary text-ant-text border border-ant-primary focus:outline-none w-20"
                    onclick={(e) => e.stopPropagation()}
                    onblur={() => saveRename(tab)}
                    onkeydown={(e) => handleRenameKeyDown(e, tab)}
                  />
                {:else}
                  <span class="truncate max-w-[120px] font-mono text-[11px] select-none" title="Double click to rename">{tab.title}</span>
                {/if}
                <Tooltip title="Kill process & close terminal tab" placement="top">
                  <button
                    type="button"
                    onclick={(e) => {
                      e.stopPropagation();
                      terminalStore.closeTerminal(sessionId, tab.id);
                    }}
                    class="text-ant-text-secondary hover:text-rose-400 p-0.5 rounded transition hover:bg-white/10"
                  >
                    <X size={11} />
                  </button>
                </Tooltip>
              </div>
            {/each}

            <Tooltip title="New Terminal" placement="top">
              <button
                type="button"
                onclick={handleAddTerminal}
                class="p-1 rounded text-ant-text-secondary hover:text-ant-text hover:bg-white/10 transition"
              >
                <Plus size={13} />
              </button>
            </Tooltip>
          </div>

          <!-- Right: Action Buttons -->
          <div class="flex items-center space-x-1.5 flex-shrink-0">
            <Tooltip title="Attach recent terminal output to composer" placement="top">
              <button
                type="button"
                onclick={handleSendToAgent}
                class="flex items-center space-x-1 px-2.5 py-1 rounded-md bg-ant-primary/10 text-ant-primary hover:bg-ant-primary/20 text-[11px] font-medium transition outline-none"
              >
                <Send size={11} />
                <span>Send to Agent</span>
              </button>
            </Tooltip>

            <Tooltip title="Clear buffer" placement="top">
              <button
                type="button"
                onclick={handleClearTerminal}
                class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
              >
                <Trash2 size={12} />
              </button>
            </Tooltip>

            <Tooltip title={dockPosition === 'bottom' ? 'Dock to right' : 'Dock to bottom'} placement="top">
              <button
                type="button"
                onclick={toggleDock}
                class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
              >
                <PanelRight size={13} />
              </button>
            </Tooltip>

            <Tooltip title={isMaximized ? 'Restore size' : 'Maximize terminal'} placement="top">
              <button
                type="button"
                onclick={() => {
                  isMaximized = !isMaximized;
                  refitActiveTerminal();
                }}
                class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
              >
                {#if isMaximized}
                  <Minimize2 size={12} />
                {:else}
                  <Maximize2 size={12} />
                {/if}
              </button>
            </Tooltip>

            <Tooltip title="Collapse terminal panel" placement="top">
              <button
                type="button"
                onclick={() => terminalStore.toggleSessionCollapse(sessionId, true)}
                class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
              >
                <ChevronDown size={14} />
              </button>
            </Tooltip>
          </div>
        </div>

        <!-- Terminal Canvas Containers -->
        <div class="flex-1 w-full relative overflow-hidden p-1.5 bg-ant-bg {isDragging ? 'pointer-events-none' : ''}">
          {#each activeTerminals as tab (tab.id)}
            <div
              use:terminalContainerAction={tab.id}
              class="w-full h-full {tab.id === activeTermId ? 'block' : 'hidden'}"
            ></div>
          {/each}
        </div>
      </div>
    {:else}
      <!-- Right Docked Panel -->
      <div
        class="flex flex-row h-full bg-ant-bg border-l border-ant-border z-20 select-none font-serif transition-all duration-75 flex-shrink-0"
        style="width: {terminalStore.panelWidth}px; min-width: 260px;"
      >
        <!-- Left Drag Resize Handle for Right Dock -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <div
          role="separator"
          class="w-1.5 h-full cursor-col-resize bg-transparent hover:bg-ant-primary/40 active:bg-ant-primary transition-colors flex flex-col items-center justify-center -ml-0.5 z-30"
          onmousedown={handleResizeStart}
        >
          <div class="w-0.5 h-10 bg-ant-border rounded-full"></div>
        </div>

        <div class="flex-1 flex flex-col h-full min-w-0 overflow-visible">
          <!-- Terminal Header & Multi-Tab Bar for Right Dock -->
          <div class="flex items-center justify-between px-2.5 h-8 bg-ant-bg-secondary border-b border-ant-border flex-shrink-0 text-xs text-ant-text relative z-30">
            <!-- Left: Terminal Sub-Tabs -->
            <div class="flex items-center space-x-1 overflow-x-auto no-scrollbar flex-1 min-w-0 pr-1">
              {#each activeTerminals as tab (tab.id)}
                {@const isActive = tab.id === activeTermId}
                <!-- svelte-ignore a11y_click_events_have_key_events -->
                <!-- svelte-ignore a11y_no_static_element_interactions -->
                <div
                  class="flex items-center space-x-1.5 px-2.5 py-1 rounded-t-md text-xs transition-colors cursor-pointer border-t-2 flex-shrink-0 {isActive ? 'bg-ant-bg text-ant-text border-ant-primary font-medium shadow-sm' : 'border-transparent text-ant-text-secondary hover:text-ant-text hover:bg-white/5'}"
                  onclick={() => terminalStore.switchTerminal(sessionId, tab.id)}
                  ondblclick={(e) => startRename(tab, e)}
                >
                  <TerminalIcon size={11} class="{isActive ? 'text-ant-primary' : 'text-ant-text-secondary'} flex-shrink-0" />
                  {#if editingTermId === tab.id}
                    <!-- svelte-ignore a11y_autofocus -->
                    <input
                      type="text"
                      use:selectOnFocus
                      bind:value={editingTitle}
                      class="px-1 py-0.5 text-[10px] font-mono rounded bg-ant-bg-tertiary text-ant-text border border-ant-primary focus:outline-none w-20"
                      onclick={(e) => e.stopPropagation()}
                      onblur={() => saveRename(tab)}
                      onkeydown={(e) => handleRenameKeyDown(e, tab)}
                    />
                  {:else}
                    <span class="font-mono text-[11px] whitespace-nowrap select-none" title="Double click to rename">{tab.title}</span>
                  {/if}
                  <Tooltip title="Kill process & close terminal tab" placement="bottom">
                    <button
                      type="button"
                      onclick={(e) => {
                        e.stopPropagation();
                        terminalStore.closeTerminal(sessionId, tab.id);
                      }}
                      class="text-ant-text-secondary hover:text-rose-400 p-0.5 rounded transition hover:bg-white/10 flex-shrink-0 ml-0.5"
                    >
                      <X size={11} />
                    </button>
                  </Tooltip>
                </div>
              {/each}

            <Tooltip title="New Terminal" placement="bottom">
              <button
                type="button"
                onclick={handleAddTerminal}
                class="p-0.5 rounded text-ant-text-secondary hover:text-ant-text hover:bg-white/10 transition"
              >
                <Plus size={12} />
              </button>
            </Tooltip>
          </div>

          <!-- Right Action Controls -->
          <div class="flex items-center space-x-1 flex-shrink-0">
            <Tooltip title="Send output to Agent" placement="bottom">
              <button
                type="button"
                onclick={handleSendToAgent}
                class="flex items-center space-x-0.5 px-1.5 py-0.5 rounded bg-ant-primary/10 text-ant-primary hover:bg-ant-primary/20 text-[10px] font-medium transition"
              >
                <Send size={10} />
              </button>
            </Tooltip>

            <Tooltip title="Clear buffer" placement="bottom">
              <button
                type="button"
                onclick={handleClearTerminal}
                class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
              >
                <Trash2 size={11} />
              </button>
            </Tooltip>

            <Tooltip title="Dock to bottom" placement="bottom">
              <button
                type="button"
                onclick={toggleDock}
                class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
              >
                <PanelBottom size={12} />
              </button>
            </Tooltip>

            <Tooltip title="Collapse terminal" placement="bottom">
              <button
                type="button"
                onclick={() => terminalStore.toggleSessionCollapse(sessionId, true)}
                class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
              >
                <ChevronDown size={13} class="rotate-[-90deg]" />
              </button>
            </Tooltip>
          </div>
          </div>

          <!-- Terminal Canvas Containers for Right Dock -->
          <div class="flex-1 w-full relative overflow-hidden p-1 bg-ant-bg {isDragging ? 'pointer-events-none' : ''}">
            {#each activeTerminals as tab (tab.id)}
              <div
                use:terminalContainerAction={tab.id}
                class="w-full h-full {tab.id === activeTermId ? 'block' : 'hidden'}"
              ></div>
            {/each}
          </div>
        </div>
      </div>
    {/if}
  {/if}
{/if}
