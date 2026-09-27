<script lang="ts">
  import type { ToolCall } from '$lib/stores/session.svelte';
  import { calculateDiffStat, computeStringDiff } from '$lib/utils/diffUtils';
  import { formatToolResult } from '$lib/utils/toolUtils';
  import { highlightCode } from '$lib/utils/codeHighlighter';
  import DiffCard from './DiffCard.svelte';
  import { renderMarkdown } from '$lib/utils/markdownRenderer';
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
    Replace,
    Bot,
    Clock,
    AlertTriangle,
    Compass,
    CheckSquare,
    BookOpen
  } from 'lucide-svelte';

  interface Props {
    toolCall: ToolCall;
  }

  let { toolCall }: Props = $props();

  let isExpanded = $state(false);
  let copiedOutput = $state(false);
  let planMarkdownContent = $state<string>('');
  let isPlanLoading = $state(false);

  // Fetch plan file content when plan card is expanded
  $effect(() => {
    if (isExpanded && (toolParsed.type === 'plan_enter' || toolParsed.type === 'plan_exit')) {
      const planPath = toolParsed.planPath;
      if (planPath && !planMarkdownContent && !isPlanLoading) {
        isPlanLoading = true;
        if (window.go?.main?.App?.GetPlanContent) {
          window.go.main.App.GetPlanContent(planPath)
            .then((content) => {
              planMarkdownContent = content;
            })
            .catch(() => {
              planMarkdownContent = '';
            })
            .finally(() => {
              isPlanLoading = false;
            });
        } else {
          isPlanLoading = false;
        }
      }
    }
  });

  // Derive formatted execution duration
  const formattedDuration = $derived.by(() => {
    if (!toolCall.startTime) return null;
    const end = toolCall.endTime || Date.now();
    const ms = Math.max(0, end - toolCall.startTime);

    if (ms < 1000) {
      return `${ms}ms`;
    }
    const secs = ms / 1000;
    if (secs < 60) {
      return `${secs >= 10 ? secs.toFixed(1) : secs.toFixed(2)}s`;
    }
    const mins = Math.floor(secs / 60);
    const remSecs = Math.round(secs % 60);
    return remSecs > 0 ? `${mins}m ${remSecs}s` : `${mins}m`;
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
      const content = String(p.content || '');
      return {
        verb: 'Write',
        type: 'write' as const,
        target: filePath || 'file',
        filePath,
        content,
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
    if (rawTool.includes('run_terminal_command') || rawTool.includes('terminal') || rawTool.includes('bash') || rawTool.includes('exec') || rawTool.includes('cmd')) {
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

    // 7. Get Command / Subagent Output
    if (rawTool.includes('get_command_or_subagent_output') || rawTool.includes('command_output') || rawTool.includes('subagent_output')) {
      let targetDesc = '';
      if (Array.isArray(p.task_ids) && p.task_ids.length > 0) {
        targetDesc = p.task_ids.join(', ');
      } else if (p.task_id) {
        targetDesc = String(p.task_id);
      }
      return {
        verb: 'Output',
        type: 'task_output' as const,
        target: targetDesc || 'task output',
        icon: Bot,
        iconClass: 'text-cyan-400'
      };
    }

    // 8. Enter Plan Mode
    if (rawTool.includes('enter_plan_mode')) {
      const reason = String(p.reason || '');
      let planPath = '';
      if (typeof toolCall.result === 'string') {
        const match = toolCall.result.match(/Plan file:\s*([^\r\n]+)/i);
        if (match) planPath = match[1].trim();
      }
      return {
        verb: 'Plan',
        type: 'plan_enter' as const,
        target: reason ? reason : (planPath ? `Plan file: ${planPath}` : 'Enter plan mode'),
        reason,
        planPath,
        icon: Compass,
        iconClass: 'text-violet-400'
      };
    }

    // 9. Exit Plan Mode
    if (rawTool.includes('exit_plan_mode')) {
      const reason = String(p.reason || '');
      let planPath = '';
      if (typeof toolCall.result === 'string') {
        const match = toolCall.result.match(/Plan file:\s*([^\r\n]+)/i);
        if (match) planPath = match[1].trim();
      }
      return {
        verb: 'Review Plan',
        type: 'plan_exit' as const,
        target: reason ? reason : (planPath ? `Plan file: ${planPath}` : 'Plan ready for review'),
        reason,
        planPath,
        icon: CheckSquare,
        iconClass: 'text-emerald-400'
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

  // Calculate comprehensive diff stat using calculateDiffStat utility
  const diffStat = $derived.by(() => {
    return calculateDiffStat(toolCall);
  });

  // Computed multi-line diff for inline edit view
  const editDiffLines = $derived.by(() => {
    if (toolParsed.type !== 'edit') return [];
    const oldStr = toolParsed.oldStr || '';
    const newStr = toolParsed.newStr || '';
    if (!oldStr && !newStr) return [];
    return computeStringDiff(oldStr, newStr).lines;
  });

  // Parse file content into numbered lines for Read preview (stripping 1→ line-number prefixes if present)
  const readLines = $derived.by(() => {
    if (toolParsed.type !== 'read' || !toolCall.result) return [];
    const lines = toolCall.result.split('\n');
    const start = ('startLine' in toolParsed && typeof toolParsed.startLine === 'number')
      ? toolParsed.startLine
      : 1;

    // Detect if the result already has anchor/line prefix format like "1→..." or "10→..."
    const anchorRegex = /^(\d+)→(.*)$/;
    const firstNonEmpty = lines.find(l => l.trim().length > 0);
    const hasAnchors = firstNonEmpty ? anchorRegex.test(firstNonEmpty) : false;

    let currentLineNo = start;

    return lines.map((rawText, idx) => {
      let lineNo = currentLineNo;
      let text = rawText;

      if (hasAnchors) {
        const match = rawText.match(anchorRegex);
        if (match) {
          lineNo = parseInt(match[1], 10);
          text = match[2];
          currentLineNo = lineNo + 1;
        } else {
          lineNo = currentLineNo;
          currentLineNo++;
        }
      } else {
        lineNo = start + idx;
      }

      // Detect language from file path extension for syntax highlighting
      const ext = toolParsed.filePath ? toolParsed.filePath.split('.').pop()?.toLowerCase() || '' : '';
      const highlightedHtml = highlightCode(text, ext);

      return {
        lineNo,
        text,
        highlightedHtml
      };
    });
  });

  // Determine whether this output is redundant noise (e.g. standard success ack for edits and writes)
  const isGenericEditAck = $derived.by(() => {
    if (toolParsed.type !== 'edit' && toolParsed.type !== 'write') return false;
    const res = (toolCall.result || '').toLowerCase();
    return (
      res.includes('successfully') ||
      res.includes('wrote file') ||
      res.includes('updated timeout') ||
      res.includes('1 replacement applied') ||
      res.includes('ok')
    );
  });

  // Parse written file content into numbered lines for Write preview
  const writeLines = $derived.by(() => {
    if (toolParsed.type !== 'write') return [];
    const content = ('content' in toolParsed && typeof toolParsed.content === 'string') ? toolParsed.content : '';
    if (!content) return [];
    const lines = content.split('\n');
    const ext = toolParsed.filePath ? toolParsed.filePath.split('.').pop()?.toLowerCase() || '' : '';

    return lines.map((text, idx) => ({
      lineNo: idx + 1,
      text,
      highlightedHtml: highlightCode(text, ext)
    }));
  });

  // Parse directory listing output into clean file/folder items
  const dirListingItems = $derived.by(() => {
    if (toolParsed.type !== 'list' || !toolCall.result) return [];
    const raw = typeof toolCall.result === 'string' ? toolCall.result : '';
    const lines = raw.split('\n').map(l => l.trim()).filter(Boolean);

    return lines.map(line => {
      // Strip bullet "- " or "• "
      const clean = line.replace(/^[-*•]\s+/, '').trim();
      const isDir = clean.endsWith('/') || clean.endsWith('\\');
      return {
        name: clean,
        isDir
      };
    });
  });

  // Parse polymorphic tool result (e.g. JSON TaskOutput from terminal/subagent tools, stripping ANSI)
  const parsedOutput = $derived.by(() => {
    return formatToolResult(toolCall.result);
  });

  async function copyOutputText() {
    let textToCopy = '';
    if (toolParsed.type === 'read' && readLines.length > 0) {
      textToCopy = readLines.map(r => r.text).join('\n');
    } else if (toolParsed.type === 'write' && writeLines.length > 0) {
      textToCopy = writeLines.map(r => r.text).join('\n');
    } else {
      textToCopy = parsedOutput.isJsonTaskOutput ? parsedOutput.output : (toolCall.result || '');
    }
    if (!textToCopy) return;
    try {
      await navigator.clipboard.writeText(textToCopy);
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

<div class="my-1 rounded-md border border-white/5 bg-ant-bg-secondary/20 overflow-hidden text-xs transition-colors hover:border-white/10">
  <!-- Minimalist Tool Header (Anti Gravity / VSCode Style) -->
  <div
    role="button"
    tabindex="0"
    onclick={toggleExpand}
    onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); isExpanded = !isExpanded; } }}
    class="flex items-center justify-between px-2.5 py-1.5 cursor-pointer select-none hover:bg-ant-bg-secondary/50 transition-colors"
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
        {:else if toolParsed.type === 'task_output'}
          <Bot size={13} class="text-cyan-400" />
        {:else if toolParsed.type === 'plan_enter'}
          <Compass size={13} class="text-violet-400" />
        {:else if toolParsed.type === 'plan_exit'}
          <CheckSquare size={13} class="text-emerald-400" />
        {:else}
          <FileText size={13} class="text-ant-primary" />
        {/if}
        <span class="ml-1 text-ant-text font-medium">{toolParsed.verb}</span>
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

      {#if formattedDuration !== null}
        <span class="text-[10px] font-mono text-ant-text-muted">
          {formattedDuration}
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
    <div class="px-2.5 pb-2 pt-1 bg-ant-bg border-t border-white/5 space-y-1.5 select-text font-mono">
      <!-- 1. Dedicated Diff Card for Edit -->
      {#if toolCall.diff}
        <div class="pt-0.5">
          <DiffCard diff={toolCall.diff} />
        </div>
      {:else if toolParsed.type === 'edit' && (toolParsed.oldStr || toolParsed.newStr)}
        <!-- Inline Diff with Line-by-Line Highlight -->
        <div class="rounded border border-white/5 bg-ant-bg-secondary/40 overflow-hidden text-[11px]">
          {#if editDiffLines.length > 0}
            <div class="overflow-x-auto max-h-64 scrollbar-thin">
              <table class="w-full border-collapse font-mono text-[11px] leading-relaxed">
                <tbody>
                  {#each editDiffLines as line, idx (idx)}
                    <tr class="hover:bg-white/[0.02] {line.type === 'add' ? 'bg-emerald-500/10 text-emerald-300' : line.type === 'del' ? 'bg-rose-500/10 text-rose-300' : 'text-ant-text-muted'}">
                      <td class="w-6 px-1.5 py-0.5 text-center select-none font-bold {line.type === 'add' ? 'text-emerald-400' : line.type === 'del' ? 'text-rose-400' : 'text-ant-text-muted/40'}">
                        {line.type === 'add' ? '+' : line.type === 'del' ? '−' : ' '}
                      </td>
                      <td class="px-2 py-0.5 whitespace-pre font-mono break-all">
                        <span>{line.content}</span>
                      </td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {:else}
            <div class="p-2 space-y-0.5">
              {#if toolParsed.oldStr}
                <div class="text-rose-400 break-all flex items-start gap-1">
                  <span class="select-none font-bold min-w-[10px] text-rose-500">−</span>
                  <span class="bg-rose-500/10 px-1 py-0.5 rounded flex-1 whitespace-pre-wrap">{toolParsed.oldStr}</span>
                </div>
              {/if}
              {#if toolParsed.newStr}
                <div class="text-emerald-400 break-all flex items-start gap-1">
                  <span class="select-none font-bold min-w-[10px] text-emerald-500">+</span>
                  <span class="bg-emerald-500/10 px-1 py-0.5 rounded flex-1 whitespace-pre-wrap">{toolParsed.newStr}</span>
                </div>
              {/if}
            </div>
          {/if}
        </div>
      {:else if toolParsed.type === 'read' && readLines.length > 0}
        <!-- 2. Clean Numbered Code Reader for Read File -->
        <div class="rounded border border-white/5 bg-ant-bg-secondary/30 overflow-hidden text-[11px]">
          <div class="flex items-center justify-between px-2.5 py-1 bg-ant-bg-tertiary/50 border-b border-white/5 text-[10px] select-none text-ant-text-muted">
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
                    <td class="w-7 px-1.5 py-0.5 text-right text-ant-text-muted/60 border-r border-white/5 select-none bg-ant-bg-secondary/20 text-[10px]">
                      {row.lineNo}
                    </td>
                    <td class="px-2.5 py-0.5 whitespace-pre font-mono">
                      <span>{@html row.highlightedHtml}</span>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      {:else if toolParsed.type === 'write' && writeLines.length > 0}
        <!-- 2b. Clean Numbered Code Reader for Write File Content -->
        <div class="rounded border border-white/5 bg-ant-bg-secondary/30 overflow-hidden text-[11px]">
          <div class="flex items-center justify-between px-2.5 py-1 bg-ant-bg-tertiary/50 border-b border-white/5 text-[10px] select-none text-ant-text-muted">
            <span class="font-mono text-emerald-400 font-medium">
              Written {writeLines.length} {writeLines.length === 1 ? 'line' : 'lines'}
            </span>
            <button
              type="button"
              onclick={copyOutputText}
              class="flex items-center space-x-1 text-ant-text-muted hover:text-white transition px-1 py-0.5 rounded hover:bg-ant-bg cursor-pointer"
              title="Copy written content"
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
                {#each writeLines as row (row.lineNo)}
                  <tr class="hover:bg-ant-bg-secondary/40 text-ant-text">
                    <td class="w-7 px-1.5 py-0.5 text-right text-emerald-400/60 border-r border-white/5 select-none bg-emerald-500/5 text-[10px]">
                      {row.lineNo}
                    </td>
                    <td class="px-2.5 py-0.5 whitespace-pre font-mono">
                      <span>{@html row.highlightedHtml}</span>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        </div>
      {:else if toolParsed.type === 'list' && dirListingItems.length > 0}
        <!-- 2c. Clean Directory Listing Grid -->
        <div class="rounded border border-white/5 bg-ant-bg-secondary/30 overflow-hidden text-[11px]">
          <div class="flex items-center justify-between px-2.5 py-1 bg-ant-bg-tertiary/50 border-b border-white/5 text-[10px] select-none text-ant-text-muted">
            <span class="font-mono text-indigo-400 font-medium">
              {dirListingItems.length} {dirListingItems.length === 1 ? 'item' : 'items'} in directory
            </span>
          </div>
          <div class="p-2 overflow-x-auto max-h-56 scrollbar-thin grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-1">
            {#each dirListingItems as item}
              <div class="flex items-center space-x-1.5 px-2 py-1 rounded bg-white/[0.02] hover:bg-white/[0.05] text-[11.5px] font-mono truncate text-ant-text border border-white/[0.03]">
                {#if item.isDir}
                  <Folder size={12} class="text-indigo-400 flex-shrink-0" />
                {:else}
                  <FileCode size={12} class="text-blue-400/80 flex-shrink-0" />
                {/if}
                <span class="truncate" title={item.name}>{item.name}</span>
              </div>
            {/each}
          </div>
        </div>
      {:else if (toolParsed.type === 'plan_enter' || toolParsed.type === 'plan_exit') && (planMarkdownContent || toolParsed.reason || toolCall.result)}
        <!-- 2d. Rich Markdown Plan Document Viewer -->
        <div class="rounded-lg border border-violet-500/20 bg-violet-950/10 overflow-hidden text-xs">
          <div class="flex items-center justify-between px-3 py-1.5 bg-violet-900/20 border-b border-violet-500/20 text-[11px] select-none text-violet-300">
            <div class="flex items-center space-x-1.5 font-medium">
              <BookOpen size={13} class="text-violet-400" />
              <span>{toolParsed.type === 'plan_enter' ? 'Active Engineering Plan Document' : 'Completed Execution Plan'}</span>
            </div>
            {#if toolParsed.planPath}
              <span class="font-mono text-[10px] text-violet-300/70 max-w-xs truncate" title={toolParsed.planPath}>
                {toolParsed.planPath}
              </span>
            {/if}
          </div>
          <div class="p-3.5 space-y-2 max-h-80 overflow-y-auto scrollbar-thin font-serif leading-relaxed text-ant-text select-text markdown-chat-body">
            {#if isPlanLoading}
              <div class="flex items-center space-x-2 py-4 text-violet-400 justify-center">
                <Loader2 size={15} class="animate-spin" />
                <span class="text-xs">Loading plan file content...</span>
              </div>
            {:else if planMarkdownContent}
              {@html renderMarkdown(planMarkdownContent)}
            {:else if toolParsed.reason}
              <div class="text-xs text-zinc-300">
                <p class="font-semibold text-violet-300 mb-1">Plan Summary:</p>
                <p>{toolParsed.reason}</p>
              </div>
            {:else}
              <div class="text-xs text-zinc-400 italic">
                No markdown plan file available.
              </div>
            {/if}
          </div>
        </div>
      {:else if toolCall.result !== undefined && (!toolCall.diff || !isGenericEditAck || toolCall.status === 'error')}
        <!-- 3. General Output Block (Terminal / TaskOutput / Search / Write) -->
        <div class="rounded border border-white/5 bg-ant-bg-secondary/50 overflow-hidden text-[11px]">
          <div class="flex items-center justify-between px-2 py-1 bg-ant-bg-tertiary/60 border-b border-white/5 text-[10px] select-none text-ant-text-muted">
            <div class="flex items-center space-x-2">
              <span class="uppercase tracking-wider font-semibold {toolCall.status === 'error' || parsedOutput.status === 'failed' || (parsedOutput.exitCode !== undefined && parsedOutput.exitCode !== 0) ? 'text-ant-error' : 'text-ant-text-secondary'}">
                {toolCall.status === 'error' || parsedOutput.status === 'failed' || (parsedOutput.exitCode !== undefined && parsedOutput.exitCode !== 0) ? 'Execution Failed' : 'Output'}
              </span>
              {#if parsedOutput.isJsonTaskOutput}
                {#if parsedOutput.command}
                  <span class="font-mono text-[10.5px] text-zinc-400 bg-white/5 px-1.5 py-0.5 rounded">
                    $ {parsedOutput.command}
                  </span>
                {/if}
                {#if parsedOutput.exitCode !== undefined}
                  <span class="font-mono text-[10px] px-1 py-0.2 rounded {parsedOutput.exitCode === 0 ? 'bg-emerald-500/10 text-emerald-400' : 'bg-rose-500/10 text-rose-400'}">
                    exit: {parsedOutput.exitCode}
                  </span>
                {/if}
                {#if parsedOutput.durationSecs !== undefined}
                  <span class="text-[10px] font-mono text-zinc-400">
                    {parsedOutput.durationSecs.toFixed(1)}s
                  </span>
                {/if}
              {/if}
            </div>
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
          <pre class="p-2 leading-relaxed overflow-x-auto max-h-64 scrollbar-thin {toolCall.status === 'error' || parsedOutput.status === 'failed' || (parsedOutput.exitCode !== undefined && parsedOutput.exitCode !== 0) ? 'text-rose-300' : 'text-ant-text'}"><code>{parsedOutput.output || '(empty)'}</code></pre>
          {#if parsedOutput.truncated && parsedOutput.outputFile}
            <div class="px-2 py-1 bg-ant-bg-tertiary/40 border-t border-white/5 text-[10px] text-ant-text-muted flex items-center justify-between">
              <span>Output truncated</span>
              <span class="font-mono truncate max-w-xs text-ant-text-secondary" title={parsedOutput.outputFile}>{parsedOutput.outputFile}</span>
            </div>
          {/if}
        </div>
      {/if}
    </div>
  {/if}
</div>
