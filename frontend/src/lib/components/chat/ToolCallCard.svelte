<script lang="ts">
  import type { ToolCall } from '$lib/stores/session.svelte';
  import DiffCard from './DiffCard.svelte';
  import {
    Terminal,
    FileText,
    FileEdit,
    Search,
    ChevronDown,
    ChevronRight,
    Loader2,
    CheckCircle2,
    XCircle,
    Clock,
    Copy,
    Check,
    CornerDownRight,
    ArrowUpRight,
    Code2
  } from 'lucide-svelte';

  interface Props {
    toolCall: ToolCall;
  }

  let { toolCall }: Props = $props();

  let isExpanded = $state(true);
  let copiedInput = $state(false);
  let copiedOutput = $state(false);

  // Derive execution duration
  const durationMs = $derived.by(() => {
    if (!toolCall.startTime) return null;
    const end = toolCall.endTime || Date.now();
    return Math.max(0, end - toolCall.startTime);
  });

  // Extract tool category / icon / normalized name
  const toolMeta = $derived.by(() => {
    const t = (toolCall.tool || '').toLowerCase();
    if (t.includes('terminal') || t.includes('bash') || t.includes('exec') || t.includes('cmd')) {
      return {
        label: 'Terminal Command',
        name: 'run_terminal_cmd',
        type: 'terminal' as const,
        badgeClass: 'bg-indigo-500/15 text-indigo-400 border-indigo-500/30'
      };
    }
    if (t.includes('write') || t.includes('save') || t.includes('create_file')) {
      return {
        label: 'File Write',
        name: 'write',
        type: 'write' as const,
        badgeClass: 'bg-emerald-500/15 text-emerald-400 border-emerald-500/30'
      };
    }
    if (t.includes('search') || t.includes('replace') || t.includes('edit')) {
      return {
        label: 'Search & Replace',
        name: 'search_replace',
        type: 'search' as const,
        badgeClass: 'bg-amber-500/15 text-amber-400 border-amber-500/30'
      };
    }
    if (t.includes('read') || t.includes('cat') || t.includes('fetch')) {
      return {
        label: 'Read File',
        name: 'read_file',
        type: 'read' as const,
        badgeClass: 'bg-blue-500/15 text-blue-400 border-blue-500/30'
      };
    }
    return {
      label: 'Tool Execution',
      name: toolCall.tool || 'custom_tool',
      type: 'custom' as const,
      badgeClass: 'bg-ant-primary/15 text-ant-primary border-ant-primary/30'
    };
  });

  // Format parameters cleanly
  const formattedParams = $derived.by(() => {
    if (!toolCall.params) return '';
    if (typeof toolCall.params === 'string') return toolCall.params;
    return JSON.stringify(toolCall.params, null, 2);
  });

  // Check if params or result contain diff data or file content
  const hasDiffData = $derived(
    !!toolCall.diff ||
    (typeof toolCall.params === 'object' && toolCall.params !== null && 'diff' in toolCall.params)
  );

  async function copyInputText() {
    try {
      await navigator.clipboard.writeText(formattedParams);
      copiedInput = true;
      setTimeout(() => (copiedInput = false), 2000);
    } catch {
      // ignore
    }
  }

  async function copyOutputText() {
    if (!toolCall.result) return;
    try {
      await navigator.clipboard.writeText(toolCall.result);
      copiedOutput = true;
      setTimeout(() => (copiedOutput = false), 2000);
    } catch {
      // ignore
    }
  }
</script>

