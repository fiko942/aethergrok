<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import type { ToolCall } from '$lib/stores/session.svelte';
  import { planStore } from '$lib/stores/plan.svelte';
  import { sessionStore } from '$lib/stores/session.svelte';
  import {
    parseTodosUpdated,
    calculatePlanMetrics,
    type SubagentPlanState,
    renderTodoContent
  } from '$lib/utils/planParser';
  import {
    type SubagentParsedMeta,
    extractSubagentMeta,
    extractSubagentIdFromText,
    cleanSubagentOutput,
    isSubagentBackgroundAck,
    formatSubagentDuration,
    formatActionSummary,
    extractActionTodos
  } from '$lib/utils/toolUtils';
  import { renderMarkdown } from '$lib/utils/markdownRenderer';
  import {
    Bot,
    X,
    Clock,
    Copy,
    Check,
    Loader2,
    CheckCircle2,
    XCircle,
    Cpu,
    Shield,
    Code,
    FileText,
    ListTodo,
    Activity,
    ChevronDown,
    ChevronRight
  } from 'lucide-svelte';

  interface Props {
    visible: boolean;
    toolCall: ToolCall;
    meta?: SubagentParsedMeta;
    onClose: () => void;
  }

  let { visible = false, toolCall, meta, onClose }: Props = $props();

  let copiedPrompt = $state(false);
  let copiedOutput = $state(false);
  let now = $state(Date.now());
  let showRawPayload = $state(false);
  let expandedChildIds = $state<Record<string, boolean>>({});
  let copiedChildId = $state<string | null>(null);
  let localChildCalls = $state<ToolCall[]>([]);
  let localSubagentStatus = $state<'running' | 'completed' | 'error' | null>(null);
  let localSubagentOutput = $state<string>('');
  let localSubagentDurationMs = $state<number | null>(null);
  let activityContainerEl = $state<HTMLElement | null>(null);
  let userHasScrolledActivity = $state(false);
  let prevActivityCount = $state(0);

  function handleActivityScroll() {
    if (!activityContainerEl) return;
    const { scrollTop, scrollHeight, clientHeight } = activityContainerEl;
    const distanceFromBottom = scrollHeight - (scrollTop + clientHeight);
    userHasScrolledActivity = distanceFromBottom > 35;
  }

  const displayChildCalls = $derived(
    localChildCalls.length > 0 ? localChildCalls : (toolCall.childToolCalls || [])
  );

  const effectiveMeta = $derived.by(() => {
    const extracted = extractSubagentMeta(toolCall.params, toolCall.result);
    return {
      description: extracted.description || meta?.description || 'Subagent Task',
      prompt: extracted.prompt || meta?.prompt || '',
      subagentType: extracted.subagentType || meta?.subagentType || 'general-purpose',
      model: extracted.model || meta?.model,
      isolation: extracted.isolation || meta?.isolation,
      isBackground: extracted.isBackground ?? meta?.isBackground ?? true,
      taskId: extracted.taskId || meta?.taskId || toolCall.childSessionId
    };
  });

  function toggleChildExpand(id: string) {
    expandedChildIds[id] = !expandedChildIds[id];
  }

  async function loadChildToolCalls() {
    const sid = toolCall.childSessionId || effectiveMeta.taskId || extractSubagentIdFromText(toolCall.result) || toolCall.id;
    if (!sid) return;

    if (!toolCall.childSessionId && sid && sid !== toolCall.id) {
      toolCall.childSessionId = sid;
    }

    // Resolve workspace path from sessionStore
    const activeSession = sessionStore.sessions.find((s) => s.id === sessionStore.activeSessionId);
    let wsPath = '';
    if (activeSession?.workspaceId) {
      const ws = sessionStore.workspaces.find((w) => w.id === activeSession.workspaceId);
      if (ws?.path) wsPath = ws.path;
    }
    if (!wsPath && sessionStore.activeWorkspace?.path) {
      wsPath = sessionStore.activeWorkspace.path;
    }

    if (window.go?.main?.App?.LoadSubagentDetails) {
      try {
        const details = await window.go.main.App.LoadSubagentDetails(wsPath, sid);
        if (details) {
          if (details.calls && Array.isArray(details.calls) && details.calls.length > 0) {
            const mapped: ToolCall[] = details.calls.map((c: any) => ({
              id: c.id,
              tool: c.tool,
              params: c.params,
              result: c.result,
              status: c.status === 'failed' ? 'error' : c.status === 'completed' ? 'completed' : 'running',
              startTime: c.startTime,
              endTime: c.endTime
            }));
            toolCall.childToolCalls = mapped;
            localChildCalls = mapped;

            // Inspect child tool calls for any subagent plan updates (e.g. todo_write)
            for (const c of details.calls) {
              planStore.inspectToolForPlan(
                sessionStore.activeSessionId || '',
                c.id,
                c.tool,
                c.params,
                c.result,
                { id: sid, name: effectiveMeta.description }
              );
            }
          }

          if (details.status === 'completed' || details.status === 'failed') {
            localSubagentStatus = details.status === 'failed' ? 'error' : 'completed';
            toolCall.status = localSubagentStatus;
            if (details.output) {
              localSubagentOutput = details.output;
              if (!toolCall.result || isSubagentBackgroundAck(toolCall.result)) {
                toolCall.result = details.output;
              }
            }
            if (details.startTime && !toolCall.startTime) {
              toolCall.startTime = details.startTime;
            }
            if (details.endTime && !toolCall.endTime) {
              toolCall.endTime = details.endTime;
            } else if (details.durationMs && details.durationMs > 0 && toolCall.startTime) {
              toolCall.endTime = toolCall.startTime + details.durationMs;
            }
            if (details.durationMs && details.durationMs > 0) {
              localSubagentDurationMs = details.durationMs;
            }
          }
          if (typeof toolCall.params === 'string') {
            try {
              toolCall.params = JSON.parse(toolCall.params);
            } catch {
              toolCall.params = { raw: toolCall.params };
            }
          }
          if (!toolCall.params || typeof toolCall.params !== 'object') {
            toolCall.params = {};
          }
          const modalP = toolCall.params as Record<string, any>;
          if (details.description && !modalP.description) {
            modalP.description = details.description;
          }
          if (details.prompt && !modalP.prompt) {
            modalP.prompt = details.prompt;
          }
          if (details.subagentType && !modalP.subagent_type) {
            modalP.subagent_type = details.subagentType;
          }
          if (details.model && !modalP.model) {
            modalP.model = details.model;
          }
          if (details.subagentId && !toolCall.childSessionId) {
            toolCall.childSessionId = details.subagentId;
          }
          sessionStore.syncToolCallUpdate(toolCall);
        }
      } catch (err) {
        console.error('Failed to load subagent execution details:', err);
      }
    } else if (window.go?.main?.App?.LoadSubagentToolCalls) {
      try {
        const calls = await window.go.main.App.LoadSubagentToolCalls(wsPath, sid);
        if (calls && Array.isArray(calls) && calls.length > 0) {
          const mapped: ToolCall[] = calls.map((c: any) => ({
            id: c.id,
            tool: c.tool,
            params: c.params,
            result: c.result,
            status: c.status === 'failed' ? 'error' : c.status === 'completed' ? 'completed' : 'running',
            startTime: c.startTime,
            endTime: c.endTime
          }));
          toolCall.childToolCalls = mapped;
          localChildCalls = mapped;
          sessionStore.syncToolCallUpdate(toolCall);

          // Inspect child tool calls for any subagent plan updates (e.g. todo_write)
          for (const c of calls) {
            planStore.inspectToolForPlan(
              sessionStore.activeSessionId || '',
              c.id,
              c.tool,
              c.params,
              c.result,
              { id: sid, name: effectiveMeta.description }
            );
          }
        }
      } catch (err) {
        console.error('Failed to load subagent tool calls:', err);
      }
    }
  }

  $effect(() => {
    if (!visible) return;
    loadChildToolCalls();
    if (isActuallyRunning) {
      const interval = setInterval(loadChildToolCalls, 1500);
      return () => clearInterval(interval);
    }
  });

  // Auto-scroll activity stream list to bottom when modal is opened or when new actions arrive
  $effect(() => {
    if (!visible) {
      userHasScrolledActivity = false;
      prevActivityCount = 0;
      return;
    }

    const currentCount = displayChildCalls.length;
    if (currentCount > 0) {
      const isNewItem = currentCount !== prevActivityCount;
      prevActivityCount = currentCount;

      if (!userHasScrolledActivity || isActuallyRunning) {
        setTimeout(() => {
          if (activityContainerEl) {
            activityContainerEl.scrollTo({
              top: activityContainerEl.scrollHeight,
              behavior: isNewItem && prevActivityCount > 1 ? 'smooth' : 'auto'
            });
          }
        }, 30);
      }
    }
  });

  async function copyText(text: string, id: string) {
    if (!text) return;
    try {
      await navigator.clipboard.writeText(text);
      copiedChildId = id;
      setTimeout(() => {
        if (copiedChildId === id) copiedChildId = null;
      }, 2000);
    } catch {
      // ignore
    }
  }

  function getChildDuration(child: ToolCall): string | null {
    if (!child.startTime) return null;
    const end = child.endTime || (child.status === 'running' ? now : child.startTime);
    const ms = Math.max(0, end - child.startTime);
    if (ms < 1000) return `${ms}ms`;
    const secs = ms / 1000;
    if (secs < 60) return `${secs.toFixed(1)}s`;
    const mins = Math.floor(secs / 60);
    const rem = Math.round(secs % 60);
    return `${mins}m ${rem}s`;
  }

  const isBackgroundAck = $derived.by(() => {
    return isSubagentBackgroundAck(toolCall.result);
  });

  const cleanedOutput = $derived.by(() => {
    const fromCall = cleanSubagentOutput(toolCall.result);
    if (fromCall) return fromCall;
    if (localSubagentOutput) return cleanSubagentOutput(localSubagentOutput);
    return '';
  });

  const effectiveStatus = $derived.by(() => {
    if (localSubagentStatus) return localSubagentStatus;
    if (cleanedOutput && toolCall.status !== 'error') return 'completed';
    return toolCall.status;
  });

  const isActuallyRunning = $derived.by(() => {
    if (effectiveStatus === 'completed' || effectiveStatus === 'error') {
      return false;
    }
    return toolCall.status === 'running' || (isBackgroundAck && !cleanedOutput) || displayChildCalls.some(c => c.status === 'running');
  });

  // Active ticking timer for running subagents
  $effect(() => {
    if (!visible || !isActuallyRunning) return;
    const interval = setInterval(() => {
      now = Date.now();
    }, 1000);
    return () => clearInterval(interval);
  });

  const elapsedDuration = $derived.by(() => {
    if (localSubagentDurationMs && localSubagentDurationMs > 0 && !isActuallyRunning) {
      return formatSubagentDuration(localSubagentDurationMs);
    }
    if (toolCall.startTime) {
      let end: number;
      if (isActuallyRunning) {
        end = now;
      } else {
        const lastChildEnd = (displayChildCalls && displayChildCalls.length > 0)
          ? Math.max(0, ...displayChildCalls.map(c => c.endTime || c.startTime || 0))
          : 0;
        end = Math.max(toolCall.endTime || 0, lastChildEnd) || now;
      }
      return formatSubagentDuration(Math.max(0, end - toolCall.startTime));
    }
    return null;
  });

  const subagentPlan = $derived.by<SubagentPlanState | null>(() => {
    const sid = effectiveMeta.taskId || toolCall.childSessionId || toolCall.id;
    const currentSessionId = sessionStore.activeSessionId;
    if (sid && currentSessionId) {
      const p = planStore.getSubagentPlan(currentSessionId, sid);
      if (p && p.todos && p.todos.length > 0) {
        return p;
      }
    }

    // Fallback: If not in planStore yet, inspect toolCall.result or effectiveMeta.prompt
    const fallbackTodos = parseTodosUpdated(toolCall.result) || parseTodosUpdated(effectiveMeta.prompt);
    if (fallbackTodos && fallbackTodos.length > 0) {
      const metrics = calculatePlanMetrics(fallbackTodos);
      return {
        subagentId: sid || 'subagent',
        subagentName: effectiveMeta.description,
        todos: fallbackTodos,
        completedCount: metrics.completedCount,
        inProgressCount: metrics.inProgressCount,
        pendingCount: metrics.pendingCount,
        totalCount: metrics.totalCount,
        progressPercent: metrics.progressPercent,
        lastUpdated: Date.now()
      };
    }

    return null;
  });

  function resolveTodoDisplay(todo: { id: string; content?: string }, knownList?: Array<{ id: string; content?: string }>): string {
    const id = todo.id || '';
    if (todo.content && todo.content.trim() && todo.content !== id) {
      return todo.content.trim();
    }
    if (knownList && knownList.length > 0) {
      const match = knownList.find((t) => t.id === id);
      if (match && match.content && match.content !== id) {
        return match.content;
      }
    }
    const activePlan = planStore.getPlan(sessionStore.activeSessionId);
    if (activePlan?.todos) {
      const match = activePlan.todos.find((t) => t.id === id);
      if (match && match.content && match.content !== id) {
        return match.content;
      }
    }
    return id || 'task';
  }

  async function copyPrompt() {
    if (!effectiveMeta.prompt) return;
    try {
      await navigator.clipboard.writeText(effectiveMeta.prompt);
      copiedPrompt = true;
      setTimeout(() => (copiedPrompt = false), 2000);
    } catch {
      // ignore
    }
  }

  async function copyOutput() {
    if (!cleanedOutput) return;
    try {
      await navigator.clipboard.writeText(cleanedOutput);
      copiedOutput = true;
      setTimeout(() => (copiedOutput = false), 2000);
    } catch {
      // ignore
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (!visible) return;
    if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeyDown);
  });
