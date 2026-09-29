import { describe, it, expect, beforeEach, afterEach, vi, beforeAll } from 'vitest';

// Define Svelte 5 runes mock on globalThis for node test environment if not defined
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

// Dynamically import after defining runes mocks on globalThis
const { InputShieldStore } = await import('../stores/inputShield.svelte');

describe('InputShieldStore', () => {
  let shield: InstanceType<typeof InputShieldStore>;

  beforeEach(() => {
    vi.useFakeTimers();
    shield = new InputShieldStore();
  });

  afterEach(() => {
    shield.resetAll();
    vi.restoreAllMocks();
  });

  it('starts with isReadOnly as false and empty active locks', () => {
    expect(shield.isReadOnly).toBe(false);
    expect(shield.getActiveLockReasons()).toEqual([]);
    expect(shield.hasLock('snapshot_key_hold')).toBe(false);
  });

  it('acquires and releases a single lock correctly', () => {
    shield.acquireLock('snapshot_key_hold');
    expect(shield.isReadOnly).toBe(true);
    expect(shield.hasLock('snapshot_key_hold')).toBe(true);
    expect(shield.getActiveLockReasons()).toContain('snapshot_key_hold');

    shield.releaseLock('snapshot_key_hold');
    expect(shield.isReadOnly).toBe(false);
    expect(shield.hasLock('snapshot_key_hold')).toBe(false);
    expect(shield.getActiveLockReasons()).toEqual([]);
  });

  it('handles multiple overlapping locks independently', () => {
    shield.acquireLock('snapshot_key_hold');
    shield.acquireLock('snapshot_capturing');

    expect(shield.isReadOnly).toBe(true);
    expect(shield.getActiveLockReasons().length).toBe(2);

    // Releasing one lock keeps the shield active for the remaining lock
    shield.releaseLock('snapshot_key_hold');
    expect(shield.isReadOnly).toBe(true);
    expect(shield.hasLock('snapshot_capturing')).toBe(true);

    // Releasing the final lock turns off read-only mode
    shield.releaseLock('snapshot_capturing');
    expect(shield.isReadOnly).toBe(false);
    expect(shield.getActiveLockReasons()).toEqual([]);
  });

  it('auto-releases locks when timeout expires (safety timeout mechanism)', () => {
    shield.acquireLock('dictation_recording', 1000);
    expect(shield.isReadOnly).toBe(true);

    // Advance 500ms: still active
    vi.advanceTimersByTime(500);
    expect(shield.isReadOnly).toBe(true);
    expect(shield.hasLock('dictation_recording')).toBe(true);

    // Advance past 1000ms: automatically released
    vi.advanceTimersByTime(501);
    expect(shield.isReadOnly).toBe(false);
    expect(shield.hasLock('dictation_recording')).toBe(false);
  });

  it('resets timer when acquiring the same lock reason again', () => {
    shield.acquireLock('test_lock', 1000);
    vi.advanceTimersByTime(600);

    // Re-acquire extends timeout for another 1000ms
    shield.acquireLock('test_lock', 1000);
    vi.advanceTimersByTime(600); // 1200ms from start, but only 600ms from re-acquire
    expect(shield.isReadOnly).toBe(true);

    vi.advanceTimersByTime(401); // 1001ms from re-acquire
    expect(shield.isReadOnly).toBe(false);
  });

  it('resetAll clears all active locks and cancels all running timers immediately', () => {
    shield.acquireLock('snapshot_key_hold', 5000);
    shield.acquireLock('dictation_recording', 10000);
    shield.acquireLock('dictation_transcribing', 15000);

    expect(shield.isReadOnly).toBe(true);
    expect(shield.getActiveLockReasons().length).toBe(3);

    shield.resetAll();

    expect(shield.isReadOnly).toBe(false);
    expect(shield.getActiveLockReasons()).toEqual([]);

    // Timers should not fire any unexpected state changes after reset
    vi.advanceTimersByTime(20000);
    expect(shield.isReadOnly).toBe(false);
  });

  it('safely handles empty string and non-existent lock operations without throwing', () => {
    expect(() => shield.acquireLock('')).not.toThrow();
    expect(shield.isReadOnly).toBe(false);

    expect(() => shield.releaseLock('')).not.toThrow();
    expect(() => shield.releaseLock('non_existent')).not.toThrow();
    expect(shield.isReadOnly).toBe(false);
  });
});
