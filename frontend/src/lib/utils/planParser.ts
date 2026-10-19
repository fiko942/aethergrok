import { marked } from 'marked';

/**
 * Utility functions for parsing TodosUpdated payloads, plan updates,
 * and structured task execution plans from tool outputs or file writes.
 */

export interface TodoItem {
  id: string;
  content: string;
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority?: 'low' | 'medium' | 'high' | string;
  subagentId?: string;
  subagentName?: string;
  createdAt?: number;
  updatedAt?: number;
}

export interface SubagentPlanState {
  subagentId: string;
  subagentName: string;
  todos: TodoItem[];
  completedCount: number;
  inProgressCount: number;
  pendingCount: number;
  totalCount: number;
  progressPercent: number;
  lastUpdated: number;
}

export interface SessionPlanState {
  todos: TodoItem[];
  mainTodos: TodoItem[];
  subagentPlans: Record<string, SubagentPlanState>;
  lastUpdated: number;
  completedCount: number;
  inProgressCount: number;
  pendingCount: number;
  totalCount: number;
  progressPercent: number;
  sourceToolId?: string;
  sourceToolName?: string;
}

/**
 * Normalizes raw todo status strings into standardized status enum
 */
export function normalizeTodoStatus(status: unknown): 'pending' | 'in_progress' | 'completed' | 'cancelled' {
  if (typeof status !== 'string') return 'pending';
  const s = status.toLowerCase().trim();
  if (s === 'completed' || s === 'done' || s === 'finished' || s === 'complete' || s === 'success') {
    return 'completed';
  }
  if (s === 'in_progress' || s === 'in-progress' || s === 'running' || s === 'doing' || s === 'active') {
    return 'in_progress';
  }
  if (s === 'cancelled' || s === 'canceled' || s === 'skipped' || s === 'abort' || s === 'failed') {
    return 'cancelled';
  }
  return 'pending';
}

/**
 * Cleans markdown noise and wrapping markers from todo content strings:
 * - Strips outer bold/italic wrappers like `**Step 1: Write test**`
 * - Strips leading markdown list bullets if accidentally preserved
 */
export function cleanTodoContent(raw: string): string {
  if (!raw) return '';
  let text = String(raw).trim();

  // Strip leading list markers if present (e.g. "- [ ] " or "* ") without stripping bold asterisks
  text = text.replace(/^[-*•](?:\s*\[[ xX/_\-a-zA-Z]*\]|\s+)\s*/, '').trim();

  // Strip outer bold/italic wrappers if they enclose the entire text
  if (text.startsWith('**') && text.endsWith('**') && text.length >= 4) {
    const inner = text.slice(2, -2).trim();
    if (!inner.includes('**')) {
      text = inner;
    }
  } else if (text.startsWith('__') && text.endsWith('__') && text.length >= 4) {
    const inner = text.slice(2, -2).trim();
    if (!inner.includes('__')) {
      text = inner;
    }
  } else if (text.startsWith('*') && text.endsWith('*') && text.length >= 2) {
    const inner = text.slice(1, -1).trim();
    if (!inner.includes('*')) {
      text = inner;
    }
  }

  return text;
}

/**
 * Formats and renders a todo item's content with clean typography:
 * - Unwraps outer markdown bold/italic tags
 * - Highlights "Step N:" or "Task N:" prefixes elegantly
 * - Parses inline markdown (code blocks, bold, italic, links) safely
 */
export function renderTodoContent(raw: string): string {
  if (!raw) return '';
  let text = cleanTodoContent(String(raw));

  // Handle case where bold was on step label only: **Step 1**: ...
  const boldPrefixMatch = text.match(/^\*\*((?:Step|Task)\s+\d+)\*\*(?:\s*[:\-]\s*|\s+)(.*)$/i);
  if (boldPrefixMatch) {
    const label = boldPrefixMatch[1];
    const rest = boldPrefixMatch[2];
    let parsedRest = '';
    try {
      parsedRest = rest ? (marked.parseInline(rest) as string) : '';
    } catch {
      parsedRest = rest;
    }
    return `<span class="font-semibold text-ant-text shrink-0">${label}:</span> ${parsedRest ? `<span class="text-ant-text/90">${parsedRest}</span>` : ''}`;
  }

  // Handle plain Step N: or Task N: prefix
  const stepMatch = text.match(/^((?:Step|Task)\s+\d+)(?:\s*[:\-]\s*|\s+)(.*)$/i);
  if (stepMatch) {
    const label = stepMatch[1];
    const rest = stepMatch[2];
    let parsedRest = '';
    try {
      parsedRest = rest ? (marked.parseInline(rest) as string) : '';
    } catch {
      parsedRest = rest;
    }
    return `<span class="font-semibold text-ant-text shrink-0">${label}:</span> ${parsedRest ? `<span class="text-ant-text/90">${parsedRest}</span>` : ''}`;
  }

  try {
    return (marked.parseInline(text) as string) || text;
  } catch {
    return text;
  }
}

