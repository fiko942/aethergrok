<script lang="ts">
  import { sessionStore } from '$lib/stores/session.svelte';
  import { Folder, GitBranch, X, PanelRightClose } from 'lucide-svelte';
  import FileExplorerTree from '../workspace/FileExplorerTree.svelte';
  import FileViewerModal from '../workspace/FileViewerModal.svelte';
  import GitChangesPanel from '../workspace/GitChangesPanel.svelte';

  let {
    workspacePath = '',
    sessionId,
  }: {
    workspacePath: string;
    sessionId?: string;
  } = $props();

  const currentSession = $derived(
    sessionId ? sessionStore.sessions.find((s) => s.id === sessionId) : sessionStore.activeSession
  );

  const isOpen = $derived(currentSession?.rightSidebarOpen ?? false);
  const activeTab = $derived(currentSession?.rightSidebarTab ?? 'files');

  // Drag resizing logic
  const DEFAULT_WIDTH = 340;
  const MIN_WIDTH = 260;
  const MAX_WIDTH = 600;
  const SNAP_CLOSE_WIDTH = 180;

  let width = $state(DEFAULT_WIDTH);
  let isDragging = $state(false);
  let startX = 0;
  let startWidth = DEFAULT_WIDTH;

  // Selected file for viewer modal
  let selectedFile = $state<string | null>(null);

  function startResize(e: MouseEvent) {
    e.preventDefault();
    isDragging = true;
    startX = e.clientX;
    startWidth = width;

    window.addEventListener('mousemove', handleResizeMove);
    window.addEventListener('mouseup', handleResizeEnd);
  }

  function handleResizeMove(e: MouseEvent) {
    if (!isDragging) return;
    // Moving mouse left increases right sidebar width
    const delta = startX - e.clientX;
    const newWidth = startWidth + delta;

    if (newWidth < SNAP_CLOSE_WIDTH) {
      // Snap to close
      if (currentSession?.id) {
        sessionStore.toggleRightSidebar(currentSession.id);
      }
      handleResizeEnd();
      return;
    }

    width = Math.min(Math.max(newWidth, MIN_WIDTH), MAX_WIDTH);
  }

  function handleResizeEnd() {
    isDragging = false;
    window.removeEventListener('mousemove', handleResizeMove);
    window.removeEventListener('mouseup', handleResizeEnd);
  }

  function resetWidth() {
    width = DEFAULT_WIDTH;
  }

  function handleClose() {
    if (currentSession?.id) {
      sessionStore.toggleRightSidebar(currentSession.id);
    }
  }

  function setTab(tab: 'files' | 'changes') {
    if (currentSession?.id) {
      sessionStore.setRightSidebarTab(tab, currentSession.id);
    }
  }
</script>

{#if isOpen}
  <aside
    style="width: {width}px;"
    class="relative h-full flex flex-col border-l border-[#27272a] bg-[#18181b] z-20 shrink-0 select-none {isDragging ? 'select-none pointer-events-none' : ''}"
  >
    <!-- Left Drag Divider Handle -->
    <div
      role="separator"
      tabindex="0"
      onmousedown={startResize}
      ondblclick={resetWidth}
      title="Drag to resize, double click to reset"
      class="absolute left-0 top-0 bottom-0 w-1.5 -translate-x-1/2 cursor-col-resize hover:bg-indigo-500/50 group z-30 transition-colors {isDragging ? 'bg-indigo-500' : ''}"
    >
      <div class="h-full w-full"></div>
    </div>

    <!-- Sidebar Header Tabs -->
    <div class="h-10 px-2 border-b border-[#27272a] bg-[#141416] flex items-center justify-between shrink-0">
      <div class="flex items-center gap-1 bg-[#1e1e22] border border-[#27272a] p-0.5 rounded-lg text-xs">
        <button
          onclick={() => setTab('files')}
          class="flex items-center gap-1.5 px-2.5 py-1 rounded-md transition-colors {activeTab === 'files' ? 'bg-zinc-800 text-zinc-100 shadow-sm font-medium' : 'text-zinc-400 hover:text-zinc-200'}"
        >
          <Folder class="w-3.5 h-3.5 text-amber-400" />
          <span>Explorer</span>
        </button>

        <button
          onclick={() => setTab('changes')}
          class="flex items-center gap-1.5 px-2.5 py-1 rounded-md transition-colors {activeTab === 'changes' ? 'bg-zinc-800 text-zinc-100 shadow-sm font-medium' : 'text-zinc-400 hover:text-zinc-200'}"
        >
          <GitBranch class="w-3.5 h-3.5 text-indigo-400" />
          <span>Changes</span>
        </button>
      </div>

      <button
        onclick={handleClose}
        title="Close Inspector (⌘⌥B)"
        class="p-1.5 rounded-lg text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
      >
        <PanelRightClose class="w-4 h-4" />
      </button>
    </div>

    <!-- Active Tab Body -->
    <div class="flex-1 overflow-hidden">
      {#if activeTab === 'files'}
        <FileExplorerTree
          workspacePath={workspacePath}
          onSelectFile={(path) => (selectedFile = path)}
        />
      {:else}
        <GitChangesPanel workspacePath={workspacePath} />
      {/if}
    </div>
  </aside>
{/if}

<!-- Read-Only File Viewer Modal mounted outside aside to avoid stacking context traps -->
{#if selectedFile}
  <FileViewerModal
    isOpen={true}
    filePath={selectedFile}
    workspacePath={workspacePath}
    onClose={() => (selectedFile = null)}
  />
{/if}
