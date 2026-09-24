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
  status?: 'streaming' | 'done' | 'error';
  isSteer?: boolean;
}

export interface Session {
  id: string;
  workspaceId: string;
  title: string;
  status: SessionStatus;
  createdAt: number;
  updatedAt: number;
  messages: ChatMessage[];
  // Windowing state per session
  visibleTurnCount: number; // Number of recent user turns currently in DOM window (default 10)
  pendingPermission?: PermissionRequest | null;
}

export interface WorkspaceFolder {
  id: string;
  name: string; // e.g. "affilia", "grok-desktop"
  path: string; // e.g. "/Users/fiko942/Desktop/affilia"
  createdAt: number;
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

export const DEFAULT_WINDOW_TURNS = 10;
export const PREPEND_CHUNK_TURNS = 10;

class SessionStore {
  workspaces = $state<WorkspaceFolder[]>([]);
  activeWorkspaceId = $state<string>('');
  sessions = $state<Session[]>([]);
  activeSessionId = $state<string | null>(null);

  // Multi-select state
  isSelectionMode = $state<boolean>(false);
  selectedSessionIds = $state<Set<string>>(new Set());

  constructor() {
    // Initialize default workspace
    const defaultWs: WorkspaceFolder = {
      id: 'ws_affilia_root',
      name: 'affilia',
      path: '/Users/fiko942/Desktop/affilia',
      createdAt: Date.now()
    };
    this.workspaces = [defaultWs];
    this.activeWorkspaceId = defaultWs.id;

    // Initialize with a default session linked to default workspace
    const initialSession = this.createNewSessionModel('Log Audit & Automation', defaultWs.id);
    this.sessions = [initialSession];
    this.activeSessionId = initialSession.id;
  }

  private createNewSessionModel(title?: string, wsId?: string): Session {
    const id = 'sess_' + Math.random().toString(36).substring(2, 9) + '_' + Date.now().toString(36);
    const workspaceId = wsId || this.activeWorkspaceId || (this.workspaces[0]?.id ?? 'ws_default');
    return {
      id,
      workspaceId,
      title: title || `Session ${this.sessions.filter((s) => s.workspaceId === workspaceId).length + 1}`,
      status: 'idle',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: [],
      visibleTurnCount: DEFAULT_WINDOW_TURNS,
      pendingPermission: null
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
      this.activeWorkspaceId = existing.id;
      return existing;
    }

    const id = 'ws_' + Math.random().toString(36).substring(2, 9);
    const newWs: WorkspaceFolder = {
      id,
      name,
      path,
      createdAt: Date.now()
    };
    this.workspaces.push(newWs);
    this.activeWorkspaceId = id;

    // Create a starter session for the newly added workspace
    this.createSession(`New ${name} Task`, id);
    return newWs;
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

    this.workspaces.splice(index, 1);
    // Remove associated sessions
    this.sessions = this.sessions.filter((s) => s.workspaceId !== id);

    if (this.workspaces.length === 0) {
      const fallback = this.addWorkspace('default', '/');
      return;
    }

    if (this.activeWorkspaceId === id) {
      this.activeWorkspaceId = this.workspaces[0].id;
      const wsSessions = this.sessions.filter((s) => s.workspaceId === this.activeWorkspaceId);
      this.activeSessionId = wsSessions[0]?.id || null;
    }
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

  selectAllSessions(): void {
    const currentSessions = this.activeWorkspaceSessions;
    const allIds = currentSessions.map((s) => s.id);
    this.selectedSessionIds = new Set(allIds);
  }

  deselectAllSessions(): void {
    this.selectedSessionIds = new Set();
  }

  deleteSelectedSessions(): void {
    if (this.selectedSessionIds.size === 0) return;
    const idsToDelete = new Set(this.selectedSessionIds);
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
  }

  // Session Actions
  createSession(title?: string, wsId?: string): Session {
    const newSession = this.createNewSessionModel(title, wsId);
    this.sessions.push(newSession);
    this.activeSessionId = newSession.id;
    return newSession;
  }

  switchSession(id: string): void {
    if (this.activeSessionId === id) return;
    const target = this.sessions.find((s) => s.id === id);
    if (target) {
      target.visibleTurnCount = DEFAULT_WINDOW_TURNS;
      this.activeSessionId = id;
      if (target.workspaceId && target.workspaceId !== this.activeWorkspaceId) {
        this.activeWorkspaceId = target.workspaceId;
      }
    }
  }

  closeSession(id: string): void {
    const index = this.sessions.findIndex((s) => s.id === id);
    if (index === -1) return;

    const wsId = this.sessions[index].workspaceId;
    this.sessions.splice(index, 1);

    const remainingInWs = this.sessions.filter((s) => s.workspaceId === wsId);
    if (remainingInWs.length === 0) {
      const fallback = this.createNewSessionModel('New Task', wsId);
      this.sessions.push(fallback);
      this.activeSessionId = fallback.id;
      return;
    }

    if (this.activeSessionId === id) {
      this.activeSessionId = remainingInWs[0].id;
      remainingInWs[0].visibleTurnCount = DEFAULT_WINDOW_TURNS;
    }
  }

  renameSession(id: string, title: string): void {
    const session = this.sessions.find((s) => s.id === id);
    if (session && title.trim()) {
      session.title = title.trim();
      session.updatedAt = Date.now();
    }
  }

  forkSession(id: string): Session | null {
    const source = this.sessions.find((s) => s.id === id);
    if (!source) return null;

    const forked: Session = {
      id: 'sess_' + Math.random().toString(36).substring(2, 9) + '_' + Date.now().toString(36),
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
    this.activeSessionId = forked.id;
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
      fromIndex >= this.sessions.length ||
      toIndex < 0 ||
      toIndex >= this.sessions.length ||
      fromIndex === toIndex
    ) {
      return;
    }

    const [moved] = this.sessions.splice(fromIndex, 1);
    this.sessions.splice(toIndex, 0, moved);
  }

  loadEarlierTurns(chunk = PREPEND_CHUNK_TURNS): boolean {
    const session = this.activeSession;
    if (!session) return false;

    const total = this.totalUserTurns;
    if (session.visibleTurnCount >= total) {
      return false;
    }

    session.visibleTurnCount = Math.min(total, session.visibleTurnCount + chunk);
    return true;
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
    session.updatedAt = Date.now();
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

  appendDelta(sessionId: string, delta: string, role: 'assistant' | 'user' = 'assistant'): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session) return;

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
  }

  updateLastMessage(sessionId: string, updater: (msg: ChatMessage) => void): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (!session || session.messages.length === 0) return;
    const last = session.messages[session.messages.length - 1];
    updater(last);
    session.updatedAt = Date.now();
  }
}

export const sessionStore = new SessionStore();
