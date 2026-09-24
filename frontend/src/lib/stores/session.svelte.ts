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
  title: string;
  status: SessionStatus;
  createdAt: number;
  updatedAt: number;
  messages: ChatMessage[];
  // Windowing state per session
  visibleTurnCount: number; // Number of recent user turns currently in DOM window (default 10)
  pendingPermission?: PermissionRequest | null;
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
  sessions = $state<Session[]>([]);
  activeSessionId = $state<string | null>(null);

  constructor() {
    // Initialize with a default session
    const initialSession = this.createNewSessionModel('Grok Session 1');
    this.sessions = [initialSession];
    this.activeSessionId = initialSession.id;
  }

  private createNewSessionModel(title?: string): Session {
    const id = 'sess_' + Math.random().toString(36).substring(2, 9) + '_' + Date.now().toString(36);
    return {
      id,
      title: title || `Session ${this.sessions.length + 1}`,
      status: 'idle',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: [],
      visibleTurnCount: DEFAULT_WINDOW_TURNS,
      pendingPermission: null
    };
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

  // Actions
  createSession(title?: string): Session {
    const newSession = this.createNewSessionModel(title);
    this.sessions.push(newSession);
    this.activeSessionId = newSession.id;
    return newSession;
  }

  switchSession(id: string): void {
    if (this.activeSessionId === id) return;
    const target = this.sessions.find((s) => s.id === id);
    if (target) {
      // Clean switch: reset visible turns to default 10-turn window on switch
      target.visibleTurnCount = DEFAULT_WINDOW_TURNS;
      this.activeSessionId = id;
    }
  }

  closeSession(id: string): void {
    const index = this.sessions.findIndex((s) => s.id === id);
    if (index === -1) return;

    this.sessions.splice(index, 1);

    if (this.sessions.length === 0) {
      const fallback = this.createNewSessionModel('New Session');
      this.sessions = [fallback];
      this.activeSessionId = fallback.id;
      return;
    }

    if (this.activeSessionId === id) {
      // Switch to nearest neighbor
      const nextIndex = Math.min(index, this.sessions.length - 1);
      this.activeSessionId = this.sessions[nextIndex].id;
      this.sessions[nextIndex].visibleTurnCount = DEFAULT_WINDOW_TURNS;
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
      title: `${source.title} (Fork)`,
      status: 'idle',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      // Deep clone messages
      messages: JSON.parse(JSON.stringify(source.messages)),
      visibleTurnCount: DEFAULT_WINDOW_TURNS,
      pendingPermission: null
    };

    // Insert next to the source session
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
      return false; // All turns already visible
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
    const lastMsg = session.messages[session.messages.length - 1];
    updater(lastMsg);
    session.updatedAt = Date.now();
  }

  clearMessages(sessionId: string): void {
    const session = this.sessions.find((s) => s.id === sessionId);
    if (session) {
      session.messages = [];
      session.visibleTurnCount = DEFAULT_WINDOW_TURNS;
      session.pendingPermission = null;
      session.updatedAt = Date.now();
    }
  }
}

export const sessionStore = new SessionStore();
