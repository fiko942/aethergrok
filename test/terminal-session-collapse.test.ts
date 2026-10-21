import { describe, it, expect, beforeEach } from 'vitest';

// Svelte 5 Runes mock for node testing environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}

const { TerminalStore } = await import('../frontend/src/lib/stores/terminal.svelte');

describe('Terminal Store Session Collapse Default & Tab Switching Behavior', () => {
  let store: InstanceType<typeof TerminalStore>;

  beforeEach(() => {
    store = new TerminalStore();
    store.sessionCollapsed = {};
  });

  it('defaults isSessionCollapsed to true for any new session', () => {
    expect(store.isSessionCollapsed('sess-new-1')).toBe(true);
    expect(store.isSessionCollapsed('sess-new-2')).toBe(true);
    expect(store.isSessionCollapsed('')).toBe(true);
  });

  it('keeps terminal open for a session once user explicitly opens it', () => {
    const sessionA = 'sess-a';
    const sessionB = 'sess-b';

    // User opens terminal in Session A
    store.toggleSessionCollapse(sessionA, false);
    expect(store.isSessionCollapsed(sessionA)).toBe(false);

    // Switching tabs to Session B (which user has not opened) defaults to collapsed/closed
    expect(store.isSessionCollapsed(sessionB)).toBe(true);

    // Switching back to Session A preserves the open state
    expect(store.isSessionCollapsed(sessionA)).toBe(false);
  });

  it('keeps terminal closed if user explicitly closes it', () => {
    const sessionA = 'sess-a';

    // User opens then closes Session A
    store.toggleSessionCollapse(sessionA, false);
    expect(store.isSessionCollapsed(sessionA)).toBe(false);

    store.toggleSessionCollapse(sessionA, true);
    expect(store.isSessionCollapsed(sessionA)).toBe(true);
  });
});
