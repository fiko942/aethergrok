import { describe, it, expect, vi } from 'vitest';

// Svelte 5 Runes mock for node testing environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const {
  SessionStore,
  sanitizeSessionForStorage
} = await import('../frontend/src/lib/stores/session.svelte');

describe('Multi-Session Performance & Memory Simulation', () => {
  it('maintains lean serialization size across 10 concurrent sessions', () => {
    const store = new SessionStore();

    // Simulate 10 sessions with 20 turns each and large outputs
    for (let i = 1; i <= 10; i++) {
      const sess = store.createSession(`Stress Test Session ${i}`, 'ws-1');
      for (let j = 0; j < 20; j++) {
        store.addMessage(sess.id, {
          role: 'user',
          content: `User query ${j}`,
          status: 'done'
        });
        store.addMessage(sess.id, {
          role: 'assistant',
          content: `Assistant response ${j} with long tool outputs: ` + 'X'.repeat(5000),
          status: 'done'
        });
      }
    }

    expect(store.sessions.length).toBe(10);

    // Verify localStorage serialization payload size is < 15 KB (metadata only)
    const serialized = JSON.stringify(store.sessions.map((s: any) => sanitizeSessionForStorage(s)));
    expect(serialized.length).toBeLessThan(15000); // Less than 15 KB instead of megabytes!
  });
});
