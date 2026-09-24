<script lang="ts">
  import { sessionStore } from '$lib/stores/session.svelte';
  import { formatMultipleSessionsAsMarkdown, downloadOrSaveMarkdown } from '$lib/utils/markdownExport';
  import { CheckSquare, Square, Download, Trash2, X, Check } from 'lucide-svelte';

  const selectedCount = $derived(sessionStore.selectedSessionIds.size);
  const totalCount = $derived(sessionStore.activeWorkspaceSessions.length);
  const isAllSelected = $derived(totalCount > 0 && selectedCount === totalCount);

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
      sessionStore.deselectAllSessions();
    } else {
      sessionStore.selectAllSessions();
    }
  }

  async function handleBatchExport() {
    if (selectedCount === 0) return;
    isExporting = true;

    try {
      const selectedSessions = sessionStore.activeWorkspaceSessions.filter((s) =>
        sessionStore.selectedSessionIds.has(s.id)
      );
      const ws = sessionStore.activeWorkspace;
      const mdContent = formatMultipleSessionsAsMarkdown(selectedSessions, ws);
      const dateStr = new Date().toISOString().slice(0, 10);
      const filename = `${ws?.name || 'workspace'}-sessions-batch-${dateStr}.md`;

      const res = await downloadOrSaveMarkdown(filename, mdContent);
      if (res.success) {
        showFeedback(`Exported ${selectedCount} session(s) successfully!`);
        sessionStore.toggleSelectionMode(false);
      } else if (res.error) {
        alert('Gagal mengekspor: ' + res.error);
      }
    } catch (err) {
      alert('Terjadi kesalahan saat batch export: ' + String(err));
    } finally {
      isExporting = false;
    }
  }

  function handleBatchDelete() {
    if (selectedCount === 0) return;
    const confirmText = `Yakin ingin menghapus ${selectedCount} session yang terpilih? Tindakan ini tidak dapat dibatalkan.`;
    if (window.confirm(confirmText)) {
      sessionStore.deleteSelectedSessions();
      showFeedback(`Berhasil menghapus session.`);
    }
  }

  function handleCancelSelection() {
    sessionStore.toggleSelectionMode(false);
  }
</script>

{#if sessionStore.isSelectionMode}
  <div class="p-2.5 bg-ant-bg border border-ant-primary/40 rounded-lg shadow-xl shadow-black/40 space-y-2 animate-in fade-in slide-in-from-bottom-2 duration-200">
    <div class="flex items-center justify-between">
      <div class="flex items-center space-x-2">
        <button
          onclick={handleToggleSelectAll}
          class="flex items-center space-x-1.5 text-xs text-ant-primary hover:text-ant-primary-hover font-medium transition"
          title={isAllSelected ? 'Batalkan pilihan semua' : 'Tandai semua session di workspace'}
        >
          {#if isAllSelected}
            <CheckSquare size={14} class="text-ant-primary" />
            <span>Batalkan Semua</span>
          {:else}
            <Square size={14} class="text-ant-text-secondary" />
            <span>Tandai Semua</span>
          {/if}
        </button>
      </div>

      <div class="flex items-center space-x-1.5">
        <span class="px-2 py-0.5 text-[11px] font-semibold bg-ant-primary/20 text-ant-primary rounded-full border border-ant-primary/30">
          {selectedCount} Terpilih
        </span>
        <button
          onclick={handleCancelSelection}
          class="p-1 text-ant-text-muted hover:text-white rounded hover:bg-ant-bg-tertiary transition"
          title="Tutup Mode Seleksi"
        >
          <X size={14} />
        </button>
      </div>
    </div>

    <!-- Batch action buttons -->
    <div class="grid grid-cols-2 gap-2 pt-1 border-t border-ant-border-secondary">
      <button
        onclick={handleBatchExport}
        disabled={selectedCount === 0 || isExporting}
        class="flex items-center justify-center space-x-1.5 px-2 py-1.5 rounded-md text-xs font-medium bg-ant-primary/20 hover:bg-ant-primary/30 text-ant-primary border border-ant-primary/30 transition disabled:opacity-40 disabled:pointer-events-none"
      >
        <Download size={13} />
        <span>{isExporting ? 'Mengekspor...' : 'Ekspor (.md)'}</span>
      </button>

      <button
        onclick={handleBatchDelete}
        disabled={selectedCount === 0}
        class="flex items-center justify-center space-x-1.5 px-2 py-1.5 rounded-md text-xs font-medium bg-ant-error/15 hover:bg-ant-error/25 text-ant-error border border-ant-error/30 transition disabled:opacity-40 disabled:pointer-events-none"
      >
        <Trash2 size={13} />
        <span>Hapus</span>
      </button>
    </div>

    {#if actionFeedback}
      <div class="text-[11px] text-ant-success text-center font-medium bg-ant-success/10 py-1 rounded border border-ant-success/20">
        {actionFeedback}
      </div>
    {/if}
  </div>
{/if}
