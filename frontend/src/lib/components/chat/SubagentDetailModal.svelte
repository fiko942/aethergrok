<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import type { ToolCall } from '$lib/stores/session.svelte';
  import {
    type SubagentParsedMeta,
    cleanSubagentOutput,
    formatSubagentDuration
  } from '$lib/utils/toolUtils';
  import { renderMarkdown } from '$lib/utils/markdownRenderer';
  import {
    Bot,
    X,
    Clock,
    Copy,
    Check,
    Loader2,
    CheckCircle2,
    XCircle,
    Cpu,
    Shield,
    Code,
    FileText
  } from 'lucide-svelte';

  interface Props {
    visible: boolean;
    toolCall: ToolCall;
    meta: SubagentParsedMeta;
    onClose: () => void;
  }

  let { visible = false, toolCall, meta, onClose }: Props = $props();

  let copiedPrompt = $state(false);
  let copiedOutput = $state(false);
  let now = $state(Date.now());
  let showRawPayload = $state(false);

  // Active ticking timer for running subagents
  $effect(() => {
    if (!visible || toolCall.status !== 'running') return;
    const interval = setInterval(() => {
      now = Date.now();
    }, 1000);
    return () => clearInterval(interval);
  });

  const elapsedDuration = $derived.by(() => {
    if (!toolCall.startTime) return null;
    const end = toolCall.endTime || (toolCall.status === 'running' ? now : toolCall.startTime);
    return formatSubagentDuration(Math.max(0, end - toolCall.startTime));
  });

  const cleanedOutput = $derived.by(() => {
    return cleanSubagentOutput(toolCall.result);
  });

  async function copyPrompt() {
    if (!meta.prompt) return;
    try {
      await navigator.clipboard.writeText(meta.prompt);
      copiedPrompt = true;
      setTimeout(() => (copiedPrompt = false), 2000);
    } catch {
      // ignore
    }
  }

  async function copyOutput() {
    if (!cleanedOutput) return;
    try {
      await navigator.clipboard.writeText(cleanedOutput);
      copiedOutput = true;
      setTimeout(() => (copiedOutput = false), 2000);
    } catch {
      // ignore
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (!visible) return;
    if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeyDown);
  });
</script>

