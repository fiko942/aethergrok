import { describe, it, expect, vi } from 'vitest';

describe('Cmd+W Keyboard Interception', () => {
  it('identifies Cmd+W key combinations across platforms', () => {
    const isCmdW = (e: { metaKey: boolean; ctrlKey: boolean; altKey: boolean; shiftKey: boolean; key: string }) => {
      const isMetaOrCtrl = e.metaKey || e.ctrlKey;
      return isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'w';
    };

    expect(isCmdW({ metaKey: true, ctrlKey: false, altKey: false, shiftKey: false, key: 'w' })).toBe(true);
    expect(isCmdW({ metaKey: false, ctrlKey: true, altKey: false, shiftKey: false, key: 'W' })).toBe(true);
    expect(isCmdW({ metaKey: true, ctrlKey: false, altKey: true, shiftKey: false, key: 'w' })).toBe(false);
    expect(isCmdW({ metaKey: false, ctrlKey: false, altKey: false, shiftKey: false, key: 'w' })).toBe(false);
  });

  it('excludes "w" from terminal pass-through keys so Cmd+W never leaks to OS performClose', () => {
    const terminalPassthroughKeys = ['k', 't', 'b', 'c', 'v', 'l', 'u', 'r', 'a', 'e', 'd', 'z', 'p', 'n', 'f'];
    expect(terminalPassthroughKeys.includes('w')).toBe(false);
  });

  it('intercepts Cmd+W in terminal to close active terminal tab without leaking to AppKit', () => {
    const preventDefault = vi.fn();
    const stopPropagation = vi.fn();
    const closeTerminalTab = vi.fn();
    const closeSessionTab = vi.fn();

    const e = {
      metaKey: true,
      ctrlKey: false,
      altKey: false,
      shiftKey: false,
      key: 'w',
      preventDefault,
      stopPropagation
    };

    const inTerminal = true;
    const sessionStore = { activeSessionId: 'sess-1', closeSessionTab };
    const terminalStore = {
      activeTerminalIdPerSession: { 'sess-1': 'term-42' },
      closeTerminalTab
    };

    const isMetaOrCtrl = e.metaKey || e.ctrlKey;
    if (isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'w') {
      e.preventDefault();
      e.stopPropagation();

      if (inTerminal && sessionStore.activeSessionId) {
        const activeTermId = terminalStore.activeTerminalIdPerSession[sessionStore.activeSessionId];
        if (activeTermId) {
          terminalStore.closeTerminalTab(sessionStore.activeSessionId, activeTermId);
        }
      } else if (sessionStore.activeSessionId) {
        sessionStore.closeSessionTab(sessionStore.activeSessionId);
      }
    }

    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(stopPropagation).toHaveBeenCalledTimes(1);
    expect(closeTerminalTab).toHaveBeenCalledWith('sess-1', 'term-42');
    expect(closeSessionTab).not.toHaveBeenCalled();
  });

  it('intercepts Cmd+W outside terminal to close active session tab', () => {
    const preventDefault = vi.fn();
    const stopPropagation = vi.fn();
    const closeTerminalTab = vi.fn();
    const closeSessionTab = vi.fn();

    const e = {
      metaKey: true,
      ctrlKey: false,
      altKey: false,
      shiftKey: false,
      key: 'w',
      preventDefault,
      stopPropagation
    };

    const inTerminal = false;
    const sessionStore = { activeSessionId: 'sess-2', closeSessionTab };
    const terminalStore = {
      activeTerminalIdPerSession: { 'sess-2': 'term-1' },
      closeTerminalTab
    };

    const isMetaOrCtrl = e.metaKey || e.ctrlKey;
    if (isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'w') {
      e.preventDefault();
      e.stopPropagation();

      if (inTerminal && sessionStore.activeSessionId) {
        const activeTermId = terminalStore.activeTerminalIdPerSession[sessionStore.activeSessionId];
        if (activeTermId) {
          terminalStore.closeTerminalTab(sessionStore.activeSessionId, activeTermId);
        }
      } else if (sessionStore.activeSessionId) {
        sessionStore.closeSessionTab(sessionStore.activeSessionId);
      }
    }

    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(stopPropagation).toHaveBeenCalledTimes(1);
    expect(closeTerminalTab).not.toHaveBeenCalled();
    expect(closeSessionTab).toHaveBeenCalledWith('sess-2');
  });
});
