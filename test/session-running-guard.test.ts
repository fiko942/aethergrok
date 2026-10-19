import { describe, it, expect } from 'vitest';

describe('Session Active Running Guard', () => {
  interface MockToolCall {
    id: string;
    tool: string;
    params?: any;
    result?: string;
    status: 'running' | 'completed' | 'error';
  }

  interface MockMessage {
    role: string;
    toolCalls?: MockToolCall[];
  }

  interface MockSession {
    id: string;
    title: string;
    status: 'idle' | 'working' | 'waiting_permission' | 'error';
    queuedPrompts?: { id: string }[];
    messages?: MockMessage[];
  }

  function isSubagentTool(tool: string): boolean {
    return tool === 'spawn_subagent' || tool === 'subagent' || tool === 'dispatch_agent';
  }

  function isSessionActive(s: MockSession): boolean {
    if (s.status === 'working' || s.status === 'waiting_permission' || (s.queuedPrompts && s.queuedPrompts.length > 0)) {
      return true;
    }
    if (s.messages && s.messages.length > 0) {
      for (const msg of s.messages) {
        if (msg.toolCalls && msg.toolCalls.some((tc) => tc.status === 'running' || (isSubagentTool(tc.tool) && (!tc.result || tc.result.includes('started in the background'))))) {
          return true;
        }
      }
    }
    return false;
  }

  function hasActiveRunningSessions(sessions: MockSession[]): boolean {
    return sessions.some((s) => isSessionActive(s));
  }

  function getActiveRunningSessionCount(sessions: MockSession[]): number {
    return sessions.filter((s) => isSessionActive(s)).length;
  }

  function getActiveRunningSessionTitles(sessions: MockSession[]): string[] {
    return sessions
      .filter((s) => isSessionActive(s))
      .map((s) => s.title || 'Untitled Session');
  }

  it('detects no running sessions when all are idle', () => {
    const sessions: MockSession[] = [
      { id: '1', title: 'Task 1', status: 'idle' },
      { id: '2', title: 'Task 2', status: 'idle' }
    ];
    expect(hasActiveRunningSessions(sessions)).toBe(false);
    expect(getActiveRunningSessionCount(sessions)).toBe(0);
    expect(getActiveRunningSessionTitles(sessions)).toEqual([]);
  });

  it('detects running sessions even when session is in an inactive/unopened background tab', () => {
    const sessions: MockSession[] = [
      { id: 'active-tab', title: 'Idle Tab', status: 'idle' },
      { id: 'background-tab-1', title: 'Background Running', status: 'working' },
      { id: 'background-tab-2', title: 'Background Waiting Perm', status: 'waiting_permission' },
      { id: 'background-tab-3', title: 'Background Queued', status: 'idle', queuedPrompts: [{ id: 'q1' }] }
    ];
    expect(hasActiveRunningSessions(sessions)).toBe(true);
    expect(getActiveRunningSessionCount(sessions)).toBe(3);
    expect(getActiveRunningSessionTitles(sessions)).toEqual([
      'Background Running',
      'Background Waiting Perm',
      'Background Queued'
    ]);
  });

  it('detects active sessions that have a running tool call or pending background subagent', () => {
    const sessions: MockSession[] = [
      {
        id: 'sess-tool',
        title: 'Running Shell Command',
        status: 'idle',
        messages: [
          {
            role: 'assistant',
            toolCalls: [
              { id: 'tc1', tool: 'run_terminal_command', status: 'running' }
            ]
          }
        ]
      },
      {
        id: 'sess-subagent',
        title: 'Background Subagent Polling',
        status: 'idle',
        messages: [
          {
            role: 'assistant',
            toolCalls: [
              { id: 'tc2', tool: 'spawn_subagent', status: 'completed', result: 'Subagent started in the background with task ID: task_abc' }
            ]
          }
        ]
      }
    ];

    expect(hasActiveRunningSessions(sessions)).toBe(true);
    expect(getActiveRunningSessionCount(sessions)).toBe(2);
    expect(getActiveRunningSessionTitles(sessions)).toEqual([
      'Running Shell Command',
      'Background Subagent Polling'
    ]);
  });
});
