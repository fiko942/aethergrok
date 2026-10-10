<script lang="ts">
  import { onMount } from 'svelte';
  import type { ChatMessage } from '$lib/stores/session.svelte';
  import ToolCallCard from './ToolCallCard.svelte';
  import TurnDiffSummary from './TurnDiffSummary.svelte';
  import PlanReviewCard from './PlanReviewCard.svelte';
  import { calculateDiffStat } from '$lib/utils/diffUtils';
  import { sessionStore } from '$lib/stores/session.svelte';
  import { dialogStore } from '$lib/stores/dialog.svelte';
  import {
    renderMarkdown,
    extractCandidateFilePaths,
    type FilePathInfo
  } from '$lib/utils/markdownRenderer';
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
    Edit3,
    AlertTriangle,
    RefreshCw,
    Sparkles,
    PlusCircle,
    ShieldAlert
  } from 'lucide-svelte';

  interface Props {
    message: ChatMessage;
    turnNumber?: number;
    isLastUserTurn?: boolean;
    onEditLastTurn?: () => void;
    onOpenImage?: (src: string, title?: string) => void;
    onPlanAction?: (action: 'approve' | 'reject' | 'custom', feedback?: string) => void;
    onRetryTurn?: (message: ChatMessage) => void;
  }

  let { message, turnNumber, isLastUserTurn = false, onEditLastTurn, onOpenImage, onPlanAction, onRetryTurn }: Props = $props();

  let showReasoning = $state(false);
  let showToolDetails = $state(true);
  let copied = $state(false);

  // Map of candidate file paths that actually exist on disk
  let existingFilesMap = $state<Record<string, FilePathInfo>>({});

  // Active workspace directory for path resolution
  const currentWorkspacePath = $derived.by(() => {
    const activeWs = sessionStore.activeWorkspace;
    return activeWs?.path || '';
  });

  // Query backend bridge whenever message content or workspace changes
  $effect(() => {
    const content = message.content;
    const wsPath = currentWorkspacePath;
    if (!content) {
      existingFilesMap = {};
      return;
    }

    const candidates = extractCandidateFilePaths(content);
    if (candidates.length === 0) {
      existingFilesMap = {};
      return;
    }

    const win = window as any;
    if (win.go?.main?.App?.CheckMultipleFilesExists) {
      win.go.main.App.CheckMultipleFilesExists(wsPath, candidates)
        .then((res: Record<string, FilePathInfo>) => {
          if (res) {
            existingFilesMap = res;
          }
        })
        .catch(() => {
          // Fallback or ignore error
        });
    }
  });

  // Check if this assistant message is presenting a pending execution plan
  const isPlanProposal = $derived.by(() => {
    if (message.role !== 'assistant') return false;
    const lower = (message.content || '').toLowerCase();
    const hasPlanHeading = lower.includes('### execution plan') || lower.includes('## plan') || lower.includes('### plan') || lower.includes('rancangan perencanaan') || lower.includes('implementation plan');
    const hasExitPlanTool = message.toolCalls?.some(tc => tc.tool.includes('exit_plan_mode') || tc.tool.includes('plan'));
    return (hasPlanHeading || hasExitPlanTool) && message.status === 'done';
  });

  // Check if message is a response truncated by max_tokens or context limit
  const isTruncationError = $derived.by(() => {
    return (
      message.errorKind === 'max_tokens_truncation' ||
      (message.content &&
        (message.content.includes('max_tokens') ||
          message.content.includes('Context limit reached') ||
          message.content.includes('response truncated')))
    );
  });

  const autoRetriesCount = $derived(sessionStore.activeSession?.autoRetryCount || 0);
  const isAutoResuming = $derived(
    sessionStore.activeSession?.status === 'working' ||
    (autoRetriesCount > 0 && autoRetriesCount < 3 && sessionStore.activeSession?.status !== 'error')
  );

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

  function handleMessageClick(e: MouseEvent) {
    const target = e.target as HTMLElement;

    // 1. Handle clicking on links (open in default OS browser instead of navigating webview)
    const linkEl = target.closest('a') as HTMLAnchorElement | null;
    if (linkEl && linkEl.href) {
      e.preventDefault();
      e.stopPropagation();
      const href = linkEl.href;
      const win = window as any;
      if (win.go?.main?.App?.OpenExternalURL) {
        win.go.main.App.OpenExternalURL(href);
      } else if (win.runtime?.BrowserOpenURL) {
        win.runtime.BrowserOpenURL(href);
      } else {
        window.open(href, '_blank');
      }
      return;
    }

    // 2. Handle clicking on verified file chip/link
    const fileChip = target.closest('.inline-file-chip') as HTMLElement;
    if (fileChip) {
      e.preventDefault();
      e.stopPropagation();
      const rawPath = decodeURIComponent(fileChip.dataset.filePath || '');
      const isDir = fileChip.dataset.isDir === 'true';

      if (rawPath) {
        if (isDir) {
          // If directory, open in native file manager (macOS Finder or Windows Explorer)
          const win = window as any;
          if (win.go?.main?.App?.OpenPathInSystem) {
            win.go.main.App.OpenPathInSystem(rawPath);
          }
        } else {
          // If regular file, open in built-in modal FileViewer
          dialogStore.openFileViewer(rawPath, currentWorkspacePath);
        }
      }
      return;
    }

    // 3. Handle clicking copy code block button
    const copyBtn = target.closest('.copy-code-btn') as HTMLElement;
    if (copyBtn) {
      const rawCode = decodeURIComponent(copyBtn.dataset.rawCode || '');
      if (rawCode) {
        navigator.clipboard.writeText(rawCode).then(() => {
          const label = copyBtn.querySelector('.copy-label');
          if (label) {
            const orig = label.textContent;
            label.textContent = 'Copied!';
            copyBtn.classList.add('text-ant-success');
            setTimeout(() => {
              label.textContent = orig;
              copyBtn.classList.remove('text-ant-success');
            }, 1800);
          }
        });
      }
    }
  }

  // Render comprehensive Markdown formatted text with syntax highlighting, diagram detection, and dynamic file chips
  const renderedHtml = $derived.by(() => {
    if (!message.content) return '';
    return renderMarkdown(message.content, existingFilesMap);
  });

  // DOM Virtualization & Viewport Pruning State
  let itemContainerEl = $state<HTMLDivElement | null>(null);
  let isInView = $state(true);
  let lastMeasuredHeight = $state<number | null>(null);

  // Viewport tolerance buffer: 800px margin above and below window
  const VIEWPORT_TOLERANCE_PX = 800;

  // Streaming or active messages must always stay rendered to preserve live streaming
  const isStreamingOrActive = $derived(
    message.status === 'streaming' ||
    (sessionStore.activeSession?.status === 'working' &&
      sessionStore.visibleMessages[sessionStore.visibleMessages.length - 1]?.id === message.id)
  );

  const shouldRenderChildren = $derived(
    isInView || isStreamingOrActive || lastMeasuredHeight === null || lastMeasuredHeight <= 0
  );

  onMount(() => {
    let resizeObserver: ResizeObserver | null = null;
    let intersectionObserver: IntersectionObserver | null = null;
    let rAFId: number | null = null;

    if (itemContainerEl) {
      const h = itemContainerEl.offsetHeight;
      if (h > 0) {
        lastMeasuredHeight = h;
      }

      if (typeof ResizeObserver !== 'undefined') {
        resizeObserver = new ResizeObserver(() => {
          if (rAFId !== null) cancelAnimationFrame(rAFId);
          rAFId = requestAnimationFrame(() => {
            rAFId = null;
            if (shouldRenderChildren && itemContainerEl) {
              const currentHeight = itemContainerEl.offsetHeight;
              if (currentHeight > 0 && (lastMeasuredHeight === null || Math.abs(currentHeight - lastMeasuredHeight) > 1)) {
                lastMeasuredHeight = currentHeight;
              }
            }
          });
        });
        resizeObserver.observe(itemContainerEl);
      }

      if (typeof IntersectionObserver !== 'undefined') {
        intersectionObserver = new IntersectionObserver(
          (entries) => {
            const entry = entries[0];
            if (entry) {
              isInView = entry.isIntersecting;
            }
          },
          {
            root: null, // window viewport
            rootMargin: `${VIEWPORT_TOLERANCE_PX}px 0px ${VIEWPORT_TOLERANCE_PX}px 0px`,
            threshold: 0
          }
        );
        intersectionObserver.observe(itemContainerEl);
      }
    }

    return () => {
      if (rAFId !== null) cancelAnimationFrame(rAFId);
      if (resizeObserver) resizeObserver.disconnect();
      if (intersectionObserver) intersectionObserver.disconnect();
    };
  });
