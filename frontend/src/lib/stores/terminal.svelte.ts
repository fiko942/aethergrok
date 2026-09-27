/**
 * Terminal state management per session with multi-tab terminal support
 */

export interface TerminalTab {
  id: string;
  sessionId: string;
  title: string;
  cwd: string;
  createdAt: number;
}

export class TerminalStore {
  // Map of sessionId -> TerminalTab[]
  sessionTerminals = $state<Record<string, TerminalTab[]>>({});
  // Map of sessionId -> active terminalId
  activeTerminalIdPerSession = $state<Record<string, string>>({});
  // Is bottom terminal panel visible
  isOpen = $state<boolean>(false);
  // Panel height in pixels
  panelHeight = $state<number>(240);

  getTerminalTabs(sessionId: string): TerminalTab[] {
    return this.sessionTerminals[sessionId] || [];
  }

  getActiveTerminalId(sessionId: string): string | null {
    return this.activeTerminalIdPerSession[sessionId] || null;
  }

  toggleOpen(open?: boolean): void {
    this.isOpen = open !== undefined ? open : !this.isOpen;
  }

  setPanelHeight(height: number): void {
    this.panelHeight = Math.max(140, Math.min(height, window.innerHeight - 150));
  }

  createTerminal(sessionId: string, cwd: string, title?: string): TerminalTab {
    const list = this.sessionTerminals[sessionId] ? [...this.sessionTerminals[sessionId]] : [];
    const count = list.length + 1;
    const termId = `term_${sessionId}_${Date.now()}_${Math.random().toString(36).substring(2, 6)}`;
    const tab: TerminalTab = {
      id: termId,
      sessionId,
      title: title || `Terminal ${count}`,
      cwd,
      createdAt: Date.now()
    };

    list.push(tab);
    this.sessionTerminals[sessionId] = list;
    this.activeTerminalIdPerSession[sessionId] = termId;

    // Call Go backend to instantiate PTY with isolated Process Group
    if (window.go?.main?.App?.CreateTerminal) {
      window.go.main.App.CreateTerminal(sessionId, termId, cwd, '')
        .catch((err) => console.error('Failed to create backend terminal:', err));
    }

    return tab;
  }

  switchTerminal(sessionId: string, termId: string): void {
    this.activeTerminalIdPerSession[sessionId] = termId;
  }

  closeTerminal(sessionId: string, termId: string): void {
    const list = this.sessionTerminals[sessionId] || [];
    const idx = list.findIndex((t) => t.id === termId);
    if (idx === -1) return;

    // Call Go backend to SIGKILL entire process group
    if (window.go?.main?.App?.CloseTerminal) {
      window.go.main.App.CloseTerminal(termId)
        .catch((err) => console.error('Failed to close backend terminal:', err));
    }

    const updated = list.filter((t) => t.id !== termId);
    this.sessionTerminals[sessionId] = updated;

    if (this.activeTerminalIdPerSession[sessionId] === termId) {
      if (updated.length > 0) {
        const nextIdx = Math.max(0, idx - 1);
        this.activeTerminalIdPerSession[sessionId] = updated[nextIdx].id;
      } else {
        delete this.activeTerminalIdPerSession[sessionId];
        this.isOpen = false;
      }
    }
  }

  closeAllForSession(sessionId: string): void {
    if (window.go?.main?.App?.CloseSessionTerminals) {
      window.go.main.App.CloseSessionTerminals(sessionId)
        .catch((err) => console.error('Failed to close session terminals:', err));
    }
    delete this.sessionTerminals[sessionId];
    delete this.activeTerminalIdPerSession[sessionId];
  }
}

export const terminalStore = new TerminalStore();
