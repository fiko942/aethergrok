import { describe, it, expect, vi } from 'vitest';

describe('Cmd+W / Ctrl+W Keyboard Interception & Terminal Passthrough', () => {
  function dispatchAppKeydown(
    e: {
      metaKey: boolean;
      ctrlKey: boolean;
      altKey: boolean;
      shiftKey: boolean;
      key: string;
      preventDefault: () => void;
      stopPropagation: () => void;
    },
    inTerminal: boolean,
    sessionStore: { activeSessionId: string | null; closeSessionTab: (id: string) => void },
    terminalStore: {
      activeTerminalIdPerSession: Record<string, string>;
      closeTerminalTab: (sessId: string, termId: string) => void;
    },
    isEditing = false
  ) {
    const isMetaOrCtrl = e.metaKey || e.ctrlKey;

    // When focused in terminal, let standard terminal control sequences pass through cleanly
    if (inTerminal && !e.altKey && (e.ctrlKey || (!isEditing && e.metaKey))) {
      const k = e.key.toLowerCase();
      // Ctrl+W is readline unix-word-rubout in terminal - pass through when ctrlKey and !metaKey
      if (k === 'w' && e.ctrlKey && !e.metaKey) {
        return;
      }
      if (['k', 't', 'b', 'c', 'v', 'l', 'u', 'r', 'a', 'e', 'd', 'z', 'p', 'n', 'f'].includes(k)) {
        return;
      }
    }

    // Intercept Cmd/Ctrl + W to close tabs and prevent macOS [NSWindow performClose:] from killing the app
    // Preserve Ctrl+W inside terminal for readline unix-word-rubout
    if (isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'w') {
      if (inTerminal && e.ctrlKey && !e.metaKey) {
        return;
      }
      e.preventDefault();
      e.stopPropagation();

      // If inside terminal, close active terminal pane/tab first
      if (inTerminal && sessionStore.activeSessionId) {
        const activeTermId = terminalStore.activeTerminalIdPerSession[sessionStore.activeSessionId];
        if (activeTermId) {
          terminalStore.closeTerminalTab(sessionStore.activeSessionId, activeTermId);
          return;
        }
      }

      // Otherwise close active session tab
      if (sessionStore.activeSessionId) {
        sessionStore.closeSessionTab(sessionStore.activeSessionId);
      }
      return;
    }
  }

  it('identifies Cmd+W and Ctrl+W key combinations across platforms', () => {
    const isCmdW = (e: { metaKey: boolean; ctrlKey: boolean; altKey: boolean; shiftKey: boolean; key: string }) => {
      const isMetaOrCtrl = e.metaKey || e.ctrlKey;
      return isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'w';
    };

    expect(isCmdW({ metaKey: true, ctrlKey: false, altKey: false, shiftKey: false, key: 'w' })).toBe(true);
    expect(isCmdW({ metaKey: false, ctrlKey: true, altKey: false, shiftKey: false, key: 'W' })).toBe(true);
    expect(isCmdW({ metaKey: true, ctrlKey: false, altKey: true, shiftKey: false, key: 'w' })).toBe(false);
    expect(isCmdW({ metaKey: false, ctrlKey: false, altKey: false, shiftKey: false, key: 'w' })).toBe(false);
  });

  it('passes Ctrl+W through in terminal without preventing default or closing any tab (readline unix-word-rubout)', () => {
    const preventDefault = vi.fn();
    const stopPropagation = vi.fn();
    const closeTerminalTab = vi.fn();
    const closeSessionTab = vi.fn();

    const e = {
      metaKey: false,
      ctrlKey: true,
      altKey: false,
      shiftKey: false,
      key: 'w',
      preventDefault,
      stopPropagation
    };

    const sessionStore = { activeSessionId: 'sess-1', closeSessionTab };
    const terminalStore = {
      activeTerminalIdPerSession: { 'sess-1': 'term-1' },
      closeTerminalTab
    };

    dispatchAppKeydown(e, true, sessionStore, terminalStore);

    expect(preventDefault).not.toHaveBeenCalled();
    expect(stopPropagation).not.toHaveBeenCalled();
    expect(closeTerminalTab).not.toHaveBeenCalled();
    expect(closeSessionTab).not.toHaveBeenCalled();
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

    const sessionStore = { activeSessionId: 'sess-1', closeSessionTab };
    const terminalStore = {
      activeTerminalIdPerSession: { 'sess-1': 'term-42' },
      closeTerminalTab
    };

    dispatchAppKeydown(e, true, sessionStore, terminalStore);

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

    const sessionStore = { activeSessionId: 'sess-2', closeSessionTab };
    const terminalStore = {
      activeTerminalIdPerSession: { 'sess-2': 'term-1' },
      closeTerminalTab
    };

    dispatchAppKeydown(e, false, sessionStore, terminalStore);

    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(stopPropagation).toHaveBeenCalledTimes(1);
    expect(closeTerminalTab).not.toHaveBeenCalled();
    expect(closeSessionTab).toHaveBeenCalledWith('sess-2');
  });

  it('intercepts Ctrl+W outside terminal to close active session tab', () => {
    const preventDefault = vi.fn();
    const stopPropagation = vi.fn();
    const closeTerminalTab = vi.fn();
    const closeSessionTab = vi.fn();

    const e = {
      metaKey: false,
      ctrlKey: true,
      altKey: false,
      shiftKey: false,
      key: 'w',
      preventDefault,
      stopPropagation
    };

    const sessionStore = { activeSessionId: 'sess-3', closeSessionTab };
    const terminalStore = {
      activeTerminalIdPerSession: { 'sess-3': 'term-2' },
      closeTerminalTab
    };

    dispatchAppKeydown(e, false, sessionStore, terminalStore);

    expect(preventDefault).toHaveBeenCalledTimes(1);
    expect(stopPropagation).toHaveBeenCalledTimes(1);
    expect(closeTerminalTab).not.toHaveBeenCalled();
    expect(closeSessionTab).toHaveBeenCalledWith('sess-3');
  });
});