</script>

<div
  bind:this={itemContainerEl}
  class="w-full relative"
  style="content-visibility: auto; contain-intrinsic-size: auto {lastMeasuredHeight ? `${lastMeasuredHeight}px` : (message.role === 'user' ? '60px' : '160px')}; min-height: {lastMeasuredHeight ? `${lastMeasuredHeight}px` : 'auto'};"
>
  {#if shouldRenderChildren}
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
            {#if message.toolCalls && message.toolCalls.some(tc => {
              const stat = calculateDiffStat(tc);
              return stat.added > 0 || stat.removed > 0 || !!tc.diff;
            })}
              <TurnDiffSummary toolCalls={message.toolCalls} />
            {/if}

            <!-- Message Markdown Content with Anthropic Serif Editorial Typography -->
            {#if message.content}
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div
                onclick={handleMessageClick}
                class="font-serif text-[14px] text-ant-text leading-[1.7] tracking-normal select-text markdown-chat-body max-w-none"
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

            <!-- Token Truncation & Error Recovery Card -->
            {#if isTruncationError}
              <div class="mt-3 p-3.5 rounded-xl border border-amber-500/30 bg-amber-500/[0.04] dark:bg-amber-950/20 text-ant-text space-y-3 font-sans shadow-sm select-none">
                <!-- Header -->
                <div class="flex items-start space-x-2.5">
                  <div class="p-1.5 rounded-lg bg-amber-500/15 text-amber-500 flex-shrink-0 mt-0.5">
                    <AlertTriangle size={15} />
                  </div>
                  <div class="flex-1 min-w-0">
                    <div class="flex items-center justify-between">
                      <span class="text-xs font-semibold text-amber-500 dark:text-amber-400">
                        Response Truncated by Token Limit (max_tokens)
                      </span>
                      {#if message.errorDetails?.totalTokens || message.tokens?.total}
                        <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20">
                          {((message.errorDetails?.totalTokens || message.tokens?.total || 0)).toLocaleString()} accumulated tokens
                        </span>
                      {/if}
                    </div>
                    <p class="text-[11.5px] text-ant-text-secondary mt-1 leading-relaxed select-text">
                      The model reached the token limit after extended iterations. In accordance with queue safety policies, upcoming tasks have been <strong>safely paused</strong> and will not run until this is resolved.
                    </p>
                  </div>
                </div>

                <!-- Usage Statistics Grid -->
                {#if message.errorDetails}
                  <div class="grid grid-cols-2 sm:grid-cols-4 gap-2 pt-1 border-t border-amber-500/15 text-[10.5px] font-mono">
                    {#if message.errorDetails.numTurns}
                      <div class="p-1.5 rounded bg-black/20 border border-white/5 flex flex-col">
                        <span class="text-ant-text-muted text-[9.5px]">Tool Turns</span>
                        <span class="text-ant-text font-semibold">{message.errorDetails.numTurns} turns</span>
                      </div>
                    {/if}
                    {#if message.errorDetails.modelCalls}
                      <div class="p-1.5 rounded bg-black/20 border border-white/5 flex flex-col">
                        <span class="text-ant-text-muted text-[9.5px]">Model Calls</span>
                        <span class="text-ant-text font-semibold">{message.errorDetails.modelCalls} calls</span>
                      </div>
                    {/if}
                    {#if message.errorDetails.reasoningTokens}
                      <div class="p-1.5 rounded bg-black/20 border border-white/5 flex flex-col">
                        <span class="text-ant-text-muted text-[9.5px]">Reasoning Tokens</span>
                        <span class="text-ant-text font-semibold">{message.errorDetails.reasoningTokens.toLocaleString()}</span>
                      </div>
                    {/if}
                    {#if message.errorDetails.totalTokens}
                      <div class="p-1.5 rounded bg-black/20 border border-white/5 flex flex-col">
                        <span class="text-ant-text-muted text-[9.5px]">Total Tokens</span>
                        <span class="text-ant-text font-semibold">{message.errorDetails.totalTokens.toLocaleString()}</span>
                      </div>
                    {/if}
                  </div>
                {/if}

                <!-- Queued Prompts Protection Banner -->
                {#if (sessionStore.activeSession?.queuedPrompts?.length || 0) > 0}
                  <div class="px-2.5 py-1.5 rounded bg-amber-500/10 border border-amber-500/20 text-[11px] text-amber-400 flex items-center space-x-2">
                    <ShieldAlert size={12} class="flex-shrink-0" />
                    <span><strong>Queue Protected:</strong> {sessionStore.activeSession?.queuedPrompts?.length} queued prompt(s) safely retained and paused.</span>
                  </div>
                {/if}

                <!-- Auto-Resuming Notice vs Manual Recovery Card -->
                {#if isAutoResuming}
                  <div class="flex items-center justify-between p-2.5 rounded-lg bg-amber-500/10 border border-amber-500/25 text-amber-400 text-xs">
                    <div class="flex items-center space-x-2">
                      <RefreshCw size={13} class="animate-spin text-amber-400 flex-shrink-0" />
                      <span class="font-medium">Auto-resuming task in progress (attempt {Math.max(1, autoRetriesCount)}/3)...</span>
                    </div>
                    <span class="text-[11px] text-amber-400/80">Continuing previous turn automatically</span>
                  </div>
                {:else}
                  <!-- Exhausted Retries Notice & Manual Recovery Actions -->
                  <div class="px-2.5 py-1.5 rounded bg-red-500/10 border border-red-500/25 text-[11.5px] text-red-400 flex items-center space-x-2">
                    <AlertTriangle size={13} class="flex-shrink-0 text-red-400" />
                    <span>Automatic continuation reached the maximum limit (3/3 attempts). Manual action is required to proceed.</span>
                  </div>

                  <!-- Interactive Action Buttons -->
                  <div class="flex items-center flex-wrap gap-2 pt-1">
                    <button
                      type="button"
                      onclick={() => onRetryTurn?.(message)}
                      class="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-amber-500 hover:bg-amber-600 text-white font-medium text-xs shadow-sm transition"
                      title="Resume the truncated task without losing progress"
                    >
                      <RefreshCw size={12} />
                      <span>Retry / Continue Task</span>
                    </button>

                    <button
                      type="button"
                      onclick={async () => { await sessionStore.compactActiveSession(); }}
                      class="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 border border-white/10 font-medium text-xs transition"
                      title="Compact conversation history to free up context space"
                    >
                      <Sparkles size={12} class="text-indigo-400" />
                      <span>Compact Context</span>
                    </button>

                    <button
                      type="button"
                      onclick={() => {
                        if (sessionStore.activeWorkspace) {
                          sessionStore.createSession('New Session', sessionStore.activeWorkspace.id);
                        }
                      }}
                      class="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 border border-white/10 font-medium text-xs transition"
                      title="Start a fresh new session in this workspace"
                    >
                      <PlusCircle size={12} class="text-emerald-400" />
                      <span>New Session</span>
                    </button>
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        </div>
      </div>
    {/if}
  {:else}
    <!-- Offscreen Virtual Spacer: retains exact measured height so scrollbar and scroll position remain completely stable -->
    <div
      class="w-full flex items-center justify-center opacity-0 pointer-events-none select-none my-2.5 px-2"
      style="height: {lastMeasuredHeight}px; min-height: {lastMeasuredHeight}px;"
      aria-hidden="true"
    ></div>
  {/if}
</div>
