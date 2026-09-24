<script lang="ts">
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
    FolderOpen,
    Edit2,
    GitFork
  } from 'lucide-svelte';

  let activeDropdownId = $state<string | null>(null);
  let workspaceDropdownOpen = $state(false);
  let editingSessionId = $state<string | null>(null);
  let editTitleText = $state('');
  let copiedToast = $state<string | null>(null);

  function showToast(text: string) {
    copiedToast = text;
    setTimeout(() => {
      copiedToast = null;
    }, 2200);
  }

  // Add workspace folder using native OS folder picker or prompt
  async function handleOpenWorkspaceFolder() {
    workspaceDropdownOpen = false;
    if (window.go?.main?.App?.SelectWorkspaceDirectory) {
      try {
        const dir = await window.go.main.App.SelectWorkspaceDirectory();
        if (dir) {
          const folderName = dir.split(/[/\\]/).filter(Boolean).pop() || 'workspace';
          sessionStore.addWorkspace(folderName, dir);
          showToast(`Workspace "${folderName}" dibuka`);
        }
      } catch (err) {
        console.error('Failed to select directory:', err);
      }
    } else {
      // Browser fallback prompt
      const path = window.prompt('Masukkan absolute path folder workspace:', '/Users/fiko942/Desktop/affilia');
      if (path && path.trim()) {
        const folderName = path.trim().split(/[/\\]/).filter(Boolean).pop() || 'workspace';
        sessionStore.addWorkspace(folderName, path.trim());
        showToast(`Workspace "${folderName}" ditambahkan`);
      }
    }
  }

  function handleCreateNewSession() {
    sessionStore.createSession();
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
      const ws = sessionStore.activeWorkspace;
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
    if (window.confirm(`Hapus session "${session.title}"?`)) {
      sessionStore.closeSession(session.id);
      showToast('Session dihapus');
    }
  }

  function handleForkSession(session: Session, e: MouseEvent) {
    e.stopPropagation();
    activeDropdownId = null;
    sessionStore.forkSession(session.id);
    showToast('Session di-fork');
  }

  // Close menus on click outside
  function handleWindowClick() {
    activeDropdownId = null;
    workspaceDropdownOpen = false;
  }
</script>

<svelte:window onclick={handleWindowClick} />

