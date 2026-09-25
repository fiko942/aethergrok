<script lang="ts">
  import { onMount } from 'svelte';
  import {
    GitBranch,
    RefreshCw,
    Plus,
    Minus,
    FileCode,
    Check,
    ArrowUp,
    ArrowDown,
    Send,
    Loader2,
  } from 'lucide-svelte';
  import DiffViewModal from './DiffViewModal.svelte';

  interface GitChange {
    path: string;
    status: string; // 'M', 'A', 'D', '?'
    addedLines: number;
    removedLines: number;
  }

  interface GitStatus {
    branch: string;
    isClean: boolean;
    aheadCount: number;
    behindCount: number;
    changedFiles: GitChange[];
    totalAdditions: number;
    totalDeletions: number;
  }

  let {
    workspacePath,
  }: {
    workspacePath: string;
  } = $props();

  let gitStatus = $state<GitStatus | null>(null);
  let isLoading = $state(false);
  let isCommitting = $state(false);
  let isPushing = $state(false);
  let isPulling = $state(false);
  let commitMessage = $state('');
  let actionMessage = $state<{ type: 'success' | 'error'; text: string } | null>(null);

  // Selected file for diff modal
  let selectedDiffFile = $state<string | null>(null);

  export async function refreshStatus(silent = false) {
    if (!workspacePath) return;
    if (!silent) isLoading = true;

    try {
      const win = window as any;
      if (win.go?.main?.App?.GetWorkspaceGitStatus) {
        const res = await win.go.main.App.GetWorkspaceGitStatus(workspacePath);
        gitStatus = res;
      }
    } catch (err: any) {
      console.error('Failed to get git status:', err);
    } finally {
      if (!silent) isLoading = false;
    }
  }

  $effect(() => {
    if (workspacePath) {
      refreshStatus();

      // Auto-refresh every 5 seconds when panel is open
      const interval = setInterval(() => {
        // Do not interrupt while actively committing, pushing, or pulling
        if (!isCommitting && !isPushing && !isPulling) {
          refreshStatus(true);
        }
      }, 5000);

      return () => {
        clearInterval(interval);
      };
    }
  });

  async function handleCommit() {
    if (!commitMessage.trim() || !workspacePath) return;
    isCommitting = true;
    actionMessage = null;

    try {
      const win = window as any;
      if (win.go?.main?.App?.CommitWorkspaceChanges) {
        await win.go.main.App.CommitWorkspaceChanges(workspacePath, commitMessage.trim());
        commitMessage = '';
        actionMessage = { type: 'success', text: 'Committed successfully' };
        await refreshStatus();
      }
    } catch (err: any) {
      actionMessage = { type: 'error', text: err?.toString() || 'Commit failed' };
    } finally {
      isCommitting = false;
    }
  }

  async function handleCommitAndPush() {
    if (!commitMessage.trim() || !workspacePath) return;
    isCommitting = true;
    actionMessage = null;

    try {
      const win = window as any;
      if (win.go?.main?.App?.CommitWorkspaceChanges && win.go?.main?.App?.PushWorkspaceChanges) {
        await win.go.main.App.CommitWorkspaceChanges(workspacePath, commitMessage.trim());
        commitMessage = '';
        isPushing = true;
        await win.go.main.App.PushWorkspaceChanges(workspacePath);
        actionMessage = { type: 'success', text: 'Committed and pushed successfully' };
        await refreshStatus();
      }
    } catch (err: any) {
      actionMessage = { type: 'error', text: err?.toString() || 'Commit & Push failed' };
    } finally {
      isCommitting = false;
      isPushing = false;
    }
  }

  async function handlePull() {
    if (!workspacePath) return;
    isPulling = true;
    actionMessage = null;

    try {
      const win = window as any;
      if (win.go?.main?.App?.PullWorkspaceChanges) {
        const out = await win.go.main.App.PullWorkspaceChanges(workspacePath);
        actionMessage = { type: 'success', text: out || 'Pulled latest changes' };
        await refreshStatus();
      }
    } catch (err: any) {
      actionMessage = { type: 'error', text: err?.toString() || 'Pull failed' };
    } finally {
      isPulling = false;
    }
  }

  function getStatusBadge(status: string) {
    switch (status) {
      case 'M':
        return { text: 'M', bg: 'text-amber-400 bg-amber-400/10 border-amber-400/20' };
      case 'A':
        return { text: 'A', bg: 'text-emerald-400 bg-emerald-400/10 border-emerald-400/20' };
      case 'D':
        return { text: 'D', bg: 'text-rose-400 bg-rose-400/10 border-rose-400/20' };
      case '?':
        return { text: 'U', bg: 'text-cyan-400 bg-cyan-400/10 border-cyan-400/20' };
      default:
        return { text: status, bg: 'text-zinc-400 bg-zinc-400/10 border-zinc-400/20' };
    }
  }
</script>

