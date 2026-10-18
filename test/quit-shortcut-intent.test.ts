import { describe, it, expect } from 'vitest';

describe('Quit Shortcut Intent & Text Selection Guard', () => {
  function shouldTriggerQuitFromCmdC(input: {
    isMetaOrCtrl: boolean;
    key: string;
    shiftKey: boolean;
    altKey: boolean;
    isEditing: boolean;
    hasSelection: boolean;
  }): boolean {
    if (!input.isMetaOrCtrl || input.shiftKey || input.altKey || input.key.toLowerCase() !== 'c') {
      return false;
    }
    // If editing or text is selected, do NOT trigger quit (must copy)
    if (input.isEditing || input.hasSelection) {
      return false;
    }
    return true;
  }

  it('preserves normal copy when text is selected', () => {
    const res = shouldTriggerQuitFromCmdC({
      isMetaOrCtrl: true,
      key: 'c',
      shiftKey: false,
      altKey: false,
      isEditing: false,
      hasSelection: true
    });
    expect(res).toBe(false);
  });

  it('preserves normal copy/kill inside terminal or text input', () => {
    const res = shouldTriggerQuitFromCmdC({
      isMetaOrCtrl: true,
      key: 'c',
      shiftKey: false,
      altKey: false,
      isEditing: true,
      hasSelection: false
    });
    expect(res).toBe(false);
  });

  it('triggers quit intent when Cmd+C is pressed outside of text inputs without selection', () => {
    const res = shouldTriggerQuitFromCmdC({
      isMetaOrCtrl: true,
      key: 'c',
      shiftKey: false,
      altKey: false,
      isEditing: false,
      hasSelection: false
    });
    expect(res).toBe(true);
  });
});
