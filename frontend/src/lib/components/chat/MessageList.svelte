<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { sessionStore, type ChatMessage } from '$lib/stores/session.svelte';
  import MessageItem from './MessageItem.svelte';
  import ImageLightboxModal from './ImageLightboxModal.svelte';
  import { ArrowUp, ArrowDown, Loader2, Sparkles, Brain, Cpu, Compass, CheckCircle2, AlertCircle } from 'lucide-svelte';

  interface Props {
    onEditLastTurn?: () => void;
    onPlanAction?: (action: 'approve' | 'reject' | 'custom', feedback?: string) => void;
  }

  let { onEditLastTurn, onPlanAction }: Props = $props();

  export function forceScrollBottom() {
    autoScrollToBottom = true;
    unreadActivityBelow = false;
    tick().then(() => {
      if (containerEl) {
        containerEl.scrollTop = containerEl.scrollHeight;
      }
    });
  }

  let containerEl = $state<HTMLDivElement | null>(null);
  let contentWrapperEl = $state<HTMLDivElement | null>(null);
  let topSentinelEl = $state<HTMLDivElement | null>(null);
  let isHydrating = $state(false);
  let autoScrollToBottom = $state(true);
  let showScrollToBottom = $state(false);
  let unreadActivityBelow = $state(false);

  // Lightbox modal state
  let lightboxVisible = $state(false);
  let lightboxSrc = $state('');
  let lightboxTitle = $state('');

  function handleOpenImage(src: string, title?: string) {
    lightboxSrc = src;
    lightboxTitle = title || 'Image Preview';
    lightboxVisible = true;
  }

  // Live thinking timer state for in-transcript activity indicator
  let elapsedMs = $state(0);
  let thinkingTimer: ReturnType<typeof setInterval> | null = null;
  let startTimestamp = 0;

  const isWorking = $derived(sessionStore.activeSession?.status === 'working');

  // Track live elapsed timer for the transcript thinking indicator (in ms and seconds)
  $effect(() => {
    if (isWorking) {
      elapsedMs = 0;
      startTimestamp = Date.now();
      if (thinkingTimer) clearInterval(thinkingTimer);
      thinkingTimer = setInterval(() => {
        elapsedMs = Date.now() - startTimestamp;
      }, 100);
    } else {
      if (thinkingTimer) {
        clearInterval(thinkingTimer);
        thinkingTimer = null;
      }
      elapsedMs = 0;
    }

    return () => {
      if (thinkingTimer) {
        clearInterval(thinkingTimer);
        thinkingTimer = null;
      }
    };
  });

  const formattedElapsed = $derived.by(() => {
    if (elapsedMs < 1000) {
      return `${elapsedMs}ms`;
    }
    const sec = (elapsedMs / 1000).toFixed(1);
    return `${sec}s`;
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

  // Identify last user message ID to allow editing
  const lastUserMessageId = $derived.by(() => {
    const session = sessionStore.activeSession;
    if (!session) return null;
    for (let i = session.messages.length - 1; i >= 0; i--) {
      if (session.messages[i].role === 'user') {
        return session.messages[i].id;
      }
    }
    return null;
  });

  // Track active session changes for clean DOM scroll reset
  let lastSessionId = $state<string | null>(null);

  $effect(() => {
    const currentId = sessionStore.activeSessionId;
    if (currentId !== lastSessionId) {
      lastSessionId = currentId;
      // Scroll to bottom on session switch and acknowledge finished state
      tick().then(() => {
        if (containerEl) {
          containerEl.scrollTop = containerEl.scrollHeight;
        }
        const activeSession = sessionStore.activeSession;
        if (activeSession && activeSession.status === 'finished') {
          sessionStore.setSessionStatus(activeSession.id, 'idle');
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
    autoScrollToBottom = distanceFromBottom < 100;
    showScrollToBottom = distanceFromBottom > 160;

    if (autoScrollToBottom) {
      unreadActivityBelow = false;
    }

    // Acknowledge finished turn when viewing the bottom of conversation (only reset from 'finished' to 'idle', never interrupt 'working')
    if (distanceFromBottom < 100) {
      const activeSession = sessionStore.activeSession;
      if (activeSession && activeSession.status === 'finished') {
        sessionStore.setSessionStatus(activeSession.id, 'idle');
      }
    }

    // Trigger progressive prepend if user scrolls within 200px of top
    if (scrollTop < 200 && sessionStore.remainingHiddenTurns > 0 && !isHydrating) {
      loadEarlier();
    }
  }

  // Smooth scroll to bottom function
  function scrollToBottom() {
    if (!containerEl) return;
    autoScrollToBottom = true;
    unreadActivityBelow = false;
    containerEl.scrollTo({
      top: containerEl.scrollHeight,
      behavior: 'smooth'
    });
  }

  // Auto-scroll on new streaming messages or DOM updates if locked to bottom
  $effect(() => {
    const msgs = sessionStore.visibleMessages;
    // Track messages length and last message content/tools changes
    const lastMsg = msgs[msgs.length - 1];
    const _trackContent = lastMsg?.content;
    const _trackTools = lastMsg?.toolCalls?.length;
    const _trackStatus = lastMsg?.status;

    if (msgs.length > 0) {
      if (autoScrollToBottom) {
        tick().then(() => {
          if (containerEl && autoScrollToBottom) {
            containerEl.scrollTop = containerEl.scrollHeight;
          }
        });
      } else {
        // User is scrolled up and new message/content arrived
        unreadActivityBelow = true;
      }
    }
  });

  onMount(() => {
    // Initial scroll to bottom
    if (containerEl) {
      containerEl.scrollTop = containerEl.scrollHeight;
    }

    // Use ResizeObserver on contentWrapperEl to instantly follow streaming text/tool UI height expansions when pinned to bottom
    let resizeObserver: ResizeObserver | null = null;
    if (typeof ResizeObserver !== 'undefined' && contentWrapperEl) {
      resizeObserver = new ResizeObserver(() => {
        if (autoScrollToBottom && containerEl) {
          containerEl.scrollTop = containerEl.scrollHeight;
        }
      });
      resizeObserver.observe(contentWrapperEl);
    }

    // Set up IntersectionObserver on top sentinel for smooth upward infinite hydration
    let intersectionObserver: IntersectionObserver | null = null;
    if ('IntersectionObserver' in window && topSentinelEl && containerEl) {
      intersectionObserver = new IntersectionObserver(
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

      intersectionObserver.observe(topSentinelEl);
    }

    return () => {
      if (resizeObserver) resizeObserver.disconnect();
      if (intersectionObserver) intersectionObserver.disconnect();
    };
  });
</script>

<div
  bind:this={containerEl}
  onscroll={handleScroll}
  class="flex-1 w-full h-full overflow-y-auto px-4 py-3 space-y-3 relative bg-ant-bg select-text"
>
  <div bind:this={contentWrapperEl} class="w-full space-y-3">
    <!-- Top Sentinel & Prepend History Header -->
    <div bind:this={topSentinelEl} class="w-full flex justify-center py-2">
      {#if sessionStore.remainingHiddenTurns > 0}
        <button
          type="button"
          onclick={loadEarlier}
          disabled={isHydrating}
          class="inline-flex items-center space-x-2 px-3 py-1.5 rounded-full text-xs font-medium bg-ant-bg-secondary hover:bg-ant-bg-tertiary text-ant-primary border border-white/10 shadow-sm transition-all duration-150 disabled:opacity-50"
        >
          {#if isHydrating}
            <Loader2 size={13} class="animate-spin text-ant-primary" />
            <span>Hydrating turns...</span>
          {:else}
            <ArrowUp size={13} class="text-ant-primary" />
            <span>Show previous ({sessionStore.remainingHiddenTurns} earlier turns)</span>
          {/if}
        </button>
      {:else if sessionStore.visibleMessages.length > 0}
        <div class="text-[11px] text-ant-text-muted font-mono flex items-center gap-1.5 py-1 select-none">
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
          isLastUserTurn={message.id === lastUserMessageId}
          onEditLastTurn={onEditLastTurn}
          onOpenImage={handleOpenImage}
          onPlanAction={onPlanAction}
        />
      {/each}

      <!-- Live In-Transcript Compaction Indicator Banner -->
      {#if sessionStore.isCompacting}
        <div class="my-2.5 p-3 rounded-xl border border-zinc-800 bg-zinc-900/90 backdrop-blur-md shadow-lg flex items-center justify-between font-sans select-none animate-in fade-in duration-200">
          <div class="flex items-center space-x-3 min-w-0">
            <div class="w-7 h-7 rounded-lg bg-indigo-500/15 flex items-center justify-center text-indigo-400 flex-shrink-0 animate-spin">
              <Loader2 size={14} />
            </div>
            <div class="flex flex-col min-w-0">
              <span class="text-xs font-semibold text-zinc-100 truncate flex items-center gap-1.5">
                <span>Compacting conversation context</span>
              </span>
              <span class="text-[11px] text-zinc-400 mt-0.5">
                Summarizing earlier conversation turns to optimize context window
              </span>
            </div>
          </div>

          <div class="px-2.5 py-1 rounded-md bg-indigo-500/10 text-indigo-300 text-[11px] font-mono flex-shrink-0 ml-3 border border-indigo-500/20">
            <span class="font-medium">In progress...</span>
          </div>
        </div>
      {:else if sessionStore.lastCompactNotice}
        <div class="my-2.5 p-3 rounded-xl border flex items-center justify-between font-sans select-none text-xs transition-all duration-300 {sessionStore.lastCompactNotice.type === 'success' ? 'border-emerald-500/20 bg-zinc-900/90 text-emerald-300' : 'border-rose-500/20 bg-zinc-900/90 text-rose-300'} shadow-md">
          <div class="flex items-center space-x-2.5">
            {#if sessionStore.lastCompactNotice.type === 'success'}
              <CheckCircle2 size={15} class="flex-shrink-0 text-emerald-400" />
            {:else}
              <AlertCircle size={15} class="flex-shrink-0 text-rose-400" />
            {/if}
            <span class="text-xs font-medium text-zinc-200">{sessionStore.lastCompactNotice.message}</span>
          </div>
          <button
            type="button"
            onclick={() => sessionStore.lastCompactNotice = null}
            class="text-[10.5px] text-zinc-400 hover:text-zinc-200 uppercase tracking-wider ml-3 px-2 py-0.5 rounded bg-zinc-800/60 hover:bg-zinc-800 transition"
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
            <span class="font-medium">{formattedElapsed}</span>
          </div>
        </div>
      {/if}
    {/if}
  </div>

  <!-- Floating Scroll To Bottom Button -->
  {#if showScrollToBottom}
    <div class="sticky bottom-3 right-4 flex justify-end pointer-events-none z-30 mr-2">
      <button
        type="button"
        onclick={scrollToBottom}
        class="pointer-events-auto flex items-center space-x-2 px-3 py-1.5 rounded-full bg-ant-bg-secondary/95 hover:bg-ant-bg-tertiary text-ant-text border border-white/10 shadow-xl backdrop-blur-md transition-all duration-200 hover:scale-105 active:scale-95 group"
        title="Scroll to bottom"
      >
        {#if unreadActivityBelow}
          <span class="relative flex h-2 w-2">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-ant-primary opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2 w-2 bg-ant-primary"></span>
          </span>
          <span class="text-xs font-serif font-medium text-ant-primary">New activity</span>
        {:else}
          <span class="text-xs font-serif text-ant-text-secondary group-hover:text-ant-text transition-colors">Scroll to bottom</span>
        {/if}
        <ArrowDown size={13} class="text-ant-primary transition-transform group-hover:translate-y-0.5" />
      </button>
    </div>
  {/if}

  <!-- Fullscreen Image Lightbox Modal -->
  <ImageLightboxModal
    visible={lightboxVisible}
    imageSrc={lightboxSrc}
    imageTitle={lightboxTitle}
    onClose={() => lightboxVisible = false}
  />
</div>
