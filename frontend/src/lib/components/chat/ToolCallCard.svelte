<script lang="ts">
  import type { ToolCall } from '$lib/stores/session.svelte';
  import { calculateDiffStat, computeStringDiff } from '$lib/utils/diffUtils';
  import { formatToolResult } from '$lib/utils/toolUtils';
  import { highlightCode } from '$lib/utils/codeHighlighter';
  import DiffCard from './DiffCard.svelte';
  import { renderMarkdown } from '$lib/utils/markdownRenderer';
  import { inputShieldStore } from '$lib/stores/inputShield.svelte';
  import {
    parseTodosUpdated,
    type TodoItem,
    extractPlanFilePath,
    extractPlanMarkdownFromResult
  } from '$lib/utils/planParser';
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
    BookOpen,
    ListTodo,
    RotateCw
  } from 'lucide-svelte';

  interface Props {
    toolCall: ToolCall;
  }

  let { toolCall }: Props = $props();

  let isExpanded = $state(false);
  let copiedOutput = $state(false);
  let copiedPlan = $state(false);
  let planMarkdownContent = $state<string>('');
  let isPlanLoading = $state(false);
  let planLoadAttempted = $state(false);
  let planLoadError = $state<string | null>(null);
  let lastLoadedPath = $state<string>('');
  let dirSearchQuery = $state('');
  let dirFilterTab = $state<'all' | 'folders' | 'files'>('all');
  let dirShowAll = $state(false);

  // Load plan content from disk via Go backend
  async function loadPlanContent(force = false) {
    const planPath = toolParsed.planPath;
    if (!planPath) {
      planLoadAttempted = true;
      return;
    }
    if (!force && planLoadAttempted && lastLoadedPath === planPath && (planMarkdownContent || planLoadError)) {
      return;
    }
    if (isPlanLoading) return;

    isPlanLoading = true;
    planLoadError = null;
    lastLoadedPath = planPath;

    try {
      if (window.go?.main?.App?.GetPlanContent) {
        const content = await window.go.main.App.GetPlanContent(planPath);
        planMarkdownContent = content || '';
      } else {
        planMarkdownContent = '';
      }
    } catch (err: unknown) {
      const errMsg = err instanceof Error ? err.message : String(err);
      planLoadError = errMsg;
      planMarkdownContent = '';
    } finally {
      isPlanLoading = false;
      planLoadAttempted = true;
    }
  }

  async function copyPlanMarkdown() {
    const text = displayPlanMarkdown || planMarkdownContent || toolParsed.reason || '';
    if (!text) return;
    try {
      await navigator.clipboard.writeText(text);
      copiedPlan = true;
      setTimeout(() => (copiedPlan = false), 2000);
    } catch {
      // ignore
    }
  }

  // Fetch plan file content when plan card is expanded or target path changes
  $effect(() => {
    if (isExpanded && (toolParsed.type === 'plan_enter' || toolParsed.type === 'plan_exit')) {
      const planPath = toolParsed.planPath;
      if (planPath && (!planLoadAttempted || lastLoadedPath !== planPath)) {
        loadPlanContent();
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

  // Detect TodosUpdated in params or result
  const detectedTodos = $derived.by<TodoItem[] | null>(() => {
    // Check result first
    const fromResult = parseTodosUpdated(toolCall.result);
    if (fromResult && fromResult.length > 0) return fromResult;
    // Check params
    const fromParams = parseTodosUpdated(toolCall.params);
    if (fromParams && fromParams.length > 0) return fromParams;
    return null;
  });

  // Extract tool category / verb metadata matching grok-build-vscode
  const toolParsed = $derived.by(() => {
    const rawTool = (toolCall.tool || '').toLowerCase();
    const p = (typeof toolCall.params === 'object' && toolCall.params !== null)
      ? (toolCall.params as Record<string, unknown>)
      : {};

    // 0. Todos / Plan Update Tool (e.g. todo_write or write/task output containing TodosUpdated)
    if (detectedTodos && detectedTodos.length > 0) {
      const completedCount = detectedTodos.filter(t => t.status === 'completed').length;
      const targetText = `${completedCount}/${detectedTodos.length} tasks completed`;
      return {
        verb: 'Update Plan',
        type: 'todos_update' as const,
        target: targetText,
        todos: detectedTodos,
        icon: ListTodo,
        iconClass: 'text-indigo-400'
      };
    }

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
      const planPath = extractPlanFilePath(p, toolCall.result);
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
      const planPath = extractPlanFilePath(p, toolCall.result);
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

    // 10. Git commands / Git plugins
    if (rawTool.includes('git_') || rawTool.includes('git-') || rawTool.includes(':git_') || rawTool.includes(':git-') || rawTool.startsWith('git')) {
      const cleanVerb = toolCall.tool?.replace(/^.*[:_]/, 'git ') || 'Git';
      return {
        verb: 'Git',
        type: 'custom' as const,
        target: toolCall.tool || 'git action',
        icon: Terminal,
        iconClass: 'text-amber-500 dark:text-amber-400'
      };
    }

    // Fallback Tool
    // Format custom/namespaced tools like "git_traffic_cop:git_status" or "search_tool" cleanly
    const displayVerb = toolCall.tool
      ? toolCall.tool.split(':').pop()?.replace(/_/g, ' ') || toolCall.tool
      : 'Tool';

    return {
      verb: displayVerb.charAt(0).toUpperCase() + displayVerb.slice(1),
      type: 'custom' as const,
      target: typeof toolCall.params === 'string'
        ? toolCall.params
        : (Object.keys(p).length > 0 ? JSON.stringify(p) : toolCall.tool || ''),
      icon: FileText,
      iconClass: 'text-ant-primary'
    };
  });

  // Calculate comprehensive diff stat using calculateDiffStat utility
  const diffStat = $derived.by(() => {
    return calculateDiffStat(toolCall);
  });

  // Extract fallback markdown content directly from tool result if available (e.g. exit_plan_mode output)
  const extractedResultMarkdown = $derived.by(() => {
    return extractPlanMarkdownFromResult(toolCall.result);
  });

  // Effective markdown to display: disk plan.md content takes priority, then result markdown
  const displayPlanMarkdown = $derived.by(() => {
    if (planMarkdownContent && planMarkdownContent.trim().length > 0) {
      return planMarkdownContent;
    }
    if (extractedResultMarkdown && extractedResultMarkdown.trim().length > 0) {
      return extractedResultMarkdown;
    }
    return '';
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
    if (detectedTodos && detectedTodos.length > 0) return true;
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
  const dirListingData = $derived.by(() => {
    if (toolParsed.type !== 'list' || !toolCall.result) {
      return { folders: [], files: [], all: [] };
    }
    const raw = typeof toolCall.result === 'string' ? toolCall.result : '';
    const lines = raw.split('\n').map(l => l.trim()).filter(Boolean);

    const targetDir = toolParsed.target?.replace(/^\.\/?/, '').replace(/\/$/, '') || '';

    const items = lines.map(line => {
      // Strip bullet "- " or "• "
      let clean = line.replace(/^[-*•]\s+/, '').trim();
      
      // If path starts with target directory, show relative name
      let displayName = clean;
      if (targetDir && displayName.startsWith(targetDir + '/')) {
        displayName = displayName.slice(targetDir.length + 1);
      }
      
      const isDir = clean.endsWith('/') || clean.endsWith('\\');
      const ext = !isDir && displayName.includes('.') ? displayName.split('.').pop()?.toLowerCase() || '' : '';
      
      return {
        fullName: clean,
        name: displayName,
        isDir,
        ext
      };
    });

    const folders = items.filter(i => i.isDir).sort((a, b) => a.name.localeCompare(b.name));
    const files = items.filter(i => !i.isDir).sort((a, b) => a.name.localeCompare(b.name));

    return {
      folders,
      files,
      all: [...folders, ...files]
    };
  });

  const filteredDirItems = $derived.by(() => {
    let source = dirListingData.all;
    if (dirFilterTab === 'folders') source = dirListingData.folders;
    if (dirFilterTab === 'files') source = dirListingData.files;

    const q = dirSearchQuery.trim().toLowerCase();
    if (!q) return source;

    return source.filter(i => i.name.toLowerCase().includes(q) || i.fullName.toLowerCase().includes(q));
  });

  const displayDirList = $derived.by(() => {
    return dirShowAll ? filteredDirItems : filteredDirItems.slice(0, 18);
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

<div
  class="my-1 rounded-md border border-white/5 bg-ant-bg-secondary/20 overflow-hidden text-xs transition-colors hover:border-white/10"
  style="content-visibility: auto; contain-intrinsic-size: auto 34px;"
>
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
            <span class="text-emerald-700 dark:text-emerald-400 font-semibold bg-emerald-500/15 dark:bg-emerald-500/10 px-1 rounded">+{diffStat.added}</span>
          {/if}
          {#if diffStat.removed > 0}
            <span class="text-rose-700 dark:text-rose-400 font-semibold bg-rose-500/15 dark:bg-rose-500/10 px-1 rounded">−{diffStat.removed}</span>
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
                    <tr class="hover:bg-white/[0.02] {line.type === 'add' ? 'bg-emerald-500/10 text-emerald-800 dark:text-emerald-300' : line.type === 'del' ? 'bg-rose-500/10 text-rose-800 dark:text-rose-300' : 'text-ant-text-muted'}">
                      <td class="w-6 px-1.5 py-0.5 text-center select-none font-bold {line.type === 'add' ? 'text-emerald-700 dark:text-emerald-400' : line.type === 'del' ? 'text-rose-700 dark:text-rose-400' : 'text-ant-text-muted/40'}">
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
                <div class="text-rose-800 dark:text-rose-400 break-all flex items-start gap-1">
                  <span class="select-none font-bold min-w-[10px] text-rose-700 dark:text-rose-500">−</span>
                  <span class="bg-rose-500/10 px-1 py-0.5 rounded flex-1 whitespace-pre-wrap">{toolParsed.oldStr}</span>
                </div>
              {/if}
              {#if toolParsed.newStr}
                <div class="text-emerald-800 dark:text-emerald-400 break-all flex items-start gap-1">
                  <span class="select-none font-bold min-w-[10px] text-emerald-700 dark:text-emerald-500">+</span>
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
      {:else if toolParsed.type === 'list' && dirListingData.all.length > 0}
        <!-- 2c. Clean Directory Listing Grid with Categorization & Search Filter -->
        <div class="rounded-lg border border-ant-border-secondary dark:border-white/5 bg-ant-bg-secondary/40 overflow-hidden text-[11px]">
          <!-- Header Bar: Stats, Filter Tabs & Search -->
          <div class="flex flex-wrap items-center justify-between gap-2 px-3 py-2 bg-ant-bg-tertiary/60 border-b border-ant-border-secondary dark:border-white/5 text-[10px] select-none text-ant-text-muted">
            <!-- Left: Counts & Filter Pills -->
            <div class="flex items-center space-x-1.5 font-mono">
              <span class="text-blue-700 dark:text-blue-400 font-semibold flex items-center gap-1">
                <Folder size={12} class="text-blue-500" />
                <span>{dirListingData.all.length} items</span>
              </span>

              <div class="flex items-center space-x-0.5 ml-2 bg-ant-bg p-0.5 rounded-md border border-ant-border-secondary dark:border-white/5 text-[10px]">
                <button
                  type="button"
                  onclick={() => dirFilterTab = 'all'}
                  class="px-2 py-0.5 rounded transition {dirFilterTab === 'all' ? 'bg-ant-primary/15 text-ant-primary font-bold' : 'text-ant-text-secondary hover:text-ant-text'}"
                >
                  All ({dirListingData.all.length})
                </button>
                <button
                  type="button"
                  onclick={() => dirFilterTab = 'folders'}
                  class="px-2 py-0.5 rounded transition flex items-center gap-1 {dirFilterTab === 'folders' ? 'bg-ant-primary/15 text-ant-primary font-bold' : 'text-ant-text-secondary hover:text-ant-text'}"
                >
                  <Folder size={10} />
                  <span>Folders ({dirListingData.folders.length})</span>
                </button>
                <button
                  type="button"
                  onclick={() => dirFilterTab = 'files'}
                  class="px-2 py-0.5 rounded transition flex items-center gap-1 {dirFilterTab === 'files' ? 'bg-ant-primary/15 text-ant-primary font-bold' : 'text-ant-text-secondary hover:text-ant-text'}"
                >
                  <FileCode size={10} />
                  <span>Files ({dirListingData.files.length})</span>
                </button>
              </div>
            </div>

            <!-- Right: Search Input -->
            <div class="relative w-40">
              <Search size={11} class="absolute left-2 top-1/2 -translate-y-1/2 text-ant-text-muted" />
              <input
                type="text"
                placeholder="Filter items..."
                bind:value={dirSearchQuery}
                readonly={inputShieldStore.isReadOnly}
                class="w-full pl-6 pr-2 py-0.5 text-[10.5px] bg-ant-bg text-ant-text rounded border border-ant-border-secondary dark:border-white/5 outline-none focus:border-ant-primary transition font-sans"
              />
            </div>
          </div>

          <!-- Items Grid -->
          {#if filteredDirItems.length === 0}
            <div class="p-4 text-center text-ant-text-muted text-[11px] italic">
              No matching files or folders found
            </div>
          {:else}
            <div class="p-2 overflow-x-auto max-h-64 scrollbar-thin grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-1.5">
              {#each displayDirList as item}
                <div class="flex items-center space-x-2 px-2.5 py-1.5 rounded-md bg-ant-bg hover:bg-ant-bg-secondary text-[11.5px] font-mono truncate text-ant-text border border-ant-border-secondary dark:border-white/5 shadow-2xs group transition-colors">
                  {#if item.isDir}
                    <div class="p-1 rounded bg-blue-500/10 text-blue-700 dark:text-blue-400 flex-shrink-0">
                      <Folder size={12} />
                    </div>
                  {:else}
                    <div class="p-1 rounded bg-ant-bg-tertiary text-ant-text-secondary group-hover:text-ant-text flex-shrink-0">
                      <FileCode size={12} />
                    </div>
                  {/if}
                  <div class="flex-1 min-w-0">
                    <span class="truncate block {item.isDir ? 'font-semibold text-blue-800 dark:text-blue-300' : 'text-ant-text'}" title={item.fullName}>
                      {item.name}
                    </span>
                  </div>
                  {#if item.isDir}
                    <span class="text-[9px] uppercase tracking-wider text-blue-700 dark:text-blue-400/80 bg-blue-500/10 px-1 py-0.2 rounded font-semibold flex-shrink-0">dir</span>
                  {:else if item.ext}
                    <span class="text-[9px] uppercase tracking-wider text-ant-text-muted bg-ant-bg-tertiary px-1 py-0.2 rounded flex-shrink-0 font-medium">.{item.ext}</span>
                  {/if}
                </div>
              {/each}
            </div>

            <!-- Expand / Collapse Bar if count > 18 -->
            {#if filteredDirItems.length > 18}
              <div class="px-3 py-1.5 bg-ant-bg border-t border-ant-border-secondary dark:border-white/5 flex items-center justify-between text-[10.5px]">
                <span class="text-ant-text-muted">
                  Showing {displayDirList.length} of {filteredDirItems.length} items
                </span>
                <button
                  type="button"
                  onclick={() => dirShowAll = !dirShowAll}
                  class="text-ant-primary hover:text-ant-primary-hover font-medium cursor-pointer"
                >
                  {dirShowAll ? 'Show less (first 18)' : `Show all ${filteredDirItems.length} items`}
                </button>
              </div>
            {/if}
          {/if}
        </div>
      {:else if detectedTodos && detectedTodos.length > 0}
        <!-- 2e. Structured Plan / TodosUpdated List View -->
        <div class="rounded-lg border border-indigo-500/20 bg-indigo-950/10 overflow-hidden text-xs">
          <div class="flex items-center justify-between px-3 py-1.5 bg-indigo-900/20 border-b border-indigo-500/20 text-[11px] select-none text-indigo-300">
            <div class="flex items-center space-x-1.5 font-medium">
              <ListTodo size={13} class="text-indigo-400" />
              <span>Execution Plan Tasks</span>
            </div>
            <span class="font-mono text-[10px] px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 font-semibold">
              {detectedTodos.filter(t => t.status === 'completed').length} / {detectedTodos.length} Completed
            </span>
          </div>

          <div class="p-2 space-y-1.5 max-h-72 overflow-y-auto scrollbar-thin">
            {#each detectedTodos as todo (todo.id)}
              <div class="flex items-start space-x-2.5 p-2 rounded-md bg-white/[0.02] border border-white/[0.04] transition-colors hover:bg-white/[0.04]">
                <div class="mt-0.5 shrink-0">
                  {#if todo.status === 'completed'}
                    <div class="w-4 h-4 rounded-full bg-emerald-500/20 text-emerald-400 flex items-center justify-center border border-emerald-500/40">
                      <Check size={10} class="stroke-[3]" />
                    </div>
                  {:else if todo.status === 'in_progress'}
                    <div class="w-4 h-4 rounded-full bg-amber-500/20 text-amber-400 flex items-center justify-center border border-amber-500/40 animate-pulse">
                      <Loader2 size={10} class="animate-spin" />
                    </div>
                  {:else}
                    <div class="w-4 h-4 rounded-full border border-zinc-600 bg-zinc-800/50 flex items-center justify-center">
                    </div>
                  {/if}
                </div>

                <div class="flex-1 min-w-0 font-serif leading-relaxed">
                  <p class="text-xs {todo.status === 'completed' ? 'line-through text-zinc-400' : 'text-zinc-200'}">
                    {todo.content}
                  </p>
                </div>

                {#if todo.priority}
                  <span class="text-[9px] uppercase tracking-wider font-mono px-1.5 py-0.2 rounded border {
                    todo.priority === 'high' ? 'bg-rose-500/10 text-rose-300 border-rose-500/20' :
                    todo.priority === 'medium' ? 'bg-amber-500/10 text-amber-300 border-amber-500/20' :
                    'bg-blue-500/10 text-blue-300 border-blue-500/20'
                  }">
                    {todo.priority}
                  </span>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      {:else if (toolParsed.type === 'plan_enter' || toolParsed.type === 'plan_exit') && (displayPlanMarkdown || toolParsed.reason || toolCall.result || toolParsed.planPath)}
        <!-- 2d. Rich Markdown Plan Document Viewer -->
        <div class="rounded-lg border border-violet-500/20 bg-violet-950/10 overflow-hidden text-xs">
          <div class="flex items-center justify-between px-3 py-1.5 bg-violet-900/20 border-b border-violet-500/20 text-[11px] select-none text-violet-300">
            <div class="flex items-center space-x-1.5 font-medium">
              <BookOpen size={13} class="text-violet-400" />
              <span>{toolParsed.type === 'plan_enter' ? 'Active Engineering Plan Document' : 'Completed Execution Plan'}</span>
            </div>

            <div class="flex items-center space-x-2">
              {#if toolParsed.planPath}
                <span class="font-mono text-[10px] text-violet-300/70 max-w-xs truncate" title={toolParsed.planPath}>
                  {toolParsed.planPath}
                </span>

                <!-- Reload / Refresh Button -->
                <button
                  type="button"
                  onclick={() => loadPlanContent(true)}
                  disabled={isPlanLoading}
                  class="p-1 rounded text-violet-300/70 hover:text-violet-200 hover:bg-violet-800/30 transition disabled:opacity-50"
                  title="Reload plan file from disk"
                >
                  <RotateCw size={11} class={isPlanLoading ? 'animate-spin' : ''} />
                </button>
              {/if}

              <!-- Copy Markdown Plan Button -->
              {#if displayPlanMarkdown || toolParsed.reason}
                <button
                  type="button"
                  onclick={copyPlanMarkdown}
                  class="flex items-center space-x-1 px-1.5 py-0.5 rounded text-[10px] text-violet-300/80 hover:text-white hover:bg-violet-800/30 transition"
                  title="Copy plan markdown"
                >
                  {#if copiedPlan}
                    <Check size={10} class="text-emerald-400" />
                    <span class="text-emerald-400">Copied</span>
                  {:else}
                    <Copy size={10} />
                    <span>Copy</span>
                  {/if}
                </button>
              {/if}
            </div>
          </div>

          <div class="p-3.5 space-y-2 max-h-80 overflow-y-auto scrollbar-thin font-serif leading-relaxed text-ant-text select-text markdown-chat-body">
            {#if isPlanLoading && !displayPlanMarkdown}
              <div class="flex items-center space-x-2 py-4 text-violet-400 justify-center">
                <Loader2 size={15} class="animate-spin" />
                <span class="text-xs">Loading plan file content...</span>
              </div>
            {:else if displayPlanMarkdown}
              {@html renderMarkdown(displayPlanMarkdown)}
            {:else if toolParsed.reason}
              <div class="p-2.5 rounded bg-violet-900/10 border border-violet-500/15 space-y-2">
                <div class="text-xs text-zinc-300">
                  <p class="font-semibold text-violet-300 mb-1">Plan Objective / Context:</p>
                  <p>{toolParsed.reason}</p>
                </div>
                {#if toolParsed.planPath}
                  <div class="pt-2 border-t border-violet-500/10 flex items-center justify-between text-[11px] text-zinc-400">
                    <span class="italic">Plan file initialized. Waiting for agent to write specifications.</span>
                    <button
                      type="button"
                      onclick={() => loadPlanContent(true)}
                      class="px-2 py-0.5 rounded bg-violet-600/20 hover:bg-violet-600/30 text-violet-300 border border-violet-500/30 flex items-center gap-1 transition"
                    >
                      <RotateCw size={10} class={isPlanLoading ? 'animate-spin' : ''} />
                      <span>Check for Updates</span>
                    </button>
                  </div>
                {/if}
              </div>
            {:else if planLoadError}
              <div class="p-3 rounded bg-rose-500/10 border border-rose-500/20 text-xs text-rose-300 flex items-center justify-between">
                <span>Failed to load plan file: {planLoadError}</span>
                <button
                  type="button"
                  onclick={() => loadPlanContent(true)}
                  class="px-2 py-0.5 rounded bg-rose-500/20 hover:bg-rose-500/30 text-rose-200 border border-rose-500/40"
                >
                  Retry
                </button>
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
          <pre class="p-2 leading-relaxed overflow-x-auto max-h-64 scrollbar-thin {toolCall.status === 'error' || parsedOutput.status === 'failed' || (parsedOutput.exitCode !== undefined && parsedOutput.exitCode !== 0) ? 'text-rose-700 dark:text-rose-300' : 'text-ant-text'}"><code>{parsedOutput.output || '(empty)'}</code></pre>
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
