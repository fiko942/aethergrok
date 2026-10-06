import { logger } from './logger.svelte';
import { terminalStore } from './terminal.svelte';

export type SessionStatus = 'working' | 'waiting_permission' | 'finished' | 'error' | 'idle';

export interface DiffData {
  oldPath?: string;
  newPath?: string;
  oldContent?: string;
  newContent?: string;
  diffUnified?: string; // Unified diff string (patch format)
  addedCount?: number;
  removedCount?: number;
}

export interface ToolCall {
  id: string;
  tool: string; // 'read_file' | 'write' | 'search_replace' | 'run_terminal_cmd' | 'bash' | string
  params?: Record<string, unknown> | string;
  result?: string;
  status: 'running' | 'completed' | 'error' | 'pending';
  startTime?: number;
  endTime?: number;
  diff?: DiffData;
}

export interface PermissionRequest {
  sessionId: string;
  requestId: string;
  toolName: string;
  description: string;
  details?: Record<string, unknown>;
  options?: Array<{ id: string; label: string }>;
}

export interface VisionImage {
  id: string;
  filePath: string;
  dataUrl: string;
  sizeBytes?: number;
  width?: number;
  height?: number;
  timestamp: number;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: number;
  tokens?: {
    input?: number;
    output?: number;
    total?: number;
  };
  toolCalls?: ToolCall[];
  images?: VisionImage[];
  attachments?: AttachedFile[];
  reasoningContent?: string;
  status?: 'streaming' | 'done' | 'error';
  isSteer?: boolean;
  errorKind?: 'max_tokens_truncation' | 'general_error';
  errorDetails?: {
    numTurns?: number;
    totalTokens?: number;
    model?: string;
    rawJson?: string;
  };
}

export interface SessionUsage {
  usedTokens: number;
  maxTokens: number;
  lastTurnInput: number;
  lastTurnOutput: number;
  lastTurnCacheRead: number;
  lastTurnReasoning: number;
  lastTurnModelCalls: number;
  totalInput: number;
  totalOutput: number;
  totalCacheRead: number;
  turnCount: number;
  primaryModelId: string;
}

export interface AttachedFile {
  id: string;
  name: string;
  filePath: string;
  sizeBytes?: number;
  mimeType?: string;
  content?: string; // Text/markdown/code content
  dataUrl?: string; // Base64 data URL for images
  isImage: boolean;
  timestamp: number;
}

export interface QueuedPrompt {
  id: string;
  text: string;
  images: VisionImage[];
  attachments: AttachedFile[];
  model: string;
  reasoningEffort: 'low' | 'medium' | 'high';
  agentMode?: string;
  timestamp: number;
}

export interface SessionDraft {
  text: string;
  images: VisionImage[];
  attachments: AttachedFile[];
}

export type RightSidebarTab = 'files' | 'changes' | 'plan';

export interface Session {
  id: string;
  grokSessionId?: string; // Real UUID discovered from Grok CLI execution
  workspaceId: string;
  title: string;
  isCustomTitle?: boolean; // Set to true when renamed manually by user so auto-naming won't overwrite it
  status: SessionStatus;
  createdAt: number;
  updatedAt: number;
  messages: ChatMessage[];
  // Windowing state per session
  visibleTurnCount: number; // Number of recent user turns currently in DOM window (default 4)
  pendingPermission?: PermissionRequest | null;
  isPinned?: boolean;
  pinnedAt?: number;
  usage?: SessionUsage;
  queuedPrompts?: QueuedPrompt[];
  draft?: SessionDraft;
  // Tab-isolated Right Sidebar State
  rightSidebarOpen?: boolean;
  rightSidebarTab?: RightSidebarTab;
  // Per-session execution mode ('agent' | 'plan' | 'yolo')
  agentMode?: 'agent' | 'plan' | 'yolo';
  // Auto-retry attempt counter for transient failures (e.g. max_tokens_truncation)
  autoRetryCount?: number;
}

export interface TokenTruncationDetails {
  numTurns?: number;
  totalTokens?: number;
  model?: string;
  rawJson?: string;
}

