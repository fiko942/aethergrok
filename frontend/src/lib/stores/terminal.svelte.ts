/**
 * Terminal state management per session with multi-tab terminal & split pane support
 * Persists layout, tabs, split groups, sizes, dock positions, and collapse state per session
 */

import { logger } from './logger.svelte';

export type TerminalDockPosition = 'bottom' | 'right';

export interface TerminalTab {
  id: string;
  sessionId: string;
  title: string;
  cwd: string;
  createdAt: number;
}

export interface TerminalSplitGroup {
  id: string;
  paneTermIds: string[]; // List of termIds side-by-side or stacked in this group
  splitDirection: 'horizontal' | 'vertical'; // default 'horizontal' (columns)
  paneSizes?: number[]; // percentage weights per pane, summing to 100
}

interface PersistedTerminalState {
  version: number;
  sessionTerminals: Record<string, TerminalTab[]>;
  activeTerminalIdPerSession: Record<string, string>;
  sessionSplitGroups: Record<string, TerminalSplitGroup[]>;
  focusedPaneTermId: Record<string, string>;
  sessionCollapsed: Record<string, boolean>;
  sessionDockPosition: Record<string, TerminalDockPosition>;
  panelHeight: number;
  panelWidth: number;
}

const STORAGE_KEY = 'aethergrok_terminal_state_v1';

export class TerminalStore {
  // Map of sessionId -> TerminalTab[]
  sessionTerminals = $state<Record<string, TerminalTab[]>>({});
  // Map of sessionId -> active top-level terminalId (or primary tab)
  activeTerminalIdPerSession = $state<Record<string, string>>({});
  // Map of sessionId -> TerminalSplitGroup[]
  sessionSplitGroups = $state<Record<string, TerminalSplitGroup[]>>({});
  // Map of sessionId -> which termId has focused keyboard input
  focusedPaneTermId = $state<Record<string, string>>({});
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

  constructor() {
    this.loadFromStorage();
  }

  private loadFromStorage(): void {
    if (typeof window === 'undefined' || !window.localStorage) return;
    try {
      const raw = window.localStorage.getItem(STORAGE_KEY);
      if (!raw) return;
      const data: PersistedTerminalState = JSON.parse(raw);
      if (data && data.version === 1) {
        if (data.sessionTerminals) this.sessionTerminals = data.sessionTerminals;
        if (data.activeTerminalIdPerSession) this.activeTerminalIdPerSession = data.activeTerminalIdPerSession;
        if (data.sessionSplitGroups) this.sessionSplitGroups = data.sessionSplitGroups;
        if (data.focusedPaneTermId) this.focusedPaneTermId = data.focusedPaneTermId;
        if (data.sessionCollapsed) this.sessionCollapsed = data.sessionCollapsed;
        if (data.sessionDockPosition) this.sessionDockPosition = data.sessionDockPosition;
        if (typeof data.panelHeight === 'number') this.panelHeight = data.panelHeight;
        if (typeof data.panelWidth === 'number') this.panelWidth = data.panelWidth;
      }
    } catch (err) {
      console.warn('Failed to load terminal layout from storage:', err);
    }
  }

