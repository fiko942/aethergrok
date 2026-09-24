<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/antd/Button.svelte';
  import Card from '$lib/antd/Card.svelte';
  import Badge from '$lib/antd/Badge.svelte';
  import Switch from '$lib/antd/Switch.svelte';
  import SessionTabs from '$lib/components/layout/SessionTabs.svelte';
  import MessageList from '$lib/components/chat/MessageList.svelte';
  import { sessionStore } from '$lib/stores/session.svelte';
  import {
    Bot,
    Sparkles,
    Terminal,
    Layers,
    Cpu,
    Zap,
    Shield,
    SlidersHorizontal,
    Camera,
    Plus,
    Send
  } from 'lucide-svelte';

  let appStatus = $state('Ready');
  let autoHideWindow = $state(true);
  let reasoningEffort = $state<'low' | 'medium' | 'high'>('medium');
  let selectedModel = $state('grok-2-latest');
  let pingResult = $state<string>('');
  let promptText = $state<string>('');

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

  function handleSendPrompt() {
    if (!promptText.trim() || !sessionStore.activeSessionId) return;

    // Add user turn
    const userMsg = sessionStore.addMessage(sessionStore.activeSessionId, {
      role: 'user',
      content: promptText.trim(),
      tokens: { input: promptText.length / 4, output: 0, total: promptText.length / 4 }
    });

    const activeId = sessionStore.activeSessionId;
    sessionStore.setSessionStatus(activeId, 'working');
    promptText = '';

    // Mock an assistant response for verification
    setTimeout(() => {
      sessionStore.addMessage(activeId, {
        role: 'assistant',
        content: `I received your command: "${userMsg.content}". Executing task via Grok engine...`,
        tokens: { input: 120, output: 65, total: 185 },
        toolCalls: [
          {
            id: 'tc_' + Math.random().toString(36).substring(2, 7),
            tool: 'bash',
            params: { command: 'echo "AetherGrok Ready"' },
            result: 'AetherGrok Ready',
            status: 'completed',
            startTime: Date.now() - 45,
            endTime: Date.now()
          }
        ]
      });
      sessionStore.setSessionStatus(activeId, 'finished');
    }, 400);
  }

  onMount(() => {
    testBridge();

    // Populate initial demo turns if session is fresh
    if (sessionStore.activeSession && sessionStore.activeSession.messages.length === 0) {
      sessionStore.addMessage(sessionStore.activeSession.id, {
        role: 'user',
        content: 'Check system readiness and status of AetherGrok GUI engine.'
      });
      sessionStore.addMessage(sessionStore.activeSession.id, {
        role: 'assistant',
        content: 'System diagnostic completed. All components **Svelte 5 Runes**, **Ant Design Dark Tokens**, and **10-Turn Windowing** are initialized and operational.',
        tokens: { input: 154, output: 86, total: 240 },
        toolCalls: [
          {
            id: 'tc_init_01',
            tool: 'system_info',
            params: { check: 'memory_and_compositor' },
            result: '{"idle_ram_mb": 48.2, "compositor_delay_ms": 50, "status": "nominal"}',
            status: 'completed',
            startTime: Date.now() - 32,
            endTime: Date.now()
          }
        ]
      });
      sessionStore.setSessionStatus(sessionStore.activeSession.id, 'idle');
    }
  });
</script>

