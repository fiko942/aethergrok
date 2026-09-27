/**
 * Terminal state management per session with multi-tab terminal & split pane support
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
      const old = tab.title;
      tab.title = newTitle.trim();
      logger.info('TERMINAL', `Renamed terminal tab ${termId} from "${old}" to "${tab.title}"`);
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
    position: 'left' | 'right' | 'top' | 'bottom' = 'right'
  ): void {
    if (!sessionId || !sourceTermId || !targetTermId || sourceTermId === targetTermId) return;

    logger.info('TERMINAL', `Splitting terminal ${sourceTermId} into ${targetTermId} (${position})`);

    const groups = this.getSplitGroups(sessionId);
    // Find if target is in an existing group
    let targetGroup = groups.find((g) => g.paneTermIds.includes(targetTermId));

    // Remove sourceTermId from any existing group first
    const cleanedGroups = groups
      .map((g) => ({
        ...g,
        paneTermIds: g.paneTermIds.filter((id) => id !== sourceTermId)
      }))
      .filter((g) => g.paneTermIds.length > 0);

    const isVertical = position === 'top' || position === 'bottom';
    const splitDirection = isVertical ? 'vertical' : 'horizontal';

    if (targetGroup) {
      // Insert source relative to target
      const panes = [...targetGroup.paneTermIds.filter((id) => id !== sourceTermId)];
      const targetIdx = panes.indexOf(targetTermId);
      if (position === 'left' || position === 'top') {
        panes.splice(Math.max(0, targetIdx), 0, sourceTermId);
      } else {
        panes.splice(targetIdx + 1, 0, sourceTermId);
      }

      const equalSize = 100 / panes.length;
      const updatedTargetGroup: TerminalSplitGroup = {
        ...targetGroup,
        paneTermIds: panes,
        splitDirection,
        paneSizes: panes.map(() => equalSize)
      };

      this.sessionSplitGroups[sessionId] = [
        ...cleanedGroups.filter((g) => g.id !== targetGroup!.id),
        updatedTargetGroup
      ];
      this.activeTerminalIdPerSession[sessionId] = panes[0];
      this.focusedPaneTermId[sessionId] = sourceTermId;
    } else {
      // Create new split group between target and source
      const panes =
        position === 'left' || position === 'top'
          ? [sourceTermId, targetTermId]
          : [targetTermId, sourceTermId];

      const newGroup: TerminalSplitGroup = {
        id: `group_${Date.now()}_${Math.random().toString(36).substring(2, 6)}`,
        paneTermIds: panes,
        splitDirection,
        paneSizes: [50, 50]
      };

      this.sessionSplitGroups[sessionId] = [...cleanedGroups, newGroup];
      this.activeTerminalIdPerSession[sessionId] = panes[0];
      this.focusedPaneTermId[sessionId] = sourceTermId;
    }
  }

  /**
   * Unsplits a terminal pane, detaching it back into its own standalone tab
   */
  unsplitTerminal(sessionId: string, termId: string): void {
    if (!sessionId || !termId) return;

    logger.info('TERMINAL', `Unsplitting terminal ${termId} back to standalone tab`);

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
  }

  /**
   * Updates proportional pane percentage sizes for a group
   */
  setGroupPaneSizes(sessionId: string, groupId: string, sizes: number[]): void {
    const groups = this.getSplitGroups(sessionId);
    const group = groups.find((g) => g.id === groupId);
    if (group) {
      group.paneSizes = sizes;
    }
  }

  switchTerminal(sessionId: string, termId: string): void {
    this.activeTerminalIdPerSession[sessionId] = termId;
    this.focusedPaneTermId[sessionId] = termId;
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
  }
}

export const terminalStore = new TerminalStore();

