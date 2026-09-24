<script lang="ts">
  import type { DiffData } from '$lib/stores/session.svelte';
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

  const displayNewPath = $derived(newPath || diff?.newPath || 'modified');

  interface ParsedLine {
    type: 'add' | 'del' | 'normal';
    oldLineNo?: number;
    newLineNo?: number;
    content: string;
  }

  interface SideBySideRow {
    left?: { lineNo?: number; content: string; type: 'del' | 'normal' | 'empty' };
    right?: { lineNo?: number; content: string; type: 'add' | 'normal' | 'empty' };
  }

  // Parse unified diff or compute simple diff from old/new contents
  const parsedDiff = $derived.by(() => {
    const rawPatch = patch || diff?.diffUnified;
    const oldC = oldContent || diff?.oldContent;
    const newC = newContent || diff?.newContent;

    if (rawPatch) {
      return parseUnifiedPatch(rawPatch);
    } else if (oldC !== undefined || newC !== undefined) {
      return computeSimpleDiff(oldC || '', newC || '');
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
      } else if (line.type === 'del') {
        // Look ahead for corresponding add
        const nextLine = lines[i + 1];
        if (nextLine && nextLine.type === 'add') {
          rows.push({
            left: { lineNo: line.oldLineNo, content: line.content, type: 'del' },
            right: { lineNo: nextLine.newLineNo, content: nextLine.content, type: 'add' }
          });
          i += 2;
        } else {
          rows.push({
            left: { lineNo: line.oldLineNo, content: line.content, type: 'del' },
            right: { content: '', type: 'empty' }
          });
          i++;
        }
      } else if (line.type === 'add') {
        rows.push({
          left: { content: '', type: 'empty' },
          right: { lineNo: line.newLineNo, content: line.content, type: 'add' }
        });
        i++;
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

  function computeSimpleDiff(oldText: string, newText: string): { lines: ParsedLine[]; addedCount: number; removedCount: number } {
    const oldLines = oldText ? oldText.split('\n') : [];
    const newLines = newText ? newText.split('\n') : [];
    const lines: ParsedLine[] = [];
    let added = 0;
    let removed = 0;

    const max = Math.max(oldLines.length, newLines.length);
    for (let i = 0; i < max; i++) {
      const o = oldLines[i];
      const n = newLines[i];

      if (o === n) {
        if (o !== undefined) {
          lines.push({ type: 'normal', oldLineNo: i + 1, newLineNo: i + 1, content: o });
        }
      } else {
        if (o !== undefined) {
          lines.push({ type: 'del', oldLineNo: i + 1, content: o });
          removed++;
        }
        if (n !== undefined) {
          lines.push({ type: 'add', newLineNo: i + 1, content: n });
          added++;
        }
      }
    }

    return { lines, addedCount: added, removedCount: removed };
  }

  async function handleCopyDiff() {
    const rawPatch = patch || diff?.diffUnified || `${oldContent}\n---\n${newContent}`;
    try {
      await navigator.clipboard.writeText(rawPatch);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch {
      // ignore
    }
  }
</script>

<div class="rounded-md border border-ant-border-secondary/70 bg-ant-bg-secondary/40 overflow-hidden my-0.5 select-text font-mono text-xs">
  <!-- Diff Card Header (Clean, Non-Redundant) -->
  <div class="flex items-center justify-between px-2.5 py-1 bg-ant-bg-tertiary/50 border-b border-ant-border-secondary/60 select-none">
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
      <div class="flex items-center bg-ant-bg border border-ant-border-secondary rounded p-0.5">
        <button
          type="button"
          onclick={() => viewMode = 'unified'}
          class="px-1.5 py-0.5 rounded text-[10px] font-sans transition flex items-center {viewMode === 'unified' ? 'bg-ant-primary text-white font-medium' : 'text-ant-text-muted hover:text-ant-text'}"
          title="Unified Diff View"
        >
          <AlignJustify size={11} class="mr-1" /> Unified
        </button>
        <button
          type="button"
          onclick={() => viewMode = 'split'}
          class="px-1.5 py-0.5 rounded text-[10px] font-sans transition flex items-center {viewMode === 'split' ? 'bg-ant-primary text-white font-medium' : 'text-ant-text-muted hover:text-ant-text'}"
          title="Side-by-Side Split View"
        >
          <Columns size={11} class="mr-1" /> Split
        </button>
      </div>

      <button
        type="button"
        onclick={handleCopyDiff}
        class="p-1 rounded text-ant-text-muted hover:text-white hover:bg-ant-bg-secondary transition"
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
              <tr class="bg-emerald-500/10 hover:bg-emerald-500/15 text-emerald-300 border-l-2 border-emerald-500">
                <td class="w-7 px-1.5 text-right text-ant-text-muted/30 border-r border-ant-border-secondary/40 select-none text-[10px]"></td>
                <td class="w-7 px-1.5 text-right text-emerald-400 font-semibold border-r border-ant-border-secondary/40 select-none text-[10px]">{line.newLineNo}</td>
                <td class="px-2.5 py-0.5 whitespace-pre font-mono flex items-start">
                  <span class="text-emerald-400 font-bold mr-2 select-none">+</span>
                  <span>{line.content}</span>
                </td>
              </tr>
            {:else if line.type === 'del'}
              <tr class="bg-rose-500/10 hover:bg-rose-500/15 text-rose-300 border-l-2 border-rose-500">
                <td class="w-7 px-1.5 text-right text-rose-400 font-semibold border-r border-ant-border-secondary/40 select-none text-[10px]">{line.oldLineNo}</td>
                <td class="w-7 px-1.5 text-right text-ant-text-muted/30 border-r border-ant-border-secondary/40 select-none text-[10px]"></td>
                <td class="px-2.5 py-0.5 whitespace-pre font-mono flex items-start">
                  <span class="text-rose-400 font-bold mr-2 select-none">-</span>
                  <span>{line.content}</span>
                </td>
              </tr>
            {:else}
              <tr class="hover:bg-ant-bg-secondary/30 text-ant-text">
                <td class="w-7 px-1.5 text-right text-ant-text-muted/70 border-r border-ant-border-secondary/40 select-none text-[10px]">{line.oldLineNo}</td>
                <td class="w-7 px-1.5 text-right text-ant-text-muted/70 border-r border-ant-border-secondary/40 select-none text-[10px]">{line.newLineNo}</td>
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
            <tr class="border-b border-ant-border-secondary/15">
              <!-- Left Side (Original / Deletions) -->
              <td class="w-7 px-1.5 text-right text-ant-text-muted border-r border-ant-border-secondary/40 select-none bg-ant-bg-secondary/20 text-[10px]">
                {row.left?.lineNo ?? ''}
              </td>
              <td
                class="w-1/2 px-2.5 py-0.5 whitespace-pre font-mono border-r border-ant-border-secondary/40 {row.left?.type === 'del'
                  ? 'bg-rose-500/10 text-rose-300 border-l-2 border-rose-500'
                  : 'text-ant-text'}"
              >
                {#if row.left?.type === 'del'}
                  <span class="text-rose-400 font-bold mr-1.5 select-none">-</span>
                {/if}
                {row.left?.content ?? ''}
              </td>

              <!-- Right Side (Modified / Additions) -->
              <td class="w-7 px-1.5 text-right text-ant-text-muted border-r border-ant-border-secondary/40 select-none bg-ant-bg-secondary/20 text-[10px]">
                {row.right?.lineNo ?? ''}
              </td>
              <td
                class="w-1/2 px-2.5 py-0.5 whitespace-pre font-mono {row.right?.type === 'add'
                  ? 'bg-emerald-500/10 text-emerald-300 border-l-2 border-emerald-500'
                  : 'text-ant-text'}"
              >
                {#if row.right?.type === 'add'}
                  <span class="text-emerald-400 font-bold mr-1.5 select-none">+</span>
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
