<script lang="ts">
  import type { ToolCall } from '$lib/stores/session.svelte';
  import { calculateDiffStat, computeStringDiff } from '$lib/utils/diffUtils';
  import {
    formatToolResult,
    isSubagentToolCall,
    extractSubagentMeta,
    extractSubagentIdFromText,
    formatSubagentDuration,
    isSubagentBackgroundAck,
    cleanSubagentOutput,
    formatActionSummary
  } from '$lib/utils/toolUtils';
  import { highlightCode } from '$lib/utils/codeHighlighter';
  import DiffCard from './DiffCard.svelte';
  import { renderMarkdown } from '$lib/utils/markdownRenderer';
  import { inputShieldStore } from '$lib/stores/inputShield.svelte';
  import { planStore } from '$lib/stores/plan.svelte';
  import { sessionStore } from '$lib/stores/session.svelte';
  import { dialogStore } from '$lib/stores/dialog.svelte';
  import {
    parseTodosUpdated,
    type TodoItem,
    extractPlanFilePath,
    extractPlanMarkdownFromResult,
    formatPlanUpdateSummary,
    resolveTodoContent,
    renderTodoContent
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
    RotateCw,
    Cpu,
    Shield,
    Maximize2,
    Activity
  } from 'lucide-svelte';

  interface Props {
    toolCall: ToolCall;
  }

  let { toolCall }: Props = $props();

  let isExpanded = $state(false);
  let copiedOutput = $state(false);
  let copiedCommand = $state(false);
  let copiedPlan = $state(false);
  let planMarkdownContent = $state<string>('');
  let isPlanLoading = $state(false);
  let planLoadAttempted = $state(false);
  let planLoadError = $state<string | null>(null);
  let lastLoadedPath = $state<string>('');
  let dirSearchQuery = $state('');
  let dirFilterTab = $state<'all' | 'folders' | 'files'>('all');
  let dirShowAll = $state(false);
  let copiedSubagentId = $state(false);
  let now = $state(Date.now());
  let actionsContainerEl = $state<HTMLElement | null>(null);
  let userHasScrolledActions = $state(false);
  let prevChildCount = $state(0);

  function handleActionsScroll() {
    if (!actionsContainerEl) return;
    const { scrollTop, scrollHeight, clientHeight } = actionsContainerEl;
    const distanceFromBottom = scrollHeight - (scrollTop + clientHeight);
    userHasScrolledActions = distanceFromBottom > 35;
  }

  function getChildDuration(child: ToolCall): string | null {
    if (!child.startTime) return null;
    const end = child.endTime || (child.status === 'running' ? now : child.startTime);
    const ms = Math.max(0, end - child.startTime);
    if (ms < 1000) return `${ms}ms`;
    const secs = ms / 1000;
    if (secs < 60) return `${secs.toFixed(1)}s`;
    const mins = Math.floor(secs / 60);
    const rem = Math.round(secs % 60);
    return `${mins}m ${rem}s`;
  }

  // Auto-scroll task actions list to bottom when expanded or when new actions arrive
  $effect(() => {
    const isSubagent = toolParsed.type === 'subagent';
    const currentCount = toolCall.childToolCalls?.length || 0;

    if (!isExpanded || !isSubagent) {
      userHasScrolledActions = false;
      prevChildCount = 0;
      return;
    }

    if (currentCount > 0) {
      const isNewItem = currentCount !== prevChildCount;
      prevChildCount = currentCount;

      if (!userHasScrolledActions || subagentIsActuallyRunning) {
        setTimeout(() => {
          if (actionsContainerEl) {
            actionsContainerEl.scrollTo({
              top: actionsContainerEl.scrollHeight,
              behavior: isNewItem && prevChildCount > 1 ? 'smooth' : 'auto'
            });
          }
        }, 30);
      }
    }
  });

  // Automatically load child tool calls and subagent execution details for subagent cards
  $effect(() => {
    if (toolParsed.type !== 'subagent') return;
    const sid = toolCall.childSessionId || toolParsed.meta?.taskId || extractSubagentIdFromText(toolCall.result);
    if (!sid || typeof window === 'undefined') return;

    if (!toolCall.childSessionId) {
      toolCall.childSessionId = sid;
    }

    // If subagent is already completed and has full output and child calls, skip
    const isCompletedWithOutput = toolCall.status === 'completed' &&
      !isSubagentBackgroundAck(toolCall.result) &&
      cleanSubagentOutput(toolCall.result) !== '' &&
      toolCall.childToolCalls && toolCall.childToolCalls.length > 0;
    if (isCompletedWithOutput) return;

    const activeSession = sessionStore.sessions.find((s) => s.id === sessionStore.activeSessionId);
    let wsPath = '';
    if (activeSession?.workspaceId) {
      const ws = sessionStore.workspaces.find((w) => w.id === activeSession.workspaceId);
      if (ws?.path) wsPath = ws.path;
    }
    if (!wsPath && sessionStore.activeWorkspace?.path) {
      wsPath = sessionStore.activeWorkspace.path;
    }

    if (window.go?.main?.App?.LoadSubagentDetails) {
      window.go.main.App.LoadSubagentDetails(wsPath, sid)
        .then((details) => {
          if (details) {
            if (details.calls && Array.isArray(details.calls) && details.calls.length > 0) {
              toolCall.childToolCalls = details.calls.map((c: any) => ({
                id: c.id,
                tool: c.tool,
                params: c.params,
                result: c.result,
                status: c.status === 'failed' ? 'error' : c.status === 'completed' ? 'completed' : 'running',
                startTime: c.startTime,
                endTime: c.endTime
              }));
            }
            if (details.status === 'completed' || details.status === 'failed') {
              toolCall.status = details.status === 'failed' ? 'error' : 'completed';
              if (details.output && (!toolCall.result || isSubagentBackgroundAck(toolCall.result))) {
                toolCall.result = details.output;
              }
              if (details.startTime && !toolCall.startTime) {
                toolCall.startTime = details.startTime;
              }
              if (details.endTime && !toolCall.endTime) {
                toolCall.endTime = details.endTime;
              } else if (details.durationMs && details.durationMs > 0 && toolCall.startTime) {
                toolCall.endTime = toolCall.startTime + details.durationMs;
              }
            }
            if (typeof toolCall.params === 'string') {
              try {
                toolCall.params = JSON.parse(toolCall.params);
              } catch {
                toolCall.params = { raw: toolCall.params };
              }
            }
            if (!toolCall.params || typeof toolCall.params !== 'object') {
              toolCall.params = {};
            }
            const tcP = toolCall.params as Record<string, any>;
            if (details.description && !tcP.description) {
              tcP.description = details.description;
            }
            if (details.prompt && !tcP.prompt) {
              tcP.prompt = details.prompt;
            }
            if (details.subagentType && !tcP.subagent_type) {
              tcP.subagent_type = details.subagentType;
            }
            if (details.model && !tcP.model) {
              tcP.model = details.model;
            }
            if (details.subagentId && !toolCall.childSessionId) {
              toolCall.childSessionId = details.subagentId;
            }
            sessionStore.syncToolCallUpdate(toolCall);
          }
        })
        .catch(() => {});
    } else if (window.go?.main?.App?.LoadSubagentToolCalls) {
      window.go.main.App.LoadSubagentToolCalls(wsPath, sid)
        .then((calls) => {
          if (calls && Array.isArray(calls) && calls.length > 0) {
            toolCall.childToolCalls = calls.map((c: any) => ({
              id: c.id,
              tool: c.tool,
              params: c.params,
              result: c.result,
              status: c.status === 'failed' ? 'error' : c.status === 'completed' ? 'completed' : 'running',
              startTime: c.startTime,
              endTime: c.endTime
            }));
            sessionStore.syncToolCallUpdate(toolCall);
          }
        })
        .catch(() => {});
    }
  });

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

  // Clean output and running state for subagent
  const subagentCleanOutput = $derived.by(() => {
    if (toolParsed.type !== 'subagent') return '';
    return cleanSubagentOutput(toolCall.result);
  });

  const subagentIsActuallyRunning = $derived.by(() => {
    if (toolParsed.type !== 'subagent') return false;
    if (toolCall.status === 'error') return false;
    if (subagentCleanOutput) return false;
    return toolCall.status === 'running' || isSubagentBackgroundAck(toolCall.result) || (toolCall.childToolCalls?.some(c => c.status === 'running') ?? false);
  });

  // Active ticking timer for running subagents and tool calls
  $effect(() => {
    const isRunning = subagentIsActuallyRunning || toolCall.status === 'running' || (toolCall.childToolCalls?.some(c => c.status === 'running') ?? false);
    if (!isRunning) return;
    const interval = setInterval(() => {
      now = Date.now();
    }, 1000);
    return () => clearInterval(interval);
  });

  // Derive formatted execution duration
  const formattedDuration = $derived.by(() => {
    if (!toolCall.startTime) return null;
    if (toolParsed.type === 'subagent') {
      let end: number;
      if (subagentIsActuallyRunning) {
        end = now;
      } else {
        const lastChildEnd = (toolCall.childToolCalls && toolCall.childToolCalls.length > 0)
          ? Math.max(0, ...toolCall.childToolCalls.map(c => c.endTime || c.startTime || 0))
          : 0;
        end = Math.max(toolCall.endTime || 0, lastChildEnd) || now;
      }
      return formatSubagentDuration(Math.max(0, end - toolCall.startTime));
    }
    const end = (toolCall.status === 'running' ? now : toolCall.endTime) || Date.now();
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
      let isMerge = true;
      if (p && typeof p.merge === 'boolean') {
        isMerge = p.merge;
      }
      const sessionPlan = planStore.getPlan(sessionStore.activeSessionId);
      const targetText = formatPlanUpdateSummary(detectedTodos, isMerge, sessionPlan);
      const verb = isMerge ? 'Update Plan' : 'Create Plan';
      return {
        verb,
        type: 'todos_update' as const,
        target: targetText,
        todos: detectedTodos,
        icon: ListTodo,
        iconClass: 'text-indigo-400'
      };
    }

    // Subagent Autonomous Delegation
    if (isSubagentToolCall(rawTool, toolCall.params || p)) {
      const subagentMeta = extractSubagentMeta(toolCall.params || p, toolCall.result);
      return {
        verb: 'Subagent',
        type: 'subagent' as const,
        target: subagentMeta.description,
        meta: subagentMeta,
        icon: Bot,
        iconClass: 'text-violet-400'
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
      const desc = String(p.description || p.desc || p.summary || p.reason || '').trim();
      return {
        verb: 'Run',
        type: 'terminal' as const,
        target: desc || (cmd ? `$ ${cmd}` : 'command'),
        description: desc,
        cmd,
        icon: Terminal,
        iconClass: 'text-indigo-400'
      };
    }

    // 7. Get Command / Subagent Output
    if (rawTool.includes('get_command_or_subagent_output') || rawTool.includes('command_output') || rawTool.includes('subagent_output')) {
      const desc = String(p.description || p.desc || p.summary || p.reason || '').trim();
      let targetDesc = '';
      if (Array.isArray(p.task_ids) && p.task_ids.length > 0) {
        targetDesc = p.task_ids.join(', ');
      } else if (p.task_id) {
        targetDesc = String(p.task_id);
      }
      return {
        verb: 'Output',
        type: 'task_output' as const,
        target: desc || targetDesc || 'task output',
        description: desc,
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

  async function copyCommandText() {
    const cmd = ('cmd' in toolParsed && toolParsed.cmd) ? toolParsed.cmd : (parsedOutput.command || '');
    if (!cmd) return;
    try {
      await navigator.clipboard.writeText(cmd);
      copiedCommand = true;
      setTimeout(() => (copiedCommand = false), 2000);
    } catch {
      // ignore
    }
  }

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

  function handleCardKeyDown(e: KeyboardEvent) {
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;

    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      isExpanded = !isExpanded;
      if (isExpanded) {
        userHasScrolledActions = false;
      }
      return;
    }

    // When card is expanded, provide quick copy shortcuts:
    // 'c' to copy output, 'i' or 'Alt+C' to copy input/command
    if (isExpanded) {
      if (e.key.toLowerCase() === 'c' && !e.metaKey && !e.ctrlKey) {
        if (e.altKey || e.shiftKey) {
          e.preventDefault();
          copyCommandText();
        } else {
          e.preventDefault();
          copyOutputText();
        }
      } else if (e.key.toLowerCase() === 'i' && !e.metaKey && !e.ctrlKey) {
        e.preventDefault();
        copyCommandText();
      }
    }
  }
  function toggleExpand(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (target.closest('button') || target.closest('a')) return;
    isExpanded = !isExpanded;
    if (isExpanded) {
      userHasScrolledActions = false;
    }
  }
</script>

<div
  class="my-1 rounded-md border {toolParsed.type === 'subagent' ? 'border-violet-500/20 hover:border-violet-500/40 bg-violet-950/10' : 'border-white/5 bg-ant-bg-secondary/20 hover:border-white/10'} overflow-hidden text-xs transition-colors"
  style="content-visibility: auto; contain-intrinsic-size: auto 34px;"
>
  <!-- Minimalist Tool Header (Anti Gravity / VSCode Style) -->
  <div
    role="button"
    tabindex="0"
    onclick={toggleExpand}
    onkeydown={handleCardKeyDown}
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
        {:else if toolParsed.type === 'subagent'}
          <Bot size={13} class="text-violet-400" />
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
        <span class="text-[10px] font-mono {toolParsed.type === 'subagent' ? 'text-violet-300' : 'text-ant-text-muted'}">
          {formattedDuration}
        </span>
      {/if}

      {#if toolParsed.type === 'subagent' && toolCall.childToolCalls && toolCall.childToolCalls.length > 0}
        <span class="inline-flex items-center text-[10px] font-mono px-1.5 py-0.5 rounded bg-violet-500/20 text-violet-300 border border-violet-500/30">
          <Activity size={10} class="mr-1 text-violet-400" />
          {toolCall.childToolCalls.length}
        </span>
      {/if}

      {#if toolParsed.type === 'subagent'}
        <button
          type="button"
          onclick={(e) => {
            e.stopPropagation();
            dialogStore.openSubagentModal(toolCall, toolParsed.meta);
          }}
          class="flex items-center space-x-1 px-1.5 py-0.5 rounded text-[10px] font-mono font-medium text-violet-300 bg-violet-500/15 border border-violet-500/30 hover:bg-violet-500/25 hover:text-white transition cursor-pointer"
          title="Inspect Subagent Details"
        >
          <Bot size={11} />
          <span>Inspect</span>
        </button>
      {/if}

      {#if toolParsed.type === 'subagent' && subagentIsActuallyRunning}
        <span class="flex items-center text-[10px] font-medium text-violet-400 animate-pulse">
          <Loader2 size={11} class="animate-spin mr-1" />
        </span>
      {:else if toolCall.status === 'running'}
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
      {#if toolParsed.type === 'subagent'}
        <!-- Dedicated Inline Subagent Delegation Card -->
        <div class="rounded-lg border border-violet-500/30 bg-violet-950/20 overflow-hidden font-sans space-y-2 p-3 my-1">
          <!-- Top Badges & Actions -->
          <div class="flex items-center justify-between flex-wrap gap-2 text-xs font-mono select-none">
            <div class="flex items-center flex-wrap gap-1.5">
              <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-violet-500/20 text-violet-300 font-semibold border border-violet-500/30 text-[11px]">
                <Bot size={12} class="text-violet-400" />
                <span>Role: {toolParsed.meta.subagentType}</span>
              </span>
              {#if toolParsed.meta.model}
                <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-white/5 text-ant-text-secondary border border-white/5 text-[11px]">
                  <Cpu size={11} class="text-indigo-400" />
                  <span>{toolParsed.meta.model}</span>
                </span>
              {/if}
              <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-white/5 text-ant-text-secondary border border-white/5 text-[11px]">
                <span>{toolParsed.meta.isBackground ? 'Background' : 'Foreground'}</span>
              </span>
              {#if toolParsed.meta.taskId}
                <button
                  type="button"
                  onclick={() => {
                    navigator.clipboard.writeText(toolParsed.meta.taskId!);
                    copiedSubagentId = true;
                    setTimeout(() => (copiedSubagentId = false), 2000);
                  }}
                  class="inline-flex items-center gap-1 px-2 py-0.5 rounded bg-white/5 text-ant-text-secondary hover:text-white border border-white/5 text-[11px] hover:bg-white/10 transition cursor-pointer"
                  title="Copy Task ID: {toolParsed.meta.taskId}"
                >
                  <span>ID: {toolParsed.meta.taskId.slice(0, 8)}...</span>
                  {#if copiedSubagentId}
                    <Check size={10} class="text-emerald-400" />
                  {:else}
                    <Copy size={10} />
                  {/if}
                </button>
              {/if}
            </div>

            <!-- Inspect Modal Button -->
            <button
              type="button"
              onclick={(e) => {
                e.stopPropagation();
                dialogStore.openSubagentModal(toolCall, toolParsed.meta);
              }}
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-violet-500/20 hover:bg-violet-500/30 text-violet-200 border border-violet-500/40 text-[11px] font-medium transition cursor-pointer"
            >
              <Maximize2 size={11} />
              <span>Full Details</span>
            </button>
          </div>

          <!-- Task Actions List (Prioritized Hero View) -->
          <div class="rounded-md border border-violet-500/20 bg-black/40 overflow-hidden text-xs font-sans">
            <div class="flex items-center justify-between px-3 py-2 bg-violet-500/10 border-b border-white/5 text-[11px]">
              <div class="flex items-center space-x-1.5 text-violet-300 font-medium">
                <Activity size={12} class="text-violet-400" />
                <span>Task Actions ({toolCall.childToolCalls?.length || 0})</span>
                {#if subagentIsActuallyRunning}
                  <Loader2 size={11} class="animate-spin text-violet-400 ml-1" />
                {/if}
              </div>
              <button
                type="button"
                onclick={(e) => {
                  e.stopPropagation();
                  dialogStore.openSubagentModal(toolCall, toolParsed.meta);
                }}
                class="text-[10.5px] font-sans text-violet-400 hover:text-violet-300 hover:underline cursor-pointer flex items-center gap-1"
              >
                <span>Full Details</span>
                <ChevronRight size={11} />
              </button>
            </div>

            <div
              bind:this={actionsContainerEl}
              onscroll={handleActionsScroll}
              class="p-2 space-y-1 max-h-72 overflow-y-auto scrollbar-thin"
            >
              {#if toolCall.childToolCalls && toolCall.childToolCalls.length > 0}
                {#each toolCall.childToolCalls as child, idx (child.id || idx)}
                  {@const knownPlanTodos = planStore.getSubagentPlan(sessionStore.activeSessionId, toolParsed.meta?.taskId || toolCall.childSessionId || toolCall.id)?.todos || planStore.getPlan(sessionStore.activeSessionId)?.todos}
                  {@const summary = formatActionSummary(child.tool, child.params, knownPlanTodos)}
                  {@const childDur = getChildDuration(child)}
                  <div
                    class="flex items-center space-x-2 px-2.5 py-1.5 rounded bg-white/[0.03] hover:bg-white/[0.06] border border-white/[0.04] text-[11px] font-mono transition group cursor-pointer"
                    onclick={(e) => {
                      e.stopPropagation();
                      dialogStore.openSubagentModal(toolCall, toolParsed.meta);
                    }}
                    role="button"
                    tabindex="0"
                    onkeydown={(e) => {
                      if (e.key === 'Enter') dialogStore.openSubagentModal(toolCall, toolParsed.meta);
                    }}
                  >
                    <span class="shrink-0 flex items-center justify-center">
                      {#if child.status === 'running'}
                        <Loader2 size={11} class="animate-spin text-violet-400" />
                      {:else if child.status === 'completed'}
                        <CheckCircle2 size={11} class="text-emerald-400" />
                      {:else}
                        <XCircle size={11} class="text-rose-400" />
                      {/if}
                    </span>
                    <span class="px-1.5 py-0.5 rounded text-[9.5px] font-semibold uppercase tracking-wider shrink-0 {
                      summary.iconType === 'read' ? 'bg-blue-500/15 text-blue-400 border border-blue-500/20' :
                      summary.iconType === 'edit' ? 'bg-amber-500/15 text-amber-400 border border-amber-500/20' :
                      summary.iconType === 'write' ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/20' :
                      summary.iconType === 'search' ? 'bg-purple-500/15 text-purple-400 border border-purple-500/20' :
                      summary.iconType === 'list' ? 'bg-cyan-500/15 text-cyan-400 border border-cyan-500/20' :
                      summary.iconType === 'terminal' ? 'bg-orange-500/15 text-orange-400 border border-orange-500/20' :
                      summary.iconType === 'plan' ? 'bg-indigo-500/15 text-indigo-400 border border-indigo-500/20' :
                      'bg-zinc-500/15 text-zinc-300 border border-zinc-500/20'
                    }">
                      {summary.verb}
                    </span>
                    <span class="truncate flex-1 min-w-0 text-ant-text" title={summary.target}>
                      {summary.target}
                    </span>
                    {#if childDur}
                      <span class="text-[10px] text-ant-text-muted shrink-0 font-mono">{childDur}</span>
                    {/if}
                  </div>
                {/each}
              {:else if subagentIsActuallyRunning}
                <div class="flex items-center space-x-2 text-[11px] text-violet-300/70 py-2.5 px-2">
                  <Loader2 size={12} class="animate-spin text-violet-400 shrink-0" />
                  <span>Waiting for subagent actions...</span>
                </div>
              {:else}
                <div class="text-[11px] text-ant-text-muted py-2 px-2">
                  No recorded subagent actions.
                </div>
              {/if}
            </div>
          </div>

          <!-- Compact Completion / Error Status (Detailed view in modal) -->
          {#if subagentCleanOutput}
            <div class="flex items-center justify-between px-3 py-2 rounded-md bg-emerald-500/10 border border-emerald-500/20 text-xs">
              <div class="flex items-center gap-1.5 text-emerald-400">
                <CheckCircle2 size={12} class="text-emerald-400 shrink-0" />
                <span class="font-medium text-[11px]">Subagent output ready</span>
              </div>
              <button
                type="button"
                onclick={(e) => {
                  e.stopPropagation();
                  dialogStore.openSubagentModal(toolCall, toolParsed.meta);
                }}
                class="text-[11px] text-emerald-400 hover:text-emerald-300 hover:underline cursor-pointer flex items-center gap-1 font-sans"
              >
                <span>View Full Report</span>
                <ChevronRight size={10} />
              </button>
            </div>
          {:else if toolCall.status === 'error'}
            <div class="flex items-center justify-between px-3 py-2 rounded-md bg-rose-500/10 border border-rose-500/20 text-xs text-rose-300">
              <span class="text-[11px]">Subagent execution encountered an error</span>
              <button
                type="button"
                onclick={(e) => {
                  e.stopPropagation();
                  dialogStore.openSubagentModal(toolCall, toolParsed.meta);
                }}
                class="text-[11px] text-rose-300 hover:text-rose-200 hover:underline cursor-pointer flex items-center gap-1 font-sans"
              >
                <span>View Details</span>
                <ChevronRight size={10} />
              </button>
            </div>
          {/if}
        </div>
      {:else if toolCall.diff}
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
        <div class="rounded-lg border border-indigo-500/30 bg-ant-bg-secondary/90 dark:bg-indigo-950/10 overflow-hidden text-xs">
          <div class="flex items-center justify-between px-3 py-1.5 bg-indigo-500/10 dark:bg-indigo-900/20 border-b border-indigo-500/20 text-[11px] select-none text-indigo-700 dark:text-indigo-300">
            <div class="flex items-center space-x-1.5 font-medium">
              <ListTodo size={13} class="text-indigo-600 dark:text-indigo-400" />
              <span>Execution Plan Tasks</span>
            </div>
            {#if planStore.getPlan(sessionStore.activeSessionId)}
              {@const sp = planStore.getPlan(sessionStore.activeSessionId)}
              <span class="font-mono text-[10px] px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-700 dark:text-indigo-300 font-semibold">
                {sp ? `${sp.completedCount} / ${sp.totalCount} Overall Completed` : `${detectedTodos.filter(t => t.status === 'completed').length} / ${detectedTodos.length} Completed`}
              </span>
            {:else}
              <span class="font-mono text-[10px] px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-700 dark:text-indigo-300 font-semibold">
                {detectedTodos.filter(t => t.status === 'completed').length} / {detectedTodos.length} Completed
              </span>
            {/if}
          </div>

          <div class="p-2 space-y-1.5 max-h-72 overflow-y-auto scrollbar-thin">
            {#each detectedTodos as todo (todo.id)}
              {@const content = resolveTodoContent(todo, planStore.getPlan(sessionStore.activeSessionId))}
              <div class="flex items-start space-x-2.5 p-2 rounded-md bg-ant-bg dark:bg-white/[0.02] border border-ant-border-secondary dark:border-white/[0.04] transition-colors hover:bg-ant-bg-tertiary/60 dark:hover:bg-white/[0.04]">
                <div class="mt-0.5 shrink-0">
                  {#if todo.status === 'completed'}
                    <div class="w-4 h-4 rounded-full bg-emerald-500/20 text-emerald-600 dark:text-emerald-400 flex items-center justify-center border border-emerald-500/40" title="Completed">
                      <Check size={10} class="stroke-[3]" />
                    </div>
                  {:else if todo.status === 'in_progress'}
                    <div class="w-4 h-4 rounded-full bg-amber-500/20 text-amber-600 dark:text-amber-400 flex items-center justify-center border border-amber-500/40 animate-pulse" title="In Progress">
                      <Loader2 size={10} class="animate-spin" />
                    </div>
                  {:else if todo.status === 'cancelled'}
                    <div class="w-4 h-4 rounded-full bg-rose-500/20 text-rose-600 dark:text-rose-400 flex items-center justify-center border border-rose-500/40" title="Cancelled">
                      <XCircle size={10} />
                    </div>
                  {:else}
                    <div class="w-4 h-4 rounded-full border border-zinc-300 dark:border-zinc-600 bg-zinc-100 dark:bg-zinc-800/50 flex items-center justify-center" title="Pending">
                    </div>
                  {/if}
                </div>

                <div class="flex-1 min-w-0 font-sans leading-relaxed">
                  <div class="text-xs select-text {todo.status === 'completed' ? 'line-through opacity-60 text-ant-text-muted' : 'text-ant-text'}">
                    {@html renderTodoContent(content)}
                  </div>
                </div>

                {#if todo.priority}
                  <span class="text-[9px] uppercase tracking-wider font-mono px-1.5 py-0.2 rounded border {
                    todo.priority === 'high' ? 'bg-rose-500/10 text-rose-600 dark:text-rose-300 border-rose-500/20' :
                    todo.priority === 'medium' ? 'bg-amber-500/10 text-amber-700 dark:text-amber-300 border-amber-500/20' :
                    'bg-blue-500/10 text-blue-700 dark:text-blue-300 border-blue-500/20'
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
        <div class="rounded-lg border border-violet-500/30 bg-ant-bg-secondary/90 dark:bg-violet-950/10 overflow-hidden text-xs">
          <div class="flex items-center justify-between px-3 py-1.5 bg-violet-500/10 dark:bg-violet-900/20 border-b border-violet-500/20 text-[11px] select-none text-violet-700 dark:text-violet-300">
            <div class="flex items-center space-x-1.5 font-medium">
              <BookOpen size={13} class="text-violet-600 dark:text-violet-400" />
              <span>{toolParsed.type === 'plan_enter' ? 'Active Engineering Plan Document' : 'Completed Execution Plan'}</span>
            </div>

            <div class="flex items-center space-x-2">
              {#if toolParsed.planPath}
                <span class="font-mono text-[10px] text-violet-700/80 dark:text-violet-300/70 max-w-xs truncate" title={toolParsed.planPath}>
                  {toolParsed.planPath}
                </span>

                <!-- Reload / Refresh Button -->
                <button
                  type="button"
                  onclick={() => loadPlanContent(true)}
                  disabled={isPlanLoading}
                  class="p-1 rounded text-violet-700 dark:text-violet-300/70 hover:text-violet-900 dark:hover:text-violet-200 hover:bg-violet-500/10 dark:hover:bg-violet-800/30 transition disabled:opacity-50"
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
                  class="flex items-center space-x-1 px-1.5 py-0.5 rounded text-[10px] text-violet-700 dark:text-violet-300/80 hover:text-violet-900 dark:hover:text-white hover:bg-violet-500/10 dark:hover:bg-violet-800/30 transition"
                  title="Copy plan markdown"
                >
                  {#if copiedPlan}
                    <Check size={10} class="text-emerald-500 dark:text-emerald-400" />
                    <span class="text-emerald-500 dark:text-emerald-400">Copied</span>
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
              <div class="p-2.5 rounded bg-violet-500/10 dark:bg-violet-900/15 border border-violet-500/20 space-y-2">
                <div class="text-xs text-ant-text">
                  <p class="font-semibold text-violet-700 dark:text-violet-300 mb-1">Plan Objective / Context:</p>
                  <p class="text-ant-text-secondary leading-relaxed">{toolParsed.reason}</p>
                </div>
                {#if toolParsed.planPath}
                  <div class="pt-2 border-t border-violet-500/15 flex items-center justify-between text-[11px] text-ant-text-muted">
                    <span class="italic">Plan file initialized. Waiting for agent to write specifications.</span>
                    <button
                      type="button"
                      onclick={() => loadPlanContent(true)}
                      class="px-2 py-0.5 rounded bg-violet-600/15 hover:bg-violet-600/25 text-violet-700 dark:text-violet-300 border border-violet-500/30 flex items-center gap-1 transition"
                    >
                      <RotateCw size={10} class={isPlanLoading ? 'animate-spin' : ''} />
                      <span>Check for Updates</span>
                    </button>
                  </div>
                {/if}
              </div>
            {:else if planLoadError}
              <div class="p-3 rounded bg-rose-500/10 border border-rose-500/20 text-xs text-rose-600 dark:text-rose-300 flex items-center justify-between">
                <span>Failed to load plan file: {planLoadError}</span>
                <button
                  type="button"
                  onclick={() => loadPlanContent(true)}
                  class="px-2 py-0.5 rounded bg-rose-500/20 hover:bg-rose-500/30 text-rose-700 dark:text-rose-200 border border-rose-500/40"
                >
                  Retry
                </button>
              </div>
            {:else}
              <div class="text-xs text-ant-text-muted italic">
                No markdown plan file available.
              </div>
            {/if}
          </div>
        </div>
      {:else if toolParsed.type === 'terminal'}
        <!-- 2f. Dedicated Terminal Command Card (Clean In & Out Inspector) -->
        <div class="space-y-1.5 text-[11px]">
          <!-- Panel IN: Full Command Script -->
          {#if toolParsed.cmd}
            <div class="rounded border border-indigo-500/20 bg-ant-bg-secondary/60 overflow-hidden">
              <div class="flex items-center justify-between px-2.5 py-1 bg-indigo-500/10 border-b border-indigo-500/15 text-[10px] select-none text-indigo-400">
                <div class="flex items-center space-x-1.5 font-medium">
                  <span class="px-1 py-0.2 rounded bg-indigo-500/20 text-indigo-300 font-bold uppercase tracking-wider text-[9px]">IN</span>
                  <span class="text-ant-text-secondary font-mono">Command</span>
                </div>
                <button
                  type="button"
                  onclick={copyCommandText}
                  class="flex items-center space-x-1 text-ant-text-muted hover:text-white transition px-1.5 py-0.5 rounded hover:bg-ant-bg"
                  title="Copy command (Press 'i' or 'Shift+C')"
                >
                  {#if copiedCommand}
                    <Check size={10} class="text-ant-success" />
                    <span class="text-ant-success">Copied</span>
                  {:else}
                    <Copy size={10} />
                    <span>Copy IN</span>
                  {/if}
                </button>
              </div>
              <pre class="p-2 leading-relaxed overflow-x-auto max-h-48 scrollbar-thin text-indigo-300 font-mono text-[11px] whitespace-pre-wrap break-all select-text"><code>$ {toolParsed.cmd}</code></pre>
            </div>
          {/if}

          <!-- Panel OUT: Terminal Execution Output -->
          {#if toolCall.result !== undefined}
            <div class="rounded border border-white/5 bg-ant-bg-secondary/50 overflow-hidden">
              <div class="flex items-center justify-between px-2.5 py-1 bg-ant-bg-tertiary/60 border-b border-white/5 text-[10px] select-none text-ant-text-muted">
                <div class="flex items-center space-x-2">
                  <span class="px-1 py-0.2 rounded bg-white/10 text-ant-text font-bold uppercase tracking-wider text-[9px]">OUT</span>
                  <span class="uppercase tracking-wider font-semibold {toolCall.status === 'error' || parsedOutput.status === 'failed' || (parsedOutput.exitCode !== undefined && parsedOutput.exitCode !== 0) ? 'text-ant-error' : 'text-ant-text-secondary'}">
                    {toolCall.status === 'error' || parsedOutput.status === 'failed' || (parsedOutput.exitCode !== undefined && parsedOutput.exitCode !== 0) ? 'Execution Failed' : 'Output'}
                  </span>
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
                </div>
                <button
                  type="button"
                  onclick={copyOutputText}
                  class="flex items-center space-x-1 text-ant-text-muted hover:text-white transition px-1.5 py-0.5 rounded hover:bg-ant-bg"
                  title="Copy output (Press 'c')"
                >
                  {#if copiedOutput}
                    <Check size={10} class="text-ant-success" />
                    <span class="text-ant-success">Copied</span>
                  {:else}
                    <Copy size={10} />
                    <span>Copy OUT</span>
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
