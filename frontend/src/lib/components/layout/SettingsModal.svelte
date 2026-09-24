<script lang="ts">
  import { onMount } from 'svelte';
  import Button from '$lib/antd/Button.svelte';
  import Badge from '$lib/antd/Badge.svelte';
  import Switch from '$lib/antd/Switch.svelte';
  import {
    settingsStore,
    type ThemeMode,
    type ReasoningEffort,
    type PermissionMode
  } from '$lib/stores/settings.svelte';
  import { themeDefinitions } from '$lib/antd/tokens';
  import {
    X,
    Settings,
    Sliders,
    Brain,
    Shield,
    Palette,
    Keyboard,
    FolderOpen,
    RotateCcw,
    Check,
    AlertTriangle,
    Lock,
    Eye,
    Zap,
    Cpu,
    Sparkles,
    Info,
    Terminal,
    Command
  } from 'lucide-svelte';

  interface Props {
    visible: boolean;
    onClose: () => void;
  }

  let { visible = false, onClose }: Props = $props();

  type TabKey = 'general' | 'models' | 'permissions' | 'theme' | 'shortcuts';
  let activeTab = $state<TabKey>('general');

  // Temporary edit state initialized from settingsStore
  let editGrokBinaryPath = $state(settingsStore.grokBinaryPath);
  let editSnapshotDelayMs = $state(settingsStore.snapshotDelayMs);
  let editActiveWindowTurnCount = $state(settingsStore.activeWindowTurnCount);
  let editDefaultModel = $state(settingsStore.defaultModel);
  let editCustomModelName = $state('');
  let editDefaultReasoningEffort = $state<ReasoningEffort>(settingsStore.defaultReasoningEffort);
  let editPermissionMode = $state<PermissionMode>(settingsStore.permissionMode);
  let editTheme = $state<ThemeMode>(settingsStore.theme);

  let saveSuccessNotice = $state(false);

  $effect(() => {
    if (visible) {
      editGrokBinaryPath = settingsStore.grokBinaryPath;
      editSnapshotDelayMs = settingsStore.snapshotDelayMs;
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
    { id: 'general', label: 'General', icon: Sliders, description: 'Engine path, snapshot timing, and DOM turn windowing' },
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
    { key: '⌘ / Ctrl + S', action: 'Trigger instantaneous screen snapshot', scope: 'Screen' },
    { key: '⌘ / Ctrl + R', action: 'Re-run or retry last agent turn', scope: 'Conversation' }
  ];

  function handleThemeChange(t: ThemeMode) {
    editTheme = t;
    // Real-time preview by updating the store dynamically
    settingsStore.updateSettings({ theme: t });
  }

  function handleSaveAll() {
    const finalModel = editDefaultModel === 'custom'
      ? (editCustomModelName.trim() || 'grok-4.6')
      : editDefaultModel;

    settingsStore.updateSettings({
      grokBinaryPath: editGrokBinaryPath.trim(),
      snapshotDelayMs: Math.max(10, Math.min(2000, Number(editSnapshotDelayMs) || 50)),
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
              Configure local Grok runtime execution, security constraints, and appearance.
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
              class="w-full flex items-center space-x-3 px-3 py-2.5 rounded-lg text-xs font-medium transition-all text-left group
                {activeTab === tab.id
                  ? 'bg-ant-primary text-white shadow-sm shadow-ant-primary/30'
                  : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-secondary'}"
              onclick={() => activeTab = tab.id}
            >
              <IconComponent
                size={15}
                class={activeTab === tab.id ? 'text-white' : 'text-ant-text-muted group-hover:text-ant-primary transition-colors'}
              />
              <span class="truncate">{tab.label}</span>
            </button>
          {/each}

          <div class="mt-auto pt-3 border-t border-ant-border-secondary p-2">
            <button
              type="button"
              class="w-full flex items-center justify-center space-x-1.5 px-3 py-2 text-xs text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-secondary rounded-lg border border-ant-border-secondary transition"
              onclick={handleResetDefaults}
            >
              <RotateCcw size={12} />
              <span>Reset Defaults</span>
            </button>
          </div>
        </nav>

        <!-- Tab Content Viewport -->
        <div class="flex-1 overflow-y-auto p-6 space-y-6 bg-ant-bg-secondary/40">
          <!-- TAB 1: GENERAL -->
          {#if activeTab === 'general'}
            <div class="space-y-6">
              <div>
                <h3 class="text-sm font-semibold text-white">General Engine Settings</h3>
                <p class="text-xs text-ant-text-secondary mt-1">Configure CLI executable path and performance latency settings.</p>
              </div>

              <!-- Binary Path Input -->
              <div class="p-4 rounded-xl bg-ant-bg border border-ant-border space-y-2">
                <label for="grok-binary-input" class="text-xs font-medium text-white flex items-center justify-between">
                  <span class="flex items-center gap-1.5">
                    <Terminal size={14} class="text-ant-primary" />
                    Grok CLI Executable Path
                  </span>
                  <span class="text-[10px] text-ant-text-muted font-mono">auto-detected</span>
                </label>
                <div class="flex items-center space-x-2">
                  <input
                    id="grok-binary-input"
                    type="text"
                    bind:value={editGrokBinaryPath}
                    class="flex-1 px-3 py-2 rounded-lg bg-ant-bg-secondary border border-ant-border text-xs text-white placeholder:text-ant-text-muted focus:border-ant-primary outline-none font-mono"
                    placeholder="/Users/fiko942/.local/bin/grok"
                  />
                  <Button
                    size="small"
                    type="default"
                    onclick={() => editGrokBinaryPath = '/Users/fiko942/.local/bin/grok'}
                  >
                    Default
                  </Button>
                </div>
                <p class="text-[11px] text-ant-text-muted">
                  AetherGrok invokes this CLI binary in headless JSON stream mode for prompt execution and tool calls.
                </p>
              </div>

              <!-- Snapshot Delay -->
              <div class="p-4 rounded-xl bg-ant-bg border border-ant-border space-y-3">
                <div class="flex items-center justify-between">
                  <div>
                    <label for="snapshot-delay-range" class="text-xs font-medium text-white block">
                      Screen Snapshot Delay
                    </label>
                    <p class="text-[11px] text-ant-text-muted mt-0.5">
                      Grace period (milliseconds) after hiding window before grabbing native screen frame.
                    </p>
                  </div>
                  <span class="px-2 py-0.5 rounded text-xs font-mono font-semibold bg-ant-bg-secondary text-ant-primary border border-ant-border">
                    {editSnapshotDelayMs} ms
                  </span>
                </div>
                <input
                  id="snapshot-delay-range"
                  type="range"
                  min="10"
                  max="500"
                  step="10"
                  bind:value={editSnapshotDelayMs}
                  class="w-full accent-ant-primary cursor-pointer"
                />
                <div class="flex justify-between text-[10px] text-ant-text-muted font-mono">
                  <span>10 ms (Ultra fast)</span>
                  <span>50 ms (Optimal default)</span>
                  <span>500 ms (Safe latency)</span>
                </div>
              </div>

              <!-- Active Window Turn Count -->
              <div class="p-4 rounded-xl bg-ant-bg border border-ant-border space-y-3">
                <div class="flex items-center justify-between">
                  <div>
                    <label for="window-turn-range" class="text-xs font-medium text-white block">
                      Active Window Turn Count
                    </label>
                    <p class="text-[11px] text-ant-text-muted mt-0.5">
                      Maximum number of recent conversation turns kept in active DOM to maintain &lt; 60MB RAM footprint.
                    </p>
                  </div>
                  <span class="px-2 py-0.5 rounded text-xs font-mono font-semibold bg-ant-bg-secondary text-ant-primary border border-ant-border">
                    {editActiveWindowTurnCount} turns
                  </span>
                </div>
                <input
                  id="window-turn-range"
                  type="range"
                  min="4"
                  max="30"
                  step="2"
                  bind:value={editActiveWindowTurnCount}
                  class="w-full accent-ant-primary cursor-pointer"
                />
                <div class="flex justify-between text-[10px] text-ant-text-muted font-mono">
                  <span>4 turns (Minimal RAM)</span>
                  <span>10 turns (Recommended)</span>
                  <span>30 turns (Long history)</span>
                </div>
              </div>
            </div>
          {/if}

          <!-- TAB 2: MODELS & REASONING -->
          {#if activeTab === 'models'}
            <div class="space-y-6">
              <div>
                <h3 class="text-sm font-semibold text-white">Default Model & Reasoning Effort</h3>
                <p class="text-xs text-ant-text-secondary mt-1">Select standard model endpoint and default thinking budget.</p>
              </div>

              <!-- Model Selector Cards -->
              <div class="space-y-3">
                <span class="text-xs font-medium text-white">Default Grok Model</span>
                <div class="grid grid-cols-3 gap-3">
                  <!-- grok-4.6 -->
                  <button
                    type="button"
                    class="p-3.5 rounded-xl border text-left transition-all flex flex-col justify-between
                      {editDefaultModel === 'grok-4.6'
                        ? 'bg-ant-primary/10 border-ant-primary ring-1 ring-ant-primary/50 text-white'
                        : 'bg-ant-bg border-ant-border text-ant-text-secondary hover:text-ant-text hover:border-ant-border-secondary'}"
                    onclick={() => editDefaultModel = 'grok-4.6'}
                  >
                    <div>
                      <div class="flex items-center justify-between">
                        <span class="font-semibold text-xs text-white">grok-4.6</span>
                        {#if editDefaultModel === 'grok-4.6'}
                          <Check size={14} class="text-ant-primary" />
                        {/if}
                      </div>
                      <p class="text-[11px] text-ant-text-muted mt-1 leading-snug">
                        Flagship general intelligence model with highest reasoning depth.
                      </p>
                    </div>
                    <span class="mt-3 text-[10px] px-1.5 py-0.5 rounded bg-ant-bg-tertiary text-ant-primary w-fit font-mono">
                      Recommended
                    </span>
                  </button>

                  <!-- grok-code -->
                  <button
                    type="button"
                    class="p-3.5 rounded-xl border text-left transition-all flex flex-col justify-between
                      {editDefaultModel === 'grok-code'
                        ? 'bg-ant-primary/10 border-ant-primary ring-1 ring-ant-primary/50 text-white'
                        : 'bg-ant-bg border-ant-border text-ant-text-secondary hover:text-ant-text hover:border-ant-border-secondary'}"
                    onclick={() => editDefaultModel = 'grok-code'}
                  >
                    <div>
                      <div class="flex items-center justify-between">
                        <span class="font-semibold text-xs text-white">grok-code</span>
                        {#if editDefaultModel === 'grok-code'}
                          <Check size={14} class="text-ant-primary" />
                        {/if}
                      </div>
                      <p class="text-[11px] text-ant-text-muted mt-1 leading-snug">
                        Optimized for fast code generation, tool calls, and syntax edits.
                      </p>
                    </div>
                    <span class="mt-3 text-[10px] px-1.5 py-0.5 rounded bg-ant-bg-tertiary text-ant-success w-fit font-mono">
                      Fast Edits
                    </span>
                  </button>

                  <!-- custom -->
                  <button
                    type="button"
                    class="p-3.5 rounded-xl border text-left transition-all flex flex-col justify-between
                      {editDefaultModel !== 'grok-4.6' && editDefaultModel !== 'grok-code'
                        ? 'bg-ant-primary/10 border-ant-primary ring-1 ring-ant-primary/50 text-white'
                        : 'bg-ant-bg border-ant-border text-ant-text-secondary hover:text-ant-text hover:border-ant-border-secondary'}"
                    onclick={() => {
                      editDefaultModel = 'custom';
                      if (!editCustomModelName) editCustomModelName = 'custom-model';
                    }}
                  >
                    <div>
                      <div class="flex items-center justify-between">
                        <span class="font-semibold text-xs text-white">Custom Model</span>
                        {#if editDefaultModel !== 'grok-4.6' && editDefaultModel !== 'grok-code'}
                          <Check size={14} class="text-ant-primary" />
                        {/if}
                      </div>
                      <p class="text-[11px] text-ant-text-muted mt-1 leading-snug">
                        Specify any custom Grok alias or internal fine-tuned endpoint.
                      </p>
                    </div>
                    <span class="mt-3 text-[10px] px-1.5 py-0.5 rounded bg-ant-bg-tertiary text-ant-text-muted w-fit font-mono">
                      Configurable
                    </span>
                  </button>
                </div>

                {#if editDefaultModel !== 'grok-4.6' && editDefaultModel !== 'grok-code'}
                  <div class="mt-3 p-3 rounded-lg bg-ant-bg border border-ant-border space-y-1.5 animate-in fade-in duration-100">
                    <label for="custom-model-input" class="text-xs text-ant-text-secondary block">Custom Model Identifier</label>
                    <input
                      id="custom-model-input"
                      type="text"
                      bind:value={editCustomModelName}
                      placeholder="e.g. grok-4-fast or internal-test-model"
                      class="w-full px-3 py-1.5 rounded-md bg-ant-bg-secondary border border-ant-border text-xs text-white placeholder:text-ant-text-muted focus:border-ant-primary outline-none font-mono"
                    />
                  </div>
                {/if}
              </div>

              <!-- Reasoning Effort Selector -->
              <div class="p-4 rounded-xl bg-ant-bg border border-ant-border space-y-3">
                <div class="flex items-center justify-between">
                  <span class="text-xs font-medium text-white flex items-center gap-1.5">
                    <Brain size={14} class="text-ant-primary" />
                    Default Reasoning Effort Level
                  </span>
                  <span class="px-2 py-0.5 rounded text-xs font-mono font-semibold uppercase bg-ant-primary/15 text-ant-primary border border-ant-primary/30">
                    {editDefaultReasoningEffort}
                  </span>
                </div>
                <div class="grid grid-cols-5 gap-2">
                  {#each (['none', 'low', 'medium', 'high', 'max'] as ReasoningEffort[]) as effort}
                    <button
                      type="button"
                      class="py-2 px-1 rounded-lg text-xs font-medium text-center transition-all capitalize
                        {editDefaultReasoningEffort === effort
                          ? 'bg-ant-primary text-white shadow-sm font-semibold'
                          : 'bg-ant-bg-secondary text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary border border-ant-border-secondary'}"
                      onclick={() => editDefaultReasoningEffort = effort}
                    >
                      {effort}
                    </button>
                  {/each}
                </div>
                <p class="text-[11px] text-ant-text-muted">
                  Controls internal reasoning token exploration budget. Higher effort levels produce richer step-by-step thinking for complex engineering tasks.
                </p>
              </div>
            </div>
          {/if}

          <!-- TAB 3: PERMISSIONS -->
          {#if activeTab === 'permissions'}
            <div class="space-y-4">
              <div>
                <h3 class="text-sm font-semibold text-white">Security & Permission Policies</h3>
                <p class="text-xs text-ant-text-secondary mt-1">Determine how tool invocations and system actions are authorized.</p>
              </div>

              <div class="space-y-3">
                {#each (Object.keys(permissionDescriptions) as PermissionMode[]) as mode}
                  {@const p = permissionDescriptions[mode]}
                  {@const isSelected = editPermissionMode === mode}
                  <button
                    type="button"
                    class="w-full p-4 rounded-xl border text-left transition-all relative block
                      {isSelected
                        ? 'bg-ant-primary/10 border-ant-primary ring-1 ring-ant-primary/40'
                        : 'bg-ant-bg border-ant-border hover:border-ant-border-secondary'}"
                    onclick={() => editPermissionMode = mode}
                  >
                    <div class="flex items-center justify-between">
                      <div class="flex items-center space-x-2">
                        <div class="w-4 h-4 rounded-full border flex items-center justify-center {isSelected ? 'border-ant-primary bg-ant-primary' : 'border-ant-border'}">
                          {#if isSelected}
                            <div class="w-1.5 h-1.5 bg-white rounded-full"></div>
                          {/if}
                        </div>
                        <span class="text-xs font-semibold {isSelected ? 'text-white' : 'text-ant-text'}">{p.title}</span>
                      </div>
                      <span class="text-[10px] font-mono px-2 py-0.5 rounded border uppercase
                        {p.risk === 'low' ? 'bg-ant-success/15 text-ant-success border-ant-success/30' :
                         p.risk === 'medium' ? 'bg-ant-warning/15 text-ant-warning border-ant-warning/30' :
                         'bg-ant-error/15 text-ant-error border-ant-error/30'}">
                        {p.risk} risk
                      </span>
                    </div>

                    <p class="text-xs text-ant-text-secondary mt-2 pl-6 leading-relaxed">
                      {p.summary}
                    </p>

                    <div class="mt-2.5 pl-6 space-y-1">
                      {#each p.details as detail}
                        <div class="flex items-center text-[11px] text-ant-text-muted">
                          <span class="w-1 h-1 rounded-full bg-ant-primary mr-2"></span>
                          <span>{detail}</span>
                        </div>
                      {/each}
                    </div>
                  </button>
                {/each}
              </div>
            </div>
          {/if}

          <!-- TAB 4: THEME & APPEARANCE -->
          {#if activeTab === 'theme'}
            <div class="space-y-6">
              <div>
                <h3 class="text-sm font-semibold text-white">Theme & Design System</h3>
                <p class="text-xs text-ant-text-secondary mt-1">Select aesthetic palette. Themes apply instantly across all Ant Design tokens.</p>
              </div>

              <!-- Theme Selection Cards with Live Previews -->
              <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                {#each (['dark-studio', 'dark-high-contrast', 'light-antd'] as ThemeMode[]) as tKey}
                  {@const tDef = themeDefinitions[tKey]}
                  {@const isSelected = editTheme === tKey}
                  <button
                    type="button"
                    class="p-4 rounded-xl border text-left transition-all flex flex-col justify-between group
                      {isSelected
                        ? 'border-ant-primary ring-2 ring-ant-primary/40 bg-ant-bg shadow-lg shadow-ant-primary/10'
                        : 'border-ant-border bg-ant-bg hover:border-ant-border-secondary'}"
                    onclick={() => handleThemeChange(tKey)}
                  >
                    <div>
                      <!-- Mini UI Mockup Preview -->
                      <div class="w-full h-24 rounded-lg overflow-hidden border border-black/20 mb-3 p-2 flex flex-col justify-between shadow-inner"
                        style="background-color: {tDef.colors.bgLayout}; color: {tDef.colors.text};"
                      >
                        <!-- Mock Header -->
                        <div class="flex items-center justify-between pb-1 border-b" style="border-color: {tDef.colors.borderSecondary}">
                          <div class="flex items-center space-x-1">
                            <div class="w-2 h-2 rounded-full" style="background-color: {tDef.colors.primary}"></div>
                            <span class="text-[9px] font-bold" style="color: {tDef.colors.text}">Grok</span>
                          </div>
                          <span class="text-[8px] px-1 rounded" style="background-color: {tDef.colors.primaryBg}; color: {tDef.colors.primary}">Active</span>
                        </div>

                        <!-- Mock Chat Bubble -->
                        <div class="p-1.5 rounded text-[8px] space-y-1" style="background-color: {tDef.colors.bgContainer}; border: 1px solid {tDef.colors.border}">
                          <div class="w-16 h-1 rounded" style="background-color: {tDef.colors.textSecondary}"></div>
                          <div class="w-24 h-1 rounded" style="background-color: {tDef.colors.primary}"></div>
                        </div>

                        <!-- Mock Color Swatches -->
                        <div class="flex items-center space-x-1 pt-1">
                          <div class="w-2.5 h-2.5 rounded-full" style="background-color: {tDef.colors.primary}"></div>
                          <div class="w-2.5 h-2.5 rounded-full" style="background-color: {tDef.colors.success}"></div>
                          <div class="w-2.5 h-2.5 rounded-full" style="background-color: {tDef.colors.warning}"></div>
                          <div class="w-2.5 h-2.5 rounded-full" style="background-color: {tDef.colors.error}"></div>
                        </div>
                      </div>

                      <div class="flex items-center justify-between">
                        <span class="text-xs font-bold text-white group-hover:text-ant-primary transition-colors">{tDef.name}</span>
                        {#if isSelected}
                          <Check size={14} class="text-ant-primary" />
                        {/if}
                      </div>
                      <p class="text-[11px] text-ant-text-muted mt-1 leading-normal">
                        {tDef.description}
                      </p>
                    </div>

                    <div class="mt-4 pt-2 border-t border-ant-border-secondary/50 flex items-center justify-between text-[10px]">
                      <span class="text-ant-text-muted font-mono">{tKey}</span>
                      <span class="font-medium {isSelected ? 'text-ant-primary' : 'text-ant-text-secondary'}">
                        {isSelected ? 'Active' : 'Select'}
                      </span>
                    </div>
                  </button>
                {/each}
              </div>

              <!-- Color Palette Reference -->
              <div class="p-4 rounded-xl bg-ant-bg border border-ant-border space-y-3">
                <span class="text-xs font-semibold text-white">Active Theme Swatch Palette</span>
                <div class="grid grid-cols-6 gap-2">
                  <div class="flex flex-col items-center p-2 rounded bg-ant-bg-secondary border border-ant-border-secondary">
                    <div class="w-5 h-5 rounded-full bg-ant-primary mb-1 border border-white/20"></div>
                    <span class="text-[9px] text-ant-text-muted font-mono">Primary</span>
                  </div>
                  <div class="flex flex-col items-center p-2 rounded bg-ant-bg-secondary border border-ant-border-secondary">
                    <div class="w-5 h-5 rounded-full bg-ant-success mb-1 border border-white/20"></div>
                    <span class="text-[9px] text-ant-text-muted font-mono">Success</span>
                  </div>
                  <div class="flex flex-col items-center p-2 rounded bg-ant-bg-secondary border border-ant-border-secondary">
                    <div class="w-5 h-5 rounded-full bg-ant-warning mb-1 border border-white/20"></div>
                    <span class="text-[9px] text-ant-text-muted font-mono">Warning</span>
                  </div>
                  <div class="flex flex-col items-center p-2 rounded bg-ant-bg-secondary border border-ant-border-secondary">
                    <div class="w-5 h-5 rounded-full bg-ant-error mb-1 border border-white/20"></div>
                    <span class="text-[9px] text-ant-text-muted font-mono">Error</span>
                  </div>
                  <div class="flex flex-col items-center p-2 rounded bg-ant-bg-secondary border border-ant-border-secondary">
                    <div class="w-5 h-5 rounded-full bg-ant-bg mb-1 border border-ant-border"></div>
                    <span class="text-[9px] text-ant-text-muted font-mono">Surface</span>
                  </div>
                  <div class="flex flex-col items-center p-2 rounded bg-ant-bg-secondary border border-ant-border-secondary">
                    <div class="w-5 h-5 rounded-full bg-ant-border mb-1 border border-white/20"></div>
                    <span class="text-[9px] text-ant-text-muted font-mono">Border</span>
                  </div>
                </div>
              </div>
            </div>
          {/if}

          <!-- TAB 5: KEYBOARD SHORTCUTS -->
          {#if activeTab === 'shortcuts'}
            <div class="space-y-4">
              <div>
                <h3 class="text-sm font-semibold text-white">Keyboard Shortcuts Reference</h3>
                <p class="text-xs text-ant-text-secondary mt-1">Accelerate workflow with global hotkeys and context-specific controls.</p>
              </div>

              <div class="rounded-xl border border-ant-border bg-ant-bg overflow-hidden shadow-sm">
                <table class="w-full text-left text-xs">
                  <thead class="bg-ant-bg-secondary border-b border-ant-border text-ant-text-secondary uppercase text-[10px] tracking-wider font-semibold select-none">
                    <tr>
                      <th class="py-2.5 px-4">Shortcut</th>
                      <th class="py-2.5 px-4">Action</th>
                      <th class="py-2.5 px-4 w-28">Scope</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-ant-border-secondary/60">
                    {#each keyboardShortcuts as shortcut}
                      <tr class="hover:bg-ant-bg-secondary/40 transition-colors">
                        <td class="py-2.5 px-4 font-mono font-medium text-white">
                          <kbd class="px-2 py-1 rounded bg-ant-bg-secondary border border-ant-border text-[11px] shadow-sm text-ant-primary font-mono font-semibold">
                            {shortcut.key}
                          </kbd>
                        </td>
                        <td class="py-2.5 px-4 text-ant-text font-normal">
                          {shortcut.action}
                        </td>
                        <td class="py-2.5 px-4">
                          <span class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-ant-bg-tertiary text-ant-text-muted border border-ant-border-secondary">
                            {shortcut.scope}
                          </span>
                        </td>
                      </tr>
                    {/each}
                  </tbody>
                </table>
              </div>

              <div class="p-3 bg-ant-bg rounded-lg border border-ant-border-secondary flex items-start space-x-2 text-[11px] text-ant-text-muted">
                <Info size={14} class="text-ant-primary flex-shrink-0 mt-0.5" />
                <span>Custom keyboard mapping adjustments can also be set within macOS System Settings &gt; Keyboard &gt; Keyboard Shortcuts.</span>
              </div>
            </div>
          {/if}
        </div>
      </div>

      <!-- Footer Action Bar -->
      <div class="px-6 py-3 border-t border-ant-border bg-ant-bg flex items-center justify-between flex-shrink-0">
        <div class="text-[11px] text-ant-text-muted font-mono flex items-center space-x-1.5">
          <span>Active theme:</span>
          <span class="text-ant-primary font-semibold">{settingsStore.theme}</span>
        </div>

        <div class="flex items-center space-x-2.5">
          <Button size="small" type="default" onclick={onClose}>
            Cancel
          </Button>
          <Button size="small" type="primary" onclick={handleSaveAll}>
            Save & Apply
          </Button>
        </div>
      </div>
    </div>
  </div>
{/if}