<div class="flex flex-col h-screen w-screen bg-ant-bg text-ant-text select-none overflow-hidden">
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

    <div class="flex items-center space-x-4">
      <div class="flex items-center space-x-2 text-xs text-ant-text-secondary bg-ant-bg px-2.5 py-1 rounded-md border border-ant-border">
        <Badge status="processing" />
        <span>Grok Agent Engine: <strong class="text-ant-text">{selectedModel}</strong></span>
      </div>
      <Button size="small" type="default" onclick={testBridge}>
        <Zap size={14} class="mr-1 text-ant-primary" /> Test Bridge
      </Button>
    </div>
  </header>

  <!-- Main Layout Grid -->
  <div class="flex flex-1 overflow-hidden">
    <!-- Left Sidebar: Session & Capability Navigation -->
    <aside class="w-64 bg-ant-bg-secondary border-r border-ant-border flex flex-col justify-between p-3 space-y-4">
      <div class="space-y-4">
        <div>
          <div class="text-[11px] font-semibold tracking-wider text-ant-text-muted uppercase px-2 mb-2">
            Active Workspace
          </div>
          <div class="space-y-1">
            <button class="w-full flex items-center space-x-2.5 px-2.5 py-2 rounded-md bg-ant-primary/15 text-ant-primary font-medium text-xs border border-ant-primary/30">
              <Bot size={15} />
              <span class="truncate">Grok Agent Session #1</span>
            </button>
            <button class="w-full flex items-center space-x-2.5 px-2.5 py-2 rounded-md hover:bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-text transition text-xs">
              <Terminal size={15} />
              <span class="truncate">Background Task Runner</span>
            </button>
          </div>
        </div>

        <div>
          <div class="text-[11px] font-semibold tracking-wider text-ant-text-muted uppercase px-2 mb-2">
            Engine Controls
          </div>
          <div class="p-3 bg-ant-bg rounded-lg border border-ant-border-secondary space-y-3">
            <div class="flex items-center justify-between text-xs">
              <span class="text-ant-text-secondary flex items-center">
                <Camera size={13} class="mr-1.5 text-ant-text-muted" /> Auto-Hide Window
              </span>
              <Switch bind:checked={autoHideWindow} size="small" />
            </div>

            <div class="text-xs space-y-1">
              <div class="text-ant-text-secondary flex items-center justify-between">
                <span class="flex items-center"><SlidersHorizontal size={13} class="mr-1.5 text-ant-text-muted" /> Reasoning Effort</span>
                <span class="text-ant-primary font-semibold uppercase text-[10px]">{reasoningEffort}</span>
              </div>
              <div class="grid grid-cols-3 gap-1 pt-1">
                {#each ['low', 'medium', 'high'] as effort}
                  <button
                    class="py-1 text-[11px] rounded font-medium transition {reasoningEffort === effort ? 'bg-ant-primary text-white' : 'bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-text'}"
                    onclick={() => reasoningEffort = effort as 'low' | 'medium' | 'high'}
                  >
                    {effort}
                  </button>
                {/each}
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="p-2 bg-ant-bg rounded-md border border-ant-border-secondary text-[11px] text-ant-text-muted space-y-1">
        <div class="flex items-center justify-between">
          <span>Memory Idle Budget:</span>
          <span class="text-ant-success font-semibold">&lt; 60 MB</span>
        </div>
        <div class="flex items-center justify-between">
          <span>Sliding Batch Interval:</span>
          <span class="text-ant-primary font-semibold">16 ms</span>
        </div>
      </div>
    </aside>

    <!-- Main Content Workspace -->
    <main class="flex-1 flex flex-col bg-ant-bg overflow-hidden">
      <!-- Multi-session Tab Bar -->
      <SessionTabs />

      <!-- Conversation Viewport & Message List with 10-turn windowing -->
      <div class="flex-1 flex flex-col min-h-0 relative">
        <MessageList />

        <!-- Quick Interactive Composer Prompt for Testing / Verification -->
        <div class="p-3 bg-ant-bg-secondary border-t border-ant-border flex items-center gap-2">
          <input
            type="text"
            bind:value={promptText}
            placeholder="Type prompt to launch agent turn..."
            onkeydown={(e) => e.key === 'Enter' && handleSendPrompt()}
            class="flex-1 bg-ant-bg border border-ant-border focus:border-ant-primary rounded-md px-3 py-1.5 text-xs text-white placeholder:text-ant-text-muted outline-none transition"
          />
          <Button type="primary" size="small" onclick={handleSendPrompt}>
            <Send size={13} class="mr-1" /> Send
          </Button>
        </div>
      </div>
    </main>
  </div>
</div>
