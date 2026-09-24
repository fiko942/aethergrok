<script lang="ts">
  import { onMount } from 'svelte';
  import {
    settingsStore,
    type ThemeMode,
    type DefaultModel,
    type ReasoningEffort,
    type PermissionMode
  } from '$lib/stores/settings.svelte';
  import {
    Settings,
    Sliders,
    Brain,
    Shield,
    Palette,
    Keyboard,
    Camera,
    Check,
    X,
    RotateCcw,
    Sparkles,
    CheckCircle2,
    Volume2,
    Zap,
    Image as ImageIcon
  } from 'lucide-svelte';
  import Button from '$lib/antd/Button.svelte';
  import Card from '$lib/antd/Card.svelte';
  import Switch from '$lib/antd/Switch.svelte';
  import Badge from '$lib/antd/Badge.svelte';

  let {
    visible = false,
    onClose = () => {}
  }: {
    visible: boolean;
    onClose: () => void;
  } = $props();

  type TabKey = 'general' | 'models' | 'permissions' | 'theme' | 'shortcuts';
  let activeTab = $state<TabKey>('general');

  // Local draft state for edits
  let editGrokBinaryPath = $state(settingsStore.grokBinaryPath);
  let editSnapshotDelayMs = $state(settingsStore.snapshotDelayMs);
  let editSnapshotSoundEnabled = $state(settingsStore.snapshotSoundEnabled);
  let editSnapshotFlashEnabled = $state(settingsStore.snapshotFlashEnabled);
  let editSnapshotAutoAttach = $state(settingsStore.snapshotAutoAttach);
  let editActiveWindowTurnCount = $state(settingsStore.activeWindowTurnCount);
  let editDefaultModel = $state<DefaultModel>(settingsStore.defaultModel);
  let editCustomModelName = $state('');
  let editDefaultReasoningEffort = $state<ReasoningEffort>(settingsStore.defaultReasoningEffort);
  let editPermissionMode = $state<PermissionMode>(settingsStore.permissionMode);
  let editTheme = $state<ThemeMode>(settingsStore.theme);

  let saveSuccessNotice = $state(false);

  $effect(() => {
    if (visible) {
      editGrokBinaryPath = settingsStore.grokBinaryPath;
      editSnapshotDelayMs = settingsStore.snapshotDelayMs;
      editSnapshotSoundEnabled = settingsStore.snapshotSoundEnabled;
      editSnapshotFlashEnabled = settingsStore.snapshotFlashEnabled;
      editSnapshotAutoAttach = settingsStore.snapshotAutoAttach;
      editActiveWindowTurnCount = settingsStore.activeWindowTurnCount;
      editDefaultModel = settingsStore.defaultModel;
      if (settingsStore.defaultModel !== 'grok-4.6' && settingsStore.defaultModel !== 'grok-code') {
        editCustomModelName = settingsStore.defaultModel;
      }
      editDefaultReasoningEffort = settingsStore.defaultReasoningEffort;
      editPermissionMode = settingsStore.permissionMode;
      editTheme = settingsStore.theme;
      saveSuccessNotice = false;
    }
  });

  const tabs: Array<{ id: TabKey; label: string; icon: typeof Settings; description: string }> = [
    { id: 'general', label: 'General & Snapshot', icon: Sliders, description: 'Engine path, smart screenshot audio/flash, and DOM turn windowing' },
    { id: 'models', label: 'Models & Reasoning', icon: Brain, description: 'Default inference model and reasoning token budget' },
    { id: 'permissions', label: 'Permissions', icon: Shield, description: 'Security boundaries for filesystem, bash, and tool execution' },
    { id: 'theme', label: 'Theme & Appearance', icon: Palette, description: 'High-contrast, Ant Design light, and dark studio styles' },
    { id: 'shortcuts', label: 'Shortcuts', icon: Keyboard, description: 'Quick access keyboard bindings and interaction triggers' }
  ];

  const permissionDescriptions: Record<PermissionMode, {
    title: string;
    badge: 'safe' | 'balanced' | 'autonomous' | 'unrestricted';
    color: string;
    summary: string;
    details: string[];
    risk: 'low' | 'medium' | 'high' | 'critical';
  }> = {
    default: {
      title: 'Standard Guarded (Default)',
      badge: 'safe',
      color: 'text-ant-primary',
      summary: 'Requires explicit confirmation before executing terminal commands or modifying files outside tracked paths.',
      details: [
        'Prompts user with a diff modal before writing files',
        'Prompts user with a command modal before running bash commands',
        'Recommended for daily coding and untrusted tool invocation'
      ],
      risk: 'low'
    },
    acceptEdits: {
      title: 'Accept File Edits Automatically',
      badge: 'balanced',
      color: 'text-ant-success',
      summary: 'Auto-approves non-destructive file edits while continuing to prompt for bash/terminal execution.',
      details: [
        'Streamlined workflow for rapid code generation and refactoring',
        'All file diffs remain visible in the chat feed',
        'Terminal commands still require explicit single-click approval'
      ],
      risk: 'medium'
    },
    auto: {
      title: 'Autonomous Session Mode',
      badge: 'autonomous',
      color: 'text-ant-warning',
      summary: 'Auto-resolves read and workspace write operations within the active project directory.',
      details: [
        'Zero interruption for read_file, search_replace, and safe scripts',
        'Dangerous operations (e.g. rm -rf, network calls) are flagged',
        'Ideal for automated debugging loops and multi-step plan execution'
      ],
      risk: 'medium'
    },
    plan: {
      title: 'Plan-Only / Dry Run',
      badge: 'safe',
      color: 'text-ant-info',
      summary: 'Blocks all mutation tools. Grok will generate execution plans and inspections without applying changes.',
      details: [
        'All write and command tools operate in read-only inspection mode',
        'Guaranteed zero filesystem mutations',
        'Ideal for architectural reviews and risk assessments'
      ],
      risk: 'low'
    },
    bypassPermissions: {
      title: 'Bypass All Approvals (Unrestricted)',
      badge: 'unrestricted',
      color: 'text-ant-error',
      summary: 'Disables all permission confirmations. Grok executes all tools immediately with full host privileges.',
      details: [
        'No approval dialogs or pauses during execution',
        'Direct bash execution with the user identity and environment permissions',
        'Strictly intended for isolated containers or fully sandboxed environments'
      ],
      risk: 'critical'
    }
  };

  const keyboardShortcuts = [
    { key: '⌘ / Ctrl + Enter', action: 'Send message / Submit turn in composer', scope: 'Composer' },
    { key: 'Esc', action: 'Dismiss modal / Clear active overlay dialog', scope: 'Global' },
    { key: '⌘ / Ctrl + K', action: 'Open Skills & MCP discovery catalog', scope: 'Global' },
    { key: '⌘ / Ctrl + ,', action: 'Open Settings modal', scope: 'Global' },
    { key: '⌘ / Ctrl + T', action: 'Create a new conversation session tab', scope: 'Tabs' },
    { key: '⌘ / Ctrl + W', action: 'Close current session tab', scope: 'Tabs' },
    { key: '⌘ / Ctrl + Shift + S', action: 'Trigger instantaneous smart screen snapshot', scope: 'Screen' },
    { key: '⌘ / Ctrl + R', action: 'Re-run or retry last agent turn', scope: 'Conversation' }
  ];

  function handleThemeChange(t: ThemeMode) {
    editTheme = t;
    settingsStore.updateSettings({ theme: t });
  }

  function handleSaveAll() {
    const finalModel = editDefaultModel === 'custom'
      ? (editCustomModelName.trim() || 'grok-4.6')
      : editDefaultModel;

    settingsStore.updateSettings({
      grokBinaryPath: editGrokBinaryPath.trim(),
      snapshotDelayMs: Math.max(10, Math.min(2000, Number(editSnapshotDelayMs) || 50)),
      snapshotSoundEnabled: editSnapshotSoundEnabled,
      snapshotFlashEnabled: editSnapshotFlashEnabled,
      snapshotAutoAttach: editSnapshotAutoAttach,
      activeWindowTurnCount: Math.max(2, Math.min(50, Number(editActiveWindowTurnCount) || 10)),
      defaultModel: finalModel,
      defaultReasoningEffort: editDefaultReasoningEffort,
      permissionMode: editPermissionMode,
      theme: editTheme
    });

    saveSuccessNotice = true;
    setTimeout(() => {
      saveSuccessNotice = false;
      onClose();
    }, 400);
  }

  function handleResetDefaults() {
    settingsStore.resetToDefaults();
    editGrokBinaryPath = settingsStore.grokBinaryPath;
    editSnapshotDelayMs = settingsStore.snapshotDelayMs;
    editSnapshotSoundEnabled = settingsStore.snapshotSoundEnabled;
    editSnapshotFlashEnabled = settingsStore.snapshotFlashEnabled;
    editSnapshotAutoAttach = settingsStore.snapshotAutoAttach;
    editActiveWindowTurnCount = settingsStore.activeWindowTurnCount;
    editDefaultModel = settingsStore.defaultModel;
    editCustomModelName = '';
    editDefaultReasoningEffort = settingsStore.defaultReasoningEffort;
    editPermissionMode = settingsStore.permissionMode;
    editTheme = settingsStore.theme;
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape' && visible) {
      onClose();
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  });
</script>

