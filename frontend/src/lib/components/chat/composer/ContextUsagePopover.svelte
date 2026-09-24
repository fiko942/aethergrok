<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { ChevronRight, ChevronDown } from 'lucide-svelte';

  interface Props {
    usedTokens?: number;
    maxTokens?: number;
    lastTurnInput?: number;
    lastTurnOutput?: number;
    lastTurnCacheRead?: number;
    lastTurnReasoning?: number;
    lastTurnModelCalls?: number;
    totalInput?: number;
    totalOutput?: number;
    totalCacheRead?: number;
    onCompact?: () => void;
  }

  let {
    usedTokens = 153036,
    maxTokens = 200000,
    lastTurnInput = 5545355,
    lastTurnOutput = 5988,
    lastTurnCacheRead = 4676263,
    lastTurnReasoning = 2105,
    lastTurnModelCalls = 42,
    totalInput = 133089013,
    totalOutput = 621193,
    totalCacheRead = 69119302,
    onCompact
  }: Props = $props();

  let isOpen = $state(false);
  let isSessionTotalExpanded = $state(true);
  let isLastTurnExpanded = $state(true);
  let containerRef = $state<HTMLDivElement | null>(null);

  const percentage = $derived.by(() => {
    if (!maxTokens || maxTokens <= 0) return 0;
    return Math.min(100, Math.round((usedTokens / maxTokens) * 100));
  });

  const formattedShortUsed = $derived.by(() => {
    return Math.round(usedTokens / 1000) + 'K';
  });

  const formattedShortMax = $derived.by(() => {
    return Math.round(maxTokens / 1000) + 'K';
  });

  // Calculate SVG circular arc values
  const radius = 6;
  const circumference = 2 * Math.PI * radius;
  const strokeDashoffset = $derived.by(() => {
    return circumference - (circumference * percentage) / 100;
  });

  function toggleOpen(e: MouseEvent) {
    e.stopPropagation();
    isOpen = !isOpen;
  }

  function handleClickOutside(e: PointerEvent) {
    if (isOpen && containerRef && !containerRef.contains(e.target as Node)) {
      isOpen = false;
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      isOpen = false;
    }
  }

  onMount(() => {
    window.addEventListener('pointerdown', handleClickOutside);
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('pointerdown', handleClickOutside);
      window.removeEventListener('keydown', handleKeyDown);
    }
  });
</script>

