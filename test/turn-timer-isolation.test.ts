import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

describe('Per-Session Turn Execution Timer & Isolation', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('isolates turnStartTime per session and calculates elapsed time correctly', () => {
    let now = 1000000;
    const getNow = () => now;

    // Session A starts prompt at t = 0s
    const sessionA: { id: string; status: string; turnStartTime?: number; lastTurnDurationMs?: number } = {
      id: 'session-a',
      status: 'working',
      turnStartTime: getNow()
    };

    // Advance 15 seconds
    now += 15000;

    // Session B starts prompt at t = 15s
    const sessionB: { id: string; status: string; turnStartTime?: number; lastTurnDurationMs?: number } = {
      id: 'session-b',
      status: 'working',
      turnStartTime: getNow()
    };

    // Advance another 10 seconds (total 25s for A, 10s for B)
    now += 10000;

    const elapsedA = Math.max(0, getNow() - (sessionA.turnStartTime || getNow()));
    const elapsedB = Math.max(0, getNow() - (sessionB.turnStartTime || getNow()));

    expect(elapsedA).toBe(25000);
    expect(elapsedB).toBe(10000);

    // Complete Session A: store lastTurnDurationMs and clear turnStartTime
    sessionA.status = 'finished';
    sessionA.lastTurnDurationMs = Math.max(0, getNow() - (sessionA.turnStartTime || 0));
    sessionA.turnStartTime = undefined;

    expect(sessionA.lastTurnDurationMs).toBe(25000);
    expect(sessionA.turnStartTime).toBeUndefined();

    // Session B is still working
    now += 5000;
    const elapsedBAfter = Math.max(0, getNow() - (sessionB.turnStartTime || getNow()));
    expect(elapsedBAfter).toBe(15000);
  });

  it('formats elapsed milliseconds matching Grok CLI precision rules', () => {
    function formatElapsed(ms: number): string {
      if (ms < 1000) {
        return `${ms}ms`;
      }
      const sec = Math.floor(ms / 1000);
      const m = Math.floor(sec / 60);
      const s = sec % 60;
      if (m === 0) {
        return `${(ms / 1000).toFixed(1)}s`;
      }
      return `${m}m ${s < 10 ? '0' : ''}${s}s`;
    }

    expect(formatElapsed(450)).toBe('450ms');
    expect(formatElapsed(4200)).toBe('4.2s');
    expect(formatElapsed(25000)).toBe('25.0s');
    expect(formatElapsed(85400)).toBe('1m 25s');
    expect(formatElapsed(124000)).toBe('2m 04s');
  });

  it('derives contextual status text between tool running and waiting for response', () => {
    function getStatusText(lastMsg?: { status?: string; toolCalls?: Array<{ status: string; tool?: string }> }) {
      if (!lastMsg) return 'Waiting for response...';
      if (lastMsg.toolCalls && lastMsg.toolCalls.length > 0) {
        const activeTool = lastMsg.toolCalls.find(tc => tc.status === 'running');
        if (activeTool) {
          const name = (activeTool.tool || '').toLowerCase();
          if (name.includes('read')) return 'Reading workspace file...';
          if (name.includes('terminal') || name.includes('bash')) return 'Executing shell command...';
          return `Running tool: ${activeTool.tool}...`;
        }
        if (lastMsg.status === 'streaming') {
          return 'Grok is formulating response...';
        }
        return 'Waiting for response...';
      }
      return 'Waiting for response...';
    }

    expect(getStatusText()).toBe('Waiting for response...');

    // Tool running
    expect(getStatusText({
      toolCalls: [{ status: 'running', tool: 'run_terminal_command' }]
    })).toBe('Executing shell command...');

    // Tool completed, assistant generating response
    expect(getStatusText({
      status: 'streaming',
      toolCalls: [{ status: 'completed', tool: 'run_terminal_command' }]
    })).toBe('Grok is formulating response...');

    // Tool completed, awaiting model output
    expect(getStatusText({
      status: 'pending',
      toolCalls: [{ status: 'completed', tool: 'run_terminal_command' }]
    })).toBe('Waiting for response...');
  });
});
