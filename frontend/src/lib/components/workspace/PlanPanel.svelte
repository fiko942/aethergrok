<script lang="ts">
  import { planStore } from '$lib/stores/plan.svelte';
  import { sessionStore } from '$lib/stores/session.svelte';
  import type { SessionPlanState } from '$lib/utils/planParser';
  import {
    ListTodo,
    CheckCircle2,
    Clock,
    Check,
    Loader2,
    X,
    Sparkles,
    AlertCircle
  } from 'lucide-svelte';

  interface Props {
    sessionId?: string;
  }

  let { sessionId }: Props = $props();

  const currentSessionId = $derived(sessionId || sessionStore.activeSessionId);
  const planState = $derived<SessionPlanState | null>(planStore.getPlan(currentSessionId));
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

        <span class="font-mono text-xs font-bold px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
          {planState.progressPercent}%
        </span>
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
    </div>

    <!-- Task List -->
    <div class="flex-1 p-3 overflow-y-auto scrollbar-thin space-y-2 select-text">
      {#each planState.todos as todo (todo.id)}
        <div class="p-2.5 rounded-xl border transition-all {
          todo.status === 'completed'
            ? 'bg-emerald-500/[0.03] border-emerald-500/20 text-ant-text-muted'
            : todo.status === 'in_progress'
              ? 'bg-amber-500/[0.04] border-amber-500/30 text-ant-text shadow-xs'
              : 'bg-ant-bg-secondary/40 border-ant-border-secondary dark:border-white/5 text-ant-text'
        }">
          <div class="flex items-start space-x-2.5">
            <!-- Status Badge/Icon -->
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

            <div class="flex-1 min-w-0">
              <p class="text-xs leading-relaxed font-serif {todo.status === 'completed' ? 'line-through text-ant-text-muted' : 'text-ant-text'}">
                {todo.content}
              </p>

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
