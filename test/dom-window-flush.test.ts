import { describe, it, expect } from 'vitest';

describe('DOM Active Turn Windowing & Flush Logic', () => {
  function getActiveWindowTurns(settings: { activeWindowTurnCount?: number }): number {
    if (typeof settings.activeWindowTurnCount === 'number') {
      return Math.max(2, Math.min(50, settings.activeWindowTurnCount));
    }
    return 4;
  }

  function computeVisibleMessages(messages: Array<{ id: string; role: string; content: string }>, windowTurns: number) {
    const userTurnIndices: number[] = [];
    for (let i = 0; i < messages.length; i++) {
      if (messages[i].role === 'user') {
        userTurnIndices.push(i);
      }
    }

    const totalTurns = userTurnIndices.length;
    if (totalTurns <= windowTurns) {
      return messages;
    }

    const cutoffTurnIdx = totalTurns - windowTurns;
    const startIndex = userTurnIndices[cutoffTurnIdx];
    return messages.slice(startIndex);
  }

  it('getActiveWindowTurns reflects activeWindowTurnCount and clamps boundaries', () => {
    expect(getActiveWindowTurns({ activeWindowTurnCount: 10 })).toBe(10);
    expect(getActiveWindowTurns({ activeWindowTurnCount: 4 })).toBe(4);
    expect(getActiveWindowTurns({ activeWindowTurnCount: 30 })).toBe(30);

    // Clamps minimum to 2
    expect(getActiveWindowTurns({ activeWindowTurnCount: 1 })).toBe(2);

    // Clamps maximum to 50
    expect(getActiveWindowTurns({ activeWindowTurnCount: 100 })).toBe(50);

    // Fallback when undefined
    expect(getActiveWindowTurns({})).toBe(4);
  });

  it('correctly calculates visible messages sliding window and flushes older turns', () => {
    const userTurnCount = 15;
    const messages = [];

    for (let i = 0; i < userTurnCount; i++) {
      messages.push({ id: `user_${i}`, role: 'user', content: `Prompt ${i}` });
      messages.push({ id: `asst_${i}`, role: 'assistant', content: `Response ${i}` });
    }

    // Window set to 4 turns
    const visible4 = computeVisibleMessages(messages, 4);
    // 4 turns = 4 user + 4 assistant = 8 messages
    expect(visible4.length).toBe(8);
    expect(visible4[0].id).toBe('user_11');
    expect(visible4[visible4.length - 1].id).toBe('asst_14');

    // Window set to 10 turns
    const visible10 = computeVisibleMessages(messages, 10);
    expect(visible10.length).toBe(20);
    expect(visible10[0].id).toBe('user_5');
    expect(visible10[visible10.length - 1].id).toBe('asst_14');
  });

  it('auto-flushes expanded turns back to configured window limit when returning to bottom', () => {
    let sessionVisibleTurnCount = 12; // Expanded by user clicking Show Earlier
    const configuredLimit = 4;

    function flushEarlierTurnsToWindow(): boolean {
      if (sessionVisibleTurnCount > configuredLimit) {
        sessionVisibleTurnCount = configuredLimit;
        return true;
      }
      return false;
    }

    expect(flushEarlierTurnsToWindow()).toBe(true);
    expect(sessionVisibleTurnCount).toBe(4);
    expect(flushEarlierTurnsToWindow()).toBe(false);
  });

  it('auto-flushes older turns during live activity when pinned to bottom but preserves them when scrolled up', () => {
    let sessionVisibleTurnCount = 15;
    const configuredLimit = 4;
    let autoScrollToBottom = true;

    function handleLiveActivity() {
      if (autoScrollToBottom && sessionVisibleTurnCount > configuredLimit) {
        sessionVisibleTurnCount = configuredLimit;
      }
    }

    // Live activity arrives while pinned to bottom
    handleLiveActivity();
    expect(sessionVisibleTurnCount).toBe(4);

    // User scrolled up and expanded turns
    autoScrollToBottom = false;
    sessionVisibleTurnCount = 12;

    // Live activity arrives while user is reading history (unpinned)
    handleLiveActivity();
    expect(sessionVisibleTurnCount).toBe(12); // Must NOT collapse under user eyes

    // User returns to bottom
    autoScrollToBottom = true;
    handleLiveActivity();
    expect(sessionVisibleTurnCount).toBe(4); // Flushes to configured setting
  });

  it('correctly handles mouse wheel direction: wheel up marks manual scroll, wheel down unmarks', () => {
    let isUserManualScrolling = false;

    function handleWheel(deltaY: number) {
      if (deltaY < 0) {
        // Scrolling up towards history
        isUserManualScrolling = true;
      } else if (deltaY > 0) {
        // Scrolling down towards bottom / live activity
        isUserManualScrolling = false;
      }
    }

    handleWheel(-120); // Mouse wheel notch up
    expect(isUserManualScrolling).toBe(true);

    handleWheel(120); // Mouse wheel notch down
    expect(isUserManualScrolling).toBe(false);
  });
});
