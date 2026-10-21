import { describe, it, expect } from 'vitest';

// Svelte 5 Runes mock for node testing environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const {
  mergeSessionHistoryWithUncommittedTurns
} = await import('../frontend/src/lib/stores/session.svelte');

describe('mergeSessionHistoryWithUncommittedTurns', () => {
  it('cleanly adopts loadedHistory when existingMessages is empty or undefined', () => {
    const loadedHistory = [
      { id: 'm1', role: 'user', content: 'hello', timestamp: 1000, status: 'done' },
      { id: 'm2', role: 'assistant', content: 'hi', timestamp: 1100, status: 'done' }
    ] as any[];

    expect(mergeSessionHistoryWithUncommittedTurns([], loadedHistory)).toEqual(loadedHistory);
    expect(mergeSessionHistoryWithUncommittedTurns(undefined as any, loadedHistory)).toEqual(loadedHistory);
  });

  it('cleanly preserves existingMessages when loadedHistory is empty or undefined', () => {
    const existingMessages = [
      { id: 'm1', role: 'user', content: 'hello', timestamp: 1000, status: 'done' }
    ] as any[];

    expect(mergeSessionHistoryWithUncommittedTurns(existingMessages, [])).toEqual(existingMessages);
    expect(mergeSessionHistoryWithUncommittedTurns(existingMessages, undefined as any)).toEqual(existingMessages);
  });

  it('returns loadedHistory when existingUserCount is less than or equal to loadedUserCount', () => {
    const existingMessages = [
      { id: 'm1', role: 'user', content: 'hello', timestamp: 1000, status: 'done' },
      { id: 'm2', role: 'assistant', content: 'hi', timestamp: 1100, status: 'done' }
    ] as any[];
    const loadedHistory = [
      { id: 'm1-disk', role: 'user', content: 'hello', timestamp: 1000, status: 'done' },
      { id: 'm2-disk', role: 'assistant', content: 'hi formatted', timestamp: 1100, status: 'done' }
    ] as any[];

    const result = mergeSessionHistoryWithUncommittedTurns(existingMessages, loadedHistory);
    expect(result).toEqual(loadedHistory);
  });

  it('preserves cancelled turn when disk history only contains turns prior to the cancelled turn', () => {
    // Turn 1 on disk and in memory
    // Turn 2 was started, then cancelled by user (never committed to grok disk session)
    const existingMessages = [
      { id: 'm1', role: 'user', content: 'First prompt', timestamp: 1000, status: 'done' },
      { id: 'm2', role: 'assistant', content: 'First response', timestamp: 1100, status: 'done' },
      { id: 'm3', role: 'user', content: 'Cancelled prompt', timestamp: 2000, status: 'done' },
      { id: 'm4', role: 'assistant', content: 'Working...\n\n*(Turn cancelled by user)*', timestamp: 2100, status: 'done' }
    ] as any[];

    const loadedHistory = [
      { id: 'disk_m1', role: 'user', content: 'First prompt', timestamp: 1000, status: 'done' },
      { id: 'disk_m2', role: 'assistant', content: 'First response', timestamp: 1100, status: 'done' }
    ] as any[];

    const merged = mergeSessionHistoryWithUncommittedTurns(existingMessages, loadedHistory);

    expect(merged).toHaveLength(4);
    expect(merged[0].id).toBe('disk_m1');
    expect(merged[1].id).toBe('disk_m2');
    expect(merged[2].id).toBe('m3');
    expect(merged[2].content).toBe('Cancelled prompt');
    expect(merged[3].id).toBe('m4');
    expect(merged[3].content).toContain('*(Turn cancelled by user)*');
  });

  it('preserves multiple uncommitted trailing turns when existingMessages has multiple newer turns', () => {
    const existingMessages = [
      { id: 'm1', role: 'user', content: 'Turn 1', timestamp: 1000, status: 'done' },
      { id: 'm2', role: 'assistant', content: 'Resp 1', timestamp: 1100, status: 'done' },
      { id: 'm3', role: 'user', content: 'Turn 2 uncommitted', timestamp: 2000, status: 'done' },
      { id: 'm4', role: 'assistant', content: 'Resp 2 uncommitted', timestamp: 2100, status: 'done' },
      { id: 'm5', role: 'user', content: 'Turn 3 uncommitted', timestamp: 3000, status: 'done' },
      { id: 'm6', role: 'assistant', content: 'Resp 3 uncommitted', timestamp: 3100, status: 'done' }
    ] as any[];

    const loadedHistory = [
      { id: 'disk_m1', role: 'user', content: 'Turn 1', timestamp: 1000, status: 'done' },
      { id: 'disk_m2', role: 'assistant', content: 'Resp 1', timestamp: 1100, status: 'done' }
    ] as any[];

    const merged = mergeSessionHistoryWithUncommittedTurns(existingMessages, loadedHistory);

    expect(merged).toHaveLength(6);
    expect(merged[0].id).toBe('disk_m1');
    expect(merged[1].id).toBe('disk_m2');
    expect(merged[2].id).toBe('m3');
    expect(merged[3].id).toBe('m4');
    expect(merged[4].id).toBe('m5');
    expect(merged[5].id).toBe('m6');
  });
});
