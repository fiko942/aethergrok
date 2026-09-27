<script lang="ts">
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
    FileText
  } from 'lucide-svelte';
  import Button from '$lib/antd/Button.svelte';
  import Tooltip from '$lib/antd/Tooltip.svelte';

  let searchQuery = $state('');
  let selectedLevel = $state<LogLevel | 'ALL'>('ALL');
  let selectedCategory = $state<LogCategory | 'ALL'>('ALL');
  let copyNotice = $state(false);
  let logContainer = $state<HTMLDivElement | null>(null);

  const levels: Array<LogLevel | 'ALL'> = ['ALL', 'INFO', 'WARN', 'ERROR', 'DEBUG'];
  const categories: Array<LogCategory | 'ALL'> = ['ALL', 'UI', 'SESSION', 'TERMINAL', 'VOICE', 'BACKEND', 'SETTINGS', 'SYSTEM'];

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
  <!-- Top Action & Filter Bar -->
  <div class="flex flex-wrap items-center justify-between gap-2.5 bg-ant-bg-tertiary/40 p-2.5 rounded-xl border border-ant-border/60">
    <!-- Search Input -->
    <div class="relative flex-1 min-w-[180px]">
      <Search size={13} class="absolute left-2.5 top-1/2 -translate-y-1/2 text-ant-text-muted" />
      <input
        type="text"
        placeholder="Filter logs or errors..."
        bind:value={searchQuery}
        class="w-full bg-ant-bg text-ant-text pl-8 pr-2.5 py-1.5 rounded-lg text-xs border border-ant-border outline-none focus:border-ant-primary transition"
      />
    </div>

    <!-- Level Filter Pills -->
    <div class="flex items-center space-x-1 bg-ant-bg p-0.5 rounded-lg border border-ant-border text-[11px] font-mono">
      {#each levels as lvl}
        <button
          type="button"
          onclick={() => selectedLevel = lvl}
          class="px-2 py-1 rounded-md transition font-medium {selectedLevel === lvl
            ? 'bg-ant-primary/15 text-ant-primary font-semibold'
            : 'text-ant-text-secondary hover:text-ant-text hover:bg-white/5'}"
        >
          {lvl}
        </button>
      {/each}
    </div>

    <!-- Category Filter Dropdown -->
    <div class="flex items-center space-x-1 bg-ant-bg p-0.5 rounded-lg border border-ant-border text-[11px]">
      <select
        bind:value={selectedCategory}
        aria-label="Filter by Category"
        class="bg-transparent text-ant-text text-xs px-2 py-1 outline-none font-mono cursor-pointer"
      >
        {#each categories as cat}
          <option value={cat} class="bg-ant-bg text-ant-text font-sans">{cat}</option>
        {/each}
      </select>
    </div>

    <!-- Actions (Export, Copy, Clear) -->
    <div class="flex items-center space-x-1.5">
      <Tooltip title="Copy logs to clipboard" placement="top">
        <button
          type="button"
          onclick={handleCopy}
          class="flex items-center space-x-1 px-2.5 py-1.5 rounded-lg bg-ant-bg hover:bg-ant-bg-secondary border border-ant-border text-xs text-ant-text transition"
        >
          {#if copyNotice}
            <Check size={13} class="text-emerald-400" />
            <span class="text-emerald-400 font-mono text-[11px]">Copied</span>
          {:else}
            <Copy size={13} class="text-ant-text-secondary" />
            <span class="text-[11px]">Copy</span>
          {/if}
        </button>
      </Tooltip>

      <Tooltip title="Export JSON" placement="top">
        <button
          type="button"
          onclick={() => logger.exportAsJSON()}
          class="flex items-center space-x-1 px-2.5 py-1.5 rounded-lg bg-ant-bg hover:bg-ant-bg-secondary border border-ant-border text-xs text-ant-text transition"
        >
          <FileJson size={13} class="text-ant-primary" />
          <span class="text-[11px]">JSON</span>
        </button>
      </Tooltip>

      <Tooltip title="Export .log file" placement="top">
        <button
          type="button"
          onclick={() => logger.exportAsText()}
          class="flex items-center space-x-1 px-2.5 py-1.5 rounded-lg bg-ant-bg hover:bg-ant-bg-secondary border border-ant-border text-xs text-ant-text transition"
        >
          <FileText size={13} class="text-ant-primary" />
          <span class="text-[11px]">Log</span>
        </button>
      </Tooltip>

      <Tooltip title="Clear in-memory buffer" placement="top">
        <button
          type="button"
          onclick={() => logger.clear()}
          class="p-1.5 rounded-lg bg-ant-bg hover:bg-rose-500/10 border border-ant-border text-ant-text-muted hover:text-rose-400 transition"
        >
          <Trash2 size={13} />
        </button>
      </Tooltip>
    </div>
  </div>

  <!-- Terminal Log Viewer Output Window -->
  <div
    bind:this={logContainer}
    class="flex-1 min-h-[320px] max-h-[440px] bg-ant-bg border border-ant-border rounded-xl p-3 font-mono text-xs overflow-y-auto space-y-1.5 select-text shadow-inner"
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
            log.level === 'ERROR' ? 'bg-rose-500/15 text-rose-400 border border-rose-500/30' :
            log.level === 'WARN' ? 'bg-amber-500/15 text-amber-400 border border-amber-500/30' :
            log.level === 'INFO' ? 'bg-indigo-500/15 text-indigo-400 border border-indigo-500/30' :
            'bg-zinc-500/15 text-zinc-400 border border-zinc-500/30'
          }">
            {log.level}
          </span>

          <!-- Category Badge -->
          <span class="px-1.5 py-0.2 rounded text-[10px] text-ant-text-secondary bg-ant-bg-tertiary border border-ant-border flex-shrink-0">
            {log.category}
          </span>

          <!-- Message and Details -->
          <div class="flex-1 min-w-0 break-words text-[11.5px] {
            log.level === 'ERROR' ? 'text-rose-300' :
            log.level === 'WARN' ? 'text-amber-200' :
            log.level === 'INFO' ? 'text-ant-text' :
            'text-ant-text-secondary'
          }">
            <span>{log.message}</span>
            {#if log.details}
              <div class="mt-1 p-1.5 rounded bg-black/30 border border-white/5 text-[10.5px] text-ant-text-secondary font-mono overflow-x-auto">
                <pre class="whitespace-pre-wrap">{typeof log.details === 'object' ? JSON.stringify(log.details, null, 2) : String(log.details)}</pre>
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
    <label class="flex items-center space-x-1.5 cursor-pointer text-ant-text select-none">
      <input
        type="checkbox"
        bind:checked={logger.autoScroll}
        class="accent-ant-primary rounded cursor-pointer"
      />
      <span>Auto-scroll to latest</span>
    </label>
  </div>
</div>
