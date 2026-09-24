<script lang="ts">
  import type { ToolCall } from '$lib/stores/session.svelte';
  import DiffCard from './DiffCard.svelte';
  import {
    Terminal,
    FileText,
    FileEdit,
    Search,
    ChevronDown,
    ChevronRight,
    Loader2,
    CheckCircle2,
    XCircle,
    Copy,
    Check,
    Folder,
    FileCode,
    Replace
  } from 'lucide-svelte';

  interface Props {
    toolCall: ToolCall;
  }

  let { toolCall }: Props = $props();

  let isExpanded = $state(false);
  let copiedOutput = $state(false);

  // Derive execution duration
  const durationMs = $derived.by(() => {
    if (!toolCall.startTime) return null;
    const end = toolCall.endTime || Date.now();
    return Math.max(0, end - toolCall.startTime);
  });

  // Extract tool category / verb metadata matching grok-build-vscode
  const toolParsed = $derived.by(() => {
    const rawTool = (toolCall.tool || '').toLowerCase();
    const p = (typeof toolCall.params === 'object' && toolCall.params !== null)
      ? (toolCall.params as Record<string, unknown>)
      : {};

    // 1. Read File
    if (rawTool.includes('read_file') || rawTool.includes('file_read') || rawTool.includes('cat')) {
      const filePath = String(p.target_file || p.file_path || p.path || p.filePath || '');
      let lineRange = '';
      let startLine = 1;
      if (p.offset && p.limit) {
        startLine = Number(p.offset);
        const end = startLine + Number(p.limit) - 1;
        lineRange = ` (lines ${startLine}-${end})`;
      } else if (p.offset) {
        startLine = Number(p.offset);
        lineRange = ` (lines ${startLine}+)`;
      } else if (p.limit) {
        lineRange = ` (lines 1-${p.limit})`;
      }
      return {
        verb: 'Read',
        type: 'read' as const,
        target: filePath ? `${filePath}${lineRange}` : 'file',
        filePath,
        lineRange,
        startLine,
        icon: FileCode,
        iconClass: 'text-blue-400'
      };
    }

    // 2. Edit / Search & Replace
    if (rawTool.includes('search_replace') || rawTool.includes('edit') || rawTool.includes('str_replace')) {
      const filePath = String(p.file_path || p.target_file || p.path || p.filePath || '');
      const oldStr = String(p.old_string || p.find || '');
      const newStr = String(p.new_string || p.replace || '');
      return {
        verb: 'Edit',
        type: 'edit' as const,
        target: filePath || 'file',
        filePath,
        oldStr,
        newStr,
        icon: Replace,
        iconClass: 'text-amber-400'
      };
    }

    // 3. Write / Save File
    if (rawTool.includes('write') || rawTool.includes('save') || rawTool.includes('create_file')) {
      const filePath = String(p.file_path || p.path || p.target_file || p.filePath || '');
      return {
        verb: 'Write',
        type: 'write' as const,
        target: filePath || 'file',
        filePath,
        icon: FileEdit,
        iconClass: 'text-emerald-400'
      };
    }

    // 4. Search / Grep / Ripgrep
    if (rawTool.includes('grep') || rawTool.includes('search') || rawTool.includes('find')) {
      const query = String(p.pattern || p.query || p.text || p.regex || p.glob_pattern || '');
      const path = String(p.path || p.target_directory || p.directory || '');
      const target = query ? `"${query}"${path && path !== '.' ? ` in ${path}` : ''}` : path || 'files';
      return {
        verb: 'Search',
        type: 'search' as const,
        target,
        query,
        path,
        icon: Search,
        iconClass: 'text-amber-400'
      };
    }

    // 5. List Directory
    if (rawTool.includes('list_dir') || rawTool.includes('list_directory') || rawTool.includes('ls')) {
      const dirPath = String(p.target_directory || p.path || p.directory || p.dir || '.');
      return {
        verb: 'List',
        type: 'list' as const,
        target: dirPath,
        icon: Folder,
        iconClass: 'text-indigo-400'
      };
    }

    // 6. Terminal / Bash / Exec
    if (rawTool.includes('terminal') || rawTool.includes('bash') || rawTool.includes('exec') || rawTool.includes('cmd')) {
      const cmd = String(p.command || p.cmd || '');
      return {
        verb: 'Run',
        type: 'terminal' as const,
        target: cmd ? `$ ${cmd}` : 'command',
        cmd,
        icon: Terminal,
        iconClass: 'text-indigo-400'
      };
    }

    // Fallback Tool
    return {
      verb: toolCall.tool || 'Tool',
      type: 'custom' as const,
      target: typeof toolCall.params === 'string' ? toolCall.params : '',
      icon: FileText,
      iconClass: 'text-ant-primary'
    };
  });

  // Extract inline diff stat if available
  const diffStat = $derived.by(() => {
    if (!toolCall.diff) return null;
    const patch = toolCall.diff.diffUnified || '';
    let added = 0;
    let removed = 0;
    for (const line of patch.split('\n')) {
      if (line.startsWith('+') && !line.startsWith('+++')) added++;
      else if (line.startsWith('-') && !line.startsWith('---')) removed++;
    }
    return { added, removed };
  });

  // Parse file content into numbered lines for Read preview
  const readLines = $derived.by(() => {
    if (toolParsed.type !== 'read' || !toolCall.result) return [];
    const lines = toolCall.result.split('\n');
    const start = ('startLine' in toolParsed && typeof toolParsed.startLine === 'number')
      ? toolParsed.startLine
      : 1;
    return lines.map((text, idx) => ({
      lineNo: start + idx,
      text
    }));
  });

  // Determine whether this output is redundant noise (e.g. standard success ack for edits)
  const isGenericEditAck = $derived.by(() => {
    if (toolParsed.type !== 'edit') return false;
    const res = (toolCall.result || '').toLowerCase();
    return (
      res.includes('successfully') ||
      res.includes('updated timeout') ||
      res.includes('1 replacement applied') ||
      res.includes('ok')
    );
  });

  async function copyOutputText() {
    if (!toolCall.result) return;
    try {
      await navigator.clipboard.writeText(toolCall.result);
      copiedOutput = true;
      setTimeout(() => (copiedOutput = false), 2000);
    } catch {
      // ignore
    }
  }

  function toggleExpand(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (target.closest('button') || target.closest('a')) return;
    isExpanded = !isExpanded;
  }
