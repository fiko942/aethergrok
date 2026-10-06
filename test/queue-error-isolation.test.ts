import { describe, it, expect } from 'vitest';

export interface QueueState {
  sessionId: string;
  queuedPrompts: string[];
  activePrompt: string | null;
  status: 'idle' | 'working' | 'error';
  retryCount: number;
}

export function handleTurnError(
  state: QueueState,
  isTruncation: boolean,
  maxRetries = 3
): { action: 'retry' | 'pause'; nextPrompt?: string } {
  state.status = 'error';

  if (isTruncation && state.retryCount < maxRetries) {
    state.retryCount += 1;
    state.status = 'working';
    return {
      action: 'retry',
      nextPrompt: 'Continue and finish the previous truncated task. Focus on the remaining uncompleted steps.'
    };
  }

  // Strictly pause queue: DO NOT dequeue or run next item!
  return { action: 'pause' };
}

describe('Queue Error Isolation & Auto-Retry', () => {
  it('strictly pauses queue and preserves remaining prompts on general error', () => {
    const state: QueueState = {
      sessionId: 's1',
      queuedPrompts: ['Prompt 2', 'Prompt 3'],
      activePrompt: 'Prompt 1',
      status: 'working',
      retryCount: 0
    };

    const res = handleTurnError(state, false);
    expect(res.action).toBe('pause');
    expect(state.queuedPrompts).toHaveLength(2);
    expect(state.queuedPrompts[0]).toBe('Prompt 2');
    expect(state.status).toBe('error');
  });

  it('triggers auto-retry on max_tokens_truncation up to 3x before pausing', () => {
    const state: QueueState = {
      sessionId: 's1',
      queuedPrompts: ['Next task'],
      activePrompt: 'Heavy reading task',
      status: 'working',
      retryCount: 0
    };

    // First truncation -> Retry 1
    const res1 = handleTurnError(state, true, 3);
    expect(res1.action).toBe('retry');
    expect(res1.nextPrompt).toContain('Continue and finish');
    expect(state.retryCount).toBe(1);
    expect(state.queuedPrompts).toHaveLength(1); // Queue untouched!

    // Second truncation -> Retry 2
    const res2 = handleTurnError(state, true, 3);
    expect(res2.action).toBe('retry');
    expect(state.retryCount).toBe(2);
    expect(state.queuedPrompts).toHaveLength(1); // Queue untouched!

    // Third truncation -> Retry 3
    const res3 = handleTurnError(state, true, 3);
    expect(res3.action).toBe('retry');
    expect(state.retryCount).toBe(3);
    expect(state.queuedPrompts).toHaveLength(1); // Queue untouched!

    // Fourth truncation -> Exhausted (3 reached) -> Pause queue
    const res4 = handleTurnError(state, true, 3);
    expect(res4.action).toBe('pause');
    expect(state.status).toBe('error');
    expect(state.queuedPrompts).toHaveLength(1); // Next task safe!
  });
});
