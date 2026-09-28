<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Button from '$lib/antd/Button.svelte';
  import Badge from '$lib/antd/Badge.svelte';
  import Switch from '$lib/antd/Switch.svelte';
  import Tooltip from '$lib/antd/Tooltip.svelte';
  import SessionTabs from '$lib/components/layout/SessionTabs.svelte';
  import TerminalPanel from '$lib/components/terminal/TerminalPanel.svelte';
  import { terminalStore } from '$lib/stores/terminal.svelte';
  import MessageList from '$lib/components/chat/MessageList.svelte';
  import Composer from '$lib/components/chat/Composer.svelte';
  import PermissionModal from '$lib/components/chat/PermissionModal.svelte';
  import SkillCatalog from '$lib/components/skills/SkillCatalog.svelte';
  import SettingsModal from '$lib/components/layout/SettingsModal.svelte';
  import FileViewerModal from '$lib/components/workspace/FileViewerModal.svelte';
  import WorkspaceSidebar from '$lib/components/layout/WorkspaceSidebar.svelte';
  import RightSidebar from '$lib/components/layout/RightSidebar.svelte';
  import ScreenFlash from '$lib/components/snapshot/ScreenFlash.svelte';
  import ModalConfirm from '$lib/antd/ModalConfirm.svelte';
  import NewSessionDropdown from '$lib/components/layout/NewSessionDropdown.svelte';
  import WorkspacePickerModal from '$lib/components/layout/WorkspacePickerModal.svelte';
  import { dialogStore } from '$lib/stores/dialog.svelte';
  import { playCameraShutterSound } from '$lib/utils/audio';
  import { settingsStore } from '$lib/stores/settings.svelte';
  import { updaterStore } from '$lib/stores/updater.svelte';
  import { ShortcutDetector, type DictationTriggerEvent } from '$lib/utils/shortcutDetector';
  import type { SkillItem, SnapshotResult } from './app.d';
  import {
    sessionStore,
    type VisionImage,
    type AttachedFile,
    type QueuedPrompt,
    type ToolCall,
    type PermissionRequest
  } from '$lib/stores/session.svelte';
  import {
    Bot,
    Sparkles,
    Terminal,
    Zap,
    Camera,
    SlidersHorizontal,
    Code2,
    Shield,
    Settings,
    Volume2,
    PanelLeftClose,
    PanelLeftOpen,
    PanelRightClose,
    PanelRightOpen,
    ChevronRight,
    Upload,
    FolderPlus
  } from 'lucide-svelte';

  let reasoningEffort = $state<'low' | 'medium' | 'high'>('medium');
  let selectedModel = $state('9router');
  let pingResult = $state<string>('');
  let skillsCatalogVisible = $state(false);
  let settingsModalVisible = $state(false);
  let workspacePickerModalVisible = $state(false);
  let emptyStateDropdownOpen = $state(false);
  let flashActive = $state(false);
  const isMac = typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform || navigator.userAgent);

  // Shortcut detector for dictation triggers (push-to-talk & double-tap lock)
  const shortcutDetector = new ShortcutDetector({
    holdThresholdMs: settingsStore.dictationHoldThresholdMs || 300,
    doubleTapThresholdMs: 350,
    onTrigger: (event: DictationTriggerEvent) => {
      if (typeof window !== 'undefined') {
        window.dispatchEvent(new CustomEvent('aethergrok:dictation-trigger', { detail: event }));
      }
    }
  });

  // Global Session Drag and Drop Overlay State
  let isSessionDragOver = $state(false);
  let sessionDragCounter = 0;

  function hasAcceptableData(dataTransfer: DataTransfer | null): boolean {
    if (!dataTransfer || !dataTransfer.types) return false;
    const types = Array.from(dataTransfer.types);
    return (
      types.includes('Files') ||
      types.includes('public.png') ||
      types.includes('public.tiff') ||
      types.includes('public.jpeg') ||
      types.includes('com.apple.traditional-mac-plain-text') ||
      types.some((t) => t.startsWith('image/'))
    );
  }

  function handleMainDragEnter(e: DragEvent) {
    if (hasAcceptableData(e.dataTransfer)) {
      e.preventDefault();
      sessionDragCounter++;
      isSessionDragOver = true;
    }
  }

  function handleMainDragOver(e: DragEvent) {
    if (hasAcceptableData(e.dataTransfer)) {
      e.preventDefault();
      if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy';
      isSessionDragOver = true;
    }
  }

  function handleMainDragLeave(e: DragEvent) {
    e.preventDefault();
    sessionDragCounter--;
    if (sessionDragCounter <= 0) {
      sessionDragCounter = 0;
      isSessionDragOver = false;
    }
  }

  function handleMainDrop(e: DragEvent) {
    e.preventDefault();
    sessionDragCounter = 0;
    isSessionDragOver = false;

    // Delegate extraction and processing to composer
    if (e.dataTransfer) {
      const files: File[] = [];
      if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
        for (let i = 0; i < e.dataTransfer.files.length; i++) {
          const f = e.dataTransfer.files[i];
          if (f) files.push(f);
        }
      }
      if (e.dataTransfer.items && e.dataTransfer.items.length > 0) {
        for (let i = 0; i < e.dataTransfer.items.length; i++) {
          const item = e.dataTransfer.items[i];
          if (item.kind === 'file') {
            const file = item.getAsFile();
            if (file && !files.some((existing) => existing.name === file.name && existing.size === file.size)) {
              files.push(file);
            }
          }
        }
      }
      if (files.length > 0) {
        composerRef?.handleExternalFiles?.(files);
      }
    }
  }

  // Sidebar Resizing & Responsive State
  const COMPACT_BREAKPOINT = 840;
  let windowWidth = $state(typeof window !== 'undefined' ? window.innerWidth : 1200);
  let isCompact = $derived(windowWidth < COMPACT_BREAKPOINT);
  let isDraggingSidebar = $state(false);
  let dragStartX = 0;
  let dragStartWidth = 288;

  function handleResizeStart(e: MouseEvent) {
    if (isCompact || settingsStore.sidebarCollapsed) return;
    e.preventDefault();
    e.stopPropagation();
    isDraggingSidebar = true;
    dragStartX = e.clientX;
    dragStartWidth = settingsStore.sidebarWidth || 288;
    document.body.style.userSelect = 'none';
    document.body.style.webkitUserSelect = 'none';
    document.body.style.cursor = 'col-resize';
    if (window.getSelection) {
      window.getSelection()?.removeAllRanges();
    }
    window.addEventListener('mousemove', handleResizeMove);
    window.addEventListener('mouseup', handleResizeEnd);
  }

  function handleResizeMove(e: MouseEvent) {
    if (!isDraggingSidebar) return;
    e.preventDefault();
    const delta = e.clientX - dragStartX;
    const newWidth = dragStartWidth + delta;
    if (newWidth < 160) {
      settingsStore.sidebarCollapsed = true;
    } else {
      settingsStore.sidebarCollapsed = false;
      settingsStore.sidebarWidth = Math.min(480, Math.max(220, newWidth));
    }
  }

  function handleResizeEnd() {
    if (isDraggingSidebar) {
      isDraggingSidebar = false;
      document.body.style.userSelect = '';
      document.body.style.webkitUserSelect = '';
      document.body.style.cursor = '';
      settingsStore.saveToStorage();
      window.removeEventListener('mousemove', handleResizeMove);
      window.removeEventListener('mouseup', handleResizeEnd);
    }
  }

  function handleResetSidebarWidth() {
    settingsStore.sidebarWidth = 288;
    settingsStore.sidebarCollapsed = false;
    settingsStore.saveToStorage();
  }

  function toggleSidebar() {
    settingsStore.sidebarCollapsed = !settingsStore.sidebarCollapsed;
    settingsStore.saveToStorage();
  }

  function handleWindowResize() {
    if (typeof window !== 'undefined') {
      const prevCompact = windowWidth < COMPACT_BREAKPOINT;
      windowWidth = window.innerWidth;
      const nowCompact = windowWidth < COMPACT_BREAKPOINT;
      if (!prevCompact && nowCompact) {
        settingsStore.sidebarCollapsed = true;
      }
    }
  }
  let composerRef = $state<{
    appendText?: (str: string) => void;
    appendPrompt?: (str: string) => void;
    attachImage?: (img: VisionImage) => void;
    focusInput?: () => void;
    restorePrompt?: (payload: { text: string; images?: VisionImage[]; attachments?: AttachedFile[] }) => void;
    handleExternalFiles?: (files: FileList | File[]) => void;
  } | null>(null);
  let messageListRef = $state<{ forceScrollBottom: () => void } | null>(null);

  // Sync settingsStore default values
  $effect(() => {
    if (settingsStore.defaultModel) {
      selectedModel = settingsStore.defaultModel;
    }
    if (['low', 'medium', 'high'].includes(settingsStore.defaultReasoningEffort)) {
      reasoningEffort = settingsStore.defaultReasoningEffort as 'low' | 'medium' | 'high';
    }
  });

  // Apply data-theme attribute on document root
  $effect(() => {
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('data-theme', settingsStore.theme);
    }
  });

  const isWorking = $derived(sessionStore.activeSession?.status === 'working');
  const pendingPermission = $derived(sessionStore.activeSession?.pendingPermission || null);
  const currentSession = $derived(sessionStore.activeSession);
  const currentSessionWorkspace = $derived.by(() => {
    if (!currentSession) return sessionStore.activeWorkspace;
    return sessionStore.workspaces.find((w) => w.id === currentSession.workspaceId) || sessionStore.activeWorkspace;
  });

  function handleSelectSkill(skill: SkillItem) {
    if (composerRef) {
      composerRef.appendText(`/${skill.name}`);
    }
  }

  // Trigger snapshot feedback (Camera Shutter Audio + Screen White Flash)
  function triggerSnapshotEffects() {
    if (settingsStore.snapshotSoundEnabled) {
      playCameraShutterSound(0.5);
    }
    if (settingsStore.snapshotFlashEnabled) {
      flashActive = true;
      setTimeout(() => {
        flashActive = false;
      }, 260);
    }
  }

  // Global Snapshot Action (Invoked via button or Cmd/Ctrl+Shift+S)
  async function performGlobalSnapshot() {
    const delay = settingsStore.snapshotDelayMs || 50;

    let snapshotResult: SnapshotResult | null = null;

    if (window.go?.main?.App?.CaptureScreenExcludingSelf) {
      try {
        const res = await window.go.main.App.CaptureScreenExcludingSelf(delay);
        if (res && res.dataUrl) {
          snapshotResult = res;
        }
      } catch (err) {
        console.error('Failed to capture native screen:', err);
      }
    }

    // Fallback simulation for web browser preview mode
    if (!snapshotResult) {
      const mockCanvas = document.createElement('canvas');
      mockCanvas.width = 1920;
      mockCanvas.height = 1080;
      const ctx = mockCanvas.getContext('2d');
      if (ctx) {
        const grad = ctx.createLinearGradient(0, 0, 1920, 1080);
        grad.addColorStop(0, '#0F1117');
        grad.addColorStop(1, '#181B26');
        ctx.fillStyle = grad;
        ctx.fillRect(0, 0, 1920, 1080);
        ctx.fillStyle = '#00F0FF';
        ctx.font = 'bold 36px monospace';
        ctx.fillText('AetherGrok Screen Capture (Preview Mode)', 80, 140);
        ctx.fillStyle = '#8C93A4';
        ctx.font = '20px sans-serif';
        ctx.fillText(`Timestamp: ${new Date().toISOString()}`, 80, 200);
      }
      snapshotResult = {
        dataUrl: mockCanvas.toDataURL('image/png'),
        filePath: '/tmp/aethergrok_snapshot_preview.png',
        width: 1920,
        height: 1080,
        sizeBytes: 15400,
        timestamp: Date.now()
      };
    }

    // Fire sound & visual flash animation
    triggerSnapshotEffects();

    // Attach image to composer vision context
    if (snapshotResult && settingsStore.snapshotAutoAttach && composerRef) {
      composerRef.attachImage({
        id: 'snap_' + Date.now(),
        dataUrl: snapshotResult.dataUrl,
        filePath: snapshotResult.filePath || `Screen Snapshot (${new Date().toLocaleTimeString()}).jpg`,
        sizeBytes: snapshotResult.sizeBytes || Math.round((snapshotResult.dataUrl.length - (snapshotResult.dataUrl.indexOf(',') + 1)) * 0.75),
        timestamp: Date.now()
      });
      composerRef.focusInput();
    }
  }

  // Handle user turn submission from rich Composer (supports queuing while working)
  async function handleSendMessage(payload: {
    text: string;
    images: VisionImage[];
    attachments?: AttachedFile[];
    model: string;
    reasoningEffort: 'low' | 'medium' | 'high';
  }) {
    const activeSession = sessionStore.activeSession;
    if (!activeSession) return;
    const sessionId = activeSession.id;

    // If Grok is currently working, enqueue the prompt turn
    if (activeSession.status === 'working') {
      const queuedId = 'q_' + Math.random().toString(36).substring(2, 9) + '_' + Date.now().toString(36);
      sessionStore.addQueuedPrompt(sessionId, {
        id: queuedId,
        text: payload.text,
        images: payload.images || [],
        attachments: payload.attachments || [],
        model: payload.model,
        reasoningEffort: payload.reasoningEffort,
        timestamp: Date.now()
      });
      return;
    }

    // Execute immediately if idle
    await executeTurn(sessionId, payload);
  }

  // Execute a prompt turn
  async function executeTurn(sessionId: string, payload: {
    text: string;
    images: VisionImage[];
    attachments?: AttachedFile[];
    model: string;
    reasoningEffort: 'low' | 'medium' | 'high';
    isSteer?: boolean;
  }) {
    // Combine text and any text/document attachments into final prompt
    let fullPromptText = payload.text;
    if (payload.attachments && payload.attachments.length > 0) {
      const docAttachments = payload.attachments.filter(a => !a.isImage && a.content);
      if (docAttachments.length > 0) {
        const docContext = docAttachments.map(d => `--- Attached File: ${d.name} ---\n${d.content}\n--- End of ${d.name} ---`).join('\n\n');
        fullPromptText = fullPromptText ? `${fullPromptText}\n\n${docContext}` : docContext;
      }
    }

    // Record user message with vision images
    const userMsg = sessionStore.addMessage(sessionId, {
      role: 'user',
      content: payload.text,
      images: payload.images.length > 0 ? payload.images : undefined,
      isSteer: payload.isSteer,
      tokens: {
        input: Math.ceil(fullPromptText.length / 4) + payload.images.length * 100,
        output: 0,
        total: Math.ceil(fullPromptText.length / 4) + payload.images.length * 100
      }
    });

    selectedModel = payload.model;
    reasoningEffort = payload.reasoningEffort;
    sessionStore.setSessionStatus(sessionId, 'working');

    // Instantly force scroll to bottom on new prompt submission
    messageListRef?.forceScrollBottom();

    // Auto-derive a provisional title only on the FIRST prompt if session has a default placeholder title
    const currentSession = sessionStore.sessions.find((s) => s.id === sessionId);
    if (currentSession && !currentSession.isCustomTitle && currentSession.messages.filter(m => m.role === 'user').length <= 1 && (currentSession.title.startsWith('Session ') || currentSession.title.startsWith('Percakapan ') || currentSession.title.startsWith('New '))) {
      const firstLine = payload.text.split('\n')[0].trim();
      if (firstLine) {
        const previewTitle = firstLine.length > 38 ? firstLine.slice(0, 38) + '...' : firstLine;
        sessionStore.updateAutoTitle(sessionId, previewTitle);
      }
    }

    // If Wails Go backend is available, run prompt stream
    if (window.go?.main?.App?.RunPromptStream) {
      try {
        const sessionObj = sessionStore.sessions.find((s) => s.id === sessionId);
        const sessionWs = sessionObj ? sessionStore.workspaces.find((w) => w.id === sessionObj.workspaceId) : null;
        const workingDir = sessionWs?.path || sessionStore.activeWorkspace?.path;
        const grokSessionId = sessionObj?.grokSessionId || (sessionObj?.id && !sessionObj.id.startsWith('sess_') ? sessionObj.id : undefined);

        await window.go.main.App.RunPromptStream({
          sessionId,
          prompt: fullPromptText,
          images: payload.images.map((img) => img.filePath),
          options: {
            model: payload.model || undefined,
            reasoningEffort: payload.reasoningEffort,
            workingDir: workingDir || undefined,
            ...(grokSessionId ? { grokSessionId } : {})
          } as any
        });
      } catch (err) {
        sessionStore.addMessage(sessionId, {
          role: 'assistant',
          content: `Failed to execute prompt stream: ${String(err)}`,
          status: 'error'
        });
        sessionStore.setSessionStatus(sessionId, 'error');
        checkAndDispatchNextQueue(sessionId);
      }
    } else {
      // Browser preview mode: simulate agent lifecycle with rich tool calls and diff card
      simulateAgentResponse(sessionId, userMsg.content, payload.images);
    }
  }

  // Handle plan review card actions (Approve, Reject, Custom Feedback)
  async function handlePlanAction(action: 'approve' | 'reject' | 'custom', feedback?: string) {
    const activeSession = sessionStore.activeSession;
    if (!activeSession) return;

    if (action === 'approve') {
      // Send prompt to implement the proposed plan
      await handleSendMessage({
        text: 'The plan is approved. Please implement the changes step by step now.',
        images: [],
        model: selectedModel,
        reasoningEffort: reasoningEffort as 'low' | 'medium' | 'high'
      });
    } else if (action === 'reject') {
      // Send prompt to cancel or rethink
      await handleSendMessage({
        text: 'I reject the proposed plan. Please stop or suggest an alternative approach.',
        images: [],
        model: selectedModel,
        reasoningEffort: reasoningEffort as 'low' | 'medium' | 'high'
      });
    } else if (action === 'custom' && feedback) {
      // Send user's specific feedback for plan revision
      await handleSendMessage({
        text: `Regarding the plan: ${feedback}\nPlease update and revise the plan accordingly.`,
        images: [],
        model: selectedModel,
        reasoningEffort: reasoningEffort as 'low' | 'medium' | 'high'
      });
    }
  }

  // Steer: cancel current turn and immediately run the chosen prompt
  async function handleSteerPrompt(promptItem: QueuedPrompt) {
    const activeSession = sessionStore.activeSession;
    if (!activeSession) return;
    const sessionId = activeSession.id;

    // 1. Remove this item from queue
    sessionStore.removeQueuedPrompt(sessionId, promptItem.id);

    // 2. Cancel current running turn
    await handleCancelSession();

    // 3. Immediately dispatch the steer prompt
    await executeTurn(sessionId, {
      text: promptItem.text,
      images: promptItem.images || [],
      attachments: promptItem.attachments || [],
      model: promptItem.model,
      reasoningEffort: promptItem.reasoningEffort,
      isSteer: true
    });
  }

  // Auto-dequeue helper
  function checkAndDispatchNextQueue(sessionId: string) {
    setTimeout(() => {
      const next = sessionStore.popNextQueuedPrompt(sessionId);
      if (next) {
        executeTurn(sessionId, {
          text: next.text,
          images: next.images || [],
          attachments: next.attachments || [],
          model: next.model,
          reasoningEffort: next.reasoningEffort
        });
      }
    }, 200);
  }

  // Cancel running session
  async function handleCancelSession() {
    const activeSession = sessionStore.activeSession;
    if (!activeSession) return;

    if (window.go?.main?.App?.CancelSession) {
      try {
        await window.go.main.App.CancelSession(activeSession.id);
      } catch (err) {
        console.error('Failed to cancel session:', err);
      }
    }

    sessionStore.setSessionStatus(activeSession.id, 'idle');
    sessionStore.updateLastMessage(activeSession.id, (msg) => {
      if (msg.status === 'streaming') {
        msg.status = 'done';
        msg.content += '\n\n*(Turn cancelled by user)*';
      }
    });
  }

  // Edit last user turn: rollback turn on disk/session and load payload into composer
  async function handleEditLastTurn() {
    const activeSession = sessionStore.activeSession;
    if (!activeSession) return;

    // 1. Cancel session if currently running
    if (activeSession.status === 'working') {
      await handleCancelSession();
    }

    // 2. Perform rollback in session store
    const rollback = sessionStore.rollbackLastUserTurn(activeSession.id);
    if (!rollback) return;

    // 3. Revert workspace files modified during this turn if in git repo
    const ws = sessionStore.activeWorkspace;
    if (ws && rollback.revertFiles.length > 0 && window.go?.main?.App?.RevertWorkspaceFiles) {
      try {
        await window.go.main.App.RevertWorkspaceFiles(ws.path, rollback.revertFiles);
      } catch (err) {
        console.error('Failed to revert workspace files on turn rollback:', err);
      }
    }

    // 4. Restore text, images, and attachments back into composer
    if (composerRef) {
      composerRef.restorePrompt({
        text: rollback.text,
        images: rollback.images,
        attachments: rollback.attachments
      });
    }
  }

  // Permission modal resolution
  async function handlePermissionDecision(decision: 'allow_once' | 'allow_always' | 'reject') {
    const session = sessionStore.activeSession;
    if (!session || !session.pendingPermission) return;

    const req = session.pendingPermission;

    if (window.go?.main?.App?.RespondPermission) {
      try {
        await window.go.main.App.RespondPermission({
          sessionId: session.id,
          requestId: req.requestId,
          decision
        });
      } catch (err) {
        console.error('Failed to respond permission:', err);
      }
    } else {
      // Mock execution in preview mode
      if (decision === 'reject') {
        sessionStore.updateToolCall(session.id, req.requestId, (tool) => {
          tool.status = 'error';
          tool.result = 'Permission denied by user.';
          tool.endTime = Date.now();
        });
      } else {
        sessionStore.updateToolCall(session.id, req.requestId, (tool) => {
          tool.status = 'completed';
          tool.result = 'Execution completed successfully.';
          tool.endTime = Date.now();
        });
      }
    }

    sessionStore.setPendingPermission(session.id, null);
  }

  // Simulate full agent cycle with tool calls and diff card for browser preview & testing
  function simulateAgentResponse(sessionId: string, prompt: string, images: VisionImage[]) {
    const assistantMsg = sessionStore.addMessage(sessionId, {
      role: 'assistant',
      content: images.length > 0
        ? `Analyzing ${images.length} vision frame(s) and prompt: "${prompt}"...`
        : `Analyzing prompt and preparing environment...`,
      status: 'streaming',
      toolCalls: []
    });

    const isEditCommand = prompt.toLowerCase().includes('edit') ||
      prompt.toLowerCase().includes('diff') ||
      prompt.toLowerCase().includes('replace') ||
      prompt.toLowerCase().includes('fix');

    setTimeout(() => {
      // Step 1: Add a tool call
      const toolId = 'tc_' + Math.random().toString(36).substring(2, 8);
      const isTerminal = prompt.toLowerCase().includes('run') || prompt.toLowerCase().includes('cmd');

      const toolCall: ToolCall = isTerminal
        ? {
            id: toolId,
            tool: 'run_terminal_cmd',
            params: { command: 'pnpm test && git status --short' },
            result: 'PASS src/app.test.ts (4 tests passed)\nM frontend/src/App.svelte',
            status: 'completed',
            startTime: Date.now() - 140,
            endTime: Date.now()
          }
        : isEditCommand
        ? {
            id: toolId,
            tool: 'search_replace',
            params: {
              file_path: 'frontend/src/App.svelte',
              old_string: '<div class="old-header">',
              new_string: '<header class="ant-header-v2">'
            },
            result: 'Successfully modified frontend/src/App.svelte (1 replacement applied)',
            status: 'completed',
            startTime: Date.now() - 95,
            endTime: Date.now(),
            diff: {
              oldPath: 'frontend/src/App.svelte',
              newPath: 'frontend/src/App.svelte',
              oldContent: '  <div class="old-header">\n    <h1>Legacy Header</h1>\n  </div>',
              newContent: '  <header class="ant-header-v2">\n    <h1>AetherGrok Studio Header</h1>\n  </header>',
              diffUnified: `--- frontend/src/App.svelte\n+++ frontend/src/App.svelte\n@@ -1,3 +1,3 @@\n-  <div class="old-header">\n-    <h1>Legacy Header</h1>\n-  </div>\n+  <header class="ant-header-v2">\n+    <h1>AetherGrok Studio Header</h1>\n+  </header>`
            }
          }
        : {
            id: toolId,
            tool: 'read_file',
            params: { target_file: 'src/lib/stores/session.svelte.ts' },
            result: 'export const sessionStore = new SessionStore();\n// File read: 284 lines, 0 errors.',
            status: 'completed',
            startTime: Date.now() - 40,
            endTime: Date.now()
          };

      sessionStore.updateLastMessage(sessionId, (msg) => {
        msg.toolCalls = [toolCall];
        msg.content = `### Execution Plan Complete\n\nI processed your request using **${selectedModel}** (Effort: *${reasoningEffort}*).\n\n1. Inspected workspace state.\n2. Executed tool \`${toolCall.tool}\`.\n3. Verified output and diff integrity.\n\nAll operations succeeded with 0 errors.`;
        msg.status = 'done';
        msg.tokens = { input: 240, output: 145, total: 385 };
      });

      sessionStore.setSessionStatus(sessionId, 'finished');
      checkAndDispatchNextQueue(sessionId);
    }, 600);
  }

  // Helper to test if a keydown matches configured shortcut string
  function matchesShortcut(e: KeyboardEvent, shortcutStr: string): boolean {
    if (!shortcutStr) return false;

    // Check if target is an input/textarea/contenteditable
    const target = e.target as HTMLElement | null;
    const isEditingText = target && (
      target.tagName === 'INPUT' ||
      target.tagName === 'TEXTAREA' ||
      target.isContentEditable
    );

    const normCode = e.code ? e.code.toLowerCase().replace(/[\s_-]/g, '') : '';
    const normShortcut = shortcutStr.toLowerCase().replace(/[\s_-]/g, '');

    // 1. Direct code check (e.g. ShiftLeft, ShiftRight, MetaLeft, MetaRight, AltLeft, AltRight, ControlLeft, ControlRight)
    const isDirectMatch = normCode && (
      normCode === normShortcut ||
      (normCode === 'shiftright' && (normShortcut === 'rightshift' || normShortcut === 'shiftright')) ||
      (normCode === 'shiftleft' && (normShortcut === 'leftshift' || normShortcut === 'shiftleft')) ||
      (normCode === 'controlright' && (normShortcut === 'rightctrl' || normShortcut === 'controlright' || normShortcut === 'rightcontrol')) ||
      (normCode === 'controlleft' && (normShortcut === 'leftctrl' || normShortcut === 'controlleft' || normShortcut === 'leftcontrol')) ||
      (normCode === 'altright' && (normShortcut === 'rightalt' || normShortcut === 'altright' || normShortcut === 'rightoption' || normShortcut === 'rightopt')) ||
      (normCode === 'altleft' && (normShortcut === 'leftalt' || normShortcut === 'altleft' || normShortcut === 'leftoption' || normShortcut === 'leftopt')) ||
      (normCode === 'metaright' && (normShortcut === 'rightcmd' || normShortcut === 'metaright' || normShortcut === 'rightwin' || normShortcut === 'rightmeta')) ||
      (normCode === 'metaleft' && (normShortcut === 'leftcmd' || normShortcut === 'metaleft' || normShortcut === 'leftwin' || normShortcut === 'leftmeta'))
    );

    if (isDirectMatch) {
      if (isEditingText) return false;
      return true;
    }

    // If editing text, do not fire single-key printable shortcuts like '/' or 'delete' unless modifiers are held
    const parts = shortcutStr.toLowerCase().split('+').map((s) => s.trim());
    const hasCmdOrCtrl = parts.includes('cmdorctrl') || parts.includes('cmd') || parts.includes('ctrl') || parts.includes('meta');
    const hasShift = parts.includes('shift');
    const hasAlt = parts.includes('alt') || parts.includes('opt') || parts.includes('option');

    const keyPart = parts.find((p) => !['cmdorctrl', 'cmd', 'ctrl', 'meta', 'shift', 'alt', 'opt', 'option'].includes(p));

    if (isEditingText && !hasCmdOrCtrl && !hasAlt) {
      // Don't intercept normal typing in inputs
      return false;
    }

    const isMetaOrCtrl = e.metaKey || e.ctrlKey;
    if (hasCmdOrCtrl && !isMetaOrCtrl) return false;
    if (!hasCmdOrCtrl && isMetaOrCtrl) return false;
    if (hasShift && !e.shiftKey) return false;
    if (!hasShift && e.shiftKey) return false;
    if (hasAlt && !e.altKey) return false;
    if (!hasAlt && e.altKey) return false;

    if (keyPart) {
      const eKey = e.key.toLowerCase();
      const eCode = e.code.toLowerCase();
      if (eKey === keyPart) return true;
      if (eCode === keyPart || eCode === `key${keyPart}` || eCode === `digit${keyPart}`) return true;
      if (keyPart === '/' && (eKey === '/' || eCode === 'slash')) return true;
      if (keyPart === 'delete' && (eKey === 'delete' || eCode === 'delete')) return true;
      if (keyPart === 'backspace' && (eKey === 'backspace' || eCode === 'backspace')) return true;
      if (keyPart === 'space' && (eKey === ' ' || eCode === 'space')) return true;
      return false;
    }

    return true;
  }

  // Handle Open New Folder in Empty State & Dropdowns
  async function handleOpenNewFolder() {
    if (window.go?.main?.App?.SelectWorkspaceDirectory) {
      try {
        const dir = await window.go.main.App.SelectWorkspaceDirectory();
        if (dir) {
          const folderName = dir.split(/[/\\]/).filter(Boolean).pop() || 'workspace';
          const newWs = sessionStore.addWorkspace(folderName, dir);
          if (window.go?.main?.App?.DiscoverGrokSessions) {
            const diskSessions = await window.go.main.App.DiscoverGrokSessions(dir);
            if (diskSessions && diskSessions.length > 0) {
              sessionStore.syncDiscoveredGrokSessions(newWs.id, diskSessions);
            }
          }
        }
      } catch (err) {
        console.error('Failed to open workspace directory:', err);
      }
    } else {
      const path = window.prompt('Enter absolute path of folder workspace:', '/Users/fiko942/Desktop/workspace');
      if (path && path.trim()) {
        const folderName = path.trim().split(/[/\\]/).filter(Boolean).pop() || 'workspace';
        sessionStore.addWorkspace(folderName, path.trim());
      }
    }
  }

  // Toggle Window Maximization (Zoom / Fill Screen on current desktop) on Header Double Click
  async function handleHeaderDoubleClick(e: MouseEvent) {
    const target = e.target as HTMLElement | null;
    if (target?.closest('button, input, select, textarea, a, [role="button"], [role="tab"]')) {
      return;
    }

    // Ensure we exit macOS Space fullscreen if accidentally active
    if (window.runtime?.WindowIsFullscreen) {
      try {
        const isFull = await window.runtime.WindowIsFullscreen();
        if (isFull) {
          window.runtime.WindowUnfullscreen();
          return;
        }
      } catch {
        // ignore
      }
    }

    // Toggle window maximize to fill the current desktop screen without creating a separate macOS space
    if (window.runtime?.WindowToggleMaximise) {
      try {
        window.runtime.WindowToggleMaximise();
      } catch (err) {
        console.error('Failed to toggle window maximize via Wails runtime:', err);
      }
    } else if (window.runtime?.WindowMaximise) {
      window.runtime.WindowMaximise();
    }
  }

  // Global Keyboard Shortcuts Handler
  function handleGlobalKeyDown(e: KeyboardEvent) {
    // Feed dictation shortcut detector
    const dictationEvent = shortcutDetector.feedKeyDown(e, settingsStore.dictationShortcut);
    if (dictationEvent) {
      window.dispatchEvent(new CustomEvent('aethergrok:dictation-trigger', { detail: dictationEvent }));
    }

    const isMetaOrCtrl = e.metaKey || e.ctrlKey;

    // Smart Screen Snapshot (Customizable via settingsStore.snapshotShortcut)
    if (matchesShortcut(e, settingsStore.snapshotShortcut) || (isMetaOrCtrl && e.shiftKey && e.key.toLowerCase() === 's')) {
      e.preventDefault();
      performGlobalSnapshot();
      return;
    }

    // Cmd/Ctrl + K: Open Skills Catalog
    if (isMetaOrCtrl && e.key.toLowerCase() === 'k') {
      e.preventDefault();
      skillsCatalogVisible = !skillsCatalogVisible;
      return;
    }

    // Cmd/Ctrl + ,: Open Settings Modal
    if (isMetaOrCtrl && e.key === ',') {
      e.preventDefault();
      settingsModalVisible = !settingsModalVisible;
      return;
    }

    // Cmd/Ctrl + T: Open Workspace Picker for New Conversation
    if (isMetaOrCtrl && !e.shiftKey && e.key.toLowerCase() === 't') {
      e.preventDefault();
      workspacePickerModalVisible = true;
      return;
    }

    // Cmd/Ctrl + W: Close active session tab (allows 0 tabs in view)
    if (isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key.toLowerCase() === 'w') {
      e.preventDefault();
      if (sessionStore.activeSessionId) {
        sessionStore.closeSessionTab(sessionStore.activeSessionId);
      }
      return;
    }

    // Cmd/Ctrl + Alt + B: Toggle right sidebar (workspace explorer / git inspector)
    if (isMetaOrCtrl && e.altKey && e.key.toLowerCase() === 'b') {
      e.preventDefault();
      sessionStore.toggleRightSidebar();
      return;
    }

    // Cmd/Ctrl + B: Toggle left sidebar collapse
    if (isMetaOrCtrl && !e.shiftKey && !e.altKey && (e.key.toLowerCase() === 'b' || e.key === '\\')) {
      e.preventDefault();
      toggleSidebar();
      return;
    }

    // Cmd/Ctrl + 1-9: Browser-style quick session tab switching
    if (isMetaOrCtrl && !e.shiftKey && !e.altKey && e.key >= '1' && e.key <= '9') {
      const tabs = sessionStore.openWorkspaceTabs;
      if (tabs.length > 0) {
        e.preventDefault();
        if (e.key === '9') {
          // Switch to last open tab
          const lastTab = tabs[tabs.length - 1];
          if (lastTab) sessionStore.switchSession(lastTab.id);
        } else {
          const tabIndex = parseInt(e.key, 10) - 1;
          if (tabIndex >= 0 && tabIndex < tabs.length) {
            const targetTab = tabs[tabIndex];
            if (targetTab) sessionStore.switchSession(targetTab.id);
          }
        }
        return;
      }
    }
  }

  function handleGlobalKeyUp(e: KeyboardEvent) {
    const dictationEvent = shortcutDetector.feedKeyUp(e, settingsStore.dictationShortcut);
    if (dictationEvent) {
      window.dispatchEvent(new CustomEvent('aethergrok:dictation-trigger', { detail: dictationEvent }));
    }
  }

  // Global link click interceptor: ensures any external link clicked in the webview
  // opens in default OS browser (macOS Safari/Chrome/etc.) rather than navigating the app window
  function handleGlobalDocumentClick(e: MouseEvent) {
    const target = e.target as HTMLElement | null;
    const linkEl = target?.closest('a') as HTMLAnchorElement | null;
    if (linkEl && linkEl.href) {
      const href = linkEl.href;
      // If it's a web URL (http:// or https://) or mailto/etc.
      if (/^https?:\/\//i.test(href) || /^mailto:/i.test(href)) {
        e.preventDefault();
        e.stopPropagation();
        const win = window as any;
        if (win.go?.main?.App?.OpenExternalURL) {
          win.go.main.App.OpenExternalURL(href);
        } else if (win.runtime?.BrowserOpenURL) {
          win.runtime.BrowserOpenURL(href);
        } else {
          window.open(href, '_blank');
        }
      }
    }
  }

  // Wails Event Listeners
  let unsubDelta: (() => void) | undefined;
  let unsubTool: (() => void) | undefined;
  let unsubPerm: (() => void) | undefined;
  let unsubComplete: (() => void) | undefined;
  let unsubError: (() => void) | undefined;
  let unsubGlobalSnapshot: (() => void) | undefined;

  // Keep native global snapshot shortcut registered on OS level
  $effect(() => {
    const sc = settingsStore.snapshotShortcut;
    if (window.go?.main?.App?.RegisterGlobalSnapshotShortcut && sc) {
      window.go.main.App.RegisterGlobalSnapshotShortcut(sc).catch((err: unknown) => {
        console.error('Failed to register OS global snapshot shortcut:', err);
      });
    }
  });

  onMount(() => {
    window.addEventListener('click', handleGlobalDocumentClick, true);
    window.addEventListener('keydown', handleGlobalKeyDown);
    window.addEventListener('keyup', handleGlobalKeyUp);
    window.addEventListener('resize', handleWindowResize);
    handleWindowResize();

    // Initialize auto-update background polling on startup
    updaterStore.initPeriodicCheck();

    // Check and request macOS Accessibility / Input Monitoring permissions on startup
    if (window.go?.main?.App?.CheckAndRequestAccessibilityPermissions) {
      window.go.main.App.CheckAndRequestAccessibilityPermissions()
        .then((status: { granted: boolean; message: string; platform: string }) => {
          if (!status.granted && status.platform === 'darwin') {
            console.warn('macOS Accessibility permission prompt triggered:', status.message);
          }
        })
        .catch((err: unknown) => {
          console.error('Error checking macOS accessibility permissions:', err);
        });
    }

    // Hook Wails native runtime events if available
    if (window.runtime?.EventsOn) {
      unsubGlobalSnapshot = window.runtime.EventsOn('snapshot:trigger_global', () => {
        performGlobalSnapshot();
      });
      unsubDelta = window.runtime.EventsOn('grok:delta_batch', (event: { sessionId: string; delta: string; role?: string }) => {
        if (event.sessionId) {
          const session = sessionStore.sessions.find((s) => s.id === event.sessionId);
          if (session && session.status !== 'working') {
            sessionStore.setSessionStatus(event.sessionId, 'working');
          }
          sessionStore.appendDelta(event.sessionId, event.delta, (event.role as 'assistant' | 'user') || 'assistant');
        }
      });

      unsubTool = window.runtime.EventsOn('grok:tool_call', (event: {
        sessionId: string;
        toolId: string;
        toolName: string;
        input?: Record<string, unknown>;
        output?: string;
        status: string;
      }) => {
        if (event.sessionId) {
          const session = sessionStore.sessions.find((s) => s.id === event.sessionId);
          if (session && session.status !== 'working') {
            sessionStore.setSessionStatus(event.sessionId, 'working');
          }
          if (!session) return;

          // Dynamically sync agentMode per session on enter_plan_mode / exit_plan_mode events
          const toolLower = (event.toolName || '').toLowerCase();
          if (toolLower.includes('enter_plan_mode')) {
            session.agentMode = 'plan';
          } else if (toolLower.includes('exit_plan_mode')) {
            session.agentMode = 'agent';
          }

          let found = false;
          for (const msg of session.messages) {
            const tool = msg.toolCalls?.find((t) => t.id === event.toolId);
            if (tool) {
              tool.status = event.status === 'completed' ? 'completed' : event.status === 'failed' ? 'error' : 'running';
              tool.result = event.output;
              tool.endTime = Date.now();
              found = true;
              break;
            }
          }

          if (!found) {
            // Ensure tool calls attach to an assistant message in the active turn
            const lastMsg = session.messages[session.messages.length - 1];
            if (lastMsg && lastMsg.role === 'assistant') {
              if (!lastMsg.toolCalls) lastMsg.toolCalls = [];
              lastMsg.toolCalls.push({
                id: event.toolId,
                tool: event.toolName,
                params: event.input,
                result: event.output,
                status: event.status === 'completed' ? 'completed' : event.status === 'failed' ? 'error' : 'running',
                startTime: Date.now()
              });
            } else {
              // Create an assistant container message for tool calls
              sessionStore.addMessage(event.sessionId, {
                role: 'assistant',
                content: '',
                status: 'streaming',
                toolCalls: [{
                  id: event.toolId,
                  tool: event.toolName,
                  params: event.input,
                  result: event.output,
                  status: event.status === 'completed' ? 'completed' : event.status === 'failed' ? 'error' : 'running',
                  startTime: Date.now()
                }]
              });
            }
          }
        }
      });

      unsubPerm = window.runtime.EventsOn('grok:permission_request', (event: PermissionRequest) => {
        if (event.sessionId) {
          // If permission mode is bypass/unrestricted or auto, auto-approve immediately
          const permMode = settingsStore.permissionMode;
          if (permMode === 'bypassPermissions') {
            handlePermissionDecision('allow_always');
            return;
          } else if (permMode === 'auto') {
            handlePermissionDecision('allow_once');
            return;
          } else if (permMode === 'acceptEdits') {
            const isFileEdit = event.toolName === 'write' || event.toolName === 'search_replace' || event.toolName === 'edit';
            if (isFileEdit) {
              handlePermissionDecision('allow_once');
              return;
            }
          }
          sessionStore.setPendingPermission(event.sessionId, event);
        }
      });

      unsubComplete = window.runtime.EventsOn('grok:complete', async (event: { sessionId: string; status: string; grokSessionId?: string; title?: string }) => {
        if (event.sessionId) {
          sessionStore.setSessionStatus(event.sessionId, event.status === 'success' ? 'finished' : 'error');
          sessionStore.updateLastMessage(event.sessionId, (msg) => {
            msg.status = 'done';
          });

          // Auto-update title if Grok emitted a summary title
          if (event.title) {
            sessionStore.updateAutoTitle(event.sessionId, event.title, event.grokSessionId);
          }

          // Rescan workspace on disk to sync official Grok titles & IDs from summary.json
          const sessionObj = sessionStore.sessions.find((s) => s.id === event.sessionId);
          const sessionWs = sessionObj ? sessionStore.workspaces.find((w) => w.id === sessionObj.workspaceId) : null;
          const ws = sessionWs || sessionStore.activeWorkspace;
          if (ws && window.go?.main?.App?.DiscoverGrokSessions) {
            try {
              const diskSessions = await window.go.main.App.DiscoverGrokSessions(ws.path);
              if (diskSessions && diskSessions.length > 0) {
                sessionStore.syncDiscoveredGrokSessions(ws.id, diskSessions);
              }
            } catch (err) {
              console.error('Failed to auto-sync sessions after turn:', err);
            }
          }

          // Automatically pop and dispatch next queued prompt if available
          checkAndDispatchNextQueue(event.sessionId);
        }
      });

      unsubError = window.runtime.EventsOn('grok:error', (event: { sessionId: string; error: string }) => {
        if (event.sessionId) {
          sessionStore.setSessionStatus(event.sessionId, 'error');
          sessionStore.addMessage(event.sessionId, {
            role: 'assistant',
            content: `Error: ${event.error}`,
            status: 'error'
          });
        }
      });
    }
  });

  onDestroy(() => {
    window.removeEventListener('click', handleGlobalDocumentClick, true);
    window.removeEventListener('keydown', handleGlobalKeyDown);
    window.removeEventListener('keyup', handleGlobalKeyUp);
    window.removeEventListener('resize', handleWindowResize);
    window.removeEventListener('mousemove', handleResizeMove);
    window.removeEventListener('mouseup', handleResizeEnd);
    unsubDelta?.();
    unsubTool?.();
    unsubPerm?.();
    unsubComplete?.();
    unsubError?.();
    unsubGlobalSnapshot?.();
  });
