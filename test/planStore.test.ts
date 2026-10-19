import { describe, it, expect, beforeEach } from 'vitest';

// Define Svelte 5 runes mock on globalThis for node test environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const { PlanStore } = await import('../frontend/src/lib/stores/plan.svelte');
import type { TodoItem } from '../frontend/src/lib/utils/planParser';

describe('planStore', () => {
  let store: InstanceType<typeof PlanStore>;

  beforeEach(() => {
    store = new PlanStore();
  });

  it('updates plan and computes metrics', () => {
    const todos: TodoItem[] = [
      { id: '1', content: 'Design architecture', status: 'completed' },
      { id: '2', content: 'Implement UI components', status: 'in_progress' },
      { id: '3', content: 'Write tests', status: 'pending' }
    ];

    const plan = store.updatePlan('sess-1', todos, 'tool-1', 'write');
    expect(plan.totalCount).toBe(3);
    expect(plan.completedCount).toBe(1);
    expect(plan.inProgressCount).toBe(1);
    expect(plan.pendingCount).toBe(1);
    expect(plan.progressPercent).toBe(33);
    expect(store.hasNewUpdate).toBe(true);
    expect(store.lastUpdatedSessionId).toBe('sess-1');

    const fetched = store.getPlan('sess-1');
    expect(fetched?.totalCount).toBe(3);
  });

  it('inspectToolForPlan correctly ingests TodosUpdated payload', () => {
    const rawResult = JSON.stringify({
      TodosUpdated: {
        state: {
          todos: {
            'task-a': { content: 'Setup database schema', status: 'completed' },
            'task-b': { content: 'Run migrations', status: 'in_progress' }
          }
        }
      }
    });

    const ingested = store.inspectToolForPlan('sess-2', 'tool-write', 'write', {}, rawResult);
    expect(ingested).toBe(true);

    const plan = store.getPlan('sess-2');
    expect(plan).not.toBeNull();
    expect(plan?.todos.length).toBe(2);
    expect(plan?.completedCount).toBe(1);
    expect(plan?.inProgressCount).toBe(1);
    expect(plan?.progressPercent).toBe(50);
  });

  it('toggles compact mode and clears flags', () => {
    expect(store.isCompact).toBe(false);
    store.hasNewUpdate = true;
    store.toggleCompact();
    expect(store.isCompact).toBe(true);
    expect(store.hasNewUpdate).toBe(false);
  });

  it('removes session plan cleanly', () => {
    store.updatePlan('sess-del', [{ id: '1', content: 'Task', status: 'pending' }]);
    expect(store.getPlan('sess-del')).not.toBeNull();
    store.removeSessionPlan('sess-del');
    expect(store.getPlan('sess-del')).toBeNull();
  });

  it('attributes plans per subagent and isolates them from main plan', () => {
    // 1. Main plan has 2 tasks
    store.inspectToolForPlan('sess-sub', 'tool-main', 'todo_write', {
      todos: [
        { id: 'main-1', content: 'Main Task 1', status: 'in_progress' },
        { id: 'main-2', content: 'Main Task 2', status: 'pending' },
      ],
      merge: false
    }, null);

    let plan = store.getPlan('sess-sub');
    expect(plan?.mainTodos).toHaveLength(2);
    expect(store.getSubagentPlans('sess-sub')).toHaveLength(0);

    // 2. Subagent 1 writes its own plan (Step 1, Step 2)
    store.inspectToolForPlan('sess-sub', 'call_sub1', 'spawn_subagent', {
      description: 'Subagent A: Fix Bug',
      prompt: 'Please fix bug',
      task_id: 'sub-uuid-1'
    }, '- [x] Step 1: Write test\n- [ ] Step 2: Implement');

    // 3. Subagent 2 writes its own plan with identical step names (Step 1, Step 2)
    store.inspectToolForPlan('sess-sub', 'call_sub2', 'spawn_subagent', {
      description: 'Subagent B: Add Feature',
      prompt: 'Please add feature',
      task_id: 'sub-uuid-2'
    }, '- [x] Step 1: Write test\n- [x] Step 2: Implement');

    plan = store.getPlan('sess-sub');
    expect(plan).not.toBeNull();
    // Total todos: 2 main + 2 subagent 1 + 2 subagent 2 = 6
    expect(plan?.todos).toHaveLength(6);
    expect(plan?.mainTodos).toHaveLength(2);

    // Subagent plans list
    const subPlans = store.getSubagentPlans('sess-sub');
    expect(subPlans).toHaveLength(2);

    const sub1 = store.getSubagentPlan('sess-sub', 'sub-uuid-1');
    expect(sub1).not.toBeNull();
    expect(sub1?.subagentName).toBe('Subagent A: Fix Bug');
    expect(sub1?.todos).toHaveLength(2);
    expect(sub1?.completedCount).toBe(1);
    expect(sub1?.pendingCount).toBe(1);
    expect(sub1?.progressPercent).toBe(50);

    const sub2 = store.getSubagentPlan('sess-sub', 'sub-uuid-2');
    expect(sub2).not.toBeNull();
    expect(sub2?.subagentName).toBe('Subagent B: Add Feature');
    expect(sub2?.todos).toHaveLength(2);
    expect(sub2?.completedCount).toBe(2);
    expect(sub2?.pendingCount).toBe(0);
    expect(sub2?.progressPercent).toBe(100);

    // 4. Main plan updates with merge: false should preserve subagent plans
    store.inspectToolForPlan('sess-sub', 'tool-main-2', 'todo_write', {
      todos: [
        { id: 'main-1', content: 'Main Task 1', status: 'completed' },
        { id: 'main-2', content: 'Main Task 2', status: 'completed' },
        { id: 'main-3', content: 'Main Task 3', status: 'in_progress' },
      ],
      merge: false
    }, null);

    const updatedPlan = store.getPlan('sess-sub');
    expect(updatedPlan?.mainTodos).toHaveLength(3);
    expect(store.getSubagentPlans('sess-sub')).toHaveLength(2);
    expect(store.getSubagentPlan('sess-sub', 'sub-uuid-1')?.todos).toHaveLength(2);
    expect(store.getSubagentPlan('sess-sub', 'sub-uuid-2')?.todos).toHaveLength(2);
  });
});