{#if visible}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150 select-text font-sans"
    onclick={(e) => {
      if (e.target === e.currentTarget) onClose();
    }}
    onkeydown={handleKeyDown}
  >
    <!-- Modal Card Container -->
    <div
      class="w-full max-w-3xl max-h-[85vh] flex flex-col bg-ant-bg border border-violet-500/30 rounded-xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-5 py-3.5 bg-ant-bg-secondary/70 border-b border-white/5 select-none">
        <div class="flex items-center space-x-3 min-w-0">
          <div class="w-8 h-8 rounded-lg bg-violet-500/15 border border-violet-500/30 flex items-center justify-center flex-shrink-0 text-violet-400">
            {#if toolCall.status === 'running'}
              <Loader2 size={18} class="animate-spin text-violet-400" />
            {:else}
              <Bot size={18} />
            {/if}
          </div>
          <div class="min-w-0">
            <div class="flex items-center space-x-2">
              <span class="text-xs font-semibold uppercase tracking-wider text-violet-400 font-mono">Subagent Process</span>
              {#if elapsedDuration !== null}
                <span class="inline-flex items-center text-[11px] font-mono font-medium px-2 py-0.5 rounded-full bg-violet-500/10 text-violet-300 border border-violet-500/20">
                  <Clock size={11} class="mr-1" />
                  {elapsedDuration}
                </span>
              {/if}
            </div>
            <h2 class="text-sm font-semibold text-ant-text truncate mt-0.5">
              {meta.description}
            </h2>
          </div>
        </div>

        <!-- Right: Status Badge & Close Button -->
        <div class="flex items-center space-x-3 flex-shrink-0">
          {#if toolCall.status === 'running'}
            <span class="inline-flex items-center text-xs font-medium text-violet-400 animate-pulse bg-violet-500/10 px-2.5 py-1 rounded-md border border-violet-500/20">
              <Loader2 size={12} class="animate-spin mr-1.5" />
              Running
            </span>
          {:else if toolCall.status === 'completed'}
            <span class="inline-flex items-center text-xs font-medium text-emerald-400 bg-emerald-500/10 px-2.5 py-1 rounded-md border border-emerald-500/20">
              <CheckCircle2 size={12} class="mr-1.5" />
              Completed
            </span>
          {:else if toolCall.status === 'error'}
            <span class="inline-flex items-center text-xs font-medium text-rose-400 bg-rose-500/10 px-2.5 py-1 rounded-md border border-rose-500/20">
              <XCircle size={12} class="mr-1.5" />
              Failed
            </span>
          {/if}

          <button
            type="button"
            onclick={onClose}
            class="text-ant-text-muted hover:text-ant-text hover:bg-white/5 p-1 rounded-md transition-colors"
            title="Close (Escape)"
          >
            <X size={16} />
          </button>
        </div>
      </div>

      <!-- Quick Metadata Bar -->
      <div class="flex items-center flex-wrap gap-2 px-5 py-2.5 bg-ant-bg-tertiary/40 border-b border-white/5 text-xs text-ant-text-secondary font-mono select-none">
        <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
          <Bot size={12} class="text-violet-400" />
          <span>Role: {meta.subagentType}</span>
        </span>
        {#if meta.model}
          <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
            <Cpu size={12} class="text-indigo-400" />
            <span>Model: {meta.model}</span>
          </span>
        {/if}
        {#if meta.isolation}
          <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
            <Shield size={12} class="text-amber-400" />
            <span>Isolation: {meta.isolation}</span>
          </span>
        {/if}
        <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
          <span>Execution: {meta.isBackground ? 'Background' : 'Foreground'}</span>
        </span>
      </div>

      <!-- Modal Body (Scrollable) -->
      <div class="flex-1 overflow-y-auto p-5 space-y-4 scrollbar-thin">
        <!-- 1. Task Instructions (Prompt) -->
        <div class="rounded-lg border border-white/10 bg-ant-bg-secondary/40 overflow-hidden">
          <div class="flex items-center justify-between px-3.5 py-2 bg-ant-bg-tertiary/60 border-b border-white/5 text-xs select-none">
            <span class="font-semibold text-ant-text flex items-center gap-1.5">
              <FileText size={13} class="text-violet-400" />
              Task Prompt
            </span>
            <button
              type="button"
              onclick={copyPrompt}
              class="flex items-center space-x-1 text-xs text-ant-text-secondary hover:text-ant-text px-1.5 py-0.5 rounded hover:bg-white/5 transition-colors"
            >
              {#if copiedPrompt}
                <Check size={12} class="text-emerald-400" />
                <span class="text-emerald-400">Copied</span>
              {:else}
                <Copy size={12} />
                <span>Copy</span>
              {/if}
            </button>
          </div>
          <div class="p-3.5 text-xs font-mono text-ant-text leading-relaxed whitespace-pre-wrap max-h-48 overflow-y-auto scrollbar-thin select-text">
            {meta.prompt || 'No prompt instructions recorded.'}
          </div>
        </div>

        <!-- 2. Subagent Result / Live Activity -->
        <div class="rounded-lg border border-white/10 bg-ant-bg-secondary/40 overflow-hidden">
          <div class="flex items-center justify-between px-3.5 py-2 bg-ant-bg-tertiary/60 border-b border-white/5 text-xs select-none">
            <span class="font-semibold text-ant-text flex items-center gap-1.5">
              <Code size={13} class="text-emerald-400" />
              Output of Subagent
            </span>
            {#if cleanedOutput}
              <button
                type="button"
                onclick={copyOutput}
                class="flex items-center space-x-1 text-xs text-ant-text-secondary hover:text-ant-text px-1.5 py-0.5 rounded hover:bg-white/5 transition-colors"
              >
                {#if copiedOutput}
                  <Check size={12} class="text-emerald-400" />
                  <span class="text-emerald-400">Copied</span>
                {:else}
                  <Copy size={12} />
                  <span>Copy</span>
                {/if}
              </button>
            {/if}
          </div>

          <div class="p-4 text-xs leading-relaxed max-h-80 overflow-y-auto scrollbar-thin select-text">
            {#if toolCall.status === 'running' && !cleanedOutput}
              <div class="flex items-center justify-center py-8 space-x-3 text-ant-text-muted font-mono">
                <Loader2 size={16} class="animate-spin text-violet-400" />
                <span>Subagent is actively executing tasks in the background...</span>
              </div>
            {:else if cleanedOutput}
              <div class="prose prose-invert prose-xs max-w-none">
                {@html renderMarkdown(cleanedOutput)}
              </div>
            {:else}
              <div class="py-4 text-center text-ant-text-muted font-mono">
                No output returned from subagent.
              </div>
            {/if}
          </div>
        </div>

        <!-- 3. Collapsible Raw Arguments -->
        <div class="rounded-lg border border-white/5 bg-ant-bg-tertiary/20 overflow-hidden text-xs">
          <button
            type="button"
            onclick={() => (showRawPayload = !showRawPayload)}
            class="w-full flex items-center justify-between px-3.5 py-2 hover:bg-white/5 text-ant-text-muted hover:text-ant-text font-mono text-left transition-colors"
          >
            <span>Raw Parameters & Config</span>
            <span class="text-[10px]">{showRawPayload ? '▲ Hide' : '▼ View'}</span>
          </button>
          {#if showRawPayload}
            <div class="p-3 border-t border-white/5 bg-black/40 overflow-x-auto max-h-40 scrollbar-thin">
              <pre class="font-mono text-[11px] text-ant-text-muted">{JSON.stringify(toolCall.params, null, 2)}</pre>
            </div>
          {/if}
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between px-5 py-3 bg-ant-bg-secondary/40 border-t border-white/5 text-xs text-ant-text-muted select-none">
        <span class="font-mono">Press Escape to close</span>
        <button
          type="button"
          onclick={onClose}
          class="px-3 py-1.5 rounded-md bg-white/10 hover:bg-white/15 text-ant-text font-medium transition-colors"
        >
          Close
        </button>
      </div>
    </div>
  </div>
{/if}
