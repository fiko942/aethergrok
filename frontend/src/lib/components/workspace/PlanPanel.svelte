<script lang="ts">
  import { planStore } from '$lib/stores/plan.svelte';
  import { sessionStore } from '$lib/stores/session.svelte';
  import { type SessionPlanState, renderTodoContent } from '$lib/utils/planParser';
  import {
    ListTodo,
    CheckCircle2,
    Clock,
    Check,
    Loader2,
    X,
    Sparkles,
    AlertCircle,
    Bot,
    Layers
  } from 'lucide-svelte';

  interface Props {
    sessionId?: string;
  }

  let { sessionId }: Props = $props();

  const currentSessionId = $derived(sessionId || sessionStore.activeSessionId);
  const planState = $derived<SessionPlanState | null>(planStore.getPlan(currentSessionId));

  const subagentPlans = $derived(planStore.getSubagentPlans(currentSessionId));
  const hasSubagentPlans = $derived(subagentPlans.length > 0);
  const mainTodos = $derived(planState?.mainTodos || []);
  const hasMainPlan = $derived(mainTodos.length > 0);

  let selectedTab = $state<string>('all');

  $effect(() => {
    if (selectedTab !== 'all' && selectedTab !== 'main') {
      const exists = subagentPlans.some(sp => sp.subagentId === selectedTab);
      if (!exists) selectedTab = 'all';
    }
  });
</script>

