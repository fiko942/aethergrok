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

{#if message.role === 'user'}
  <!-- USER MESSAGE: Right Aligned (Z-Pattern Flow) -->
  <div class="flex justify-end w-full my-2.5 px-2 select-text">
    <div class="max-w-[85%] sm:max-w-[75%] flex flex-col items-end group">
      <!-- User Info Header -->
      <div class="flex items-center space-x-1.5 mb-1 text-[11px] text-ant-text-muted font-mono select-none">
        {#if turnNumber !== undefined}
          <span class="text-[9.5px] px-1 py-0.5 rounded bg-ant-bg-tertiary text-ant-text-secondary border border-ant-border-secondary">
            #{turnNumber}
          </span>
        {/if}
        <span>{formatTime(message.timestamp)}</span>
        <div class="w-4 h-4 rounded-full bg-ant-primary/20 text-ant-primary flex items-center justify-center border border-ant-primary/40 ml-1">
          <User size={10} />
        </div>
      </div>

      <!-- User Bubble Card -->
      <div class="relative bg-ant-primary/10 border border-ant-primary/30 text-white rounded-2xl rounded-tr-sm px-4 py-2.5 text-xs shadow-sm hover:border-ant-primary/60 transition">
        <div class="whitespace-pre-wrap leading-relaxed select-text font-sans">
          {message.content}
        </div>

        <!-- Vision Images (if user attached images) -->
        {#if message.images && message.images.length > 0}
          <div class="flex flex-wrap gap-1.5 mt-2 pt-2 border-t border-ant-primary/20">
            {#each message.images as img}
              <div class="relative rounded overflow-hidden border border-ant-border-secondary group/img">
                {#if img.dataUrl}
                  <img src={img.dataUrl} alt={img.filePath || "Attachment"} class="w-20 h-14 object-cover" />
                {:else}
                  <div class="px-2 py-1 text-[10px] bg-ant-bg-secondary text-ant-text-secondary">
                    {img.filePath}
                  </div>
                {/if}
              </div>
            {/each}
          </div>
        {/if}

        <!-- Hover Copy Action -->
        <button
          type="button"
          onclick={copyContent}
          class="absolute -left-7 top-2 opacity-0 group-hover:opacity-100 p-1 rounded hover:bg-ant-bg-secondary text-ant-text-muted hover:text-ant-text transition"
          title="Salin pesan"
        >
          {#if copied}
            <Check size={12} class="text-ant-success" />
          {:else}
            <Copy size={12} />
          {/if}
        </button>
      </div>
    </div>
  </div>
{:else}
  <!-- ASSISTANT / SYSTEM MESSAGE: Left Aligned (Z-Pattern Flow) -->
  <div class="flex justify-start w-full my-2.5 px-2 select-text">
    <div class="w-full max-w-[96%] flex items-start space-x-2.5 group">
      <!-- Assistant Avatar -->
      <div class="flex-shrink-0 mt-0.5">
        {#if message.role === 'assistant'}
          <div class="w-6 h-6 rounded-lg bg-ant-bg-secondary border border-ant-border-secondary flex items-center justify-center text-ant-primary shadow-sm">
            <Bot size={13} />
          </div>
        {:else}
          <div class="w-6 h-6 rounded-lg bg-ant-warning/10 border border-ant-warning/30 flex items-center justify-center text-ant-warning shadow-sm">
            <Terminal size={13} />
          </div>
        {/if}
      </div>

      <!-- Main Assistant Content Column -->
      <div class="flex-1 min-w-0 flex flex-col">
        <!-- Header Info Bar -->
        <div class="flex items-center justify-between mb-1 select-none text-[11px] text-ant-text-muted font-mono">
          <div class="flex items-center space-x-2">
            <span class="font-semibold text-xs {message.role === 'assistant' ? 'text-ant-text' : 'text-ant-warning'}">
              {message.role === 'assistant' ? 'Grok' : 'System'}
            </span>
            <span>{formatTime(message.timestamp)}</span>
            {#if message.tokens?.total}
              <span class="text-[9.5px] px-1 py-0.5 rounded bg-ant-bg-tertiary text-ant-text-secondary border border-ant-border-secondary">
                {message.tokens.total.toLocaleString()} tokens
              </span>
            {/if}
          </div>

          <div class="opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              type="button"
              onclick={copyContent}
              class="p-1 rounded text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-secondary transition"
              title="Salin jawaban"
            >
              {#if copied}
                <Check size={12} class="text-ant-success" />
              {:else}
                <Copy size={12} />
              {/if}
            </button>
          </div>
        </div>

        <!-- Collapsible Thought Process Pill -->
        {#if message.reasoningContent}
          <div class="mb-2 rounded-lg border border-ant-border-secondary/60 bg-ant-bg-secondary/40 overflow-hidden text-xs max-w-2xl">
            <button
              type="button"
              onclick={() => showReasoning = !showReasoning}
              class="flex items-center justify-between w-full px-2.5 py-1 text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-secondary transition font-mono select-none"
            >
              <div class="flex items-center space-x-1.5">
                <Brain size={12} class="text-ant-primary" />
                <span class="font-medium text-[10.5px]">Thought Process</span>
                <span class="text-[9.5px] text-ant-text-muted">({message.reasoningContent.length} chars)</span>
              </div>
              {#if showReasoning}
                <ChevronDown size={12} />
              {:else}
                <ChevronRight size={12} />
              {/if}
            </button>

            {#if showReasoning}
              <div class="p-2.5 border-t border-ant-border-secondary/40 bg-ant-bg text-ant-text-secondary text-xs leading-relaxed font-mono whitespace-pre-wrap select-text max-h-56 overflow-y-auto scrollbar-thin">
                {message.reasoningContent}
              </div>
            {/if}
          </div>
        {/if}

        <!-- Tool Calls Card Section -->
        {#if message.toolCalls && message.toolCalls.length > 0}
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

        <!-- Message Markdown Content -->
        {#if message.content}
          <div
            class="prose prose-invert max-w-none text-xs text-ant-text leading-relaxed select-text space-y-2 prose-headings:font-semibold prose-headings:text-white prose-h1:text-sm prose-h2:text-xs prose-h3:text-xs prose-p:my-1.5 prose-ul:my-1.5 prose-ul:list-disc prose-ul:pl-4 prose-ol:my-1.5 prose-ol:list-decimal prose-ol:pl-4 prose-li:my-0.5 prose-code:text-[11px] prose-code:font-mono prose-code:bg-ant-bg-tertiary prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:text-ant-primary prose-pre:my-2 prose-pre:bg-ant-bg-secondary prose-pre:border prose-pre:border-ant-border-secondary prose-pre:rounded-lg prose-pre:p-3 prose-blockquote:border-l-2 prose-blockquote:border-ant-primary prose-blockquote:pl-3 prose-blockquote:text-ant-text-secondary prose-hr:my-3 prose-hr:border-ant-border-secondary/60 prose-strong:text-white prose-table:my-2 prose-table:border-collapse prose-th:border prose-th:border-ant-border-secondary prose-th:p-1.5 prose-th:bg-ant-bg-secondary prose-td:border prose-td:border-ant-border-secondary prose-td:p-1.5"
          >
            {@html renderedHtml}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
