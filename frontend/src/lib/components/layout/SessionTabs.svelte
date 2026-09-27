<script lang="ts">
  import { sessionStore, STATUS_META, type Session } from '$lib/stores/session.svelte';
  import { settingsStore } from '$lib/stores/settings.svelte';
  import { terminalStore } from '$lib/stores/terminal.svelte';
  import { Plus, X, GitFork, Edit2, Check, Loader2, AlertCircle, CheckCircle2, MessageSquare, Folder, Terminal } from 'lucide-svelte';
  import NewSessionDropdown from './NewSessionDropdown.svelte';

  let draggedIndex = $state<number | null>(null);
  let dragOverIndex = $state<number | null>(null);
  let editingId = $state<string | null>(null);
  let editTitleInput = $state<string>('');
  let isDropdownOpen = $state<boolean>(false);

  function handleDragStart(e: DragEvent, index: number) {
    draggedIndex = index;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', index.toString());
    }
  }

  function handleDragOver(e: DragEvent, index: number) {
    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }
    dragOverIndex = index;
  }

  function handleDrop(e: DragEvent, targetIndex: number) {
    e.preventDefault();
    if (draggedIndex !== null && draggedIndex !== targetIndex) {
      sessionStore.reorderSessions(draggedIndex, targetIndex);
    }
    draggedIndex = null;
    dragOverIndex = null;
  }

  function handleDragEnd() {
    draggedIndex = null;
    dragOverIndex = null;
  }

  function startEditing(session: Session, e: MouseEvent) {
    e.stopPropagation();
    editingId = session.id;
    editTitleInput = session.title;
  }

  function saveEditing(session: Session) {
    if (editingId === session.id) {
      if (editTitleInput.trim()) {
        sessionStore.renameSession(session.id, editTitleInput.trim());
      }
      editingId = null;
    }
  }

  function handleKeyDown(e: KeyboardEvent, session: Session) {
    if (e.key === 'Enter') {
      saveEditing(session);
    } else if (e.key === 'Escape') {
      editingId = null;
    }
  }

  function handleFork(session: Session, e: MouseEvent) {
    e.stopPropagation();
    sessionStore.forkSession(session.id);
  }

  function handleClose(session: Session, e: MouseEvent) {
    e.stopPropagation();
    sessionStore.closeSessionTab(session.id);
  }

  function handleNewSession(e: MouseEvent) {
    e.stopPropagation();
    isDropdownOpen = !isDropdownOpen;
  }
</script>

