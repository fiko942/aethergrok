<script lang="ts">
  import { planStore } from '$lib/stores/plan.svelte';
  import { sessionStore } from '$lib/stores/session.svelte';
  import { settingsStore } from '$lib/stores/settings.svelte';
  import {
    type SessionPlanState,
    type TodoItem,
    renderTodoContent
  } from '$lib/utils/planParser';
  import {
    ListTodo,
    CheckCircle2,
    Clock,
    ChevronDown,
    ChevronUp,
    Minimize2,
    Maximize2,
    X,
    Sparkles,
    Check,
    Loader2,
    Layers,
    PanelRightOpen,
    Bot
  } from 'lucide-svelte';

  interface Props {
    sessionId?: string;
    onOpenRightSidebarTab?: (tab: 'plan' | 'files' | 'changes') => void;
  }

  let { sessionId, onOpenRightSidebarTab }: Props = $props();

  const currentSessionId = $derived(sessionId || sessionStore.activeSessionId);
  const activeSession = $derived(sessionStore.sessions.find(s => s.id === currentSessionId) || sessionStore.activeSession);
  const planState = $derived<SessionPlanState | null>(planStore.getPlan(currentSessionId));

  const subagentPlans = $derived(planStore.getSubagentPlans(currentSessionId));
  const hasSubagentPlans = $derived(subagentPlans.length > 0);
  const mainTodos = $derived(planState?.mainTodos || []);
  const hasMainPlan = $derived(mainTodos.length > 0);

  // Active filter tab: 'all' | 'main' | subagentId
  let selectedTab = $state<string>('all');

  $effect(() => {
    if (selectedTab !== 'all' && selectedTab !== 'main') {
      const exists = subagentPlans.some(sp => sp.subagentId === selectedTab);
      if (!exists) selectedTab = 'all';
    }
  });

  // Determine if the Right Sidebar is currently open showing the Plan tab
  const isSidebarPlanOpen = $derived(
    Boolean(activeSession && activeSession.rightSidebarOpen && activeSession.rightSidebarTab === 'plan')
  );

  let isHovered = $state(false);

  // Active or running task item
  const activeTask = $derived.by<TodoItem | null>(() => {
    if (!planState || !planState.todos) return null;
    return planState.todos.find(t => t.status === 'in_progress') ||
           planState.todos.find(t => t.status === 'pending') ||
           null;
  });

  function handleDismiss() {
    planStore.isFloatingVisible = false;
  }

  function handleToggleCompact() {
    planStore.toggleCompact();
  }

  function handleOpenSidebarPlan() {
    if (onOpenRightSidebarTab) {
      onOpenRightSidebarTab('plan');
    } else if (currentSessionId) {
      sessionStore.setRightSidebarTab('plan', currentSessionId);
      const sess = sessionStore.sessions.find(s => s.id === currentSessionId);
      if (sess && !sess.rightSidebarOpen) {
        sessionStore.toggleRightSidebar(currentSessionId);
      }
    }
  }
</script>

