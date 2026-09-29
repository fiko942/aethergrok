<script lang="ts">
  import { sessionStore, type WorkspaceFolder } from '$lib/stores/session.svelte';
  import { Folder, FolderPlus, Check, X, Search, Terminal } from 'lucide-svelte';
  import { onMount, tick } from 'svelte';

  interface Props {
    visible: boolean;
    onClose: () => void;
  }

  let { visible = $bindable(false), onClose }: Props = $props();

  let searchQuery = $state('');
  let selectedIndex = $state(0);
  let searchInputRef = $state<HTMLInputElement | null>(null);

  // Filtered workspaces based on search query
  let filteredWorkspaces = $derived(
    sessionStore.workspaces.filter((ws) =>
      ws.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      ws.path.toLowerCase().includes(searchQuery.toLowerCase())
    )
  );

  // Focus input when modal opens & reset state
  $effect(() => {
    if (visible) {
      searchQuery = '';
      // Default selectedIndex to the active workspace index if possible
      const activeIdx = sessionStore.workspaces.findIndex((w) => w.id === sessionStore.activeWorkspaceId);
      selectedIndex = activeIdx >= 0 ? activeIdx : 0;
      tick().then(() => {
        searchInputRef?.focus();
      });
    }
  });

  function selectWorkspace(ws: WorkspaceFolder) {
    sessionStore.activeWorkspaceId = ws.id;
    sessionStore.createSession(undefined, ws.id);
    onClose();
  }

  async function handleOpenNewFolder() {
    onClose();
    if (window.go?.main?.App?.SelectWorkspaceDirectory) {
      try {
        const dir = await window.go.main.App.SelectWorkspaceDirectory();
        if (dir) {
          const folderName = dir.split(/[/\\]/).filter(Boolean).pop() || 'workspace';
          const newWs = sessionStore.addWorkspace(folderName, dir);
          // Sync any existing grok sessions on disk
          if (window.go?.main?.App?.DiscoverGrokSessions) {
            const diskSessions = await window.go.main.App.DiscoverGrokSessions(dir);
            if (diskSessions && diskSessions.length > 0) {
              sessionStore.syncDiscoveredGrokSessions(newWs.id, diskSessions);
            }
          }
          // Automatically create a new session in this newly added folder
          sessionStore.createSession(undefined, newWs.id);
        }
      } catch (err) {
        console.error('Failed to open workspace directory:', err);
      }
    } else {
      const defaultPath = typeof navigator !== 'undefined' && /Win/.test(navigator.platform || navigator.userAgent) ? 'C:\\workspace' : '/workspace';
      const path = window.prompt('Enter absolute path of folder workspace:', defaultPath);
      if (path && path.trim()) {
        const folderName = path.trim().split(/[/\\]/).filter(Boolean).pop() || 'workspace';
        const newWs = sessionStore.addWorkspace(folderName, path.trim());
        sessionStore.createSession(undefined, newWs.id);
      }
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (!visible) return;

    if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
      return;
    }

    if (e.key === 'ArrowDown') {
      e.preventDefault();
      const totalItems = filteredWorkspaces.length + 1; // +1 for "Open New Folder"
      selectedIndex = (selectedIndex + 1) % totalItems;
      return;
    }

    if (e.key === 'ArrowUp') {
      e.preventDefault();
      const totalItems = filteredWorkspaces.length + 1;
      selectedIndex = (selectedIndex - 1 + totalItems) % totalItems;
      return;
    }

    if (e.key === 'Enter') {
      e.preventDefault();
      if (selectedIndex < filteredWorkspaces.length) {
        const targetWs = filteredWorkspaces[selectedIndex];
        if (targetWs) selectWorkspace(targetWs);
      } else {
        // "Open New Folder" action
        handleOpenNewFolder();
      }
    }
  }
</script>

<svelte:window onkeydown={handleKeyDown} />

