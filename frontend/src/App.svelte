<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Button from '$lib/antd/Button.svelte';
  import Badge from '$lib/antd/Badge.svelte';
  import Switch from '$lib/antd/Switch.svelte';
  import SessionTabs from '$lib/components/layout/SessionTabs.svelte';
  import MessageList from '$lib/components/chat/MessageList.svelte';
  import Composer from '$lib/components/chat/Composer.svelte';
  import PermissionModal from '$lib/components/chat/PermissionModal.svelte';
  import SkillCatalog from '$lib/components/skills/SkillCatalog.svelte';
  import SettingsModal from '$lib/components/layout/SettingsModal.svelte';
  import WorkspaceSidebar from '$lib/components/layout/WorkspaceSidebar.svelte';
  import ScreenFlash from '$lib/components/snapshot/ScreenFlash.svelte';
  import { playCameraShutterSound } from '$lib/utils/audio';
  import { settingsStore } from '$lib/stores/settings.svelte';
  import type { SkillItem } from './app.d';
  import {
    sessionStore,
    type VisionImage,
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
    Volume2
  } from 'lucide-svelte';

  let autoHideWindow = $state(true);
  let reasoningEffort = $state<'low' | 'medium' | 'high'>('medium');
  let selectedModel = $state('9router');
  let pingResult = $state<string>('');
  let skillsCatalogVisible = $state(false);
  let settingsModalVisible = $state(false);
  let flashActive = $state(false);
  let composerRef = $state<{
    appendText: (str: string) => void;
    attachImage: (img: VisionImage) => void;
    focusInput: () => void;
  } | null>(null);

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

    let snapshotResult: { dataUrl: string; filePath?: string; width?: number; height?: number } | null = null;

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
        height: 1080
      };
    }

    // Fire sound & visual flash animation
    triggerSnapshotEffects();

    // Attach image to composer vision context
    if (snapshotResult && settingsStore.snapshotAutoAttach && composerRef) {
      composerRef.attachImage({
        id: 'snap_' + Date.now(),
        dataUrl: snapshotResult.dataUrl,
        filePath: snapshotResult.filePath || `Screen Snapshot (${new Date().toLocaleTimeString()}).png`,
        sizeBytes: Math.round(snapshotResult.dataUrl.length * 0.75),
        timestamp: Date.now()
      });
      composerRef.focusInput();
    }
  }

  // Wails bridge greeting check
  async function testBridge() {
    if (window.go?.main?.App?.Greet) {
      try {
        pingResult = await window.go.main.App.Greet('Agent');
      } catch (err) {
        pingResult = `Error: ${String(err)}`;
      }
    } else {
      pingResult = 'Go Wails bridge ready (browser preview mode)';
    }
  }

  // Handle user turn submission from rich Composer
  async function handleSendMessage(payload: {
    text: string;
    images: VisionImage[];
    model: string;
    reasoningEffort: 'low' | 'medium' | 'high';
  }) {
    const activeSession = sessionStore.activeSession;
    if (!activeSession) return;
    const sessionId = activeSession.id;

    // Record user message with vision images
    const userMsg = sessionStore.addMessage(sessionId, {
      role: 'user',
      content: payload.text,
      images: payload.images.length > 0 ? payload.images : undefined,
      tokens: {
        input: Math.ceil(payload.text.length / 4) + payload.images.length * 100,
        output: 0,
        total: Math.ceil(payload.text.length / 4) + payload.images.length * 100
      }
    });

    selectedModel = payload.model;
    reasoningEffort = payload.reasoningEffort;
    sessionStore.setSessionStatus(sessionId, 'working');

    // If Wails Go backend is available, run prompt stream
    if (window.go?.main?.App?.RunPromptStream) {
      try {
        const workingDir = sessionStore.activeWorkspace?.path;
        await window.go.main.App.RunPromptStream({
          sessionId,
          prompt: payload.text,
          images: payload.images.map((img) => img.filePath),
          options: {
            model: payload.model || undefined,
            reasoningEffort: payload.reasoningEffort,
            workingDir: workingDir || undefined
          }
        });
      } catch (err) {
        sessionStore.addMessage(sessionId, {
          role: 'assistant',
          content: `Failed to execute prompt stream: ${String(err)}`,
          status: 'error'
        });
        sessionStore.setSessionStatus(sessionId, 'error');
      }
    } else {
      // Browser preview mode: simulate agent lifecycle with rich tool calls and diff card
      simulateAgentResponse(sessionId, userMsg.content, payload.images);
    }
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
    }, 600);
  }

  // Global Keyboard Shortcuts Handler
  function handleGlobalKeyDown(e: KeyboardEvent) {
    const isMetaOrCtrl = e.metaKey || e.ctrlKey;

    // Cmd/Ctrl + Shift + S: Instantaneous Smart Screen Snapshot
    if (isMetaOrCtrl && e.shiftKey && e.key.toLowerCase() === 's') {
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

    // Cmd/Ctrl + T: Create new session
    if (isMetaOrCtrl && !e.shiftKey && e.key.toLowerCase() === 't') {
      e.preventDefault();
      sessionStore.createSession();
      return;
    }
  }

  // Wails Event Listeners
  let unsubDelta: (() => void) | undefined;
  let unsubTool: (() => void) | undefined;
  let unsubPerm: (() => void) | undefined;
  let unsubComplete: (() => void) | undefined;
  let unsubError: (() => void) | undefined;

  onMount(() => {
    testBridge();
    window.addEventListener('keydown', handleGlobalKeyDown);

    // Hook Wails native runtime events if available
    if (window.runtime?.EventsOn) {
      unsubDelta = window.runtime.EventsOn('grok:delta_batch', (event: { sessionId: string; delta: string; role?: string }) => {
        if (event.sessionId) {
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
          if (!session) return;

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
            // Append tool call to last message
            sessionStore.updateLastMessage(event.sessionId, (msg) => {
              if (!msg.toolCalls) msg.toolCalls = [];
              msg.toolCalls.push({
                id: event.toolId,
                tool: event.toolName,
                params: event.input,
                result: event.output,
                status: event.status === 'completed' ? 'completed' : event.status === 'failed' ? 'error' : 'running',
                startTime: Date.now()
              });
            });
          }
        }
      });

      unsubPerm = window.runtime.EventsOn('grok:permission_request', (event: PermissionRequest) => {
        if (event.sessionId) {
          sessionStore.setPendingPermission(event.sessionId, event);
        }
      });

      unsubComplete = window.runtime.EventsOn('grok:complete', (event: { sessionId: string; status: string }) => {
        if (event.sessionId) {
          sessionStore.setSessionStatus(event.sessionId, event.status === 'success' ? 'finished' : 'error');
          sessionStore.updateLastMessage(event.sessionId, (msg) => {
            msg.status = 'done';
          });
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

    // Default sample session setup
    if (sessionStore.activeSession && sessionStore.activeSession.messages.length === 0) {
      sessionStore.addMessage(sessionStore.activeSession.id, {
        role: 'user',
        content: 'Check system readiness and status of AetherGrok GUI engine.'
      });
      sessionStore.addMessage(sessionStore.activeSession.id, {
        role: 'assistant',
        content: 'System diagnostic completed. All components **Svelte 5 Runes**, **Ant Design Dark Tokens**, **Diff Viewer**, and **Smart Snapshot** are initialized and operational.',
        tokens: { input: 154, output: 86, total: 240 },
        toolCalls: [
          {
            id: 'tc_init_01',
            tool: 'read_file',
            params: { target_file: 'frontend/src/App.svelte' },
            result: '// Initialized AetherGrok Desktop UI\nstatus: nominal',
            status: 'completed',
            startTime: Date.now() - 32,
            endTime: Date.now()
          },
          {
            id: 'tc_init_02',
            tool: 'search_replace',
            params: {
              file_path: 'src/config.ts',
              old_string: 'const TIMEOUT = 1000;',
              new_string: 'const TIMEOUT = 5000;'
            },
            result: 'Updated timeout parameter successfully.',
            status: 'completed',
            startTime: Date.now() - 20,
            endTime: Date.now(),
            diff: {
              oldPath: 'src/config.ts',
              newPath: 'src/config.ts',
              diffUnified: '--- src/config.ts\n+++ src/config.ts\n@@ -1,2 +1,2 @@\n-const TIMEOUT = 1000;\n+const TIMEOUT = 5000;'
            }
          }
        ]
      });
      sessionStore.setSessionStatus(sessionStore.activeSession.id, 'idle');
    }
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleGlobalKeyDown);
    unsubDelta?.();
    unsubTool?.();
    unsubPerm?.();
    unsubComplete?.();
    unsubError?.();
  });
</script>

<div class="flex flex-col h-screen w-screen bg-ant-bg text-ant-text select-none overflow-hidden font-sans">
  <!-- Top Navigation Bar -->
  <header
    class="flex items-center justify-between px-4 h-12 bg-ant-bg-secondary border-b border-ant-border flex-shrink-0"
    style="--wails-draggable:drag"
  >
    <div class="flex items-center space-x-3">
      <div class="flex items-center justify-center w-7 h-7 rounded-lg bg-ant-primary/10 border border-ant-primary/30 text-ant-primary shadow-sm">
        <Sparkles size={16} />
      </div>
      <div class="flex items-center space-x-2">
        <span class="font-bold text-sm tracking-tight text-white">AetherGrok</span>
        <span class="px-1.5 py-0.5 text-[10px] font-semibold bg-ant-primary/20 text-ant-primary rounded border border-ant-primary/30">v1.0.0</span>
      </div>
    </div>

    <div class="flex items-center space-x-3">
      <Button size="small" type="primary" onclick={performGlobalSnapshot}>
        <Camera size={13} class="mr-1" /> Snapshot
      </Button>
      <Button size="small" type="default" onclick={() => skillsCatalogVisible = true}>
        <Sparkles size={13} class="mr-1 text-ant-primary" /> Skills Hub
      </Button>
      <div class="flex items-center space-x-2 text-xs text-ant-text-secondary bg-ant-bg px-2.5 py-1 rounded-md border border-ant-border">
        <Badge status={isWorking ? 'processing' : 'success'} />
        <span>Model: <strong class="text-white">{selectedModel}</strong></span>
      </div>
      <Button size="small" type="default" onclick={testBridge}>
        <Zap size={13} class="mr-1 text-ant-primary" /> Test Bridge
      </Button>
      <Button size="small" type="default" onclick={() => settingsModalVisible = true} class="!px-2">
        <Settings size={14} class="text-ant-text-secondary hover:text-ant-primary transition-colors" />
      </Button>
    </div>
  </header>

  <!-- Main Layout Grid -->
  <div class="flex flex-1 overflow-hidden">
    <!-- Left Sidebar: Workspace & Session Management -->
    <aside class="w-72 bg-ant-bg-secondary border-r border-ant-border flex flex-col justify-between p-3 overflow-hidden">
      <!-- Workspace Folders & Sessions List -->
      <div class="flex-1 overflow-hidden min-h-0">
        <WorkspaceSidebar />
      </div>

      <!-- Compact Engine Controls & Diff Status Footer -->
      <div class="pt-3 mt-2 border-t border-ant-border-secondary space-y-2 flex-shrink-0">
        <div class="p-2.5 bg-ant-bg rounded-lg border border-ant-border-secondary space-y-2 text-xs">
          <div class="flex items-center justify-between">
            <span class="text-ant-text-secondary flex items-center">
              <Camera size={13} class="mr-1.5 text-ant-text-muted" /> Auto-Hide
            </span>
            <Switch bind:checked={autoHideWindow} size="small" />
          </div>
          <div class="flex items-center justify-between pt-1.5 border-t border-ant-border/40">
            <span class="text-ant-text-secondary flex items-center">
              <Volume2 size={13} class="mr-1.5 text-ant-text-muted" /> Cekrek Sound
            </span>
            <Switch bind:checked={settingsStore.snapshotSoundEnabled} size="small" />
          </div>
        </div>

        <div class="p-2 bg-ant-bg rounded-lg border border-ant-border flex items-center justify-between text-xs">
          <div class="flex items-center space-x-2">
            <Code2 size={13} class="text-ant-primary" />
            <span class="text-ant-text-secondary text-[11px]">Diff Previewer</span>
          </div>
          <Badge status="success" />
        </div>
      </div>
    </aside>

    <!-- Center Workspace: Tabs & Chat Engine -->
    <main class="flex-1 flex flex-col min-w-0 bg-ant-bg overflow-hidden relative">
      <!-- Session Tabs Bar (Drag & Drop + Badges) -->
      <SessionTabs />

      <!-- Chat Feed Viewport (10-Turn Windowing) -->
      <div class="flex-1 overflow-hidden relative">
        <MessageList />
      </div>

      <!-- Rich Prompt Composer with Snapshot & Model Selectors -->
      <Composer
        bind:this={composerRef}
        {isWorking}
        onSend={handleSendMessage}
        onCancel={handleCancelSession}
        onOpenSkillsCatalog={() => skillsCatalogVisible = true}
      />
    </main>
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

  <!-- Interactive Permission Modal -->
  {#if pendingPermission}
    <PermissionModal
      request={pendingPermission}
      onDecision={handlePermissionDecision}
      onClose={() => sessionStore.setPendingPermission(sessionStore.activeSessionId || '', null)}
    />
  {/if}
</div>
