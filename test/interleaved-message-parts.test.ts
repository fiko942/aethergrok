import { describe, it, expect, beforeEach } from 'vitest';

// Svelte 5 Runes mock for node testing environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const { SessionStore } = await import('../frontend/src/lib/stores/session.svelte');
import type { ToolCall } from '../frontend/src/lib/stores/session.svelte';

describe('Chronological Interleaved Message Parts', () => {
  let store: SessionStore;
  const sessionId = 'test-session-interleave';

  beforeEach(() => {
    store = new SessionStore();
    store.sessions = [
      {
        id: sessionId,
        workspaceId: 'ws-1',
        title: 'Interleaved Test',
        createdAt: 1000,
        updatedAt: 1000,
        status: 'idle',
        messages: []
      }
    ];
    store.activeSessionId = sessionId;
  });

  it('places initial text first and subsequent tool call below it', () => {
    // 1. User prompt
    store.addMessage(sessionId, {
      role: 'user',
      content: 'Please inspect the codebase'
    });

    // 2. Assistant streams initial text
    store.appendDelta(sessionId, 'I will inspect the workspace files.', 'assistant');

    const session = store.sessions.find((s) => s.id === sessionId)!;
    expect(session.messages.length).toBe(2);
    const asstMsg = session.messages[1];
    expect(asstMsg.parts).toBeDefined();
    expect(asstMsg.parts?.length).toBe(1);
    expect(asstMsg.parts?.[0]).toEqual({
      type: 'text',
      text: 'I will inspect the workspace files.'
    });

    // 3. Assistant invokes a tool call
    const toolCall: ToolCall = {
      id: 'tc-1',
      tool: 'read_file',
      params: { target_file: 'src/main.ts' },
      status: 'running',
      startTime: Date.now()
    };
    store.addToolCall(sessionId, toolCall);

    expect(session.messages.length).toBe(2);
    expect(asstMsg.parts?.length).toBe(2);
    // Crucial: Part 0 is introductory text, Part 1 is the tool call (below the text)
    expect(asstMsg.parts?.[0].type).toBe('text');
    expect((asstMsg.parts?.[0] as any).text).toBe('I will inspect the workspace files.');
    expect(asstMsg.parts?.[1].type).toBe('tool_calls');
    expect((asstMsg.parts?.[1] as any).toolCalls[0].id).toBe('tc-1');
  });

  it('interleaves text, tool call, subsequent text, and second tool call in exact chronological order', () => {
    // 1. Initial text
    store.appendDelta(sessionId, 'Starting inspection...', 'assistant');

    // 2. First tool call
    store.addToolCall(sessionId, {
      id: 'tool-read',
      tool: 'read_file',
      params: { file: 'app.ts' },
      status: 'completed',
      startTime: 1010
    });

    // 3. Subsequent text after first tool
    store.appendDelta(sessionId, 'Found the bug in app.ts. Now fixing.', 'assistant');

    // 4. Second tool call
    store.addToolCall(sessionId, {
      id: 'tool-write',
      tool: 'write_file',
      params: { file: 'app.ts' },
      status: 'completed',
      startTime: 1020
    });

    // 5. Final conclusion text
    store.appendDelta(sessionId, 'Verification complete!', 'assistant');

    const session = store.sessions.find((s) => s.id === sessionId)!;
    const msg = session.messages[0];

    expect(msg.parts?.length).toBe(5);
    expect(msg.parts?.[0]).toEqual({
      type: 'text',
      text: 'Starting inspection...'
    });
    expect(msg.parts?.[1].type).toBe('tool_calls');
    expect((msg.parts?.[1] as any).toolCalls[0].id).toBe('tool-read');
    expect(msg.parts?.[2]).toEqual({
      type: 'text',
      text: 'Found the bug in app.ts. Now fixing.'
    });
    expect(msg.parts?.[3].type).toBe('tool_calls');
    expect((msg.parts?.[3] as any).toolCalls[0].id).toBe('tool-write');
    expect(msg.parts?.[4]).toEqual({
      type: 'text',
      text: 'Verification complete!'
    });
  });

  it('defaults fallback reconstruction to text first and activity list below text', () => {
    const msg = store.addMessage(sessionId, {
      role: 'assistant',
      content: 'Here is what I found during analysis.',
      toolCalls: [
        {
          id: 'tc-fallback',
          tool: 'list_dir',
          params: {},
          status: 'completed'
        }
      ]
    });

    // When no explicit parts passed, addMessage derives parts with text first and tool calls below
    expect(msg.parts).toBeDefined();
    expect(msg.parts?.length).toBe(2);
    expect(msg.parts?.[0].type).toBe('text');
    expect((msg.parts?.[0] as any).text).toBe('Here is what I found during analysis.');
    expect(msg.parts?.[1].type).toBe('tool_calls');
    expect((msg.parts?.[1] as any).toolCalls[0].id).toBe('tc-fallback');
  });

  it('correctly handles tool calls starting before text when tool call runs first', () => {
    const toolCall: ToolCall = {
      id: 'tc-first',
      tool: 'read_file',
      params: {},
      status: 'completed',
      startTime: 1000
    };

    // Tool call added first to an empty turn
    store.addToolCall(sessionId, toolCall);

    // Later delta text arrives
    store.appendDelta(sessionId, 'Result from reading file', 'assistant');

    const session = store.sessions.find((s) => s.id === sessionId)!;
    const msg = session.messages[0];
    expect(msg.parts?.length).toBe(2);
    expect(msg.parts?.[0].type).toBe('tool_calls');
    expect((msg.parts?.[0] as any).toolCalls[0].id).toBe('tc-first');
    expect(msg.parts?.[1].type).toBe('text');
    expect((msg.parts?.[1] as any).text).toBe('Result from reading file');
  });
});
