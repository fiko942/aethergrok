/**
 * Terminal state management per session with multi-tab terminal support
 */

export type TerminalDockPosition = 'bottom' | 'right';

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
  // Map of sessionId -> collapsed state (true = collapsed to slim bar)
  sessionCollapsed = $state<Record<string, boolean>>({});
  // Map of sessionId -> dock position ('bottom' | 'right')
  sessionDockPosition = $state<Record<string, TerminalDockPosition>>({});
  // Fallback global open state (for backwards compatibility if needed)
  isOpen = $state<boolean>(true);
  // Panel height in pixels (when docked at bottom)
  panelHeight = $state<number>(240);
  // Panel width in pixels (when docked at right)
  panelWidth = $state<number>(440);

  getTerminalTabs(sessionId: string): TerminalTab[] {
    return this.sessionTerminals[sessionId] || [];
  }

  getActiveTerminalId(sessionId: string): string | null {
    return this.activeTerminalIdPerSession[sessionId] || null;
  }

  isSessionCollapsed(sessionId: string): boolean {
    if (!sessionId) return false;
    return this.sessionCollapsed[sessionId] ?? false;
  }

  toggleSessionCollapse(sessionId: string, collapsed?: boolean): void {
    if (!sessionId) return;
    const current = this.isSessionCollapsed(sessionId);
    this.sessionCollapsed[sessionId] = collapsed !== undefined ? collapsed : !current;
  }

  getDockPosition(sessionId: string): TerminalDockPosition {
    if (!sessionId) return 'bottom';
    return this.sessionDockPosition[sessionId] || 'bottom';
  }

  setDockPosition(sessionId: string, pos: TerminalDockPosition): void {
    if (!sessionId) return;
    this.sessionDockPosition[sessionId] = pos;
  }

  toggleDockPosition(sessionId: string): void {
    if (!sessionId) return;
    const current = this.getDockPosition(sessionId);
    this.sessionDockPosition[sessionId] = current === 'bottom' ? 'right' : 'bottom';
  }

  toggleOpen(open?: boolean): void {
    this.isOpen = open !== undefined ? open : !this.isOpen;
  }

  setPanelHeight(height: number): void {
    this.panelHeight = Math.max(140, Math.min(height, window.innerHeight - 150));
  }

  setPanelWidth(width: number): void {
    this.panelWidth = Math.max(260, Math.min(width, window.innerWidth - 350));
  }

  renameTerminal(sessionId: string, termId: string, newTitle: string): void {
    if (!sessionId || !termId) return;
    const list = this.sessionTerminals[sessionId];
    if (!list) return;
    const tab = list.find((t) => t.id === termId);
    if (tab && newTitle.trim()) {
      tab.title = newTitle.trim();
    }
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
    delete this.sessionCollapsed[sessionId];
    delete this.sessionDockPosition[sessionId];
  }
}

export const terminalStore = new TerminalStore();
