// @vitest-environment happy-dom
import { describe, it, expect, vi } from 'vitest';

// Define Svelte 5 runes mock on globalThis for test environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const { loggerStore } = await import('../frontend/src/lib/stores/logger.svelte');

describe('Logger Filter for Benign Engine Errors', () => {
  it('suppresses ResizeObserver loop notifications from logging as SYSTEM error', () => {
    const errorSpy = vi.spyOn(loggerStore, 'error');

    // Dispatch error event mimicking ResizeObserver loop notification
    const evt = new ErrorEvent('error', {
      message: 'ResizeObserver loop completed with undelivered notifications.',
      filename: 'wails://wails/',
      lineno: 0,
      colno: 0
    });

    window.dispatchEvent(evt);

    // Should not log this benign notification
    expect(errorSpy).not.toHaveBeenCalledWith('SYSTEM', expect.stringContaining('ResizeObserver'), expect.anything());
    expect(errorSpy.mock.calls.some(call => call[0] === 'SYSTEM' && String(call[1]).includes('ResizeObserver'))).toBe(false);
    errorSpy.mockRestore();
  });

  it('also suppresses ResizeObserver loop limit exceeded notifications', () => {
    const errorSpy = vi.spyOn(loggerStore, 'error');

    const evt = new ErrorEvent('error', {
      message: 'ResizeObserver loop limit exceeded',
      filename: 'wails://wails/',
      lineno: 0,
      colno: 0
    });

    window.dispatchEvent(evt);

    expect(errorSpy).not.toHaveBeenCalledWith('SYSTEM', expect.stringContaining('ResizeObserver'), expect.anything());
    expect(errorSpy.mock.calls.some(call => call[0] === 'SYSTEM' && String(call[1]).includes('ResizeObserver'))).toBe(false);
    errorSpy.mockRestore();
  });

  it('still logs genuine uncaught window errors', () => {
    const errorSpy = vi.spyOn(loggerStore, 'error');

    const evt = new ErrorEvent('error', {
      message: 'Uncaught TypeError: Cannot read properties of undefined',
      filename: 'wails://wails/app.js',
      lineno: 42,
      colno: 10
    });

    window.dispatchEvent(evt);

    expect(errorSpy).toHaveBeenCalledWith('SYSTEM', expect.stringContaining('Uncaught TypeError'), expect.anything());
    errorSpy.mockRestore();
  });
});
