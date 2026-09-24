<script lang="ts">
  import type { ChatMessage } from '$lib/stores/session.svelte';
  import ToolCallCard from './ToolCallCard.svelte';
  import {
    Bot,
    User,
    Terminal,
    Clock,
    Zap,
    Copy,
    Check,
    Image as ImageIcon
  } from 'lucide-svelte';

  interface Props {
    message: ChatMessage;
    turnNumber?: number;
  }

  let { message, turnNumber }: Props = $props();

  let copied = $state(false);

  function formatTime(timestamp: number): string {
    const d = new Date(timestamp);
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
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
      const languageBadge = lang
        ? `<span class="absolute top-2 right-2 text-[10px] uppercase font-mono text-ant-text-muted bg-ant-bg-tertiary px-1.5 py-0.5 rounded border border-ant-border-secondary">${lang}</span>`
        : '';
      return `<div class="relative my-3 rounded-lg overflow-hidden border border-ant-border bg-ant-bg-secondary">${languageBadge}<pre class="p-3.5 text-xs font-mono overflow-x-auto text-ant-text leading-relaxed"><code>${code.trim()}</code></pre></div>`;
    });

    // Inline code `code`
    escaped = escaped.replace(
      /`([^`]+)`/g,
      '<code class="px-1.5 py-0.5 mx-0.5 text-[11px] font-mono rounded bg-ant-bg-tertiary text-ant-primary border border-ant-border-secondary">$1</code>'
    );

    // Bold **text**
    escaped = escaped.replace(/\*\*([^*]+)\*\*/g, '<strong class="font-semibold text-ant-text">$1</strong>');

    // Italic *text*
    escaped = escaped.replace(/\*([^*]+)\*/g, '<em class="italic text-ant-text-secondary">$1</em>');

    // Blockquotes > text
    escaped = escaped.replace(
      /^>\s*(.+)$/gm,
      '<blockquote class="border-l-2 border-ant-primary pl-3 py-0.5 my-1.5 text-ant-text-secondary italic">$1</blockquote>'
    );

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
        <span class="text-xs font-semibold text-ant-text">You</span>
      {:else if message.role === 'assistant'}
        <div class="flex items-center justify-center w-6 h-6 rounded-md bg-ant-success/20 text-ant-success border border-ant-success/30">
          <Bot size={13} />
        </div>
        <span class="text-xs font-semibold text-ant-text">Grok</span>
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

  <!-- Vision Images Preview (if present) -->
  {#if message.images && message.images.length > 0}
    <div class="flex items-center gap-2 mb-2 flex-wrap">
      {#each message.images as img (img.id)}
        <div class="relative rounded-md overflow-hidden border border-ant-border bg-ant-bg-secondary group/img max-w-[140px] max-h-[90px]">
          {#if img.dataUrl}
            <img src={img.dataUrl} alt={img.filePath} class="object-cover w-full h-full rounded" />
          {:else}
            <div class="p-2 text-[10px] text-ant-text-muted flex items-center gap-1">
              <ImageIcon size={12} />
              <span class="truncate">{img.filePath}</span>
            </div>
          {/if}
        </div>
      {/each}
    </div>
  {/if}

  <!-- Message Body (Formatted Markdown) -->
  {#if message.content}
    <div class="text-xs leading-relaxed text-ant-text font-normal space-y-2 select-text">
      {@html renderSimpleMarkdown(message.content)}
    </div>
  {/if}

  <!-- Tool Invocations Section using ToolCallCard -->
  {#if message.toolCalls && message.toolCalls.length > 0}
    <div class="mt-3 space-y-2">
      <div class="text-[11px] font-semibold text-ant-text-secondary flex items-center gap-1.5">
        <Terminal size={12} class="text-ant-primary" />
        <span>Tool Invocations ({message.toolCalls.length})</span>
      </div>

      <div class="space-y-2">
        {#each message.toolCalls as tool (tool.id)}
          <ToolCallCard toolCall={tool} />
        {/each}
      </div>
    </div>
  {/if}

  <!-- Streaming pulse indicator -->
  {#if message.status === 'streaming'}
    <div class="flex items-center space-x-1.5 mt-2 text-ant-primary text-xs font-mono">
      <span class="inline-block w-2 h-2 rounded-full bg-ant-primary animate-ping mr-1"></span>
      <span>Grok is thinking & executing...</span>
    </div>
  {/if}
</div>
