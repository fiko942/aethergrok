<script lang="ts">
  import { sessionStore } from '$lib/stores/session.svelte';
  import { formatMultipleSessionsAsMarkdown, downloadOrSaveMarkdown } from '$lib/utils/markdownExport';
  import { Download, Trash2, X } from 'lucide-svelte';
  import CustomCheckbox from '$lib/components/ui/CustomCheckbox.svelte';

  let { searchQuery = '' }: { searchQuery?: string } = $props();

  // Find all visible/filtered sessions currently displayed in sidebar across all workspaces
  const filteredSessions = $derived.by(() => {
    const q = searchQuery.trim().toLowerCase();
    if (!q) {
      return sessionStore.sessions;
    }
    return sessionStore.sessions.filter((s) => {
      const matchTitle = s.title.toLowerCase().includes(q);
      const ws = sessionStore.workspaces.find((w) => w.id === s.workspaceId);
      const matchWs = ws ? ws.name.toLowerCase().includes(q) : false;
      return matchTitle || matchWs;
    });
  });

  const totalFilteredCount = $derived(filteredSessions.length);
  const selectedFilteredCount = $derived(
    filteredSessions.filter((s) => sessionStore.selectedSessionIds.has(s.id)).length
  );
  const selectedCount = $derived(sessionStore.selectedSessionIds.size);

  const isAllSelected = $derived(totalFilteredCount > 0 && selectedFilteredCount === totalFilteredCount);
  const isIndeterminate = $derived(selectedFilteredCount > 0 && selectedFilteredCount < totalFilteredCount);

  let isExporting = $state(false);
  let actionFeedback = $state<string | null>(null);

  function showFeedback(msg: string) {
    actionFeedback = msg;
    setTimeout(() => {
      actionFeedback = null;
    }, 2800);
  }

  function handleToggleSelectAll() {
    if (isAllSelected) {
      // Deselect only the currently visible/filtered sessions
      const next = new Set(sessionStore.selectedSessionIds);
      for (const s of filteredSessions) {
        next.delete(s.id);
      }
      sessionStore.selectedSessionIds = next;
    } else {
      // Select all currently visible/filtered sessions across all workspaces
      const next = new Set(sessionStore.selectedSessionIds);
      for (const s of filteredSessions) {
        next.add(s.id);
      }
      sessionStore.selectedSessionIds = next;
    }
  }

  async function handleBatchExport() {
    if (selectedCount === 0) return;
    isExporting = true;

    try {
      const selectedSessions = sessionStore.sessions.filter((s) =>
        sessionStore.selectedSessionIds.has(s.id)
      );
      const ws = sessionStore.activeWorkspace;
      const mdContent = formatMultipleSessionsAsMarkdown(selectedSessions, ws);
      const dateStr = new Date().toISOString().slice(0, 10);
      const filename = `${ws?.name || 'workspace'}-sessions-batch-${dateStr}.md`;

      const res = await downloadOrSaveMarkdown(filename, mdContent);
      if (res.success) {
        showFeedback(`Exported ${selectedCount} session(s) successfully.`);
        sessionStore.toggleSelectionMode(false);
      } else if (res.error) {
        alert('Failed to export: ' + res.error);
      }
    } catch (err) {
      alert('Error during batch export: ' + String(err));
    } finally {
      isExporting = false;
    }
  }

  function handleBatchDelete() {
    if (selectedCount === 0) return;
    const confirmText = `Are you sure you want to delete ${selectedCount} selected session(s)? This action cannot be undone.`;
    if (window.confirm(confirmText)) {
      sessionStore.deleteSelectedSessions();
      showFeedback(`Successfully deleted session(s).`);
    }
  }

  function handleCancelSelection() {
    sessionStore.toggleSelectionMode(false);
  }
</script>

{#if sessionStore.isSelectionMode}
  <div class="p-2.5 bg-[#18181b] border border-[#27272a] rounded-xl shadow-2xl space-y-2 animate-in fade-in slide-in-from-bottom-2 duration-200">
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-2">
        <button
          onclick={handleToggleSelectAll}
          class="flex items-center space-x-2 text-xs text-zinc-300 hover:text-white font-medium transition group"
          title={isAllSelected ? 'Deselect visible sessions' : 'Select all visible sessions'}
        >
          <CustomCheckbox checked={isAllSelected} indeterminate={isIndeterminate} size="sm" />
          <span>{isAllSelected ? 'Deselect All' : 'Select All'}</span>
        </button>
      </div>

      <div class="flex items-center space-x-1.5">
        <span class="px-2 py-0.5 text-[11px] font-mono font-medium bg-[#222226] text-zinc-300 rounded-md border border-[#2e2e34]">
          {selectedCount} selected
        </span>
        <button
          onclick={handleCancelSelection}
          class="p-1 text-zinc-500 hover:text-zinc-200 rounded hover:bg-zinc-800 transition"
          title="Close Selection Mode"
        >
          <X size={14} />
        </button>
      </div>
    </div>

    <!-- Batch action buttons -->
    <div class="grid grid-cols-2 gap-2 pt-1.5 border-t border-[#27272a]">
      <button
        onclick={handleBatchExport}
        disabled={selectedCount === 0 || isExporting}
        class="flex items-center justify-center space-x-1.5 px-2 py-1.5 rounded-lg text-xs font-medium bg-[#222226] hover:bg-zinc-800 text-zinc-200 border border-transparent hover:border-zinc-700 transition disabled:opacity-40 disabled:pointer-events-none"
      >
        <Download size={13} class="text-zinc-400" />
        <span>{isExporting ? 'Exporting...' : 'Export (.md)'}</span>
      </button>

      <button
        onclick={handleBatchDelete}
        disabled={selectedCount === 0}
        class="flex items-center justify-center space-x-1.5 px-2 py-1.5 rounded-lg text-xs font-medium bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-transparent hover:border-rose-500/30 transition disabled:opacity-40 disabled:pointer-events-none"
      >
        <Trash2 size={13} class="text-rose-400" />
        <span>Delete</span>
      </button>
    </div>

    {#if actionFeedback}
      <div class="text-[11px] text-emerald-400 text-center font-medium bg-emerald-950/40 py-1 rounded border border-emerald-800/30">
        {actionFeedback}
      </div>
    {/if}
  </div>
{/if}