  private saveToStorage(): void {
    if (typeof window === 'undefined' || !window.localStorage) return;
    try {
      const payload: PersistedTerminalState = {
        version: 1,
        sessionTerminals: this.sessionTerminals,
        activeTerminalIdPerSession: this.activeTerminalIdPerSession,
        sessionSplitGroups: this.sessionSplitGroups,
        focusedPaneTermId: this.focusedPaneTermId,
        sessionCollapsed: this.sessionCollapsed,
        sessionDockPosition: this.sessionDockPosition,
        panelHeight: this.panelHeight,
        panelWidth: this.panelWidth
      };
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(payload));
    } catch (err) {
      console.warn('Failed to save terminal layout to storage:', err);
    }
  }

  getTerminalTabs(sessionId: string): TerminalTab[] {
    return this.sessionTerminals[sessionId] || [];
  }

  getActiveTerminalId(sessionId: string): string | null {
    return this.activeTerminalIdPerSession[sessionId] || null;
  }

  getFocusedPaneId(sessionId: string): string | null {
    return this.focusedPaneTermId[sessionId] || this.getActiveTerminalId(sessionId);
  }

  setFocusedPaneId(sessionId: string, termId: string): void {
    if (!sessionId || !termId) return;
    this.focusedPaneTermId[sessionId] = termId;
    this.saveToStorage();
  }

  getSplitGroups(sessionId: string): TerminalSplitGroup[] {
    return this.sessionSplitGroups[sessionId] || [];
  }

  /**
   * Returns the split group containing the active terminal tab,
   * or a synthetic single-terminal group if it is not split.
   */
  getActiveSplitGroup(sessionId: string): TerminalSplitGroup | null {
    const activeId = this.getActiveTerminalId(sessionId);
    if (!activeId) return null;

    const groups = this.getSplitGroups(sessionId);
    const found = groups.find((g) => g.paneTermIds.includes(activeId));
    if (found) {
      return found;
    }

    // Default 1-pane group
    return {
      id: `group_${activeId}`,
      paneTermIds: [activeId],
      splitDirection: this.getDockPosition(sessionId) === 'right' ? 'vertical' : 'horizontal',
      paneSizes: [100]
    };
  }

  isSessionCollapsed(sessionId: string): boolean {
    if (!sessionId) return false;
    return this.sessionCollapsed[sessionId] ?? false;
  }

  toggleSessionCollapse(sessionId: string, collapsed?: boolean): void {
    if (!sessionId) return;
    const current = this.isSessionCollapsed(sessionId);
    this.sessionCollapsed[sessionId] = collapsed !== undefined ? collapsed : !current;
    this.saveToStorage();
  }

  getDockPosition(sessionId: string): TerminalDockPosition {
    if (!sessionId) return 'bottom';
    return this.sessionDockPosition[sessionId] || 'bottom';
  }

  setDockPosition(sessionId: string, pos: TerminalDockPosition): void {
    if (!sessionId) return;
    this.sessionDockPosition[sessionId] = pos;
    this.saveToStorage();
  }

  toggleDockPosition(sessionId: string): void {
    if (!sessionId) return;
    const current = this.getDockPosition(sessionId);
    this.sessionDockPosition[sessionId] = current === 'bottom' ? 'right' : 'bottom';
    this.saveToStorage();
  }

  toggleOpen(open?: boolean): void {
    this.isOpen = open !== undefined ? open : !this.isOpen;
    this.saveToStorage();
  }

  setPanelHeight(height: number): void {
    this.panelHeight = Math.max(140, Math.min(height, window.innerHeight - 150));
    this.saveToStorage();
  }

  setPanelWidth(width: number): void {
    this.panelWidth = Math.max(260, Math.min(width, window.innerWidth - 350));
    this.saveToStorage();
  }

  renameTerminal(sessionId: string, termId: string, newTitle: string): void {
    if (!sessionId || !termId) return;
    const list = this.sessionTerminals[sessionId];
    if (!list) return;
    const tab = list.find((t) => t.id === termId);
    if (tab && newTitle.trim()) {
      const old = tab.title;
      tab.title = newTitle.trim();
      logger.info('TERMINAL', `Renamed terminal tab ${termId} from "${old}" to "${tab.title}"`);
      this.saveToStorage();
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
    this.focusedPaneTermId[sessionId] = termId;
    this.saveToStorage();
    logger.info('TERMINAL', `Created terminal tab: ${tab.title} (${termId}) for session ${sessionId}`);

    // Call Go backend to instantiate PTY with isolated Process Group
    if (window.go?.main?.App?.CreateTerminal) {
      window.go.main.App.CreateTerminal(sessionId, termId, cwd, '')
        .catch((err) => {
          logger.error('TERMINAL', `Failed to create backend terminal: ${termId}`, err);
          console.error('Failed to create backend terminal:', err);
        });
    }

    return tab;
  }

  /**
   * Splits the currently active terminal or creates a new one side-by-side.
   */
  splitNewTerminal(sessionId: string, cwd: string): TerminalTab {
    const newTab = this.createTerminal(sessionId, cwd);
    const activeSplit = this.getActiveSplitGroup(sessionId);

    if (activeSplit) {
      // Add new terminal into current split group
      const existingPanes = activeSplit.paneTermIds.filter((id) => id !== newTab.id);
      const updatedPanes = [...existingPanes, newTab.id];
      const equalSize = 100 / updatedPanes.length;
      const newSizes = updatedPanes.map(() => equalSize);

      const groups = this.getSplitGroups(sessionId).filter((g) => g.id !== activeSplit.id);
      const updatedGroup: TerminalSplitGroup = {
        ...activeSplit,
        paneTermIds: updatedPanes,
        paneSizes: newSizes
      };
      this.sessionSplitGroups[sessionId] = [...groups, updatedGroup];
      this.activeTerminalIdPerSession[sessionId] = updatedPanes[0];
      this.focusedPaneTermId[sessionId] = newTab.id;
      this.saveToStorage();
    }

    return newTab;
  }

  /**
   * Splits an existing source terminal tab with a target terminal tab in a given direction/position
   */
  splitTerminal(
    sessionId: string,
    sourceTermId: string,
    targetTermId: string,
    position: 'left' | 'right' | 'top' | 'bottom'
  ): void {
    if (!sessionId || !sourceTermId || !targetTermId || sourceTermId === targetTermId) return;

    const groups = this.getSplitGroups(sessionId);
    const targetGroup = groups.find((g) => g.paneTermIds.includes(targetTermId));

    const direction: 'horizontal' | 'vertical' =
      position === 'top' || position === 'bottom' ? 'vertical' : 'horizontal';

    // Remove source from existing group if any
    const cleanedGroups = groups
      .map((g) => {
        if (!g.paneTermIds.includes(sourceTermId)) return g;
        const remaining = g.paneTermIds.filter((id) => id !== sourceTermId);
        const equalSize = 100 / Math.max(1, remaining.length);
        return {
          ...g,
          paneTermIds: remaining,
          paneSizes: remaining.map(() => equalSize)
        };
      })
      .filter((g) => g.paneTermIds.length > 1);

    if (targetGroup) {
      // Add source next to target in existing group
      const panes = targetGroup.paneTermIds.filter((id) => id !== sourceTermId);
      const targetIdx = panes.indexOf(targetTermId);
      const insertIdx = position === 'right' || position === 'bottom' ? targetIdx + 1 : targetIdx;
      panes.splice(Math.max(0, insertIdx), 0, sourceTermId);

      const equalSize = 100 / panes.length;
      const newSizes = panes.map(() => equalSize);

      const remainingGroups = cleanedGroups.filter((g) => g.id !== targetGroup.id);
      const updatedTargetGroup: TerminalSplitGroup = {
        ...targetGroup,
        paneTermIds: panes,
        splitDirection: direction,
        paneSizes: newSizes
      };

      this.sessionSplitGroups[sessionId] = [...remainingGroups, updatedTargetGroup];
    } else {
      // Create new split group with target and source
      const panes =
        position === 'right' || position === 'bottom'
          ? [targetTermId, sourceTermId]
          : [sourceTermId, targetTermId];

      const newGroup: TerminalSplitGroup = {
        id: `split_${Date.now()}_${Math.random().toString(36).substring(2, 6)}`,
        paneTermIds: panes,
        splitDirection: direction,
        paneSizes: [50, 50]
      };

      this.sessionSplitGroups[sessionId] = [...cleanedGroups, newGroup];
    }

    this.activeTerminalIdPerSession[sessionId] = targetTermId;
    this.focusedPaneTermId[sessionId] = sourceTermId;
    this.saveToStorage();
    logger.info('TERMINAL', `Split terminal ${sourceTermId} into group with ${targetTermId} (${position})`);
  }

  /**
   * Unsplits a terminal, removing it from its split group back into a standalone tab
   */
  unsplitTerminal(sessionId: string, termId: string): void {
    if (!sessionId || !termId) return;

    const groups = this.getSplitGroups(sessionId);
    const updatedGroups = groups
      .map((g) => {
        if (!g.paneTermIds.includes(termId)) return g;
        const remaining = g.paneTermIds.filter((id) => id !== termId);
        const equalSize = 100 / Math.max(1, remaining.length);
        return {
          ...g,
          paneTermIds: remaining,
          paneSizes: remaining.map(() => equalSize)
        };
      })
      .filter((g) => g.paneTermIds.length > 1); // Groups with <= 1 are dissolved to single tabs

    this.sessionSplitGroups[sessionId] = updatedGroups;
    this.activeTerminalIdPerSession[sessionId] = termId;
    this.focusedPaneTermId[sessionId] = termId;
    this.saveToStorage();
  }

  /**
   * Updates proportional pane percentage sizes for a group
   */
  setGroupPaneSizes(sessionId: string, groupId: string, sizes: number[]): void {
    const groups = this.getSplitGroups(sessionId);
    const group = groups.find((g) => g.id === groupId);
    if (group) {
      group.paneSizes = sizes;
      this.saveToStorage();
    }
  }

  switchTerminal(sessionId: string, termId: string): void {
    this.activeTerminalIdPerSession[sessionId] = termId;
    this.focusedPaneTermId[sessionId] = termId;
    this.saveToStorage();
  }

  closeTerminal(sessionId: string, termId: string): void {
    const list = this.sessionTerminals[sessionId] || [];
    const idx = list.findIndex((t) => t.id === termId);
    if (idx === -1) return;

    logger.info('TERMINAL', `Closing terminal tab: ${termId} for session ${sessionId}`);

    // Call Go backend to SIGKILL entire process group
    if (window.go?.main?.App?.CloseTerminal) {
      window.go.main.App.CloseTerminal(termId)
        .catch((err) => {
          logger.error('TERMINAL', `Failed to close backend terminal: ${termId}`, err);
          console.error('Failed to close backend terminal:', err);
        });
    }

    // Clean up split groups
    const groups = this.getSplitGroups(sessionId);
    const updatedGroups = groups
      .map((g) => {
        if (!g.paneTermIds.includes(termId)) return g;
        const remaining = g.paneTermIds.filter((id) => id !== termId);
        const equalSize = 100 / Math.max(1, remaining.length);
        return {
          ...g,
          paneTermIds: remaining,
          paneSizes: remaining.map(() => equalSize)
        };
      })
      .filter((g) => g.paneTermIds.length > 1);

    this.sessionSplitGroups[sessionId] = updatedGroups;

    const updated = list.filter((t) => t.id !== termId);
    this.sessionTerminals[sessionId] = updated;

    if (this.activeTerminalIdPerSession[sessionId] === termId) {
      if (updated.length > 0) {
        const nextIdx = Math.max(0, idx - 1);
        this.activeTerminalIdPerSession[sessionId] = updated[nextIdx].id;
        this.focusedPaneTermId[sessionId] = updated[nextIdx].id;
      } else {
        delete this.activeTerminalIdPerSession[sessionId];
        delete this.focusedPaneTermId[sessionId];
        this.isOpen = false;
      }
    } else if (this.focusedPaneTermId[sessionId] === termId) {
      this.focusedPaneTermId[sessionId] = this.activeTerminalIdPerSession[sessionId] || '';
    }

    this.saveToStorage();
  }

  closeAllForSession(sessionId: string): void {
    if (window.go?.main?.App?.CloseSessionTerminals) {
      window.go.main.App.CloseSessionTerminals(sessionId)
        .catch((err) => console.error('Failed to close session terminals:', err));
    }
    delete this.sessionTerminals[sessionId];
    delete this.activeTerminalIdPerSession[sessionId];
    delete this.focusedPaneTermId[sessionId];
    delete this.sessionSplitGroups[sessionId];
    delete this.sessionCollapsed[sessionId];
    delete this.sessionDockPosition[sessionId];
    this.saveToStorage();
  }
}

export const terminalStore = new TerminalStore();