<div class="h-full flex flex-col bg-ant-bg select-none font-serif overflow-hidden">
  {#if planState && planState.todos && planState.todos.length > 0}
    <!-- Panel Header: Title, Progress Bar, Metrics -->
    <div class="p-3.5 border-b border-ant-border-secondary dark:border-white/5 bg-ant-bg-secondary shrink-0 space-y-3">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-2">
          <div class="w-7 h-7 rounded-lg bg-indigo-500/15 text-indigo-400 flex items-center justify-center">
            <ListTodo size={15} />
          </div>
          <div>
            <h3 class="font-serif-display text-xs font-semibold text-ant-text">
              Active Plan Tracker
            </h3>
            <p class="text-[10.5px] font-mono text-ant-text-secondary mt-0.5">
              {planState.completedCount} of {planState.totalCount} completed ({planState.progressPercent}%)
            </p>
          </div>
        </div>

        <div class="flex items-center space-x-1.5">
          <span class="font-mono text-xs font-bold px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
            {planState.progressPercent}%
          </span>
          <button
            type="button"
            onclick={() => currentSessionId && planStore.resetPlan(currentSessionId)}
            class="p-1 rounded text-ant-text-muted hover:text-rose-400 hover:bg-rose-500/10 transition cursor-pointer"
            title="Clear Plan"
          >
            <X size={13} />
          </button>
        </div>
      </div>

      <!-- Segmented / Continuous Progress Bar -->
      <div class="w-full bg-ant-bg-tertiary rounded-full h-1.5 overflow-hidden">
        <div
          class="h-full bg-indigo-500 transition-all duration-300 ease-out"
          style="width: {planState.progressPercent}%;"
        ></div>
      </div>

      <!-- Summary Stat Chips -->
      <div class="grid grid-cols-3 gap-2 font-mono text-[10.5px]">
        <div class="bg-ant-bg p-2 rounded-lg border border-ant-border-secondary dark:border-white/5 text-center">
          <div class="text-emerald-500 font-bold text-xs">{planState.completedCount}</div>
          <div class="text-[10px] text-ant-text-muted uppercase tracking-wider">Done</div>
        </div>
        <div class="bg-ant-bg p-2 rounded-lg border border-ant-border-secondary dark:border-white/5 text-center">
          <div class="text-amber-500 font-bold text-xs">{planState.inProgressCount}</div>
          <div class="text-[10px] text-ant-text-muted uppercase tracking-wider">Active</div>
        </div>
        <div class="bg-ant-bg p-2 rounded-lg border border-ant-border-secondary dark:border-white/5 text-center">
          <div class="text-ant-text-secondary font-bold text-xs">{planState.pendingCount}</div>
          <div class="text-[10px] text-ant-text-muted uppercase tracking-wider">Pending</div>
        </div>
      </div>

      <!-- Filter Tabs (shown when subagent plans exist) -->
      {#if hasSubagentPlans}
        <div class="flex items-center space-x-1 pt-1 overflow-x-auto scrollbar-none text-[10.5px] font-mono border-t border-ant-border-secondary/40 dark:border-white/[0.04]">
          <button
            type="button"
            onclick={() => (selectedTab = 'all')}
            class="px-2 py-0.5 rounded-md transition-colors shrink-0 cursor-pointer {selectedTab === 'all' ? 'bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 font-semibold border border-indigo-500/30' : 'text-ant-text-secondary hover:text-ant-text hover:bg-white/5'}"
          >
            All ({planState.totalCount})
          </button>
          {#if hasMainPlan}
            <button
              type="button"
              onclick={() => (selectedTab = 'main')}
              class="px-2 py-0.5 rounded-md transition-colors shrink-0 cursor-pointer {selectedTab === 'main' ? 'bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 font-semibold border border-indigo-500/30' : 'text-ant-text-secondary hover:text-ant-text hover:bg-white/5'}"
            >
              Main ({mainTodos.length})
            </button>
          {/if}
          {#each subagentPlans as sp (sp.subagentId)}
            <button
              type="button"
              onclick={() => (selectedTab = sp.subagentId)}
              class="px-2 py-0.5 rounded-md transition-colors shrink-0 flex items-center space-x-1 cursor-pointer {selectedTab === sp.subagentId ? 'bg-violet-500/20 text-violet-700 dark:text-violet-300 font-semibold border border-violet-500/30' : 'text-ant-text-secondary hover:text-ant-text hover:bg-white/5'}"
              title={sp.subagentName}
            >
              <Bot size={11} class="text-violet-500 dark:text-violet-400" />
              <span class="max-w-[90px] truncate">{sp.subagentName}</span>
              <span class="text-[9.5px] opacity-75">({sp.completedCount}/{sp.totalCount})</span>
            </button>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Task List -->
    <div class="flex-1 p-3 overflow-y-auto scrollbar-thin space-y-2 select-text">
      {#if selectedTab === 'all' && hasSubagentPlans}
        {#if hasMainPlan}
          <div class="space-y-1.5 mb-3">
            <div class="flex items-center justify-between px-2.5 py-1 rounded-lg bg-indigo-500/10 border border-indigo-500/20 text-[10.5px] font-mono text-indigo-700 dark:text-indigo-300 font-semibold">
              <div class="flex items-center space-x-1.5 truncate">
                <Layers size={12} />
                <span>Main Orchestrator Plan</span>
              </div>
              <span>{mainTodos.filter(t => t.status === 'completed').length}/{mainTodos.length}</span>
            </div>
            {#each mainTodos as todo (todo.id)}
              <div class="p-2.5 rounded-xl border transition-all {
                todo.status === 'completed'
                  ? 'bg-emerald-500/[0.03] border-emerald-500/20 text-ant-text-muted'
                  : todo.status === 'in_progress'
                    ? 'bg-amber-500/[0.04] border-amber-500/30 text-ant-text shadow-xs'
                    : 'bg-ant-bg-secondary/40 border-ant-border-secondary dark:border-white/5 text-ant-text'
              }">
                <div class="flex items-start space-x-2.5">
                  <div class="mt-0.5 shrink-0">
                    {#if todo.status === 'completed'}
                      <div class="w-4 h-4 rounded-full bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 flex items-center justify-center border border-emerald-500/40">
                        <Check size={10} class="stroke-[3]" />
                      </div>
                    {:else if todo.status === 'in_progress'}
                      <div class="w-4 h-4 rounded-full bg-amber-500/20 text-amber-600 dark:text-amber-400 flex items-center justify-center border border-amber-500/40 animate-pulse">
                        <Loader2 size={10} class="animate-spin" />
                      </div>
                    {:else if todo.status === 'cancelled'}
                      <div class="w-4 h-4 rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-500 flex items-center justify-center border border-zinc-300 dark:border-zinc-700">
                        <X size={10} />
                      </div>
                    {:else}
                      <div class="w-4 h-4 rounded-full border border-zinc-300 dark:border-zinc-600 bg-zinc-100 dark:bg-zinc-800/40 flex items-center justify-center">
                      </div>
                    {/if}
                  </div>

                  <div class="flex-1 min-w-0 font-sans">
                    <div class="text-xs leading-relaxed select-text {todo.status === 'completed' ? 'line-through opacity-60 text-ant-text-muted' : 'text-ant-text'}">
                      {@html renderTodoContent(todo.content)}
                    </div>

                    <div class="flex items-center space-x-2 mt-1.5 font-mono text-[9.5px]">
                      <span class="px-1.5 py-0.2 rounded uppercase tracking-wider font-semibold {
                        todo.status === 'completed' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20' :
                        todo.status === 'in_progress' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20' :
                        'bg-ant-bg-tertiary text-ant-text-secondary border border-ant-border-secondary'
                      }">
                        {todo.status.replace('_', ' ')}
                      </span>

                      {#if todo.priority}
                        <span class="text-ant-text-muted">•</span>
                        <span class="text-ant-text-secondary capitalize">
                          {todo.priority} priority
                        </span>
                      {/if}
                    </div>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {/if}

        {#each subagentPlans as sp (sp.subagentId)}
          <div class="space-y-1.5 mb-3">
            <div class="flex items-center justify-between px-2.5 py-1 rounded-lg bg-violet-500/10 border border-violet-500/20 text-[10.5px] font-mono text-violet-700 dark:text-violet-300 font-semibold">
              <div class="flex items-center space-x-1.5 truncate">
                <Bot size={12} class="text-violet-500 dark:text-violet-400" />
                <span class="truncate">Subagent: {sp.subagentName}</span>
              </div>
              <span>{sp.completedCount}/{sp.totalCount} ({sp.progressPercent}%)</span>
            </div>
            {#each sp.todos as todo (todo.id)}
              <div class="p-2.5 rounded-xl border transition-all {
                todo.status === 'completed'
                  ? 'bg-emerald-500/[0.03] border-emerald-500/20 text-ant-text-muted'
                  : todo.status === 'in_progress'
                    ? 'bg-amber-500/[0.04] border-amber-500/30 text-ant-text shadow-xs'
                    : 'bg-ant-bg-secondary/40 border-ant-border-secondary dark:border-white/5 text-ant-text'
              }">
                <div class="flex items-start space-x-2.5">
                  <div class="mt-0.5 shrink-0">
                    {#if todo.status === 'completed'}
                      <div class="w-4 h-4 rounded-full bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 flex items-center justify-center border border-emerald-500/40">
                        <Check size={10} class="stroke-[3]" />
                      </div>
                    {:else if todo.status === 'in_progress'}
                      <div class="w-4 h-4 rounded-full bg-amber-500/20 text-amber-600 dark:text-amber-400 flex items-center justify-center border border-amber-500/40 animate-pulse">
                        <Loader2 size={10} class="animate-spin" />
                      </div>
                    {:else if todo.status === 'cancelled'}
                      <div class="w-4 h-4 rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-500 flex items-center justify-center border border-zinc-300 dark:border-zinc-700">
                        <X size={10} />
                      </div>
                    {:else}
                      <div class="w-4 h-4 rounded-full border border-zinc-300 dark:border-zinc-600 bg-zinc-100 dark:bg-zinc-800/40 flex items-center justify-center">
                      </div>
                    {/if}
                  </div>

                  <div class="flex-1 min-w-0 font-sans">
                    <div class="text-xs leading-relaxed select-text {todo.status === 'completed' ? 'line-through opacity-60 text-ant-text-muted' : 'text-ant-text'}">
                      {@html renderTodoContent(todo.content)}
                    </div>

                    <div class="flex items-center space-x-2 mt-1.5 font-mono text-[9.5px]">
                      <span class="px-1.5 py-0.2 rounded uppercase tracking-wider font-semibold {
                        todo.status === 'completed' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20' :
                        todo.status === 'in_progress' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20' :
                        'bg-ant-bg-tertiary text-ant-text-secondary border border-ant-border-secondary'
                      }">
                        {todo.status.replace('_', ' ')}
                      </span>

                      {#if todo.priority}
                        <span class="text-ant-text-muted">•</span>
                        <span class="text-ant-text-secondary capitalize">
                          {todo.priority} priority
                        </span>
                      {/if}
                    </div>
                  </div>
                </div>
              </div>
            {/each}
          </div>
        {/each}
      {:else}
        {@const activeTodos = selectedTab === 'main' ? mainTodos : (selectedTab !== 'all' ? (subagentPlans.find(s => s.subagentId === selectedTab)?.todos || []) : planState.todos)}
        {#each activeTodos as todo (todo.id)}
          <div class="p-2.5 rounded-xl border transition-all {
            todo.status === 'completed'
              ? 'bg-emerald-500/[0.03] border-emerald-500/20 text-ant-text-muted'
              : todo.status === 'in_progress'
                ? 'bg-amber-500/[0.04] border-amber-500/30 text-ant-text shadow-xs'
                : 'bg-ant-bg-secondary/40 border-ant-border-secondary dark:border-white/5 text-ant-text'
          }">
            <div class="flex items-start space-x-2.5">
              <div class="mt-0.5 shrink-0">
                {#if todo.status === 'completed'}
                  <div class="w-4 h-4 rounded-full bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 flex items-center justify-center border border-emerald-500/40">
                    <Check size={10} class="stroke-[3]" />
                  </div>
                {:else if todo.status === 'in_progress'}
                  <div class="w-4 h-4 rounded-full bg-amber-500/20 text-amber-600 dark:text-amber-400 flex items-center justify-center border border-amber-500/40 animate-pulse">
                    <Loader2 size={10} class="animate-spin" />
                  </div>
                {:else if todo.status === 'cancelled'}
                  <div class="w-4 h-4 rounded-full bg-zinc-200 dark:bg-zinc-800 text-zinc-500 flex items-center justify-center border border-zinc-300 dark:border-zinc-700">
                    <X size={10} />
                  </div>
                {:else}
                  <div class="w-4 h-4 rounded-full border border-zinc-300 dark:border-zinc-600 bg-zinc-100 dark:bg-zinc-800/40 flex items-center justify-center">
                  </div>
                {/if}
              </div>

              <div class="flex-1 min-w-0 font-sans">
                <div class="text-xs leading-relaxed select-text {todo.status === 'completed' ? 'line-through opacity-60 text-ant-text-muted' : 'text-ant-text'}">
                  {@html renderTodoContent(todo.content)}
                </div>

                {#if todo.subagentName && selectedTab === 'all'}
                  <p class="text-[9.5px] font-mono text-violet-600 dark:text-violet-400 truncate mt-1">
                    Subagent: {todo.subagentName}
                  </p>
                {/if}

                <div class="flex items-center space-x-2 mt-1.5 font-mono text-[9.5px]">
                  <span class="px-1.5 py-0.2 rounded uppercase tracking-wider font-semibold {
                    todo.status === 'completed' ? 'bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20' :
                    todo.status === 'in_progress' ? 'bg-amber-500/10 text-amber-600 dark:text-amber-400 border border-amber-500/20' :
                    'bg-ant-bg-tertiary text-ant-text-secondary border border-ant-border-secondary'
                  }">
                    {todo.status.replace('_', ' ')}
                  </span>

                  {#if todo.priority}
                    <span class="text-ant-text-muted">•</span>
                    <span class="text-ant-text-secondary capitalize">
                      {todo.priority} priority
                    </span>
                  {/if}
                </div>
              </div>
            </div>
          </div>
        {/each}
      {/if}
    </div>
  {:else}
    <!-- Empty State -->
    <div class="h-full flex flex-col items-center justify-center p-6 text-center text-ant-text-secondary space-y-3">
      <div class="w-12 h-12 rounded-2xl bg-ant-bg-tertiary flex items-center justify-center text-ant-text-muted border border-ant-border-secondary dark:border-white/5">
        <ListTodo size={22} />
      </div>
      <div>
        <h4 class="font-serif font-semibold text-xs text-ant-text">No Active Plan</h4>
        <p class="text-[11px] text-ant-text-muted mt-1 max-w-xs leading-relaxed">
          When an execution plan or todo list is updated in this session, it will be automatically tracked here.
        </p>
      </div>
    </div>
  {/if}
</div>
