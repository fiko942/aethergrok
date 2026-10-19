import {
  parseTodosUpdated,
  calculatePlanMetrics,
  type TodoItem,
  type SessionPlanState,
  type SubagentPlanState
} from '../utils/planParser';
import { isSubagentToolCall, extractSubagentMeta } from '../utils/toolUtils';

export interface PlanTrackerState {
  // Session ID -> SessionPlanState
  plans: Record<string, SessionPlanState>;
  // Floating widget UI state per session or global
  isFloatingOpen: boolean;
  isFloatingCompact: boolean;
  activePlanSessionId: string | null;
  // Has unread update notification flag
  hasUnreadUpdate: boolean;
}

/**
 * Determines whether a tool call is explicitly designed to produce or update plans/todos.
 * File inspection, searching, and general execution tools are strictly excluded to prevent
 * accidental ingestion of markdown checkboxes and task lists from read project files.
 */
export function isPlanRelatedTool(toolName?: string): boolean {
  if (!toolName) return false;
  const lower = toolName.toLowerCase().trim().replace(/[-_]/g, '');

  // Strict blacklist: tools that read files, list directories, grep, or run commands
  if (
    lower.startsWith('read') ||
    lower.endsWith('read') ||
    lower === 'fileread' ||
    lower === 'cat' ||
    lower === 'viewfile' ||
    lower === 'searchreplace' ||
    lower === 'edit' ||
    lower === 'listdir' ||
    lower === 'dirlist' ||
    lower === 'ls' ||
    lower === 'grep' ||
    lower === 'glob' ||
    lower === 'search' ||
    lower === 'runterminalcommand' ||
    lower === 'terminal' ||
    lower === 'bash' ||
    lower === 'exec' ||
    lower === 'websearch' ||
    lower === 'webfetch' ||
    lower === 'fetch'
  ) {
    return false;
  }

  return (
    lower.includes('todo') ||
    lower.includes('plan') ||
    lower === 'subagentoutput' ||
    lower.includes('subagent')
  );
}

export class PlanStore {
  // Map from sessionId to current SessionPlanState
  sessionPlans = $state<Record<string, SessionPlanState>>({});
  
  // Floating widget display state
  isFloatingVisible = $state<boolean>(true);
  isCompact = $state<boolean>(false);
  hasNewUpdate = $state<boolean>(false);
  lastUpdatedSessionId = $state<string | null>(null);

  /**
   * Get the active plan state for a given session
   */
  getPlan(sessionId?: string | null): SessionPlanState | null {
    if (!sessionId) return null;
    return this.sessionPlans[sessionId] || null;
  }

  /**
   * Get the plan state specifically for a subagent within a session
   */
  getSubagentPlan(sessionId?: string | null, subagentId?: string | null): SubagentPlanState | null {
    if (!sessionId || !subagentId) return null;
    const plan = this.getPlan(sessionId);
    if (!plan || !plan.subagentPlans) return null;
    return plan.subagentPlans[subagentId] || null;
  }

  /**
   * Get all subagent plans within a session
   */
  getSubagentPlans(sessionId?: string | null): SubagentPlanState[] {
    if (!sessionId) return [];
    const plan = this.getPlan(sessionId);
    if (!plan || !plan.subagentPlans) return [];
    return Object.values(plan.subagentPlans);
  }

  /**
   * Record or update plan state for a session.
   * If todos have changed, trigger a subtle visual notification pulse and expand if appropriate.
   */
  updatePlan(sessionId: string, todos: TodoItem[], sourceToolId?: string, sourceToolName?: string): SessionPlanState {
    const existing = this.sessionPlans[sessionId];
    const newMetrics = calculatePlanMetrics(todos, sourceToolId, sourceToolName);

    // Check if there was actual status or count progress
    const isDifferent = !existing ||
      existing.todos.length !== todos.length ||
      existing.completedCount !== newMetrics.completedCount ||
      existing.inProgressCount !== newMetrics.inProgressCount ||
      JSON.stringify(existing.todos) !== JSON.stringify(todos);

    this.sessionPlans[sessionId] = newMetrics;

    if (isDifferent) {
      this.lastUpdatedSessionId = sessionId;
      this.hasNewUpdate = true;
      // Auto-reopen or ensure visible when there is a fresh plan update
      this.isFloatingVisible = true;
    }

    return newMetrics;
  }

