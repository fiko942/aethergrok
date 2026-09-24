<script lang="ts">
  import { onMount } from 'svelte';
  import { sessionStore, STATUS_META, type Session, type WorkspaceFolder } from '$lib/stores/session.svelte';
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
    CheckSquare,
    Square,
    ChevronDown,
    ChevronRight,
    FolderOpen,
    Edit2,
    Pin,
    Search,
    RefreshCw
  } from 'lucide-svelte';

  let activeDropdownId = $state<string | null>(null);
  let editingSessionId = $state<string | null>(null);
  let editTitleText = $state('');
  let copiedToast = $state<string | null>(null);
  let searchQuery = $state('');

  function showToast(text: string) {
    copiedToast = text;
    setTimeout(() => {
      copiedToast = null;
    }, 2200);
  }

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

  // Auto-sync sessions on mount for all known workspaces
  onMount(() => {
    for (const ws of sessionStore.workspaces) {
      syncGrokSessionsForWorkspace(ws.id, ws.path);
    }
  });

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
      const path = window.prompt('Masukkan absolute path folder workspace:', '/Users/fiko942/Desktop/affilia');
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
    showToast(session.isPinned ? 'Pin dilepas' : 'Percakapan disematkan (Pinned)!');
  }

  async function handleCopySessionId(session: Session, e: MouseEvent) {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(session.id);
      showToast('Session ID copied!');
    } catch {
      showToast(session.id);
    }
    activeDropdownId = null;
  }

  async function handleCopySessionName(session: Session, e: MouseEvent) {
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(session.title);
      showToast('Session Name copied!');
    } catch {
      showToast(session.title);
    }
    activeDropdownId = null;
  }

  async function handleExportMarkdown(session: Session, e: MouseEvent) {
    e.stopPropagation();
    activeDropdownId = null;
    try {
      const ws = sessionStore.workspaces.find(w => w.id === session.workspaceId) || sessionStore.activeWorkspace;
      const mdContent = formatSessionAsMarkdown(session, ws);
      const safeTitle = session.title.toLowerCase().replace(/[^a-z0-9_-]/g, '-').slice(0, 30);
      const dateStr = new Date().toISOString().slice(0, 10);
      const filename = `${safeTitle}-${dateStr}.md`;

      const res = await downloadOrSaveMarkdown(filename, mdContent);
      if (res.success) {
        showToast('Berhasil diekspor ke Markdown (.md)!');
      } else if (res.error) {
        alert('Gagal mengekspor: ' + res.error);
      }
    } catch (err) {
      alert('Error ekspor Markdown: ' + String(err));
    }
  }

  function handleDeleteSession(session: Session, e: MouseEvent) {
    e.stopPropagation();
    activeDropdownId = null;
    const ws = sessionStore.workspaces.find(w => w.id === session.workspaceId);
    if (confirm(`Hapus sesi "${session.title}"?`)) {
      sessionStore.closeSession(session.id);
      if (ws && window.go?.main?.App?.DeleteGrokSession) {
        window.go.main.App.DeleteGrokSession(ws.path, session.id).catch(console.error);
      }
      showToast('Sesi berhasil dihapus');
    }
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

  function getWorkspaceSessions(wsId: string): { pinned: Session[]; recent: Session[] } {
    const q = searchQuery.toLowerCase().trim();
    let wsSessions = sessionStore.sessions.filter(s => s.workspaceId === wsId);
    if (q) {
      wsSessions = wsSessions.filter(s => s.title.toLowerCase().includes(q));
    }

    const pinned = wsSessions.filter(s => s.isPinned).sort((a, b) => (b.pinnedAt || 0) - (a.pinnedAt || 0));
    const recent = wsSessions.filter(s => !s.isPinned).sort((a, b) => b.updatedAt - a.updatedAt);
    return { pinned, recent };
  }
</script>

<aside class="flex flex-col w-64 h-full bg-ant-bg-secondary border-r border-ant-border select-none relative z-20">
  <!-- 1. Search Bar & Action Controls -->
  <div class="p-2.5 border-b border-ant-border space-y-2">
    <!-- Search Bar -->
    <div class="relative flex items-center">
      <Search size={13} class="absolute left-2.5 text-ant-text-muted" />
      <input
        type="text"
        bind:value={searchQuery}
        placeholder="Filter project & session..."
        class="w-full bg-ant-bg border border-ant-border rounded-md pl-8 pr-2.5 py-1 text-xs text-white placeholder-ant-text-muted focus:border-ant-primary focus:outline-none transition-colors"
      />
    </div>

    <!-- Multi-select / Batch Action Toggle Bar -->
    <div class="flex items-center justify-between text-xs px-0.5">
      <button
        onclick={handleOpenWorkspaceFolder}
        class="flex items-center space-x-1.5 text-xs text-ant-primary hover:text-ant-primary-hover font-medium transition"
        title="Buka Folder Workspace Baru"
      >
        <FolderPlus size={14} />
        <span>Buka Folder</span>
      </button>

      <button
        onclick={() => sessionStore.toggleSelectionMode()}
        class="text-[11px] font-medium px-2 py-0.5 rounded transition {sessionStore.isSelectionMode ? 'bg-ant-primary text-white' : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary'}"
        title="Mode Multi-select"
      >
        {sessionStore.isSelectionMode ? 'Batal' : 'Tandai'}
      </button>
    </div>
  </div>

  <!-- 2. Hierarchical Multi-Workspace Tree Explorer -->
  <div class="flex-1 overflow-y-auto p-2 space-y-2.5 custom-scrollbar">
    {#each filteredWorkspaces as ws (ws.id)}
      {@const isExpanded = ws.isExpanded !== false}
      {@const { pinned, recent } = getWorkspaceSessions(ws.id)}
      {@const totalCount = pinned.length + recent.length}

      <div class="rounded-lg bg-ant-bg/60 border border-ant-border-secondary/60 overflow-hidden">
        <!-- Workspace Folder Header -->
        <div
          role="button"
          tabindex="0"
          onclick={() => sessionStore.toggleWorkspaceExpanded(ws.id)}
          onkeydown={(e) => e.key === 'Enter' && sessionStore.toggleWorkspaceExpanded(ws.id)}
          class="flex items-center justify-between px-2.5 py-1.5 bg-ant-bg-tertiary/60 hover:bg-ant-bg-tertiary cursor-pointer transition select-none group"
        >
          <div class="flex items-center space-x-1.5 min-w-0 flex-1">
            <span class="text-ant-text-muted transition-transform">
              {#if isExpanded}
                <ChevronDown size={13} />
              {:else}
                <ChevronRight size={13} />
              {/if}
            </span>
            <Folder size={14} class="text-ant-primary flex-shrink-0" />
            <span class="text-xs font-semibold text-white truncate max-w-[110px]" title={ws.path}>{ws.name}</span>
            <span class="text-[10px] text-ant-text-muted font-mono">({totalCount})</span>
          </div>

          <!-- Workspace Actions -->
          <div class="flex items-center space-x-1 opacity-80 group-hover:opacity-100">
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
          </div>
        </div>

        <!-- Collapsible Sessions under this Workspace -->
        {#if isExpanded}
          <div class="p-1.5 space-y-1">
            <!-- Pinned Sessions Group -->
            {#if pinned.length > 0}
              <div class="space-y-1 pb-1">
                <div class="px-1.5 pt-0.5 text-[10px] font-semibold tracking-wider text-amber-400 uppercase flex items-center gap-1">
                  <Pin size={10} class="fill-current" />
                  <span>Pinned ({pinned.length})</span>
                </div>
                {#each pinned as session (session.id)}
                  {@const isActive = sessionStore.activeSessionId === session.id}
                  {@const isChecked = sessionStore.selectedSessionIds.has(session.id)}
                  {@const meta = STATUS_META[session.status]}
                  {@const isEditing = editingSessionId === session.id}

                  <div
                    role="button"
                    tabindex="0"
                    onclick={() => {
                      if (sessionStore.isSelectionMode) {
                        sessionStore.toggleSessionSelected(session.id);
                      } else {
                        sessionStore.switchSession(session.id);
                      }
                    }}
                    onkeydown={(e) => e.key === 'Enter' && sessionStore.switchSession(session.id)}
                    class="group relative flex items-center justify-between px-2 py-1.5 rounded-md text-xs font-medium cursor-pointer transition border {isActive
                      ? 'bg-ant-primary/15 text-ant-primary border-ant-primary/40 shadow-sm'
                      : 'bg-ant-bg-secondary/70 text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border-transparent'}"
                  >
                    <!-- Left: Checkbox/Dot & Title -->
                    <div class="flex items-center space-x-1.5 min-w-0 flex-1">
                      {#if sessionStore.isSelectionMode}
                        <button
                          type="button"
                          onclick={(e) => { e.stopPropagation(); sessionStore.toggleSessionSelected(session.id); }}
                          class="flex-shrink-0 text-ant-primary"
                        >
                          {#if isChecked}
                            <CheckSquare size={13} class="text-ant-primary" />
                          {:else}
                            <Square size={13} class="text-ant-text-muted" />
                          {/if}
                        </button>
                      {:else}
                        <Pin size={11} class="text-amber-400 fill-current flex-shrink-0" />
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
                          class="w-full bg-ant-bg border border-ant-primary rounded px-1.5 py-0.5 text-xs text-white focus:outline-none"
                        />
                      {:else}
                        <span class="truncate text-[11px] {isActive ? 'font-semibold text-white' : ''}">{session.title}</span>
                      {/if}
                    </div>

                    <!-- Right Action Menu Trigger -->
                    <div class="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        type="button"
                        onclick={(e) => {
                          e.stopPropagation();
                          activeDropdownId = activeDropdownId === session.id ? null : session.id;
                        }}
                        class="p-0.5 text-ant-text-muted hover:text-white rounded hover:bg-ant-bg"
                        title="Opsi Sesi"
                      >
                        <MoreVertical size={13} />
                      </button>
                    </div>

                    <!-- Dropdown Menu -->
                    {#if activeDropdownId === session.id}
                      <div
                        role="menu"
                        tabindex="-1"
                        onclick={(e) => e.stopPropagation()}
                        onkeydown={(e) => { if (e.key === 'Escape') activeDropdownId = null; }}
                        class="absolute right-2 top-8 w-44 bg-ant-bg border border-ant-border rounded-lg shadow-2xl z-50 py-1 text-xs select-none"
                      >
                        <button
                          onclick={(e) => handleTogglePin(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-amber-300 transition text-left"
                        >
                          <Pin size={12} />
                          <span>Unpin Conversation</span>
                        </button>
                        <button
                          onclick={(e) => handleStartRename(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-ant-text transition text-left"
                        >
                          <Edit2 size={12} />
                          <span>Ganti Nama (Rename)</span>
                        </button>
                        <button
                          onclick={(e) => handleExportMarkdown(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-ant-text transition text-left"
                        >
                          <FileText size={12} />
                          <span>Export as .md</span>
                        </button>
                        <button
                          onclick={(e) => handleCopySessionId(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-ant-text transition text-left"
                        >
                          <Copy size={12} />
                          <span>Salin Session ID</span>
                        </button>
                        <button
                          onclick={(e) => handleCopySessionName(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-ant-text transition text-left"
                        >
                          <Tag size={12} />
                          <span>Salin Nama Sesi</span>
                        </button>
                        <div class="border-t border-ant-border my-1"></div>
                        <button
                          onclick={(e) => handleDeleteSession(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-error/10 text-ant-error transition text-left"
                        >
                          <Trash2 size={12} />
                          <span>Hapus Sesi</span>
                        </button>
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
            {/if}

            <!-- Recent Sessions List -->
            {#if recent.length > 0}
              <div class="space-y-1">
                {#each recent as session (session.id)}
                  {@const isActive = sessionStore.activeSessionId === session.id}
                  {@const isChecked = sessionStore.selectedSessionIds.has(session.id)}
                  {@const meta = STATUS_META[session.status]}
                  {@const isEditing = editingSessionId === session.id}

                  <div
                    role="button"
                    tabindex="0"
                    onclick={() => {
                      if (sessionStore.isSelectionMode) {
                        sessionStore.toggleSessionSelected(session.id);
                      } else {
                        sessionStore.switchSession(session.id);
                      }
                    }}
                    onkeydown={(e) => e.key === 'Enter' && sessionStore.switchSession(session.id)}
                    class="group relative flex items-center justify-between px-2 py-1.5 rounded-md text-xs font-medium cursor-pointer transition border {isActive
                      ? 'bg-ant-primary/15 text-ant-primary border-ant-primary/40 shadow-sm'
                      : 'bg-ant-bg-secondary/40 text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border-transparent'}"
                  >
                    <!-- Left: Checkbox/Dot & Title -->
                    <div class="flex items-center space-x-1.5 min-w-0 flex-1">
                      {#if sessionStore.isSelectionMode}
                        <button
                          type="button"
                          onclick={(e) => { e.stopPropagation(); sessionStore.toggleSessionSelected(session.id); }}
                          class="flex-shrink-0 text-ant-primary"
                        >
                          {#if isChecked}
                            <CheckSquare size={13} class="text-ant-primary" />
                          {:else}
                            <Square size={13} class="text-ant-text-muted" />
                          {/if}
                        </button>
                      {:else}
                        <span
                          class="w-1.5 h-1.5 rounded-full flex-shrink-0 {meta.dotClass}"
                          title="Status: {meta.label}"
                        ></span>
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
                          class="w-full bg-ant-bg border border-ant-primary rounded px-1.5 py-0.5 text-xs text-white focus:outline-none"
                        />
                      {:else}
                        <span class="truncate text-[11px] {isActive ? 'font-semibold text-white' : ''}">{session.title}</span>
                      {/if}
                    </div>

                    <!-- Right Action Menu Trigger -->
                    <div class="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        type="button"
                        onclick={(e) => {
                          e.stopPropagation();
                          activeDropdownId = activeDropdownId === session.id ? null : session.id;
                        }}
                        class="p-0.5 text-ant-text-muted hover:text-white rounded hover:bg-ant-bg"
                        title="Opsi Sesi"
                      >
                        <MoreVertical size={13} />
                      </button>
                    </div>

                    <!-- Dropdown Menu -->
                    {#if activeDropdownId === session.id}
                      <div
                        role="menu"
                        tabindex="-1"
                        onclick={(e) => e.stopPropagation()}
                        onkeydown={(e) => { if (e.key === 'Escape') activeDropdownId = null; }}
                        class="absolute right-2 top-8 w-44 bg-ant-bg border border-ant-border rounded-lg shadow-2xl z-50 py-1 text-xs select-none"
                      >
                        <button
                          onclick={(e) => handleTogglePin(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-amber-300 transition text-left"
                        >
                          <Pin size={12} />
                          <span>Pin Conversation</span>
                        </button>
                        <button
                          onclick={(e) => handleStartRename(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-ant-text transition text-left"
                        >
                          <Edit2 size={12} />
                          <span>Ganti Nama (Rename)</span>
                        </button>
                        <button
                          onclick={(e) => handleExportMarkdown(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-ant-text transition text-left"
                        >
                          <FileText size={12} />
                          <span>Export as .md</span>
                        </button>
                        <button
                          onclick={(e) => handleCopySessionId(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-ant-text transition text-left"
                        >
                          <Copy size={12} />
                          <span>Salin Session ID</span>
                        </button>
                        <button
                          onclick={(e) => handleCopySessionName(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-bg-tertiary text-ant-text transition text-left"
                        >
                          <Tag size={12} />
                          <span>Salin Nama Sesi</span>
                        </button>
                        <div class="border-t border-ant-border my-1"></div>
                        <button
                          onclick={(e) => handleDeleteSession(session, e)}
                          class="w-full flex items-center space-x-2 px-2.5 py-1.5 hover:bg-ant-error/10 text-ant-error transition text-left"
                        >
                          <Trash2 size={12} />
                          <span>Hapus Sesi</span>
                        </button>
                      </div>
                    {/if}
                  </div>
                {/each}
              </div>
            {:else if pinned.length === 0}
              <div class="px-3 py-2 text-[11px] text-ant-text-muted italic text-center">
                Belum ada sesi di folder ini
              </div>
            {/if}
          </div>
        {/if}
      </div>
    {/each}
  </div>

  <!-- 3. Multi-Select Batch Actions Footer Bar -->
  <BatchActionBar />

  <!-- Toast Notification -->
  {#if copiedToast}
    <div class="absolute bottom-14 left-1/2 -translate-x-1/2 bg-ant-primary text-white text-[11px] font-medium px-3 py-1 rounded-full shadow-lg z-50 animate-fade-in whitespace-nowrap">
      {copiedToast}
    </div>
  {/if}
</aside>
