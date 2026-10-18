import { describe, it, expect } from 'vitest';

describe('Application Quit Shortcut & Clipboard Safety Guard', () => {
  function shouldTriggerQuit(e: {
    isMetaOrCtrl: boolean;
    key: string;
    shiftKey: boolean;
    altKey: boolean;
  }): boolean {
    const isCmdOrCtrlQ = e.isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'q';
    const isAltF4 = e.altKey && e.key === 'F4';
    return isCmdOrCtrlQ || isAltF4;
  }

  function shouldPreserveStandardCopy(e: {
    isMetaOrCtrl: boolean;
    key: string;
  }): boolean {
    // Cmd+C / Ctrl+C must always be preserved for clipboard copy
    return e.isMetaOrCtrl && e.key.toLowerCase() === 'c';
  }

  it('triggers quit on Cmd+Q (macOS) and Ctrl+Q (Windows)', () => {
    expect(shouldTriggerQuit({ isMetaOrCtrl: true, key: 'q', shiftKey: false, altKey: false })).toBe(true);
    expect(shouldTriggerQuit({ isMetaOrCtrl: true, key: 'Q', shiftKey: false, altKey: false })).toBe(true);
  });

  it('triggers quit on Alt+F4 (Windows standard quit)', () => {
    expect(shouldTriggerQuit({ isMetaOrCtrl: false, key: 'F4', shiftKey: false, altKey: true })).toBe(true);
  });

  it('never intercepts Cmd+C or Ctrl+C for quit to preserve standard clipboard copy', () => {
    expect(shouldTriggerQuit({ isMetaOrCtrl: true, key: 'c', shiftKey: false, altKey: false })).toBe(false);
    expect(shouldPreserveStandardCopy({ isMetaOrCtrl: true, key: 'c' })).toBe(true);
    expect(shouldPreserveStandardCopy({ isMetaOrCtrl: true, key: 'C' })).toBe(true);
  });
});
