<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { sessionStore, type ChatMessage } from '$lib/stores/session.svelte';
  import MessageItem from './MessageItem.svelte';
  import { ArrowUp, Loader2, Sparkles, Brain, Cpu, Compass, CheckCircle2, AlertCircle } from 'lucide-svelte';

  let containerEl = $state<HTMLDivElement | null>(null);
  let topSentinelEl = $state<HTMLDivElement | null>(null);
  let isHydrating = $state(false);
  let autoScrollToBottom = $state(true);

  // Live thinking timer state for in-transcript activity indicator
  let elapsedSeconds = $state(0);
  let thinkingTimer: ReturnType<typeof setInterval> | null = null;

  const isWorking = $derived(sessionStore.activeSession?.status === 'working');

  // Track live elapsed timer for the transcript thinking indicator
  $effect(() => {
    if (isWorking) {
      elapsedSeconds = 0;
      if (thinkingTimer) clearInterval(thinkingTimer);
      thinkingTimer = setInterval(() => {
        elapsedSeconds += 1;
      }, 1000);
    } else {
      if (thinkingTimer) {
        clearInterval(thinkingTimer);
        thinkingTimer = null;
      }
      elapsedSeconds = 0;
    }

    return () => {
      if (thinkingTimer) {
        clearInterval(thinkingTimer);
        thinkingTimer = null;
      }
    };
  });

  // Determine intelligent contextual thinking status text
  const thinkingContextText = $derived.by(() => {
    const session = sessionStore.activeSession;
    if (!session || !isWorking) return 'Thinking and analyzing...';

    const lastMsg = session.messages[session.messages.length - 1];
    if (lastMsg && lastMsg.toolCalls && lastMsg.toolCalls.length > 0) {
      const activeTool = lastMsg.toolCalls.find(tc => tc.status === 'running');
      if (activeTool) {
        const name = (activeTool.tool || '').toLowerCase();
        if (name.includes('read')) return 'Reading workspace file...';
        if (name.includes('search') || name.includes('grep')) return 'Searching codebase & patterns...';
        if (name.includes('edit') || name.includes('replace')) return 'Applying file modifications...';
        if (name.includes('terminal') || name.includes('bash')) return 'Executing shell command...';
        return `Running tool: ${activeTool.tool}...`;
      }
      return 'Evaluating tool outputs & preparing response...';
    }

    if (elapsedSeconds > 8) {
      return 'Formulating comprehensive response...';
    }
    return 'Grok is reasoning and planning actions...';
  });

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
        <h3 class="font-serif-display text-lg font-semibold text-ant-text tracking-tight">Start a new conversation</h3>
        <p class="font-serif text-[13px] text-ant-text-secondary mt-1 max-w-sm leading-relaxed">
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

    <!-- Live In-Transcript Compaction Indicator Banner -->
    {#if sessionStore.isCompacting}
      <div class="my-2 p-3 rounded-lg border border-white/5 bg-ant-bg-secondary/70 backdrop-blur-md shadow-lg shadow-black/10 dark:shadow-black/40 flex items-center justify-between font-mono select-none animate-in fade-in duration-200">
        <div class="flex items-center space-x-2.5 min-w-0">
          <div class="w-6 h-6 rounded-md bg-ant-primary/15 border border-transparent flex items-center justify-center text-ant-primary flex-shrink-0 animate-spin">
            <Loader2 size={13} />
          </div>
          <div class="flex flex-col min-w-0">
            <span class="text-xs font-semibold text-ant-text truncate flex items-center gap-1.5">
              <span>Compacting conversation context</span>
              <span class="inline-flex space-x-0.5">
                <span class="w-1 h-1 rounded-full bg-ant-primary animate-bounce"></span>
                <span class="w-1 h-1 rounded-full bg-ant-primary animate-bounce [animation-delay:0.2s]"></span>
                <span class="w-1 h-1 rounded-full bg-ant-primary animate-bounce [animation-delay:0.4s]"></span>
              </span>
            </span>
            <span class="text-[10px] text-ant-text-secondary mt-0.5">
              Executing compaction routine to summarize turns and reclaim context window
            </span>
          </div>
        </div>

        <div class="px-2 py-0.5 rounded bg-ant-primary/15 border border-transparent text-ant-primary text-[11px] font-mono flex-shrink-0 ml-3">
          <span class="font-medium">Compacting...</span>
        </div>
      </div>
    {:else if sessionStore.lastCompactNotice}
      <div class="my-2 p-2.5 rounded-lg border flex items-center justify-between font-mono select-none text-xs transition-all duration-300 {sessionStore.lastCompactNotice.type === 'success' ? 'border-ant-success/20 bg-ant-success/10 text-ant-success' : 'border-ant-error/20 bg-ant-error/10 text-ant-error'}">
        <div class="flex items-center space-x-2">
          {#if sessionStore.lastCompactNotice.type === 'success'}
            <CheckCircle2 size={14} class="flex-shrink-0 text-ant-success" />
          {:else}
            <AlertCircle size={14} class="flex-shrink-0 text-ant-error" />
          {/if}
          <span class="text-[11.5px] font-medium">{sessionStore.lastCompactNotice.message}</span>
        </div>
        <button
          type="button"
          onclick={() => sessionStore.lastCompactNotice = null}
          class="text-[10px] opacity-70 hover:opacity-100 uppercase tracking-wider ml-2 hover:underline"
        >
          Dismiss
        </button>
      </div>
    {/if}

    <!-- Live In-Transcript Activity & Thinking Indicator -->
    {#if isWorking}
      <div class="my-2 p-3 rounded-lg border border-white/5 bg-ant-bg-secondary/70 backdrop-blur-md shadow-sm flex items-center justify-between font-mono select-none">
        <div class="flex items-center space-x-2.5 min-w-0">
          <div class="w-6 h-6 rounded-md bg-ant-primary/10 text-ant-primary flex items-center justify-center flex-shrink-0 animate-pulse">
            <Brain size={14} />
          </div>
          <div class="flex flex-col min-w-0">
            <span class="text-xs font-semibold text-ant-text truncate flex items-center gap-1.5">
              <span>{thinkingContextText}</span>
              <span class="inline-flex space-x-0.5">
                <span class="w-1 h-1 rounded-full bg-ant-primary animate-bounce"></span>
                <span class="w-1 h-1 rounded-full bg-ant-primary animate-bounce [animation-delay:0.2s]"></span>
                <span class="w-1 h-1 rounded-full bg-ant-primary animate-bounce [animation-delay:0.4s]"></span>
              </span>
            </span>
            <span class="text-[10px] text-ant-text-secondary mt-0.5">
              Processing actions and analyzing workspace files
            </span>
          </div>
        </div>

        <!-- Right Side: Live Timer Pill -->
        <div class="flex items-center space-x-1.5 px-2 py-0.5 rounded bg-ant-bg-tertiary/60 border border-white/5 text-ant-primary text-[11px] font-mono flex-shrink-0 ml-3">
          <Loader2 size={11} class="animate-spin text-ant-primary" />
          <span class="font-medium">{elapsedSeconds}s</span>
        </div>
      </div>
    {/if}
  {/if}
</div>
