<script lang="ts">
  import type { ChatMessage } from '$lib/stores/session.svelte';
  import ToolCallCard from './ToolCallCard.svelte';
  import TurnDiffSummary from './TurnDiffSummary.svelte';
  import PlanReviewCard from './PlanReviewCard.svelte';
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
    Check,
    Edit3
  } from 'lucide-svelte';

  interface Props {
    message: ChatMessage;
    turnNumber?: number;
    isLastUserTurn?: boolean;
    onEditLastTurn?: () => void;
    onOpenImage?: (src: string, title?: string) => void;
    onPlanAction?: (action: 'approve' | 'reject' | 'custom', feedback?: string) => void;
  }

  let { message, turnNumber, isLastUserTurn = false, onEditLastTurn, onOpenImage, onPlanAction }: Props = $props();

  let showReasoning = $state(false);
  let showToolDetails = $state(true);
  let copied = $state(false);

  // Check if this assistant message is presenting a pending execution plan
  const isPlanProposal = $derived.by(() => {
    if (message.role !== 'assistant') return false;
    const lower = (message.content || '').toLowerCase();
    const hasPlanHeading = lower.includes('### execution plan') || lower.includes('## plan') || lower.includes('### plan') || lower.includes('rancangan perencanaan') || lower.includes('implementation plan');
    const hasExitPlanTool = message.toolCalls?.some(tc => tc.tool.includes('exit_plan_mode') || tc.tool.includes('plan'));
    return (hasPlanHeading || hasExitPlanTool) && message.status === 'done';
  });

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
          <span class="text-[9.5px] px-1 py-0.5 rounded bg-ant-bg-tertiary text-ant-text-secondary border border-white/5">
            #{turnNumber}
          </span>
        {/if}
        <span>{formatTime(message.timestamp)}</span>
        <div class="w-4 h-4 rounded-full bg-ant-primary/15 text-ant-primary flex items-center justify-center border border-ant-primary/20 ml-1">
          <User size={10} />
        </div>
      </div>

      <!-- User Bubble Card -->
      <div class="relative bg-ant-bg-secondary border border-white/5 text-ant-text rounded-2xl rounded-tr-sm px-4 py-2.5 text-xs shadow-sm hover:border-white/10 transition">
        <div class="whitespace-pre-wrap leading-relaxed select-text font-serif text-[13.5px]">
          {message.content}
        </div>

        <!-- Vision Images (if user attached images) -->
        {#if message.images && message.images.length > 0}
          <div class="flex flex-wrap gap-1.5 mt-2 pt-2 border-t border-white/5">
            {#each message.images as img}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div
                class="relative rounded overflow-hidden border border-white/5 group/img cursor-pointer hover:border-ant-primary/50 transition-all hover:scale-[1.02]"
                onclick={() => onOpenImage?.(img.dataUrl || img.filePath, img.filePath || 'Attached Image')}
              >
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

        <!-- Hover Actions: Copy & Edit (Edit only on last user prompt) -->
        <div class="absolute -left-14 top-2 opacity-0 group-hover:opacity-100 flex items-center space-x-1 transition">
          {#if isLastUserTurn && onEditLastTurn}
            <button
              type="button"
              onclick={onEditLastTurn}
              class="p-1 rounded hover:bg-ant-bg-secondary text-ant-text-muted hover:text-ant-primary transition"
              title="Edit prompt (roll back turn and edit in composer)"
            >
              <Edit3 size={12} />
            </button>
          {/if}

          <button
            type="button"
            onclick={copyContent}
            class="p-1 rounded hover:bg-ant-bg-secondary text-ant-text-muted hover:text-ant-text transition"
            title="Copy message"
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
  </div>
{:else}
  <!-- ASSISTANT / SYSTEM MESSAGE: Left Aligned (Z-Pattern Flow) -->
  <div class="flex justify-start w-full my-2.5 px-2 select-text">
    <div class="w-full max-w-[96%] flex items-start space-x-2.5 group">
      <!-- Assistant Avatar -->
      <div class="flex-shrink-0 mt-0.5">
        {#if message.role === 'assistant'}
          <div class="w-6 h-6 rounded-lg bg-ant-bg-secondary border border-white/5 flex items-center justify-center text-ant-primary shadow-sm">
            <Bot size={13} />
          </div>
        {:else}
          <div class="w-6 h-6 rounded-lg bg-ant-warning/10 border border-ant-warning/20 flex items-center justify-center text-ant-warning shadow-sm">
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
              <span class="text-[9.5px] px-1 py-0.5 rounded bg-ant-bg-tertiary text-ant-text-secondary border border-white/5">
                {message.tokens.total.toLocaleString()} tokens
              </span>
            {/if}
          </div>

          <div class="opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              type="button"
              onclick={copyContent}
              class="p-1 rounded text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-secondary transition"
              title="Copy response"
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

        <!-- Multi-File Turn Diff Summary Rollup -->
        {#if message.toolCalls && message.toolCalls.some(tc => tc.diff)}
          <TurnDiffSummary toolCalls={message.toolCalls} />
        {/if}

        <!-- Message Markdown Content with Anthropic Serif Editorial Typography -->
        {#if message.content}
          <div
            class="font-serif text-[14px] text-ant-text leading-[1.7] tracking-normal select-text space-y-3 prose dark:prose-invert max-w-none prose-headings:font-serif-display prose-headings:font-semibold prose-headings:text-ant-text prose-headings:tracking-tight prose-h1:text-[18px] prose-h2:text-[16px] prose-h3:text-[14.5px] prose-p:my-2.5 prose-ul:my-2 prose-ul:list-disc prose-ul:pl-5 prose-ol:my-2 prose-ol:list-decimal prose-ol:pl-5 prose-li:my-1 prose-code:font-mono prose-code:text-[11.5px] prose-code:bg-ant-bg-tertiary prose-code:px-1.5 prose-code:py-0.5 prose-code:rounded prose-code:text-ant-primary prose-pre:my-2.5 prose-pre:bg-ant-bg-secondary prose-pre:border prose-pre:border-white/5 prose-pre:rounded-lg prose-pre:p-3 prose-pre:font-mono prose-blockquote:font-serif prose-blockquote:italic prose-blockquote:border-l-2 prose-blockquote:border-ant-primary/60 prose-blockquote:pl-3.5 prose-blockquote:text-ant-text-secondary prose-hr:my-4 prose-hr:border-white/5 dark:prose-hr:border-white/5 prose-strong:text-ant-text prose-strong:font-semibold prose-table:my-2.5 prose-table:border-collapse prose-table:font-sans prose-th:border prose-th:border-white/5 prose-th:p-1.5 prose-th:bg-ant-bg-secondary prose-th:text-xs prose-td:border prose-td:border-white/5 prose-td:p-1.5 prose-td:text-xs"
          >
            {@html renderedHtml}
          </div>
        {/if}

        <!-- Interactive Plan Review Card -->
        {#if isPlanProposal && onPlanAction}
          <PlanReviewCard
            onAction={onPlanAction}
          />
        {/if}
      </div>
    </div>
  </div>
{/if}