  /**
   * Scan and ingest a tool call result or params to see if it contains plan updates.
   * Supports optional subagentContext to attribute plan steps directly to a subagent.
   */
  inspectToolForPlan(
    sessionId: string,
    toolId: string,
    toolName: string,
    params: unknown,
    result: unknown,
    subagentContext?: { id: string; name?: string }
  ): boolean {
    const isPlanTool = isPlanRelatedTool(toolName) || isSubagentToolCall(toolName, params) || Boolean(subagentContext);
    if (!isPlanTool) {
      // Non-plan tools (e.g. read_file, list_dir, run_terminal_command, grep) must NEVER
      // parse arbitrary text/markdown outputs as plan checklists.
      // Only allow if params or result explicitly contain a structured JSON object
      // with a dedicated "todos" or "TodosUpdated" property.
      let hasStructuredTodos =
        (params && typeof params === 'object' && ('todos' in (params as any) || 'TodosUpdated' in (params as any))) ||
        (result && typeof result === 'object' && ('todos' in (result as any) || 'TodosUpdated' in (result as any)));

      if (!hasStructuredTodos && typeof result === 'string') {
        const trimmed = result.trim();
        if (trimmed.startsWith('{') && (trimmed.includes('"TodosUpdated"') || trimmed.includes('"todos"'))) {
          hasStructuredTodos = true;
        }
      }

      if (!hasStructuredTodos) {
        return false;
      }
    }

    const isTodoWrite = (toolName || '').toLowerCase().includes('todo_write');
    let isMerge = true;

    if (params && typeof params === 'object') {
      const p = params as Record<string, any>;
      if (typeof p.merge === 'boolean') {
        isMerge = p.merge;
      }
    }

    // Infer subagent context if not explicitly provided
    let effectiveSubagent = subagentContext;
    if (!effectiveSubagent) {
      if (isSubagentToolCall(toolName, params)) {
        const meta = extractSubagentMeta(params, typeof result === 'string' ? result : undefined);
        effectiveSubagent = {
          id: meta.taskId || toolId,
          name: meta.description
        };
      } else if (params && typeof params === 'object') {
        const p = params as Record<string, any>;
        const sId = p.subagentId || p.subagent_id || p.taskId || p.task_id;
        const sName = p.subagentName || p.subagent_name || p.description;
        if (sId) {
          effectiveSubagent = {
            id: String(sId),
            name: sName ? String(sName) : undefined
          };
        }
      }
    }

    // 1. Check result first
    let todos = parseTodosUpdated(result);
    // 2. Check params/input next (e.g. todo_write tool)
    if (!todos && params) {
      todos = parseTodosUpdated(params);
    }

    if (todos && todos.length > 0) {
      // Attribute todos to subagent if context is known
      if (effectiveSubagent) {
        for (const item of todos) {
          item.subagentId = effectiveSubagent.id;
          item.subagentName = effectiveSubagent.name || item.subagentName || 'Subagent';
          // Namespace ID with subagent prefix to avoid cross-subagent collisions
          if (!item.id.startsWith(`${effectiveSubagent.id}:`)) {
            item.id = `${effectiveSubagent.id}:${item.id}`;
          }
        }
      }

      const existing = this.sessionPlans[sessionId];
      if (existing && existing.todos && existing.todos.length > 0) {
        const mergedMap = new Map<string, TodoItem>();
        for (const item of existing.todos) {
          mergedMap.set(item.id, { ...item });
        }

        if (effectiveSubagent) {
          // If updating a subagent plan
          if (!isMerge) {
            // Clear prior todos for this specific subagent
            for (const [id, item] of mergedMap.entries()) {
              if (item.subagentId === effectiveSubagent.id) {
                mergedMap.delete(id);
              }
            }
          }

          for (const update of todos) {
            const prev = mergedMap.get(update.id);
            if (prev) {
              mergedMap.set(update.id, {
                ...prev,
                ...update,
                status: update.status || prev.status,
                content: update.content || prev.content,
                subagentId: effectiveSubagent.id,
                subagentName: effectiveSubagent.name || prev.subagentName,
                updatedAt: Date.now()
              });
            } else {
              mergedMap.set(update.id, update);
            }
          }

          this.updatePlan(sessionId, Array.from(mergedMap.values()), toolId, toolName);
        } else {
          // Updating main / orchestrator plan
          if (isMerge && isTodoWrite) {
            for (const update of todos) {
              const prev = mergedMap.get(update.id);
              if (prev) {
                mergedMap.set(update.id, {
                  ...prev,
                  ...update,
                  status: update.status || prev.status,
                  content: update.content || prev.content,
                  updatedAt: Date.now()
                });
              } else {
                mergedMap.set(update.id, update);
              }
            }
            this.updatePlan(sessionId, Array.from(mergedMap.values()), toolId, toolName);
          } else {
            // Full replacement of main todos, preserving active subagent plans
            const subagentExisting = Array.from(mergedMap.values()).filter(t => Boolean(t.subagentId));
            this.updatePlan(sessionId, [...todos, ...subagentExisting], toolId, toolName);
          }
        }
      } else {
        // Complete replacement / initial session plan
        this.updatePlan(sessionId, todos, toolId, toolName);
      }
      return true;
    }
    return false;
  }

  /**
   * Clear unread badge / notification status
   */
  clearUpdateFlag() {
    this.hasNewUpdate = false;
  }

  /**
   * Toggle floating plan widget visibility
   */
  toggleFloatingVisible() {
    this.isFloatingVisible = !this.isFloatingVisible;
    if (this.isFloatingVisible) {
      this.hasNewUpdate = false;
    }
  }

  /**
   * Toggle between compact pill and expanded card
   */
  toggleCompact() {
    this.isCompact = !this.isCompact;
    this.hasNewUpdate = false;
  }

  /**
   * Reset or clear plan for a session (e.g. on new prompt submission or deleted session)
   */
  resetPlan(sessionId: string) {
    if (this.sessionPlans[sessionId]) {
      const copy = { ...this.sessionPlans };
      delete copy[sessionId];
      this.sessionPlans = copy;
      this.hasNewUpdate = false;
      if (this.lastUpdatedSessionId === sessionId) {
        this.lastUpdatedSessionId = null;
      }
    }
  }

  /**
   * Reset or clear plan for a deleted session
   */
  removeSessionPlan(sessionId: string) {
    this.resetPlan(sessionId);
  }
}

export const planStore = new PlanStore();
