<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { sessionStore, STATUS_META, type Session, type WorkspaceFolder } from '$lib/stores/session.svelte';
  import { settingsStore } from '$lib/stores/settings.svelte';
  import { dialogStore } from '$lib/stores/dialog.svelte';
  import { formatSessionAsMarkdown, downloadOrSaveMarkdown } from '$lib/utils/markdownExport';
  import BatchActionBar from './BatchActionBar.svelte';
  import {
    Folder,
    FolderPlus,
    Plus,
    MoreVertical,
    FileText,
    Copy,
    Tag,
    Trash2,
    Check,
    Square,
    ChevronDown,
    ChevronRight,
    FolderOpen,
    Edit2,
    Pin,
    Search,
    RefreshCw,
    MessageSquare,
    Loader2,
    AlertCircle,
    CheckCircle2
  } from 'lucide-svelte';
  import CustomCheckbox from '$lib/components/ui/CustomCheckbox.svelte';

  let activeDropdownId = $state<string | null>(null);
  let activeDropdownCoords = $state<{ top: number; right: number } | null>(null);
  let editingSessionId = $state<string | null>(null);
  let editTitleText = $state('');
  let copiedToast = $state<string | null>(null);
  let searchQuery = $state('');

  // Track pagination count per workspace (defaults to 8 items, loads +8 each click)
  const PAGE_SIZE = 8;
  let workspaceVisibleCounts = $state<Record<string, number>>({});

  function getVisibleLimit(wsId: string): number {
    return workspaceVisibleCounts[wsId] ?? PAGE_SIZE;
  }

  function handleShowMore(wsId: string) {
    const current = getVisibleLimit(wsId);
    workspaceVisibleCounts = {
      ...workspaceVisibleCounts,
      [wsId]: current + PAGE_SIZE
    };
  }

  function handleShowLess(wsId: string) {
    workspaceVisibleCounts = {
      ...workspaceVisibleCounts,
      [wsId]: PAGE_SIZE
    };
  }

  function handleToggleWorkspace(wsId: string) {
    // When closing or reopening an accordion folder, reset pagination view count back to 8
    workspaceVisibleCounts = {
      ...workspaceVisibleCounts,
      [wsId]: PAGE_SIZE
    };
    sessionStore.toggleWorkspaceExpanded(wsId);
  }

  function handleSelectSession(sessionId: string) {
    if (sessionStore.isSelectionMode) {
      sessionStore.toggleSessionSelected(sessionId);
    } else {
      sessionStore.openSessionInTab(sessionId);
      // If viewport is in compact mode (< 840px), auto-collapse the sidebar drawer
      if (typeof window !== 'undefined' && window.innerWidth < 840) {
        settingsStore.sidebarCollapsed = true;
        settingsStore.saveToStorage();
      }
    }
  }

  function toggleDropdown(sessionId: string, e: MouseEvent) {
    e.stopPropagation();
    if (activeDropdownId === sessionId) {
      activeDropdownId = null;
      activeDropdownCoords = null;
    } else {
      activeDropdownId = sessionId;
      const target = (e.currentTarget as HTMLElement) || (e.target as HTMLElement);
      const rect = target.getBoundingClientRect();
      const menuHeight = 220; // approximate menu height
      const windowHeight = window.innerHeight;

      // Flip upwards if near bottom
      let top = rect.bottom + 4;
      if (top + menuHeight > windowHeight) {
        top = Math.max(10, rect.top - menuHeight - 4);
      }
      activeDropdownCoords = {
        top,
        right: window.innerWidth - rect.right
      };
    }
  }

  function showToast(text: string) {
    copiedToast = text;
    setTimeout(() => {
      copiedToast = null;
    }, 2200);
  }

  // Handle global click-outside and Escape key for 3-dots action dropdown
  function handleWindowPointerDown(e: PointerEvent) {
    if (activeDropdownId) {
      const target = e.target as HTMLElement | null;
      if (target && !target.closest('.session-dropdown-menu') && !target.closest('.session-more-btn')) {
        activeDropdownId = null;
        activeDropdownCoords = null;
      }
    }
  }

  function handleWindowKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      activeDropdownId = null;
      activeDropdownCoords = null;
      editingSessionId = null;
    }
  }

  // Auto-sync sessions on mount for all known workspaces & register click-outside handlers
  onMount(async () => {
    await sessionStore.verifyAllWorkspaces();
    for (const ws of sessionStore.workspaces) {
      if (ws.existsOnDisk !== false) {
        syncGrokSessionsForWorkspace(ws.id, ws.path);
      }
    }
    window.addEventListener('pointerdown', handleWindowPointerDown);
    window.addEventListener('keydown', handleWindowKeyDown);
  });

  onDestroy(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('pointerdown', handleWindowPointerDown);
      window.removeEventListener('keydown', handleWindowKeyDown);
    }
  });

  // Scan and discover Grok sessions saved in ~/.grok/sessions
  async function syncGrokSessionsForWorkspace(wsId: string, wsPath: string) {
    if (window.go?.main?.App?.DiscoverGrokSessions) {
      try {
        const sessions = await window.go.main.App.DiscoverGrokSessions(wsPath);
        if (sessions && sessions.length > 0) {
          sessionStore.syncDiscoveredGrokSessions(wsId, sessions);
        }
      } catch (err) {
        console.error('Failed to discover Grok sessions:', err);
      }
    }
  }

  // Add workspace folder using native OS folder picker or prompt
  async function handleOpenWorkspaceFolder() {
    if (window.go?.main?.App?.SelectWorkspaceDirectory) {
      try {
        const dir = await window.go.main.App.SelectWorkspaceDirectory();
        if (dir) {
          const folderName = dir.split(/[/\\]/).filter(Boolean).pop() || 'workspace';
          const newWs = sessionStore.addWorkspace(folderName, dir);
          showToast(`Workspace "${folderName}" dibuka`);
          syncGrokSessionsForWorkspace(newWs.id, dir);
        }
      } catch (err) {
        console.error('Failed to select directory:', err);
      }
    } else {
      const defaultPath = typeof navigator !== 'undefined' && /Win/.test(navigator.platform || navigator.userAgent) ? 'C:\\workspace' : '/workspace';
      const path = window.prompt('Masukkan absolute path folder workspace:', defaultPath);
      if (path && path.trim()) {
        const folderName = path.trim().split(/[/\\]/).filter(Boolean).pop() || 'workspace';
        const newWs = sessionStore.addWorkspace(folderName, path.trim());
        showToast(`Workspace "${folderName}" ditambahkan`);
      }
    }
  }

  function handleCreateNewSession(wsId?: string) {
    sessionStore.createSession(undefined, wsId);
  }

  function handleStartRename(session: Session, e: MouseEvent) {
    e.stopPropagation();
    editingSessionId = session.id;
    editTitleText = session.title;
    activeDropdownId = null;
  }

  function handleSaveRename(session: Session) {
    if (editingSessionId === session.id) {
      if (editTitleText.trim()) {
        sessionStore.renameSession(session.id, editTitleText.trim());
      }
      editingSessionId = null;
    }
  }

  function handleTogglePin(session: Session, e: MouseEvent) {
    e.stopPropagation();
    sessionStore.togglePinSession(session.id);
    activeDropdownId = null;
    activeDropdownCoords = null;
    showToast(session.isPinned ? 'Session unpinned' : 'Session pinned');
  }

  async function handleCopySessionId(session: Session, e: MouseEvent) {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(session.id);
      showToast('Session ID copied');
    } catch {
      showToast(session.id);
    }
    activeDropdownId = null;
    activeDropdownCoords = null;
  }

  async function handleCopySessionName(session: Session, e: MouseEvent) {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(session.title);
      showToast('Session title copied');
    } catch {
      showToast(session.title);
    }
    activeDropdownId = null;
    activeDropdownCoords = null;
  }

  async function handleExportMarkdown(session: Session, e: MouseEvent) {
    e.stopPropagation();
    activeDropdownId = null;
    activeDropdownCoords = null;
    try {
      const ws = sessionStore.workspaces.find(w => w.id === session.workspaceId) || sessionStore.activeWorkspace;
      const mdContent = formatSessionAsMarkdown(session, ws);
      const safeTitle = session.title.toLowerCase().replace(/[^a-z0-9_-]/g, '-').slice(0, 30);
      const dateStr = new Date().toISOString().slice(0, 10);
      const filename = `${safeTitle}-${dateStr}.md`;

      const res = await downloadOrSaveMarkdown(filename, mdContent);
      if (res.success) {
        showToast('Exported to Markdown (.md)');
      } else if (res.error) {
        alert('Failed to export: ' + res.error);
      }
    } catch (err) {
      alert('Markdown export error: ' + String(err));
    }
  }

  function handleDeleteSession(session: Session, e: MouseEvent) {
    e.stopPropagation();
    activeDropdownId = null;
    activeDropdownCoords = null;
    const ws = sessionStore.workspaces.find(w => w.id === session.workspaceId);
    
    dialogStore.openConfirm({
      title: 'Delete Chat Session',
      content: `Are you sure you want to delete "${session.title}"? This conversation will be removed from AetherGrok and Grok CLI.`,
      confirmText: 'Delete',
      cancelText: 'Cancel',
      type: 'danger',
      onConfirm: () => {
        sessionStore.closeSession(session.id);
        if (ws && window.go?.main?.App?.DeleteGrokSession) {
          const targetGrokId = session.grokSessionId || session.id;
          window.go.main.App.DeleteGrokSession(ws.path, targetGrokId).catch(console.error);
        }
        showToast('Session deleted');
      }
    });
  }

  function handleRemoveWorkspace(ws: WorkspaceFolder, e: MouseEvent) {
    e.stopPropagation();
    dialogStore.openConfirm({
      title: 'Remove Workspace',
      content: `Remove "${ws.name}" from your workspace list? Files on disk (${ws.path}) will not be modified or deleted.`,
      confirmText: 'Remove',
      cancelText: 'Cancel',
      type: 'danger',
      onConfirm: () => {
        sessionStore.removeWorkspace(ws.id);
        showToast(`Workspace "${ws.name}" removed`);
      }
    });
  }

  // Filter workspaces and sessions based on search query
  const filteredWorkspaces = $derived.by(() => {
    const q = searchQuery.toLowerCase().trim();
    if (!q) return sessionStore.workspaces;

    return sessionStore.workspaces.filter(ws => {
      const matchWs = ws.name.toLowerCase().includes(q) || ws.path.toLowerCase().includes(q);
      const matchSession = sessionStore.sessions.some(s => s.workspaceId === ws.id && s.title.toLowerCase().includes(q));
      return matchWs || matchSession;
    });
  });

  // Global Pinned Sessions across all workspaces
  const globalPinnedSessions = $derived.by(() => {
    const q = searchQuery.toLowerCase().trim();
    let pinned = sessionStore.sessions.filter(s => s.isPinned);
    if (q) {
      pinned = pinned.filter(s => {
        const ws = sessionStore.workspaces.find(w => w.id === s.workspaceId);
        return s.title.toLowerCase().includes(q) || (ws && ws.name.toLowerCase().includes(q));
      });
    }
    return pinned.sort((a, b) => (b.pinnedAt || 0) - (a.pinnedAt || 0));
  });

  function getWorkspaceSessions(wsId: string): { visible: Session[]; totalCount: number; hasMore: boolean; hiddenCount: number; isExpanded: boolean } {
    const q = searchQuery.toLowerCase().trim();
    let wsSessions = sessionStore.sessions.filter(s => s.workspaceId === wsId && !s.isPinned);

    // Sort newest first based on updatedAt or createdAt
    const sorted = wsSessions.sort((a, b) => {
      const timeA = a.updatedAt || a.createdAt || 0;
      const timeB = b.updatedAt || b.createdAt || 0;
      return timeB - timeA;
    });

    if (q) {
      // When searching, match across all sessions (including ones normally hidden under Show More)
      const filtered = sorted.filter(s => s.title.toLowerCase().includes(q));
      return {
        visible: filtered,
        totalCount: filtered.length,
        hasMore: false,
        hiddenCount: 0,
        isExpanded: false
      };
    }

    const totalCount = sorted.length;
    const currentLimit = getVisibleLimit(wsId);
    const visible = sorted.slice(0, currentLimit);
    const hasMore = totalCount > currentLimit;
    const hiddenCount = totalCount - currentLimit;
    const isExpanded = currentLimit > PAGE_SIZE;

    return {
      visible,
      totalCount,
      hasMore,
      hiddenCount,
      isExpanded
    };
  }

  function getWorkspaceForSession(sessionId: string): WorkspaceFolder | undefined {
    const sess = sessionStore.sessions.find(s => s.id === sessionId);
    if (!sess) return undefined;
    return sessionStore.workspaces.find(w => w.id === sess.workspaceId);
  }