/**
 * Extracts and parses TodosUpdated or plan todos from a raw string or object.
 * Supports:
 * 1. { "TodosUpdated": { "state": { "todos": { "id": { "content": "...", "status": "..." } } } } }
 * 2. { "TodosUpdated": { "todos": [ ... ] } }
 * 3. { "state": { "todos": { ... } } }
 * 4. { "todos": [ { "id": "...", "content": "..." } ] }
 * 5. todo_write tool params { "todos": [...] }
 * 6. Markdown todo lists "- [completed] id: content" or "- [x] content"
 */
export function parseTodosUpdated(raw: unknown): TodoItem[] | null {
  if (!raw) return null;

  let obj: any = raw;

  if (typeof raw === 'string') {
    const trimmed = raw.trim();
    if (!trimmed.startsWith('{') && !trimmed.startsWith('[')) {
      // Check for markdown list items (e.g. "- [completed] task: description" or "- [ ] task")
      const lines = trimmed.split('\n').map((l) => l.trim()).filter(Boolean);
      const markdownTodos: TodoItem[] = [];
      const mdRegex = /^[-*]\s*\[([ xX/_\-a-zA-Z]+)\]\s*(?:([a-zA-Z0-9_\-]+):\s*)?(.*)$/;

      let currentGroupTitle: string | undefined = undefined;
      let currentGroupId: string | undefined = undefined;

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];

        // Detect section headings like "### Task 1: Component Name" or "## Subagent: Name"
        const headingMatch = line.match(/^#{1,4}\s+(?:Task\s*(\d+)[:\-]?\s*)?(.*)$/i);
        if (headingMatch && !line.includes('- [')) {
          const taskNum = headingMatch[1];
          const taskTitle = headingMatch[2]?.trim();
          if (
            taskTitle &&
            !taskTitle.toLowerCase().includes('global constraints') &&
            !taskTitle.toLowerCase().includes('implementation plan') &&
            !taskTitle.toLowerCase().includes('overview')
          ) {
            currentGroupTitle = taskNum ? `Task ${taskNum}: ${taskTitle}` : taskTitle;
            currentGroupId = taskNum ? `task-${taskNum}` : taskTitle.toLowerCase().replace(/[^a-z0-9]+/g, '-').slice(0, 30);
          }
        }

        const match = line.match(mdRegex);
        if (match) {
          const rawStatus = match[1].trim().toLowerCase();
          const validStatuses = new Set([
            '', ' ', 'x', '/', '-', 'completed', 'done', 'finished', 'complete', 'success',
            'in_progress', 'in-progress', 'running', 'doing', 'active',
            'cancelled', 'canceled', 'skipped', 'abort', 'failed',
            'pending', 'todo', 'open'
          ]);
          if (!validStatuses.has(rawStatus)) {
            continue;
          }

          const namedId = match[2]?.trim();
          const content = (match[3] || '').trim();
          let status: 'pending' | 'in_progress' | 'completed' | 'cancelled' = 'pending';

          if (rawStatus === 'x' || rawStatus === 'completed' || rawStatus === 'done') {
            status = 'completed';
          } else if (rawStatus === '/' || rawStatus === 'in_progress' || rawStatus === 'running' || rawStatus === 'active') {
            status = 'in_progress';
          } else if (rawStatus === 'cancelled' || rawStatus === 'canceled' || rawStatus === '-') {
            status = 'cancelled';
          }

          if (content || namedId) {
            const cleaned = cleanTodoContent(content || namedId || `Task ${i + 1}`);
            markdownTodos.push({
              id: namedId || (currentGroupId ? `${currentGroupId}-step-${i + 1}` : `task-${i + 1}`),
              content: cleaned,
              status,
              priority: undefined,
              subagentId: currentGroupId,
              subagentName: currentGroupTitle,
              updatedAt: Date.now()
            });
          }
        }
      }

      if (markdownTodos.length > 0) {
        return deduplicateTodos(markdownTodos);
      }

      return null;
    }

    try {
      obj = JSON.parse(trimmed);
    } catch {
      // Check if there's an embedded JSON block inside text (e.g. inside a markdown code fence)
      const jsonMatch = trimmed.match(/\{[\s\S]*"TodosUpdated"[\s\S]*\}/);
      if (jsonMatch) {
        try {
          obj = JSON.parse(jsonMatch[0]);
        } catch {
          return null;
        }
      } else {
        return null;
      }
    }
  }

  if (!obj || typeof obj !== 'object') return null;

  // Case 1: TodosUpdated root
  let todosSource = obj.TodosUpdated || obj.todosUpdated || obj.todos_updated;
  if (todosSource) {
    if (todosSource.state && todosSource.state.todos) {
      todosSource = todosSource.state.todos;
    } else if (todosSource.todos) {
      todosSource = todosSource.todos;
    }
  } else if (obj.state && obj.state.todos) {
    todosSource = obj.state.todos;
  } else if (obj.todos) {
    todosSource = obj.todos;
  } else if (obj.result) {
    // Nested result property
    return parseTodosUpdated(obj.result);
  } else if (typeof obj.prompt === 'string') {
    return parseTodosUpdated(obj.prompt);
  } else if (typeof obj.content === 'string') {
    return parseTodosUpdated(obj.content);
  }

  if (!todosSource) return null;

  const result: TodoItem[] = [];

  // Check if root has subagent context
  const rootSubagentId = typeof obj.subagentId === 'string' ? obj.subagentId : (typeof obj.subagent_id === 'string' ? obj.subagent_id : undefined);
  const rootSubagentName = typeof obj.subagentName === 'string' ? obj.subagentName : (typeof obj.subagent_name === 'string' ? obj.subagent_name : undefined);

  // If todosSource is an Object dictionary (e.g., { "task-1": { content: "...", status: "..." } })
  if (typeof todosSource === 'object' && !Array.isArray(todosSource)) {
    for (const [key, val] of Object.entries(todosSource)) {
      if (val && typeof val === 'object') {
        const item = val as Record<string, any>;
        const sId = typeof item.subagentId === 'string' ? item.subagentId : (typeof item.subagent_id === 'string' ? item.subagent_id : rootSubagentId);
        const sName = typeof item.subagentName === 'string' ? item.subagentName : (typeof item.subagent_name === 'string' ? item.subagent_name : rootSubagentName);
        result.push({
          id: String(item.id || key),
          content: cleanTodoContent(String(item.content || item.title || item.description || key)),
          status: normalizeTodoStatus(item.status),
          priority: typeof item.priority === 'string' && item.priority.trim() ? item.priority.trim() : undefined,
          subagentId: sId,
          subagentName: sName,
          updatedAt: Date.now()
        });
      }
    }
  } else if (Array.isArray(todosSource)) {
    // If todosSource is an Array of items
    todosSource.forEach((item, idx) => {
      if (typeof item === 'string') {
        result.push({
          id: `task-${idx + 1}`,
          content: cleanTodoContent(item),
          status: 'pending',
          priority: undefined,
          subagentId: rootSubagentId,
          subagentName: rootSubagentName,
          updatedAt: Date.now()
        });
      } else if (item && typeof item === 'object') {
        const id = String(item.id || `task-${idx + 1}`);
        const sId = typeof item.subagentId === 'string' ? item.subagentId : (typeof item.subagent_id === 'string' ? item.subagent_id : rootSubagentId);
        const sName = typeof item.subagentName === 'string' ? item.subagentName : (typeof item.subagent_name === 'string' ? item.subagent_name : rootSubagentName);
        result.push({
          id,
          content: cleanTodoContent(String(item.content || item.title || item.description || id)),
          status: normalizeTodoStatus(item.status),
          priority: typeof item.priority === 'string' && item.priority.trim() ? item.priority.trim() : undefined,
          subagentId: sId,
          subagentName: sName,
          updatedAt: Date.now()
        });
      }
    });
  }

  return result.length > 0 ? deduplicateTodos(result) : null;
}