<div class="relative inline-flex items-center" bind:this={containerRef}>
  <!-- Context Donut Button Trigger (Matching screenshot: [Donut] 153K/200K) -->
  <button
    type="button"
    onclick={toggleOpen}
    class="flex items-center space-x-1.5 px-2 py-1 rounded-lg text-xs font-mono text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition-colors {isOpen ? 'bg-ant-bg-tertiary text-ant-text' : ''}"
    title="View context window and token usage"
  >
    <!-- SVG Circular Donut Chart -->
    <svg width="15" height="15" viewBox="0 0 16 16" class="flex-shrink-0 -rotate-90">
      <circle
        cx="8"
        cy="8"
        r={radius}
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        class="text-ant-bg-tertiary"
      />
      <circle
        cx="8"
        cy="8"
        r={radius}
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        stroke-dasharray={circumference}
        stroke-dashoffset={strokeDashoffset}
        stroke-linecap="round"
        class="{percentage > 85 ? 'text-amber-400' : 'text-amber-300'}"
      />
    </svg>

    <span class="text-[11px] font-mono tracking-tight">{formattedShortUsed}/{formattedShortMax}</span>
  </button>

  <!-- Detailed Usage Popover (Matching screenshot) -->
  {#if isOpen}
    <div
      class="absolute bottom-full left-0 mb-2.5 w-80 bg-ant-bg border border-ant-border-secondary rounded-xl shadow-2xl z-50 p-4 select-none text-xs animate-in fade-in zoom-in-95 duration-100"
      role="dialog"
      aria-label="Context usage"
    >
      <!-- Header Row: Context used & Compact link -->
      <div class="flex items-center justify-between pb-2 border-b border-ant-border-secondary/60">
        <span class="font-medium text-ant-text">Context used</span>
        <span class="font-mono text-ant-text-secondary text-[11px]">
          {usedTokens.toLocaleString()} / {maxTokens.toLocaleString()} ({percentage}%)
        </span>
      </div>

      <!-- Action: Compact Conversation Link -->
      <div class="py-2">
        <button
          type="button"
          onclick={() => {
            onCompact?.();
            isOpen = false;
          }}
          class="text-ant-primary hover:underline text-xs font-medium transition"
        >
          Compact conversation
        </button>
      </div>

      <!-- Breakdown: SESSION TOTAL Section -->
      <div class="border-t border-ant-border-secondary/60 pt-2 space-y-1.5">
        <button
          type="button"
          onclick={() => isSessionTotalExpanded = !isSessionTotalExpanded}
          class="w-full flex items-center justify-between text-[10px] font-semibold text-ant-text-muted uppercase tracking-wider hover:text-ant-text py-0.5"
        >
          <span>Session Total</span>
          {#if isSessionTotalExpanded}
            <ChevronDown size={12} />
          {:else}
            <ChevronRight size={12} />
          {/if}
        </button>

        {#if isSessionTotalExpanded}
          <div class="pl-1 space-y-1 font-mono text-[11px] text-ant-text-secondary">
            <div class="flex items-center justify-between">
              <span>Input</span>
              <span class="text-ant-text">{totalInput.toLocaleString()}</span>
            </div>
            <div class="flex items-center justify-between pl-3 text-ant-text-muted text-[10.5px]">
              <span>↳ cache read</span>
              <span>{totalCacheRead.toLocaleString()}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>Output</span>
              <span class="text-ant-text">{totalOutput.toLocaleString()}</span>
            </div>
          </div>
        {/if}
      </div>

      <!-- Breakdown: LAST TURN Section -->
      <div class="border-t border-ant-border-secondary/60 pt-2 mt-2 space-y-1.5">
        <button
          type="button"
          onclick={() => isLastTurnExpanded = !isLastTurnExpanded}
          class="w-full flex items-center justify-between text-[10px] font-semibold text-ant-text-muted uppercase tracking-wider hover:text-ant-text py-0.5"
        >
          <span>Last Turn</span>
          {#if isLastTurnExpanded}
            <ChevronDown size={12} />
          {:else}
            <ChevronRight size={12} />
          {/if}
        </button>

        {#if isLastTurnExpanded}
          <div class="pl-1 space-y-1 font-mono text-[11px] text-ant-text-secondary">
            <div class="flex items-center justify-between">
              <span>Input</span>
              <span class="text-ant-text">{lastTurnInput.toLocaleString()}</span>
            </div>
            <div class="flex items-center justify-between pl-3 text-ant-text-muted text-[10.5px]">
              <span>↳ cache read</span>
              <span>{lastTurnCacheRead.toLocaleString()}</span>
            </div>
            <div class="flex items-center justify-between">
              <span>Output</span>
              <span class="text-ant-text">{lastTurnOutput.toLocaleString()}</span>
            </div>
            <div class="flex items-center justify-between pl-3 text-ant-text-muted text-[10.5px]">
              <span>↳ reasoning</span>
              <span>{lastTurnReasoning.toLocaleString()}</span>
            </div>
            <div class="flex items-center justify-between pt-1 border-t border-ant-border-secondary/30">
              <span>Model calls</span>
              <span class="text-ant-text">{lastTurnModelCalls}</span>
            </div>
          </div>
        {/if}
      </div>

      <!-- Footer Note (Matching screenshot copy verbatim) -->
      <div class="mt-3 pt-2 border-t border-ant-border-secondary/60 text-[10px] text-ant-text-muted leading-relaxed">
        Context is how full the window is. Token counts are billed usage tracked here — each model call re-sends the conversation, so a turn bills far more than the context it holds.
      </div>
    </div>
  {/if}
</div>