<div class="rounded-lg border border-ant-border bg-ant-bg-secondary overflow-hidden my-2.5 transition-all shadow-sm">
  <!-- Card Header Bar -->
  <div
    class="flex items-center justify-between px-3 py-2 bg-ant-bg-secondary hover:bg-ant-bg-tertiary border-b border-ant-border-secondary select-none cursor-pointer transition-colors"
    onclick={() => (isExpanded = !isExpanded)}
    role="button"
    tabindex="0"
    onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (isExpanded = !isExpanded)}
  >
    <div class="flex items-center space-x-2.5 min-w-0">
      <button
        type="button"
        class="text-ant-text-muted hover:text-white p-0.5 rounded transition"
        onclick={(e) => {
          e.stopPropagation();
          isExpanded = !isExpanded;
        }}
      >
        {#if isExpanded}
          <ChevronDown size={14} />
        {:else}
          <ChevronRight size={14} />
        {/if}
      </button>

      <!-- Tool Icon & Identifier -->
      <div class="flex items-center space-x-1.5 px-2 py-0.5 rounded text-[11px] font-mono border {toolMeta.badgeClass}">
        {#if toolMeta.type === 'terminal'}
          <Terminal size={12} />
        {:else if toolMeta.type === 'write'}
          <FileEdit size={12} />
        {:else if toolMeta.type === 'search'}
          <Search size={12} />
        {:else if toolMeta.type === 'read'}
          <FileText size={12} />
        {:else}
          <Code2 size={12} />
        {/if}
        <span class="font-semibold">{toolCall.tool}</span>
      </div>

      <!-- Quick summary if available in params -->
      {#if typeof toolCall.params === 'object' && toolCall.params !== null}
        {@const p = toolCall.params as Record<string, unknown>}
        <span class="text-xs text-ant-text-muted font-mono truncate max-w-sm hidden sm:inline">
          {p.command || p.path || p.file_path || p.target_file || p.query || ''}
        </span>
      {/if}
    </div>

    <!-- Right Badges: Status, Elapsed Time -->
    <div class="flex items-center space-x-3 flex-shrink-0">
      {#if durationMs !== null}
        <div class="flex items-center text-[10px] font-mono text-ant-text-muted">
          <Clock size={11} class="mr-1" />
          <span>{durationMs}ms</span>
        </div>
      {/if}

      <!-- Status Badge -->
      {#if toolCall.status === 'running'}
        <span class="flex items-center text-[10px] font-medium text-ant-primary bg-ant-primary/10 border border-ant-primary/30 px-2 py-0.5 rounded-full animate-pulse">
          <Loader2 size={11} class="animate-spin mr-1" />
          <span>Running</span>
        </span>
      {:else if toolCall.status === 'completed'}
        <span class="flex items-center text-[10px] font-medium text-ant-success bg-ant-success/10 border border-ant-success/30 px-2 py-0.5 rounded-full">
          <CheckCircle2 size={11} class="mr-1" />
          <span>Success</span>
        </span>
      {:else if toolCall.status === 'error'}
        <span class="flex items-center text-[10px] font-medium text-ant-error bg-ant-error/10 border border-ant-error/30 px-2 py-0.5 rounded-full">
          <XCircle size={11} class="mr-1" />
          <span>Failed</span>
        </span>
      {:else}
        <span class="flex items-center text-[10px] font-medium text-ant-text-muted bg-ant-bg border border-ant-border px-2 py-0.5 rounded-full">
          <span>Pending</span>
        </span>
      {/if}
    </div>
  </div>

  <!-- Collapsible IN / OUT Details -->
  {#if isExpanded}
    <div class="p-3 bg-ant-bg space-y-3 select-text border-t border-ant-border-secondary">
      <!-- Input Execution Block -->
      {#if formattedParams}
        <div class="rounded-md border border-ant-border-secondary bg-ant-bg-secondary overflow-hidden">
          <div class="flex items-center justify-between px-2.5 py-1.5 bg-ant-bg-tertiary border-b border-ant-border-secondary text-[11px] select-none">
            <div class="flex items-center space-x-1.5 text-ant-text-secondary font-mono">
              <CornerDownRight size={12} class="text-ant-primary" />
              <span class="font-semibold uppercase text-[10px] tracking-wider text-ant-primary">INPUT / PARAMS</span>
            </div>
            <button
              type="button"
              onclick={copyInputText}
              class="flex items-center space-x-1 text-[10px] text-ant-text-muted hover:text-white transition px-1.5 py-0.5 rounded hover:bg-ant-bg"
              title="Copy input params"
            >
              {#if copiedInput}
                <Check size={11} class="text-ant-success" />
                <span class="text-ant-success">Copied</span>
              {:else}
                <Copy size={11} />
                <span>Copy</span>
              {/if}
            </button>
          </div>
          <pre class="p-2.5 text-[11px] font-mono text-ant-text leading-relaxed overflow-x-auto max-h-56 scrollbar-thin"><code>{formattedParams}</code></pre>
        </div>
      {/if}

      <!-- Output / Result Block -->
      {#if toolCall.result !== undefined}
        <div class="rounded-md border border-ant-border-secondary bg-ant-bg-secondary overflow-hidden">
          <div class="flex items-center justify-between px-2.5 py-1.5 bg-ant-bg-tertiary border-b border-ant-border-secondary text-[11px] select-none">
            <div class="flex items-center space-x-1.5 text-ant-text-secondary font-mono">
              <ArrowUpRight size={12} class={toolCall.status === 'error' ? 'text-ant-error' : 'text-ant-success'} />
              <span class="font-semibold uppercase text-[10px] tracking-wider {toolCall.status === 'error' ? 'text-ant-error' : 'text-ant-success'}">
                {toolCall.status === 'error' ? 'ERROR OUTPUT' : 'OUTPUT / RESULT'}
              </span>
            </div>
            <button
              type="button"
              onclick={copyOutputText}
              class="flex items-center space-x-1 text-[10px] text-ant-text-muted hover:text-white transition px-1.5 py-0.5 rounded hover:bg-ant-bg"
              title="Copy output"
            >
              {#if copiedOutput}
                <Check size={11} class="text-ant-success" />
                <span class="text-ant-success">Copied</span>
              {:else}
                <Copy size={11} />
                <span>Copy</span>
              {/if}
            </button>
          </div>
          <pre class="p-2.5 text-[11px] font-mono leading-relaxed overflow-x-auto max-h-72 scrollbar-thin {toolCall.status === 'error' ? 'text-rose-300' : 'text-ant-text'}"><code>{toolCall.result || '(empty output)'}</code></pre>
        </div>
      {/if}

      <!-- Optional Embedded Diff Inspector -->
      {#if toolCall.diff}
        <div class="pt-1">
          <DiffCard diff={toolCall.diff} />
        </div>
      {/if}
    </div>
  {/if}
</div>
