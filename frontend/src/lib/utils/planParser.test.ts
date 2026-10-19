import { describe, it, expect, beforeEach } from 'vitest';

// Define Svelte 5 runes mock on globalThis for node test environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

import {
  parseTodosUpdated,
  calculatePlanMetrics,
  deduplicateTodos,
  formatPlanUpdateSummary,
  resolveTodoContent,
  cleanTodoContent,
  renderTodoContent,
  type TodoItem
} from './planParser';
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

  it('cleans wrapping markdown bold/italic and noise from todo content', () => {
    expect(cleanTodoContent('**Step 1: Write the failing test**')).toBe('Step 1: Write the failing test');
    expect(cleanTodoContent('*Step 2: Run test to verify it fails*')).toBe('Step 2: Run test to verify it fails');
    expect(cleanTodoContent('__Step 3: Write minimal implementation__')).toBe('Step 3: Write minimal implementation');
    expect(cleanTodoContent('- [ ] **Step 4: Run test to verify it passes**')).toBe('Step 4: Run test to verify it passes');
    expect(cleanTodoContent('**Step 1** and **Step 2**')).toBe('**Step 1** and **Step 2**');
  });

  it('renders todo content with highlighted step label and clean inline markdown', () => {
    const rendered1 = renderTodoContent('**Step 1: Write the failing test**');
    expect(rendered1).toContain('font-semibold text-ant-text shrink-0');
    expect(rendered1).toContain('Step 1:');
    expect(rendered1).toContain('Write the failing test');
    expect(rendered1).not.toContain('**');

    const renderedCode = renderTodoContent('Step 2: Run `npm test` to verify');
    expect(renderedCode).toContain('<code>npm test</code>');

    const plain = renderTodoContent('Investigate bug in authentication pipeline');
    expect(plain).toBe('Investigate bug in authentication pipeline');
  });

  it('parses markdown plan checklists with section headings and attributes tasks', () => {
    const planMarkdown = `
# Realtime Channel Database Plan

### Task 1: Fix Database Query Semantics in Account.ts
- [x] **Step 1: Write the failing test for query semantics**
- [ ] **Step 2: Run test to verify it fails**

### Task 2: Browser Enforcement
- [ ] **Step 1: Write the failing test for browser**
- [ ] **Step 2: Run test to verify it passes**
`;

    const parsed = parseTodosUpdated(planMarkdown);
    expect(parsed).toHaveLength(4);

    expect(parsed?.[0].subagentId).toBe('task-1');
    expect(parsed?.[0].subagentName).toBe('Task 1: Fix Database Query Semantics in Account.ts');
    expect(parsed?.[0].content).toBe('Step 1: Write the failing test for query semantics');
    expect(parsed?.[0].status).toBe('completed');
    expect(parsed?.[0].priority).toBeUndefined();

    expect(parsed?.[2].subagentId).toBe('task-2');
    expect(parsed?.[2].subagentName).toBe('Task 2: Browser Enforcement');
    expect(parsed?.[2].content).toBe('Step 1: Write the failing test for browser');
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

  it('formats descriptive summary for plan creation and task completion', () => {
    const freshTodos: TodoItem[] = [
      { id: 't1', content: 'Investigate bug', status: 'in_progress' },
      { id: 't2', content: 'Implement fix', status: 'pending' },
      { id: 't3', content: 'Verify tests', status: 'pending' },
    ];

    // Non-merge creation
    const freshSummary = formatPlanUpdateSummary(freshTodos, false);
    expect(freshSummary).toBe('Created plan with 3 tasks');

    // Single task completion with session plan context
    const sessionPlan = calculatePlanMetrics(freshTodos);
    sessionPlan.completedCount = 1;
    sessionPlan.totalCount = 3;

    const singleDone = formatPlanUpdateSummary(
      [{ id: 't1', content: 'Investigate bug', status: 'completed' }],
      true,
      sessionPlan
    );
    expect(singleDone).toContain('Completed "Investigate bug"');
    expect(singleDone).toContain('1/3 completed');

    // Transition: one done, one started
    const transition = formatPlanUpdateSummary(
      [
        { id: 't1', content: 'Investigate bug', status: 'completed' },
        { id: 't2', content: 'Implement fix', status: 'in_progress' },
      ],
      true,
      sessionPlan
    );
    expect(transition).toContain('Completed "Investigate bug" → Started "Implement fix"');
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
