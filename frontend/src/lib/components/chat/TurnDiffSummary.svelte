<script lang="ts">
  import type { ToolCall } from '$lib/stores/session.svelte';
  import { calculateDiffStat, extractToolCallFileChange, type ToolCallFileChange } from '$lib/utils/diffUtils';
  import DiffCard from './DiffCard.svelte';
  import { FileCode, ChevronDown, ChevronRight, Plus, Minus, ExternalLink } from 'lucide-svelte';

  interface Props {
    toolCalls: ToolCall[];
  }

  let { toolCalls }: Props = $props();

  let isExpanded = $state(false);
  let expandedFileIndex = $state<number | null>(null);

  // Aggregate all tool calls that modified files with diffs or line counts
  const changedFiles = $derived.by(() => {
    const list: ToolCallFileChange[] = [];
    for (const tc of toolCalls) {
      const change = extractToolCallFileChange(tc);
      if (
        change.addedCount > 0 ||
        change.removedCount > 0 ||
        !!change.diff ||
        !!change.diffUnified ||
        change.oldContent !== undefined ||
        change.newContent !== undefined
      ) {
        list.push(change);
      }
    }
    return list;
  });

  const totalAdded = $derived(changedFiles.reduce((acc, f) => acc + f.addedCount, 0));
  const totalRemoved = $derived(changedFiles.reduce((acc, f) => acc + f.removedCount, 0));

  function toggleFile(idx: number, e: MouseEvent) {
    e.stopPropagation();
    expandedFileIndex = expandedFileIndex === idx ? null : idx;
  }
</script>

{#if changedFiles.length > 0}
  <div
    class="my-2 rounded-xl border border-ant-border-secondary dark:border-white/5 bg-ant-bg-secondary/70 overflow-hidden shadow-sm font-sans select-none"
    style="content-visibility: auto; contain-intrinsic-size: auto 38px;"
  >
    <!-- Top Summary Banner -->
    <div
      role="button"
      tabindex="0"
      onclick={() => isExpanded = !isExpanded}
      onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && (isExpanded = !isExpanded)}
      class="flex items-center justify-between px-3.5 py-2 cursor-pointer hover:bg-white/[0.03] transition-colors"
    >
      <div class="flex items-center space-x-2.5 min-w-0">
        <span class="text-ant-text-secondary transition-transform">
          {#if isExpanded}
            <ChevronDown size={14} />
          {:else}
            <ChevronRight size={14} />
          {/if}
        </span>
        <FileCode size={14} class="text-ant-primary flex-shrink-0" />
        <span class="text-xs font-semibold text-ant-text">
          Changed {changedFiles.length} {changedFiles.length === 1 ? 'file' : 'files'}
        </span>
      </div>

      <!-- Add/Remove Totals Pills -->
      <div class="flex items-center space-x-1.5 font-mono text-[11px]">
        {#if totalAdded > 0}
          <span class="flex items-center space-x-0.5 px-1.5 py-0.5 rounded bg-emerald-500/15 text-emerald-400 font-semibold border border-emerald-500/20">
            <Plus size={10} />
            <span>{totalAdded}</span>
          </span>
        {/if}
        {#if totalRemoved > 0}
          <span class="flex items-center space-x-0.5 px-1.5 py-0.5 rounded bg-rose-500/15 text-rose-400 font-semibold border border-rose-500/20">
            <Minus size={10} />
            <span>{totalRemoved}</span>
          </span>
        {/if}
      </div>
    </div>

    <!-- Collapsible File List Breakdown -->
    {#if isExpanded}
      <div class="px-2 pb-2 pt-0.5 space-y-1 border-t border-white/5 divide-y divide-white/5 bg-ant-bg/60">
        {#each changedFiles as item, idx}
          {@const isThisFileExpanded = expandedFileIndex === idx}
          <div class="pt-1 first:pt-0">
            <!-- File Row Header -->
            <div
              role="button"
              tabindex="0"
              onclick={(e) => toggleFile(idx, e)}
              onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && toggleFile(idx, e as any)}
              class="flex items-center justify-between p-1.5 rounded-lg hover:bg-ant-bg-tertiary/70 cursor-pointer transition text-xs"
            >
              <div class="flex items-center space-x-2 min-w-0 pr-2">
                <span class="text-ant-text-muted">
                  {#if isThisFileExpanded}
                    <ChevronDown size={12} />
                  {:else}
                    <ChevronRight size={12} />
                  {/if}
                </span>
                <span class="font-mono text-[11.5px] text-ant-text truncate" title={item.path}>
                  {item.path}
                </span>
              </div>

              <div class="flex items-center space-x-1 font-mono text-[10.5px] flex-shrink-0">
                {#if item.addedCount > 0}
                  <span class="text-emerald-400 font-medium">+{item.addedCount}</span>
                {/if}
                {#if item.removedCount > 0}
                  <span class="text-rose-400 font-medium">−{item.removedCount}</span>
                {/if}
              </div>
            </div>

            <!-- Inline Diff Preview for Selected File -->
            {#if isThisFileExpanded}
              <div class="mt-1 pl-4 pr-1 pb-1 animate-in fade-in duration-150">
                {#if item.diff}
                  <DiffCard diff={item.diff} newPath={item.path} showHeaderTitle={false} />
                {:else if item.diffUnified}
                  <DiffCard patch={item.diffUnified} newPath={item.path} showHeaderTitle={false} />
                {:else if item.oldContent !== undefined || item.newContent !== undefined}
                  <DiffCard
                    newPath={item.path}
                    oldContent={item.oldContent ?? ''}
                    newContent={item.newContent ?? ''}
                    showHeaderTitle={false}
                  />
                {:else}
                  <div class="p-2 rounded bg-ant-bg-secondary/60 text-xs font-mono text-ant-text-secondary">
                    Changes applied in tool call ({item.toolCallId})
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/if}