</script>

{#if visible}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-[1000] flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-150 select-text font-sans"
    onclick={(e) => {
      if (e.target === e.currentTarget) onClose();
    }}
    onkeydown={handleKeyDown}
  >
    <!-- Modal Card Container -->
    <div
      class="w-full max-w-4xl max-h-[90vh] flex flex-col bg-ant-bg border border-violet-500/40 rounded-xl shadow-2xl shadow-violet-950/40 overflow-hidden animate-in zoom-in-95 duration-150"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-5 py-3.5 bg-ant-bg-secondary/80 border-b border-white/5 select-none">
        <div class="flex items-center space-x-3 min-w-0">
          <div class="w-8 h-8 rounded-lg bg-violet-500/15 border border-violet-500/30 flex items-center justify-center flex-shrink-0 text-violet-400">
            {#if isActuallyRunning}
              <Loader2 size={18} class="animate-spin text-violet-400" />
            {:else if effectiveStatus === 'completed'}
              <Bot size={18} class="text-violet-400" />
            {:else}
              <Bot size={18} class="text-rose-400" />
            {/if}
          </div>
          <div class="min-w-0">
            <div class="flex items-center space-x-2">
              <span class="text-xs font-semibold uppercase tracking-wider text-violet-400 font-mono">Subagent Process</span>
              {#if elapsedDuration !== null}
                <span class="inline-flex items-center text-[11px] font-mono font-medium px-2 py-0.5 rounded-full bg-violet-500/10 text-violet-300 border border-violet-500/20">
                  <Clock size={11} class="mr-1" />
                  {elapsedDuration}
                </span>
              {/if}
            </div>
            <h2 class="text-sm font-semibold text-ant-text truncate mt-0.5">
              {effectiveMeta.description}
            </h2>
          </div>
        </div>

        <!-- Right: Status Badge & Close Button -->
        <div class="flex items-center space-x-3 flex-shrink-0">
          {#if isActuallyRunning}
            <span class="inline-flex items-center text-xs font-medium text-violet-400 animate-pulse bg-violet-500/10 px-2.5 py-1 rounded-md border border-violet-500/20">
              <Loader2 size={12} class="animate-spin mr-1.5" />
              Running
            </span>
          {:else if effectiveStatus === 'completed'}
            <span class="inline-flex items-center text-xs font-medium text-emerald-400 bg-emerald-500/10 px-2.5 py-1 rounded-md border border-emerald-500/20">
              <CheckCircle2 size={12} class="mr-1.5" />
              Completed
            </span>
          {:else if effectiveStatus === 'error'}
            <span class="inline-flex items-center text-xs font-medium text-rose-400 bg-rose-500/10 px-2.5 py-1 rounded-md border border-rose-500/20">
              <XCircle size={12} class="mr-1.5" />
              Failed
            </span>
          {/if}

          <button
            type="button"
            onclick={onClose}
            class="text-ant-text-muted hover:text-ant-text hover:bg-white/5 p-1 rounded-md transition-colors"
            title="Close (Escape)"
          >
            <X size={16} />
          </button>
        </div>
      </div>

      <!-- Quick Metadata Bar -->
      <div class="flex items-center flex-wrap gap-2 px-5 py-2.5 bg-ant-bg-tertiary/40 border-b border-white/5 text-xs text-ant-text-secondary font-mono select-none">
        <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
          <Bot size={12} class="text-violet-400" />
          <span>Role: {effectiveMeta.subagentType}</span>
        </span>
        {#if effectiveMeta.model}
          <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
            <Cpu size={12} class="text-indigo-400" />
            <span>Model: {effectiveMeta.model}</span>
          </span>
        {/if}
        {#if effectiveMeta.isolation}
          <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
            <Shield size={12} class="text-amber-400" />
            <span>Isolation: {effectiveMeta.isolation}</span>
          </span>
        {/if}
        <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
          <span>Execution: {effectiveMeta.isBackground ? 'Background' : 'Foreground'}</span>
        </span>
        {#if effectiveMeta.taskId || toolCall.childSessionId}
          <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-violet-500/10 border border-violet-500/20 text-violet-300">
            <span>ID: {(effectiveMeta.taskId || toolCall.childSessionId || '').slice(0, 8)}...</span>
          </span>
        {/if}
      </div>

      <!-- Modal Body (Scrollable) -->
      <div class="flex-1 overflow-y-auto p-5 space-y-4 scrollbar-thin">
        <!-- 1. Task Instructions (Prompt) -->
        <div class="rounded-lg border border-white/10 bg-ant-bg-secondary/40 overflow-hidden">
          <div class="flex items-center justify-between px-3.5 py-2 bg-ant-bg-tertiary/60 border-b border-white/5 text-xs select-none">
            <span class="font-semibold text-ant-text flex items-center gap-1.5">
              <FileText size={13} class="text-violet-400" />
              Task Prompt
            </span>
            <button
              type="button"
              onclick={copyPrompt}
              class="flex items-center space-x-1 text-xs text-ant-text-secondary hover:text-ant-text px-1.5 py-0.5 rounded hover:bg-white/5 transition-colors cursor-pointer"
            >
              {#if copiedPrompt}
                <Check size={12} class="text-emerald-400" />
                <span class="text-emerald-400">Copied</span>
              {:else}
                <Copy size={12} />
                <span>Copy</span>
              {/if}
            </button>
          </div>
          <div class="p-3.5 text-xs text-ant-text leading-relaxed whitespace-pre-wrap max-h-48 overflow-y-auto scrollbar-thin select-text">
            {effectiveMeta.prompt || 'No prompt instructions recorded.'}
          </div>
        </div>

        <!-- 1.5 Subagent Execution Plan (Tasks) -->
        {#if subagentPlan && subagentPlan.todos && subagentPlan.todos.length > 0}
          <div class="rounded-lg border border-violet-500/30 bg-violet-500/[0.04] overflow-hidden">
            <div class="flex items-center justify-between px-3.5 py-2.5 bg-violet-500/[0.08] border-b border-violet-500/20 text-xs select-none">
              <div class="flex items-center space-x-2">
                <ListTodo size={14} class="text-violet-400" />
                <span class="font-semibold text-ant-text">Subagent Execution Plan</span>
                <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-violet-500/20 text-violet-300 font-semibold border border-violet-500/30">
                  {subagentPlan.completedCount}/{subagentPlan.totalCount} completed ({subagentPlan.progressPercent}%)
                </span>
              </div>
              {#if subagentPlan.inProgressCount > 0}
                <span class="inline-flex items-center text-[11px] font-mono text-amber-400">
                  <Loader2 size={11} class="animate-spin mr-1" />
                  {subagentPlan.inProgressCount} active
                </span>
              {/if}
            </div>

            <!-- Progress bar line -->
            <div class="w-full bg-ant-bg-tertiary h-1 overflow-hidden">
              <div
                class="h-full bg-gradient-to-r from-violet-500 to-indigo-400 transition-all duration-300 ease-out"
                style="width: {subagentPlan.progressPercent}%;"
              ></div>
            </div>

            <!-- Step items -->
            <div class="p-3 space-y-1.5 max-h-56 overflow-y-auto scrollbar-thin">
              {#each subagentPlan.todos as todo (todo.id)}
                <div class="flex items-start space-x-2.5 p-2 rounded-lg bg-ant-bg/60 border border-white/[0.04]">
                  <div class="mt-0.5 shrink-0">
                    {#if todo.status === 'completed'}
                      <div class="w-4 h-4 rounded-full bg-emerald-500/20 text-emerald-400 flex items-center justify-center border border-emerald-500/40">
                        <Check size={10} class="stroke-[3]" />
                      </div>
                    {:else if todo.status === 'in_progress'}
                      <div class="w-4 h-4 rounded-full bg-amber-500/20 text-amber-400 flex items-center justify-center border border-amber-500/40 animate-pulse">
                        <Loader2 size={10} class="animate-spin" />
                      </div>
                    {:else if todo.status === 'cancelled'}
                      <div class="w-4 h-4 rounded-full bg-zinc-800 text-zinc-500 flex items-center justify-center border border-zinc-700">
                        <X size={10} />
                      </div>
                    {:else}
                      <div class="w-4 h-4 rounded-full border border-zinc-600 bg-zinc-800/40 flex items-center justify-center">
                      </div>
                    {/if}
                  </div>
                  <div class="flex-1 min-w-0 font-sans">
                    <div class="text-xs leading-relaxed select-text {todo.status === 'completed' ? 'line-through opacity-60 text-ant-text-muted' : 'text-ant-text'}">
                      {@html renderTodoContent(todo.content)}
                    </div>
                  </div>
                  {#if todo.priority}
                    <span class="text-[8.5px] uppercase tracking-wider font-mono px-1.5 py-0.2 rounded border shrink-0 text-violet-300 bg-violet-500/10 border-violet-500/20">
                      {todo.priority}
                    </span>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- 1.8 Subagent Activity Stream (Tool Calls) -->
        {#if displayChildCalls && displayChildCalls.length > 0}
          <div class="rounded-lg border border-violet-500/30 bg-violet-500/[0.04] overflow-hidden">
            <div class="flex items-center justify-between px-3.5 py-2.5 bg-violet-500/[0.08] border-b border-violet-500/20 text-xs select-none">
              <div class="flex items-center space-x-2">
                <Activity size={14} class="text-violet-400" />
                <span class="font-semibold text-ant-text">Subagent Activity Stream</span>
                <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-violet-500/20 text-violet-300 font-semibold border border-violet-500/30">
                  {displayChildCalls.length} {displayChildCalls.length === 1 ? 'action' : 'actions'}
                </span>
              </div>
              {#if displayChildCalls.some(c => c.status === 'running')}
                <span class="inline-flex items-center text-[11px] font-mono text-violet-400 animate-pulse">
                  <Loader2 size={11} class="animate-spin mr-1" />
                  Running action...
                </span>
              {/if}
            </div>

            <!-- List of Child Actions -->
            <div
              bind:this={activityContainerEl}
              onscroll={handleActivityScroll}
              class="p-2 space-y-1.5 max-h-72 overflow-y-auto scrollbar-thin"
            >
              {#each displayChildCalls as child, idx (child.id || idx)}
                {@const summary = formatActionSummary(child.tool, child.params, subagentPlan?.todos)}
                {@const dur = getChildDuration(child)}
                {@const isChildExpanded = Boolean(expandedChildIds[child.id || String(idx)])}
                {@const planActionData = extractActionTodos(child.params)}
                <div class="rounded-md border border-white/[0.06] bg-ant-bg/70 overflow-hidden text-xs">
                  <!-- Row Header -->
                  <button
                    type="button"
                    onclick={() => toggleChildExpand(child.id || String(idx))}
                    class="w-full flex items-center justify-between px-2.5 py-1.5 hover:bg-white/[0.04] text-left transition-colors cursor-pointer select-none"
                  >
                    <div class="flex items-center space-x-2 min-w-0 flex-1">
                      <!-- Status Icon -->
                      <div class="shrink-0 flex items-center justify-center">
                        {#if child.status === 'running'}
                          <Loader2 size={12} class="animate-spin text-violet-400" />
                        {:else if child.status === 'completed'}
                          <CheckCircle2 size={12} class="text-emerald-400" />
                        {:else}
                          <XCircle size={12} class="text-rose-400" />
                        {/if}
                      </div>

                      <!-- Verb Badge -->
                      <span class="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold uppercase tracking-wider shrink-0 {
                        summary.iconType === 'read' ? 'bg-blue-500/15 text-blue-400 border border-blue-500/30' :
                        summary.iconType === 'edit' ? 'bg-amber-500/15 text-amber-400 border border-amber-500/30' :
                        summary.iconType === 'write' ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30' :
                        summary.iconType === 'search' ? 'bg-purple-500/15 text-purple-400 border border-purple-500/30' :
                        summary.iconType === 'list' ? 'bg-cyan-500/15 text-cyan-400 border border-cyan-500/30' :
                        summary.iconType === 'terminal' ? 'bg-orange-500/15 text-orange-400 border border-orange-500/30' :
                        summary.iconType === 'plan' ? 'bg-indigo-500/15 text-indigo-400 border border-indigo-500/30' :
                        'bg-zinc-500/15 text-zinc-300 border border-zinc-500/30'
                      }">
                        {summary.verb}
                      </span>

                      <!-- Target -->
                      <span class="font-mono text-[11px] text-ant-text truncate flex-1 min-w-0" title={summary.target}>
                        {summary.target}
                      </span>
                    </div>

                    <!-- Right Duration & Chevron -->
                    <div class="flex items-center space-x-2 shrink-0 ml-2 font-mono text-[10px] text-ant-text-muted">
                      {#if dur}
                        <span>{dur}</span>
                      {/if}
                      {#if isChildExpanded}
                        <ChevronDown size={12} class="text-ant-text-secondary" />
                      {:else}
                        <ChevronRight size={12} class="text-ant-text-secondary" />
                      {/if}
                    </div>
                  </button>

                  <!-- Expandable In/Out Details -->
                  {#if isChildExpanded}
                    <div class="p-2.5 border-t border-white/[0.04] bg-black/40 space-y-2.5 font-mono text-[11px]">
                      {#if planActionData}
                        <!-- Formatted Execution Plan Task Status Update View -->
                        <div class="p-2 rounded-lg bg-indigo-500/[0.06] border border-indigo-500/25 space-y-2 font-sans">
                          <div class="flex items-center justify-between text-[11px]">
                            <div class="flex items-center space-x-1.5 font-semibold text-indigo-300">
                              <ListTodo size={13} class="text-indigo-400 shrink-0" />
                              <span>{planActionData.isMerge ? 'Plan Task Status Update' : 'Full Execution Plan Creation'}</span>
                            </div>
                            <span class="text-[10px] font-mono px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                              {planActionData.todos.length} {planActionData.todos.length === 1 ? 'task' : 'tasks'}
                            </span>
                          </div>

                          <div class="space-y-1">
                            {#each planActionData.todos as todo}
                              {@const resolvedContent = resolveTodoDisplay(todo, subagentPlan?.todos)}
                              <div class="flex items-center justify-between px-2 py-1.5 rounded bg-black/30 border border-white/[0.04] text-xs">
                                <div class="flex items-center space-x-2 min-w-0 flex-1">
                                  {#if todo.status === 'completed'}
                                    <div class="w-3.5 h-3.5 rounded-full bg-emerald-500/20 text-emerald-400 flex items-center justify-center border border-emerald-500/40 shrink-0" title="Completed">
                                      <Check size={9} class="stroke-[3]" />
                                    </div>
                                  {:else if todo.status === 'in_progress'}
                                    <div class="w-3.5 h-3.5 rounded-full bg-amber-500/20 text-amber-400 flex items-center justify-center border border-amber-500/40 shrink-0 animate-pulse" title="In Progress">
                                      <Loader2 size={9} class="animate-spin" />
                                    </div>
                                  {:else if todo.status === 'cancelled'}
                                    <div class="w-3.5 h-3.5 rounded-full bg-rose-500/20 text-rose-400 flex items-center justify-center border border-rose-500/40 shrink-0" title="Cancelled">
                                      <XCircle size={9} />
                                    </div>
                                  {:else}
                                    <div class="w-3.5 h-3.5 rounded-full border border-zinc-600 bg-zinc-800/50 shrink-0" title="Pending"></div>
                                  {/if}
                                  <span class="font-mono text-[11px] truncate text-ant-text {todo.status === 'completed' ? 'line-through opacity-60' : ''}">
                                    {resolvedContent}
                                  </span>
                                </div>
                                <span class="text-[9.5px] font-mono uppercase tracking-wider px-1.5 py-0.5 rounded shrink-0 ml-2 {
                                  todo.status === 'completed' ? 'text-emerald-400 bg-emerald-500/15 border border-emerald-500/30' :
                                  todo.status === 'in_progress' ? 'text-amber-400 bg-amber-500/15 border border-amber-500/30 font-semibold' :
                                  todo.status === 'cancelled' ? 'text-rose-400 bg-rose-500/15 border border-rose-500/30' :
                                  'text-zinc-400 bg-zinc-800/60 border border-zinc-700'
                                }">
                                  {todo.status === 'in_progress' ? 'In Progress' : todo.status || 'Pending'}
                                </span>
                              </div>
                            {/each}
                          </div>
                        </div>

                        <!-- Output Status for Plan Tool -->
                        <div class="flex items-center justify-between text-[11px] p-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 font-sans">
                          <div class="flex items-center space-x-1.5">
                            <CheckCircle2 size={13} class="text-emerald-400 shrink-0" />
                            <span>Plan state updated in orchestrator</span>
                          </div>
                          {#if child.result && child.result.trim() && child.result.trim() !== '{}'}
                            <span class="text-[10px] font-mono text-emerald-400/80">{child.result.slice(0, 30)}</span>
                          {/if}
                        </div>

                        <!-- Collapsible Raw Parameters & Output -->
                        <details class="group text-[10.5px]">
                          <summary class="text-ant-text-muted hover:text-ant-text cursor-pointer select-none py-0.5 flex items-center space-x-1">
                            <span class="text-[9.5px] font-mono">View raw parameters & output JSON</span>
                          </summary>
                          <div class="pt-2 space-y-2">
                            <div>
                              <div class="flex items-center justify-between text-[10px] text-ant-text-muted mb-1 select-none">
                                <span class="font-semibold text-violet-300">Raw Input / Parameters</span>
                                <button
                                  type="button"
                                  onclick={() => copyText(typeof child.params === 'string' ? child.params : JSON.stringify(child.params, null, 2), `in-${child.id || idx}`)}
                                  class="hover:text-white flex items-center space-x-1 cursor-pointer"
                                >
                                  {#if copiedChildId === `in-${child.id || idx}`}
                                    <Check size={10} class="text-emerald-400" />
                                    <span class="text-emerald-400">Copied</span>
                                  {:else}
                                    <Copy size={10} />
                                    <span>Copy</span>
                                  {/if}
                                </button>
                              </div>
                              <pre class="p-2 rounded bg-ant-bg-secondary/60 text-ant-text max-h-36 overflow-y-auto scrollbar-thin whitespace-pre-wrap break-all">{typeof child.params === 'string' ? child.params : JSON.stringify(child.params, null, 2)}</pre>
                            </div>
                            {#if child.result}
                              <div>
                                <div class="flex items-center justify-between text-[10px] text-ant-text-muted mb-1 select-none">
                                  <span class="font-semibold text-emerald-400">Raw Output</span>
                                  <button
                                    type="button"
                                    onclick={() => copyText(child.result || '', `out-${child.id || idx}`)}
                                    class="hover:text-white flex items-center space-x-1 cursor-pointer"
                                  >
                                    {#if copiedChildId === `out-${child.id || idx}`}
                                      <Check size={10} class="text-emerald-400" />
                                      <span class="text-emerald-400">Copied</span>
                                    {:else}
                                      <Copy size={10} />
                                      <span>Copy</span>
                                    {/if}
                                  </button>
                                </div>
                                <pre class="p-2 rounded bg-ant-bg-secondary/60 text-ant-text max-h-48 overflow-y-auto scrollbar-thin whitespace-pre-wrap break-all">{child.result}</pre>
                              </div>
                            {/if}
                          </div>
                        </details>
                      {:else}
                        <!-- In: Parameters / Command -->
                        <div>
                          <div class="flex items-center justify-between text-[10px] text-ant-text-muted mb-1 select-none">
                            <span class="font-semibold text-violet-300">Input / Parameters</span>
                            <button
                              type="button"
                              onclick={() => copyText(typeof child.params === 'string' ? child.params : JSON.stringify(child.params, null, 2), `in-${child.id || idx}`)}
                              class="hover:text-white flex items-center space-x-1 cursor-pointer"
                            >
                              {#if copiedChildId === `in-${child.id || idx}`}
                                <Check size={10} class="text-emerald-400" />
                                <span class="text-emerald-400">Copied</span>
                              {:else}
                                <Copy size={10} />
                                <span>Copy</span>
                              {/if}
                            </button>
                          </div>
                          <pre class="p-2 rounded bg-ant-bg-secondary/60 text-ant-text max-h-36 overflow-y-auto scrollbar-thin whitespace-pre-wrap break-all">{typeof child.params === 'string' ? child.params : JSON.stringify(child.params, null, 2)}</pre>
                        </div>

                        <!-- Out: Result -->
                        {#if child.result}
                          <div>
                            <div class="flex items-center justify-between text-[10px] text-ant-text-muted mb-1 select-none">
                              <span class="font-semibold text-emerald-400">Output</span>
                              <button
                                type="button"
                                onclick={() => copyText(child.result || '', `out-${child.id || idx}`)}
                                class="hover:text-white flex items-center space-x-1 cursor-pointer"
                              >
                                {#if copiedChildId === `out-${child.id || idx}`}
                                  <Check size={10} class="text-emerald-400" />
                                  <span class="text-emerald-400">Copied</span>
                                {:else}
                                  <Copy size={10} />
                                  <span>Copy</span>
                                {/if}
                              </button>
                            </div>
                            <pre class="p-2 rounded bg-ant-bg-secondary/60 text-ant-text max-h-48 overflow-y-auto scrollbar-thin whitespace-pre-wrap break-all">{child.result}</pre>
                          </div>
                        {/if}
                      {/if}
                    </div>
                  {/if}
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- 2. Subagent Result / Live Activity -->
        <div class="rounded-lg border border-white/10 bg-ant-bg-secondary/40 overflow-hidden">
          <div class="flex items-center justify-between px-4 py-2.5 bg-ant-bg-tertiary/60 border-b border-white/5 text-xs select-none">
            <span class="font-semibold text-ant-text flex items-center gap-1.5 font-sans">
              <Code size={13} class="text-emerald-400" />
              Subagent Output
            </span>
            {#if cleanedOutput}
              <button
                type="button"
                onclick={copyOutput}
                class="flex items-center space-x-1 text-xs text-ant-text-secondary hover:text-ant-text px-2 py-1 rounded hover:bg-white/5 transition-colors cursor-pointer font-sans"
              >
                {#if copiedOutput}
                  <Check size={12} class="text-emerald-400" />
                  <span class="text-emerald-400">Copied</span>
                {:else}
                  <Copy size={12} />
                  <span>Copy</span>
                {/if}
              </button>
            {/if}
          </div>

          <div class="p-5 select-text overflow-y-auto max-h-[38rem] scrollbar-thin">
            {#if isActuallyRunning && !cleanedOutput}
              <div class="flex flex-col items-center justify-center py-8 space-y-3 text-center">
                <div class="w-10 h-10 rounded-full bg-violet-500/10 border border-violet-500/30 flex items-center justify-center text-violet-400">
                  <Loader2 size={20} class="animate-spin text-violet-400" />
                </div>
                <div class="space-y-1 font-sans">
                  <p class="font-medium text-xs text-white">Subagent is actively working in background</p>
                  <p class="text-[11px] text-ant-text-muted">Elapsed time: {elapsedDuration || 'In progress...'} · Results will appear once completed.</p>
                </div>

                {#if displayChildCalls && displayChildCalls.length > 0}
                  {@const latestChild = displayChildCalls[displayChildCalls.length - 1]}
                  {@const latestSummary = formatActionSummary(latestChild.tool, latestChild.params, subagentPlan?.todos)}
                  <div class="w-full max-w-md p-2.5 rounded-lg bg-black/40 border border-violet-500/30 text-left font-mono space-y-1 mt-2">
                    <div class="flex items-center justify-between text-[10.5px]">
                      <span class="text-violet-300 font-semibold flex items-center gap-1.5">
                        <Activity size={12} class="text-violet-400" />
                        Live Activity ({displayChildCalls.length} {displayChildCalls.length === 1 ? 'action' : 'actions'})
                      </span>
                      {#if latestChild.status === 'running'}
                        <span class="text-amber-400 text-[10px] flex items-center gap-1">
                          <Loader2 size={10} class="animate-spin" /> Running
                        </span>
                      {:else}
                        <span class="text-emerald-400 text-[10px]">Latest completed</span>
                      {/if}
                    </div>
                    <div class="text-[11px] text-ant-text flex items-center gap-1.5 truncate">
                      <span class="px-1 py-0.2 rounded text-[9px] font-semibold uppercase shrink-0 {latestSummary.iconType === 'plan' ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30' : 'bg-violet-500/20 text-violet-300 border border-violet-500/30'}">
                        {latestSummary.verb}
                      </span>
                      <span class="truncate flex-1" title={latestSummary.target}>{latestSummary.target}</span>
                    </div>
                  </div>
                {/if}
              </div>
            {:else if cleanedOutput}
              <div class="font-serif text-[14px] text-ant-text leading-[1.75] tracking-normal select-text markdown-chat-body max-w-none">
                {@html renderMarkdown(cleanedOutput)}
              </div>
            {:else}
              <div class="py-6 text-center text-ant-text-muted font-mono text-xs">
                No output returned from subagent.
              </div>
            {/if}
          </div>
        </div>

        <!-- 3. Collapsible Raw Arguments -->
        <div class="rounded-lg border border-white/5 bg-ant-bg-tertiary/20 overflow-hidden text-xs">
          <button
            type="button"
            onclick={() => (showRawPayload = !showRawPayload)}
            class="w-full flex items-center justify-between px-3.5 py-2 hover:bg-white/5 text-ant-text-muted hover:text-ant-text font-mono text-left transition-colors"
          >
            <span>Raw Parameters & Config</span>
            <span class="text-[10px]">{showRawPayload ? '▲ Hide' : '▼ View'}</span>
          </button>
          {#if showRawPayload}
            <div class="p-3 border-t border-white/5 bg-black/40 overflow-x-auto max-h-40 scrollbar-thin">
              <pre class="font-mono text-[11px] text-ant-text-muted">{JSON.stringify(toolCall.params, null, 2)}</pre>
            </div>
          {/if}
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between px-5 py-3 bg-ant-bg-secondary/40 border-t border-white/5 text-xs text-ant-text-muted select-none">
        <span class="font-mono">Press Escape to close</span>
        <button
          type="button"
          onclick={onClose}
          class="px-3 py-1.5 rounded-md bg-white/10 hover:bg-white/15 text-ant-text font-medium transition-colors"
        >
          Close
        </button>
      </div>
    </div>
  </div>
{/if}