/**
 * Deduplicates todos list keeping the latest status and unique ID
 */
export function deduplicateTodos(todos: TodoItem[]): TodoItem[] {
  const map = new Map<string, TodoItem>();
  for (const item of todos) {
    const key = (item.subagentId ? `${item.subagentId}:${item.id}` : item.id.trim()) || item.content.trim();
    if (!key) continue;
    map.set(key, item);
  }
  return Array.from(map.values());
}

/**
 * Calculates summary metrics for a list of Todo items
 */
export function calculatePlanMetrics(todos: TodoItem[], sourceToolId?: string, sourceToolName?: string): SessionPlanState {
  const total = todos.length;
  const completed = todos.filter((t) => t.status === 'completed').length;
  const inProgress = todos.filter((t) => t.status === 'in_progress').length;
  const pending = todos.filter((t) => t.status === 'pending').length;
  const progressPercent = total > 0 ? Math.round((completed / total) * 100) : 0;

  const mainTodos: TodoItem[] = [];
  const subagentPlans: Record<string, SubagentPlanState> = {};

  for (const item of todos) {
    if (item.subagentId) {
      const sId = item.subagentId;
      if (!subagentPlans[sId]) {
        subagentPlans[sId] = {
          subagentId: sId,
          subagentName: item.subagentName || `Subagent ${sId.slice(0, 8)}`,
          todos: [],
          completedCount: 0,
          inProgressCount: 0,
          pendingCount: 0,
          totalCount: 0,
          progressPercent: 0,
          lastUpdated: item.updatedAt || Date.now()
        };
      }
      subagentPlans[sId].todos.push(item);
    } else {
      mainTodos.push(item);
    }
  }

  for (const sp of Object.values(subagentPlans)) {
    sp.totalCount = sp.todos.length;
    sp.completedCount = sp.todos.filter((t) => t.status === 'completed').length;
    sp.inProgressCount = sp.todos.filter((t) => t.status === 'in_progress').length;
    sp.pendingCount = sp.todos.filter((t) => t.status === 'pending').length;
    sp.progressPercent = sp.totalCount > 0 ? Math.round((sp.completedCount / sp.totalCount) * 100) : 0;
  }

  return {
    todos,
    mainTodos,
    subagentPlans,
    lastUpdated: Date.now(),
    completedCount: completed,
    inProgressCount: inProgress,
    pendingCount: pending,
    totalCount: total,
    progressPercent,
    sourceToolId,
    sourceToolName
  };
}

