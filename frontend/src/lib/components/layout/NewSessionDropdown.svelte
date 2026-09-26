<script lang="ts">
  import { sessionStore, type WorkspaceFolder } from '$lib/stores/session.svelte';
  import { Folder, FolderPlus, Check, Sparkles } from 'lucide-svelte';
  import { onMount } from 'svelte';

  interface Props {
    open: boolean;
    placement?: 'bottom-start' | 'bottom-end' | 'bottom-center';
    onClose: () => void;
  }

  let { open = $bindable(false), placement = 'bottom-end', onClose }: Props = $props();

  let dropdownRef = $state<HTMLDivElement | null>(null);

  async function handleSelectWorkspace(ws: WorkspaceFolder) {
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
          sessionStore.createSession(undefined, newWs.id);
        }
      } catch (err) {
        console.error('Failed to open workspace directory:', err);
      }
    } else {
      const path = window.prompt('Enter absolute path of folder workspace:', '/Users/fiko942/Desktop/workspace');
      if (path && path.trim()) {
        const folderName = path.trim().split(/[/\\]/).filter(Boolean).pop() || 'workspace';
        const newWs = sessionStore.addWorkspace(folderName, path.trim());
        sessionStore.createSession(undefined, newWs.id);
      }
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onClose();
    }
  }

  function handleClickOutside(e: MouseEvent) {
    if (dropdownRef && !dropdownRef.contains(e.target as Node)) {
      onClose();
    }
  }

  onMount(() => {
    window.addEventListener('mousedown', handleClickOutside);
    window.addEventListener('keydown', handleKeydown);
    return () => {
      window.removeEventListener('mousedown', handleClickOutside);
      window.removeEventListener('keydown', handleKeydown);
    };
  });
</script>

{#if open}
  <div
    bind:this={dropdownRef}
    class="absolute z-50 min-w-[260px] max-w-[320px] bg-[#18181c]/95 border border-white/10 rounded-xl shadow-2xl backdrop-blur-md p-1.5 space-y-1 animate-in fade-in zoom-in-95 duration-150 select-none {
      placement === 'bottom-end' ? 'right-0 top-full mt-1.5' : placement === 'bottom-start' ? 'left-0 top-full mt-1.5' : 'left-1/2 -translate-x-1/2 top-full mt-2'
    }"
  >
    <!-- Dropdown Header -->
    <div class="px-2.5 py-1.5 text-[11px] font-mono uppercase tracking-wider text-ant-text-muted flex items-center justify-between border-b border-white/5">
      <span>Select Workspace</span>
      <span class="text-[10px] text-ant-text-muted/60 lowercase">for new conversation</span>
    </div>

    <!-- Workspace List -->
    <div class="max-h-[220px] overflow-y-auto space-y-0.5 py-1">
      {#each sessionStore.workspaces as ws (ws.id)}
        {@const isActive = sessionStore.activeWorkspaceId === ws.id}
        {@const sessionCount = sessionStore.sessions.filter(s => s.workspaceId === ws.id).length}
        <button
          type="button"
          onclick={() => handleSelectWorkspace(ws)}
          class="w-full flex items-center justify-between px-2.5 py-2 rounded-lg text-xs font-serif text-left transition group {
            isActive
              ? 'bg-blue-500/10 text-blue-400 font-medium'
              : 'text-ant-text hover:bg-white/5'
          }"
        >
          <div class="flex items-center space-x-2.5 min-w-0">
            <div class="w-6 h-6 rounded-md flex items-center justify-center flex-shrink-0 {
              isActive ? 'bg-blue-500/20 text-blue-400' : 'bg-white/5 text-ant-text-secondary group-hover:text-ant-text'
            }">
              <Folder size={13} />
            </div>
            <div class="min-w-0">
              <div class="truncate text-xs font-serif">{ws.name}</div>
              <div class="truncate text-[10px] font-mono text-ant-text-muted">{ws.path}</div>
            </div>
          </div>

          <div class="flex items-center space-x-1.5 flex-shrink-0 ml-2">
            <span class="text-[10.5px] font-mono px-1.5 py-0.2 rounded bg-white/5 text-ant-text-muted">
              {sessionCount}
            </span>
            {#if isActive}
              <Check size={13} class="text-blue-400" />
            {/if}
          </div>
        </button>
      {/each}
    </div>

    <!-- Divider & Open New Workspace Action -->
    <div class="pt-1 border-t border-white/5">
      <button
        type="button"
        onclick={handleOpenNewFolder}
        class="w-full flex items-center space-x-2.5 px-2.5 py-2 rounded-lg text-xs font-serif text-ant-text hover:text-ant-primary hover:bg-ant-primary/10 transition cursor-pointer group"
      >
        <div class="w-6 h-6 rounded-md bg-ant-primary/15 text-ant-primary flex items-center justify-center flex-shrink-0 group-hover:scale-105 transition-transform">
          <FolderPlus size={13} />
        </div>
        <div class="text-left">
          <div class="font-medium text-xs">Open New Folder...</div>
          <div class="text-[10px] font-mono text-ant-text-muted">Choose workspace folder from disk</div>
        </div>
      </button>
    </div>
  </div>
{/if}