export function parseTokenTruncationDetails(rawError: string): TokenTruncationDetails | null {
  if (!rawError) return null;
  const lower = rawError.toLowerCase();
  const isTrunc = lower.includes('max_tokens') || lower.includes('response truncated') || lower.includes('context limit reached');
  if (!isTrunc) return null;

  let numTurns: number | undefined;
  let totalTokens: number | undefined;
  let model: string | undefined;

  const matchTurns = rawError.match(/"numTurns":\s*(\d+)/i) || rawError.match(/after (\d+) tool iterations/i);
  if (matchTurns) numTurns = parseInt(matchTurns[1], 10);

  const matchTokens = rawError.match(/"totalTokens":\s*(\d+)/i) || rawError.match(/\((\d+) accumulated tokens\)/i);
  if (matchTokens) totalTokens = parseInt(matchTokens[1], 10);

  const matchModel = rawError.match(/"([a-zA-Z0-9\.\-_]+)":\s*{\s*"inputTokens"/);
  if (matchModel) model = matchModel[1];

  return {
    numTurns,
    totalTokens,
    model,
    rawJson: rawError
  };
}

export interface WorkspaceFolder {
  id: string;
  name: string; // e.g. "affilia", "grok-desktop"
  path: string; // e.g. "/Users/fiko942/Desktop/affilia"
  createdAt: number;
  isExpanded?: boolean;
  existsOnDisk?: boolean;
}

export const STATUS_META: Record<SessionStatus, { label: string; dotClass: string; hex: string; icon: string }> = {
  working: {
    label: 'Working',
    dotClass: 'bg-ant-primary animate-pulse',
    hex: '#1677ff',
    icon: '🔵'
  },
  waiting_permission: {
    label: 'Waiting Permission',
    dotClass: 'bg-ant-warning animate-bounce',
    hex: '#faad14',
    icon: '🟡'
  },
  finished: {
    label: 'Finished',
    dotClass: 'bg-ant-success',
    hex: '#52c41a',
    icon: '🟢'
  },
  error: {
    label: 'Error',
    dotClass: 'bg-ant-error',
    hex: '#ff4d4f',
    icon: '🔴'
  },
  idle: {
    label: 'Idle',
    dotClass: 'bg-ant-text-muted',
    hex: '#6b7280',
    icon: '⚪'
  }
};

export const DEFAULT_WINDOW_TURNS = 4;
export const PREPEND_CHUNK_TURNS = 4;

const WORKSPACES_STORAGE_KEY = 'aethergrok_workspaces_v1';
const SESSIONS_STORAGE_KEY = 'aethergrok_sessions_v1';
const OPEN_TABS_STORAGE_KEY = 'aethergrok_open_tabs_v1';
const ACTIVE_SESSION_STORAGE_KEY = 'aethergrok_active_session_id_v1';
const ACTIVE_WS_STORAGE_KEY = 'aethergrok_active_workspace_id_v1';

class SessionStore {
  workspaces = $state<WorkspaceFolder[]>([]);
  activeWorkspaceId = $state<string>('');
  sessions = $state<Session[]>([]);
  activeSessionId = $state<string | null>(null);
  openTabSessionIds = $state<string[]>([]); // Track sessions opened as active tabs

  // Multi-select state
  isSelectionMode = $state<boolean>(false);
  selectedSessionIds = $state<Set<string>>(new Set());

  // Compaction state
  isCompacting = $state<boolean>(false);
  lastCompactNotice = $state<{ type: 'in_progress' | 'success' | 'error'; message: string; tokensBefore?: number; tokensAfter?: number } | null>(null);

  constructor() {
    this.loadWorkspacesFromStorage();
    this.loadSessionsFromStorage();

    // Workspaces default to collapsed unless explicitly expanded by user
    for (const ws of this.workspaces) {
      if (ws.isExpanded === undefined) {
        ws.isExpanded = false;
      }
    }

    if (this.workspaces.length === 0) {
      this.activeWorkspaceId = '';
    } else if (!this.activeWorkspaceId || !this.workspaces.some(w => w.id === this.activeWorkspaceId)) {
      this.activeWorkspaceId = this.workspaces[0].id;
    }

    // Validate activeSessionId and openTabSessionIds against actual sessions
    this.reconcileSessionState();

    // Trigger verification of workspaces and active session hydration on disk
    if (typeof window !== 'undefined') {
      setTimeout(async () => {
        await this.hydrateFromBackend();
        logger.info('SESSION', 'Starting session and workspace reconciliation cycle');
        await this.verifyAllWorkspaces();
        this.reconcileSessionState();
        await this.verifySessionsOnDisk();

        // Auto-hydrate the active session transcript and token usage stats
        if (this.activeSessionId) {
          const activeSess = this.sessions.find((s) => s.id === this.activeSessionId);
          if (activeSess) {
            logger.info('SESSION', `Auto-hydrating active session on launch: ${activeSess.id}`, { title: activeSess.title });
            if (activeSess.messages.length === 0 && activeSess.status !== 'working') {
              await this.loadSessionHistoryFromDisk(activeSess);
            }
            await this.loadSessionUsage(activeSess);
          }
        }
      }, 50);
    }
  }

  async hydrateFromBackend(): Promise<void> {
    const win = typeof window !== 'undefined' ? (window as any) : null;
    if (!win?.go?.main?.App) return;

    try {
      // 1. Load workspaces from backend storage
      if (win.go.main.App.GetWorkspaces) {
        const storedWs = await win.go.main.App.GetWorkspaces();
        if (Array.isArray(storedWs) && storedWs.length > 0) {
          this.workspaces = storedWs.map((w: any) => ({
            id: w.id,
            name: w.name,
            path: w.path,
            createdAt: w.createdAt || Date.now(),
            isExpanded: false
          }));
        } else if (this.workspaces.length > 0 && win.go.main.App.SaveWorkspaces) {
          // Auto-migrate local workspaces to backend
          await win.go.main.App.SaveWorkspaces(this.workspaces.map(w => ({
            id: w.id,
            name: w.name,
            path: w.path,
            createdAt: w.createdAt
          })));
        }
      }

      // 2. Load UI state from backend storage
      if (win.go.main.App.GetUIState) {
        const uiState = await win.go.main.App.GetUIState();
        if (uiState) {
          if (uiState.activeWorkspaceId) {
            this.activeWorkspaceId = uiState.activeWorkspaceId;
          }
          if (Array.isArray(uiState.openTabSessionIds) && uiState.openTabSessionIds.length > 0) {
            this.openTabSessionIds = uiState.openTabSessionIds;
          }
          if (uiState.activeSessionId) {
            this.activeSessionId = uiState.activeSessionId;
          }
        }
      }
    } catch (err) {
      logger.error('SESSION', 'Failed to hydrate session/workspace state from backend', err);
    }
  }

  private loadWorkspacesFromStorage() {
    if (typeof window === 'undefined' || !window.localStorage) return;
    try {
      const raw = window.localStorage.getItem(WORKSPACES_STORAGE_KEY);
      if (raw) {
        const parsed = JSON.parse(raw);
        if (Array.isArray(parsed) && parsed.length > 0) {
          this.workspaces = parsed;
        }
      }
      const rawActiveWs = window.localStorage.getItem(ACTIVE_WS_STORAGE_KEY);
      if (rawActiveWs) {
        this.activeWorkspaceId = rawActiveWs;
      }
    } catch (e) {
      console.warn('Failed to load workspaces from storage:', e);
    }
  }

  private saveWorkspacesToStorage() {
    // 1. LocalStorage fallback
    if (typeof window !== 'undefined' && window.localStorage) {
      try {
        window.localStorage.setItem(WORKSPACES_STORAGE_KEY, JSON.stringify(this.workspaces));
        if (this.activeWorkspaceId) {
          window.localStorage.setItem(ACTIVE_WS_STORAGE_KEY, this.activeWorkspaceId);
        } else {
          window.localStorage.removeItem(ACTIVE_WS_STORAGE_KEY);
        }
      } catch (e) {
        console.warn('Failed to save workspaces to storage:', e);
      }
    }

    // 2. Persist to Go backend storage
    const win = typeof window !== 'undefined' ? (window as any) : null;
    if (win?.go?.main?.App?.SaveWorkspaces) {
      win.go.main.App.SaveWorkspaces(this.workspaces.map(w => ({
        id: w.id,
        name: w.name,
        path: w.path,
        createdAt: w.createdAt
      }))).catch(() => {});
    }
  }

  private loadSessionsFromStorage() {
    if (typeof window === 'undefined' || !window.localStorage) return;
    try {
      const rawSessions = window.localStorage.getItem(SESSIONS_STORAGE_KEY);
      if (rawSessions) {
        const parsed = JSON.parse(rawSessions);
        if (Array.isArray(parsed)) {
          // Sanitize any stale session or message states:
          // If the app was closed while a turn was running or stalled, reset 'working' or 'waiting_permission'
          // to 'idle' so the UI never starts up locked in an infinite loading spinner.
          for (const s of parsed) {
            if (s.status === 'working' || s.status === 'waiting_permission') {
              s.status = 'idle';
            }
            if (s.pendingPermission) {
              s.pendingPermission = null;
            }
            if (Array.isArray(s.messages)) {
              for (const m of s.messages) {
                if (m.status === 'streaming') {
                  m.status = 'done';
                }
                if (Array.isArray(m.toolCalls)) {
                  for (const tc of m.toolCalls) {
                    if (tc.status === 'running') {
                      tc.status = 'error';
                    }
                  }
                }
              }
            }
          }
          this.sessions = parsed;
          this.deduplicateSessions();
        }
      }

      const rawTabs = window.localStorage.getItem(OPEN_TABS_STORAGE_KEY);
      if (rawTabs) {
        const parsed = JSON.parse(rawTabs);
        if (Array.isArray(parsed)) {
          this.openTabSessionIds = parsed;
        }
      }

      const rawActiveSession = window.localStorage.getItem(ACTIVE_SESSION_STORAGE_KEY);
      if (rawActiveSession !== null) {
        this.activeSessionId = rawActiveSession || null;
      }

      this.deduplicateSessions();
    } catch (e) {
      console.warn('Failed to load sessions from storage:', e);
    }
  }

  // Deduplicate sessions by (workspaceId + grokSessionId) and session ID to prevent ghost duplicate proliferation
  deduplicateSessions(): void {
    const seenGrokIds = new Map<string, Session>();
    const seenIds = new Set<string>();
    const uniqueSessions: Session[] = [];

    for (const session of this.sessions) {
      if (!session || !session.id) continue;

      // Deduplicate by workspaceId and grokSessionId when grokSessionId exists
      const grokKey = session.grokSessionId ? `${session.workspaceId}:${session.grokSessionId}` : null;
      if (grokKey && seenGrokIds.has(grokKey)) {
        const existing = seenGrokIds.get(grokKey)!;
        // Merge messages and preserve best data
        if (session.messages && session.messages.length > existing.messages.length) {
          existing.messages = session.messages;
        }
        if (session.isCustomTitle) {
          existing.title = session.title;
          existing.isCustomTitle = true;
        }
        if (session.updatedAt > existing.updatedAt) {
          existing.updatedAt = session.updatedAt;
        }
        if (session.isPinned) {
          existing.isPinned = true;
        }
        // If the open tab references this duplicate session id, remap it to existing.id
        const tabIdx = this.openTabSessionIds.indexOf(session.id);
        if (tabIdx !== -1) {
          this.openTabSessionIds.splice(tabIdx, 1);
          if (!this.openTabSessionIds.includes(existing.id)) {
            this.openTabSessionIds.push(existing.id);
          }
        }
        if (this.activeSessionId === session.id) {
          this.activeSessionId = existing.id;
        }
        continue;
      }

      // Check if session.id is already in uniqueSessions
      if (seenIds.has(session.id)) {
        continue;
      }

      seenIds.add(session.id);
      if (grokKey) {
        seenGrokIds.set(grokKey, session);
      }
      uniqueSessions.push(session);
    }

    this.sessions = uniqueSessions;

    // Deduplicate and validate openTabSessionIds
    const validSessionIds = new Set(this.sessions.map((s) => s.id));
    const dedupedTabs: string[] = [];
    for (const tid of this.openTabSessionIds) {
      if (validSessionIds.has(tid) && !dedupedTabs.includes(tid)) {
        dedupedTabs.push(tid);
      }
    }
    this.openTabSessionIds = dedupedTabs;
  }

  saveSessionsToStorage() {
    // 1. LocalStorage fallback
    if (typeof window !== 'undefined' && window.localStorage) {
      try {
        window.localStorage.setItem(SESSIONS_STORAGE_KEY, JSON.stringify(this.sessions));
        window.localStorage.setItem(OPEN_TABS_STORAGE_KEY, JSON.stringify(this.openTabSessionIds));
        if (this.activeSessionId) {
          window.localStorage.setItem(ACTIVE_SESSION_STORAGE_KEY, this.activeSessionId);
        } else {
          window.localStorage.removeItem(ACTIVE_SESSION_STORAGE_KEY);
        }
      } catch (e) {
        console.warn('Failed to save sessions to storage:', e);
      }
    }

    // 2. Persist UIState (open tabs and active workspace/session) to backend
    const win = typeof window !== 'undefined' ? (window as any) : null;
    if (win?.go?.main?.App?.SaveUIState) {
      win.go.main.App.SaveUIState({
        activeWorkspaceId: this.activeWorkspaceId,
        activeSessionId: this.activeSessionId || '',
        openTabSessionIds: this.openTabSessionIds
      }).catch(() => {});
    }
  }

  // Ensure tabs and active session point to existing sessions
  reconcileSessionState() {
    // Keep only open tab IDs that exist in this.sessions and belong to workspaces that are not missing
    const missingWsIds = new Set(
      this.workspaces.filter((w) => w.existsOnDisk === false).map((w) => w.id)
    );

    this.openTabSessionIds = this.openTabSessionIds.filter((id) => {
      const sess = this.sessions.find((s) => s.id === id);
      return sess && !missingWsIds.has(sess.workspaceId);
    });

    // If activeSessionId is set but doesn't exist or is missing, pick another open tab or null
    const activeSess = this.activeSessionId ? this.sessions.find((s) => s.id === this.activeSessionId) : null;
    if (this.activeSessionId && (!activeSess || missingWsIds.has(activeSess.workspaceId))) {
      if (this.openTabSessionIds.length > 0) {
        this.activeSessionId = this.openTabSessionIds[0];
      } else {
        this.activeSessionId = null;
      }
    }

    // If activeSessionId is set but not in openTabSessionIds, add it
    if (this.activeSessionId && !this.openTabSessionIds.includes(this.activeSessionId)) {
      this.openTabSessionIds.push(this.activeSessionId);
    }

    this.saveSessionsToStorage();
  }

  async verifySessionsOnDisk(): Promise<void> {
    if (typeof window === 'undefined') return;
    if (!window.go?.main?.App?.DiscoverGrokSessions) return;

    for (const ws of this.workspaces) {
      try {
        const diskSessions = await window.go.main.App.DiscoverGrokSessions(ws.path);
        if (diskSessions) {
          this.syncDiscoveredGrokSessions(ws.id, diskSessions);
        }
      } catch (err) {
        console.warn('verifySessionsOnDisk failed for', ws.path, err);
      }
    }
  }

  async verifyAllWorkspaces(): Promise<void> {
    if (typeof window === 'undefined') return;

    for (const ws of this.workspaces) {
      if (window.go?.main?.App?.CheckDirectoryExists) {
        try {
          const exists = await window.go.main.App.CheckDirectoryExists(ws.path);
          ws.existsOnDisk = exists;
        } catch {
          ws.existsOnDisk = true;
        }
      } else {
        // Fallback for browser preview
        ws.existsOnDisk = true;
      }
    }
    this.saveWorkspacesToStorage();
  }

  private generateSessionUUID(): string {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
      return crypto.randomUUID();
    }
    // Fallback standard RFC4122 v4 UUID generator
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
      const r = (Math.random() * 16) | 0;
      const v = c === 'x' ? r : (r & 0x3) | 0x8;
      return v.toString(16);
    });
  }

  private createNewSessionModel(title?: string, wsId?: string): Session {
    const id = this.generateSessionUUID();
    const workspaceId = wsId || this.activeWorkspaceId || (this.workspaces[0]?.id ?? '');
    return {
      id,
      grokSessionId: id,
      workspaceId,
      title: title || (workspaceId ? `Session ${this.sessions.filter((s) => s.workspaceId === workspaceId).length + 1}` : `Session ${this.sessions.length + 1}`),
      status: 'idle',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: [],
      visibleTurnCount: DEFAULT_WINDOW_TURNS,
      pendingPermission: null,
      usage: {
        usedTokens: 2500, // Accurate baseline token calculation (system prompt, tools & harness overhead)
        maxTokens: 200000,
        lastTurnInput: 2500,
        lastTurnOutput: 0,
        lastTurnCacheRead: 0,
        lastTurnReasoning: 0,
        lastTurnModelCalls: 0,
        totalInput: 2500,
        totalOutput: 0,
        totalCacheRead: 0,
        turnCount: 0,
        primaryModelId: '9router'
      }
    };
  }

  // Active Workspace Getter
  get activeWorkspace(): WorkspaceFolder | undefined {
    return this.workspaces.find((w) => w.id === this.activeWorkspaceId);
  }

  // Sessions filtered by active workspace
  get activeWorkspaceSessions(): Session[] {
    if (!this.activeWorkspaceId) return this.sessions;
    return this.sessions.filter((s) => s.workspaceId === this.activeWorkspaceId);
  }

  // All open tabs across active workspaces (excluding sessions from missing workspaces)
  get openTabs(): Session[] {
    // Collect valid workspace IDs that either exist on disk or are active
    const validWorkspaceIds = new Set(
      this.workspaces
        .filter((w) => w.existsOnDisk !== false)
        .map((w) => w.id)
    );

    // Map open tab session IDs to actual session objects, preserving order
    const validSessions: Session[] = [];
    for (const tabId of this.openTabSessionIds) {
      const session = this.sessions.find((s) => s.id === tabId);
      if (session && validWorkspaceIds.has(session.workspaceId)) {
        validSessions.push(session);
      }
    }
    return validSessions;
  }

  // Open tabs belonging to active workspace (kept for backward compatibility)
  get openWorkspaceTabs(): Session[] {
    return this.openTabs;
  }

  // Active Session Getter
  get activeSession(): Session | undefined {
    return this.sessions.find((s) => s.id === this.activeSessionId);
  }

  // Active Messages window calculation
  get activeSessionMessages(): ChatMessage[] {
    const session = this.activeSession;
    if (!session) return [];
    return session.messages;
  }

  // Count user turns in active session
  get totalUserTurns(): number {
    const session = this.activeSession;
    if (!session) return 0;
    return session.messages.filter((m) => m.role === 'user').length;
  }

  // Messages sliced by the 10-turn (or expanded) active window
  get visibleMessages(): ChatMessage[] {
    const session = this.activeSession;
    if (!session) return [];
    
    const messages = session.messages;
    if (messages.length === 0) return [];

    // Find user turn indices from the end
    const userTurnIndices: number[] = [];
    for (let i = 0; i < messages.length; i++) {
      if (messages[i].role === 'user') {
        userTurnIndices.push(i);
      }
    }

    const totalTurns = userTurnIndices.length;
    if (totalTurns <= session.visibleTurnCount) {
      return messages;
    }

    // Cutoff is the start of the (totalTurns - visibleTurnCount)th user turn
    const cutoffTurnIdx = totalTurns - session.visibleTurnCount;
    const startIndex = userTurnIndices[cutoffTurnIdx];
    return messages.slice(startIndex);
  }

  // Remaining user turns that are parked before the current window
  get remainingHiddenTurns(): number {
    const total = this.totalUserTurns;
    const session = this.activeSession;
    if (!session) return 0;
    const remaining = total - session.visibleTurnCount;
    return remaining > 0 ? remaining : 0;
  }

  // Workspace Actions
  addWorkspace(name: string, path: string): WorkspaceFolder {
    const existing = this.workspaces.find((w) => w.path === path);
    if (existing) {
      existing.existsOnDisk = true;
      this.activeWorkspaceId = existing.id;
      this.saveWorkspacesToStorage();
      return existing;
    }

    const id = 'ws_' + Math.random().toString(36).substring(2, 9);
    const newWs: WorkspaceFolder = {
      id,
      name,
      path,
      createdAt: Date.now(),
      existsOnDisk: true,
      isExpanded: false
    };
    this.workspaces.push(newWs);
    this.activeWorkspaceId = id;
    this.saveWorkspacesToStorage();
    return newWs;
  }

  toggleWorkspaceExpanded(wsId: string): void {
    const ws = this.workspaces.find((w) => w.id === wsId);
    if (ws) {
      ws.isExpanded = ws.isExpanded === false ? true : false;
      this.saveWorkspacesToStorage();
    }
  }

  switchWorkspace(id: string): void {
    if (this.activeWorkspaceId === id) return;
    this.activeWorkspaceId = id;
    
    // Switch to first session in this workspace or create one
    const wsSessions = this.sessions.filter((s) => s.workspaceId === id);
    if (wsSessions.length > 0) {
      this.activeSessionId = wsSessions[0].id;
    } else {
      const ws = this.workspaces.find((w) => w.id === id);
      this.createSession(ws ? `Task for ${ws.name}` : 'New Task', id);
    }
  }

  removeWorkspace(id: string): void {
    const index = this.workspaces.findIndex((w) => w.id === id);
    if (index === -1) return;

    // Purge temporary files from disk for all sessions of this workspace
    const sessionsToRemove = this.sessions.filter((s) => s.workspaceId === id);
    this.purgeTempFilesForSessions(sessionsToRemove);

    this.workspaces.splice(index, 1);
    // Remove associated sessions
    this.sessions = this.sessions.filter((s) => s.workspaceId !== id);

    if (this.workspaces.length === 0) {
      this.activeWorkspaceId = '';
      this.activeSessionId = null;
      this.openTabSessionIds = [];
      this.saveWorkspacesToStorage();
      this.saveSessionsToStorage();
      return;
    }

    if (this.activeWorkspaceId === id) {
      this.activeWorkspaceId = this.workspaces[0].id;
      const wsSessions = this.sessions.filter((s) => s.workspaceId === this.activeWorkspaceId);
      this.activeSessionId = wsSessions[0]?.id || null;
    }
    this.saveWorkspacesToStorage();
  }

  // Selection Mode Actions
  toggleSelectionMode(enabled?: boolean): void {
    this.isSelectionMode = enabled !== undefined ? enabled : !this.isSelectionMode;
    if (!this.isSelectionMode) {
      this.selectedSessionIds = new Set();
    }
  }

  toggleSessionSelected(sessionId: string): void {
    const next = new Set(this.selectedSessionIds);
    if (next.has(sessionId)) {
      next.delete(sessionId);
    } else {
      next.add(sessionId);
    }
    this.selectedSessionIds = next;
  }

  selectAllSessions(customIds?: string[]): void {
    if (customIds) {
      this.selectedSessionIds = new Set(customIds);
      return;
    }
    // Select all sessions across all workspaces
    const allIds = this.sessions.map((s) => s.id);
    this.selectedSessionIds = new Set(allIds);
  }

  deselectAllSessions(): void {
    this.selectedSessionIds = new Set();
  }

  // Extract all temporary file paths (such as snapshots) linked to a session
  collectSessionTempFilePaths(session: Session): string[] {
    const paths: string[] = [];
    // From messages images
    for (const msg of session.messages) {
      if (msg.images) {
        for (const img of msg.images) {
          if (img.filePath) paths.push(img.filePath);
        }
      }
      if (msg.attachments) {
        for (const att of msg.attachments) {
          if (att.filePath) paths.push(att.filePath);
        }
      }
    }
    // From session draft
    if (session.draft?.images) {
      for (const img of session.draft.images) {
        if (img.filePath) paths.push(img.filePath);
      }
    }
    if (session.draft?.attachments) {
      for (const att of session.draft.attachments) {
        if (att.filePath) paths.push(att.filePath);
      }
    }
    // From queued prompts
    if (session.queuedPrompts) {
      for (const q of session.queuedPrompts) {
        if (q.images) {
          for (const img of q.images) {
            if (img.filePath) paths.push(img.filePath);
          }
        }
        if (q.attachments) {
          for (const att of q.attachments) {
            if (att.filePath) paths.push(att.filePath);
          }
        }
      }
    }
    return paths;
  }

  // Safely trigger backend deletion of temporary files
  private purgeTempFilesForSessions(sessions: Session[]): void {
    if (typeof window === 'undefined' || !window.go?.main?.App?.DeleteSessionTempFiles) return;
    const allPaths: string[] = [];
    for (const s of sessions) {
      allPaths.push(...this.collectSessionTempFilePaths(s));
    }
    if (allPaths.length > 0) {
      window.go.main.App.DeleteSessionTempFiles(allPaths).catch((err: any) => {
        console.warn('Failed to delete session temp files:', err);
      });
    }
  }

  deleteSelectedSessions(): void {
    if (this.selectedSessionIds.size === 0) return;
    const idsToDelete = new Set(this.selectedSessionIds);
    const sessionsToDelete = this.sessions.filter((s) => idsToDelete.has(s.id));

    // Purge temporary vision files on disk for all selected sessions
    this.purgeTempFilesForSessions(sessionsToDelete);

    this.sessions = this.sessions.filter((s) => !idsToDelete.has(s.id));
    this.selectedSessionIds = new Set();
    this.isSelectionMode = false;

    // Check if active session was deleted
    const currentWsSessions = this.activeWorkspaceSessions;
    if (currentWsSessions.length === 0) {
      this.createSession();
    } else if (!this.activeSessionId || idsToDelete.has(this.activeSessionId)) {
      this.activeSessionId = currentWsSessions[0].id;
    }
    this.saveSessionsToStorage();
  }

  // Right Sidebar Session-Isolated Toggle & Tab Switching
  toggleRightSidebar(sessionId?: string): void {
    const id = sessionId || this.activeSessionId;
    if (!id) return;
    const session = this.sessions.find((s) => s.id === id);
    if (session) {
      session.rightSidebarOpen = !session.rightSidebarOpen;
    }
  }

  setRightSidebarTab(tab: RightSidebarTab, sessionId?: string): void {
    const id = sessionId || this.activeSessionId;
    if (!id) return;
    const session = this.sessions.find((s) => s.id === id);
    if (session) {
      session.rightSidebarTab = tab;
      session.rightSidebarOpen = true;
    }
  }

  // Session Actions
  createSession(title?: string, wsId?: string): Session {
    const newSession = this.createNewSessionModel(title, wsId);
    this.sessions.push(newSession);
    this.openTabSessionIds.push(newSession.id);
    this.activeSessionId = newSession.id;
    this.saveSessionsToStorage();
    this.loadSessionUsage(newSession).catch(() => {});
    return newSession;
  }

  // Open a session in a tab (e.g. clicked from sidebar)
  async openSessionInTab(id: string): Promise<void> {
    const target = this.sessions.find((s) => s.id === id);
    if (!target) {
      logger.warn('UI', `Attempted to open non-existent session ID: ${id}`);
      return;
    }

    logger.info('UI', `Opening session in tab: ${target.title} (${id})`);

    if (!this.openTabSessionIds.includes(id)) {
      // Limit open tabs to 12 max to prevent visual overflow
      if (this.openTabSessionIds.length >= 12) {
        this.openTabSessionIds.shift();
      }
      this.openTabSessionIds.push(id);
    }

    target.visibleTurnCount = DEFAULT_WINDOW_TURNS;
    this.activeSessionId = id;
    if (target.workspaceId && target.workspaceId !== this.activeWorkspaceId) {
      this.activeWorkspaceId = target.workspaceId;
    }
    this.saveSessionsToStorage();

    // Lazy load real chat history from disk only if the session is NOT actively working
    // and currently has 0 in-memory messages. Never overwrite an in-flight prompt!
    if (target.messages.length === 0 && target.status !== 'working') {
      await this.loadSessionHistoryFromDisk(target);
    }

    // Load or update token usage stats for this session
    await this.loadSessionUsage(target);
  }

  // Load token usage stats from Go backend
  async loadSessionUsage(session: Session): Promise<void> {
    const ws = this.workspaces.find((w) => w.id === session.workspaceId) || this.activeWorkspace;
    if (!ws || !ws.path) return;

    if (window.go?.main?.App?.GetSessionUsage) {
      try {
        const targetId = session.grokSessionId || session.id;
        const stats = await window.go.main.App.GetSessionUsage(ws.path, targetId);
        if (stats) {
          if (stats.sessionId && !session.grokSessionId && !stats.sessionId.startsWith('sess_')) {
            session.grokSessionId = stats.sessionId;
          }
          session.usage = {
            usedTokens: stats.usedTokens || 0,
            maxTokens: stats.maxTokens || 200000,
            lastTurnInput: stats.lastTurnInput || 0,
            lastTurnOutput: stats.lastTurnOutput || 0,
            lastTurnCacheRead: stats.lastTurnCacheRead || 0,
            lastTurnReasoning: stats.lastTurnReasoning || 0,
            lastTurnModelCalls: stats.lastTurnModelCalls || 0,
            totalInput: stats.totalInput || 0,
            totalOutput: stats.totalOutput || 0,
            totalCacheRead: stats.totalCacheRead || 0,
            turnCount: stats.turnCount || session.messages.length,
            primaryModelId: stats.primaryModelId || ''
          };
          logger.debug('SESSION', `Loaded session usage: ${session.id}`, { usedTokens: session.usage.usedTokens });
        }
      } catch (err) {
        logger.error('SESSION', `Failed to load session usage for ${session.id}`, err);
        console.error('Failed to load session usage for', session.id, err);
      }
    }

    // Fallback: If session has messages but usedTokens is 0, estimate from message characters
    if ((!session.usage || session.usage.usedTokens === 0) && session.messages.length > 0) {
      let charCount = 0;
      for (const m of session.messages) {
        charCount += (m.content || '').length;
        if (m.reasoningContent) charCount += m.reasoningContent.length;
        if (m.toolCalls) {
          for (const tc of m.toolCalls) {
            charCount += (tc.result || '').length;
          }
        }
      }
      const estimated = Math.max(2500, Math.round(charCount / 4) + 2500);
      session.usage = {
        usedTokens: estimated,
        maxTokens: session.usage?.maxTokens || 200000,
        lastTurnInput: session.usage?.lastTurnInput || 0,
        lastTurnOutput: session.usage?.lastTurnOutput || 0,
        lastTurnCacheRead: session.usage?.lastTurnCacheRead || 0,
        lastTurnReasoning: session.usage?.lastTurnReasoning || 0,
        lastTurnModelCalls: session.usage?.lastTurnModelCalls || 0,
        totalInput: session.usage?.totalInput || estimated,
        totalOutput: session.usage?.totalOutput || 0,
        totalCacheRead: session.usage?.totalCacheRead || 0,
        turnCount: session.messages.length,
        primaryModelId: session.usage?.primaryModelId || ''
      };
    }
  }

  // Refresh token usage metrics for the active session
  async refreshActiveSessionUsage(): Promise<void> {
    const active = this.activeSession;
    if (active) {
      await this.loadSessionUsage(active);
    }
  }

  // Compact conversation via Go backend with rich loading state and notifications
  async compactActiveSession(): Promise<{ success: boolean; before: number; after: number; error?: string }> {
    const session = this.activeSession;
    if (!session) return { success: false, before: 0, after: 0, error: 'No active session' };
    const ws = this.workspaces.find((w) => w.id === session.workspaceId) || this.activeWorkspace;
    if (!ws || !ws.path) return { success: false, before: 0, after: 0, error: 'Workspace path not found' };

    if (this.isCompacting) {
      return { success: false, before: session.usage?.usedTokens || 0, after: session.usage?.usedTokens || 0, error: 'Compaction already in progress' };
    }

    const tokensBefore = session.usage?.usedTokens || 0;
    this.isCompacting = true;
    this.lastCompactNotice = {
      type: 'in_progress',
      message: `Compacting session conversation (${Math.round(tokensBefore / 1000)}K tokens)...`,
      tokensBefore
    };

    try {
      if (window.go?.main?.App?.CompactSession) {
        const targetId = session.grokSessionId || session.id;
        const stats = await window.go.main.App.CompactSession(ws.path, targetId);
        if (stats) {
          if (stats.sessionId && !session.grokSessionId && !stats.sessionId.startsWith('sess_')) {
            session.grokSessionId = stats.sessionId;
          }
          session.usage = {
            usedTokens: stats.usedTokens || Math.round(tokensBefore * 0.4),
            maxTokens: stats.maxTokens || 200000,
            lastTurnInput: stats.lastTurnInput || 0,
            lastTurnOutput: stats.lastTurnOutput || 0,
            lastTurnCacheRead: stats.lastTurnCacheRead || 0,
            lastTurnReasoning: stats.lastTurnReasoning || 0,
            lastTurnModelCalls: stats.lastTurnModelCalls || 0,
            totalInput: stats.totalInput || 0,
            totalOutput: stats.totalOutput || 0,
            totalCacheRead: stats.totalCacheRead || 0,
            turnCount: stats.turnCount || session.messages.length,
            primaryModelId: stats.primaryModelId || ''
          };
        }
        // Refresh session history from disk
        await this.loadSessionHistoryFromDisk(session);

        const tokensAfter = session.usage?.usedTokens || Math.round(tokensBefore * 0.4);
        this.lastCompactNotice = {
          type: 'success',
          message: `Conversation compacted successfully! Context reduced to ${Math.round(tokensAfter / 1000)}K tokens.`,
          tokensBefore,
          tokensAfter
        };

        // Auto clear notice after 5 seconds
        setTimeout(() => {
          if (this.lastCompactNotice?.type === 'success') {
            this.lastCompactNotice = null;
          }
        }, 5000);

        this.saveSessionsToStorage();
        return { success: true, before: tokensBefore, after: tokensAfter };
      } else {
        // Fallback preview mode simulation
        await new Promise((r) => setTimeout(r, 1200));
        const tokensAfter = Math.max(2500, Math.round(tokensBefore * 0.35));
        if (session.usage) {
          session.usage.usedTokens = tokensAfter;
        }
        this.lastCompactNotice = {
          type: 'success',
          message: `Conversation compacted successfully (Preview Mode)! Reduced from ${Math.round(tokensBefore / 1000)}K to ${Math.round(tokensAfter / 1000)}K tokens.`,
          tokensBefore,
          tokensAfter
        };
        setTimeout(() => {
          if (this.lastCompactNotice?.type === 'success') {
            this.lastCompactNotice = null;
          }
        }, 5000);
        return { success: true, before: tokensBefore, after: tokensAfter };
      }
    } catch (err) {
      console.error('Failed to compact session:', err);
      const errMsg = err instanceof Error ? err.message : String(err);
      this.lastCompactNotice = {
        type: 'error',
        message: `Failed to compact conversation: ${errMsg}`,
        tokensBefore
      };
      return { success: false, before: tokensBefore, after: tokensBefore, error: errMsg };
    } finally {
      this.isCompacting = false;
    }
  }

  // Load chat history from disk via Go backend
  async loadSessionHistoryFromDisk(session: Session): Promise<void> {
    const ws = this.workspaces.find((w) => w.id === session.workspaceId);
    if (!ws || !ws.path) return;

    if (window.go?.main?.App?.LoadGrokSessionHistory) {
      try {
        const targetGrokId = session.grokSessionId || session.id;
        const history = await window.go.main.App.LoadGrokSessionHistory(ws.path, targetGrokId);
        if (history && history.length > 0) {
          session.messages = history.map((m: any, idx: number) => ({
            id: m.id || `${session.id}_msg_${idx}`,
            role: m.role || 'assistant',
            content: m.content || '',
            timestamp: m.timestamp || (session.createdAt + idx * 1000),
            reasoningContent: m.reasoningContent,
            toolCalls: m.toolCalls,
            tokens: m.tokens,
            status: m.status || 'done'
          }));
          session.visibleTurnCount = DEFAULT_WINDOW_TURNS;

          // Reconcile agentMode from session message tool calls if not explicitly set
          if (!session.agentMode) {
            let lastPlanMode: 'agent' | 'plan' | null = null;
            for (const m of session.messages) {
              if (m.toolCalls) {
                for (const tc of m.toolCalls) {
                  const toolLower = (tc.tool || '').toLowerCase();
                  if (toolLower.includes('enter_plan_mode')) {
                    lastPlanMode = 'plan';
                  } else if (toolLower.includes('exit_plan_mode')) {
                    lastPlanMode = 'agent';
                  }
                }
              }
            }
            if (lastPlanMode) {
              session.agentMode = lastPlanMode;
            }
          }
        }
      } catch (err) {
        console.error('Failed to load session history for', session.id, err);
      }
    }
  }

  async switchSession(id: string): Promise<void> {
    if (this.activeSessionId === id) {
      const target = this.sessions.find((s) => s.id === id);
      if (target && target.messages.length === 0 && target.status !== 'working') {
        await this.loadSessionHistoryFromDisk(target);
        await this.loadSessionUsage(target);
      }
      return;
    }
    await this.openSessionInTab(id);
  }

  // Close tab only (preserves session in sidebar and disk, terminates session terminals while preserving background Grok runner)
  closeSessionTab(id: string): void {
    // Terminate all background pseudo-terminals and child processes for this session
    terminalStore.closeAllForSession(id);
    if (window.go?.main?.App?.CloseSessionTerminals) {
      window.go.main.App.CloseSessionTerminals(id)
        .catch(() => {});
    }

    const tabIdx = this.openTabSessionIds.indexOf(id);
    if (tabIdx !== -1) {
      this.openTabSessionIds.splice(tabIdx, 1);
    }

    const currentOpenInWs = this.openWorkspaceTabs;
    if (currentOpenInWs.length === 0) {
      // Allow closing down to 0 tabs without auto-creating a new session
      this.activeSessionId = null;
      this.saveSessionsToStorage();
      return;
    }

    if (this.activeSessionId === id) {
      const nextTab = currentOpenInWs[Math.max(0, tabIdx - 1)] || currentOpenInWs[0];
      this.openSessionInTab(nextTab.id);
    } else {
      this.saveSessionsToStorage();
    }
  }

  closeSession(id: string): void {
    const targetSession = this.sessions.find((s) => s.id === id);
    if (targetSession) {
      // Cancel active runner turn if in progress
      if (targetSession.status === 'working' && window.go?.main?.App?.CancelSession) {
        window.go.main.App.CancelSession(id).catch(() => {});
      }
      // Purge temporary files from disk for this session
      this.purgeTempFilesForSessions([targetSession]);
    }

    this.closeSessionTab(id);
    const index = this.sessions.findIndex((s) => s.id === id);
    if (index === -1) return;

    const wsId = this.sessions[index].workspaceId;
    this.sessions.splice(index, 1);

    const remainingInWs = this.sessions.filter((s) => s.workspaceId === wsId);
    if (remainingInWs.length === 0) {
      this.activeSessionId = null;
      this.saveSessionsToStorage();
      return;
    }

    if (this.activeSessionId === id) {
      this.activeSessionId = remainingInWs[0].id;
      remainingInWs[0].visibleTurnCount = DEFAULT_WINDOW_TURNS;
    }
    this.saveSessionsToStorage();
  }

  togglePinSession(id: string): void {
    const session = this.sessions.find((s) => s.id === id);
    if (session) {
      session.isPinned = !session.isPinned;
      session.pinnedAt = session.isPinned ? Date.now() : undefined;
    }
  }

  // Load external Grok sessions discovered from ~/.grok/sessions
  syncDiscoveredGrokSessions(wsId: string, grokSessions: Array<{ id: string; title: string; createdAt: number; updatedAt: number }>): void {
    if (!grokSessions || grokSessions.length === 0) return;

    for (const gs of grokSessions) {
      // Find existing session in this workspace matching either id or grokSessionId
      const existing = this.sessions.find(
        (s) => s.workspaceId === wsId && (s.id === gs.id || s.grokSessionId === gs.id)
      );

      if (existing) {
        if (!existing.grokSessionId || existing.grokSessionId !== gs.id) {
          existing.grokSessionId = gs.id;
        }
        // Update generic title only if user has not explicitly edited it and existing has not already been derived
        if (gs.title && !existing.isCustomTitle && (existing.title.startsWith('Session ') || existing.title.startsWith('Percakapan ') || existing.title.startsWith('New '))) {
          existing.title = gs.title;
        }
        if (gs.updatedAt && gs.updatedAt > existing.updatedAt) {
          existing.updatedAt = gs.updatedAt;
        }
      } else {
        // Also ensure not already present anywhere in this.sessions under matching ID
        const duplicate = this.sessions.find((s) => s.id === gs.id || s.grokSessionId === gs.id);
        if (!duplicate) {
          this.sessions.push({
            id: gs.id,
            grokSessionId: gs.id,
            workspaceId: wsId,
            title: gs.title || `Session ${this.sessions.filter((s) => s.workspaceId === wsId).length + 1}`,
            status: 'idle',
            createdAt: gs.createdAt || Date.now(),
            updatedAt: gs.updatedAt || Date.now(),
            messages: [],
            visibleTurnCount: DEFAULT_WINDOW_TURNS,
            pendingPermission: null
          });
        }
      }
    }

    this.deduplicateSessions();
    this.reconcileSessionState();
    this.saveSessionsToStorage();
  }

  renameSession(id: string, title: string): void {
    const session = this.sessions.find((s) => s.id === id);
    if (session && title.trim()) {
      session.title = title.trim();
      session.isCustomTitle = true; // Lock manual title
      session.updatedAt = Date.now();
      this.saveSessionsToStorage();
    }
  }

  updateAutoTitle(id: string, newTitle: string, grokSessionId?: string): void {
    const session = this.sessions.find((s) => s.id === id);
    if (!session) return;
    if (session.isCustomTitle) {
      // Still update grokSessionId link if provided, but preserve custom title
      if (grokSessionId) {
        session.grokSessionId = grokSessionId;
        this.saveSessionsToStorage();
      }
      return;
    }
    const clean = newTitle.trim();
    if (clean) {
      session.title = clean;
      if (grokSessionId) {
        session.grokSessionId = grokSessionId;
      }
      session.updatedAt = Date.now();
      this.saveSessionsToStorage();
    }
  }

  forkSession(id: string): Session | null {
    const source = this.sessions.find((s) => s.id === id);
    if (!source) return null;

    const forkedId = this.generateSessionUUID();
    const forked: Session = {
      id: forkedId,
      grokSessionId: forkedId,
      workspaceId: source.workspaceId,
      title: `${source.title} (Fork)`,
      status: 'idle',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: JSON.parse(JSON.stringify(source.messages)),
      visibleTurnCount: DEFAULT_WINDOW_TURNS,
      pendingPermission: null
    };

    const sourceIndex = this.sessions.findIndex((s) => s.id === id);
    this.sessions.splice(sourceIndex + 1, 0, forked);
    this.openTabSessionIds.push(forked.id);
    this.activeSessionId = forked.id;
    this.saveSessionsToStorage();
    return forked;
  }

  setSessionStatus(id: string, status: SessionStatus): void {
    const session = this.sessions.find((s) => s.id === id);
    if (session) {
      session.status = status;
      session.updatedAt = Date.now();
    }
  }

  setPendingPermission(sessionId: string, perm: PermissionRequest | null): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (session) {
      session.pendingPermission = perm;
      if (perm) {
        session.status = 'waiting_permission';
      } else if (session.status === 'waiting_permission') {
        session.status = 'working';
      }
      session.updatedAt = Date.now();
    }
  }

  reorderSessions(fromIndex: number, toIndex: number): void {
    if (
      fromIndex < 0 ||
      fromIndex >= this.openTabSessionIds.length ||
      toIndex < 0 ||
      toIndex >= this.openTabSessionIds.length ||
      fromIndex === toIndex
    ) {
      return;
    }

    const [moved] = this.openTabSessionIds.splice(fromIndex, 1);
    this.openTabSessionIds.splice(toIndex, 0, moved);
    this.saveSessionsToStorage();
  }

  loadEarlierTurns(chunk = PREPEND_CHUNK_TURNS): boolean {
    const session = this.activeSession;
    if (!session) return false;

    const totalTurns = session.messages.filter((m) => m.role === 'user').length;
    if (session.visibleTurnCount >= totalTurns) {
      return false; // Already fully loaded
    }

    session.visibleTurnCount = Math.min(totalTurns, session.visibleTurnCount + chunk);
    return true;
  }

  // Queue and Steer management methods
  addQueuedPrompt(sessionId: string, prompt: QueuedPrompt): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session) return;
    if (!session.queuedPrompts) {
      session.queuedPrompts = [];
    }
    session.queuedPrompts.push(prompt);
  }

  removeQueuedPrompt(sessionId: string, promptId: string): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session || !session.queuedPrompts) return;
    session.queuedPrompts = session.queuedPrompts.filter((p) => p.id !== promptId);
  }

  updateQueuedPrompt(sessionId: string, promptId: string, newText: string): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session || !session.queuedPrompts) return;
    const item = session.queuedPrompts.find((p) => p.id === promptId);
    if (item) {
      item.text = newText;
    }
  }

  reorderQueuedPrompt(sessionId: string, fromIndex: number, toIndex: number): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session || !session.queuedPrompts) return;
    if (fromIndex < 0 || fromIndex >= session.queuedPrompts.length) return;
    if (toIndex < 0 || toIndex >= session.queuedPrompts.length) return;
    const [moved] = session.queuedPrompts.splice(fromIndex, 1);
    session.queuedPrompts.splice(toIndex, 0, moved);
  }

  popNextQueuedPrompt(sessionId: string): QueuedPrompt | undefined {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session || !session.queuedPrompts || session.queuedPrompts.length === 0) return undefined;
    return session.queuedPrompts.shift();
  }

  clearQueuedPrompts(sessionId: string): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session || !session.queuedPrompts) return;
    session.queuedPrompts = [];
  }

  addMessage(sessionId: string, message: Omit<ChatMessage, 'id' | 'timestamp'> & { id?: string; timestamp?: number }): ChatMessage {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session) throw new Error(`Session ${sessionId} not found`);

    const msg: ChatMessage = {
      id: message.id || 'msg_' + Math.random().toString(36).substring(2, 9) + '_' + Date.now().toString(36),
      timestamp: message.timestamp || Date.now(),
      role: message.role,
      content: message.content,
      tokens: message.tokens,
      toolCalls: message.toolCalls ? [...message.toolCalls] : undefined,
      images: message.images ? [...message.images] : undefined,
      status: message.status,
      isSteer: message.isSteer
    };

    session.messages.push(msg);
    // When a new user turn arrives, anchor the active DOM window to the default 4-turn window
    if (msg.role === 'user') {
      session.visibleTurnCount = DEFAULT_WINDOW_TURNS;
    }
    session.updatedAt = Date.now();
    this.saveSessionsToStorage();
    return msg;
  }

  updateToolCall(sessionId: string, toolId: string, updater: (tool: ToolCall) => void): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session) return;
    for (const msg of session.messages) {
      if (msg.toolCalls) {
        const target = msg.toolCalls.find((t) => t.id === toolId);
        if (target) {
          updater(target);
          session.updatedAt = Date.now();
          return;
        }
      }
    }
  }

  appendDelta(
    sessionId: string,
    delta: string,
    role: 'assistant' | 'user' = 'assistant',
    grokSessionId?: string,
    liveTokens?: number
  ): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session) return;

    if (grokSessionId && !session.grokSessionId && !grokSessionId.startsWith('sess_')) {
      session.grokSessionId = grokSessionId;
    }

    if (session.messages.length === 0 || session.messages[session.messages.length - 1].role !== role) {
      this.addMessage(sessionId, {
        role,
        content: delta,
        status: 'streaming'
      });
    } else {
      const lastMsg = session.messages[session.messages.length - 1];
      lastMsg.content += delta;
      lastMsg.status = 'streaming';
      session.updatedAt = Date.now();
    }

    // Refresh live context token consumption during active streaming
    if (liveTokens && liveTokens > 0) {
      if (!session.usage) {
        session.usage = {
          usedTokens: liveTokens,
          maxTokens: 200000,
          lastTurnInput: 0,
          lastTurnOutput: 0,
          lastTurnCacheRead: 0,
          lastTurnReasoning: 0,
          lastTurnModelCalls: 0,
          totalInput: liveTokens,
          totalOutput: 0,
          totalCacheRead: 0,
          turnCount: session.messages.length,
          primaryModelId: ''
        };
      } else {
        session.usage.usedTokens = Math.max(session.usage.usedTokens, liveTokens);
      }
    } else if (delta) {
      // Incremental estimation: ~1 token per 3.8 characters of stream delta
      const deltaTokens = Math.max(1, Math.ceil(delta.length / 3.8));
      if (!session.usage) {
        session.usage = {
          usedTokens: 2500 + deltaTokens,
          maxTokens: 200000,
          lastTurnInput: 2500,
          lastTurnOutput: deltaTokens,
          lastTurnCacheRead: 0,
          lastTurnReasoning: 0,
          lastTurnModelCalls: 0,
          totalInput: 2500,
          totalOutput: deltaTokens,
          totalCacheRead: 0,
          turnCount: session.messages.length,
          primaryModelId: ''
        };
      } else {
        session.usage.usedTokens += deltaTokens;
        session.usage.lastTurnOutput = (session.usage.lastTurnOutput || 0) + deltaTokens;
        session.usage.totalOutput = (session.usage.totalOutput || 0) + deltaTokens;
      }
    }
  }

  updateLastMessage(sessionId: string, updater: (msg: ChatMessage) => void): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session || session.messages.length === 0) return;
    const last = session.messages[session.messages.length - 1];
    updater(last);
    session.updatedAt = Date.now();
  }

  // Rollback last user turn: finds last user message and subsequent assistant messages,
  // removes them from history, and returns the payload to re-load into composer
  rollbackLastUserTurn(sessionId: string): {
    text: string;
    images: VisionImage[];
    attachments: AttachedFile[];
    revertFiles: string[];
  } | null {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session || session.messages.length === 0) return null;

    // Find index of the last message with role === 'user'
    let lastUserIdx = -1;
    for (let i = session.messages.length - 1; i >= 0; i--) {
      if (session.messages[i].role === 'user') {
        lastUserIdx = i;
        break;
      }
    }

    if (lastUserIdx === -1) return null;

    const userMsg = session.messages[lastUserIdx];
    const removedMessages = session.messages.slice(lastUserIdx);

    // Extract any files that were modified by tools in this turn to revert
    const revertFilesSet = new Set<string>();
    for (const msg of removedMessages) {
      if (msg.toolCalls) {
        for (const tc of msg.toolCalls) {
          if (tc.diff?.newPath) {
            revertFilesSet.add(tc.diff.newPath);
          } else if (tc.diff?.oldPath) {
            revertFilesSet.add(tc.diff.oldPath);
          } else if (tc.params) {
            const p = tc.params as Record<string, any>;
            if (p.path || p.file_path || p.filePath || p.target_file) {
              revertFilesSet.add(p.path || p.file_path || p.filePath || p.target_file);
            }
          }
        }
      }
    }

    // Truncate session messages back to before this user turn
    session.messages = session.messages.slice(0, lastUserIdx);
    session.updatedAt = Date.now();

    return {
      text: userMsg.content,
      images: userMsg.images ? [...userMsg.images] : [],
      attachments: userMsg.attachments ? [...userMsg.attachments] : [],
      revertFiles: Array.from(revertFilesSet)
    };
  }
}

export const sessionStore = new SessionStore();