<div class="flex items-center w-full bg-ant-bg-secondary border-b border-ant-border-secondary dark:border-white/5 px-2 h-10 select-none gap-1.5 font-serif relative overflow-visible z-20">
  <div class="flex items-center space-x-1 flex-1 min-w-0 overflow-x-auto no-scrollbar py-1">
    {#each sessionStore.openTabs as session, index (session.id)}
      {@const isActive = sessionStore.activeSessionId === session.id}
      {@const meta = STATUS_META[session.status]}
      {@const isDragging = draggedIndex === index}
      {@const isOver = dragOverIndex === index}
      {@const ws = sessionStore.workspaces.find((w) => w.id === session.workspaceId)}
      {@const showWsBadge = sessionStore.workspaces.filter((w) => w.existsOnDisk !== false).length > 1 && ws}

      <div
        role="tab"
        tabindex="0"
        aria-selected={isActive}
        draggable={editingId !== session.id}
        ondragstart={(e) => handleDragStart(e, index)}
        ondragover={(e) => handleDragOver(e, index)}
        ondrop={(e) => handleDrop(e, index)}
        ondragend={handleDragEnd}
        onclick={() => sessionStore.switchSession(session.id)}
        onkeydown={(e) => e.key === 'Enter' && sessionStore.switchSession(session.id)}
        class="group relative flex items-center h-8 pl-2.5 pr-2 rounded-md text-xs font-medium cursor-pointer transition-all duration-200 border max-w-[220px] min-w-[120px] flex-shrink-0 {isActive
          ? 'bg-ant-bg text-ant-primary border-ant-border-secondary dark:border-white/10 shadow-2xs font-semibold'
          : 'bg-ant-bg-tertiary/20 text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary/60 border-transparent hover:border-ant-border-secondary dark:hover:border-white/5'} {isDragging ? 'opacity-40 scale-95' : ''} {isOver ? 'border-r-2 border-r-ant-primary' : ''} {settingsStore.animationsEnabled ? 'tab-peel-transition' : ''}"
      >
        <!-- Status Icon Matching Sidebar -->
        {#if session.status === 'working'}
          <span class="flex items-center justify-center flex-shrink-0 mr-1.5 text-ant-primary" title="Status: Working / Generating...">
            <Loader2 size={12.5} class="animate-spin" />
          </span>
        {:else if session.status === 'waiting_permission'}
          <span class="flex items-center justify-center flex-shrink-0 mr-1.5 text-amber-400" title="Status: Waiting Permission">
            <AlertCircle size={12.5} class="animate-bounce" />
          </span>
        {:else if session.status === 'finished'}
          <span class="flex items-center justify-center flex-shrink-0 mr-1.5 text-ant-success/80 group-hover:text-ant-success transition" title="Status: Completed">
            <CheckCircle2 size={12.5} />
          </span>
        {:else}
          <span class="flex items-center justify-center flex-shrink-0 mr-1.5 {isActive ? 'text-ant-primary' : 'text-ant-text-muted/70 group-hover:text-ant-text-secondary'} transition" title="Chat Session">
            <MessageSquare size={12.5} />
          </span>
        {/if}

        <!-- Title or Edit Input -->
        {#if editingId === session.id}
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div
            class="flex items-center flex-1 min-w-0 mr-1 gap-1"
            onclick={(e) => e.stopPropagation()}
            onkeydown={(e) => e.stopPropagation()}
          >
            <input
              type="text"
              bind:value={editTitleInput}
              onkeydown={(e) => handleKeyDown(e, session)}
              class="w-full bg-ant-bg-tertiary text-ant-text px-1.5 py-0.5 rounded text-[11.5px] outline-none border border-ant-primary/40 focus:border-ant-primary transition font-serif"
              autofocus
            />
            <!-- Checkmark Confirm Button -->
            <button
              type="button"
              onclick={() => saveEditing(session)}
              class="p-1 rounded text-zinc-400 hover:text-zinc-200 hover:bg-white/5 transition flex-shrink-0"
              title="Apply changes (Enter)"
            >
              <Check size={12} class="stroke-[2.2]" />
            </button>
            <!-- Cross Cancel Button (Always visible during edit) -->
            <button
              type="button"
              onclick={() => editingId = null}
              class="p-1 rounded text-zinc-400 hover:text-rose-400 hover:bg-white/5 transition flex-shrink-0"
              title="Cancel (Esc)"
            >
              <X size={12} class="stroke-[2.2]" />
            </button>
          </div>
        {:else}
          <div class="flex items-center min-w-0 flex-1 gap-1">
            <span
              role="button"
              tabindex="0"
              class="truncate flex-1 min-w-0 font-serif text-[12px]"
              ondblclick={(e) => startEditing(session, e)}
              onkeydown={(e) => e.key === 'F2' && startEditing(session, e as unknown as MouseEvent)}
              title={`${session.title}${ws ? ` (${ws.name})` : ''} - Double click to rename`}
            >
              {session.title}
            </span>
            {#if showWsBadge}
              <span
                class="flex-shrink-0 text-[9.5px] font-mono px-1 py-0.2 rounded bg-ant-bg-tertiary/70 text-ant-text-muted border border-ant-border-secondary dark:border-white/5 max-w-[65px] truncate"
                title={`Workspace: ${ws?.name} (${ws?.path})`}
              >
                {ws?.name}
              </span>
            {/if}
          </div>
        {/if}

        <!-- Hover Actions: Fork, Rename & Close with fixed space reservation -->
        {#if editingId !== session.id}
          <div class="items-center space-x-1 ml-1 flex-shrink-0 hidden group-hover:flex">
            <button
              type="button"
              onclick={(e) => startEditing(session, e)}
              class="p-0.5 rounded text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-secondary transition"
              title="Rename tab"
            >
              <Edit2 size={11} />
            </button>
            <button
              type="button"
              onclick={(e) => handleFork(session, e)}
              class="p-0.5 rounded text-ant-text-muted hover:text-ant-primary hover:bg-ant-bg-secondary transition"
              title="Fork session to new tab (duplicate conversation history)"
            >
              <GitFork size={11} />
            </button>
            <button
              type="button"
              onclick={(e) => handleClose(session, e)}
              class="p-0.5 rounded text-ant-text-muted hover:text-ant-error hover:bg-ant-bg-secondary transition"
              title="Close tab (⌘W / Ctrl+W)"
            >
              <X size={12} />
            </button>
          </div>
        {/if}

        <!-- Active Bottom Line Accent -->
        {#if isActive}
          <div class="absolute -bottom-[1px] left-0 right-0 h-[2px] bg-ant-primary rounded-t"></div>
        {/if}
      </div>
    {/each}
  </div>

  <!-- Right Actions: Terminal Toggle + New Session Action Button -->
  <div class="relative flex items-center space-x-1 flex-shrink-0 z-30">
    <!-- Tab-Isolated Terminal Toggle Button -->
    {#if sessionStore.activeSession}
      {@const activeSessionId = sessionStore.activeSession.id}
      {@const hasTerminals = terminalStore.getTerminalTabs(activeSessionId).length > 0}
      {@const isCollapsed = terminalStore.isSessionCollapsed(activeSessionId)}
      <button
        type="button"
        onclick={() => {
          if (!hasTerminals) {
            terminalStore.createTerminal(activeSessionId, sessionStore.activeWorkspace?.path || '');
            terminalStore.toggleSessionCollapse(activeSessionId, false);
          } else {
            terminalStore.toggleSessionCollapse(activeSessionId);
          }
        }}
        class="flex items-center space-x-1 px-2 h-7 rounded text-xs transition {hasTerminals && !isCollapsed ? 'bg-ant-primary/15 text-ant-primary font-medium' : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary'}"
        title="Toggle Session Terminal (Multi-tab PTY with full process cleanup)"
      >
        <Terminal size={13} />
        <span class="text-[11px]">Terminal</span>
      </button>
    {/if}

    <!-- New Session Action Button with Workspace Dropdown Picker -->
    <button
      type="button"
      onclick={handleNewSession}
      class="flex items-center justify-center w-7 h-7 rounded bg-ant-bg hover:bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-primary border border-ant-border-secondary dark:border-white/5 transition flex-shrink-0 cursor-pointer shadow-2xs {isDropdownOpen ? 'bg-ant-bg-tertiary text-ant-primary' : ''}"
      title="Create new conversation (Choose workspace)"
    >
      <Plus size={15} />
    </button>

    <NewSessionDropdown
      bind:open={isDropdownOpen}
      placement="bottom-end"
      onClose={() => isDropdownOpen = false}
    />
  </div>
</div>

<style>
  .no-scrollbar::-webkit-scrollbar {
    display: none;
  }
  .no-scrollbar {
    -ms-overflow-style: none;
    scrollbar-width: none;
  }
</style>
