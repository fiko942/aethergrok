import { describe, it, expect, vi } from 'vitest';

// Svelte 5 Runes mock for node testing environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const {
  SessionStore
} = await import('../frontend/src/lib/stores/session.svelte');

describe('Background Delta Reactivity & Turn Preservation', () => {
  it('updates background session content and refreshes session.updatedAt', () => {
    const store = new SessionStore();
    store.activeSessionId = 'active-session';

    const backgroundSession: any = {
      id: 'bg-session',
      title: 'Background Task',
      workspaceId: 'ws-1',
      status: 'working',
      createdAt: 1000,
      updatedAt: 1000,
      messages: [
        {
          id: 'm1',
          role: 'user',
          content: 'Run task in background',
          timestamp: 1000,
          status: 'done'
        },
        {
          id: 'm2',
          role: 'assistant',
          content: 'Starting',
          timestamp: 1010,
          status: 'streaming',
          parts: [{ type: 'text', text: 'Starting' }]
        }
      ]
    };

    store.sessions = [backgroundSession];

    const initialParts = backgroundSession.messages[1].parts;
    const initialUpdatedAt = backgroundSession.updatedAt;

    const nowSpy = vi.spyOn(Date, 'now').mockReturnValue(2000);

    store.appendDelta('bg-session', ' step 1', 'assistant');

    const updatedMsg = backgroundSession.messages[1];
    expect(updatedMsg.content).toBe('Starting step 1');
    expect(updatedMsg.parts).toEqual([{ type: 'text', text: 'Starting step 1' }]);
    // Parts array reference preserved in background to save allocations
    expect(updatedMsg.parts).toBe(initialParts);
    // session.updatedAt must be refreshed
    expect(backgroundSession.updatedAt).toBe(2000);
    expect(backgroundSession.updatedAt).not.toBe(initialUpdatedAt);

    nowSpy.mockRestore();
  });

  it('clones parts array reference when activating session in tab to trigger Svelte 5 reactivity', async () => {
    const store = new SessionStore();
    store.saveSessionsToStorage = vi.fn();
    store.loadSessionHistoryFromDisk = vi.fn().mockResolvedValue(undefined);

    const initialParts = [{ type: 'text', text: 'Background stream text' }];
    const session: any = {
      id: 'bg-sess',
      title: 'Background Session',
      workspaceId: 'ws-1',
      status: 'working',
      createdAt: 1000,
      updatedAt: 1000,
      messages: [
        { id: 'm1', role: 'user', content: 'hello', timestamp: 1000, status: 'done' },
        { id: 'm2', role: 'assistant', content: 'Background stream text', timestamp: 1010, status: 'streaming', parts: initialParts }
      ],
      visibleTurnCount: 4
    };

    store.sessions = [session];
    store.activeSessionId = 'other-sess';

    await store.openSessionInTab('bg-sess');

    const activeMsg = session.messages[1];
    // When opened in tab, parts array is cloned to trigger Svelte 5 re-render
    expect(activeMsg.parts).not.toBe(initialParts);
    expect(activeMsg.parts).toEqual(initialParts);
  });

  it('openSessionInTab preserves existing user-expanded visibleTurnCount', async () => {
    const store = new SessionStore();
    store.saveSessionsToStorage = vi.fn();
    store.loadSessionHistoryFromDisk = vi.fn().mockResolvedValue(undefined);

    const targetSession: any = {
      id: 'session-expanded',
      title: 'Expanded Session',
      workspaceId: 'ws-1',
      status: 'idle',
      createdAt: 1000,
      updatedAt: 1000,
      messages: [{ id: 'm1', role: 'user', content: 'test', timestamp: 1000, status: 'done' }],
      visibleTurnCount: 25
    };

    store.sessions = [targetSession];

    await store.openSessionInTab('session-expanded');

    // visibleTurnCount should be preserved at 25 and not reset to default active window turns
    expect(targetSession.visibleTurnCount).toBe(25);
  });
});
