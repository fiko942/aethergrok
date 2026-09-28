/**
 * shortcutDetector.ts
 *
 * Dual-Mode Shortcut Engine supporting Push-to-Talk (press & hold > 300ms)
 * and Double-Tap hands-free lock (< 350ms between two taps), as well as
 * single-tap unlock to stop a locked recording session.
 */

export type DictationTriggerEventType =
  | 'hold-start'
  | 'hold-release'
  | 'double-tap-lock'
  | 'single-tap-unlock';

export interface DictationTriggerEvent {
  type: DictationTriggerEventType;
}

export interface ShortcutDetectorOptions {
  holdThresholdMs?: number;
  doubleTapThresholdMs?: number;
  onTrigger?: (event: DictationTriggerEvent) => void;
}

export class ShortcutDetector {
  private holdThresholdMs: number;
  private doubleTapThresholdMs: number;
  private onTrigger?: (event: DictationTriggerEvent) => void;

  private isDown: boolean = false;
  private isHolding: boolean = false;
  private isLocked: boolean = false;
  private holdStartEmitted: boolean = false;

  private pressStartTime: number = 0;
  private lastTapReleaseTime: number = 0;
  private holdTimer: any = null;

  constructor(options?: ShortcutDetectorOptions) {
    this.holdThresholdMs = options?.holdThresholdMs ?? 300;
    this.doubleTapThresholdMs = options?.doubleTapThresholdMs ?? 350;
    this.onTrigger = options?.onTrigger;
  }

  /**
   * Evaluates if a given KeyboardEvent matches the configured shortcut string.
   */
  public matchesShortcut(e: KeyboardEvent, shortcutStr: string): boolean {
    if (!shortcutStr) return false;

    const normCode = (e.code || '').toLowerCase().replace(/[\s_-]/g, '');
    const normKey = (e.key || '').toLowerCase().replace(/[\s_-]/g, '');
    const normShortcut = shortcutStr.toLowerCase().replace(/[\s_-]/g, '');

    // 1. Single character / Backslash handling
    if (normShortcut === '\\' || normShortcut === 'backslash') {
      if (e.key === '\\' || e.code === 'Backslash') {
        return true;
      }
      return false;
    }

    // 2. Direct single key / single modifier code check
    const isDirectMatch = normCode && (
      normCode === normShortcut ||
      (normCode === 'shiftright' && (normShortcut === 'rightshift' || normShortcut === 'shiftright')) ||
      (normCode === 'shiftleft' && (normShortcut === 'leftshift' || normShortcut === 'shiftleft')) ||
      (normCode === 'controlright' && (normShortcut === 'rightctrl' || normShortcut === 'controlright' || normShortcut === 'rightcontrol')) ||
      (normCode === 'controlleft' && (normShortcut === 'leftctrl' || normShortcut === 'controlleft' || normShortcut === 'leftcontrol')) ||
      (normCode === 'altright' && (normShortcut === 'rightalt' || normShortcut === 'altright' || normShortcut === 'rightoption' || normShortcut === 'rightopt')) ||
      (normCode === 'altleft' && (normShortcut === 'leftalt' || normShortcut === 'altleft' || normShortcut === 'leftoption' || normShortcut === 'leftopt')) ||
      (normCode === 'metaright' && (normShortcut === 'rightcmd' || normShortcut === 'metaright' || normShortcut === 'rightwin' || normShortcut === 'rightmeta')) ||
      (normCode === 'metaleft' && (normShortcut === 'leftcmd' || normShortcut === 'metaleft' || normShortcut === 'leftwin' || normShortcut === 'leftmeta'))
    );

    if (isDirectMatch) {
      return true;
    }

    // 3. Modifier combination check (e.g. "Cmd+D", "Ctrl+Shift+D", "Alt+Space")
    const parts = shortcutStr.toLowerCase().split('+').map((s) => s.trim());
    const hasCmd = parts.includes('cmd') || parts.includes('meta') || parts.includes('cmdorctrl');
    const hasCtrl = parts.includes('ctrl') || parts.includes('control') || parts.includes('cmdorctrl');
    const hasShift = parts.includes('shift');
    const hasAlt = parts.includes('alt') || parts.includes('opt') || parts.includes('option');

    const keyPart = parts.find((p) => !['cmdorctrl', 'cmd', 'ctrl', 'control', 'meta', 'shift', 'alt', 'opt', 'option'].includes(p));

    // Modifiers matching
    const isMetaOrCtrl = e.metaKey || e.ctrlKey;
    if (parts.includes('cmdorctrl')) {
      if (!isMetaOrCtrl) return false;
    } else {
      if (hasCmd && !e.metaKey) return false;
      if (!hasCmd && !hasCtrl && isMetaOrCtrl) return false;
      if (hasCtrl && !e.ctrlKey) return false;
    }

    if (hasShift && !e.shiftKey) return false;
    if (!hasShift && e.shiftKey) return false;
    if (hasAlt && !e.altKey) return false;
    if (!hasAlt && e.altKey) return false;

    if (keyPart) {
      const eKey = (e.key || '').toLowerCase();
      const eCode = (e.code || '').toLowerCase();
      if (eKey === keyPart) return true;
      if (eCode === keyPart || eCode === `key${keyPart}` || eCode === `digit${keyPart}`) return true;
      if (keyPart === '/' && (eKey === '/' || eCode === 'slash')) return true;
      if (keyPart === 'delete' && (eKey === 'delete' || eCode === 'delete')) return true;
      if (keyPart === 'backspace' && (eKey === 'backspace' || eCode === 'backspace')) return true;
      if (keyPart === 'space' && (eKey === ' ' || eCode === 'space')) return true;
      return false;
    }

    return true;
  }

