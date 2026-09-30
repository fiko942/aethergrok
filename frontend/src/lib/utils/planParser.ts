/**
 * Utility functions for parsing TodosUpdated payloads, plan updates,
 * and structured task execution plans from tool outputs or file writes.
 */

export interface TodoItem {
  id: string;
  content: string;
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled';
  priority?: 'low' | 'medium' | 'high' | string;
  createdAt?: number;
  updatedAt?: number;
}

export interface SessionPlanState {
  todos: TodoItem[];
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

      for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        const match = line.match(mdRegex);
        if (match) {
          const rawStatus = match[1].trim().toLowerCase();
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
            markdownTodos.push({
              id: namedId || `task-${i + 1}`,
              content: content || namedId || `Task ${i + 1}`,
              status,
              priority: 'medium',
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
  }

  if (!todosSource) return null;

  const result: TodoItem[] = [];

  // If todosSource is an Object dictionary (e.g., { "task-1": { content: "...", status: "..." } })
  if (typeof todosSource === 'object' && !Array.isArray(todosSource)) {
    for (const [key, val] of Object.entries(todosSource)) {
      if (val && typeof val === 'object') {
        const item = val as Record<string, any>;
        result.push({
          id: String(item.id || key),
          content: String(item.content || item.title || item.description || key),
          status: normalizeTodoStatus(item.status),
          priority: typeof item.priority === 'string' ? item.priority : 'medium',
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
          content: item,
          status: 'pending',
          priority: 'medium',
          updatedAt: Date.now()
        });
      } else if (item && typeof item === 'object') {
        const id = String(item.id || `task-${idx + 1}`);
        result.push({
          id,
          content: String(item.content || item.title || item.description || id),
          status: normalizeTodoStatus(item.status),
          priority: typeof item.priority === 'string' ? item.priority : 'medium',
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
    const key = item.id.trim() || item.content.trim();
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

  return {
    todos,
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
