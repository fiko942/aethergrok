import { describe, it, expect, beforeEach } from 'vitest';

// Svelte 5 Runes mock for node testing environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const { sessionStore } = await import('../frontend/src/lib/stores/session.svelte');

describe('Stream Delta Batching and Background Session Buffering', () => {
  const wsId = 'ws-test-delta-batch';
  const activeSessId = 'sess-active';
  const backgroundSessId = 'sess-background';

  beforeEach(() => {
    if (typeof localStorage !== 'undefined') {
      localStorage.clear();
    }
    sessionStore.workspaces = [
      { id: wsId, name: 'Workspace', path: 'C:\\fake\\path' }
    ];
    sessionStore.sessions = [
      {
        id: activeSessId,
        workspaceId: wsId,
        title: 'Active Session',
        createdAt: Date.now(),
        updatedAt: Date.now(),
        messages: [],
        status: 'idle'
      },
      {
        id: backgroundSessId,
        workspaceId: wsId,
        title: 'Background Session',
        createdAt: Date.now(),
        updatedAt: Date.now(),
        messages: [],
        status: 'idle'
      }
    ];
    sessionStore.activeSessionId = activeSessId;
    sessionStore.activeWorkspaceId = wsId;
  });

  it('updates background session content in-place without creating fresh parts array clones', () => {
    // Initial message added to background session
    sessionStore.appendDelta(backgroundSessId, 'Hello');
    const bgSession = sessionStore.sessions.find((s) => s.id === backgroundSessId)!;
    expect(bgSession.messages.length).toBe(1);

    const msg = bgSession.messages[0];
    expect(msg.content).toBe('Hello');
    expect(msg.parts).toBeDefined();
    expect(msg.parts!.length).toBe(1);

    // Save initial parts array reference
    const initialPartsRef = msg.parts;

    // Stream subsequent tokens to background session
    sessionStore.appendDelta(backgroundSessId, ' world');
    sessionStore.appendDelta(backgroundSessId, ' from');
    sessionStore.appendDelta(backgroundSessId, ' background');

    // Message content and text part must be updated with the full streamed content
    expect(msg.content).toBe('Hello world from background');
    expect(msg.status).toBe('streaming');
    expect(msg.parts![0].type).toBe('text');
    expect(msg.parts![0].text).toBe('Hello world from background');

    // Parts array reference must be preserved (in-place modification, no clone)
    expect(msg.parts).toBe(initialPartsRef);
  });

  it('clones parts array reference for active session to trigger Svelte reactivity', () => {
    // Initial message added to active session
    sessionStore.appendDelta(activeSessId, 'Hello');
    const activeSession = sessionStore.sessions.find((s) => s.id === activeSessId)!;
    expect(activeSession.messages.length).toBe(1);

    const msg = activeSession.messages[0];
    const initialPartsRef = msg.parts;

    // Stream next token to active session
    sessionStore.appendDelta(activeSessId, ' active');

    expect(msg.content).toBe('Hello active');
    // Parts array reference must be cloned for active session UI reactivity
    expect(msg.parts).not.toBe(initialPartsRef);
    expect(msg.parts![0].text).toBe('Hello active');
  });

  it('handles background session delta when last part is not text', () => {
    // Setup message with existing tool call part
    const bgSession = sessionStore.sessions.find((s) => s.id === backgroundSessId)!;
    bgSession.messages.push({
      id: 'msg-bg-1',
      role: 'assistant',
      content: '',
      status: 'streaming',
      parts: [
        {
          type: 'tool_calls',
          toolCalls: [
            { id: 'call-1', name: 'read_file', status: 'completed' } as any
          ]
        }
      ]
    });

    const msg = bgSession.messages[0];
    const partsRef = msg.parts;

    sessionStore.appendDelta(backgroundSessId, 'Tool call completed.');

    expect(msg.content).toBe('Tool call completed.');
    expect(msg.parts!.length).toBe(2);
    expect(msg.parts![1]).toEqual({ type: 'text', text: 'Tool call completed.' });
    // In-place push, parts array reference preserved
    expect(msg.parts).toBe(partsRef);
  });
});
