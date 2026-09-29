<script lang="ts">
  import type { DiffData } from '$lib/stores/session.svelte';
  import { computeStringDiff, type DiffParsedLine } from '$lib/utils/diffUtils';
  import { Columns, AlignJustify, Copy, Check, Plus, Minus } from 'lucide-svelte';

  interface Props {
    diff?: DiffData;
    patch?: string;
    oldPath?: string;
    newPath?: string;
    oldContent?: string;
    newContent?: string;
    showHeaderTitle?: boolean;
  }

  let {
    diff,
    patch,
    oldPath,
    newPath,
    oldContent,
    newContent,
    showHeaderTitle = false
  }: Props = $props();

  let viewMode = $state<'unified' | 'split'>('unified');
  let copied = $state(false);

  const displayNewPath = $derived(newPath || diff?.newPath || oldPath || diff?.oldPath || 'modified');

  interface ParsedLine extends DiffParsedLine {}

  interface SideBySideRow {
    left?: { lineNo?: number; content: string; type: 'del' | 'normal' | 'empty' };
    right?: { lineNo?: number; content: string; type: 'add' | 'normal' | 'empty' };
  }

  // Parse unified diff or compute accurate LCS diff from old/new contents
  const parsedDiff = $derived.by(() => {
    const rawPatch = patch || diff?.diffUnified;
    const oldC = oldContent !== undefined ? oldContent : diff?.oldContent;
    const newC = newContent !== undefined ? newContent : diff?.newContent;

    if (rawPatch) {
      return parseUnifiedPatch(rawPatch);
    } else if (oldC !== undefined || newC !== undefined) {
      return computeStringDiff(oldC || '', newC || '');
    }
    return { lines: [], addedCount: 0, removedCount: 0 };
  });

  const sideBySideRows = $derived.by(() => {
    const lines = parsedDiff.lines;
    const rows: SideBySideRow[] = [];

    let i = 0;
    while (i < lines.length) {
      const line = lines[i];

      if (line.type === 'normal') {
        rows.push({
          left: { lineNo: line.oldLineNo, content: line.content, type: 'normal' },
          right: { lineNo: line.newLineNo, content: line.content, type: 'normal' }
        });
        i++;
      } else {
        // Collect contiguous deletions and additions to align nicely side-by-side
        const delLines: ParsedLine[] = [];
        const addLines: ParsedLine[] = [];

        while (i < lines.length && (lines[i].type === 'del' || lines[i].type === 'add')) {
          if (lines[i].type === 'del') {
            delLines.push(lines[i]);
          } else {
            addLines.push(lines[i]);
          }
          i++;
        }

        const maxLen = Math.max(delLines.length, addLines.length);
        for (let k = 0; k < maxLen; k++) {
          const d = delLines[k];
          const a = addLines[k];
          rows.push({
            left: d ? { lineNo: d.oldLineNo, content: d.content, type: 'del' } : { content: '', type: 'empty' },
            right: a ? { lineNo: a.newLineNo, content: a.content, type: 'add' } : { content: '', type: 'empty' }
          });
        }
      }
    }

    return rows;
  });

  // Filter out raw git noise (---, +++, @@, diff --git, index) and parse actual changed lines
  function parseUnifiedPatch(raw: string): { lines: ParsedLine[]; addedCount: number; removedCount: number } {
    const rawLines = raw.split('\n');
    const lines: ParsedLine[] = [];
    let oldLine = 1;
    let newLine = 1;
    let added = 0;
    let removed = 0;

    for (const l of rawLines) {
      if (l.startsWith('@@')) {
        const match = l.match(/@@\s*-(\d+)(?:,\d+)?\s*\+(\d+)(?:,\d+)?\s*@@/);
        if (match) {
          oldLine = parseInt(match[1], 10);
          newLine = parseInt(match[2], 10);
        }
        // Omit raw hunk headers from user view
      } else if (l.startsWith('---') || l.startsWith('+++') || l.startsWith('diff --git') || l.startsWith('index ')) {
        // Suppress raw git header metadata
      } else if (l.startsWith('+')) {
        lines.push({ type: 'add', newLineNo: newLine++, content: l.slice(1) });
        added++;
      } else if (l.startsWith('-')) {
        lines.push({ type: 'del', oldLineNo: oldLine++, content: l.slice(1) });
        removed++;
      } else if (l.startsWith(' ') || l === '') {
        const text = l.startsWith(' ') ? l.slice(1) : l;
        lines.push({ type: 'normal', oldLineNo: oldLine++, newLineNo: newLine++, content: text });
      }
    }

    return { lines, addedCount: added, removedCount: removed };
  }

  async function handleCopyDiff() {
    let rawPatch = patch || diff?.diffUnified;
    if (!rawPatch && (oldContent !== undefined || newContent !== undefined || diff?.oldContent !== undefined || diff?.newContent !== undefined)) {
      const o = oldContent ?? diff?.oldContent ?? '';
      const n = newContent ?? diff?.newContent ?? '';
      rawPatch = `--- a/${displayNewPath}\n+++ b/${displayNewPath}\n` +
        parsedDiff.lines.map(l => (l.type === 'add' ? `+${l.content}` : l.type === 'del' ? `-${l.content}` : ` ${l.content}`)).join('\n');
    }
    rawPatch = rawPatch || `${oldContent || ''}\n---\n${newContent || ''}`;
    try {
      await navigator.clipboard.writeText(rawPatch);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      // ignore
    }
  }
