<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/antd/Button.svelte';
  import Card from '$lib/antd/Card.svelte';
  import Badge from '$lib/antd/Badge.svelte';
  import Switch from '$lib/antd/Switch.svelte';
  import {
    Bot,
    Sparkles,
    Terminal,
    Layers,
    Cpu,
    Zap,
    Shield,
    SlidersHorizontal,
    Camera
  } from 'lucide-svelte';

  let appStatus = $state('Ready');
  let autoHideWindow = $state(true);
  let reasoningEffort = $state<'low' | 'medium' | 'high'>('medium');
  let selectedModel = $state('grok-2-latest');
  let pingResult = $state<string>('');

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

  onMount(() => {
    testBridge();
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
    <main class="flex-1 flex flex-col bg-ant-bg overflow-y-auto p-6 space-y-6">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-xl font-bold text-white tracking-tight flex items-center gap-2">
            AetherGrok Agentic AI Studio
            <Badge status="success" text="Online" />
          </h1>
          <p class="text-xs text-ant-text-secondary mt-0.5">
            Native Wails v2 + Svelte 5 High-Performance AI Orchestration Engine
          </p>
        </div>
        <div class="flex space-x-2">
          <Button type="primary" size="middle">
            <Sparkles size={14} class="mr-1.5" /> Launch Turn
          </Button>
        </div>
      </div>

      <!-- Overview Cards -->
      <div class="grid grid-cols-3 gap-4">
        <Card title="Process Engine" hoverable>
          {#snippet extra()}
            <Cpu size={14} class="text-ant-primary" />
          {/snippet}
          <div class="space-y-2 text-xs">
            <div class="flex justify-between">
              <span class="text-ant-text-secondary">Stream Parser:</span>
              <span class="font-mono text-ant-primary">NDJSON 16ms Batching</span>
            </div>
            <div class="flex justify-between">
              <span class="text-ant-text-secondary">IPC Latency:</span>
              <span class="font-mono text-ant-success">&lt; 1ms (Wails IPC)</span>
            </div>
            <div class="flex justify-between">
              <span class="text-ant-text-secondary">Process Supervisor:</span>
              <span class="font-mono text-white">Process Group PGID</span>
            </div>
          </div>
        </Card>

        <Card title="Smart Screen Engine" hoverable>
          {#snippet extra()}
            <Camera size={14} class="text-ant-warning" />
          {/snippet}
          <div class="space-y-2 text-xs">
            <div class="flex justify-between">
              <span class="text-ant-text-secondary">macOS Pipeline:</span>
              <span class="font-mono text-white">CoreGraphics Direct</span>
            </div>
            <div class="flex justify-between">
              <span class="text-ant-text-secondary">Windows Pipeline:</span>
              <span class="font-mono text-white">GDI BitBlt Direct</span>
            </div>
            <div class="flex justify-between">
              <span class="text-ant-text-secondary">Window Auto-Hiding:</span>
              <span class="font-mono text-ant-success">{autoHideWindow ? 'Active (0-Artifact)' : 'Disabled'}</span>
            </div>
          </div>
        </Card>

        <Card title="Security & Isolation" hoverable>
          {#snippet extra()}
            <Shield size={14} class="text-ant-success" />
          {/snippet}
          <div class="space-y-2 text-xs">
            <div class="flex justify-between">
              <span class="text-ant-text-secondary">Tool Approvals:</span>
              <span class="font-mono text-ant-warning">User-Gated Modal</span>
            </div>
            <div class="flex justify-between">
              <span class="text-ant-text-secondary">Local Persistence:</span>
              <span class="font-mono text-white">SQLite (~/.grok)</span>
            </div>
            <div class="flex justify-between">
              <span class="text-ant-text-secondary">Context Sliding:</span>
              <span class="font-mono text-ant-primary">10-Turn Windowing</span>
            </div>
          </div>
        </Card>
      </div>

      <!-- Bridge Status Console -->
      <Card title="Go / Wails Bridge Status" bodyClass="p-3">
        {#snippet extra()}
          <Badge status="processing" text={appStatus} />
        {/snippet}
        <div class="bg-ant-bg rounded p-3 border border-ant-border-secondary font-mono text-xs text-ant-text-secondary flex items-center justify-between">
          <span>{pingResult || 'Initializing Wails IPC bridge connection...'}</span>
          <Button size="small" type="dashed" onclick={testBridge}>Refresh</Button>
        </div>
      </Card>
    </main>
  </div>
</div>
