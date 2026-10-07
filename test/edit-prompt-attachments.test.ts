import { describe, it, expect, beforeEach } from 'vitest';

// Svelte 5 Runes mock for node testing environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

const { sessionStore } = await import('../frontend/src/lib/stores/session.svelte');

describe('Edit and Rollback Prompt Attachments Integration', () => {
  const wsId = 'ws-test-edit-turn';
  const sessId = 'sess-test-edit-turn';

  beforeEach(() => {
    if (typeof localStorage !== 'undefined') {
      localStorage.clear();
    }
    sessionStore.workspaces = [
      { id: wsId, name: 'Workspace', path: 'C:\\fake\\path' }
    ];
    sessionStore.sessions = [
      {
        id: sessId,
        workspaceId: wsId,
        title: 'New Session',
        createdAt: Date.now(),
        updatedAt: Date.now(),
        messages: [],
        status: 'idle'
      }
    ];
    sessionStore.activeSessionId = sessId;
    sessionStore.activeWorkspaceId = wsId;
  });

  it('preserves attachments and images in addMessage and returns them in rollbackLastUserTurn', () => {
    const mockImage = {
      id: 'img-1',
      filePath: 'C:\\test\\snapshot.png',
      dataUrl: 'data:image/png;base64,fakeimage',
      sizeBytes: 1024,
      timestamp: Date.now()
    };

    const mockAttachment = {
      id: 'file-1',
      name: 'notes.md',
      filePath: 'C:\\test\\notes.md',
      content: '# My Notes',
      isImage: false,
      timestamp: Date.now()
    };

    // 1. Add user turn with both images and attachments
    const userMsg = sessionStore.addMessage(sessId, {
      role: 'user',
      content: 'Analyze this image and note file',
      images: [mockImage],
      attachments: [mockAttachment]
    });

    expect(userMsg.images).toHaveLength(1);
    expect(userMsg.attachments).toHaveLength(1);
    expect(userMsg.attachments?.[0].name).toBe('notes.md');

    // 2. Add assistant response
    sessionStore.addMessage(sessId, {
      role: 'assistant',
      content: 'Here is the analysis.',
      status: 'done'
    });

    // 3. Perform rollback
    const rollback = sessionStore.rollbackLastUserTurn(sessId);
    expect(rollback).not.toBeNull();
    expect(rollback?.text).toBe('Analyze this image and note file');
    expect(rollback?.images).toHaveLength(1);
    expect(rollback?.images[0].filePath).toBe('C:\\test\\snapshot.png');
    expect(rollback?.attachments).toHaveLength(1);
    expect(rollback?.attachments[0].name).toBe('notes.md');

    // Verify session history has been truncated
    const session = sessionStore.sessions.find(s => s.id === sessId);
    expect(session?.messages).toHaveLength(0);
  });
});
