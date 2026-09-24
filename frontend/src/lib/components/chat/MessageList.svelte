<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { sessionStore, type ChatMessage } from '$lib/stores/session.svelte';
  import MessageItem from './MessageItem.svelte';
  import { ArrowUp, Loader2, Sparkles } from 'lucide-svelte';

  let containerEl = $state<HTMLDivElement | null>(null);
  let topSentinelEl = $state<HTMLDivElement | null>(null);
  let isHydrating = $state(false);
  let autoScrollToBottom = $state(true);

  // Derive turn numbers for all messages in the active session
  const turnMap = $derived.by(() => {
    const map = new Map<string, number>();
    let currentTurn = 0;
    const session = sessionStore.activeSession;
    if (!session) return map;

    for (const msg of session.messages) {
      if (msg.role === 'user') {
        currentTurn++;
      }
      map.set(msg.id, currentTurn);
    }
    return map;
  });

  // Track active session changes for clean DOM scroll reset
  let lastSessionId = $state<string | null>(null);

  $effect(() => {
    const currentId = sessionStore.activeSessionId;
    if (currentId !== lastSessionId) {
      lastSessionId = currentId;
      // Scroll to bottom on session switch
      tick().then(() => {
        if (containerEl) {
          containerEl.scrollTop = containerEl.scrollHeight;
        }
      });
    }
  });

  // Scroll anchor preservation when loading earlier turns
  async function loadEarlier() {
    if (isHydrating || sessionStore.remainingHiddenTurns === 0 || !containerEl) return;

    isHydrating = true;
    const previousScrollHeight = containerEl.scrollHeight;
    const previousScrollTop = containerEl.scrollTop;

    try {
      const loaded = sessionStore.loadEarlierTurns(10);
      if (loaded) {
        await tick();
        // Restore precise relative scroll position to avoid jump
        if (containerEl) {
          const heightDiff = containerEl.scrollHeight - previousScrollHeight;
          containerEl.scrollTop = previousScrollTop + heightDiff;
        }
      }
    } finally {
      isHydrating = false;
    }
  }

  // Handle scroll events: trigger auto-load near top & detect manual upward scrolling
  function handleScroll() {
    if (!containerEl) return;

    const { scrollTop, scrollHeight, clientHeight } = containerEl;

    // Check if user is near bottom to maintain stick-to-bottom
    const distanceFromBottom = scrollHeight - (scrollTop + clientHeight);
    autoScrollToBottom = distanceFromBottom < 80;

    // Trigger progressive prepend if user scrolls within 200px of top
    if (scrollTop < 200 && sessionStore.remainingHiddenTurns > 0 && !isHydrating) {
      loadEarlier();
    }
  }

  // Auto-scroll on new streaming messages if locked to bottom
  $effect(() => {
    const msgs = sessionStore.visibleMessages;
    if (msgs.length > 0 && autoScrollToBottom) {
      tick().then(() => {
        if (containerEl && autoScrollToBottom) {
          containerEl.scrollTop = containerEl.scrollHeight;
        }
      });
    }
  });

  onMount(() => {
    // Initial scroll to bottom
    if (containerEl) {
      containerEl.scrollTop = containerEl.scrollHeight;
    }

    // Set up IntersectionObserver on top sentinel for smooth upward infinite hydration
    if ('IntersectionObserver' in window && topSentinelEl && containerEl) {
      const observer = new IntersectionObserver(
        (entries) => {
          if (entries[0].isIntersecting && sessionStore.remainingHiddenTurns > 0 && !isHydrating) {
            loadEarlier();
          }
        },
        {
          root: containerEl,
          rootMargin: '300px 0px 0px 0px',
          threshold: 0.1
        }
      );

      observer.observe(topSentinelEl);

      return () => {
        observer.disconnect();
      };
    }
  });
</script>

<div
  bind:this={containerEl}
  onscroll={handleScroll}
  class="flex-1 w-full h-full overflow-y-auto px-4 py-3 space-y-3 relative bg-ant-bg select-text"
>
  <!-- Top Sentinel & Prepend History Header -->
  <div bind:this={topSentinelEl} class="w-full flex justify-center py-2">
    {#if sessionStore.remainingHiddenTurns > 0}
      <button
        type="button"
        onclick={loadEarlier}
        disabled={isHydrating}
        class="inline-flex items-center space-x-2 px-3 py-1.5 rounded-full text-xs font-medium bg-ant-bg-secondary hover:bg-ant-bg-tertiary text-ant-primary border border-ant-border shadow-sm transition-all duration-150 disabled:opacity-50"
      >
        {#if isHydrating}
          <Loader2 size={13} class="animate-spin text-ant-primary" />
          <span>Hydrating turns...</span>
        {:else}
          <ArrowUp size={13} class="text-ant-primary" />
          <span>Load earlier turns ({sessionStore.remainingHiddenTurns} remaining)</span>
        {/if}
      </button>
    {:else if sessionStore.visibleMessages.length > 0}
      <div class="text-[11px] text-ant-text-muted font-mono flex items-center gap-1.5 py-1">
        <span>Beginning of session conversation</span>
      </div>
    {/if}
  </div>

  <!-- Empty state if no messages -->
  {#if sessionStore.visibleMessages.length === 0}
    <div class="flex flex-col items-center justify-center h-64 text-center space-y-3">
      <div class="w-12 h-12 rounded-xl bg-ant-primary/10 border border-ant-primary/20 flex items-center justify-center text-ant-primary">
        <Sparkles size={24} />
      </div>
      <div>
        <h3 class="text-sm font-semibold text-white">Start a new conversation</h3>
        <p class="text-xs text-ant-text-secondary mt-1 max-w-sm">
          Ask Grok to inspect your code, execute terminal commands, or orchestrate autonomous agent tasks.
        </p>
      </div>
    </div>
  {:else}
    <!-- Render 10-turn windowed messages -->
    {#each sessionStore.visibleMessages as message (message.id)}
      <MessageItem
        {message}
        turnNumber={turnMap.get(message.id)}
      />
    {/each}
  {/if}
</div>