{#if visible}
  <!-- Backdrop Overlay with Smooth Blur & Entrance -->
  <div
    role="presentation"
    class="fixed inset-0 bg-black/60 backdrop-blur-sm z-50 flex items-center justify-center p-4 animate-in fade-in duration-150"
    onclick={onClose}
  >
    <!-- Modal Card (Spotlight / Ant Design Dark Command Palette) -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
      role="dialog"
      aria-modal="true"
      aria-label="Select Workspace for New Conversation"
      class="w-full max-w-lg bg-ant-bg-secondary border border-ant-border-secondary dark:border-white/5 rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150 outline-none flex flex-col max-h-[80vh]"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Search & Title Bar -->
      <div class="px-4 py-3.5 border-b border-white/5 flex items-center space-x-3 bg-ant-bg/60">
        <Search size={16} class="text-ant-primary shrink-0" />
        <input
          bind:this={searchInputRef}
          bind:value={searchQuery}
          type="text"
          placeholder="Choose a workspace folder for new conversation..."
          class="w-full bg-transparent border-none text-xs font-serif text-ant-text placeholder:text-ant-text-muted outline-none"
        />
        <button
          type="button"
          onclick={onClose}
          class="p-1 text-ant-text-muted hover:text-ant-text rounded-md hover:bg-white/5 transition"
          aria-label="Close modal"
        >
          <X size={15} />
        </button>
      </div>

      <!-- Workspace List -->
      <div class="p-2 space-y-1 overflow-y-auto max-h-[320px]">
        {#if filteredWorkspaces.length === 0}
          <div class="py-8 text-center space-y-2">
            <p class="text-xs font-serif text-ant-text-secondary">No matching workspace folders found</p>
            <p class="text-[11px] font-mono text-ant-text-muted">Press Enter to choose a new folder from disk</p>
          </div>
        {:else}
          {#each filteredWorkspaces as ws, idx (ws.id)}
            {@const isSelected = selectedIndex === idx}
            {@const isActive = sessionStore.activeWorkspaceId === ws.id}
            {@const sessionCount = sessionStore.sessions.filter((s) => s.workspaceId === ws.id).length}

            <button
              type="button"
              onclick={() => selectWorkspace(ws)}
              onmouseenter={() => (selectedIndex = idx)}
              class="w-full flex items-center justify-between px-3 py-2.5 rounded-xl text-left transition {
                isSelected
                  ? 'bg-blue-500/15 border border-blue-500/30 text-ant-text shadow-sm'
                  : 'hover:bg-white/5 border border-transparent text-ant-text-secondary hover:text-ant-text'
              }"
            >
              <div class="flex items-center space-x-3 min-w-0">
                <div class="w-8 h-8 rounded-lg flex items-center justify-center shrink-0 {
                  isActive
                    ? 'bg-blue-500/20 text-blue-400'
                    : 'bg-white/5 text-ant-text-secondary'
                }">
                  <Folder size={15} />
                </div>
                <div class="min-w-0">
                  <div class="flex items-center space-x-2">
                    <span class="text-xs font-serif font-semibold {isSelected || isActive ? 'text-ant-text' : 'text-ant-text'}">{ws.name}</span>
                    {#if isActive}
                      <span class="px-1.5 py-0.2 rounded bg-blue-500/20 text-blue-400 text-[9.5px] font-mono font-medium">Active</span>
                    {/if}
                  </div>
                  <div class="truncate text-[10.5px] font-mono text-ant-text-muted mt-0.5">{ws.path}</div>
                </div>
              </div>

              <div class="flex items-center space-x-2 shrink-0 ml-3">
                <span class="text-[11px] font-mono px-2 py-0.5 rounded-md bg-ant-bg border border-white/5 text-ant-text-muted">
                  {sessionCount} {sessionCount === 1 ? 'chat' : 'chats'}
                </span>
                {#if isSelected}
                  <kbd class="px-1.5 py-0.5 text-[9px] font-mono bg-blue-500/30 text-blue-300 rounded border border-blue-500/40">↵ Enter</kbd>
                {/if}
              </div>
            </button>
          {/each}
        {/if}

        <!-- Add New Workspace Folder Item -->
        {#snippet addFolderItem()}
          {@const isAddSelected = selectedIndex === filteredWorkspaces.length}
          <button
            type="button"
            onclick={handleOpenNewFolder}
            onmouseenter={() => (selectedIndex = filteredWorkspaces.length)}
            class="w-full flex items-center justify-between px-3 py-2.5 rounded-xl text-left transition {
              isAddSelected
                ? 'bg-ant-primary/15 border border-ant-primary/30 text-ant-text shadow-sm'
                : 'hover:bg-white/5 border border-transparent text-ant-text-secondary hover:text-ant-text'
            }"
          >
            <div class="flex items-center space-x-3 min-w-0">
              <div class="w-8 h-8 rounded-lg bg-ant-primary/10 text-ant-primary flex items-center justify-center shrink-0">
                <FolderPlus size={16} />
              </div>
              <div>
                <div class="text-xs font-serif font-semibold text-ant-text">Open New Folder...</div>
                <div class="text-[10.5px] font-mono text-ant-text-muted mt-0.5">Pick another repository or folder from your disk</div>
              </div>
            </div>

            {#if isAddSelected}
              <kbd class="px-1.5 py-0.5 text-[9px] font-mono bg-ant-primary/30 text-ant-primary rounded border border-ant-primary/40">↵ Enter</kbd>
            {/if}
          </button>
        {/snippet}

        {@render addFolderItem()}
      </div>

      <!-- Footer Quick Tips -->
      <div class="px-4 py-2.5 bg-ant-bg/80 border-t border-white/5 flex items-center justify-between text-[10.5px] font-mono text-ant-text-muted">
        <div class="flex items-center space-x-3">
          <span><kbd class="px-1 py-0.5 bg-white/5 rounded">↑</kbd> <kbd class="px-1 py-0.5 bg-white/5 rounded">↓</kbd> Navigate</span>
          <span><kbd class="px-1 py-0.5 bg-white/5 rounded">↵</kbd> Select</span>
          <span><kbd class="px-1 py-0.5 bg-white/5 rounded">esc</kbd> Dismiss</span>
        </div>
        <div class="text-ant-text-muted/70 font-serif text-[10px]">
          AetherGrok Workspace Isolation
        </div>
      </div>
    </div>
  </div>
{/if}