<div class="flex flex-col h-full bg-[#18181b] text-zinc-300 font-sans text-xs select-none">
  <!-- Header: Branch & Actions -->
  <div class="p-2.5 border-b border-[#27272a] bg-[#141416] flex items-center justify-between shrink-0">
    <div class="flex items-center gap-2 min-w-0">
      <GitBranch class="w-3.5 h-3.5 text-indigo-400 shrink-0" />
      <span class="font-mono font-medium text-zinc-200 truncate">
        {gitStatus?.branch || 'main'}
      </span>
      {#if gitStatus?.isClean}
        <span class="px-1.5 py-0.2 text-[10px] rounded bg-emerald-950/60 text-emerald-400 border border-emerald-800/40">
          Clean
        </span>
      {:else}
        <span class="px-1.5 py-0.2 text-[10px] rounded bg-amber-950/60 text-amber-400 border border-amber-800/40">
          Dirty ({gitStatus?.changedFiles?.length || 0})
        </span>
      {/if}
    </div>

    <div class="flex items-center gap-1">
      <button
        onclick={handlePull}
        disabled={isPulling}
        title="Pull Changes (git pull)"
        class="p-1 rounded text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors disabled:opacity-50"
      >
        <ArrowDown class="w-3.5 h-3.5 {isPulling ? 'animate-bounce' : ''}" />
      </button>

      <button
        onclick={refreshStatus}
        disabled={isLoading}
        title="Refresh Status"
        class="p-1 rounded text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
      >
        <RefreshCw class="w-3.5 h-3.5 {isLoading ? 'animate-spin' : ''}" />
      </button>
    </div>
  </div>

  <!-- Changes List -->
  <div class="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
    {#if isLoading && !gitStatus}
      <div class="flex items-center justify-center py-8 text-zinc-500 gap-2">
        <Loader2 class="w-4 h-4 animate-spin text-zinc-400" />
        <span>Inspecting repository...</span>
      </div>
    {:else if gitStatus?.isClean || !gitStatus?.changedFiles || gitStatus.changedFiles.length === 0}
      <div class="text-center py-8 text-zinc-500 italic space-y-1">
        <div>No uncommitted changes</div>
        <div class="text-[11px] text-zinc-600">Working tree clean</div>
      </div>
    {:else}
      <div class="text-[11px] font-medium text-zinc-400 px-1 py-0.5 flex items-center justify-between">
        <span>CHANGED FILES ({gitStatus.changedFiles.length})</span>
        <div class="flex items-center gap-1.5 font-mono">
          <span class="text-emerald-400">+{gitStatus.totalAdditions}</span>
          <span class="text-rose-400">-{gitStatus.totalDeletions}</span>
        </div>
      </div>

      {#each gitStatus.changedFiles as change}
        {@const badge = getStatusBadge(change.status)}
        <div
          role="button"
          tabindex="0"
          onclick={() => (selectedDiffFile = change.path)}
          onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (selectedDiffFile = change.path)}
          class="flex items-center justify-between gap-2 p-1.5 rounded-lg bg-[#1a1a1d] hover:bg-zinc-800/80 border border-[#27272a] hover:border-zinc-700 cursor-pointer group transition-all"
        >
          <div class="flex items-center gap-2 min-w-0 flex-1">
            <span class="font-mono text-[10px] px-1 py-0.2 rounded border font-semibold {badge.bg}">
              {badge.text}
            </span>
            <span class="truncate font-mono text-[11px] text-zinc-300 group-hover:text-zinc-100">
              {change.path}
            </span>
          </div>

          <div class="flex items-center gap-1.5 shrink-0 font-mono text-[10px]">
            {#if change.addedLines > 0}
              <span class="text-emerald-400">+{change.addedLines}</span>
            {/if}
            {#if change.removedLines > 0}
              <span class="text-rose-400">-{change.removedLines}</span>
            {/if}
          </div>
        </div>
      {/each}
    {/if}
  </div>

  <!-- Commit & Push Section -->
  {#if !gitStatus?.isClean && gitStatus?.changedFiles && gitStatus.changedFiles.length > 0}
    <div class="p-2.5 border-t border-[#27272a] bg-[#141416] space-y-2 shrink-0">
      {#if actionMessage}
        <div class="text-[11px] px-2 py-1 rounded {actionMessage.type === 'success' ? 'bg-emerald-950/60 text-emerald-300 border border-emerald-800/40' : 'bg-rose-950/60 text-rose-300 border border-rose-800/40'}">
          {actionMessage.text}
        </div>
      {/if}

      <div class="space-y-1.5">
        <textarea
          bind:value={commitMessage}
          placeholder="Commit message..."
          rows="2"
          class="w-full bg-[#1e1e22] text-zinc-200 placeholder-zinc-500 rounded-lg p-2 text-xs border border-[#27272a] focus:border-zinc-600 focus:outline-none resize-none transition-colors"
        ></textarea>

        <div class="flex items-center gap-1.5">
          <button
            onclick={handleCommit}
            disabled={isCommitting || !commitMessage.trim()}
            class="flex-1 py-1.5 px-3 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-200 text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-1.5"
          >
            {#if isCommitting && !isPushing}
              <Loader2 class="w-3.5 h-3.5 animate-spin" />
              <span>Committing...</span>
            {:else}
              <Check class="w-3.5 h-3.5 text-zinc-400" />
              <span>Commit</span>
            {/if}
          </button>

          <button
            onclick={handleCommitAndPush}
            disabled={isCommitting || isPushing || !commitMessage.trim()}
            class="flex-1 py-1.5 px-3 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-1.5"
          >
            {#if isPushing}
              <Loader2 class="w-3.5 h-3.5 animate-spin" />
              <span>Pushing...</span>
            {:else}
              <Send class="w-3.5 h-3.5" />
              <span>Commit & Push</span>
            {/if}
          </button>
        </div>
      </div>
    </div>
  {/if}
</div>

<!-- Diff Viewer Modal mounted outside container with high z-index -->
{#if selectedDiffFile}
  <DiffViewModal
    isOpen={true}
    filePath={selectedDiffFile}
    workspacePath={workspacePath}
    onClose={() => (selectedDiffFile = null)}
  />
{/if}

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 5px;
    height: 5px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: #27272a;
    border-radius: 3px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: #3f3f46;
  }
</style>
