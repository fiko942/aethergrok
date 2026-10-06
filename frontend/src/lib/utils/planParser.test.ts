import { describe, it, expect, beforeEach } from 'vitest';
import { parseTodosUpdated, calculatePlanMetrics, deduplicateTodos, type TodoItem } from './planParser';

// Define Svelte 5 runes mock on globalThis for node test environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const { PlanStore } = await import('../stores/plan.svelte');

describe('planParser', () => {
  it('parses todo_write parameters correctly', () => {
    const params = {
      todos: [
        { id: 'step-1', content: 'Step 1 description', status: 'completed' },
        { id: 'step-2', content: 'Step 2 description', status: 'in_progress' },
        { id: 'step-3', content: 'Step 3 description', status: 'pending' },
      ],
      merge: false,
    };

    const parsed = parseTodosUpdated(params);
    expect(parsed).toHaveLength(3);
    expect(parsed?.[0].id).toBe('step-1');
    expect(parsed?.[0].status).toBe('completed');
    expect(parsed?.[1].status).toBe('in_progress');
    expect(parsed?.[2].status).toBe('pending');
  });

  it('parses markdown checkbox list output from CLI', () => {
    const raw = `- [completed] check_status: Check current git status
- [in_progress] verify_tests: Verify Go tests pass
- [pending] git_commit_push: Stage, commit, and push all changes`;

    const parsed = parseTodosUpdated(raw);
    expect(parsed).toHaveLength(3);
    expect(parsed?.[0].id).toBe('check_status');
    expect(parsed?.[0].status).toBe('completed');
    expect(parsed?.[0].content).toBe('Check current git status');
    expect(parsed?.[1].id).toBe('verify_tests');
    expect(parsed?.[1].status).toBe('in_progress');
    expect(parsed?.[2].id).toBe('git_commit_push');
    expect(parsed?.[2].status).toBe('pending');
  });

  it('deduplicates todos by id', () => {
    const todos: TodoItem[] = [
      { id: 'task-1', content: 'Old task 1', status: 'pending' },
      { id: 'task-2', content: 'Task 2', status: 'pending' },
      { id: 'task-1', content: 'Updated task 1', status: 'completed' },
    ];

    const deduplicated = deduplicateTodos(todos);
    expect(deduplicated).toHaveLength(2);
    expect(deduplicated.find((t) => t.id === 'task-1')?.status).toBe('completed');
  });
});

describe('PlanStore', () => {
  let store: InstanceType<typeof PlanStore>;

  beforeEach(() => {
    store = new PlanStore();
  });

  it('replaces all todos when merge is false', () => {
    const sessionId = 'session-1';

    // Initial plan
    store.inspectToolForPlan(
      sessionId,
      't1',
      'todo_write',
      {
        todos: [
          { id: 'old-1', content: 'Old 1', status: 'completed' },
          { id: 'old-2', content: 'Old 2', status: 'completed' },
        ],
        merge: false,
      },
      null
    );

    expect(store.getPlan(sessionId)?.todos).toHaveLength(2);

    // New prompt generates fresh plan with merge: false
    store.inspectToolForPlan(
      sessionId,
      't2',
      'todo_write',
      {
        todos: [
          { id: 'new-1', content: 'New Task', status: 'pending' },
        ],
        merge: false,
      },
      null
    );

    const updated = store.getPlan(sessionId);
    expect(updated?.todos).toHaveLength(1);
    expect(updated?.todos[0].id).toBe('new-1');
  });

  it('resets plan cleanly on session reset', () => {
    const sessionId = 'session-1';
    store.inspectToolForPlan(
      sessionId,
      't1',
      'todo_write',
      {
        todos: [{ id: 'task-1', content: 'Task 1', status: 'pending' }],
      },
      null
    );

    expect(store.getPlan(sessionId)).not.toBeNull();
    store.resetPlan(sessionId);
    expect(store.getPlan(sessionId)).toBeNull();
  });
});
