<script lang="ts">
  import { sessionStore, STATUS_META, type Session } from '$lib/stores/session.svelte';
  import { Plus, X, GitFork, Edit2, Check } from 'lucide-svelte';

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
        class="group relative flex items-center h-8 pl-2.5 pr-2 rounded-md text-xs font-medium cursor-pointer transition-all duration-150 border max-w-[200px] min-w-[120px] flex-shrink-0 {isActive
          ? 'bg-ant-bg text-ant-primary border-white/10 shadow-sm font-semibold'
          : 'bg-ant-bg-tertiary/40 text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border-transparent'} {isDragging ? 'opacity-40 scale-95' : ''} {isOver ? 'border-r-2 border-r-ant-primary' : ''}"
      >
        <!-- Status Indicator Dot (Ant Design palette) -->
        <span
          class="w-2 h-2 rounded-full mr-2 flex-shrink-0 {meta.dotClass}"
          title={`Status: ${meta.label}`}
        ></span>

        <!-- Title or Edit Input -->
        {#if editingId === session.id}
          <div class="flex items-center flex-1 min-w-0 mr-1">
            <input
              type="text"
              bind:value={editTitleInput}
              onkeydown={(e) => handleKeyDown(e, session)}
              onblur={() => saveEditing(session)}
              class="w-full bg-ant-bg-secondary text-white px-1 py-0.5 rounded text-xs outline-none border border-ant-primary/40"
            />
            <button
              type="button"
              onclick={() => saveEditing(session)}
              class="p-0.5 ml-1 text-ant-success hover:text-white"
            >
              <Check size={12} />
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
        <div class="items-center space-x-1 ml-1 flex-shrink-0 hidden group-hover:flex">
          {#if editingId !== session.id}
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
          {/if}

          <button
            type="button"
            onclick={(e) => handleClose(session, e)}
            class="p-0.5 rounded text-ant-text-muted hover:text-ant-error hover:bg-ant-bg-secondary transition"
            title="Close session tab"
          >
            <X size={12} />
          </button>
        </div>

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
