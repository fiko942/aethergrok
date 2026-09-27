<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { logger, type LogLevel, type LogCategory, type LogEntry } from '$lib/stores/logger.svelte';
  import {
    Search,
    Download,
    Copy,
    Trash2,
    Check,
    AlertCircle,
    Info,
    AlertTriangle,
    Bug,
    Terminal,
    Filter,
    ScrollText,
    ArrowDown,
    FileJson,
    FileText,
    ChevronDown,
    Layers,
    Cpu,
    Mic,
    Sliders,
    Monitor
  } from 'lucide-svelte';
  import Button from '$lib/antd/Button.svelte';
  import Tooltip from '$lib/antd/Tooltip.svelte';
  import CustomCheckbox from '$lib/components/ui/CustomCheckbox.svelte';

  let searchQuery = $state('');
  let selectedLevel = $state<LogLevel | 'ALL'>('ALL');
  let selectedCategory = $state<LogCategory | 'ALL'>('ALL');
  let copyNotice = $state(false);
  let exportNotice = $state<string | null>(null);
  let logContainer = $state<HTMLDivElement | null>(null);

  // Custom Category Dropdown State
  let categoryDropdownOpen = $state(false);
  let categoryDropdownRef = $state<HTMLDivElement | null>(null);

  const levels: Array<LogLevel | 'ALL'> = ['ALL', 'INFO', 'WARN', 'ERROR', 'DEBUG'];
  const categories: Array<{ id: LogCategory | 'ALL'; label: string; icon: any }> = [
    { id: 'ALL', label: 'All Categories', icon: Filter },
    { id: 'SYSTEM', label: 'System', icon: Monitor },
    { id: 'SESSION', label: 'Session', icon: ScrollText },
    { id: 'TERMINAL', label: 'Terminal', icon: Terminal },
    { id: 'VOICE', label: 'Voice & Mic', icon: Mic },
    { id: 'BACKEND', label: 'Backend IPC', icon: Cpu },
    { id: 'UI', label: 'Interface', icon: Layers },
    { id: 'SETTINGS', label: 'Settings', icon: Sliders }
  ];

  let activeCategoryObj = $derived(categories.find((c) => c.id === selectedCategory) || categories[0]);
  let ActiveCategoryIcon = $derived(activeCategoryObj.icon);

  let filteredLogs = $derived.by(() => {
    let list = logger.entries;
    if (selectedLevel !== 'ALL') {
      list = list.filter((l) => l.level === selectedLevel);
    }
    if (selectedCategory !== 'ALL') {
      list = list.filter((l) => l.category === selectedCategory);
    }
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase().trim();
      list = list.filter((l) => 
        l.message.toLowerCase().includes(q) ||
        l.category.toLowerCase().includes(q) ||
        (l.details && JSON.stringify(l.details).toLowerCase().includes(q))
      );
    }
    return list;
  });

  async function handleCopy() {
    const success = await logger.copyToClipboard();
    if (success) {
      copyNotice = true;
      setTimeout(() => {
        copyNotice = false;
      }, 2000);
    }
  }

  async function handleExportJSON() {
    const result = await logger.exportAsJSON();
    if (result) {
      exportNotice = 'JSON Export Saved';
      setTimeout(() => exportNotice = null, 2500);
    }
  }

  async function handleExportText() {
    const result = await logger.exportAsText();
    if (result) {
      exportNotice = 'Log Export Saved';
      setTimeout(() => exportNotice = null, 2500);
    }
  }

  function handleClickOutside(e: MouseEvent) {
    if (categoryDropdownRef && !categoryDropdownRef.contains(e.target as Node)) {
      categoryDropdownOpen = false;
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape' && categoryDropdownOpen) {
      categoryDropdownOpen = false;
    }
  }

  onMount(() => {
    window.addEventListener('click', handleClickOutside);
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('click', handleClickOutside);
      window.removeEventListener('keydown', handleKeyDown);
    }
  });

  function formatTime(ts: number): string {
    const d = new Date(ts);
    return d.toTimeString().split(' ')[0] + '.' + String(d.getMilliseconds()).padStart(3, '0');
  }

  $effect(() => {
    if (logger.autoScroll && logContainer && filteredLogs.length) {
      logContainer.scrollTop = logContainer.scrollHeight;
    }
  });
</script>