</script>

<div class="my-1 rounded-md border border-ant-border-secondary/60 bg-ant-bg-secondary/40 overflow-hidden text-xs transition-colors hover:border-ant-border">
  <!-- Minimalist Tool Header (Anti Gravity / VSCode Style) -->
  <div
    role="button"
    tabindex="0"
    onclick={toggleExpand}
    onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); isExpanded = !isExpanded; } }}
    class="flex items-center justify-between px-2.5 py-1.5 cursor-pointer select-none hover:bg-ant-bg-secondary/80 transition-colors"
  >
    <!-- Left: Chevron, Icon, Clean Verb & Target -->
    <div class="flex items-center space-x-2 min-w-0 flex-1 mr-2 font-mono">
      <span class="text-ant-text-muted transition-transform duration-150 flex-shrink-0">
        {#if isExpanded}
          <ChevronDown size={13} />
        {:else}
          <ChevronRight size={13} />
        {/if}
      </span>

      <!-- Verb Badge -->
      <span class="font-semibold text-ant-text tracking-wide flex items-center space-x-1 flex-shrink-0">
        {#if toolParsed.type === 'read'}
          <FileCode size={13} class="text-blue-400" />
        {:else if toolParsed.type === 'edit'}
          <Replace size={13} class="text-amber-400" />
        {:else if toolParsed.type === 'write'}
          <FileEdit size={13} class="text-emerald-400" />
        {:else if toolParsed.type === 'search'}
          <Search size={13} class="text-amber-400" />
        {:else if toolParsed.type === 'list'}
          <Folder size={13} class="text-indigo-400" />
        {:else if toolParsed.type === 'terminal'}
          <Terminal size={13} class="text-indigo-400" />
        {:else}
          <FileText size={13} class="text-ant-primary" />
        {/if}
        <span class="ml-1 text-white">{toolParsed.verb}</span>
      </span>

      <!-- Clean Single-Line Target -->
      <span class="text-[11px] text-ant-text-secondary truncate max-w-xl">
        {toolParsed.target}
      </span>
    </div>

    <!-- Right: Diff Stat, Duration & Status Indicator -->
    <div class="flex items-center space-x-2 flex-shrink-0">
      {#if diffStat && (diffStat.added > 0 || diffStat.removed > 0)}
        <div class="flex items-center space-x-1 text-[10px] font-mono">
          {#if diffStat.added > 0}
            <span class="text-emerald-400 font-medium bg-emerald-500/10 px-1 rounded">+{diffStat.added}</span>
          {/if}
          {#if diffStat.removed > 0}
            <span class="text-rose-400 font-medium bg-rose-500/10 px-1 rounded">−{diffStat.removed}</span>
          {/if}
        </div>
      {/if}

      {#if durationMs !== null}
        <span class="text-[10px] font-mono text-ant-text-muted">
          {durationMs}ms
        </span>
      {/if}

      {#if toolCall.status === 'running'}
        <span class="flex items-center text-[10px] font-medium text-ant-primary animate-pulse">
          <Loader2 size={11} class="animate-spin mr-1" />
        </span>
      {:else if toolCall.status === 'completed'}
        <CheckCircle2 size={12} class="text-ant-success/80" />
      {:else if toolCall.status === 'error'}
        <XCircle size={12} class="text-ant-error" />
      {/if}
    </div>
  </div>

  <!-- Collapsible Details Container (Zero Redundancy) -->
  {#if isExpanded}
    <div class="px-2.5 pb-2 pt-1 bg-ant-bg border-t border-ant-border-secondary/60 space-y-1.5 select-text font-mono">
      <!-- 1. Dedicated Diff Card for Edit -->
      {#if toolCall.diff}
        <div class="pt-0.5">
          <DiffCard diff={toolCall.diff} />
        </div>
      {:else if toolParsed.type === 'edit' && (toolParsed.oldStr || toolParsed.newStr)}
        <!-- Inline Diff Fallback -->
        <div class="p-2 rounded bg-ant-bg-secondary/70 border border-ant-border-secondary/60 text-[11px] space-y-0.5">
          {#if toolParsed.oldStr}
            <div class="text-rose-400 break-all flex items-start gap-1">
              <span class="select-none font-bold min-w-[10px] text-rose-500">-</span>
              <span class="bg-rose-500/10 px-1 py-0.5 rounded flex-1">{toolParsed.oldStr}</span>
            </div>
          {/if}
          {#if toolParsed.newStr}
            <div class="text-emerald-400 break-all flex items-start gap-1">
              <span class="select-none font-bold min-w-[10px] text-emerald-500">+</span>
              <span class="bg-emerald-500/10 px-1 py-0.5 rounded flex-1">{toolParsed.newStr}</span>
            </div>
          {/if}
        </div>
      {:else if toolParsed.type === 'read' && readLines.length > 0}
        <!-- 2. Clean Numbered Code Reader for Read File -->
        <div class="rounded border border-ant-border-secondary/70 bg-ant-bg-secondary/30 overflow-hidden text-[11px]">
          <div class="flex items-center justify-between px-2.5 py-1 bg-ant-bg-tertiary/50 border-b border-ant-border-secondary/60 text-[10px] select-none text-ant-text-muted">
            <span class="font-mono text-ant-text-secondary">
              {readLines.length} {readLines.length === 1 ? 'line' : 'lines'}
            </span>
            <button
              type="button"
              onclick={copyOutputText}
              class="flex items-center space-x-1 text-ant-text-muted hover:text-white transition px-1 py-0.5 rounded hover:bg-ant-bg"
              title="Copy file excerpt"
            >
              {#if copiedOutput}
                <Check size={10} class="text-ant-success" />
                <span class="text-ant-success">Copied</span>
              {:else}
                <Copy size={10} />
                <span>Copy</span>
              {/if}
            </button>
          </div>
          <div class="overflow-x-auto max-h-56 scrollbar-thin">
            <table class="w-full border-collapse font-mono text-[11px] leading-relaxed">
              <tbody>
                {#each readLines as row (row.lineNo)}
                  <tr class="hover:bg-ant-bg-secondary/40 text-ant-text">
                    <td class="w-7 px-1.5 py-0.5 text-right text-ant-text-muted/60 border-r border-ant-border-secondary/40 select-none bg-ant-bg-secondary/20 text-[10px]">
                      {row.lineNo}
                    </td>
                    <td class="px-2.5 py-0.5 whitespace-pre font-mono">
                      <span>{row.text}</span>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      {:else if toolCall.result !== undefined && (!toolCall.diff || !isGenericEditAck || toolCall.status === 'error')}
        <!-- 3. General Output Block (Terminal / Search / Write) -->
        <div class="rounded border border-ant-border-secondary/60 bg-ant-bg-secondary/50 overflow-hidden text-[11px]">
          <div class="flex items-center justify-between px-2 py-1 bg-ant-bg-tertiary/60 border-b border-ant-border-secondary/60 text-[10px] select-none text-ant-text-muted">
            <span class="uppercase tracking-wider font-semibold {toolCall.status === 'error' ? 'text-ant-error' : 'text-ant-text-secondary'}">
              {toolCall.status === 'error' ? 'Error Output' : 'Output'}
            </span>
            <button
              type="button"
              onclick={copyOutputText}
              class="flex items-center space-x-1 text-ant-text-muted hover:text-white transition px-1 py-0.5 rounded hover:bg-ant-bg"
              title="Copy output"
            >
              {#if copiedOutput}
                <Check size={10} class="text-ant-success" />
                <span class="text-ant-success">Copied</span>
              {:else}
                <Copy size={10} />
                <span>Copy</span>
              {/if}
            </button>
          </div>
          <pre class="p-2 leading-relaxed overflow-x-auto max-h-56 scrollbar-thin {toolCall.status === 'error' ? 'text-rose-300' : 'text-ant-text'}"><code>{toolCall.result || '(empty)'}</code></pre>
        </div>
      {/if}
    </div>
  {/if}
</div>