{#if planState && planState.todos && planState.todos.length > 0 && planStore.isFloatingVisible && !isSidebarPlanOpen}
  <!-- Floating Plan Tracker Container (Positioned gracefully at top-right of main chat stage) -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="absolute top-3 right-4 z-30 transition-all duration-200 select-none {planStore.isCompact ? 'w-auto' : 'w-80 sm:w-96'}"
    onmouseenter={() => {
      isHovered = true;
      planStore.clearUpdateFlag();
    }}
    onmouseleave={() => isHovered = false}
  >
    {#if planStore.isCompact}
      <!-- Compact Pill Badge Mode -->
      <button
        type="button"
        onclick={handleToggleCompact}
        class="group flex items-center space-x-2.5 px-3 py-1.5 rounded-full bg-ant-bg-secondary/95 dark:bg-[#16161a]/90 hover:bg-ant-bg-tertiary dark:hover:bg-[#1f1f26]/95 border border-indigo-500/30 hover:border-indigo-500/50 shadow-lg shadow-black/10 dark:shadow-black/40 backdrop-blur-md transition-all text-left cursor-pointer"
        title="Click to expand full execution plan tracker"
      >
        <!-- Pulse Indicator / Icon -->
        <div class="relative flex items-center justify-center">
          <div class="w-6 h-6 rounded-full bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 flex items-center justify-center">
            <ListTodo size={13} />
          </div>
          {#if planStore.hasNewUpdate}
            <span class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full bg-indigo-400 animate-ping"></span>
            <span class="absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full bg-indigo-500"></span>
          {/if}
        </div>

        <!-- Progress Summary -->
        <div class="flex items-center space-x-2 text-xs font-sans">
          <span class="font-medium text-ant-text">Plan Progress</span>
          <span class="font-mono text-[11px] font-semibold text-indigo-600 dark:text-indigo-400 bg-indigo-500/10 px-1.5 py-0.2 rounded border border-indigo-500/20">
            {planState.completedCount}/{planState.totalCount} ({planState.progressPercent}%)
          </span>
        </div>

        <Maximize2 size={12} class="text-ant-text-muted group-hover:text-ant-text transition-colors" />
      </button>
    {:else}
      <!-- Expanded Floating Plan Card -->
      <div class="rounded-xl bg-ant-bg border border-indigo-500/25 shadow-2xl shadow-black/15 dark:shadow-black/60 backdrop-blur-xl overflow-hidden font-sans animate-in fade-in zoom-in-95 duration-150">
        <!-- Card Header -->
        <div class="flex items-center justify-between px-3.5 py-2.5 bg-indigo-500/[0.06] dark:bg-gradient-to-r dark:from-indigo-950/40 dark:to-transparent border-b border-ant-border-secondary dark:border-white/[0.06]">
          <div class="flex items-center space-x-2 min-w-0">
            <div class="w-6 h-6 rounded-lg bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 flex items-center justify-center shrink-0">
              <ListTodo size={14} />
            </div>
            <div class="min-w-0">
              <div class="flex items-center space-x-2">
                <span class="font-sans text-xs font-semibold text-ant-text tracking-tight">Active Execution Plan</span>
                {#if planStore.hasNewUpdate}
                  <span class="text-[9px] font-mono px-1.5 py-0.2 rounded-full bg-indigo-500/20 text-indigo-700 dark:text-indigo-300 border border-indigo-500/30 font-semibold animate-pulse">
                    Updated
                  </span>
                {/if}
              </div>
              <p class="text-[10px] text-ant-text-secondary truncate font-mono">
                {planState.completedCount} of {planState.totalCount} tasks finished ({planState.progressPercent}%)
              </p>
            </div>
          </div>

          <!-- Actions: Dock to sidebar, Minimize, Close -->
          <div class="flex items-center space-x-1 shrink-0">
            <button
              type="button"
              onclick={handleOpenSidebarPlan}
              class="p-1 rounded-md text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition cursor-pointer"
              title="Pin to Right Inspector Panel"
            >
              <PanelRightOpen size={13} />
            </button>
            <button
              type="button"
              onclick={handleToggleCompact}
              class="p-1 rounded-md text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition cursor-pointer"
              title="Compact view"
            >
              <Minimize2 size={13} />
            </button>
            <button
              type="button"
              onclick={handleDismiss}
              class="p-1 rounded-md text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition cursor-pointer"
              title="Hide floating tracker"
            >
              <X size={13} />
            </button>
          </div>
        </div>

        <!-- Filter Tabs (shown when subagents have execution plans) -->
        {#if hasSubagentPlans}
          <div class="flex items-center space-x-1 px-3 py-1.5 bg-ant-bg-secondary/70 dark:bg-white/[0.02] border-b border-ant-border-secondary dark:border-white/[0.04] overflow-x-auto scrollbar-none text-[10.5px] font-mono">
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
                <span class="max-w-[100px] truncate">{sp.subagentName}</span>
                <span class="text-[9.5px] opacity-75">({sp.completedCount}/{sp.totalCount})</span>
              </button>
            {/each}
          </div>
        {/if}

        <!-- Progress Bar Line -->
        <div class="w-full bg-ant-bg-tertiary h-1 overflow-hidden">
          <div
            class="h-full bg-gradient-to-r from-indigo-500 to-indigo-400 transition-all duration-300 ease-out"
            style="width: {planState.progressPercent}%;"
          ></div>
        </div>

        <!-- Scrollable Task List -->
        <div class="p-2.5 space-y-2 max-h-64 overflow-y-auto scrollbar-thin">
          {#if selectedTab === 'all' && hasSubagentPlans}
            <!-- Grouped by Main Plan & Subagents -->
            {#if hasMainPlan}
              <div class="space-y-1">
                <div class="flex items-center justify-between px-2 py-1 rounded bg-indigo-500/10 border border-indigo-500/20 text-[10px] font-mono text-indigo-700 dark:text-indigo-300 font-semibold">
                  <div class="flex items-center space-x-1.5 truncate">
                    <Layers size={11} />
                    <span>Main Orchestrator Plan</span>
                  </div>
                  <span>{mainTodos.filter(t => t.status === 'completed').length}/{mainTodos.length}</span>
                </div>
                {#each mainTodos as todo (todo.id)}
                  <div class="flex items-start space-x-2.5 p-2 rounded-lg bg-ant-bg-secondary/60 dark:bg-white/[0.02] border border-ant-border-secondary dark:border-white/[0.03] transition-colors hover:bg-ant-bg-tertiary/60 dark:hover:bg-white/[0.04]">
                    <!-- Status Icon -->
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

                    <!-- Task Content -->
                    <div class="flex-1 min-w-0 font-sans">
                      <div class="text-[12px] leading-relaxed select-text {todo.status === 'completed' ? 'line-through opacity-60 text-ant-text-muted' : 'text-ant-text'}">
                        {@html renderTodoContent(todo.content)}
                      </div>
                    </div>

                    <!-- Priority Badge (only if explicitly set) -->
                    {#if todo.priority}
                      <span class="text-[8.5px] uppercase tracking-wider font-mono px-1.5 py-0.2 rounded border shrink-0 {
                        todo.priority === 'high' || todo.priority === 'urgent' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-300 border-rose-500/20' :
                        todo.priority === 'medium' ? 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/20' :
                        'bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/20'
                      }">
                        {todo.priority}
                      </span>
                    {/if}
                  </div>
                {/each}
              </div>
            {/if}

            {#each subagentPlans as sp (sp.subagentId)}
              <div class="space-y-1">
                <div class="flex items-center justify-between px-2 py-1 rounded bg-violet-500/10 border border-violet-500/20 text-[10px] font-mono text-violet-700 dark:text-violet-300 font-semibold">
                  <div class="flex items-center space-x-1.5 truncate">
                    <Bot size={11} class="text-violet-500 dark:text-violet-400" />
                    <span class="truncate">{sp.subagentName.startsWith('Task') || sp.subagentName.startsWith('Subagent') ? sp.subagentName : `Subagent: ${sp.subagentName}`}</span>
                  </div>
                  <span>{sp.completedCount}/{sp.totalCount} ({sp.progressPercent}%)</span>
                </div>
                {#each sp.todos as todo (todo.id)}
                  <div class="flex items-start space-x-2.5 p-2 rounded-lg bg-ant-bg-secondary/60 dark:bg-white/[0.02] border border-ant-border-secondary dark:border-white/[0.03] transition-colors hover:bg-ant-bg-tertiary/60 dark:hover:bg-white/[0.04]">
                    <!-- Status Icon -->
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

                    <!-- Task Content -->
                    <div class="flex-1 min-w-0 font-sans">
                      <div class="text-[12px] leading-relaxed select-text {todo.status === 'completed' ? 'line-through opacity-60 text-ant-text-muted' : 'text-ant-text'}">
                        {@html renderTodoContent(todo.content)}
                      </div>
                    </div>

                    <!-- Priority Badge (only if explicitly set) -->
                    {#if todo.priority}
                      <span class="text-[8.5px] uppercase tracking-wider font-mono px-1.5 py-0.2 rounded border shrink-0 text-violet-700 dark:text-violet-300 bg-violet-500/10 border-violet-500/20">
                        {todo.priority}
                      </span>
                    {/if}
                  </div>
                {/each}
              </div>
            {/each}
          {:else}
            <!-- Flat / Tab-filtered List -->
            {@const activeDisplayTodos = selectedTab === 'main' ? mainTodos : (selectedTab !== 'all' ? (subagentPlans.find(s => s.subagentId === selectedTab)?.todos || []) : planState.todos)}
            {#each activeDisplayTodos as todo (todo.id)}
              <div class="flex items-start space-x-2.5 p-2 rounded-lg bg-ant-bg-secondary/60 dark:bg-white/[0.02] border border-ant-border-secondary dark:border-white/[0.03] transition-colors hover:bg-ant-bg-tertiary/60 dark:hover:bg-white/[0.04]">
                <!-- Status Icon -->
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

                <!-- Task Content -->
                <div class="flex-1 min-w-0 font-sans">
                  <div class="text-[12px] leading-relaxed select-text {todo.status === 'completed' ? 'line-through opacity-60 text-ant-text-muted' : 'text-ant-text'}">
                    {@html renderTodoContent(todo.content)}
                  </div>
                  {#if todo.subagentName && selectedTab === 'all' && !hasSubagentPlans}
                    <p class="text-[9.5px] font-mono text-violet-600 dark:text-violet-400 truncate mt-0.5">
                      Subagent: {todo.subagentName}
                    </p>
                  {/if}
                </div>

                <!-- Priority Badge (only if explicitly set) -->
                {#if todo.priority}
                  <span class="text-[8.5px] uppercase tracking-wider font-mono px-1.5 py-0.2 rounded border shrink-0 {
                    todo.priority === 'high' || todo.priority === 'urgent' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-300 border-rose-500/20' :
                    todo.priority === 'medium' ? 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/20' :
                    'bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/20'
                  }">
                    {todo.priority}
                  </span>
                {/if}
              </div>
            {/each}
          {/if}
        </div>

        <!-- Footer: Summary Count & Compact Trigger -->
        <div class="px-3 py-2 bg-ant-bg-secondary/80 dark:bg-white/[0.02] border-t border-ant-border-secondary dark:border-white/[0.05] flex items-center justify-between text-[10.5px] font-mono text-ant-text-secondary">
          <div class="flex items-center space-x-2">
            <span class="text-emerald-600 dark:text-emerald-400 font-semibold">{planState.completedCount} done</span>
            <span>•</span>
            <span class="text-amber-600 dark:text-amber-400 font-semibold">{planState.inProgressCount} active</span>
            <span>•</span>
            <span class="text-ant-text-muted">{planState.pendingCount} left</span>
          </div>
          <button
            type="button"
            onclick={handleToggleCompact}
            class="text-indigo-600 dark:text-indigo-400 hover:text-indigo-500 dark:hover:text-indigo-300 transition text-[10px] cursor-pointer"
          >
            Compact Pill
          </button>
        </div>
      </div>
    {/if}
  </div>
{/if}
