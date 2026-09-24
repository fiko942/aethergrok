<script lang="ts">
  import type { ChatMessage, ToolCall } from '$lib/stores/session.svelte';
  import {
    Bot,
    User,
    Terminal,
    ChevronDown,
    ChevronRight,
    CheckCircle2,
    XCircle,
    Loader2,
    Clock,
    Zap,
    Copy,
    Check
  } from 'lucide-svelte';

  interface Props {
    message: ChatMessage;
    turnNumber?: number;
  }

  let { message, turnNumber }: Props = $props();

  let copied = $state(false);
  let expandedToolCalls = $state<Record<string, boolean>>({});

  function formatTime(timestamp: number): string {
    const d = new Date(timestamp);
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  }

  function toggleToolCall(id: string) {
    expandedToolCalls[id] = !expandedToolCalls[id];
  }

  async function copyContent() {
    try {
      await navigator.clipboard.writeText(message.content);
      copied = true;
      setTimeout(() => {
        copied = false;
      }, 2000);
    } catch {
      // Fallback or ignore
    }
  }

  // Simple, robust inline markdown formatting parser (bold, italic, inline code, code blocks, lists)
  function renderSimpleMarkdown(raw: string): string {
    if (!raw) return '';

    // Escape basic html
    let escaped = raw
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');

    // Code blocks ```lang\ncode\n```
    escaped = escaped.replace(/```([a-zA-Z0-9_-]*)\n([\s\S]*?)```/g, (_match, lang, code) => {
      const languageBadge = lang ? `<span class="absolute top-2 right-2 text-[10px] uppercase font-mono text-ant-text-muted bg-ant-bg-tertiary px-1.5 py-0.5 rounded border border-ant-border-secondary">${lang}</span>` : '';
      return `<div class="relative my-3 rounded-lg overflow-hidden border border-ant-border bg-ant-bg-secondary">${languageBadge}<pre class="p-3.5 text-xs font-mono overflow-x-auto text-ant-text leading-relaxed"><code>${code.trim()}</code></pre></div>`;
    });

    // Inline code `code`
    escaped = escaped.replace(/`([^`]+)`/g, '<code class="px-1.5 py-0.5 mx-0.5 text-[11px] font-mono rounded bg-ant-bg-tertiary text-ant-primary border border-ant-border-secondary">$1</code>');

    // Bold **text**
    escaped = escaped.replace(/\*\*([^*]+)\*\*/g, '<strong class="font-semibold text-white">$1</strong>');

    // Italic *text*
    escaped = escaped.replace(/\*([^*]+)\*/g, '<em class="italic text-ant-text-secondary">$1</em>');

    // Blockquotes > text
    escaped = escaped.replace(/^>\s*(.+)$/gm, '<blockquote class="border-l-2 border-ant-primary pl-3 py-0.5 my-1.5 text-ant-text-secondary italic">$1</blockquote>');

    // Bullet points
    escaped = escaped.replace(/^\s*[-*]\s+(.+)$/gm, '<li class="ml-4 list-disc text-ant-text my-0.5">$1</li>');

    // Line breaks
    escaped = escaped.replace(/\n\n/g, '<br/><br/>').replace(/\n/g, '<br/>');

    return escaped;
  }
</script>

<div
  class="group relative flex flex-col w-full px-4 py-3 rounded-lg transition-colors border {message.role === 'user'
    ? 'bg-ant-bg-secondary/40 border-ant-border-secondary/60 hover:border-ant-border'
    : 'bg-ant-bg border-transparent hover:border-ant-border-secondary'}"
>
  <!-- Header: Role avatar, Turn ID, Timestamp, Token stats, Copy button -->
  <div class="flex items-center justify-between mb-2">
    <div class="flex items-center space-x-2">
      <!-- Role Icon -->
      {#if message.role === 'user'}
        <div class="flex items-center justify-center w-6 h-6 rounded-md bg-ant-primary/20 text-ant-primary border border-ant-primary/30">
          <User size={13} />
        </div>
        <span class="text-xs font-semibold text-white">You</span>
      {:else if message.role === 'assistant'}
        <div class="flex items-center justify-center w-6 h-6 rounded-md bg-ant-success/20 text-ant-success border border-ant-success/30">
          <Bot size={13} />
        </div>
        <span class="text-xs font-semibold text-white">Grok</span>
      {:else}
        <div class="flex items-center justify-center w-6 h-6 rounded-md bg-ant-warning/20 text-ant-warning border border-ant-warning/30">
          <Terminal size={13} />
        </div>
        <span class="text-xs font-semibold text-ant-warning">System</span>
      {/if}

      <!-- Turn Badge -->
      {#if turnNumber !== undefined}
        <span class="text-[10px] font-mono px-1.5 py-0.2 rounded bg-ant-bg-tertiary text-ant-text-muted border border-ant-border-secondary">
          Turn #{turnNumber}
        </span>
      {/if}

      <!-- Timestamp -->
      <span class="flex items-center text-[11px] text-ant-text-muted ml-1">
        <Clock size={11} class="mr-1" />
        {formatTime(message.timestamp)}
      </span>

      <!-- Steer badge if applicable -->
      {#if message.isSteer}
        <span class="text-[10px] font-medium px-1.5 py-0.2 rounded bg-ant-warning/15 text-ant-warning border border-ant-warning/30">
          Steer
        </span>
      {/if}
    </div>

    <!-- Right Side: Token info & Copy Action -->
    <div class="flex items-center space-x-2 opacity-70 group-hover:opacity-100 transition-opacity">
      {#if message.tokens?.total}
        <div class="flex items-center text-[10px] font-mono text-ant-text-muted bg-ant-bg-secondary px-2 py-0.5 rounded border border-ant-border-secondary">
          <Zap size={10} class="mr-1 text-ant-primary" />
          <span>{message.tokens.total.toLocaleString()} tokens</span>
        </div>
      {/if}

      <button
        type="button"
        onclick={copyContent}
        class="p-1 rounded text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-secondary transition"
        title="Copy message content"
      >
        {#if copied}
          <Check size={12} class="text-ant-success" />
        {:else}
          <Copy size={12} />
        {/if}
      </button>
    </div>
  </div>

  <!-- Message Body (Formatted Markdown) -->
  {#if message.content}
    <div class="text-xs leading-relaxed text-ant-text font-normal space-y-2 select-text">
      {@html renderSimpleMarkdown(message.content)}
    </div>
  {/if}

  <!-- Tool Invocations Section -->
  {#if message.toolCalls && message.toolCalls.length > 0}
    <div class="mt-3 space-y-2">
      <div class="text-[11px] font-semibold text-ant-text-secondary flex items-center gap-1.5">
        <Terminal size={12} class="text-ant-primary" />
        <span>Tool Invocations ({message.toolCalls.length})</span>
      </div>

      <div class="space-y-1.5">
        {#each message.toolCalls as tool (tool.id)}
          {@const isExpanded = !!expandedToolCalls[tool.id]}
          <div class="rounded-md border border-ant-border bg-ant-bg-secondary overflow-hidden">
            <!-- Tool Header Bar -->
            <button
              type="button"
              onclick={() => toggleToolCall(tool.id)}
              class="w-full flex items-center justify-between px-2.5 py-1.5 text-xs text-left bg-ant-bg-secondary hover:bg-ant-bg-tertiary transition"
            >
              <div class="flex items-center space-x-2 min-w-0">
                {#if isExpanded}
                  <ChevronDown size={13} class="text-ant-text-muted flex-shrink-0" />
                {:else}
                  <ChevronRight size={13} class="text-ant-text-muted flex-shrink-0" />
                {/if}
                <span class="font-mono font-medium text-white truncate">{tool.tool}</span>
                {#if tool.status === 'running'}
                  <span class="flex items-center text-[10px] text-ant-primary bg-ant-primary/10 px-1.5 py-0.2 rounded border border-ant-primary/20">
                    <Loader2 size={10} class="animate-spin mr-1" /> Running
                  </span>
                {:else if tool.status === 'completed'}
                  <span class="flex items-center text-[10px] text-ant-success bg-ant-success/10 px-1.5 py-0.2 rounded border border-ant-success/20">
                    <CheckCircle2 size={10} class="mr-1" /> Success
                  </span>
                {:else if tool.status === 'error'}
                  <span class="flex items-center text-[10px] text-ant-error bg-ant-error/10 px-1.5 py-0.2 rounded border border-ant-error/20">
                    <XCircle size={10} class="mr-1" /> Failed
                  </span>
                {/if}
              </div>

              {#if tool.startTime && tool.endTime}
                <span class="text-[10px] font-mono text-ant-text-muted">
                  {tool.endTime - tool.startTime}ms
                </span>
              {/if}
            </button>

            <!-- Expanded Details: Parameters & Output -->
            {#if isExpanded}
              <div class="p-2.5 bg-ant-bg border-t border-ant-border text-[11px] font-mono space-y-2">
                {#if tool.params}
                  <div>
                    <div class="text-[10px] font-semibold text-ant-text-muted uppercase mb-1">Parameters</div>
                    <pre class="p-2 rounded bg-ant-bg-secondary text-ant-text border border-ant-border-secondary overflow-x-auto"><code>{typeof tool.params === 'string' ? tool.params : JSON.stringify(tool.params, null, 2)}</code></pre>
                  </div>
                {/if}

                {#if tool.result}
                  <div>
                    <div class="text-[10px] font-semibold text-ant-text-muted uppercase mb-1">Result</div>
                    <pre class="p-2 rounded bg-ant-bg-secondary text-ant-text border border-ant-border-secondary overflow-x-auto max-h-48"><code>{tool.result}</code></pre>
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Streaming pulse indicator -->
  {#if message.status === 'streaming'}
    <div class="flex items-center space-x-1.5 mt-2 text-ant-primary text-xs font-mono">
      <Loader2 size={12} class="animate-spin" />
      <span>Grok is thinking...</span>
    </div>
  {/if}
</div>