<div class="flex flex-col h-full space-y-3 font-serif">
  <!-- Top Unified Action & Filter Bar (Single Row) -->
  <div class="flex items-center gap-2 bg-ant-bg-tertiary/40 px-3 py-2 rounded-xl border border-ant-border-secondary dark:border-white/5 w-full">
    <!-- Search Input -->
    <div class="relative flex-1 min-w-[140px]">
      <Search size={13} class="absolute left-2.5 top-1/2 -translate-y-1/2 text-ant-text-muted pointer-events-none" />
      <input
        type="text"
        placeholder="Filter logs or errors..."
        bind:value={searchQuery}
        class="w-full bg-ant-bg text-ant-text pl-8 pr-2.5 py-1.5 rounded-lg text-xs border border-ant-border-secondary dark:border-white/5 outline-none focus:border-ant-primary transition"
      />
    </div>

    <!-- Level Filter Pills -->
    <div class="flex items-center space-x-0.5 bg-ant-bg p-0.5 rounded-lg border border-ant-border-secondary dark:border-white/5 text-[10.5px] font-mono shrink-0">
      {#each levels as lvl}
        <button
          type="button"
          onclick={() => selectedLevel = lvl}
          class="px-1.5 py-1 rounded-md transition font-medium {selectedLevel === lvl
            ? 'bg-ant-primary/15 text-ant-primary font-semibold'
            : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary'}"
        >
          {lvl}
        </button>
      {/each}
    </div>

    <!-- Custom Category Filter Dropdown -->
    <div class="relative shrink-0" bind:this={categoryDropdownRef}>
      <button
        type="button"
        onclick={() => categoryDropdownOpen = !categoryDropdownOpen}
        class="flex items-center space-x-1.5 px-2 py-1.5 bg-ant-bg hover:bg-ant-bg-secondary rounded-lg border border-ant-border-secondary dark:border-white/5 text-xs text-ant-text transition cursor-pointer {categoryDropdownOpen ? 'ring-1 ring-ant-primary/30 border-white/10' : ''}"
        title="Filter by Subsystem Category"
      >
        <ActiveCategoryIcon size={12} class="text-ant-primary shrink-0" />
        <span class="font-sans text-[11px] font-medium max-w-[70px] truncate">{activeCategoryObj.label}</span>
        <ChevronDown size={11} class="text-ant-text-muted transition-transform duration-150 shrink-0 {categoryDropdownOpen ? 'rotate-180 text-ant-primary' : ''}" />
      </button>

      {#if categoryDropdownOpen}
        <div
          class="absolute top-full right-0 mt-1.5 w-44 bg-ant-bg-elevated dark:bg-[#18181c] border border-ant-border-secondary dark:border-white/10 rounded-xl shadow-2xl z-50 overflow-hidden flex flex-col p-1 space-y-0.5 backdrop-blur-xl animate-in fade-in zoom-in-95 duration-100"
        >
          <div class="px-2.5 py-1 text-[10px] font-semibold text-ant-text-muted uppercase tracking-wider border-b border-ant-border-secondary dark:border-white/5 mb-0.5">
            Log Categories
          </div>
          {#each categories as cat}
            {@const isSelected = selectedCategory === cat.id}
            {@const CatIcon = cat.icon}
            <button
              type="button"
              class="w-full flex items-center justify-between px-2.5 py-1.5 rounded-lg text-xs transition-colors {isSelected ? 'bg-ant-primary/15 text-ant-primary font-medium' : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-secondary dark:hover:bg-white/5'}"
              onclick={() => {
                selectedCategory = cat.id;
                categoryDropdownOpen = false;
              }}
            >
              <div class="flex items-center space-x-2 min-w-0">
                <CatIcon size={12} class={isSelected ? 'text-ant-primary' : 'text-ant-text-muted'} />
                <span class="truncate font-sans text-[11px]">{cat.label}</span>
              </div>
              {#if isSelected}
                <Check size={12} class="text-ant-primary shrink-0" />
              {/if}
            </button>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Vertical Divider -->
    <div class="h-4 w-[1px] bg-ant-border-secondary dark:bg-white/10 shrink-0"></div>

    <!-- Actions (Export, Copy, Clear) -->
    <div class="flex items-center space-x-1 shrink-0">
      <Tooltip title="Copy logs to clipboard" placement="top">
        <button
          type="button"
          onclick={handleCopy}
          class="flex items-center space-x-1 px-2 py-1.5 rounded-lg bg-ant-bg hover:bg-ant-bg-secondary border border-ant-border-secondary dark:border-white/5 text-xs text-ant-text transition shrink-0"
        >
          {#if copyNotice}
            <Check size={12} class="text-emerald-400" />
            <span class="text-emerald-400 font-mono text-[11px]">Copied</span>
          {:else}
            <Copy size={12} class="text-ant-text-secondary" />
            <span class="text-[11px]">Copy</span>
          {/if}
        </button>
      </Tooltip>

      <Tooltip title="Export JSON to file" placement="top">
        <button
          type="button"
          onclick={handleExportJSON}
          class="flex items-center space-x-1 px-2 py-1.5 rounded-lg bg-ant-bg hover:bg-ant-bg-secondary border border-ant-border-secondary dark:border-white/5 text-xs text-ant-text transition shrink-0"
        >
          <FileJson size={12} class="text-ant-primary" />
          <span class="text-[11px]">JSON</span>
        </button>
      </Tooltip>

      <Tooltip title="Export .log file" placement="top">
        <button
          type="button"
          onclick={handleExportText}
          class="flex items-center space-x-1 px-2 py-1.5 rounded-lg bg-ant-bg hover:bg-ant-bg-secondary border border-ant-border-secondary dark:border-white/5 text-xs text-ant-text transition shrink-0"
        >
          <FileText size={12} class="text-ant-primary" />
          <span class="text-[11px]">Log</span>
        </button>
      </Tooltip>

      <Tooltip title="Clear persistent log files and buffer" placement="top">
        <button
          type="button"
          onclick={() => logger.clear()}
          class="p-1.5 rounded-lg bg-ant-bg hover:bg-rose-500/10 border border-ant-border-secondary dark:border-white/5 text-ant-text-muted hover:text-rose-400 transition shrink-0"
        >
          <Trash2 size={12} />
        </button>
      </Tooltip>
    </div>
  </div>

  {#if exportNotice}
    <div class="px-3 py-1.5 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs font-serif flex items-center justify-between animate-in fade-in duration-150">
      <div class="flex items-center space-x-2">
        <Check size={13} class="text-emerald-400" />
        <span>{exportNotice}</span>
      </div>
      <span class="text-[10px] text-emerald-400/80 font-mono">Saved</span>
    </div>
  {/if}

  <!-- Terminal Log Viewer Output Window -->
  <div
    bind:this={logContainer}
    class="flex-1 min-h-[320px] max-h-[440px] bg-ant-bg border border-ant-border-secondary dark:border-white/5 rounded-xl p-3 font-mono text-xs overflow-y-auto space-y-1.5 select-text shadow-inner"
  >
    {#if filteredLogs.length === 0}
      <div class="h-full min-h-[260px] flex flex-col items-center justify-center text-center text-ant-text-muted space-y-2 select-none">
        <ScrollText size={28} class="opacity-40" />
        <span class="text-xs">No log entries found matching criteria</span>
      </div>
    {:else}
      {#each filteredLogs as log (log.id)}
        <div class="flex items-start space-x-2 py-0.5 hover:bg-white/[0.03] px-1 rounded transition group leading-relaxed">
          <!-- Time -->
          <span class="text-ant-text-muted text-[10.5px] whitespace-nowrap flex-shrink-0 pt-0.5">
            {formatTime(log.timestamp)}
          </span>

          <!-- Level Badge -->
          <span class="px-1.5 py-0.2 rounded text-[10px] font-bold flex-shrink-0 uppercase {
            log.level === 'ERROR' ? 'bg-rose-500/15 text-rose-700 dark:text-rose-400 border border-rose-500/30' :
            log.level === 'WARN' ? 'bg-amber-500/15 text-amber-800 dark:text-amber-300 border border-amber-500/30' :
            log.level === 'INFO' ? 'bg-blue-500/15 text-blue-700 dark:text-blue-400 border border-blue-500/30' :
            'bg-zinc-500/15 text-zinc-700 dark:text-zinc-400 border border-zinc-500/30'
          }">
            {log.level}
          </span>

          <!-- Category Badge -->
          <span class="px-1.5 py-0.2 rounded text-[10px] font-medium text-ant-text-secondary bg-ant-bg-tertiary border border-ant-border-secondary dark:border-white/5 flex-shrink-0">
            {log.category}
          </span>

          <!-- Message and Details -->
          <div class="flex-1 min-w-0 break-words text-[11.5px] {
            log.level === 'ERROR' ? 'text-rose-700 dark:text-rose-300' :
            log.level === 'WARN' ? 'text-amber-800 dark:text-amber-200' :
            log.level === 'INFO' ? 'text-ant-text' :
            'text-ant-text-secondary'
          }">
            <span>{log.message}</span>
            {#if log.details}
              <div class="mt-1.5 p-2.5 rounded-lg bg-zinc-100 dark:bg-black/50 border border-zinc-200 dark:border-white/10 text-[11px] text-zinc-900 dark:text-zinc-200 font-mono overflow-x-auto shadow-xs">
                <pre class="whitespace-pre-wrap leading-relaxed text-zinc-900 dark:text-zinc-200 font-medium">{typeof log.details === 'object' ? JSON.stringify(log.details, null, 2) : String(log.details)}</pre>
              </div>
            {/if}
          </div>
        </div>
      {/each}
    {/if}
  </div>

  <!-- Bottom Log Stats & Auto-Scroll Toggle -->
  <div class="flex items-center justify-between text-[11px] text-ant-text-secondary px-1 font-mono">
    <div class="flex items-center space-x-3">
      <span>Total: <strong>{logger.entries.length}</strong> / {logger.maxEntries} events</span>
      <span>Filtered: <strong>{filteredLogs.length}</strong></span>
    </div>
    <button
      type="button"
      onclick={() => logger.autoScroll = !logger.autoScroll}
      class="flex items-center space-x-2 cursor-pointer text-ant-text select-none group"
    >
      <CustomCheckbox
        checked={logger.autoScroll}
        size="sm"
      />
      <span class="group-hover:text-ant-primary transition-colors text-xs font-sans">Auto-scroll to latest</span>
    </button>
  </div>
</div>
