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
    Sparkles,
    Columns2,
    ExternalLink
  } from 'lucide-svelte';
  import Tooltip from '$lib/antd/Tooltip.svelte';
  import { terminalStore, type TerminalTab, type TerminalSplitGroup } from '$lib/stores/terminal.svelte';
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
    // Delay slightly to bypass the browser double-click selection collapse
    setTimeout(() => {
      node.focus();
      node.select();
      node.setSelectionRange(0, node.value.length);
    }, 20);
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

  let activeSplitGroup = $derived(terminalStore.getActiveSplitGroup(sessionId));
  let focusedPaneTermId = $derived(terminalStore.getFocusedPaneId(sessionId));

  // Drag and drop state for terminal tabs
  let draggedTermId = $state<string | null>(null);
  let activeDropZone = $state<{
    targetTermId: string;
    position: 'left' | 'right' | 'top' | 'bottom';
  } | null>(null);

  // ResizeObserver registry for split containers
  const resizeObservers = new Map<string, ResizeObserver>();

  function handleTabDragStart(e: DragEvent, tabId: string) {
    if (!e.dataTransfer) return;
    draggedTermId = tabId;
    e.dataTransfer.setData('text/plain', tabId);
    e.dataTransfer.setData('application/x-terminal-tab', tabId);
    e.dataTransfer.effectAllowed = 'move';
  }

  function handleTabDragEnd() {
    draggedTermId = null;
    activeDropZone = null;
  }

  function handlePaneDragOver(e: DragEvent, targetTermId: string) {
    if (!draggedTermId || draggedTermId === targetTermId) return;
    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }

    const target = e.currentTarget as HTMLElement;
    if (!target) return;

    const rect = target.getBoundingClientRect();
    const x = e.clientX - rect.left;
    const y = e.clientY - rect.top;
    const w = rect.width;
    const h = rect.height;

    // Calculate zone
    const isRightDock = dockPosition === 'right';
    if (isRightDock) {
      // Prioritize vertical splitting when docked at the right
      if (y < h * 0.5) {
        activeDropZone = { targetTermId, position: 'top' };
      } else {
        activeDropZone = { targetTermId, position: 'bottom' };
      }
    } else {
      // Prioritize horizontal splitting when docked at the bottom
      if (x < w * 0.4) {
        activeDropZone = { targetTermId, position: 'left' };
      } else if (x > w * 0.6) {
        activeDropZone = { targetTermId, position: 'right' };
      } else if (y > h * 0.6) {
        activeDropZone = { targetTermId, position: 'bottom' };
      } else {
        activeDropZone = { targetTermId, position: 'right' };
      }
    }
  }

  function handlePaneDragLeave(e: DragEvent) {
    // Only clear if leaving to an outside element
    const currentTarget = e.currentTarget as HTMLElement;
    const relatedTarget = e.relatedTarget as Node | null;
    if (!currentTarget.contains(relatedTarget)) {
      activeDropZone = null;
    }
  }

  function handlePaneDrop(e: DragEvent, targetTermId: string) {
    e.preventDefault();
    if (!draggedTermId || draggedTermId === targetTermId) {
      activeDropZone = null;
      draggedTermId = null;
      return;
    }

    const pos = activeDropZone?.position || 'right';
    terminalStore.splitTerminal(sessionId, draggedTermId, targetTermId, pos);

    activeDropZone = null;
    draggedTermId = null;
    refitAllSplitTerminals();
  }

  function handleQuickSplit() {
    terminalStore.splitNewTerminal(sessionId, workspacePath);
    refitAllSplitTerminals();
  }

  // Debounced xterm refit to prevent jank, layout thrashing, and backend PTY spam
  const pendingFitMap = new Map<string, number>();

  function scheduleXtermFit(termId: string) {
    // If the panel or divider is actively being dragged, skip fitting until drag ends
    if (isDragging || isResizingPanel || isDraggingSplitDivider) {
      return;
    }

    if (pendingFitMap.has(termId)) {
      cancelAnimationFrame(pendingFitMap.get(termId)!);
    }
    const id = requestAnimationFrame(() => {
      pendingFitMap.delete(termId);
      if (terminalInstances.has(termId)) {
        const inst = terminalInstances.get(termId)!;
        try {
          inst.fitAddon.fit();
          if (window.go?.main?.App?.ResizeTerminal) {
            window.go.main.App.ResizeTerminal(termId, inst.term.cols, inst.term.rows);
          }
        } catch (e) {}
      }
    });
    pendingFitMap.set(termId, id);
  }

  function refitAllSplitTerminals() {
    forceRefitAllSplitTerminals();
  }

  // Interactive Split Divider dragging
  let isDraggingSplitDivider = $state(false);
  let activeDividerIndex = $state<number | null>(null);
  let dividerStartPos = 0;
  let dividerInitialSizes: number[] = [];
  let dividerContainerSize = 0;
  let dividerRafId: number | null = null;

  function handleSplitDividerStart(e: MouseEvent, index: number, isVertical: boolean, containerEl: HTMLElement) {
    e.preventDefault();
    e.stopPropagation();

    const activeSplit = terminalStore.getActiveSplitGroup(sessionId);
    if (!activeSplit) return;

    isDraggingSplitDivider = true;
    activeDividerIndex = index;
    dividerStartPos = isVertical ? e.clientY : e.clientX;
    dividerContainerSize = isVertical ? containerEl.clientHeight : containerEl.clientWidth;

    const count = activeSplit.paneTermIds.length;
    dividerInitialSizes = activeSplit.paneSizes && activeSplit.paneSizes.length === count
      ? [...activeSplit.paneSizes]
      : Array(count).fill(100 / count);

    document.body.style.userSelect = 'none';
    document.body.style.cursor = isVertical ? 'row-resize' : 'col-resize';

    const onMove = (moveEvt: MouseEvent) => {
      if (!isDraggingSplitDivider || activeDividerIndex === null) return;
      if (dividerRafId) return;

      dividerRafId = requestAnimationFrame(() => {
        dividerRafId = null;
        if (!isDraggingSplitDivider || activeDividerIndex === null) return;

        const currentPos = isVertical ? moveEvt.clientY : moveEvt.clientX;
        const deltaPx = currentPos - dividerStartPos;
        const deltaPercent = (deltaPx / Math.max(1, dividerContainerSize)) * 100;

        const newSizes = [...dividerInitialSizes];
        const i = activeDividerIndex;
        const minPercent = 12; // Minimum size for any pane

        if (newSizes[i] + deltaPercent >= minPercent && newSizes[i + 1] - deltaPercent >= minPercent) {
          newSizes[i] += deltaPercent;
          newSizes[i + 1] -= deltaPercent;

          if (activeSplit) {
            terminalStore.setGroupPaneSizes(sessionId, activeSplit.id, newSizes);
          }
        }
      });
    };

    const onEnd = () => {
      isDraggingSplitDivider = false;
      activeDividerIndex = null;
      if (dividerRafId) {
        cancelAnimationFrame(dividerRafId);
        dividerRafId = null;
      }
      document.body.style.userSelect = '';
      document.body.style.cursor = '';
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onEnd);
      refitAllSplitTerminals();
    };

    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onEnd);
  }

  function terminalContainerAction(node: HTMLElement, termId: string) {
    termContainerMap.set(termId, node);
    initXterm(termId, node);

    // Attach ResizeObserver with debounced requestAnimationFrame to avoid jank
    const ro = new ResizeObserver(() => {
      scheduleXtermFit(termId);
    });
    ro.observe(node);
    resizeObservers.set(termId, ro);

    return {
      destroy() {
        const obs = resizeObservers.get(termId);
        if (obs) {
          obs.disconnect();
          resizeObservers.delete(termId);
        }
        if (pendingFitMap.has(termId)) {
          cancelAnimationFrame(pendingFitMap.get(termId)!);
          pendingFitMap.delete(termId);
        }
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

    // Fetch any prior terminal buffer history (e.g. initial shell prompt or pre-rendered output)
    if (window.go?.main?.App?.GetTerminalBuffer) {
      window.go.main.App.GetTerminalBuffer(termId)
        .then((buf: string) => {
          if (buf) {
            term.write(buf);
          }
        })
        .catch(() => {});
    }
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

  let isResizingPanel = $state(false);

  function handleResizeStart(e: MouseEvent) {
    e.preventDefault();
    isDragging = true;
    isResizingPanel = true;
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
    isResizingPanel = false;
    if (rafId) {
      cancelAnimationFrame(rafId);
      rafId = null;
    }

    document.body.style.userSelect = '';
    document.body.style.cursor = '';

    window.removeEventListener('mousemove', handleResizeMove);
    window.removeEventListener('mouseup', handleResizeEnd);
    refitAllSplitTerminals();
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
      // Ensure focus on active terminal
      if (terminalInstances.has(activeTermId)) {
        setTimeout(() => {
          terminalInstances.get(activeTermId)?.term.focus();
        }, 30);
      }
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
      <div class="h-8 w-full bg-ant-bg-secondary border-t border-ant-border-secondary dark:border-white/5 flex items-center justify-between px-3 text-xs text-ant-text-secondary select-none flex-shrink-0 z-20">
        <!-- Left: Horizontal Clickable Tab Chips -->
        <div class="flex items-center space-x-1.5 overflow-x-auto no-scrollbar flex-1 min-w-0 pr-2">
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
                class="flex items-center space-x-1.5 px-2.5 py-1 rounded-md text-[11px] font-mono transition border flex-shrink-0 {isActive ? 'bg-blue-500/10 text-blue-400 border-blue-500/30 shadow-sm font-medium' : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border-transparent'}"
              >
                <TerminalIcon size={11} class="{isActive ? 'text-blue-400' : 'text-ant-text-secondary'} flex-shrink-0" />
                <span class="truncate max-w-[140px] whitespace-nowrap">{tab.title}</span>
              </button>
            </Tooltip>
          {/each}

          <Tooltip title="New Terminal" placement="top">
            <button
              type="button"
              onclick={() => {
                terminalStore.createTerminal(sessionId, workspacePath);
                terminalStore.toggleSessionCollapse(sessionId, false);
                refitActiveTerminal();
              }}
              class="p-1 rounded text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition flex-shrink-0"
            >
              <Plus size={13} />
            </button>
          </Tooltip>
        </div>

        <!-- Right: Action Controls -->
        <div class="flex items-center space-x-1 flex-shrink-0">
          <Tooltip title="Dock to right" placement="top">
            <button
              type="button"
              onclick={toggleDock}
              class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
            >
              <PanelRight size={13} />
            </button>
          </Tooltip>
          <Tooltip title="Expand terminal panel" placement="top">
            <button
              type="button"
              onclick={() => {
                terminalStore.toggleSessionCollapse(sessionId, false);
                refitActiveTerminal();
              }}
              class="flex items-center space-x-1 px-2.5 py-1 text-[11px] rounded bg-ant-primary/10 text-ant-primary hover:bg-ant-primary/20 transition font-medium border border-ant-primary/20"
            >
              <ChevronUp size={13} />
              <span>Expand</span>
            </button>
          </Tooltip>
        </div>
      </div>
    {:else}
      <!-- Right Dock Slim Collapsed Bar with Individual Clickable Tab Pills -->
      <div class="w-9 h-full bg-ant-bg-secondary border-l border-ant-border-secondary dark:border-white/5 flex flex-col items-center justify-between py-2 text-xs text-ant-text-secondary select-none flex-shrink-0 z-20">
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
                class="flex flex-col items-center py-2 px-1 rounded-md transition w-full {isActive ? 'bg-blue-500/10 text-blue-400 font-medium border border-blue-500/30 shadow-sm' : 'text-ant-text-secondary hover:text-ant-text hover:bg-white/5 border border-transparent'}"
              >
                <TerminalIcon size={12} class="{isActive ? 'text-blue-400' : 'text-ant-text-secondary'} mb-1.5 flex-shrink-0" />
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
        <div class="flex flex-col items-center space-y-1.5 pt-2 border-t border-ant-border-secondary dark:border-white/5 w-full">
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
        class="flex flex-col w-full bg-ant-bg border-t border-ant-border-secondary dark:border-white/5 z-20 select-none font-serif flex-shrink-0"
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
        <div class="flex items-center justify-between px-3 h-8 bg-ant-bg-secondary border-b border-ant-border-secondary dark:border-white/5 flex-shrink-0 text-xs text-ant-text">
          <!-- Left: Terminal Sub-Tabs -->
          <div class="flex items-center space-x-1 overflow-x-auto no-scrollbar flex-1 min-w-0 pr-2">
            {#each activeTerminals as tab (tab.id)}
              {@const isTabActive = activeSplitGroup?.paneTermIds.includes(tab.id) || tab.id === activeTermId}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div
                draggable="true"
                ondragstart={(e) => handleTabDragStart(e, tab.id)}
                ondragend={handleTabDragEnd}
                class="flex items-center space-x-1.5 px-2.5 py-1 rounded-t-md text-xs transition-all cursor-grab active:cursor-grabbing border-t-2 flex-shrink-0 {isTabActive ? 'bg-ant-bg text-ant-text border-ant-primary font-medium shadow-sm' : 'border-transparent text-ant-text-secondary hover:text-ant-text hover:bg-white/5'}"
                onclick={() => terminalStore.switchTerminal(sessionId, tab.id)}
                ondblclick={(e) => startRename(tab, e)}
                title="Drag to terminal viewport to split"
              >
                <TerminalIcon size={12} class="{isTabActive ? 'text-ant-primary' : 'text-ant-text-secondary'} flex-shrink-0" />
                {#if editingTermId === tab.id}
                  <!-- svelte-ignore a11y_autofocus -->
                  <input
                    type="text"
                    use:selectOnFocus
                    bind:value={editingTitle}
                    class="px-1.5 py-0.5 text-[11px] font-mono rounded bg-ant-bg-tertiary text-ant-text border border-ant-primary focus:outline-none w-28"
                    onclick={(e) => {
                      e.stopPropagation();
                    }}
                    onfocus={(e) => {
                      e.currentTarget.select();
                    }}
                    onblur={() => saveRename(tab)}
                    onkeydown={(e) => handleRenameKeyDown(e, tab)}
                  />
                {:else}
                  <span class="font-mono text-[11.5px] whitespace-nowrap select-none" title="Double click to rename">{tab.title}</span>
                {/if}
                <Tooltip title="Kill process & close terminal tab" placement="top">
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
            <Tooltip title="Split Terminal (Ctrl+Shift+5)" placement="top">
              <button
                type="button"
                onclick={handleQuickSplit}
                class="flex items-center space-x-1 px-2 py-1 rounded-md bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-text hover:bg-ant-primary/10 hover:text-ant-primary text-[11px] font-medium transition outline-none border border-ant-border-secondary dark:border-white/5"
              >
                <Columns2 size={12} />
                <span>Split</span>
              </button>
            </Tooltip>

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
                  refitAllSplitTerminals();
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

        <!-- Terminal Split Viewport Containers (Bottom Dock) -->
        {#if true}
          {@const splitGroup = activeSplitGroup}
          {@const visibleTermIds = splitGroup?.paneTermIds || (activeTermId ? [activeTermId] : [])}
          {@const isMultiSplit = visibleTermIds.length > 1}
          <div
            class="flex-1 w-full relative overflow-hidden p-1 bg-ant-bg {isDragging || isResizingPanel || isDraggingSplitDivider ? 'select-none' : ''}"
          >
            <!-- Split Flex Container -->
            <div
              class="w-full h-full flex {splitGroup?.splitDirection === 'vertical' ? 'flex-col' : 'flex-row'} gap-0"
            >
              {#each visibleTermIds as termId, idx (termId)}
                {@const tabObj = activeTerminals.find((t) => t.id === termId)}
                {@const panePct = splitGroup?.paneSizes?.[idx] ?? (100 / visibleTermIds.length)}
                {@const isPaneFocused = focusedPaneTermId === termId}

                <!-- Resizable Divider Handle between panes -->
                {#if idx > 0}
                  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
                  <div
                    role="separator"
                    class="{splitGroup?.splitDirection === 'vertical' ? 'h-1.5 w-full cursor-row-resize' : 'w-1.5 h-full cursor-col-resize'} bg-transparent hover:bg-ant-primary/50 active:bg-ant-primary transition-colors flex items-center justify-center flex-shrink-0 z-30 group"
                    onmousedown={(e) => {
                      const parent = (e.currentTarget as HTMLElement).parentElement;
                      if (parent) {
                        handleSplitDividerStart(e, idx - 1, splitGroup?.splitDirection === 'vertical', parent);
                      }
                    }}
                  >
                    <div class="{splitGroup?.splitDirection === 'vertical' ? 'w-8 h-0.5' : 'w-0.5 h-8'} bg-ant-border-secondary dark:bg-white/10 group-hover:bg-ant-primary rounded-full transition-colors"></div>
                  </div>
                {/if}

                <!-- Single Split Pane Box -->
                <!-- svelte-ignore a11y_click_events_have_key_events -->
                <!-- svelte-ignore a11y_no_static_element_interactions -->
                <div
                  class="relative flex flex-col min-w-0 min-h-0 rounded-lg overflow-hidden border {isPaneFocused ? 'border-ant-primary shadow-sm shadow-ant-primary/10' : 'border-ant-border-secondary dark:border-white/5'}"
                  style="{splitGroup?.splitDirection === 'vertical' ? `height: ${panePct}%;` : `width: ${panePct}%;`}"
                  onclick={() => terminalStore.setFocusedPaneId(sessionId, termId)}
                  ondragover={(e) => handlePaneDragOver(e, termId)}
                  ondragleave={handlePaneDragLeave}
                  ondrop={(e) => handlePaneDrop(e, termId)}
                >
                  <!-- Mini Pane Header when in Multi-Split mode -->
                  {#if isMultiSplit && tabObj}
                    <div class="h-6 px-2 bg-ant-bg-secondary/80 border-b border-ant-border-secondary dark:border-white/5 flex items-center justify-between text-[10.5px] select-none flex-shrink-0">
                      <div class="flex items-center space-x-1 min-w-0">
                        <TerminalIcon size={10} class="{isPaneFocused ? 'text-ant-primary' : 'text-ant-text-muted'}" />
                        <span class="font-mono font-medium truncate {isPaneFocused ? 'text-ant-text' : 'text-ant-text-secondary'}">
                          {tabObj.title}
                        </span>
                      </div>

                      <div class="flex items-center space-x-1">
                        <Tooltip title="Unsplit (Move back to standalone tab)" placement="top">
                          <button
                            type="button"
                            onclick={(e) => {
                              e.stopPropagation();
                              terminalStore.unsplitTerminal(sessionId, termId);
                              refitAllSplitTerminals();
                            }}
                            class="p-0.5 text-ant-text-secondary hover:text-ant-primary hover:bg-white/10 rounded transition"
                          >
                            <ExternalLink size={10} />
                          </button>
                        </Tooltip>

                        <Tooltip title="Close split pane" placement="top">
                          <button
                            type="button"
                            onclick={(e) => {
                              e.stopPropagation();
                              terminalStore.closeTerminal(sessionId, termId);
                              refitAllSplitTerminals();
                            }}
                            class="p-0.5 text-ant-text-secondary hover:text-rose-400 hover:bg-white/10 rounded transition"
                          >
                            <X size={10} />
                          </button>
                        </Tooltip>
                      </div>
                    </div>
                  {/if}

                  <!-- Terminal Canvas Node -->
                  <div
                    use:terminalContainerAction={termId}
                    class="flex-1 w-full h-full p-1 bg-ant-bg"
                  ></div>

                  <!-- Visual Drop Indicator Overlay when dragging a tab over this pane -->
                  {#if activeDropZone && activeDropZone.targetTermId === termId}
                    <div
                      class="absolute pointer-events-none z-40 bg-ant-primary/20 border-2 border-dashed border-ant-primary backdrop-blur-[1px] flex items-center justify-center transition-all duration-150 animate-pulse {
                        activeDropZone.position === 'left' ? 'left-0 top-0 bottom-0 w-1/2 rounded-l-lg' :
                        activeDropZone.position === 'right' ? 'right-0 top-0 bottom-0 w-1/2 rounded-r-lg' :
                        activeDropZone.position === 'top' ? 'top-0 left-0 right-0 h-1/2 rounded-t-lg' :
                        'bottom-0 left-0 right-0 h-1/2 rounded-b-lg'
                      }"
                    >
                      <div class="px-2.5 py-1 rounded bg-ant-primary text-white text-[11px] font-serif font-medium shadow-md shadow-black/30">
                        Split {activeDropZone.position.toUpperCase()}
                      </div>
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {:else}
      <!-- Right Docked Panel -->
      <div
        class="flex flex-row h-full bg-ant-bg border-l border-ant-border-secondary dark:border-white/5 z-20 select-none font-serif flex-shrink-0"
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
          <div class="flex items-center justify-between px-2.5 h-8 bg-ant-bg-secondary border-b border-ant-border-secondary dark:border-white/5 flex-shrink-0 text-xs text-ant-text relative z-30">
            <!-- Left: Terminal Sub-Tabs -->
            <div class="flex items-center space-x-1 overflow-x-auto no-scrollbar flex-1 min-w-0 pr-1">
              {#each activeTerminals as tab (tab.id)}
                {@const isTabActive = activeSplitGroup?.paneTermIds.includes(tab.id) || tab.id === activeTermId}
                <!-- svelte-ignore a11y_click_events_have_key_events -->
                <!-- svelte-ignore a11y_no_static_element_interactions -->
                <div
                  draggable="true"
                  ondragstart={(e) => handleTabDragStart(e, tab.id)}
                  ondragend={handleTabDragEnd}
                  class="flex items-center space-x-1.5 px-2 py-1 rounded-t-md text-xs transition-all cursor-grab active:cursor-grabbing border-t-2 flex-shrink-0 {isTabActive ? 'bg-ant-bg text-ant-text border-ant-primary font-medium shadow-sm' : 'border-transparent text-ant-text-secondary hover:text-ant-text hover:bg-white/5'}"
                  onclick={() => terminalStore.switchTerminal(sessionId, tab.id)}
                  ondblclick={(e) => startRename(tab, e)}
                  title="Drag to terminal viewport to split"
                >
                  <TerminalIcon size={11} class="{isTabActive ? 'text-ant-primary' : 'text-ant-text-secondary'} flex-shrink-0" />
                  {#if editingTermId === tab.id}
                    <!-- svelte-ignore a11y_autofocus -->
                    <input
                      type="text"
                      use:selectOnFocus
                      bind:value={editingTitle}
                      class="px-1 py-0.5 text-[10px] font-mono rounded bg-ant-bg-tertiary text-ant-text border border-ant-primary focus:outline-none w-20"
                      onclick={(e) => {
                        e.stopPropagation();
                      }}
                      onfocus={(e) => {
                        e.currentTarget.select();
                      }}
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
            <Tooltip title="Split Terminal" placement="bottom">
              <button
                type="button"
                onclick={handleQuickSplit}
                class="p-1 text-ant-text-secondary hover:text-ant-text hover:bg-white/10 rounded transition"
              >
                <Columns2 size={11} />
              </button>
            </Tooltip>

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

          <!-- Terminal Split Viewport Containers (Right Dock) -->
          {#if true}
            {@const rightSplitGroup = activeSplitGroup}
            {@const visibleRightTermIds = rightSplitGroup?.paneTermIds || (activeTermId ? [activeTermId] : [])}
            {@const isRightMultiSplit = visibleRightTermIds.length > 1}
            <div
              class="flex-1 w-full relative overflow-hidden p-1 bg-ant-bg {isDragging || isResizingPanel || isDraggingSplitDivider ? 'select-none' : ''}"
            >
              <!-- Split Flex Container for Right Dock -->
              <div
                class="w-full h-full flex flex-col gap-0"
              >
                {#each visibleRightTermIds as termId, idx (termId)}
                  {@const tabObj = activeTerminals.find((t) => t.id === termId)}
                  {@const panePct = rightSplitGroup?.paneSizes?.[idx] ?? (100 / visibleRightTermIds.length)}
                  {@const isPaneFocused = focusedPaneTermId === termId}

                  <!-- Resizable Divider Handle between stacked panes in right dock -->
                  {#if idx > 0}
                    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
                    <div
                      role="separator"
                      class="h-1.5 w-full cursor-row-resize bg-transparent hover:bg-ant-primary/50 active:bg-ant-primary transition-colors flex items-center justify-center flex-shrink-0 z-30 group"
                      onmousedown={(e) => {
                        const parent = (e.currentTarget as HTMLElement).parentElement;
                        if (parent) {
                          handleSplitDividerStart(e, idx - 1, true, parent);
                        }
                      }}
                    >
                      <div class="w-8 h-0.5 bg-ant-border-secondary dark:bg-white/10 group-hover:bg-ant-primary rounded-full transition-colors"></div>
                    </div>
                  {/if}

                  <!-- Single Split Pane Box (Right Dock) -->
                  <!-- svelte-ignore a11y_click_events_have_key_events -->
                  <!-- svelte-ignore a11y_no_static_element_interactions -->
                  <div
                    class="relative flex flex-col min-w-0 min-h-0 rounded-lg overflow-hidden border {isPaneFocused ? 'border-ant-primary shadow-sm shadow-ant-primary/10' : 'border-ant-border-secondary dark:border-white/5'}"
                    style="height: {panePct}%;"
                    onclick={() => terminalStore.setFocusedPaneId(sessionId, termId)}
                    ondragover={(e) => handlePaneDragOver(e, termId)}
                    ondragleave={handlePaneDragLeave}
                    ondrop={(e) => handlePaneDrop(e, termId)}
                  >
                    <!-- Mini Pane Header when in Multi-Split mode -->
                    {#if isRightMultiSplit && tabObj}
                      <div class="h-5 px-1.5 bg-ant-bg-secondary/80 border-b border-ant-border-secondary dark:border-white/5 flex items-center justify-between text-[10px] select-none flex-shrink-0">
                        <div class="flex items-center space-x-1 min-w-0">
                          <TerminalIcon size={9} class="{isPaneFocused ? 'text-ant-primary' : 'text-ant-text-muted'}" />
                          <span class="font-mono font-medium truncate {isPaneFocused ? 'text-ant-text' : 'text-ant-text-secondary'}">
                            {tabObj.title}
                          </span>
                        </div>

                        <div class="flex items-center space-x-0.5">
                          <Tooltip title="Unsplit" placement="bottom">
                            <button
                              type="button"
                              onclick={(e) => {
                                e.stopPropagation();
                                terminalStore.unsplitTerminal(sessionId, termId);
                                refitAllSplitTerminals();
                              }}
                              class="p-0.5 text-ant-text-secondary hover:text-ant-primary hover:bg-white/10 rounded transition"
                            >
                              <ExternalLink size={9} />
                            </button>
                          </Tooltip>

                          <Tooltip title="Close split pane" placement="bottom">
                            <button
                              type="button"
                              onclick={(e) => {
                                e.stopPropagation();
                                terminalStore.closeTerminal(sessionId, termId);
                                refitAllSplitTerminals();
                              }}
                              class="p-0.5 text-ant-text-secondary hover:text-rose-400 hover:bg-white/10 rounded transition"
                            >
                              <X size={9} />
                            </button>
                          </Tooltip>
                        </div>
                      </div>
                    {/if}

                    <!-- Terminal Canvas Node -->
                    <div
                      use:terminalContainerAction={termId}
                      class="flex-1 w-full h-full p-0.5 bg-ant-bg"
                    ></div>

                    <!-- Visual Drop Indicator Overlay (Right Dock) -->
                    {#if activeDropZone && activeDropZone.targetTermId === termId}
                      <div
                        class="absolute pointer-events-none z-40 bg-ant-primary/20 border-2 border-dashed border-ant-primary backdrop-blur-[1px] flex items-center justify-center transition-all duration-150 animate-pulse {
                          activeDropZone.position === 'top' ? 'top-0 left-0 right-0 h-1/2 rounded-t-lg' :
                          'bottom-0 left-0 right-0 h-1/2 rounded-b-lg'
                        }"
                      >
                        <div class="px-2 py-0.5 rounded bg-ant-primary text-white text-[10px] font-serif font-medium shadow-md shadow-black/30">
                          Split {activeDropZone.position.toUpperCase()}
                        </div>
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      </div>
    {/if}
  {/if}
{/if}
