import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import { ShortcutDetector, type DictationTriggerEvent } from './shortcutDetector';

describe('ShortcutDetector', () => {
  let detector: ShortcutDetector;

  beforeEach(() => {
    vi.useFakeTimers();
    detector = new ShortcutDetector({
      holdThresholdMs: 300,
      doubleTapThresholdMs: 350
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  function createKeyEvent(opts: {
    key?: string;
    code?: string;
    metaKey?: boolean;
    ctrlKey?: boolean;
    shiftKey?: boolean;
    altKey?: boolean;
  }): KeyboardEvent {
    return {
      key: opts.key || '',
      code: opts.code || '',
      metaKey: opts.metaKey || false,
      ctrlKey: opts.ctrlKey || false,
      shiftKey: opts.shiftKey || false,
      altKey: opts.altKey || false,
      preventDefault: vi.fn(),
      stopPropagation: vi.fn()
    } as unknown as KeyboardEvent;
  }

  describe('Shortcut Matching', () => {
    it('matches Fn / Globe key in various browser representations', () => {
      expect(detector.matchesShortcut(createKeyEvent({ code: 'Backslash', key: '\\' }), '\\')).toBe(true);
      expect(detector.matchesShortcut(createKeyEvent({ key: 'a' }), '\\')).toBe(false);
    });

    it('matches single modifier keys like ShiftLeft or ControlRight', () => {
      expect(detector.matchesShortcut(createKeyEvent({ code: 'ShiftLeft' }), 'ShiftLeft')).toBe(true);
      expect(detector.matchesShortcut(createKeyEvent({ code: 'ShiftRight' }), 'RightShift')).toBe(true);
      expect(detector.matchesShortcut(createKeyEvent({ code: 'ControlRight' }), 'RightCtrl')).toBe(true);
    });

    it('matches modifier combinations (Cmd+D, Ctrl+Shift+D)', () => {
      expect(detector.matchesShortcut(createKeyEvent({ key: 'd', code: 'KeyD', metaKey: true }), 'Cmd+D')).toBe(true);
      expect(detector.matchesShortcut(createKeyEvent({ key: 'd', code: 'KeyD', ctrlKey: true, shiftKey: true }), 'Ctrl+Shift+D')).toBe(true);
      expect(detector.matchesShortcut(createKeyEvent({ key: 'd', code: 'KeyD', ctrlKey: true }), 'Ctrl+Shift+D')).toBe(false);
    });
  });

  describe('Hold / Push-to-Talk Detection (> 300ms)', () => {
    it('detects hold-start via timer or repeat after threshold and hold-release on keyup', () => {
      const fnKey = createKeyEvent({ key: 'Fn' });

      // Initial keydown
      const e1 = detector.feedKeyDown(fnKey, 'Fn');
      expect(e1).toBeNull();
      expect(detector.getIsHolding()).toBe(false);

      // Advance time beyond hold threshold (300ms) - hold timer fires
      vi.advanceTimersByTime(310);
      expect(detector.getIsHolding()).toBe(true);

      // Subsequent keydown returns null because hold-start was already triggered by timer
      const e2 = detector.feedKeyDown(fnKey, 'Fn');
      expect(e2).toBeNull();

      // Release emits hold-release
      const e3 = detector.feedKeyUp(fnKey, 'Fn');
      expect(e3).toEqual<DictationTriggerEvent>({ type: 'hold-release' });
    });

    it('fires onTrigger callback on hold-start and hold-release', () => {
      const triggers: DictationTriggerEvent[] = [];
      const cbDetector = new ShortcutDetector({
        holdThresholdMs: 300,
        onTrigger: (e) => triggers.push(e)
      });
      const fnKey = createKeyEvent({ key: 'Fn' });

      cbDetector.feedKeyDown(fnKey, 'Fn');
      vi.advanceTimersByTime(310);

      expect(triggers).toEqual([{ type: 'hold-start' }]);

      cbDetector.feedKeyUp(fnKey, 'Fn');
      expect(triggers).toEqual([
        { type: 'hold-start' },
        { type: 'hold-release' }
      ]);
    });

    it('emits hold-release on keyup even if keydown repeat did not fire if held > 300ms', () => {
      const fnKey = createKeyEvent({ key: 'Fn' });

      detector.feedKeyDown(fnKey, 'Fn');
      vi.advanceTimersByTime(350);

      const eUp = detector.feedKeyUp(fnKey, 'Fn');
      expect(eUp).toEqual<DictationTriggerEvent>({ type: 'hold-release' });
    });
  });

  describe('Double-Tap Hands-Free Lock (< 350ms)', () => {
    it('detects double-tap lock within 350ms interval', () => {
      const fnKey = createKeyEvent({ key: 'Fn' });

      // Tap 1: press and release fast (50ms)
      detector.feedKeyDown(fnKey, 'Fn');
      vi.advanceTimersByTime(50);
      const e1 = detector.feedKeyUp(fnKey, 'Fn');
      expect(e1).toBeNull();

      // Wait 100ms (< 350ms total between releases)
      vi.advanceTimersByTime(100);

      // Tap 2: press and release fast
      detector.feedKeyDown(fnKey, 'Fn');
      vi.advanceTimersByTime(50);
      const e2 = detector.feedKeyUp(fnKey, 'Fn');
      expect(e2).toEqual<DictationTriggerEvent>({ type: 'double-tap-lock' });
      expect(detector.getIsLocked()).toBe(true);
    });

    it('does not trigger double-tap if interval exceeds 350ms', () => {
      const fnKey = createKeyEvent({ key: 'Fn' });

      // Tap 1
      detector.feedKeyDown(fnKey, 'Fn');
      vi.advanceTimersByTime(50);
      detector.feedKeyUp(fnKey, 'Fn');

      // Wait 400ms (> 350ms)
      vi.advanceTimersByTime(400);

      // Tap 2
      detector.feedKeyDown(fnKey, 'Fn');
      vi.advanceTimersByTime(50);
      const e2 = detector.feedKeyUp(fnKey, 'Fn');
      expect(e2).toBeNull();
      expect(detector.getIsLocked()).toBe(false);
    });
  });

  describe('Single-Tap Unlock while locked', () => {
    it('emits single-tap-unlock when shortcut is pressed while in locked state', () => {
      const fnKey = createKeyEvent({ key: 'Fn' });

      // Enter locked state via double tap
      detector.feedKeyDown(fnKey, 'Fn');
      vi.advanceTimersByTime(40);
      detector.feedKeyUp(fnKey, 'Fn');
      vi.advanceTimersByTime(60);
      detector.feedKeyDown(fnKey, 'Fn');
      vi.advanceTimersByTime(40);
      detector.feedKeyUp(fnKey, 'Fn');

      expect(detector.getIsLocked()).toBe(true);

      // User taps key again to unlock / stop recording
      vi.advanceTimersByTime(1000);
      const unlockEvent = detector.feedKeyDown(fnKey, 'Fn');
      expect(unlockEvent).toEqual<DictationTriggerEvent>({ type: 'single-tap-unlock' });
      expect(detector.getIsLocked()).toBe(false);

      // Keyup after unlock is consumed cleanly
      const upEvent = detector.feedKeyUp(fnKey, 'Fn');
      expect(upEvent).toBeNull();
    });
  });

  describe('Reset', () => {
    it('clears state on reset()', () => {
      const fnKey = createKeyEvent({ key: 'Fn' });
      detector.feedKeyDown(fnKey, 'Fn');
      detector.reset();
      expect(detector.getIsHolding()).toBe(false);
      expect(detector.getIsLocked()).toBe(false);
    });
  });
});