<div class="flex flex-col h-full space-y-3">
  <!-- 1. Workspace Header & Folder Selector -->
  <div>
    <div class="flex items-center justify-between px-2 mb-1.5">
      <span class="text-[11px] font-bold tracking-wider text-ant-text-muted uppercase">Workspace</span>
      <button
        onclick={(e) => { e.stopPropagation(); handleOpenWorkspaceFolder(); }}
        class="flex items-center space-x-1 text-[11px] text-ant-primary hover:text-ant-primary-hover font-medium px-1.5 py-0.5 rounded hover:bg-ant-primary/10 transition"
        title="Buka / Tambah Folder Workspace Baru"
      >
        <FolderPlus size={12} />
        <span>Folder</span>
      </button>
    </div>

    <!-- Active Workspace Dropdown Trigger -->
    <div class="relative">
      <button
        onclick={(e) => { e.stopPropagation(); workspaceDropdownOpen = !workspaceDropdownOpen; }}
        class="w-full flex items-center justify-between px-2.5 py-2 rounded-lg bg-ant-bg border border-ant-border hover:border-ant-primary/50 transition text-xs font-medium text-ant-text group shadow-sm"
      >
        <div class="flex items-center space-x-2 min-w-0">
          <div class="w-6 h-6 rounded flex items-center justify-center bg-ant-primary/10 text-ant-primary flex-shrink-0">
            <FolderOpen size={14} />
          </div>
          <div class="text-left min-w-0">
            <div class="truncate text-white font-semibold">{sessionStore.activeWorkspace?.name || 'Pilih Workspace'}</div>
            <div class="text-[10px] text-ant-text-muted truncate max-w-[150px]">{sessionStore.activeWorkspace?.path || 'Belum ada folder'}</div>
          </div>
        </div>
        <ChevronDown size={14} class="text-ant-text-muted group-hover:text-ant-text transition-transform {workspaceDropdownOpen ? 'rotate-180' : ''}" />
      </button>

      <!-- Workspace List Popup -->
      {#if workspaceDropdownOpen}
        <div
          role="dialog"
          aria-label="Workspace Dropdown"
          tabindex="-1"
          onclick={(e) => e.stopPropagation()}
          onkeydown={(e) => { if (e.key === 'Escape') workspaceDropdownOpen = false; }}
          class="absolute left-0 right-0 top-full mt-1 bg-ant-bg border border-ant-border rounded-lg shadow-xl shadow-black/50 z-30 py-1 max-h-56 overflow-y-auto"
        >
          <div class="px-2 py-1 text-[10px] uppercase font-semibold text-ant-text-muted tracking-wider">
            Daftar Folder Workspace
          </div>
          {#each sessionStore.workspaces as ws (ws.id)}
            {@const isSelected = sessionStore.activeWorkspaceId === ws.id}
            <button
              onclick={() => { sessionStore.switchWorkspace(ws.id); workspaceDropdownOpen = false; }}
              class="w-full flex items-center justify-between px-2.5 py-1.5 text-xs hover:bg-ant-bg-tertiary transition text-left {isSelected ? 'text-ant-primary font-semibold bg-ant-primary/10' : 'text-ant-text-secondary'}"
            >
              <div class="flex items-center space-x-2 min-w-0">
                <Folder size={13} class={isSelected ? 'text-ant-primary' : 'text-ant-text-muted'} />
                <span class="truncate">{ws.name}</span>
              </div>
              {#if isSelected}
                <Check size={13} class="text-ant-primary" />
              {/if}
            </button>
          {/each}

          <div class="border-t border-ant-border mt-1 pt-1">
            <button
              onclick={handleOpenWorkspaceFolder}
              class="w-full flex items-center space-x-2 px-2.5 py-1.5 text-xs text-ant-primary hover:bg-ant-primary/10 transition"
            >
              <FolderPlus size={13} />
              <span>+ Buka Folder Lain...</span>
            </button>
          </div>
        </div>
      {/if}
    </div>
  </div>

  <!-- 2. Sessions Header & Mode Actions -->
  <div class="flex items-center justify-between px-2 pt-1 border-t border-ant-border-secondary">
    <div class="flex items-center space-x-1.5">
      <span class="text-[11px] font-bold tracking-wider text-ant-text-muted uppercase">Sessions</span>
      <span class="text-[10px] text-ant-text-muted">({sessionStore.activeWorkspaceSessions.length})</span>
    </div>

    <div class="flex items-center space-x-1">
      <button
        onclick={() => sessionStore.toggleSelectionMode()}
        class="text-[11px] font-medium px-2 py-0.5 rounded transition {sessionStore.isSelectionMode ? 'bg-ant-primary text-white' : 'text-ant-text-secondary hover:text-white hover:bg-ant-bg-tertiary'}"
        title="Aktifkan mode multi-select untuk tandai dan hapus/ekspor massal"
      >
        {sessionStore.isSelectionMode ? 'Batal' : 'Tandai'}
      </button>

      <button
        onclick={handleCreateNewSession}
        class="flex items-center space-x-0.5 text-[11px] text-ant-primary hover:text-ant-primary-hover font-semibold px-2 py-0.5 rounded hover:bg-ant-primary/10 transition"
        title="Buat Session Baru (Cmd+T)"
      >
        <Plus size={13} />
        <span>Baru</span>
      </button>
    </div>
  </div>

  <!-- 3. List of Sessions in Active Workspace -->
  <div class="flex-1 overflow-y-auto space-y-1 pr-1 custom-scrollbar">
    {#each sessionStore.activeWorkspaceSessions as session (session.id)}
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
        class="group relative flex items-center justify-between px-2.5 py-2 rounded-lg text-xs font-medium cursor-pointer transition border {isActive
          ? 'bg-ant-primary/15 text-ant-primary border-ant-primary/40 shadow-sm'
          : 'bg-ant-bg/40 text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border-transparent'}"
      >
        <!-- Left: Status or Checkbox & Title -->
        <div class="flex items-center space-x-2 min-w-0 flex-1">
          {#if sessionStore.isSelectionMode}
            <button
              onclick={(e) => {
                e.stopPropagation();
                sessionStore.toggleSessionSelected(session.id);
              }}
              class="flex-shrink-0 text-ant-primary hover:scale-110 transition-transform"
            >
              {#if isChecked}
                <CheckSquare size={15} class="text-ant-primary" />
              {:else}
                <Square size={15} class="text-ant-text-muted" />
              {/if}
            </button>
          {:else}
            <span
              class="w-2 h-2 rounded-full flex-shrink-0 {meta.dotClass}"
              title="Status: {meta.label}"
            ></span>
          {/if}

          {#if isEditing}
            <input
              type="text"
              bind:value={editTitleText}
              onclick={(e) => e.stopPropagation()}
              onblur={() => handleSaveRename(session)}
              onkeydown={(e) => {
                if (e.key === 'Enter') handleSaveRename(session);
                if (e.key === 'Escape') editingSessionId = null;
              }}
              class="bg-ant-bg px-1.5 py-0.5 rounded border border-ant-primary text-xs text-white outline-none w-full"
              autoFocus
            />
          {:else}
            <span class="truncate {isActive ? 'font-semibold text-white' : ''}">
              {session.title}
            </span>
          {/if}
        </div>

        <!-- Right: Actions Menu (•••) visible on hover -->
        {#if !sessionStore.isSelectionMode && !isEditing}
          <div class="relative flex-shrink-0 opacity-0 group-hover:opacity-100 transition-opacity">
            <button
              onclick={(e) => {
                e.stopPropagation();
                activeDropdownId = activeDropdownId === session.id ? null : session.id;
              }}
              class="p-1 rounded hover:bg-ant-bg-tertiary text-ant-text-muted hover:text-white transition"
              title="Menu Opsi Session"
            >
              <MoreVertical size={13} />
            </button>

            <!-- Context Dropdown -->
            {#if activeDropdownId === session.id}
              <div
                role="menu"
                tabindex="-1"
                onclick={(e) => e.stopPropagation()}
                onkeydown={(e) => { if (e.key === 'Escape') activeDropdownId = null; }}
                class="absolute right-0 top-full mt-1 w-44 bg-ant-bg border border-ant-border rounded-lg shadow-2xl z-40 py-1 text-xs"
              >
                <button
                  onclick={(e) => handleExportMarkdown(session, e)}
                  class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-tertiary text-ant-text hover:text-ant-primary transition text-left"
                >
                  <FileText size={13} class="text-ant-primary" />
                  <span>Ekspor Log (.md)</span>
                </button>

                <button
                  onclick={(e) => handleCopySessionId(session, e)}
                  class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-tertiary text-ant-text hover:text-white transition text-left"
                >
                  <Copy size={13} class="text-ant-text-muted" />
                  <span>Salin Session ID</span>
                </button>

                <button
                  onclick={(e) => handleCopySessionName(session, e)}
                  class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-tertiary text-ant-text hover:text-white transition text-left"
                >
                  <Tag size={13} class="text-ant-text-muted" />
                  <span>Salin Nama Session</span>
                </button>

                <button
                  onclick={(e) => handleStartRename(session, e)}
                  class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-tertiary text-ant-text hover:text-white transition text-left"
                >
                  <Edit2 size={13} class="text-ant-text-muted" />
                  <span>Ubah Nama</span>
                </button>

                <button
                  onclick={(e) => handleForkSession(session, e)}
                  class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-bg-tertiary text-ant-text hover:text-white transition text-left"
                >
                  <GitFork size={13} class="text-ant-text-muted" />
                  <span>Duplikat / Fork</span>
                </button>

                <div class="border-t border-ant-border-secondary my-1"></div>

                <button
                  onclick={(e) => handleDeleteSession(session, e)}
                  class="w-full flex items-center space-x-2 px-3 py-1.5 hover:bg-ant-error/15 text-ant-error transition text-left"
                >
                  <Trash2 size={13} />
                  <span>Hapus Session</span>
                </button>
              </div>
            {/if}
          </div>
        {/if}
      </div>
    {/each}

    {#if sessionStore.activeWorkspaceSessions.length === 0}
      <div class="text-center py-6 px-2 text-ant-text-muted text-xs">
        <Folder size={20} class="mx-auto mb-1.5 opacity-40" />
        <p>Belum ada session di workspace ini.</p>
        <button
          onclick={handleCreateNewSession}
          class="mt-2 text-ant-primary hover:underline font-medium"
        >
          + Buat session pertama
        </button>
      </div>
    {/if}
  </div>

  <!-- 4. Floating Batch Action Bar (Visible when Selection Mode Active) -->
  <BatchActionBar />

  <!-- Feedback Toast Notification -->
  {#if copiedToast}
    <div class="fixed bottom-4 left-6 z-50 px-3 py-1.5 rounded-lg bg-ant-bg-elevated border border-ant-primary/40 text-ant-primary text-xs font-semibold shadow-2xl animate-in fade-in slide-in-from-bottom-2 duration-150 flex items-center space-x-1.5">
      <Check size={13} />
      <span>{copiedToast}</span>
    </div>
  {/if}
</div>
