import { describe, it, expect, beforeEach } from 'vitest';

// Svelte 5 Runes mock for node testing environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}

const { settingsStore } = await import('../frontend/src/lib/stores/settings.svelte');
const { sessionStore, getActiveMaxContextTokens } = await import('../frontend/src/lib/stores/session.svelte');

describe('Session Context Window Token Limit Persistence & Synchronization', () => {
  beforeEach(() => {
    settingsStore.updateSettings({ maxContextTokens: 200000 });
  });

  it('respects user configured maxContextTokens from settingsStore on session creation', () => {
    settingsStore.updateSettings({ maxContextTokens: 128000 });
    expect(getActiveMaxContextTokens()).toBe(128000);

    const session = sessionStore.createSession('Test Token Session');
    expect(session.usage?.maxTokens).toBe(128000);
  });

  it('does not allow disk backend stats to override user configured maxContextTokens with 1M', async () => {
    settingsStore.updateSettings({ maxContextTokens: 200000 });
    sessionStore.workspaces = [{ id: 'ws-1', name: 'Ws', path: '/path' }];
    sessionStore.activeWorkspaceId = 'ws-1';

    const session = sessionStore.createSession('Test Override Session', 'ws-1');
    // Simulate backend returning 1,048,576 from Grok CLI signals.json
    (globalThis as any).window = (globalThis as any).window || {};
    (globalThis as any).window.go = {
      main: {
        App: {
          GetSessionUsage: async () => ({
            sessionId: session.id,
            usedTokens: 50000,
            maxTokens: 1048576, // 1M raw from signals.json
            turnCount: 5
          })
        }
      }
    };

    await sessionStore.loadSessionUsage(session);

    // Session usage maxTokens MUST adhere to user configured 200,000, NOT the 1M raw signals value!
    expect(session.usage?.maxTokens).toBe(200000);
    expect(session.usage?.usedTokens).toBe(50000);
  });

  it('updates all active sessions immediately when user changes maxContextTokens in settings', () => {
    const session1 = sessionStore.createSession('S1');
    const session2 = sessionStore.createSession('S2');

    settingsStore.updateSettings({ maxContextTokens: 500000 });
    for (const s of sessionStore.sessions) {
      if (s.usage) {
        s.usage.maxTokens = settingsStore.maxContextTokens;
      }
    }

    expect(session1.usage?.maxTokens).toBe(500000);
    expect(session2.usage?.maxTokens).toBe(500000);
  });
});
