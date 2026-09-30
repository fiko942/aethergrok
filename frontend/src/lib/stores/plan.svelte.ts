import { parseTodosUpdated, calculatePlanMetrics, type TodoItem, type SessionPlanState } from '../utils/planParser';

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
   * Scan and ingest a tool call result or params to see if it contains plan updates
   */
  inspectToolForPlan(sessionId: string, toolId: string, toolName: string, params: unknown, result: unknown): boolean {
    // 1. Check result first
    let todos = parseTodosUpdated(result);
    // 2. Check params/input next (e.g. todo_write tool)
    if (!todos && params) {
      todos = parseTodosUpdated(params);
    }

    if (todos && todos.length > 0) {
      this.updatePlan(sessionId, todos, toolId, toolName);
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