</script>

<div class="rounded-md border border-white/5 bg-ant-bg-secondary/30 overflow-hidden my-0.5 select-text font-mono text-xs">
  <!-- Diff Card Header (Clean, Non-Redundant) -->
  <div class="flex items-center justify-between px-2.5 py-1 bg-ant-bg-tertiary/40 border-b border-white/5 select-none">
    <div class="flex items-center space-x-2 min-w-0">
      {#if showHeaderTitle}
        <span class="font-medium text-ant-text truncate max-w-xs">{displayNewPath}</span>
      {:else}
        <span class="text-[11px] font-semibold text-ant-text-secondary tracking-wide uppercase">Diff View</span>
      {/if}

      <!-- Stats chip -->
      <div class="flex items-center space-x-1 ml-1 text-[10px] font-mono">
        <span class="text-emerald-400 flex items-center bg-emerald-500/15 px-1.5 py-0.2 rounded font-medium">
          <Plus size={10} class="mr-0.5" />{parsedDiff.addedCount}
        </span>
        <span class="text-rose-400 flex items-center bg-rose-500/15 px-1.5 py-0.2 rounded font-medium">
          <Minus size={10} class="mr-0.5" />{parsedDiff.removedCount}
        </span>
      </div>
    </div>

    <!-- Actions: View Mode toggle & Copy -->
    <div class="flex items-center space-x-1.5 flex-shrink-0">
      <div class="flex items-center bg-ant-bg border border-white/5 rounded p-0.5">
        <button
          type="button"
          onclick={() => viewMode = 'unified'}
          class="px-1.5 py-0.5 rounded text-[10px] font-serif transition flex items-center {viewMode === 'unified' ? 'bg-ant-primary text-white font-medium' : 'text-ant-text-muted hover:text-ant-text'}"
          title="Unified Diff View"
        >
          <AlignJustify size={11} class="mr-1" /> Unified
        </button>
        <button
          type="button"
          onclick={() => viewMode = 'split'}
          class="px-1.5 py-0.5 rounded text-[10px] font-serif transition flex items-center {viewMode === 'split' ? 'bg-ant-primary text-white font-medium' : 'text-ant-text-muted hover:text-ant-text'}"
          title="Side-by-Side Split View"
        >
          <Columns size={11} class="mr-1" /> Split
        </button>
      </div>

      <button
        type="button"
        onclick={handleCopyDiff}
        class="p-1 rounded text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-secondary transition"
        title="Copy raw diff"
      >
        {#if copied}
          <Check size={12} class="text-ant-success" />
        {:else}
          <Copy size={12} />
        {/if}
      </button>
    </div>
  </div>

  <!-- Diff Body -->
  <div class="overflow-x-auto max-h-[360px] scrollbar-thin">
    {#if viewMode === 'unified'}
      <!-- Unified View -->
      <table class="w-full border-collapse font-mono text-[11px] leading-relaxed">
        <tbody>
          {#each parsedDiff.lines as line, idx (idx)}
            {#if line.type === 'add'}
              <tr class="bg-emerald-500/10 hover:bg-emerald-500/15 text-emerald-800 dark:text-emerald-300 border-l-2 border-emerald-500/50">
                <td class="w-7 px-1.5 text-right text-ant-text-muted/30 border-r border-white/5 select-none text-[10px]"></td>
                <td class="w-7 px-1.5 text-right text-emerald-700 dark:text-emerald-400 font-semibold border-r border-white/5 select-none text-[10px]">{line.newLineNo}</td>
                <td class="px-2.5 py-0.5 whitespace-pre font-mono flex items-start">
                  <span class="text-emerald-700 dark:text-emerald-400 font-bold mr-2 select-none">+</span>
                  <span>{line.content}</span>
                </td>
              </tr>
            {:else if line.type === 'del'}
              <tr class="bg-rose-500/10 hover:bg-rose-500/15 text-rose-800 dark:text-rose-300 border-l-2 border-rose-500/50">
                <td class="w-7 px-1.5 text-right text-rose-700 dark:text-rose-400 font-semibold border-r border-white/5 select-none text-[10px]">{line.oldLineNo}</td>
                <td class="w-7 px-1.5 text-right text-ant-text-muted/30 border-r border-white/5 select-none text-[10px]"></td>
                <td class="px-2.5 py-0.5 whitespace-pre font-mono flex items-start">
                  <span class="text-rose-700 dark:text-rose-400 font-bold mr-2 select-none">-</span>
                  <span>{line.content}</span>
                </td>
              </tr>
            {:else}
              <tr class="hover:bg-ant-bg-secondary/30 text-ant-text">
                <td class="w-7 px-1.5 text-right text-ant-text-muted/70 border-r border-white/5 select-none text-[10px]">{line.oldLineNo}</td>
                <td class="w-7 px-1.5 text-right text-ant-text-muted/70 border-r border-white/5 select-none text-[10px]">{line.newLineNo}</td>
                <td class="px-2.5 py-0.5 whitespace-pre font-mono flex items-start">
                  <span class="text-ant-text-muted mr-2 select-none opacity-0"> </span>
                  <span>{line.content}</span>
                </td>
              </tr>
            {/if}
          {/each}
        </tbody>
      </table>
    {:else}
      <!-- Side-by-side Split View -->
      <table class="w-full border-collapse font-mono text-[11px] leading-relaxed">
        <tbody>
          {#each sideBySideRows as row, idx (idx)}
            <tr class="border-b border-white/5">
              <!-- Left Side (Original / Deletions) -->
              <td class="w-7 px-1.5 text-right text-ant-text-muted border-r border-white/5 select-none bg-ant-bg-secondary/20 text-[10px]">
                {row.left?.lineNo ?? ''}
              </td>
              <td
                class="w-1/2 px-2.5 py-0.5 whitespace-pre font-mono border-r border-white/5 {row.left?.type === 'del'
                  ? 'bg-rose-500/10 text-rose-800 dark:text-rose-300 border-l-2 border-rose-500/50'
                  : 'text-ant-text'}"
              >
                {#if row.left?.type === 'del'}
                  <span class="text-rose-700 dark:text-rose-400 font-bold mr-1.5 select-none">-</span>
                {/if}
                {row.left?.content ?? ''}
              </td>

              <!-- Right Side (Modified / Additions) -->
              <td class="w-7 px-1.5 text-right text-ant-text-muted border-r border-white/5 select-none bg-ant-bg-secondary/20 text-[10px]">
                {row.right?.lineNo ?? ''}
              </td>
              <td
                class="w-1/2 px-2.5 py-0.5 whitespace-pre font-mono {row.right?.type === 'add'
                  ? 'bg-emerald-500/10 text-emerald-800 dark:text-emerald-300 border-l-2 border-emerald-500/50'
                  : 'text-ant-text'}"
              >
                {#if row.right?.type === 'add'}
                  <span class="text-emerald-700 dark:text-emerald-400 font-bold mr-1.5 select-none">+</span>
                {/if}
                {row.right?.content ?? ''}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>
