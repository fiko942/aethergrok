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
    Sparkles
  } from 'lucide-svelte';
  import { terminalStore, type TerminalTab } from '$lib/stores/terminal.svelte';
  import { sessionStore } from '$lib/stores/session.svelte';

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
  let startY = 0;
  let startHeight = 0;

  // Map of termId -> { term: Xterm, fitAddon: FitAddon, container: HTMLElement, unsub: () => void }
  const terminalInstances = new Map<string, { term: Xterm; fitAddon: FitAddon; unsub?: () => void }>();
  let termContainerMap = new Map<string, HTMLElement>();

  // Auto create initial terminal if none exists and panel is opened
  $effect(() => {
    if (terminalStore.isOpen && activeTerminals.length === 0 && sessionId) {
      terminalStore.createTerminal(sessionId, workspacePath);
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
      theme: {
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
      },
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

  function handleResizeStart(e: MouseEvent) {
    isDragging = true;
    startY = e.clientY;
    startHeight = terminalStore.panelHeight;
    window.addEventListener('mousemove', handleResizeMove);
    window.addEventListener('mouseup', handleResizeEnd);
  }

  function handleResizeMove(e: MouseEvent) {
    if (!isDragging) return;
    const delta = startY - e.clientY;
    terminalStore.setPanelHeight(startHeight + delta);
    refitActiveTerminal();
  }

  function handleResizeEnd() {
    isDragging = false;
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

{#if terminalStore.isOpen}
  <div
    class="flex flex-col w-full bg-[#0e0e11] border-t border-white/10 z-20 select-none font-serif transition-all duration-75 flex-shrink-0"
    style="height: {isMaximized ? 'calc(100vh - 120px)' : `${terminalStore.panelHeight}px`}; min-height: 140px;"
  >
    <!-- 4px Drag Resize Handle Divider -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
      role="separator"
      class="h-1.5 w-full cursor-row-resize bg-transparent hover:bg-ant-primary/40 active:bg-ant-primary transition-colors flex items-center justify-center -mt-0.5 z-30"
      onmousedown={handleResizeStart}
    >
      <div class="w-10 h-0.5 bg-white/20 rounded-full"></div>
    </div>

    <!-- Terminal Header & Multi-Tab Bar -->
    <div class="flex items-center justify-between px-3 h-8 bg-[#141418] border-b border-white/5 flex-shrink-0 text-xs text-zinc-300">
      <!-- Left: Terminal Sub-Tabs -->
      <div class="flex items-center space-x-1 overflow-x-auto no-scrollbar flex-1 min-w-0 pr-2">
        {#each activeTerminals as tab (tab.id)}
          {@const isActive = tab.id === activeTermId}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="flex items-center space-x-1.5 px-2.5 py-1 rounded-t-md text-xs transition-colors cursor-pointer border-t-2 {isActive ? 'bg-[#0e0e11] text-zinc-100 border-ant-primary font-medium shadow-sm' : 'border-transparent text-zinc-400 hover:text-zinc-200 hover:bg-white/5'}"
            onclick={() => terminalStore.switchTerminal(sessionId, tab.id)}
          >
            <TerminalIcon size={12} class={isActive ? 'text-ant-primary' : 'text-zinc-500'} />
            <span class="truncate max-w-[120px] font-mono text-[11px]">{tab.title}</span>
            <button
              type="button"
              onclick={(e) => {
                e.stopPropagation();
                terminalStore.closeTerminal(sessionId, tab.id);
              }}
              class="text-zinc-500 hover:text-rose-400 p-0.5 rounded transition hover:bg-white/10"
              title="Kill process & close terminal tab"
            >
              <X size={11} />
            </button>
          </div>
        {/each}

        <!-- Add New Terminal Sub-Tab Button -->
        <button
          type="button"
          onclick={handleAddTerminal}
          class="p-1 rounded text-zinc-400 hover:text-zinc-100 hover:bg-white/10 transition"
          title="New Terminal"
        >
          <Plus size={13} />
        </button>
      </div>

      <!-- Right: Action Buttons (Send to Agent, Clear, Maximize, Close) -->
      <div class="flex items-center space-x-1.5 flex-shrink-0">
        <button
          type="button"
          onclick={handleSendToAgent}
          class="flex items-center space-x-1 px-2.5 py-1 rounded-md bg-ant-primary/10 text-ant-primary hover:bg-ant-primary/20 text-[11px] font-medium transition outline-none"
          title="Attach recent terminal output to agent composer"
        >
          <Send size={11} />
          <span>Send to Agent</span>
        </button>

        <button
          type="button"
          onclick={handleClearTerminal}
          class="p-1 text-zinc-400 hover:text-zinc-200 hover:bg-white/10 rounded transition"
          title="Clear terminal buffer"
        >
          <Trash2 size={12} />
        </button>

        <button
          type="button"
          onclick={() => {
            isMaximized = !isMaximized;
            refitActiveTerminal();
          }}
          class="p-1 text-zinc-400 hover:text-zinc-200 hover:bg-white/10 rounded transition"
          title={isMaximized ? 'Restore size' : 'Maximize terminal'}
        >
          {#if isMaximized}
            <Minimize2 size={12} />
          {:else}
            <Maximize2 size={12} />
          {/if}
        </button>

        <button
          type="button"
          onclick={() => terminalStore.toggleOpen(false)}
          class="p-1 text-zinc-400 hover:text-zinc-200 hover:bg-white/10 rounded transition"
          title="Hide terminal panel"
        >
          <ChevronDown size={14} />
        </button>
      </div>
    </div>

    <!-- Terminal Canvas Containers (Preserves instance in DOM) -->
    <div class="flex-1 w-full relative overflow-hidden p-1.5 bg-[#0e0e11]">
      {#each activeTerminals as tab (tab.id)}
        <div
          use:terminalContainerAction={tab.id}
          class="w-full h-full {tab.id === activeTermId ? 'block' : 'hidden'}"
        ></div>
      {/each}
    </div>
  </div>
{/if}
