export type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';
export type LogCategory = 'UI' | 'SESSION' | 'TERMINAL' | 'VOICE' | 'BACKEND' | 'SETTINGS' | 'SYSTEM' | 'UPDATER';

export interface LogEntry {
  id: string;
  timestamp: number;
  level: LogLevel;
  category: LogCategory;
  message: string;
  details?: any;
}

const MAX_LOG_ENTRIES = 1000;
const STORAGE_LOG_RETENTION_KEY = 'aethergrok_log_retention_v1';

class LoggerStore {
  entries = $state<LogEntry[]>([]);
  maxEntries = $state<number>(MAX_LOG_ENTRIES);
  autoScroll = $state<boolean>(true);
  isLoadedFromDisk = $state<boolean>(false);

  constructor() {
    this.setupGlobalErrorHandlers();
    this.initLogsFromBackend();
  }

  private async initLogsFromBackend(): Promise<void> {
    if (typeof window !== 'undefined' && window.go?.main?.App?.LoadPersistedLogs) {
      try {
        const persisted = await window.go.main.App.LoadPersistedLogs(this.maxEntries);
        if (persisted && Array.isArray(persisted) && persisted.length > 0) {
          this.entries = persisted.map((p) => ({
            id: p.id || 'log_' + Math.random().toString(36).substring(2, 8),
            timestamp: p.timestamp || Date.now(),
            level: (p.level as LogLevel) || 'INFO',
            category: (p.category as LogCategory) || 'SYSTEM',
            message: p.message,
            details: p.details
          }));
        }
      } catch (err) {
        console.warn('Failed to load persisted logs from disk:', err);
      }
    }

    this.isLoadedFromDisk = true;
    this.info('SYSTEM', 'Application logging engine initialized (persistent file backing active)', {
      maxEntries: this.maxEntries,
      userAgent: typeof navigator !== 'undefined' ? navigator.userAgent : 'Unknown'
    });
  }

  private setupGlobalErrorHandlers(): void {
    if (typeof window === 'undefined') return;

    window.addEventListener('error', (event) => {
      this.error('SYSTEM', `Uncaught window error: ${event.message}`, {
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
        error: event.error ? (event.error.stack || event.error.toString()) : undefined
      });
    });

    window.addEventListener('unhandledrejection', (event) => {
      let reasonText = 'Unknown rejection';
      let details: any = null;
      if (event.reason instanceof Error) {
        reasonText = event.reason.message;
        details = { stack: event.reason.stack };
      } else if (typeof event.reason === 'string') {
        reasonText = event.reason;
      } else {
        try {
          details = JSON.stringify(event.reason);
        } catch {
          details = String(event.reason);
        }
      }
      this.error('SYSTEM', `Unhandled promise rejection: ${reasonText}`, details);
    });
  }

  private addEntry(level: LogLevel, category: LogCategory, message: string, details?: any): void {
    const entry: LogEntry = {
      id: 'log_' + Date.now().toString(36) + '_' + Math.random().toString(36).substring(2, 7),
      timestamp: Date.now(),
      level,
      category,
      message,
      details
    };

    // Circular ring buffer: keep under maxEntries
    if (this.entries.length >= this.maxEntries) {
      this.entries.shift();
    }
    this.entries.push(entry);

    // Persist to disk via Wails backend (non-blocking)
    if (typeof window !== 'undefined' && window.go?.main?.App?.AppendSystemLog) {
      window.go.main.App.AppendSystemLog({
        id: entry.id,
        timestamp: entry.timestamp,
        level: entry.level,
        category: entry.category,
        message: entry.message,
        details: entry.details
      }).catch((e: any) => {
        // Ignore background disk append logging errors
      });
    }

    // Standard devtools console mirror (non-blocking)
    const formatted = `[${category}] ${message}`;
    if (level === 'ERROR') {
      console.error(formatted, details ?? '');
    } else if (level === 'WARN') {
      console.warn(formatted, details ?? '');
    } else if (level === 'INFO') {
      console.info(formatted, details ?? '');
    } else {
      console.debug(formatted, details ?? '');
    }
  }

  debug(category: LogCategory, message: string, details?: any): void {
    this.addEntry('DEBUG', category, message, details);
  }