  /**
   * Feed a keydown event.
   * If auto-repeat fires, it's ignored or used to trigger hold-start after holdThresholdMs.
   */
  public feedKeyDown(e: KeyboardEvent, configuredShortcut: string): DictationTriggerEvent | null {
    if (!this.matchesShortcut(e, configuredShortcut)) {
      return null;
    }

    const now = Date.now();

    // If key is already down (e.g. browser key repeat)
    if (this.isDown) {
      if (!this.holdStartEmitted && !this.isLocked && now - this.pressStartTime >= this.holdThresholdMs) {
        this.isHolding = true;
        this.holdStartEmitted = true;
        if (this.holdTimer) {
          clearTimeout(this.holdTimer);
          this.holdTimer = null;
        }
        const event: DictationTriggerEvent = { type: 'hold-start' };
        this.onTrigger?.(event);
        return event;
      }
      return null;
    }

    this.isDown = true;
    this.pressStartTime = now;
    this.holdStartEmitted = false;

    // If we are currently locked, pressing the shortcut again acts as an unlock trigger
    if (this.isLocked) {
      this.isLocked = false;
      this.isHolding = false;
      this.holdStartEmitted = false;
      const event: DictationTriggerEvent = { type: 'single-tap-unlock' };
      this.onTrigger?.(event);
      return event;
    }

    // Set timer for hold detection when key repeat is delayed or disabled
    if (this.holdTimer) {
      clearTimeout(this.holdTimer);
    }
    this.holdTimer = setTimeout(() => {
      if (this.isDown && !this.holdStartEmitted && !this.isLocked) {
        this.isHolding = true;
        this.holdStartEmitted = true;
        this.onTrigger?.({ type: 'hold-start' });
      }
    }, this.holdThresholdMs);

    return null;
  }

  /**
   * Feed a keyup event.
   */
  public feedKeyUp(e: KeyboardEvent, configuredShortcut: string): DictationTriggerEvent | null {
    if (!this.matchesShortcut(e, configuredShortcut)) {
      return null;
    }

    const now = Date.now();
    const pressDuration = now - this.pressStartTime;

    if (this.holdTimer) {
      clearTimeout(this.holdTimer);
      this.holdTimer = null;
    }

    const wasHolding = this.holdStartEmitted || this.isHolding || pressDuration >= this.holdThresholdMs;
    this.isDown = false;
    this.isHolding = false;
    this.holdStartEmitted = false;

    // 1. If it was a hold (> holdThresholdMs), emit hold-release
    if (wasHolding) {
      this.lastTapReleaseTime = 0;
      const event: DictationTriggerEvent = { type: 'hold-release' };
      this.onTrigger?.(event);
      return event;
    }

    // 2. Otherwise, it is a quick tap (< holdThresholdMs)
    const timeSinceLastTap = now - this.lastTapReleaseTime;
    if (this.lastTapReleaseTime > 0 && timeSinceLastTap <= this.doubleTapThresholdMs) {
      // Double tap detected!
      this.isLocked = true;
      this.lastTapReleaseTime = 0;
      const event: DictationTriggerEvent = { type: 'double-tap-lock' };
      this.onTrigger?.(event);
      return event;
    } else {
      // First tap of a potential double-tap
      this.lastTapReleaseTime = now;
      return null;
    }
  }

  /**
   * Resets internal state (e.g. when recording is cancelled or completed).
   */
  public reset(): void {
    if (this.holdTimer) {
      clearTimeout(this.holdTimer);
      this.holdTimer = null;
    }
    this.isDown = false;
    this.isHolding = false;
    this.isLocked = false;
    this.holdStartEmitted = false;
    this.pressStartTime = 0;
    this.lastTapReleaseTime = 0;
  }

  public getIsLocked(): boolean {
    return this.isLocked;
  }

  public getIsHolding(): boolean {
    return this.isHolding;
  }
}
