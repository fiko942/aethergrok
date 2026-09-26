<script lang="ts">
  import { sessionStore, STATUS_META, type Session } from '$lib/stores/session.svelte';
  import { settingsStore } from '$lib/stores/settings.svelte';
  import { Plus, X, GitFork, Edit2, Check, Loader2, AlertCircle, CheckCircle2, MessageSquare } from 'lucide-svelte';

  let draggedIndex = $state<number | null>(null);
  let dragOverIndex = $state<number | null>(null);
  let editingId = $state<string | null>(null);
  let editTitleInput = $state<string>('');

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

  function handleNewSession() {
    sessionStore.createSession();
  }
</script>

<div class="flex items-center w-full bg-ant-bg-secondary border-b border-white/5 px-2 h-10 select-none overflow-x-auto no-scrollbar gap-1.5 font-serif">
  <div class="flex items-center space-x-1 flex-1 min-w-0 overflow-x-auto">
    {#each sessionStore.openWorkspaceTabs as session, index (session.id)}
      {@const isActive = sessionStore.activeSessionId === session.id}
      {@const meta = STATUS_META[session.status]}
      {@const isDragging = draggedIndex === index}
      {@const isOver = dragOverIndex === index}

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
        class="group relative flex items-center h-8 pl-2.5 pr-2 rounded-md text-xs font-medium cursor-pointer transition-all duration-200 border max-w-[200px] min-w-[120px] flex-shrink-0 {isActive
          ? 'bg-ant-bg text-ant-primary border-white/[0.08] shadow-sm font-semibold'
          : 'bg-ant-bg-tertiary/40 text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border-transparent'} {isDragging ? 'opacity-40 scale-95' : ''} {isOver ? 'border-r-2 border-r-ant-primary' : ''} {settingsStore.animationsEnabled ? 'tab-peel-transition' : ''}"
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
          <span
            role="button"
            tabindex="0"
            class="truncate flex-1 min-w-0 font-serif text-[12px]"
            ondblclick={(e) => startEditing(session, e)}
            onkeydown={(e) => e.key === 'F2' && startEditing(session, e as unknown as MouseEvent)}
            title={`${session.title} (Double click to rename)`}
          >
            {session.title}
          </span>
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

  <!-- New Session Action Button -->
  <button
    type="button"
    onclick={handleNewSession}
    class="flex items-center justify-center w-7 h-7 rounded hover:bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-primary border border-transparent hover:border-ant-border transition flex-shrink-0"
    title="Create new session (Ctrl+N / Cmd+N)"
  >
    <Plus size={15} />
  </button>
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