{#if visible}
  <!-- Modal Backdrop with blur -->
  <div
    class="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 select-none animate-in fade-in duration-150"
    role="dialog"
    aria-modal="true"
    aria-labelledby="settings-modal-title"
  >
    <!-- Modal Container -->
    <div
      class="w-full max-w-4xl h-[680px] max-h-[90vh] flex flex-col bg-ant-bg-secondary border border-ant-border rounded-2xl shadow-2xl overflow-hidden"
    >
      <!-- Header -->
      <div class="px-6 py-4 border-b border-ant-border flex items-center justify-between bg-ant-bg-secondary flex-shrink-0">
        <div class="flex items-center space-x-3">
          <div class="w-8 h-8 rounded-lg bg-ant-primary/10 border border-ant-primary/30 flex items-center justify-center text-ant-primary shadow-sm">
            <Settings size={18} />
          </div>
          <div>
            <h2 id="settings-modal-title" class="text-sm font-bold text-white tracking-tight flex items-center gap-2">
              Settings & Workspace Preferences
              <span class="px-2 py-0.5 text-[10px] font-semibold bg-ant-primary/15 text-ant-primary rounded-full border border-ant-primary/30">
                Ant Design
              </span>
            </h2>
            <p class="text-xs text-ant-text-secondary mt-0.5">
              Configure local Grok runtime execution, smart screen snapshot, and theme appearance.
            </p>
          </div>
        </div>

        <div class="flex items-center space-x-2">
          {#if saveSuccessNotice}
            <span class="inline-flex items-center text-xs text-ant-success font-medium animate-pulse">
              <Check size={14} class="mr-1" /> Saved
            </span>
          {/if}
          <button
            type="button"
            class="p-1.5 rounded-lg text-ant-text-secondary hover:text-white hover:bg-ant-bg-tertiary transition"
            onclick={onClose}
            aria-label="Close settings"
          >
            <X size={16} />
          </button>
        </div>
      </div>

      <!-- Main Body: Sidebar + Content -->
      <div class="flex flex-1 min-h-0 overflow-hidden">
        <!-- Sidebar Navigation Tabs -->
        <nav class="w-56 bg-ant-bg border-r border-ant-border flex flex-col p-2 space-y-1 flex-shrink-0" aria-label="Settings navigation">
          {#each tabs as tab}
            {@const IconComponent = tab.icon}
            <button
              type="button"
              class="w-full flex items-center space-x-2.5 px-3 py-2 rounded-lg text-xs font-medium text-left transition {activeTab === tab.id
                ? 'bg-ant-primary/15 text-ant-primary border border-ant-primary/30'
                : 'text-ant-text-secondary hover:text-white hover:bg-ant-bg-tertiary/50 border border-transparent'}"
              onclick={() => activeTab = tab.id}
            >
              <IconComponent size={15} class="flex-shrink-0 {activeTab === tab.id ? 'text-ant-primary' : 'text-ant-text-secondary'}" />
              <span class="truncate">{tab.label}</span>
            </button>
          {/each}

          <div class="mt-auto pt-4 border-t border-ant-border/50 px-2 pb-2">
            <div class="text-[11px] text-ant-text-secondary font-mono leading-tight">
              AetherGrok Studio
              <div class="text-[10px] opacity-60">Build 1.0.0 (Wails/Go)</div>
            </div>
          </div>
        </nav>

        <!-- Tab Content Viewport -->
        <div class="flex-1 min-h-0 overflow-y-auto p-6 space-y-6 custom-scrollbar bg-ant-bg-secondary/60">
          
          <!-- TAB 1: GENERAL & SNAPSHOT -->
          {#if activeTab === 'general'}
            <div class="space-y-6 animate-in fade-in duration-100">
              <div>
                <h3 class="text-sm font-semibold text-white">General & Smart Screen Snapshot</h3>
                <p class="text-xs text-ant-text-secondary mt-0.5">
                  Configure CLI runner location, native snapshot behavior on macOS & Windows, and DOM memory bounds.
                </p>
              </div>

              <!-- Grok Binary Path Card -->
              <Card>
                <div class="space-y-3">
                  <div class="flex items-center justify-between">
                    <div>
                      <div class="text-xs font-semibold text-white">Grok CLI Executable Path</div>
                      <div class="text-[11px] text-ant-text-secondary">
                        Path to the native <code class="text-ant-primary font-mono bg-ant-bg px-1 py-0.5 rounded">grok</code> binary.
                      </div>
                    </div>
                    <button
                      type="button"
                      class="text-[11px] text-ant-primary hover:underline flex items-center gap-1"
                      onclick={() => editGrokBinaryPath = '/Users/fiko942/.local/bin/grok'}
                    >
                      <RotateCcw size={12} /> Auto-Detect
                    </button>
                  </div>
                  <input
                    type="text"
                    bind:value={editGrokBinaryPath}
                    placeholder="/Users/fiko942/.local/bin/grok"
                    class="w-full px-3 py-2 text-xs font-mono bg-ant-bg border border-ant-border rounded-lg text-white focus:outline-none focus:border-ant-primary focus:ring-1 focus:ring-ant-primary transition"
                  />
                </div>
              </Card>

              <!-- Smart Screen Snapshot Settings Card -->
              <Card>
                <div class="space-y-4">
                  <div class="flex items-center justify-between border-b border-ant-border/50 pb-2">
                    <div class="flex items-center gap-2">
                      <Camera size={16} class="text-ant-primary" />
                      <div class="text-xs font-semibold text-white">Non-Intrusive Smart Snapshot (macOS & Windows)</div>
                    </div>
                    <Badge status="success">Active Native Grab</Badge>
                  </div>

                  <!-- Audio Shutter Sound Toggle -->
                  <div class="flex items-center justify-between pt-1">
                    <div class="space-y-0.5">
                      <div class="text-xs font-medium text-white flex items-center gap-1.5">
                        <Volume2 size={13} class="text-ant-primary" />
                        Camera Shutter Audio ("Cekrek" Sound)
                      </div>
                      <div class="text-[11px] text-ant-text-secondary">
                        Synthesizes a realistic dual-stage mechanical shutter click on capture via Web Audio API.
                      </div>
                    </div>
                    <Switch
                      bind:checked={editSnapshotSoundEnabled}
                      size="small"
                    />
                  </div>

                  <!-- Screen Flash FX Toggle -->
                  <div class="flex items-center justify-between pt-2 border-t border-ant-border/40">
                    <div class="space-y-0.5">
                      <div class="text-xs font-medium text-white flex items-center gap-1.5">
                        <Zap size={13} class="text-ant-warning" />
                        Screen White Flash Animation
                      </div>
                      <div class="text-[11px] text-ant-text-secondary">
                        Displays an instantaneous white flash overlay when screenshot completes.
                      </div>
                    </div>
                    <Switch
                      bind:checked={editSnapshotFlashEnabled}
                      size="small"
                    />
                  </div>

                  <!-- Auto-Attach to Composer Toggle -->
                  <div class="flex items-center justify-between pt-2 border-t border-ant-border/40">
                    <div class="space-y-0.5">
                      <div class="text-xs font-medium text-white flex items-center gap-1.5">
                        <ImageIcon size={13} class="text-ant-success" />
                        Auto-Attach to Prompt Composer
                      </div>
                      <div class="text-[11px] text-ant-text-secondary">
                        Automatically embeds the captured screen thumbnail as a vision context chip.
                      </div>
                    </div>
                    <Switch
                      bind:checked={editSnapshotAutoAttach}
                      size="small"
                    />
                  </div>

                  <!-- Compositor Delay Slider -->
                  <div class="space-y-2 pt-2 border-t border-ant-border/40">
                    <div class="flex items-center justify-between">
                      <div>
                        <div class="text-xs font-medium text-white">OS Compositor Flush Delay</div>
                        <div class="text-[11px] text-ant-text-secondary">
                          Delay between window hide and screen capture to prevent window ghost frames.
                        </div>
                      </div>
                      <span class="text-xs font-mono font-bold text-ant-primary bg-ant-primary/10 px-2 py-0.5 rounded border border-ant-primary/30">
                        {editSnapshotDelayMs} ms
                      </span>
                    </div>
                    <input
                      type="range"
                      min="10"
                      max="500"
                      step="10"
                      bind:value={editSnapshotDelayMs}
                      class="w-full accent-ant-primary cursor-pointer h-1.5 bg-ant-bg rounded-lg"
                    />
                    <div class="flex justify-between text-[10px] text-ant-text-secondary font-mono">
                      <span>10ms (Fastest)</span>
                      <span>50ms (Recommended Mac)</span>
                      <span>80ms (Recommended Win)</span>
                      <span>500ms (Safe)</span>
                    </div>
                  </div>
                </div>
              </Card>

              <!-- DOM Windowing Turn Bounds -->
              <Card>
                <div class="space-y-3">
                  <div class="flex items-center justify-between">
                    <div>
                      <div class="text-xs font-semibold text-white">DOM Active Turn Windowing (RAM Guard)</div>
                      <div class="text-[11px] text-ant-text-secondary">
                        Limits active DOM messages to preserve low memory usage (&lt;60MB RAM).
                      </div>
                    </div>
                    <span class="text-xs font-mono font-bold text-ant-primary bg-ant-primary/10 px-2 py-0.5 rounded border border-ant-primary/30">
                      {editActiveWindowTurnCount} turns
                    </span>
                  </div>
                  <input
                    type="range"
                    min="4"
                    max="30"
                    step="2"
                    bind:value={editActiveWindowTurnCount}
                    class="w-full accent-ant-primary cursor-pointer h-1.5 bg-ant-bg rounded-lg"
                  />
                  <div class="flex justify-between text-[10px] text-ant-text-secondary font-mono">
                    <span>4 turns (Ultra-Low Spec)</span>
                    <span>10 turns (Default Optimal)</span>
                    <span>30 turns (Max Viewport)</span>
                  </div>
                </div>
              </Card>
            </div>
          {/if}

          <!-- TAB 2: MODELS & REASONING -->
          {#if activeTab === 'models'}
            <div class="space-y-6 animate-in fade-in duration-100">
              <div>
                <h3 class="text-sm font-semibold text-white">Inference Models & Reasoning Depth</h3>
                <p class="text-xs text-ant-text-secondary mt-0.5">
                  Select default model engine and reasoning effort stops for complex tasks.
                </p>
              </div>

              <!-- Model Selector Cards -->
              <div class="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  class="p-4 rounded-xl border text-left transition {editDefaultModel === 'grok-4.6'
                    ? 'border-ant-primary bg-ant-primary/10 shadow-sm'
                    : 'border-ant-border bg-ant-bg hover:border-ant-primary/50'}"
                  onclick={() => editDefaultModel = 'grok-4.6'}
                >
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-xs font-bold text-white flex items-center gap-1.5">
                      <Sparkles size={14} class="text-ant-primary" /> Grok 4.6
                    </span>
                    {#if editDefaultModel === 'grok-4.6'}
                      <CheckCircle2 size={15} class="text-ant-primary" />
                    {/if}
                  </div>
                  <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                    Flagship reasoning engine with deep multi-step analysis, complex logic, and comprehensive system architecture understanding.
                  </p>
                </button>

                <button
                  type="button"
                  class="p-4 rounded-xl border text-left transition {editDefaultModel === 'grok-code'
                    ? 'border-ant-primary bg-ant-primary/10 shadow-sm'
                    : 'border-ant-border bg-ant-bg hover:border-ant-primary/50'}"
                  onclick={() => editDefaultModel = 'grok-code'}
                >
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-xs font-bold text-white flex items-center gap-1.5">
                      <Brain size={14} class="text-ant-primary" /> Grok Code
                    </span>
                    {#if editDefaultModel === 'grok-code'}
                      <CheckCircle2 size={15} class="text-ant-primary" />
                    {/if}
                  </div>
                  <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                    High-throughput code synthesis engine optimized for refactoring, test-driven development, and git diff generation.
                  </p>
                </button>
              </div>

              <!-- Reasoning Effort Levels -->
              <Card>
                <div class="space-y-3">
                  <div>
                    <div class="text-xs font-semibold text-white">Default Reasoning Effort</div>
                    <div class="text-[11px] text-ant-text-secondary">
                      Controls internal thinking token budget before emitting tool and assistant actions.
                    </div>
                  </div>

                  <div class="grid grid-cols-5 gap-2">
                    {#each ['none', 'low', 'medium', 'high', 'max'] as effort}
                      <button
                        type="button"
                        class="px-3 py-2 rounded-lg text-xs font-medium border capitalize transition {editDefaultReasoningEffort === effort
                          ? 'bg-ant-primary text-white border-ant-primary shadow-sm'
                          : 'bg-ant-bg text-ant-text-secondary border-ant-border hover:text-white hover:bg-ant-bg-tertiary'}"
                        onclick={() => editDefaultReasoningEffort = effort as ReasoningEffort}
                      >
                        {effort}
                      </button>
                    {/each}
                  </div>
                </div>
              </Card>
            </div>
          {/if}

          <!-- TAB 3: PERMISSIONS -->
          {#if activeTab === 'permissions'}
            <div class="space-y-6 animate-in fade-in duration-100">
              <div>
                <h3 class="text-sm font-semibold text-white">Security & Execution Boundaries</h3>
                <p class="text-xs text-ant-text-secondary mt-0.5">
                  Control how AetherGrok asks for confirmation before executing bash commands or modifying project files.
                </p>
              </div>

              <div class="space-y-3">
                {#each (Object.keys(permissionDescriptions) as Array<PermissionMode>) as modeKey}
                  {@const mode = permissionDescriptions[modeKey]}
                  <button
                    type="button"
                    class="w-full p-4 rounded-xl border text-left transition {editPermissionMode === modeKey
                      ? 'border-ant-primary bg-ant-primary/10 shadow-sm'
                      : 'border-ant-border bg-ant-bg hover:border-ant-primary/40'}"
                    onclick={() => editPermissionMode = modeKey}
                  >
                    <div class="flex items-center justify-between mb-2">
                      <div class="flex items-center space-x-2">
                        <span class="text-xs font-bold text-white">{mode.title}</span>
                        <Badge
                          status={mode.badge === 'safe' ? 'success' : mode.badge === 'balanced' ? 'processing' : mode.badge === 'autonomous' ? 'warning' : 'error'}
                        >
                          {mode.risk.toUpperCase()} RISK
                        </Badge>
                      </div>
                      {#if editPermissionMode === modeKey}
                        <CheckCircle2 size={16} class="text-ant-primary" />
                      {/if}
                    </div>

                    <p class="text-[11px] text-ant-text-secondary mb-2">
                      {mode.summary}
                    </p>

                    <ul class="space-y-1 pl-4 border-l-2 border-ant-border/60 text-[10px] text-ant-text-secondary">
                      {#each mode.details as detail}
                        <li>• {detail}</li>
                      {/each}
                    </ul>
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- TAB 4: THEME & APPEARANCE -->
          {#if activeTab === 'theme'}
            <div class="space-y-6 animate-in fade-in duration-100">
              <div>
                <h3 class="text-sm font-semibold text-white">Theme & Color Customization</h3>
                <p class="text-xs text-ant-text-secondary mt-0.5">
                  Choose from calibrated Ant Design themes with high contrast legibility.
                </p>
              </div>

              <div class="grid grid-cols-3 gap-3">
                <!-- Dark Studio -->
                <button
                  type="button"
                  class="p-4 rounded-xl border text-left transition flex flex-col space-y-3 {editTheme === 'dark-studio'
                    ? 'border-ant-primary ring-2 ring-ant-primary/20 bg-ant-bg-tertiary'
                    : 'border-ant-border bg-ant-bg hover:border-ant-primary/40'}"
                  onclick={() => handleThemeChange('dark-studio')}
                >
                  <div class="flex items-center justify-between">
                    <span class="text-xs font-bold text-white">Dark Studio</span>
                    {#if editTheme === 'dark-studio'}
                      <CheckCircle2 size={15} class="text-ant-primary" />
                    {/if}
                  </div>

                  <!-- Theme Swatches -->
                  <div class="h-16 rounded-lg bg-[#0F1117] border border-[#232738] p-2 flex flex-col justify-between">
                    <div class="flex space-x-1.5">
                      <div class="w-3 h-3 rounded-full bg-[#1677FF]"></div>
                      <div class="w-3 h-3 rounded-full bg-[#00F0FF]"></div>
                      <div class="w-3 h-3 rounded-full bg-[#52C41A]"></div>
                    </div>
                    <div class="h-2 w-16 bg-[#181B26] rounded"></div>
                  </div>
                  <p class="text-[10px] text-ant-text-secondary">
                    Classic deep slate canvas (#0F1117) with vibrant neon-cyan & AntD blue accents.
                  </p>
                </button>

                <!-- Dark High Contrast -->
                <button
                  type="button"
                  class="p-4 rounded-xl border text-left transition flex flex-col space-y-3 {editTheme === 'dark-high-contrast'
                    ? 'border-ant-primary ring-2 ring-ant-primary/20 bg-ant-bg-tertiary'
                    : 'border-ant-border bg-ant-bg hover:border-ant-primary/40'}"
                  onclick={() => handleThemeChange('dark-high-contrast')}
                >
                  <div class="flex items-center justify-between">
                    <span class="text-xs font-bold text-white">High Contrast</span>
                    {#if editTheme === 'dark-high-contrast'}
                      <CheckCircle2 size={15} class="text-ant-primary" />
                    {/if}
                  </div>

                  <!-- Theme Swatches -->
                  <div class="h-16 rounded-lg bg-[#000000] border border-[#333333] p-2 flex flex-col justify-between">
                    <div class="flex space-x-1.5">
                      <div class="w-3 h-3 rounded-full bg-[#3B82F6]"></div>
                      <div class="w-3 h-3 rounded-full bg-[#FFFFFF]"></div>
                      <div class="w-3 h-3 rounded-full bg-[#22C55E]"></div>
                    </div>
                    <div class="h-2 w-16 bg-[#111111] rounded"></div>
                  </div>
                  <p class="text-[10px] text-ant-text-secondary">
                    Pure black (#000000) OLED canvas for maximum contrast and readability.
                  </p>
                </button>

                <!-- Light AntD -->
                <button
                  type="button"
                  class="p-4 rounded-xl border text-left transition flex flex-col space-y-3 {editTheme === 'light-antd'
                    ? 'border-ant-primary ring-2 ring-ant-primary/20 bg-ant-bg-tertiary'
                    : 'border-ant-border bg-ant-bg hover:border-ant-primary/40'}"
                  onclick={() => handleThemeChange('light-antd')}
                >
                  <div class="flex items-center justify-between">
                    <span class="text-xs font-bold text-white">Light AntD</span>
                    {#if editTheme === 'light-antd'}
                      <CheckCircle2 size={15} class="text-ant-primary" />
                    {/if}
                  </div>

                  <!-- Theme Swatches -->
                  <div class="h-16 rounded-lg bg-[#F5F5F5] border border-[#D9D9D9] p-2 flex flex-col justify-between">
                    <div class="flex space-x-1.5">
                      <div class="w-3 h-3 rounded-full bg-[#1677FF]"></div>
                      <div class="w-3 h-3 rounded-full bg-[#1890FF]"></div>
                      <div class="w-3 h-3 rounded-full bg-[#52C41A]"></div>
                    </div>
                    <div class="h-2 w-16 bg-[#FFFFFF] rounded"></div>
                  </div>
                  <p class="text-[10px] text-ant-text-secondary">
                    Clean official Ant Design light palette for daytime programming.
                  </p>
                </button>
              </div>
            </div>
          {/if}

          <!-- TAB 5: KEYBOARD SHORTCUTS -->
          {#if activeTab === 'shortcuts'}
            <div class="space-y-6 animate-in fade-in duration-100">
              <div>
                <h3 class="text-sm font-semibold text-white">Keyboard Shortcuts Reference</h3>
                <p class="text-xs text-ant-text-secondary mt-0.5">
                  Accelerate your workflow with dedicated keystroke accelerators.
                </p>
              </div>

              <div class="border border-ant-border rounded-xl overflow-hidden bg-ant-bg">
                <table class="w-full text-left border-collapse text-xs">
                  <thead>
                    <tr class="border-b border-ant-border bg-ant-bg-tertiary/40">
                      <th class="px-4 py-2.5 font-semibold text-white">Key Combination</th>
                      <th class="px-4 py-2.5 font-semibold text-white">Action Trigger</th>
                      <th class="px-4 py-2.5 font-semibold text-ant-text-secondary">Scope</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-ant-border/50">
                    {#each keyboardShortcuts as shortcut}
                      <tr class="hover:bg-ant-bg-secondary/40 transition">
                        <td class="px-4 py-2.5 font-mono text-ant-primary font-medium">{shortcut.key}</td>
                        <td class="px-4 py-2.5 text-white">{shortcut.action}</td>
                        <td class="px-4 py-2.5 text-ant-text-secondary">
                          <span class="px-2 py-0.5 text-[10px] font-medium bg-ant-bg-tertiary rounded border border-ant-border">
                            {shortcut.scope}
                          </span>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>
            </div>
          {/if}

        </div>
      </div>

      <!-- Footer Action Toolbar -->
      <div class="px-6 py-3 border-t border-ant-border flex items-center justify-between bg-ant-bg-secondary flex-shrink-0">
        <button
          type="button"
          class="text-xs text-ant-text-secondary hover:text-white flex items-center gap-1.5 transition"
          onclick={handleResetDefaults}
        >
          <RotateCcw size={13} /> Reset to Defaults
        </button>

        <div class="flex items-center space-x-2">
          <Button size="small" type="default" onclick={onClose}>
            Cancel
          </Button>
          <Button size="small" type="primary" onclick={handleSaveAll}>
            <Check size={14} class="mr-1" /> Save & Apply
          </Button>
        </div>
      </div>
    </div>
  </div>
{/if}
