<script lang="ts">
  import type { ChatMessage } from '$lib/stores/session.svelte';
  import ToolCallCard from './ToolCallCard.svelte';
  import { marked } from 'marked';
  import {
    User,
    Bot,
    Terminal,
    ChevronDown,
    ChevronRight,
    Brain,
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

  let showReasoning = $state(false);
  let showToolDetails = $state(true);
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

  // Configure marked for GitHub Flavored Markdown (tables, lists, breaks, headings)
  marked.setOptions({
    gfm: true,
    breaks: true
  });

  // Render comprehensive Markdown formatted text
  const renderedHtml = $derived.by(() => {
    if (!message.content) return '';
    try {
      return marked.parse(message.content) as string;
    } catch {
      return message.content;
    }
  });
</script>

<div
  class="group relative flex flex-col w-full px-4 py-3 rounded-lg transition-colors border {message.role === 'user'
    ? 'bg-ant-bg-secondary/30 border-ant-border-secondary/40 hover:border-ant-border-secondary/80'
    : 'bg-transparent border-transparent hover:border-ant-border-secondary/30'}"
>
  <!-- Header: Role avatar, Turn ID, Timestamp, Token stats, Copy button -->
  <div class="flex items-center justify-between mb-2 select-none">
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
      <span class="flex items-center text-[11px] text-ant-text-muted ml-1 font-mono">
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

  <!-- Reasoning / Thinking trace accordion (if available) -->
  {#if message.reasoningContent}
    <div class="mb-3 rounded-lg border border-ant-border-secondary/60 bg-ant-bg-secondary/30 overflow-hidden text-xs">
      <button
        type="button"
        onclick={() => showReasoning = !showReasoning}
        class="flex items-center justify-between w-full px-3 py-1.5 text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-secondary transition font-mono select-none"
      >
        <div class="flex items-center space-x-2">
          <Brain size={13} class="text-ant-primary" />
          <span class="font-medium text-[11px]">Thought Process</span>
          <span class="text-[10px] text-ant-text-muted">({message.reasoningContent.length} chars)</span>
        </div>
        {#if showReasoning}
          <ChevronDown size={13} />
        {:else}
          <ChevronRight size={13} />
        {/if}
      </button>

      {#if showReasoning}
        <div class="p-3 border-t border-ant-border-secondary/40 bg-ant-bg text-ant-text-secondary text-xs leading-relaxed font-mono whitespace-pre-wrap select-text max-h-60 overflow-y-auto scrollbar-thin">
          {message.reasoningContent}
        </div>
      {/if}
    </div>
  {/if}

  <!-- Tool Calls Section (Always attached exclusively to assistant messages) -->
  {#if message.role === 'assistant' && message.toolCalls && message.toolCalls.length > 0}
    <div class="mb-2 space-y-1">
      <div class="flex items-center justify-between text-[11px] text-ant-text-muted font-mono select-none px-1">
        <span class="flex items-center space-x-1 font-semibold text-ant-text-secondary">
          <Terminal size={11} class="text-ant-primary" />
          <span>Executed {message.toolCalls.length} {message.toolCalls.length === 1 ? 'action' : 'actions'}</span>
        </span>
        <button
          type="button"
          onclick={() => showToolDetails = !showToolDetails}
          class="hover:text-ant-text transition text-[10px]"
        >
          {showToolDetails ? 'Hide details' : 'Show details'}
        </button>
      </div>

      {#if showToolDetails}
        <div class="space-y-1">
          {#each message.toolCalls as toolCall (toolCall.id)}
            <ToolCallCard {toolCall} />
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <!-- Message Body: Rendered Markdown Content with Proper Styling -->
  {#if message.content}
    <div
      class="prose prose-invert max-w-none text-xs text-ant-text leading-relaxed select-text space-y-2 prose-headings:font-semibold prose-headings:text-white prose-h1:text-sm prose-h2:text-xs prose-h3:text-xs prose-p:my-1.5 prose-ul:my-1.5 prose-ul:list-disc prose-ul:pl-4 prose-ol:my-1.5 prose-ol:list-decimal prose-ol:pl-4 prose-li:my-0.5 prose-code:text-[11px] prose-code:font-mono prose-code:bg-ant-bg-tertiary prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:text-ant-primary prose-pre:my-2 prose-pre:bg-ant-bg-secondary prose-pre:border prose-pre:border-ant-border-secondary prose-pre:rounded-lg prose-pre:p-3 prose-blockquote:border-l-2 prose-blockquote:border-ant-primary prose-blockquote:pl-3 prose-blockquote:text-ant-text-secondary prose-hr:my-3 prose-hr:border-ant-border-secondary/60 prose-strong:text-white prose-table:my-2 prose-table:border-collapse prose-th:border prose-th:border-ant-border-secondary prose-th:p-1.5 prose-th:bg-ant-bg-secondary prose-td:border prose-td:border-ant-border-secondary prose-td:p-1.5"
    >
      {@html renderedHtml}
    </div>
  {/if}
</div>