</script>

<div class="flex flex-col h-screen w-screen bg-ant-bg text-ant-text select-none overflow-hidden font-serif">
  <!-- Top Navigation Bar -->
  <header
    ondblclick={handleHeaderDoubleClick}
    class="flex items-center justify-between pl-20 pr-4 h-[38px] bg-ant-bg-secondary border-b border-ant-border-secondary dark:border-white/5 flex-shrink-0 cursor-default"
    style="--wails-draggable:drag"
  >
    <div class="flex items-center space-x-2 shrink-0">
      <!-- Sidebar Toggle Button in Header -->
      <Tooltip
        title={settingsStore.sidebarCollapsed ? "Expand Sidebar" : "Collapse Sidebar"}
        shortcut={isMac ? "⌘B" : "Ctrl+B"}
        placement="bottom"
      >
        <button
          type="button"
          onclick={toggleSidebar}
          class="w-6 h-6 flex items-center justify-center rounded-md text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg transition-colors cursor-pointer shrink-0"
          style="--wails-draggable:no-drag"
        >
          {#if settingsStore.sidebarCollapsed}
            <PanelLeftOpen size={15} />
          {:else}
            <PanelLeftClose size={15} />
          {/if}
        </button>
      </Tooltip>

      <img
        src="/brand-emblem-64.png"
        alt="AetherGrok Logo"
        class="w-5 h-5 shrink-0 object-contain drop-shadow-sm select-none pointer-events-none"
        draggable="false"
      />
      <div class="flex items-center space-x-2 shrink-0">
        <span class="font-serif-display font-bold text-sm tracking-tight text-ant-text leading-none select-none">AetherGrok</span>
        <button
          type="button"
          onclick={() => settingsModalVisible = true}
          class="px-1.5 py-0.5 text-[9px] font-mono font-medium bg-ant-bg-tertiary/70 hover:bg-ant-primary/15 hover:text-ant-primary text-ant-text-muted rounded border border-ant-border-secondary dark:border-white/5 leading-none select-none cursor-pointer transition flex items-center gap-1"
          title="Current app version (Click to view updates & changelog)"
        >
          <span>v{__APP_VERSION__}</span>
          {#if updaterStore.updateAvailable}
            <span class="w-1.5 h-1.5 rounded-full bg-amber-500 animate-pulse"></span>
          {/if}
        </button>
      </div>
    </div>

    <div class="flex items-center space-x-2 shrink-0" style="--wails-draggable:no-drag">
      <Button size="small" type="primary" onclick={performGlobalSnapshot}>
        <Camera size={13} class="mr-1" /> Snapshot
      </Button>
      <Button size="small" type="default" onclick={() => skillsCatalogVisible = true}>
        <Sparkles size={13} class="mr-1 text-ant-primary" /> Skills Hub
      </Button>
      <div class="h-6.5 flex items-center space-x-2 text-xs text-ant-text-secondary bg-ant-bg px-2 rounded-md border border-ant-border-secondary dark:border-white/5 shrink-0 whitespace-nowrap shadow-2xs">
        <Badge status={isWorking ? 'processing' : 'success'} />
        <span class="whitespace-nowrap">Model: <strong class="text-ant-text font-medium">{selectedModel}</strong></span>
      </div>
      <Button size="small" type="default" onclick={() => settingsModalVisible = true} class="!px-2 !h-6.5 relative">
        <Settings size={14} class="text-ant-text-secondary hover:text-ant-primary transition-colors" />
        {#if updaterStore.updateAvailable}
          <span class="absolute -top-1 -right-1 w-2 h-2 rounded-full bg-amber-500 animate-pulse ring-2 ring-amber-500/20"></span>
        {/if}
      </Button>

      <!-- Right Sidebar (Inspector) Toggle in Header -->
      <Tooltip
        title={currentSession?.rightSidebarOpen ? "Close Inspector" : "Open Inspector"}
        shortcut={isMac ? "⌘⌥B" : "Ctrl+Alt+B"}
        placement="bottom"
      >
        <button
          type="button"
          onclick={() => sessionStore.toggleRightSidebar()}
          class="w-6 h-6 flex items-center justify-center rounded-md text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg transition-colors cursor-pointer shrink-0"
        >
          {#if currentSession?.rightSidebarOpen}
            <PanelRightClose size={15} class="text-indigo-400" />
          {:else}
            <PanelRightOpen size={15} />
          {/if}
        </button>
      </Tooltip>
    </div>
  </header>

  <!-- Main Layout Grid -->
  <div class="flex flex-1 overflow-hidden relative">
    <!-- Compact Backdrop Overlay (<840px breakpoint) -->
    {#if isCompact && !settingsStore.sidebarCollapsed}
      <div
        role="button"
        tabindex="0"
        aria-label="Close sidebar overlay"
        onclick={toggleSidebar}
        onkeydown={(e) => (e.key === 'Escape' || e.key === 'Enter') && toggleSidebar()}
        class="fixed inset-0 top-12 bg-black/50 backdrop-blur-xs z-30 transition-opacity duration-200 cursor-pointer"
      ></div>
    {/if}

    <!-- Left Sidebar: Workspace & Session Management (Resizable & Collapsible) -->
    <aside
      class="{isCompact ? 'fixed top-12 bottom-0 left-0 z-40 shadow-2xl transition-transform duration-200 ease-out' : `relative ${isDraggingSidebar ? 'transition-none' : 'transition-[width] duration-200 ease-out'}`} bg-ant-bg-secondary border-r border-ant-border-secondary dark:border-white/5 flex flex-col justify-between overflow-hidden flex-shrink-0 {isDraggingSidebar ? 'select-none pointer-events-none' : ''}"
      style="{isCompact ? (settingsStore.sidebarCollapsed ? 'transform: translateX(-100%); width: 288px;' : 'transform: translateX(0); width: 288px;') : (settingsStore.sidebarCollapsed ? 'width: 0px; padding: 0px; border-right: none;' : `width: ${settingsStore.sidebarWidth || 288}px;`)}"
    >
      <!-- Workspace Folders & Sessions List -->
      <div class="flex-1 overflow-hidden min-h-0">
        <WorkspaceSidebar />
      </div>

      <!-- Compact Engine Controls & Quick Settings Footer -->
      <div class="p-2.5 border-t border-ant-border-secondary dark:border-white/5 flex-shrink-0 bg-ant-bg-secondary">
        <Tooltip
          title="Settings"
          shortcut={isMac ? '⌘,' : 'Ctrl+,'}
          placement="top"
          class="w-full block"
        >
          <button
            type="button"
            onclick={() => settingsModalVisible = true}
            class="w-full flex items-center justify-between px-2.5 py-2 bg-ant-bg hover:bg-ant-bg-tertiary border border-ant-border-secondary dark:border-white/5 hover:border-blue-500/40 rounded-lg text-xs text-ant-text transition group cursor-pointer shadow-2xs"
          >
            <div class="flex items-center space-x-2 min-w-0">
              <div class="w-5 h-5 rounded-md bg-ant-primary/10 flex items-center justify-center text-ant-primary group-hover:scale-105 transition-transform flex-shrink-0">
                <Settings size={13} />
              </div>
              <span class="font-serif text-xs font-medium text-ant-text truncate">Settings</span>
            </div>
            {#if updaterStore.updateAvailable}
              <span class="w-2 h-2 rounded-full bg-amber-500 animate-pulse ring-2 ring-amber-500/20 shrink-0" title="Update available"></span>
            {/if}
          </button>
        </Tooltip>
      </div>
    </aside>

    <!-- 6px Drag Resize Handle Divider (Desktop Mode only) -->
    {#if !isCompact && !settingsStore.sidebarCollapsed}
      <div
        role="separator"
        aria-orientation="vertical"
        tabindex="0"
        onmousedown={handleResizeStart}
        ondblclick={handleResetSidebarWidth}
        title="Drag to resize (220-480px), Double-click to reset"
        class="w-[6px] -ml-[3px] hover:w-[6px] hover:bg-ant-primary/40 active:bg-ant-primary transition-colors cursor-col-resize z-20 select-none flex-shrink-0 relative group flex items-center justify-center"
      >
        <div class="w-[2px] h-8 rounded-full bg-ant-border group-hover:bg-ant-primary transition-colors"></div>
      </div>
    {/if}

    <!-- Collapsed Floating Edge Indicator Button -->
    {#if settingsStore.sidebarCollapsed}
      <div class="absolute top-3 left-2 z-20">
        <Tooltip
          title="Expand Sidebar"
          shortcut={isMac ? "⌘B" : "Ctrl+B"}
          placement="right"
        >
          <button
            type="button"
            onclick={toggleSidebar}
            class="flex items-center justify-center w-7 h-7 rounded-md bg-ant-bg-secondary hover:bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-primary border border-ant-border-secondary dark:border-white/5 shadow-md backdrop-blur-sm transition-all cursor-pointer"
          >
            <ChevronRight size={15} />
          </button>
        </Tooltip>
      </div>
    {/if}

    <!-- Center Workspace: Tabs & Chat Engine -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <main
      class="flex-1 flex flex-col min-w-0 bg-ant-bg overflow-hidden relative"
      ondragenter={handleMainDragEnter}
      ondragover={handleMainDragOver}
      ondragleave={handleMainDragLeave}
      ondrop={handleMainDrop}
    >
      <!-- Visual Drag-and-Drop Active Overlay across entire Chat Session -->
      {#if isSessionDragOver}
        <div class="absolute inset-0 z-50 bg-[#121316]/95 border-2 border-dashed border-ant-primary/60 rounded-xl m-2 flex flex-col items-center justify-center space-y-3 backdrop-blur-md pointer-events-none animate-in fade-in zoom-in-95 duration-150 select-none shadow-2xl">
          <div class="w-14 h-14 rounded-2xl bg-ant-primary/15 text-ant-primary flex items-center justify-center shadow-lg border border-ant-primary/25">
            <Upload size={28} />
          </div>
          <div class="text-center space-y-1">
            <p class="text-base font-serif font-semibold text-ant-text">
              Drop files or images anywhere to attach
            </p>
            <p class="text-xs font-mono text-ant-text-secondary">
              Images, Markdown, PDF, or code files will be attached to prompt
            </p>
          </div>
        </div>
      {/if}

      <!-- Session Tabs Bar (Drag & Drop + Badges) -->
      <SessionTabs />

      {#if sessionStore.activeSession}
        <!-- Main Chat & Right-Docked Terminal Container -->
        <div class="flex-1 flex min-h-0 overflow-hidden relative">
          <!-- Left Column: Chat Feed & Bottom Terminal & Composer -->
          <div class="flex-1 flex flex-col min-w-0 min-h-0 overflow-hidden relative">
            <!-- Chat Feed Viewport (10-Turn Windowing) -->
            <div class="flex-1 overflow-hidden relative">
              <MessageList
                bind:this={messageListRef}
                onEditLastTurn={handleEditLastTurn}
                onPlanAction={handlePlanAction}
              />
            </div>

            <!-- Bottom Docked Terminal Panel -->
            {#if terminalStore.getDockPosition(sessionStore.activeSession.id) === 'bottom'}
              <TerminalPanel
                sessionId={sessionStore.activeSession.id}
                workspacePath={sessionStore.activeWorkspace?.path || ''}
                onAttachLogToComposer={(logText) => {
                  if (composerRef) {
                    composerRef.appendPrompt(logText);
                  }
                }}
              />
            {/if}

            <!-- Rich Prompt Composer with Snapshot & Model Selectors -->
            <Composer
              bind:this={composerRef}
              {isWorking}
              onSend={handleSendMessage}
              onSteer={handleSteerPrompt}
              onCancel={handleCancelSession}
              onOpenSkillsCatalog={() => skillsCatalogVisible = true}
            />
          </div>

          <!-- Right Docked Terminal Panel -->
          {#if terminalStore.getDockPosition(sessionStore.activeSession.id) === 'right'}
            <TerminalPanel
              sessionId={sessionStore.activeSession.id}
              workspacePath={sessionStore.activeWorkspace?.path || ''}
              onAttachLogToComposer={(logText) => {
                if (composerRef) {
                  composerRef.appendPrompt(logText);
                }
              }}
            />
          {/if}
        </div>
      {:else}
        <!-- Zero-Tab Empty Workspace State -->
        <div class="flex-1 flex flex-col items-center justify-center p-8 select-none text-center bg-radial from-ant-bg-secondary/40 via-ant-bg to-ant-bg">
          <div class="w-16 h-16 rounded-2xl bg-ant-bg-elevated border border-white/5 flex items-center justify-center mb-5 shadow-xl shadow-black/40 overflow-hidden group">
            <img
              src="/brand-emblem-128.png"
              alt="AetherGrok Logo"
              class="w-11 h-11 object-contain drop-shadow-md select-none pointer-events-none"
              draggable="false"
            />
          </div>
          <h2 class="text-xl font-serif font-bold text-ant-text tracking-tight mb-2">No Active Session</h2>
          <p class="text-sm font-serif text-ant-text-secondary max-w-md mb-6 leading-relaxed">
            All conversation tabs are closed. Any background tasks or queued runs continue working automatically.
          </p>
          <div class="flex items-center space-x-3 relative">
            <div class="relative">
              <button
                type="button"
                onclick={(e) => {
                  e.stopPropagation();
                  emptyStateDropdownOpen = !emptyStateDropdownOpen;
                }}
                class="flex items-center space-x-2 px-4 py-2 rounded-lg bg-ant-primary hover:bg-ant-primary/90 text-white font-serif text-xs font-semibold shadow-md shadow-ant-primary/20 transition cursor-pointer"
              >
                <span>New Conversation</span>
                <kbd class="px-1.5 py-0.5 text-[10px] font-mono bg-white/20 rounded text-white">{isMac ? '⌘T' : 'Ctrl+T'}</kbd>
              </button>

              <NewSessionDropdown
                bind:open={emptyStateDropdownOpen}
                placement="bottom-center"
                onClose={() => emptyStateDropdownOpen = false}
              />
            </div>
            <button
              type="button"
              onclick={handleOpenNewFolder}
              class="flex items-center space-x-2 px-4 py-2 rounded-lg bg-ant-bg-tertiary hover:bg-white/10 text-ant-text font-serif text-xs font-medium border border-white/5 transition cursor-pointer"
            >
              <FolderPlus size={14} class="text-ant-primary" />
              <span>Open New Folder</span>
            </button>
          </div>
        </div>
      {/if}
    </main>

    <!-- Right Sidebar: Workspace Explorer & Git Changes Inspector (Session-Isolated) -->
    <RightSidebar
      workspacePath={currentSessionWorkspace?.path || ''}
      sessionId={currentSession?.id}
    />
  </div>

  <!-- Screen White Flash Visual Animation -->
  <ScreenFlash active={flashActive} />

  <!-- Skills & MCP Discovery Catalog Modal -->
  <SkillCatalog
    visible={skillsCatalogVisible}
    onClose={() => skillsCatalogVisible = false}
    onSelectSkill={handleSelectSkill}
  />

  <!-- Ant Design Settings & Theme Preferences Modal -->
  <SettingsModal
    visible={settingsModalVisible}
    onClose={() => settingsModalVisible = false}
  />

  <!-- Universal Workspace Selector Modal for New Conversation (Cmd+T / Menu) -->
  <WorkspacePickerModal
    visible={workspacePickerModalVisible}
    onClose={() => workspacePickerModalVisible = false}
  />

  <!-- Interactive Permission Modal -->
  {#if pendingPermission}
    <PermissionModal
      request={pendingPermission}
      onDecision={handlePermissionDecision}
      onClose={() => sessionStore.setPendingPermission(sessionStore.activeSessionId || '', null)}
    />
  {/if}

  <!-- Global Full-Window Confirmation Modal -->
  <ModalConfirm
    open={dialogStore.confirmState.open}
    title={dialogStore.confirmState.title}
    content={dialogStore.confirmState.content}
    confirmText={dialogStore.confirmState.confirmText}
    cancelText={dialogStore.confirmState.cancelText}
    type={dialogStore.confirmState.type}
    onConfirm={dialogStore.confirmState.onConfirm}
    onCancel={dialogStore.confirmState.onCancel}
  />

  <!-- Global File Viewer Modal (Opens any clicked file from chat or workspace) -->
  {#if dialogStore.fileViewerState.open}
    <FileViewerModal
      isOpen={true}
      filePath={dialogStore.fileViewerState.filePath}
      workspacePath={dialogStore.fileViewerState.workspacePath || currentSessionWorkspace?.path || ''}
      onClose={() => dialogStore.closeFileViewer()}
    />
  {/if}
</div>