/**
 * Strips quotes, sentence markers, and trailing punctuation from extracted path strings.
 */
export function cleanExtractedPath(raw: string): string {
  if (!raw || typeof raw !== 'string') return '';
  let p = raw.trim();
  // Strip surrounding quotes or backticks or markdown formatting
  p = p.replace(/^[`"'<(\[]+/, '').replace(/[`"'>)\]]+$/, '').trim();

  // If the path is followed by a sentence like ". The file exists and is empty."
  const sentenceMarkers = ['. The file', '.\nThe file', '. The plan', '.\nThe plan'];
  for (const marker of sentenceMarkers) {
    const idx = p.indexOf(marker);
    if (idx !== -1) {
      p = p.substring(0, idx).trim();
    }
  }

  // Strip trailing punctuation
  p = p.replace(/[.,;:!?]+$/, '').trim();
  // Strip again in case quotes were inside
  p = p.replace(/^[`"']+/, '').replace(/[`"']+$/, '').trim();
  return p;
}

/**
 * Robustly extracts the plan.md file path from enter_plan_mode or exit_plan_mode
 * params, result string, or structured result object.
 */
export function extractPlanFilePath(params: unknown, result: unknown): string {
  // 1. Check params object
  if (typeof params === 'object' && params !== null) {
    const p = params as Record<string, unknown>;
    const directPath = p.plan_file_path || p.planPath || p.plan_path || p.filePath || p.file_path || p.path || p.target_file;
    if (typeof directPath === 'string' && directPath.trim()) {
      const cleaned = cleanExtractedPath(directPath);
      if (cleaned) return cleaned;
    }
  }

  // 2. Check result object
  if (typeof result === 'object' && result !== null) {
    const r = result as Record<string, unknown>;
    const directPath = r.plan_file_path || r.planPath || r.plan_path || r.path;
    if (typeof directPath === 'string' && directPath.trim()) {
      const cleaned = cleanExtractedPath(directPath);
      if (cleaned) return cleaned;
    }
    if (typeof r.Entered === 'object' && r.Entered !== null) {
      const entered = r.Entered as Record<string, unknown>;
      if (typeof entered.plan_file_path === 'string' && entered.plan_file_path.trim()) {
        const cleaned = cleanExtractedPath(entered.plan_file_path);
        if (cleaned) return cleaned;
      }
    }
  }

  // 3. Check result string
  if (typeof result === 'string') {
    const res = result;
    // Prefix regex matches: "Plan file: <path>", "saved at: <path>", "plan saved to: <path>", "Write your plan to <path>"
    const prefixMatch = res.match(/(?:Plan file:|saved at:|plan saved to:|plan to|Write your plan to|plan at:)\s*([^\r\n]+)/i);
    if (prefixMatch && prefixMatch[1]) {
      const candidate = cleanExtractedPath(prefixMatch[1]);
      if (candidate) return candidate;
    }

    // Direct match for session path containing plan.md or ending in .md
    const pathMatch = res.match(/(?:[a-zA-Z]:[\\\/]|\/|~[\\\/])[^\s"'\r\n]+?plan\.md/i);
    if (pathMatch && pathMatch[0]) {
      const candidate = cleanExtractedPath(pathMatch[0]);
      if (candidate) return candidate;
    }
  }

  return '';
}

/**
 * Extracts embedded markdown plan content directly from exit_plan_mode or tool result
 * when the result payload includes "## Plan:\n..." or markdown headings.
 */
export function extractPlanMarkdownFromResult(result: unknown): string {
  if (!result || typeof result !== 'string') return '';
  const text = result.trim();

  // If it starts with or contains "## Plan:\n"
  const planIdx = text.indexOf('## Plan:');
  if (planIdx !== -1) {
    return text.substring(planIdx + 8).trim();
  }

  // If text starts with a markdown title "# " or "## "
  if (text.startsWith('# ') || text.startsWith('## ')) {
    return text;
  }

  // If there is an implementation plan heading inside the text
  const headingMatch = text.match(/(#+\s*(?:Rencana|Plan|Implementation Plan|Engineering Plan)[\s\S]*)/i);
  if (headingMatch && headingMatch[1]) {
    return headingMatch[1].trim();
  }

  return '';
}

/**
 * Resolves the display content for a Todo item, checking the full session plan if content is missing or placeholder.
 */
export function resolveTodoContent(item: TodoItem, sessionPlan?: SessionPlanState | null): string {
  if (item.content && item.content !== item.id && !item.content.startsWith('task-')) {
    return item.content;
  }
  if (sessionPlan && sessionPlan.todos) {
    const found = sessionPlan.todos.find((t) => t.id === item.id);
    if (found && found.content) {
      return found.content;
    }
  }
  return item.content || item.id;
}

/**
 * Truncates text with an ellipsis if it exceeds maxLen.
 */
export function truncatePlanText(text: string, maxLen = 32): string {
  if (!text) return '';
  return text.length > maxLen ? text.slice(0, maxLen - 1) + '…' : text;
}

/**
 * Generates a human-friendly, informative summary label for a plan update / todo_write action.
 */
export function formatPlanUpdateSummary(
  todos: TodoItem[],
  isMerge = true,
  sessionPlan?: SessionPlanState | null
): string {
  if (!todos || todos.length === 0) {
    return 'Plan updated';
  }

  // If not a merge (fresh plan creation or full replacement)
  if (!isMerge) {
    const count = todos.length;
    return `Created plan with ${count} ${count === 1 ? 'task' : 'tasks'}`;
  }

  const completedItems = todos.filter((t) => t.status === 'completed');
  const inProgressItems = todos.filter((t) => t.status === 'in_progress');
  const cancelledItems = todos.filter((t) => t.status === 'cancelled');

  let actionText = '';

  if (completedItems.length === 1 && inProgressItems.length === 1) {
    const doneName = resolveTodoContent(completedItems[0], sessionPlan);
    const startName = resolveTodoContent(inProgressItems[0], sessionPlan);
    actionText = `Completed "${truncatePlanText(doneName, 24)}" → Started "${truncatePlanText(startName, 24)}"`;
  } else if (completedItems.length === 1 && inProgressItems.length === 0) {
    const doneName = resolveTodoContent(completedItems[0], sessionPlan);
    actionText = `Completed "${truncatePlanText(doneName, 36)}"`;
  } else if (inProgressItems.length === 1 && completedItems.length === 0) {
    const startName = resolveTodoContent(inProgressItems[0], sessionPlan);
    actionText = `Started "${truncatePlanText(startName, 36)}"`;
  } else if (cancelledItems.length === 1 && todos.length === 1) {
    const cancelName = resolveTodoContent(cancelledItems[0], sessionPlan);
    actionText = `Cancelled "${truncatePlanText(cancelName, 36)}"`;
  } else if (completedItems.length > 0 && inProgressItems.length > 0) {
    actionText = `Completed ${completedItems.length} & Started ${inProgressItems.length} tasks`;
  } else if (completedItems.length > 0) {
    actionText = `Completed ${completedItems.length} ${completedItems.length === 1 ? 'task' : 'tasks'}`;
  } else if (inProgressItems.length > 0) {
    actionText = `In Progress: ${inProgressItems.length} ${inProgressItems.length === 1 ? 'task' : 'tasks'}`;
  } else {
    actionText = `Updated ${todos.length} ${todos.length === 1 ? 'task' : 'tasks'}`;
  }

  // Append cumulative session progress if available
  if (sessionPlan && sessionPlan.totalCount > 0) {
    return `${actionText} · ${sessionPlan.completedCount}/${sessionPlan.totalCount} completed`;
  }

  return actionText;
}
