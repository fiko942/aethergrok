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
});
