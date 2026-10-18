import { describe, it, expect } from 'vitest';

describe('Session Active Running Guard', () => {
  interface MockSession {
    id: string;
    title: string;
    status: 'idle' | 'working' | 'waiting_permission' | 'error';
    queuedPrompts?: { id: string }[];
  }

  function hasActiveRunningSessions(sessions: MockSession[]): boolean {
    return sessions.some(
      (s) => s.status === 'working' || s.status === 'waiting_permission' || (s.queuedPrompts && s.queuedPrompts.length > 0)
    );
  }

  function getActiveRunningSessionCount(sessions: MockSession[]): number {
    return sessions.filter(
      (s) => s.status === 'working' || s.status === 'waiting_permission' || (s.queuedPrompts && s.queuedPrompts.length > 0)
    ).length;
  }

  function getActiveRunningSessionTitles(sessions: MockSession[]): string[] {
    return sessions
      .filter((s) => s.status === 'working' || s.status === 'waiting_permission' || (s.queuedPrompts && s.queuedPrompts.length > 0))
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
});
