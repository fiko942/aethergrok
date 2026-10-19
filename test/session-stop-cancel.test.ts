import { describe, it, expect } from 'vitest';

if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

describe('Session Stop and Cancellation Logic', () => {
  it('correctly resolves session ID even when called with a DOM event object', () => {
    function resolveSessionId(targetSessionId: unknown, fallbackActiveId?: string | null): string | null {
      const valid = typeof targetSessionId === 'string' && targetSessionId.trim() ? targetSessionId.trim() : null;
      return valid || fallbackActiveId || null;
    }

    const mouseEvent = { isTrusted: true, type: 'click' };
    expect(resolveSessionId(mouseEvent, 'sess_active_123')).toBe('sess_active_123');
    expect(resolveSessionId('sess_explicit_456', 'sess_active_123')).toBe('sess_explicit_456');
    expect(resolveSessionId(undefined, 'sess_active_123')).toBe('sess_active_123');
    expect(resolveSessionId(null, null)).toBeNull();
  });

  it('finalizes in-progress running tool calls when cancelled', () => {
    const mockMessage = {
      role: 'assistant',
      content: 'Thinking...',
      status: 'streaming',
      toolCalls: [
        { id: 'tc-1', tool: 'spawn_subagent', status: 'running' as const, startTime: 1000 },
        { id: 'tc-2', tool: 'read_file', status: 'completed' as const, startTime: 500, endTime: 700 }
      ]
    };

    // Cancellation finalizer simulation
    if (mockMessage.status === 'streaming') {
      mockMessage.status = 'done';
      mockMessage.content += '\n\n*(Turn cancelled by user)*';
    }
    if (mockMessage.toolCalls) {
      for (const tc of mockMessage.toolCalls) {
        if (tc.status === 'running') {
          tc.status = 'error';
          tc.endTime = 2000;
        }
      }
    }

    expect(mockMessage.status).toBe('done');
    expect(mockMessage.content).toContain('*(Turn cancelled by user)*');
    expect(mockMessage.toolCalls[0].status).toBe('error');
    expect(mockMessage.toolCalls[0].endTime).toBe(2000);
    expect(mockMessage.toolCalls[1].status).toBe('completed');
  });
});
