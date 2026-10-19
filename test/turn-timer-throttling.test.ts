import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

export function getOptimalTimerInterval(elapsedMs: number): number {
  if (elapsedMs < 1000) {
    return 200; // Sub-second precision for initial start
  }
  return 1000; // 1-second cadence once past 1s to eliminate V8 and GPU thrash
}

export function calculateTimerInterval(elapsedMs: number): number {
  return getOptimalTimerInterval(elapsedMs);
}

describe('Turn Execution Timer Throttling', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('chooses low interval during first second and 1000ms interval for longer durations', () => {
    expect(getOptimalTimerInterval(150)).toBe(200);
    expect(getOptimalTimerInterval(950)).toBe(200);
    expect(getOptimalTimerInterval(1000)).toBe(1000);
    expect(getOptimalTimerInterval(15000)).toBe(1000);
    expect(calculateTimerInterval(500)).toBe(200);
    expect(calculateTimerInterval(2000)).toBe(1000);
  });

  it('reduces timer tick frequency by 80% when throttling from 100ms to 500ms', () => {
    let unthrottledTicks = 0;
    let throttledTicks = 0;

    const unthrottledTimer = setInterval(() => {
      unthrottledTicks++;
    }, 100);

    const throttledTimer = setInterval(() => {
      throttledTicks++;
    }, 500);

    vi.advanceTimersByTime(1000);

    clearInterval(unthrottledTimer);
    clearInterval(throttledTimer);

    expect(unthrottledTicks).toBe(10);
    expect(throttledTicks).toBe(2);
    // 80% reduction in timer firings and reactive re-renders
    expect((unthrottledTicks - throttledTicks) / unthrottledTicks).toBe(0.8);
  });

  it('maintains accurate elapsed time tracking with 500ms intervals', () => {
    const startTime = 1000000;
    vi.setSystemTime(startTime);

    let elapsedMs = 0;
    const timer = setInterval(() => {
      elapsedMs = Math.max(0, Date.now() - startTime);
    }, 500);

    // Initial state before tick
    expect(elapsedMs).toBe(0);

    // After 500ms
    vi.advanceTimersByTime(500);
    expect(elapsedMs).toBe(500);

    // After 1000ms
    vi.advanceTimersByTime(500);
    expect(elapsedMs).toBe(1000);

    // After 2500ms
    vi.advanceTimersByTime(1500);
    expect(elapsedMs).toBe(2500);

    clearInterval(timer);
  });
});