</script>

<div class="flex flex-col w-full h-full bg-ant-bg-secondary select-none relative z-20 font-serif">
  <!-- 1. Search Bar & Action Controls -->
  <div class="p-2.5 border-b border-ant-border space-y-2">
    <!-- Search Bar -->
    <div class="relative flex items-center">
      <Search size={13} class="absolute left-2.5 text-ant-text-muted" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Filter project & session..."
        class="w-full bg-ant-bg border border-ant-border rounded-md pl-8 pr-2.5 py-1 text-xs text-ant-text placeholder-ant-text-muted focus:border-ant-primary focus:bg-ant-bg focus:outline-none transition-colors shadow-2xs"
      />
    </div>

    <!-- Multi-select / Batch Action Toggle Bar -->
    <div class="flex items-center justify-between text-xs px-0.5">
      <button
        onclick={handleOpenWorkspaceFolder}
        class="flex items-center space-x-1.5 text-xs text-ant-primary hover:text-ant-primary-hover font-medium transition"
        title="Open Workspace Folder"
      >
        <FolderPlus size={14} />
        <span>Open Folder</span>
      </button>

      <button
        onclick={() => sessionStore.toggleSelectionMode()}
        class="text-[11px] font-medium px-2 py-0.5 rounded transition border {sessionStore.isSelectionMode ? 'bg-ant-primary text-white border-ant-primary shadow-xs' : 'text-ant-text-secondary hover:text-ant-text bg-ant-bg-tertiary/40 hover:bg-ant-bg-tertiary border-ant-border-secondary dark:border-white/5'}"
        title="Toggle Multi-Select"
      >
        {sessionStore.isSelectionMode ? 'Cancel' : 'Select'}
      </button>
    </div>
  </div>

  <!-- 2. Hierarchical Multi-Workspace Tree Explorer -->
  <div class="flex-1 overflow-y-auto p-2 space-y-2.5 custom-scrollbar">
    <!-- GLOBAL PINNED SESSIONS SECTION -->
    {#if globalPinnedSessions.length > 0}
      <div class="rounded-lg bg-amber-500/5 border border-amber-500/15 overflow-hidden pb-1">
        <div class="flex items-center justify-between px-2.5 py-1.5 bg-amber-500/10 border-b border-amber-500/15">
          <div class="flex items-center space-x-1.5">
            <Pin size={12} class="text-amber-400 fill-current" />
            <span class="text-[11px] font-semibold text-amber-300 uppercase tracking-wider">Pinned</span>
            <span class="text-[10px] text-amber-400/80 font-mono">({globalPinnedSessions.length})</span>
          </div>
        </div>

        <div class="px-0.5 py-0.5 space-y-0.5">
          {#each globalPinnedSessions as session (session.id)}
            {@const isActive = sessionStore.activeSessionId === session.id}
            {@const isChecked = sessionStore.selectedSessionIds.has(session.id)}
            {@const isEditing = editingSessionId === session.id}
            {@const ws = getWorkspaceForSession(session.id)}

            <div
              role="button"
              tabindex="0"
              onclick={() => handleSelectSession(session.id)}
              onkeydown={(e) => e.key === 'Enter' && handleSelectSession(session.id)}
              class="group relative flex items-center justify-between px-2.5 py-1.5 rounded-md text-[13px] font-sans cursor-pointer transition-colors duration-150 {isActive
                ? 'bg-ant-primary/15 text-ant-primary font-medium'
                : 'bg-ant-bg-secondary/40 text-ant-text/80 hover:text-ant-text hover:bg-ant-bg-tertiary/70'}"
            >
              <!-- Left: Title & Folder Badge -->
              <div class="flex flex-col min-w-0 flex-1 pr-1">
                <div class="flex items-center space-x-2.5 min-w-0">
                  {#if sessionStore.isSelectionMode}
                    <button
                      type="button"
                      onclick={(e) => { e.stopPropagation(); sessionStore.toggleSessionSelected(session.id); }}
                      class="flex-shrink-0"
                    >
                      <CustomCheckbox checked={isChecked} size="sm" />
                    </button>
                  {:else}
                    <Pin size={13} class="text-amber-400 fill-current flex-shrink-0" />
                  {/if}

                  {#if isEditing}
                    <input
                      type="text"
                      bind:value={editTitleText}
                      onkeydown={(e) => {
                        if (e.key === 'Enter') handleSaveRename(session);
                        if (e.key === 'Escape') editingSessionId = null;
                      }}
                      onblur={() => handleSaveRename(session)}
                      class="w-full bg-ant-bg border border-ant-primary rounded px-1.5 py-0.5 text-xs text-ant-text focus:outline-none"
                    />
                  {:else}
                    <span class="truncate leading-normal tracking-tight {isActive ? 'font-semibold text-ant-primary' : 'text-ant-text'}" title={session.title}>{session.title}</span>
                  {/if}
                </div>

                <!-- Workspace Badge Tag -->
                {#if ws}
                  <div class="flex items-center space-x-1 pl-5 pt-0.5">
                    <Folder size={10.5} class="text-ant-primary flex-shrink-0" />
                    <span class="text-[10px] text-ant-text-secondary font-medium truncate max-w-[140px]" title={ws.path}>{ws.name}</span>
                  </div>
                {/if}
              </div>

              <!-- Right Action Menu Trigger -->
              <div class="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0">
                <button
                  type="button"
                  onclick={(e) => toggleDropdown(session.id, e)}
                  class="session-more-btn p-0.5 text-ant-text-secondary hover:text-ant-text rounded hover:bg-ant-bg-tertiary transition"
                  title="Session Options"
                >
                  <MoreVertical size={13.5} />
                </button>
              </div>
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- WORKSPACE TREES -->
    {#if filteredWorkspaces.length === 0}
      <div class="flex flex-col items-center justify-center p-5 text-center my-6 space-y-3 rounded-xl border border-dashed border-ant-border-secondary dark:border-white/5 bg-ant-bg-secondary/30">
        <div class="w-10 h-10 rounded-xl bg-ant-primary/10 flex items-center justify-center text-ant-primary/70">
          <FolderOpen size={20} />
        </div>
        <div class="space-y-1">
          <p class="text-xs font-medium text-ant-text">No workspaces open</p>
          <p class="text-[11px] text-ant-text-muted leading-relaxed max-w-[190px]">
            Open any project directory on your machine to start collaborating with Grok.
          </p>
        </div>
        <button
          type="button"
          onclick={handleOpenWorkspaceFolder}
          class="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg text-xs font-medium bg-ant-primary text-white hover:bg-ant-primary/90 transition shadow-xs"
        >
          <FolderPlus size={13} />
          <span>Open Folder</span>
        </button>
      </div>
    {:else}
      {#each filteredWorkspaces as ws (ws.id)}
      {@const isExpanded = ws.isExpanded === true}
      {@const { visible, totalCount, hasMore, hiddenCount } = getWorkspaceSessions(ws.id)}
      {@const isMissing = ws.existsOnDisk === false}

      <div class="rounded-lg {isMissing ? 'bg-rose-500/5 border border-rose-500/20' : 'bg-ant-bg border border-ant-border-secondary dark:border-white/5'} overflow-hidden shadow-2xs">
        <!-- Workspace Folder Header -->
        <div
          role="button"
          tabindex="0"
          onclick={() => {
            if (!isMissing) {
              handleToggleWorkspace(ws.id);
            }
          }}
          onkeydown={(e) => {
            if (e.key === 'Enter' && !isMissing) handleToggleWorkspace(ws.id);
          }}
          class="flex items-center justify-between px-2.5 py-1.5 {isMissing ? 'bg-rose-500/10 cursor-not-allowed opacity-90' : 'bg-ant-bg-secondary hover:bg-ant-bg-tertiary cursor-pointer'} transition select-none group border-b {isExpanded ? 'border-ant-border-secondary dark:border-white/5' : 'border-transparent'}"
          title={isMissing ? `Folder ini sudah tidak ada lagi di disk: ${ws.path}` : ws.path}
        >
          <div class="flex items-center space-x-1.5 min-w-0 flex-1">
            <span class="{isMissing ? 'text-rose-400' : 'text-ant-text-muted'} transition-transform">
              {#if isExpanded && !isMissing}
                <ChevronDown size={13} />
              {:else}
                <ChevronRight size={13} />
              {/if}
            </span>
            <Folder size={14} class="{isMissing ? 'text-rose-400' : 'text-ant-primary'} flex-shrink-0" />
            <span class="text-xs font-semibold {isMissing ? 'text-rose-300 line-through' : 'text-ant-text'} truncate flex-1 min-w-0" title={ws.path}>{ws.name}</span>
            {#if isMissing}
              <span class="text-[9px] bg-rose-500/20 text-rose-300 px-1 py-0.2 rounded font-medium ml-1">Missing</span>
            {:else}
              <span class="text-[10px] text-ant-text-muted font-mono">({totalCount})</span>
            {/if}
          </div>

          <!-- Workspace Actions -->
          <div class="flex items-center space-x-0.5 opacity-80 group-hover:opacity-100">
            {#if isMissing}
              <button
                type="button"
                onclick={(e) => handleRemoveWorkspace(ws, e)}
                class="p-1 rounded text-rose-400 hover:bg-rose-500/20 transition flex items-center"
                title="Hapus workspace yang hilang dari daftar"
              >
                <Trash2 size={12} />
              </button>
            {:else}
              <button
                type="button"
                onclick={(e) => { e.stopPropagation(); syncGrokSessionsForWorkspace(ws.id, ws.path); showToast('Memperbarui session Grok...'); }}
                class="p-1 rounded text-ant-text-muted hover:text-ant-text hover:bg-ant-bg transition"
                title="Refresh / Sync Grok Sessions"
              >
                <RefreshCw size={11} />
              </button>
              <button
                type="button"
                onclick={(e) => { e.stopPropagation(); handleCreateNewSession(ws.id); }}
                class="p-1 rounded text-ant-primary hover:bg-ant-primary/15 transition flex items-center"
                title="Buat Sesi Baru di Folder ini"
              >
                <Plus size={13} />
              </button>
              <button
                type="button"
                onclick={(e) => handleRemoveWorkspace(ws, e)}
                class="p-1 rounded text-ant-text-muted hover:text-ant-error hover:bg-ant-error/15 transition flex items-center"
                title="Hapus Folder Workspace dari Daftar"
              >
                <Trash2 size={12} />
              </button>
            {/if}
          </div>
        </div>

        <!-- Collapsible Sessions under this Workspace -->
        {#if isExpanded && !isMissing}
          <div class="px-0.5 py-0.5 space-y-0.5">
            <!-- Recent Sessions List -->
            {#if visible.length > 0}
              <div class="space-y-0.5">
                {#each visible as session (session.id)}
                  {@const isActive = sessionStore.activeSessionId === session.id}
                  {@const isChecked = sessionStore.selectedSessionIds.has(session.id)}
                  {@const meta = STATUS_META[session.status]}
                  {@const isEditing = editingSessionId === session.id}

                  <div
                    role="button"
                    tabindex="0"
                    onclick={() => handleSelectSession(session.id)}
                    onkeydown={(e) => e.key === 'Enter' && handleSelectSession(session.id)}
                    class="group relative flex items-center justify-between px-2.5 py-1.5 rounded-md text-[13px] font-sans cursor-pointer transition-colors duration-150 {isActive
                      ? 'bg-ant-primary/15 text-ant-primary font-medium'
                      : 'bg-transparent text-ant-text/80 hover:text-ant-text hover:bg-ant-bg-tertiary/70'}"
                  >
                    <!-- Left: Checkbox/Icon Status & Title -->
                    <div class="flex items-center space-x-2.5 min-w-0 flex-1">
                      {#if sessionStore.isSelectionMode}
                        <button
                          type="button"
                          onclick={(e) => { e.stopPropagation(); sessionStore.toggleSessionSelected(session.id); }}
                          class="flex-shrink-0"
                        >
                          <CustomCheckbox checked={isChecked} size="sm" />
                        </button>
                      {:else if session.status === 'working'}
                        <span class="flex items-center justify-center flex-shrink-0 text-ant-primary" title="Status: Working / Generating...">
                          <Loader2 size={13.5} class="animate-spin" />
                        </span>
                      {:else if session.status === 'waiting_permission'}
                        <span class="flex items-center justify-center flex-shrink-0 text-amber-400" title="Status: Waiting Permission">
                          <AlertCircle size={13.5} class="animate-bounce" />
                        </span>
                      {:else if session.status === 'finished'}
                        <span class="flex items-center justify-center flex-shrink-0 text-ant-success/70 group-hover:text-ant-success transition" title="Status: Completed">
                          <CheckCircle2 size={13.5} />
                        </span>
                      {:else}
                        <!-- Idle / Normal Session Icon -->
                        <span class="flex items-center justify-center flex-shrink-0 {isActive ? 'text-ant-primary' : 'text-ant-text-muted/70 group-hover:text-ant-text-secondary'} transition" title="Chat Session">
                          <MessageSquare size={13.5} />
                        </span>
                      {/if}

                      {#if isEditing}
                        <input
                          type="text"
                          bind:value={editTitleText}
                          onkeydown={(e) => {
                            if (e.key === 'Enter') handleSaveRename(session);
                            if (e.key === 'Escape') editingSessionId = null;
                          }}
                          onblur={() => handleSaveRename(session)}
                          class="w-full bg-ant-bg border border-ant-primary rounded px-1.5 py-0.5 text-xs text-ant-text focus:outline-none"
                        />
                      {:else}
                        <span class="truncate leading-normal tracking-tight {isActive ? 'font-semibold text-ant-primary' : 'text-ant-text'}" title={session.title}>{session.title}</span>
                      {/if}
                    </div>

                    <!-- Right Action Menu Trigger -->
                    <div class="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        type="button"
                        onclick={(e) => toggleDropdown(session.id, e)}
                        class="session-more-btn p-0.5 text-ant-text-secondary hover:text-ant-text rounded hover:bg-ant-bg-tertiary transition"
                        title="Session Options"
                      >
                        <MoreVertical size={13.5} />
                      </button>
                    </div>
                  </div>
                {/each}
              </div>

              <!-- Show More / Show Less Toggle Button -->
              {#if hasMore || isExpanded}
                <div class="pt-0.5 pb-1 px-1 flex items-center space-x-1">
                  {#if hasMore}
                    <button
                      type="button"
                      onclick={(e) => { e.stopPropagation(); handleShowMore(ws.id); }}
                      class="flex-1 flex items-center justify-center space-x-1 py-1 px-2 rounded text-[11px] font-medium text-ant-primary hover:bg-ant-primary/10 transition-colors"
                    >
                      <span>Show More (+{Math.min(PAGE_SIZE, hiddenCount)})</span>
                      <span class="text-xs">▾</span>
                    </button>
                  {/if}
                  {#if isExpanded}
                    <button
                      type="button"
                      onclick={(e) => { e.stopPropagation(); handleShowLess(ws.id); }}
                      class="{hasMore ? 'px-2' : 'w-full'} flex items-center justify-center space-x-1 py-1 px-2 rounded text-[11px] font-medium text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-tertiary transition-colors"
                    >
                      <span>Show Less</span>
                      <span class="text-xs">▴</span>
                    </button>
                  {/if}
                </div>
              {/if}
            {:else}
              <div class="px-3 py-2 text-[11px] text-ant-text-muted italic text-center">
                No sessions in this folder yet
              </div>
            {/if}
          </div>
        {/if}
      </div>
    {/each}
    {/if}
  </div>

  <!-- Global Fixed Dropdown Menu (Portal) to prevent parent overflow clipping -->
  {#if activeDropdownId && activeDropdownCoords}
    {@const activeSession = sessionStore.sessions.find(s => s.id === activeDropdownId)}
    {#if activeSession}
      <div
        role="menu"
        tabindex="-1"
        onclick={(e) => e.stopPropagation()}
        onkeydown={(e) => { if (e.key === 'Escape') activeDropdownId = null; }}
        style="top: {activeDropdownCoords.top}px; right: {activeDropdownCoords.right}px;"
        class="session-dropdown-menu fixed w-48 bg-ant-bg border border-ant-border rounded-lg shadow-2xl z-[9999] py-1 text-xs select-none font-serif animate-fade-in"
      >
        <button
          onclick={(e) => handleTogglePin(activeSession, e)}
          class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-secondary text-amber-500 transition text-left"
        >
          <Pin size={12.5} />
          <span>{activeSession.isPinned ? 'Unpin Conversation' : 'Pin Conversation'}</span>
        </button>
        <button
          onclick={(e) => handleStartRename(activeSession, e)}
          class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-secondary text-ant-text transition text-left"
        >
          <Edit2 size={12.5} />
          <span>Rename</span>
        </button>
        <button
          onclick={(e) => handleExportMarkdown(activeSession, e)}
          class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-secondary text-ant-text transition text-left"
        >
          <FileText size={12.5} />
          <span>Export as .md</span>
        </button>
        <button
          onclick={(e) => handleCopySessionId(activeSession, e)}
          class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-secondary text-ant-text transition text-left"
        >
          <Copy size={12.5} />
          <span>Copy Session ID</span>
        </button>
        <button
          onclick={(e) => handleCopySessionName(activeSession, e)}
          class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-secondary text-ant-text transition text-left"
        >
          <Tag size={12.5} />
          <span>Copy Session Title</span>
        </button>
        <div class="border-t border-ant-border/60 my-1"></div>
        <button
          onclick={(e) => handleDeleteSession(activeSession, e)}
          class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-error/15 text-ant-error transition text-left"
        >
          <Trash2 size={12.5} />
          <span>Delete Session</span>
        </button>
      </div>
    {/if}
  {/if}

  <!-- 3. Multi-Select Batch Actions Footer Bar -->
  <BatchActionBar {searchQuery} />

  <!-- Toast Notification -->
  {#if copiedToast}
    <div class="absolute bottom-14 left-1/2 -translate-x-1/2 bg-ant-primary text-white text-[11px] font-medium px-3 py-1 rounded-full shadow-lg z-50 animate-fade-in whitespace-nowrap">
      {copiedToast}
    </div>
  {/if}
</div>
