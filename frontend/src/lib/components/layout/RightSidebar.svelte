<script lang="ts">
  import { sessionStore } from '$lib/stores/session.svelte';
  import { settingsStore } from '$lib/stores/settings.svelte';
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
    e.stopPropagation();
    isDragging = true;
    startX = e.clientX;
    startWidth = width;
    document.body.style.userSelect = 'none';
    document.body.style.webkitUserSelect = 'none';
    document.body.style.cursor = 'col-resize';
    if (window.getSelection) {
      window.getSelection()?.removeAllRanges();
    }

    window.addEventListener('mousemove', handleResizeMove);
    window.addEventListener('mouseup', handleResizeEnd);
  }

  function handleResizeMove(e: MouseEvent) {
    if (!isDragging) return;
    e.preventDefault();
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
    if (isDragging) {
      isDragging = false;
      document.body.style.userSelect = '';
      document.body.style.webkitUserSelect = '';
      document.body.style.cursor = '';
      window.removeEventListener('mousemove', handleResizeMove);
      window.removeEventListener('mouseup', handleResizeEnd);
    }
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
    class="relative h-full flex flex-col border-l border-ant-border bg-ant-bg z-20 shrink-0 select-none {isDragging ? 'select-none pointer-events-none transition-none' : (settingsStore.animationsEnabled ? 'transition-all duration-300 ease-out' : '')}"
  >
    <!-- Left Drag Divider Handle -->
    <div
      role="separator"
      tabindex="0"
      onmousedown={startResize}
      ondblclick={resetWidth}
      title="Drag to resize, double click to reset"
      class="absolute left-0 top-0 bottom-0 w-1.5 -translate-x-1/2 cursor-col-resize hover:bg-ant-primary/50 group z-30 transition-colors {isDragging ? 'bg-ant-primary' : ''}"
    >
      <div class="h-full w-full"></div>
    </div>

    <!-- Sidebar Header Tabs -->
    <div class="h-10 px-2 border-b border-ant-border bg-ant-bg-secondary flex items-center justify-between shrink-0">
      <div class="flex items-center gap-1 bg-ant-bg-tertiary border border-ant-border p-0.5 rounded-lg text-xs">
        <button
          onclick={() => setTab('files')}
          class="flex items-center gap-1.5 px-2.5 py-1 rounded-md transition-colors {activeTab === 'files' ? 'bg-ant-bg text-ant-text shadow-sm font-medium' : 'text-ant-text-secondary hover:text-ant-text'}"
        >
          <Folder class="w-3.5 h-3.5 text-amber-500" />
          <span>Explorer</span>
        </button>

        <button
          onclick={() => setTab('changes')}
          class="flex items-center gap-1.5 px-2.5 py-1 rounded-md transition-colors {activeTab === 'changes' ? 'bg-ant-bg text-ant-text shadow-sm font-medium' : 'text-ant-text-secondary hover:text-ant-text'}"
        >
          <GitBranch class="w-3.5 h-3.5 text-ant-primary" />
          <span>Changes</span>
        </button>
      </div>

      <button
        onclick={handleClose}
        title="Close Inspector (⌘⌥B)"
        class="p-1.5 rounded-lg text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition-colors"
      >
        <PanelRightClose class="w-4 h-4" />
      </button>
    </div>

    <!-- Active Tab Body with Build / Peel Animation -->
    <div class="flex-1 overflow-hidden {settingsStore.animationsEnabled ? 'tab-build-container' : ''}">
      {#if activeTab === 'files'}
        <div class="h-full {settingsStore.animationsEnabled ? 'animate-tab-build-in' : ''}">
          <FileExplorerTree
            workspacePath={workspacePath}
            onSelectFile={(path) => (selectedFile = path)}
          />
        </div>
      {:else}
        <div class="h-full {settingsStore.animationsEnabled ? 'animate-tab-build-in' : ''}">
          <GitChangesPanel workspacePath={workspacePath} />
        </div>
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