  info(category: LogCategory, message: string, details?: any): void {
    this.addEntry('INFO', category, message, details);
  }

  warn(category: LogCategory, message: string, details?: any): void {
    this.addEntry('WARN', category, message, details);
  }

  error(category: LogCategory, message: string, details?: any): void {
    this.addEntry('ERROR', category, message, details);
  }

  clear(): void {
    this.entries = [];
    if (typeof window !== 'undefined' && window.go?.main?.App?.ClearPersistedLogs) {
      window.go.main.App.ClearPersistedLogs().catch((e: any) => {
        console.warn('Failed to clear persisted logs on disk:', e);
      });
    }
    this.info('SYSTEM', 'Log buffer and disk files cleared by user');
  }

  async exportAsJSON(): Promise<string | null> {
    const filename = `aethergrok-logs-${new Date().toISOString().replace(/[:.]/g, '-')}.json`;
    const content = JSON.stringify(this.entries, null, 2);

    if (typeof window !== 'undefined' && window.go?.main?.App?.SaveLogExport) {
      try {
        const savedPath = await window.go.main.App.SaveLogExport(filename, content, 'json');
        if (savedPath) {
          this.info('SYSTEM', 'Exported logs as JSON to disk', { path: savedPath, count: this.entries.length });
          return savedPath;
        }
        return null; // Cancelled
      } catch (err) {
        this.error('SYSTEM', 'Failed to save JSON log export via dialog', err);
      }
    }

    // Browser download fallback
    const dataStr = 'data:text/json;charset=utf-8,' + encodeURIComponent(content);
    const downloadAnchor = document.createElement('a');
    downloadAnchor.setAttribute('href', dataStr);
    downloadAnchor.setAttribute('download', filename);
    document.body.appendChild(downloadAnchor);
    downloadAnchor.click();
    downloadAnchor.remove();
    this.info('SYSTEM', 'Exported logs as JSON (browser fallback)', { count: this.entries.length });
    return filename;
  }

  async exportAsText(): Promise<string | null> {
    const lines = this.entries.map((e) => {
      const timeStr = new Date(e.timestamp).toISOString();
      const detailsStr = e.details ? ` | details: ${JSON.stringify(e.details)}` : '';
      return `[${timeStr}] [${e.level.padEnd(5)}] [${e.category.padEnd(8)}] ${e.message}${detailsStr}`;
    });
    const filename = `aethergrok-logs-${new Date().toISOString().replace(/[:.]/g, '-')}.log`;
    const content = lines.join('\n');

    if (typeof window !== 'undefined' && window.go?.main?.App?.SaveLogExport) {
      try {
        const savedPath = await window.go.main.App.SaveLogExport(filename, content, 'log');
        if (savedPath) {
          this.info('SYSTEM', 'Exported logs as Text .log to disk', { path: savedPath, count: this.entries.length });
          return savedPath;
        }
        return null; // Cancelled
      } catch (err) {
        this.error('SYSTEM', 'Failed to save Text log export via dialog', err);
      }
    }

    // Browser download fallback
    const dataStr = 'data:text/plain;charset=utf-8,' + encodeURIComponent(content);
    const downloadAnchor = document.createElement('a');
    downloadAnchor.setAttribute('href', dataStr);
    downloadAnchor.setAttribute('download', filename);
    document.body.appendChild(downloadAnchor);
    downloadAnchor.click();
    downloadAnchor.remove();
    this.info('SYSTEM', 'Exported logs as Text .log (browser fallback)', { count: this.entries.length });
    return filename;
  }

  async copyToClipboard(): Promise<boolean> {
    const lines = this.entries.map((e) => {
      const timeStr = new Date(e.timestamp).toISOString();
      const detailsStr = e.details ? ` | details: ${JSON.stringify(e.details)}` : '';
      return `[${timeStr}] [${e.level.padEnd(5)}] [${e.category.padEnd(8)}] ${e.message}${detailsStr}`;
    });

    try {
      await navigator.clipboard.writeText(lines.join('\n'));
      this.info('SYSTEM', 'Copied logs to clipboard', { count: this.entries.length });
      return true;
    } catch (err) {
      this.error('SYSTEM', 'Failed to copy logs to clipboard', err);
      return false;
    }
  }
}

export const logger = new LoggerStore();
