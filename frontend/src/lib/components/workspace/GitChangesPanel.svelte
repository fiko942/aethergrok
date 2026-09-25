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

<div class="flex flex-col h-full bg-ant-bg text-ant-text font-sans text-xs select-none">
  <!-- Header: Branch & Actions -->
  <div class="p-2.5 border-b border-ant-border bg-ant-bg-secondary flex items-center justify-between shrink-0">
    <div class="flex items-center gap-2 min-w-0">
      <GitBranch class="w-3.5 h-3.5 text-ant-primary shrink-0" />
      <span class="font-mono font-medium text-ant-text truncate">
        {gitStatus?.branch || 'main'}
      </span>
      {#if gitStatus?.isClean}
        <span class="px-1.5 py-0.2 text-[10px] rounded bg-emerald-500/10 text-emerald-500 border border-emerald-500/20">
          Clean
        </span>
      {:else}
        <span class="px-1.5 py-0.2 text-[10px] rounded bg-amber-500/10 text-amber-500 border border-amber-500/20">
          Dirty ({gitStatus?.changedFiles?.length || 0})
        </span>
      {/if}
    </div>

    <div class="flex items-center gap-1.5 shrink-0">
      <button
        onclick={handlePull}
        disabled={isPulling}
        title="Pull latest changes from remote (git pull)"
        class="flex items-center gap-1 px-2 py-1 rounded-md text-[11px] font-medium bg-ant-bg-tertiary text-ant-text hover:bg-ant-primary/10 hover:text-ant-primary border border-ant-border transition-colors disabled:opacity-50"
      >
        <ArrowDown class="w-3.5 h-3.5 {isPulling ? 'animate-bounce text-ant-primary' : 'text-ant-text-muted'}" />
        <span>{isPulling ? 'Pulling...' : 'Pull'}</span>
      </button>

      <button
        onclick={() => refreshStatus(false)}
        disabled={isLoading}
        title="Refresh Git status and changes"
        class="flex items-center gap-1 px-2 py-1 rounded-md text-[11px] font-medium bg-ant-bg-tertiary text-ant-text hover:bg-ant-primary/10 hover:text-ant-primary border border-ant-border transition-colors disabled:opacity-50"
      >
        <RefreshCw class="w-3.5 h-3.5 {isLoading ? 'animate-spin text-ant-primary' : 'text-ant-text-muted'}" />
        <span>{isLoading ? 'Checking...' : 'Refresh'}</span>
      </button>
    </div>
  </div>

  <!-- Changes List -->
  <div class="flex-1 overflow-y-auto p-2 space-y-1 custom-scrollbar">
    {#if isLoading && !gitStatus}
      <div class="flex items-center justify-center py-8 text-ant-text-muted gap-2">
        <Loader2 class="w-4 h-4 animate-spin text-ant-primary" />
        <span>Inspecting repository...</span>
      </div>
    {:else if gitStatus?.isClean || !gitStatus?.changedFiles || gitStatus.changedFiles.length === 0}
      <div class="text-center py-8 text-ant-text-muted italic space-y-1">
        <div>No uncommitted changes</div>
        <div class="text-[11px] text-ant-text-muted/70">Working tree clean</div>
      </div>
    {:else}
      <div class="text-[11px] font-medium text-ant-text-secondary px-1 py-0.5 flex items-center justify-between">
        <span>CHANGED FILES ({gitStatus.changedFiles.length})</span>
        <div class="flex items-center gap-1.5 font-mono">
          <span class="text-emerald-500">+{gitStatus.totalAdditions}</span>
          <span class="text-rose-500">-{gitStatus.totalDeletions}</span>
        </div>
      </div>

      {#each gitStatus.changedFiles as change}
        {@const badge = getStatusBadge(change.status)}
        <div
          role="button"
          tabindex="0"
          onclick={() => (selectedDiffFile = change.path)}
          onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (selectedDiffFile = change.path)}
          class="flex items-center justify-between gap-2 p-1.5 rounded-lg bg-ant-bg-secondary hover:bg-ant-bg-tertiary border border-ant-border cursor-pointer group transition-all"
        >
          <div class="flex items-center gap-2 min-w-0 flex-1">
            <span class="font-mono text-[10px] px-1 py-0.2 rounded border font-semibold {badge.bg}">
              {badge.text}
            </span>
            <span class="truncate font-mono text-[11px] text-ant-text">
              {change.path}
            </span>
          </div>

          <div class="flex items-center gap-1.5 shrink-0 font-mono text-[10px]">
            {#if change.addedLines > 0}
              <span class="text-emerald-500">+{change.addedLines}</span>
            {/if}
            {#if change.removedLines > 0}
              <span class="text-rose-500">-{change.removedLines}</span>
            {/if}
          </div>
        </div>
      {/each}
    {/if}
  </div>

  <!-- Commit & Push Section -->
  {#if !gitStatus?.isClean && gitStatus?.changedFiles && gitStatus.changedFiles.length > 0}
    <div class="p-2.5 border-t border-ant-border bg-ant-bg-secondary space-y-2 shrink-0">
      {#if actionMessage}
        <div class="text-[11px] px-2 py-1 rounded {actionMessage.type === 'success' ? 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20' : 'bg-rose-500/10 text-rose-500 border border-rose-500/20'}">
          {actionMessage.text}
        </div>
      {/if}

      <div class="space-y-1.5">
        <textarea
          bind:value={commitMessage}
          placeholder="Commit message..."
          rows="2"
          class="w-full bg-ant-bg-tertiary text-ant-text placeholder-ant-text-muted rounded-lg p-2 text-xs border border-ant-border focus:border-ant-primary/40 focus:outline-none resize-none transition-colors"
        ></textarea>

        <div class="flex items-center gap-1.5">
          <button
            onclick={handleCommit}
            disabled={isCommitting || !commitMessage.trim()}
            class="flex-1 py-1.5 px-3 rounded-lg bg-ant-bg-tertiary hover:bg-ant-primary/15 text-ant-text hover:text-ant-primary text-xs font-medium border border-ant-border transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-1.5"
          >
            {#if isCommitting && !isPushing}
              <Loader2 class="w-3.5 h-3.5 animate-spin" />
              <span>Committing...</span>
            {:else}
              <Check class="w-3.5 h-3.5 text-ant-text-muted" />
              <span>Commit</span>
            {/if}
          </button>

          <button
            onclick={handleCommitAndPush}
            disabled={isCommitting || isPushing || !commitMessage.trim()}
            class="flex-1 py-1.5 px-3 rounded-lg bg-ant-primary hover:bg-ant-primary-hover text-white text-xs font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-1.5 shadow-sm"
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
