import { describe, it, expect, afterEach } from 'vitest';

// Svelte 5 Runes mock for node testing environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const {
  SessionStore,
  sanitizeSessionForStorage,
  shouldEvictSessionMessages,
  evictInactiveSessionMessages
} = await import('../frontend/src/lib/stores/session.svelte');

describe('Session Memory & Storage Optimization', () => {
  const originalWindow = (globalThis as any).window;

  afterEach(() => {
    (globalThis as any).window = originalWindow;
  });
  it('strips heavy messages and tool outputs when serializing session metadata for localStorage', () => {
    const heavySession = {
      id: 'sess-123',
      title: 'Heavy Session',
      workspaceId: 'ws-1',
      createdAt: 1000,
      updatedAt: 2000,
      status: 'idle' as const,
      agentMode: 'agent' as const,
      grokSessionId: 'grok-uuid-456',
      visibleTurnCount: 4,
      messages: [
        {
          id: 'm1',
          role: 'user' as const,
          content: 'Run heavy command',
          timestamp: 1000,
          status: 'done' as const
        },
        {
          id: 'm2',
          role: 'assistant' as const,
          content: 'Here is the huge output',
          timestamp: 1500,
          status: 'done' as const,
          toolCalls: [
            {
              id: 'tc1',
              tool: 'run_terminal_command',
              status: 'completed' as const,
              result: 'A'.repeat(500000) // 500 KB result
            }
          ]
        }
      ]
    };

    const sanitized = sanitizeSessionForStorage(heavySession as any);

    expect(sanitized.id).toBe('sess-123');
    expect(sanitized.title).toBe('Heavy Session');
    expect(sanitized.grokSessionId).toBe('grok-uuid-456');
    // Messages must not be serialized into localStorage
    expect((sanitized as any).messages).toBeUndefined();
  });

  it('identifies inactive completed sessions eligible for in-memory message eviction', () => {
    const activeWorkingSession = { id: 's1', status: 'working' as const };
    const inactiveWorkingSession = { id: 's2', status: 'working' as const };
    const inactiveIdleSession = { id: 's3', status: 'idle' as const };
    const activeIdleSession = { id: 's4', status: 'idle' as const };

    // Active session should never be evicted
    expect(shouldEvictSessionMessages(activeWorkingSession.id, 's1', activeWorkingSession.status)).toBe(false);
    expect(shouldEvictSessionMessages(activeIdleSession.id, 's4', activeIdleSession.status)).toBe(false);

    // Inactive working session must keep messages in memory while streaming
    expect(shouldEvictSessionMessages(inactiveWorkingSession.id, 's1', inactiveWorkingSession.status)).toBe(false);

    // Inactive idle session can safely be evicted to save RAM (rehydrated on switch)
    expect(shouldEvictSessionMessages(inactiveIdleSession.id, 's1', inactiveIdleSession.status)).toBe(true);
  });

  it('never evicts messages from sessions that are currently open in tabs', () => {
    const openTabIds = ['s1', 's2'];
    const activeSessionId = 's1';

    // s2 is an open inactive tab with idle status
    expect(shouldEvictSessionMessages('s2', activeSessionId, 'idle', openTabIds)).toBe(false);

    // s3 is closed and inactive, making it eligible for eviction
    expect(shouldEvictSessionMessages('s3', activeSessionId, 'idle', openTabIds)).toBe(true);
  });

  it('evicts messages from inactive non-working sessions', () => {
    const sessions = [
      { id: 's1', status: 'working' as const, messages: [{ id: 'm1' }] },
      { id: 's2', status: 'idle' as const, messages: [{ id: 'm2' }] },
      { id: 's3', status: 'idle' as const, messages: [{ id: 'm3' }] },
    ] as any[];

    evictInactiveSessionMessages(sessions, 's2');

    expect(sessions[0].messages.length).toBe(1); // s1 is working -> kept
    expect(sessions[1].messages.length).toBe(1); // s2 is active -> kept
    expect(sessions[2].messages.length).toBe(0); // s3 is inactive idle -> evicted
  });

  it('saveSessionsToStorage persists only metadata to localStorage without messages', () => {
    const storageMap = new Map<string, string>();
    const fakeLocalStorage = {
      getItem: (k: string) => storageMap.get(k) ?? null,
      setItem: (k: string, v: string) => storageMap.set(k, v),
      removeItem: (k: string) => storageMap.delete(k)
    };
    (globalThis as any).window = { localStorage: fakeLocalStorage };

    const store = new SessionStore();
    store.sessions = [
      {
        id: 'sess-1',
        title: 'Title 1',
        workspaceId: 'ws-1',
        createdAt: 100,
        updatedAt: 200,
        status: 'idle',
        messages: [{ id: 'm1', content: 'Huge text' }] as any,
        visibleTurnCount: 4
      }
    ];

    store.saveSessionsToStorage();

    const storedRaw = storageMap.get('aethergrok_sessions_v1');
    expect(storedRaw).toBeDefined();
    const parsed = JSON.parse(storedRaw!);
    expect(parsed).toHaveLength(1);
    expect(parsed[0].id).toBe('sess-1');
    expect(parsed[0].title).toBe('Title 1');
    expect(parsed[0].messages).toBeUndefined();
  });

  it('switchSession triggers on-demand history load and evicts inactive sessions', async () => {
    const store = new SessionStore();
    let loadedSessionId = '';
    store.loadSessionHistoryFromDisk = async (session: any) => {
      loadedSessionId = session.id;
      session.messages = [{ id: 'loaded-m1', content: 'Restored from disk' }];
    };

    store.sessions = [
      {
        id: 's-active',
        title: 'Active Session',
        workspaceId: 'ws-1',
        status: 'idle',
        messages: [{ id: 'm-active', content: 'Active text' }] as any,
        createdAt: 100,
        updatedAt: 100,
        visibleTurnCount: 4
      },
      {
        id: 's-target',
        title: 'Target Session',
        workspaceId: 'ws-1',
        status: 'idle',
        messages: [] as any,
        createdAt: 200,
        updatedAt: 200,
        visibleTurnCount: 4
      },
      {
        id: 's-closed',
        title: 'Closed Session',
        workspaceId: 'ws-1',
        status: 'idle',
        messages: [{ id: 'm-closed', content: 'Closed tab text' }] as any,
        createdAt: 300,
        updatedAt: 300,
        visibleTurnCount: 4
      }
    ];
    store.activeSessionId = 's-active';
    store.openTabSessionIds = ['s-active', 's-target'];

    await store.switchSession('s-target');

    expect(loadedSessionId).toBe('s-target');
    expect(store.activeSessionId).toBe('s-target');
    // s-target should have messages loaded
    expect(store.sessions.find(s => s.id === 's-target')?.messages.length).toBe(1);
    // s-active remains open in tabs, so messages remain in memory
    expect(store.sessions.find(s => s.id === 's-active')?.messages.length).toBe(1);
    // s-closed is closed and idle, so messages are evicted to conserve memory
    expect(store.sessions.find(s => s.id === 's-closed')?.messages.length).toBe(0);
  });
});
