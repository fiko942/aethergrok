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
    Image as ImageIcon,
    Info,
    Globe,
    Github,
    ExternalLink,
    Code2,
    Heart,
    FolderOpen,
    FileCode2,
    ShieldCheck,
    ZapOff
  } from 'lucide-svelte';
  import Button from '$lib/antd/Button.svelte';
  import Card from '$lib/antd/Card.svelte';
  import Switch from '$lib/antd/Switch.svelte';
  import Badge from '$lib/antd/Badge.svelte';
  import KeyRecorderModal from '$lib/components/ui/KeyRecorderModal.svelte';

  let {
    visible = false,
    onClose = () => {}
  }: {
    visible: boolean;
    onClose: () => void;
  } = $props();

  type TabKey = 'general' | 'models' | 'permissions' | 'theme' | 'shortcuts' | 'about';
  let activeTab = $state<TabKey>('general');

  // Local draft state for edits
  let editGrokBinaryPath = $state(settingsStore.grokBinaryPath);
  let editSnapshotShortcut = $state(settingsStore.snapshotShortcut);
  let editSnapshotDelayMs = $state(settingsStore.snapshotDelayMs);
  let editSnapshotSoundEnabled = $state(settingsStore.snapshotSoundEnabled);
  let editSnapshotFlashEnabled = $state(settingsStore.snapshotFlashEnabled);
  let editSnapshotAutoAttach = $state(settingsStore.snapshotAutoAttach);
  let editActiveWindowTurnCount = $state(settingsStore.activeWindowTurnCount);
  let editMaxContextTokens = $state(settingsStore.maxContextTokens || 200000);
  let editDefaultModel = $state<DefaultModel>(settingsStore.defaultModel);
  let editCustomModelName = $state('');
  let editDefaultReasoningEffort = $state<ReasoningEffort>(settingsStore.defaultReasoningEffort);
  let editPermissionMode = $state<PermissionMode>(settingsStore.permissionMode);
  let editPlanGateMode = $state<PlanGateMode>(settingsStore.planGateMode);
  let editAnimationsEnabled = $state<boolean>(settingsStore.animationsEnabled);
  let editTheme = $state<ThemeMode>(settingsStore.theme);

  let isRecordingShortcut = $state(false);
  let saveSuccessNotice = $state(false);

  function openExternal(url: string) {
    if (window.go?.main?.App?.OpenExternalURL) {
      window.go.main.App.OpenExternalURL(url);
    } else if (window.runtime?.BrowserOpenURL) {
      window.runtime.BrowserOpenURL(url);
    } else {
      window.open(url, '_blank');
    }
  }

  $effect(() => {
    if (visible) {
      editGrokBinaryPath = settingsStore.grokBinaryPath;
      editSnapshotShortcut = settingsStore.snapshotShortcut;
      editSnapshotDelayMs = settingsStore.snapshotDelayMs;
      editSnapshotSoundEnabled = settingsStore.snapshotSoundEnabled;
      editSnapshotFlashEnabled = settingsStore.snapshotFlashEnabled;
      editSnapshotAutoAttach = settingsStore.snapshotAutoAttach;
      editActiveWindowTurnCount = settingsStore.activeWindowTurnCount;
      editMaxContextTokens = settingsStore.maxContextTokens || 200000;
      editDefaultModel = settingsStore.defaultModel;
      if (settingsStore.defaultModel !== 'grok-4.6' && settingsStore.defaultModel !== 'grok-code') {
        editCustomModelName = settingsStore.defaultModel;
      }
      editDefaultReasoningEffort = settingsStore.defaultReasoningEffort;
      editPermissionMode = settingsStore.permissionMode;
      editPlanGateMode = settingsStore.planGateMode;
      editAnimationsEnabled = settingsStore.animationsEnabled;
      editTheme = settingsStore.theme;
      isRecordingShortcut = false;
      saveSuccessNotice = false;
    }
  });

  const tabs: Array<{ id: TabKey; label: string; icon: typeof Settings; description: string }> = [
    { id: 'general', label: 'General & Snapshot', icon: Sliders, description: 'Engine path, smart screenshot audio/flash, and DOM turn windowing' },
    { id: 'models', label: 'Models & Reasoning', icon: Brain, description: 'Default inference model and reasoning token budget' },
    { id: 'permissions', label: 'Permissions', icon: Shield, description: 'Security boundaries for filesystem, bash, and tool execution' },
    { id: 'theme', label: 'Theme & Appearance', icon: Palette, description: 'High-contrast, Ant Design light, and dark studio styles' },
    { id: 'shortcuts', label: 'Shortcuts', icon: Keyboard, description: 'Quick access keyboard bindings and interaction triggers' },
    { id: 'about', label: 'About AetherGrok', icon: Info, description: 'Mission, target audience, open-source repository, and developer portfolio' }
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

  const isMac = typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform || navigator.userAgent);

  interface ShortcutItem {
    keys: string[];
    action: string;
    scope: string;
  }

  const keyboardShortcuts: ShortcutItem[] = [
    { keys: [isMac ? '⌘' : 'Ctrl', 'Enter'], action: 'Send message / Submit turn in composer', scope: 'Composer' },
    { keys: ['Esc'], action: 'Dismiss modal / Clear active overlay dialog', scope: 'Global' },
    { keys: [isMac ? '⌘' : 'Ctrl', 'K'], action: 'Open Skills & MCP discovery catalog', scope: 'Global' },
    { keys: [isMac ? '⌘' : 'Ctrl', ','], action: 'Open Settings & Preferences modal', scope: 'Global' },
    { keys: [isMac ? '⌘' : 'Ctrl', 'T'], action: 'Create a new conversation session tab', scope: 'Tabs' },
    { keys: [isMac ? '⌘' : 'Ctrl', 'W'], action: 'Close current session tab', scope: 'Tabs' },
    { keys: [isMac ? '⌘' : 'Ctrl', 'B'], action: 'Toggle left sidebar collapse & expand', scope: 'Navigation' },
    { keys: [isMac ? '⌘' : 'Ctrl', '1-8'], action: 'Switch to session tab 1 through 8', scope: 'Tabs' },
    { keys: [isMac ? '⌘' : 'Ctrl', '9'], action: 'Switch to the last open session tab', scope: 'Tabs' },
    { keys: [isMac ? '⌘' : 'Ctrl', 'Shift', 'S'], action: 'Trigger instantaneous smart screen snapshot', scope: 'Screen' },
    { keys: [isMac ? '⌘' : 'Ctrl', 'R'], action: 'Re-run or retry last agent turn', scope: 'Conversation' }
  ];

  function handleThemeChange(t: ThemeMode) {
    editTheme = t;
    settingsStore.updateSettings({ theme: t });
  }

  function handleSaveAll() {
    const finalModel = editDefaultModel === 'custom'
      ? (editCustomModelName.trim() || '9router')
      : editDefaultModel;

    settingsStore.updateSettings({
      grokBinaryPath: editGrokBinaryPath.trim() || '/Users/fiko942/.local/bin/grok',
      snapshotShortcut: editSnapshotShortcut.trim() || 'CmdOrCtrl+Shift+S',
      snapshotDelayMs: Math.max(10, Math.min(2000, Number(editSnapshotDelayMs) || 50)),
      snapshotSoundEnabled: editSnapshotSoundEnabled,
      snapshotFlashEnabled: editSnapshotFlashEnabled,
      snapshotAutoAttach: editSnapshotAutoAttach,
      activeWindowTurnCount: Math.max(2, Math.min(50, Number(editActiveWindowTurnCount) || 10)),
      maxContextTokens: Math.max(16000, Number(editMaxContextTokens) || 200000),
      defaultModel: finalModel,
      defaultReasoningEffort: editDefaultReasoningEffort,
      permissionMode: editPermissionMode,
      planGateMode: editPlanGateMode,
      animationsEnabled: editAnimationsEnabled,
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
    editSnapshotShortcut = settingsStore.snapshotShortcut;
    editSnapshotDelayMs = settingsStore.snapshotDelayMs;
    editSnapshotSoundEnabled = settingsStore.snapshotSoundEnabled;
    editSnapshotFlashEnabled = settingsStore.snapshotFlashEnabled;
    editSnapshotAutoAttach = settingsStore.snapshotAutoAttach;
    editActiveWindowTurnCount = settingsStore.activeWindowTurnCount;
    editMaxContextTokens = 200000;
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
      class="w-full max-w-4xl h-[680px] max-h-[90vh] flex flex-col bg-ant-bg-secondary border border-white/10 rounded-2xl shadow-2xl overflow-hidden"
    >
      <!-- Header -->
      <div class="px-6 py-4 border-b border-white/5 flex items-center justify-between bg-ant-bg-secondary flex-shrink-0">
        <div class="flex items-center space-x-3">
          <div class="w-8 h-8 rounded-lg bg-ant-primary/10 flex items-center justify-center text-ant-primary shadow-sm">
            <Settings size={18} />
          </div>
          <div>
            <h2 id="settings-modal-title" class="font-serif-display text-base font-semibold text-ant-text tracking-tight flex items-center gap-2">
              Settings & Workspace Preferences
              <button
                type="button"
                onclick={() => openExternal('https://github.com/fiko942/grok-build')}
                class="px-2 py-0.5 text-[10px] font-serif font-medium bg-ant-primary/15 hover:bg-ant-primary/25 text-ant-primary rounded-full transition cursor-pointer flex items-center gap-1 border-0"
                title="View AetherGrok repository on GitHub"
              >
                <span>AetherGrok Studio</span>
                <ExternalLink size={10} class="opacity-70" />
              </button>
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
            class="p-1.5 rounded-lg text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition"
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
        <nav class="w-56 bg-ant-bg border-r border-white/5 flex flex-col p-2 space-y-1 flex-shrink-0" aria-label="Settings navigation">
          {#each tabs as tab}
            {@const IconComponent = tab.icon}
            <button
              type="button"
              class="w-full flex items-center space-x-2.5 px-3 py-2 rounded-lg text-xs font-medium text-left transition {activeTab === tab.id
                ? 'bg-ant-primary/15 text-ant-primary font-semibold'
                : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary'}"
              onclick={() => activeTab = tab.id}
            >
              <IconComponent size={15} class="flex-shrink-0 {activeTab === tab.id ? 'text-ant-primary' : 'text-ant-text-secondary'}" />
              <span class="truncate">{tab.label}</span>
            </button>
          {/each}

          <div class="mt-auto pt-4 border-t border-white/5 px-2 pb-2">
            <button
              type="button"
              onclick={() => openExternal('https://github.com/fiko942/grok-build')}
              class="w-full text-left p-1.5 rounded-lg hover:bg-ant-bg-tertiary transition group cursor-pointer"
              title="Open GitHub repository"
            >
              <div class="text-[11px] text-ant-text group-hover:text-ant-primary font-mono leading-tight flex items-center justify-between">
                <span>AetherGrok Studio</span>
                <ExternalLink size={10} class="opacity-40 group-hover:opacity-100" />
              </div>
              <div class="text-[10px] text-ant-text-muted mt-0.5 font-mono">Build {__APP_VERSION__} (Wails/Go)</div>
            </button>
          </div>
        </nav>

        <!-- Tab Content Viewport -->
        <div class="flex-1 min-h-0 overflow-y-auto p-6 space-y-6 custom-scrollbar bg-ant-bg-secondary/60">
          
          <!-- TAB 1: GENERAL & SNAPSHOT -->
          {#if activeTab === 'general'}
            <div class="space-y-6 animate-in fade-in duration-100">
              <div>
                <h3 class="font-serif-display text-base font-semibold text-ant-text">General & Smart Screen Snapshot</h3>
                <p class="font-serif text-xs text-ant-text-secondary mt-0.5">
                  Configure CLI runner location, native snapshot behavior on macOS & Windows, and DOM memory bounds.
                </p>
              </div>

              <!-- Grok Binary Path Card -->
              <Card>
                <div class="space-y-3">
                  <div class="flex items-center justify-between">
                    <div>
                      <div class="text-xs font-semibold text-ant-text">Grok CLI Executable Path</div>
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
                    class="w-full px-3 py-2 text-xs font-mono bg-ant-bg border border-white/10 rounded-lg text-ant-text focus:outline-none focus:border-ant-primary focus:ring-1 focus:ring-ant-primary transition"
                  />
                </div>
              </Card>

              <!-- Smart Screen Snapshot Settings Card -->
              <Card>
                <div class="space-y-4">
                  <div class="flex items-center justify-between border-b border-white/5 pb-2">
                    <div class="flex items-center gap-2">
                      <Camera size={16} class="text-ant-primary" />
                      <div class="text-xs font-semibold text-ant-text">Non-Intrusive Smart Snapshot (macOS & Windows)</div>
                    </div>
                    <Badge status="success">Active Native Grab</Badge>
                  </div>

                  <!-- Snapshot Shortcut Key Binding -->
                  <div class="space-y-2 pt-2 border-t border-white/5">
                    <div class="flex items-center justify-between">
                      <div class="space-y-0.5">
                        <div class="text-xs font-medium text-ant-text flex items-center gap-1.5">
                          <Keyboard size={13} class="text-ant-primary" />
                          Global Screen Snapshot Shortcut
                        </div>
                        <div class="text-[11px] text-ant-text-secondary">
                          Custom keybinding to trigger native screen capture instantly.
                        </div>
                      </div>
                      <div class="flex items-center gap-1.5">
                        {#each (editSnapshotShortcut ? editSnapshotShortcut.split('+') : ['CmdOrCtrl', 'Shift', 'S']) as keySegment}
                          <kbd class="px-2 py-0.5 text-xs font-mono font-medium text-ant-primary bg-ant-primary/10 rounded">
                            {keySegment.trim()}
                          </kbd>
                        {/each}
                      </div>
                    </div>

                    <div class="flex items-center gap-2">
                      <button
                        type="button"
                        onclick={() => isRecordingShortcut = true}
                        class="flex-1 px-3 py-1.5 text-xs font-serif rounded-lg bg-ant-primary/10 hover:bg-ant-primary/20 text-ant-primary transition flex items-center justify-center gap-2"
                      >
                        <Keyboard size={14} />
                        Record / Change Shortcut
                      </button>
                      <button
                        type="button"
                        class="px-2.5 py-1.5 text-xs rounded-lg bg-white/[0.04] hover:bg-white/[0.08] text-ant-text-secondary hover:text-ant-text transition"
                        onclick={() => editSnapshotShortcut = 'CmdOrCtrl+Shift+S'}
                      >
                        Reset Default
                      </button>
                    </div>
                  </div>

                  <!-- Audio Shutter Sound Toggle -->
                  <div class="flex items-center justify-between pt-1">
                    <div class="space-y-0.5">
                      <div class="text-xs font-medium text-ant-text flex items-center gap-1.5">
                        <Volume2 size={13} class="text-ant-primary" />
                        Camera Shutter Audio (Shutter Sound)
                      </div>
                      <div class="text-[11px] text-ant-text-secondary">
                        Synthesizes a realistic mechanical shutter click on capture via Web Audio API.
                      </div>
                    </div>
                    <Switch
                      bind:checked={editSnapshotSoundEnabled}
                      size="small"
                    />
                  </div>

                  <!-- Screen Flash FX Toggle -->
                  <div class="flex items-center justify-between pt-2 border-t border-white/5">
                    <div class="space-y-0.5">
                      <div class="text-xs font-medium text-ant-text flex items-center gap-1.5">
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
                  <div class="flex items-center justify-between pt-2 border-t border-white/5">
                    <div class="space-y-0.5">
                      <div class="text-xs font-medium text-ant-text flex items-center gap-1.5">
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
                  <div class="space-y-2 pt-2 border-t border-white/5">
                    <div class="flex items-center justify-between">
                      <div>
                        <div class="text-xs font-medium text-ant-text">OS Compositor Flush Delay</div>
                        <div class="text-[11px] text-ant-text-secondary">
                          Delay between window hide and screen capture to prevent window ghost frames.
                        </div>
                      </div>
                      <span class="text-xs font-mono font-bold text-ant-primary bg-ant-primary/10 px-2 py-0.5 rounded">
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
                      <div class="text-xs font-semibold text-ant-text">DOM Active Turn Windowing (RAM Guard)</div>
                      <div class="text-[11px] text-ant-text-secondary">
                        Limits active DOM messages to preserve low memory usage (&lt;60MB RAM).
                      </div>
                    </div>
                    <span class="text-xs font-mono font-bold text-ant-primary bg-ant-primary/10 px-2 py-0.5 rounded">
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
                <h3 class="font-serif-display text-base font-semibold text-ant-text">Inference Models & Reasoning Depth</h3>
                <p class="font-serif text-xs text-ant-text-secondary mt-0.5">
                  Select default model engine and reasoning effort stops for complex tasks.
                </p>
              </div>

              <!-- Model Selector Cards -->
              <div class="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  class="p-4 rounded-xl border text-left transition {editDefaultModel === '9router'
                    ? 'border-ant-primary/40 bg-ant-primary/10 shadow-sm'
                    : 'border-white/5 bg-ant-bg hover:border-white/20'}"
                  onclick={() => editDefaultModel = '9router'}
                >
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-xs font-bold text-ant-text flex items-center gap-1.5">
                      <Sparkles size={14} class="text-ant-primary" /> 9router (Default)
                    </span>
                    {#if editDefaultModel === '9router'}
                      <CheckCircle2 size={15} class="text-ant-primary" />
                    {/if}
                  </div>
                  <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                    Primary multi-provider gateway supporting ultra-fast streaming, reasoning synthesis, and vision multimodal inputs.
                  </p>
                </button>

                <button
                  type="button"
                  class="p-4 rounded-xl border text-left transition {editDefaultModel === '9router-general-purpose'
                    ? 'border-ant-primary/40 bg-ant-primary/10 shadow-sm'
                    : 'border-white/5 bg-ant-bg hover:border-white/20'}"
                  onclick={() => editDefaultModel = '9router-general-purpose'}
                >
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-xs font-bold text-ant-text flex items-center gap-1.5">
                      <Brain size={14} class="text-ant-primary" /> 9router General Purpose
                    </span>
                    {#if editDefaultModel === '9router-general-purpose'}
                      <CheckCircle2 size={15} class="text-ant-primary" />
                    {/if}
                  </div>
                  <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                    Balanced configuration for broad tasks, planning, refactoring, and general desktop workflows.
                  </p>
                </button>

                <button
                  type="button"
                  class="p-4 rounded-xl border text-left transition {editDefaultModel === '9router-explore'
                    ? 'border-ant-primary/40 bg-ant-primary/10 shadow-sm'
                    : 'border-white/5 bg-ant-bg hover:border-white/20'}"
                  onclick={() => editDefaultModel = '9router-explore'}
                >
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-xs font-bold text-ant-text flex items-center gap-1.5">
                      <Sparkles size={14} class="text-ant-primary" /> 9router Explore
                    </span>
                    {#if editDefaultModel === '9router-explore'}
                      <CheckCircle2 size={15} class="text-ant-primary" />
                    {/if}
                  </div>
                  <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                    Fast exploratory model for rapid codebase search, syntax checks, and file discovery.
                  </p>
                </button>

                <button
                  type="button"
                  class="p-4 rounded-xl border text-left transition {editDefaultModel === '9router-plan'
                    ? 'border-ant-primary/40 bg-ant-primary/10 shadow-sm'
                    : 'border-white/5 bg-ant-bg hover:border-white/20'}"
                  onclick={() => editDefaultModel = '9router-plan'}
                >
                  <div class="flex items-center justify-between mb-1">
                    <span class="text-xs font-bold text-ant-text flex items-center gap-1.5">
                      <Brain size={14} class="text-ant-primary" /> 9router Plan
                    </span>
                    {#if editDefaultModel === '9router-plan'}
                      <CheckCircle2 size={15} class="text-ant-primary" />
                    {/if}
                  </div>
                  <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                    Architectural planning model dedicated to breaking down complex engineering requirements.
                  </p>
                </button>
              </div>

              <!-- Reasoning Effort Levels -->
              <Card>
                <div class="space-y-3">
                  <div>
                    <div class="text-xs font-semibold text-ant-text">Default Reasoning Effort</div>
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
                          : 'bg-ant-bg text-ant-text-secondary border-white/5 hover:text-ant-text hover:bg-ant-bg-tertiary'}"
                        onclick={() => editDefaultReasoningEffort = effort as ReasoningEffort}
                      >
                        {effort}
                      </button>
                    {/each}
                  </div>
                </div>
              </Card>

              <!-- Context Window Token Limit Setting -->
              <Card>
                <div class="space-y-3.5">
                  <div class="flex items-center justify-between">
                    <div>
                      <div class="text-xs font-semibold text-ant-text flex items-center gap-1.5">
                        <Brain size={14} class="text-ant-primary" />
                        Session Context Window Limit
                      </div>
                      <div class="text-[11px] text-ant-text-secondary">
                        Maximum token capacity for conversation history and active reasoning context (Default: 200k).
                      </div>
                    </div>
                    <span class="text-xs font-mono font-bold text-ant-primary bg-ant-primary/10 px-2.5 py-1 rounded-lg">
                      {editMaxContextTokens >= 1000000 ? `${(editMaxContextTokens / 1000000).toFixed(0)}M` : `${Math.round(editMaxContextTokens / 1000)}k`} tokens
                    </span>
                  </div>

                  <!-- Quick Presets -->
                  <div class="grid grid-cols-4 gap-2">
                    {#each [
                      { label: '128k', value: 128000, desc: 'Fast / Compact' },
                      { label: '200k (Default)', value: 200000, desc: 'Recommended' },
                      { label: '500k', value: 500000, desc: 'Large Transcripts' },
                      { label: '1M', value: 1000000, desc: 'Ultra Deep Context' }
                    ] as preset}
                      <button
                        type="button"
                        class="p-2.5 rounded-lg border text-left transition {editMaxContextTokens === preset.value
                          ? 'bg-ant-primary/15 border-ant-primary text-ant-text font-semibold shadow-sm'
                          : 'bg-ant-bg border-white/5 text-ant-text-secondary hover:border-white/15 hover:text-ant-text'}"
                        onclick={() => editMaxContextTokens = preset.value}
                      >
                        <div class="text-xs font-bold {editMaxContextTokens === preset.value ? 'text-ant-primary' : 'text-ant-text'}">{preset.label}</div>
                        <div class="text-[10px] text-ant-text-muted mt-0.5">{preset.desc}</div>
                      </button>
                    {/each}
                  </div>

                  <!-- Custom Numeric Token Limit Input -->
                  <div class="pt-2 border-t border-white/5 flex items-center gap-3">
                    <div class="text-[11px] text-ant-text-secondary whitespace-nowrap">
                      Custom Token Limit:
                    </div>
                    <input
                      type="number"
                      min="16000"
                      max="2000000"
                      step="1000"
                      bind:value={editMaxContextTokens}
                      class="flex-1 px-3 py-1.5 text-xs font-mono bg-ant-bg border border-white/10 rounded-lg text-ant-text focus:outline-none focus:border-ant-primary focus:ring-1 focus:ring-ant-primary transition"
                      placeholder="200000"
                    />
                    <button
                      type="button"
                      class="px-2.5 py-1.5 text-xs rounded-lg border border-white/10 bg-ant-bg hover:bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-text transition"
                      onclick={() => editMaxContextTokens = 200000}
                    >
                      Reset 200k
                    </button>
                  </div>
                </div>
              </Card>

              <!-- Open Config File Card -->
              <Card>
                <div class="flex items-center justify-between">
                  <div class="space-y-0.5">
                    <div class="text-xs font-semibold text-ant-text flex items-center gap-1.5">
                      <FolderOpen size={14} class="text-ant-primary" />
                      Grok Configuration File
                    </div>
                    <div class="text-[11px] text-ant-text-secondary">
                      Open ~/.grok/config.toml in your system file manager (Finder / Explorer) with the file focused.
                    </div>
                  </div>
                  <button
                    type="button"
                    class="px-3 py-1.5 rounded-lg bg-ant-primary/10 hover:bg-ant-primary/20 text-ant-primary text-xs font-medium border border-ant-primary/20 hover:border-ant-primary/40 transition flex items-center gap-1.5 cursor-pointer"
                    onclick={async () => {
                      try {
                        if (window.go?.main?.App?.RevealGrokConfigFile) {
                          await window.go.main.App.RevealGrokConfigFile();
                        }
                      } catch (err) {
                        console.error('Failed to open config file:', err);
                      }
                    }}
                  >
                    <FileCode2 size={13} />
                    <span>Open Config File</span>
                  </button>
                </div>
              </Card>
            </div>
          {/if}

          <!-- TAB 3: PERMISSIONS & PLAN GATE -->
          {#if activeTab === 'permissions'}
            <div class="space-y-6 animate-in fade-in duration-100">
              <div>
                <h3 class="font-serif-display text-base font-semibold text-ant-text">Security & Execution Boundaries</h3>
                <p class="font-serif text-xs text-ant-text-secondary mt-0.5">
                  Control how AetherGrok asks for confirmation before executing bash commands or modifying project files.
                </p>
              </div>

              <!-- Plan Gate Mode Configuration -->
              <Card>
                <div class="space-y-3">
                  <div class="flex items-center justify-between">
                    <div>
                      <div class="text-xs font-semibold text-ant-text flex items-center gap-1.5">
                        <ShieldCheck size={14} class="text-ant-primary" />
                        Plan Gate Mode
                      </div>
                      <div class="text-[11px] text-ant-text-secondary">
                        Choose whether plan proposals require manual approval or are automatically approved and executed.
                      </div>
                    </div>
                    <span class="text-xs font-mono font-bold {editPlanGateMode === 'bypass' ? 'text-amber-400 bg-amber-400/10' : 'text-ant-primary bg-ant-primary/10'} px-2 py-0.5 rounded capitalize">
                      {editPlanGateMode === 'bypass' ? 'Bypass (Auto)' : 'Active (Manual)'}
                    </span>
                  </div>

                  <div class="grid grid-cols-2 gap-2.5 pt-1">
                    <button
                      type="button"
                      class="p-3 rounded-lg border text-left transition {editPlanGateMode === 'active'
                        ? 'bg-ant-primary/15 border-ant-primary text-ant-text shadow-sm'
                        : 'bg-ant-bg border-white/5 text-ant-text-secondary hover:border-white/15 hover:text-ant-text'}"
                      onclick={() => editPlanGateMode = 'active'}
                    >
                      <div class="flex items-center justify-between">
                        <div class="text-xs font-semibold text-ant-text">Active (Manual Review)</div>
                        {#if editPlanGateMode === 'active'}
                          <CheckCircle2 size={14} class="text-ant-primary" />
                        {/if}
                      </div>
                      <div class="text-[10px] text-ant-text-muted mt-1 leading-relaxed">
                        Presents interactive Approve, Implement, and Custom Feedback action cards before proceeding.
                      </div>
                    </button>

                    <button
                      type="button"
                      class="p-3 rounded-lg border text-left transition {editPlanGateMode === 'bypass'
                        ? 'bg-amber-500/15 border-amber-500 text-ant-text shadow-sm'
                        : 'bg-ant-bg border-white/5 text-ant-text-secondary hover:border-white/15 hover:text-ant-text'}"
                      onclick={() => editPlanGateMode = 'bypass'}
                    >
                      <div class="flex items-center justify-between">
                        <div class="text-xs font-semibold text-amber-300">Bypass (Auto-Approve)</div>
                        {#if editPlanGateMode === 'bypass'}
                          <CheckCircle2 size={14} class="text-amber-400" />
                        {/if}
                      </div>
                      <div class="text-[10px] text-ant-text-muted mt-1 leading-relaxed">
                        Instantly auto-approves and implements any plan proposals generated by the agent without blocking.
                      </div>
                    </button>
                  </div>
                </div>
              </Card>

              <div class="space-y-3">
                {#each (Object.keys(permissionDescriptions) as Array<PermissionMode>) as modeKey}
                  {@const mode = permissionDescriptions[modeKey]}
                  <button
                    type="button"
                    class="w-full p-4 rounded-xl border text-left transition {editPermissionMode === modeKey
                      ? 'border-ant-primary/40 bg-ant-primary/10 shadow-sm'
                      : 'border-white/5 bg-ant-bg hover:border-white/20'}"
                    onclick={() => editPermissionMode = modeKey}
                  >
                    <div class="flex items-center justify-between mb-2">
                      <div class="flex items-center space-x-2">
                        <span class="text-xs font-bold text-ant-text">{mode.title}</span>
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

                    <ul class="space-y-1 pl-4 border-l-2 border-white/10 text-[10px] text-ant-text-secondary">
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
                <h3 class="font-serif-display text-base font-semibold text-ant-text">Theme & Color Customization</h3>
                <p class="font-serif text-xs text-ant-text-secondary mt-0.5">
                  Select a tailored desktop aesthetic optimized for deep contrast, minimal eye strain, and readability.
                </p>
              </div>

              <div class="grid grid-cols-3 gap-3.5">
                <!-- Dark Studio -->
                <button
                  type="button"
                  class="group p-4 rounded-xl border text-left transition-all duration-150 flex flex-col justify-between space-y-3.5 relative overflow-hidden {editTheme === 'dark-studio'
                    ? 'border-ant-primary/40 ring-1 ring-ant-primary/30 bg-ant-bg-tertiary/80 shadow-lg shadow-black/20'
                    : 'border-white/5 bg-ant-bg hover:border-white/20 hover:bg-ant-bg-tertiary/40'}"
                  onclick={() => handleThemeChange('dark-studio')}
                >
                  <div class="flex items-center justify-between w-full">
                    <div class="flex items-center space-x-2">
                      <span class="w-2.5 h-2.5 rounded-full bg-[#1677FF]"></span>
                      <span class="text-xs font-bold text-ant-text">Dark Studio</span>
                    </div>
                    {#if editTheme === 'dark-studio'}
                      <CheckCircle2 size={16} class="text-ant-primary" />
                    {/if}
                  </div>

                  <!-- Theme Swatches & Mini Window Mockup -->
                  <div class="h-20 rounded-lg bg-[#0F1117] border border-white/5 p-2.5 flex flex-col justify-between shadow-inner w-full">
                    <div class="flex items-center justify-between">
                      <div class="flex space-x-1.5">
                        <div class="w-2.5 h-2.5 rounded-full bg-[#FF5F56]"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-[#FFBD2E]"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-[#27C93F]"></div>
                      </div>
                      <span class="text-[9px] font-mono text-[#4096FF]">#0F1117</span>
                    </div>
                    <div class="space-y-1.5">
                      <div class="h-2 w-24 bg-[#181B26] rounded border border-white/5"></div>
                      <div class="h-1.5 w-16 bg-[#1677FF]/40 rounded"></div>
                    </div>
                  </div>

                  <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                    Deep slate canvas with neon cyan & vibrant blue accents. Ideal for long coding sessions.
                  </p>
                </button>

                <!-- Dark High Contrast -->
                <button
                  type="button"
                  class="group p-4 rounded-xl border text-left transition-all duration-150 flex flex-col justify-between space-y-3.5 relative overflow-hidden {editTheme === 'dark-high-contrast'
                    ? 'border-ant-primary/40 ring-1 ring-ant-primary/30 bg-ant-bg-tertiary/80 shadow-lg shadow-black/20'
                    : 'border-white/5 bg-ant-bg hover:border-white/20 hover:bg-ant-bg-tertiary/40'}"
                  onclick={() => handleThemeChange('dark-high-contrast')}
                >
                  <div class="flex items-center justify-between w-full">
                    <div class="flex items-center space-x-2">
                      <span class="w-2.5 h-2.5 rounded-full bg-[#388BFD]"></span>
                      <span class="text-xs font-bold text-ant-text">High Contrast OLED</span>
                    </div>
                    {#if editTheme === 'dark-high-contrast'}
                      <CheckCircle2 size={16} class="text-ant-primary" />
                    {/if}
                  </div>

                  <!-- Theme Swatches & Mini Window Mockup -->
                  <div class="h-20 rounded-lg bg-[#000000] border border-white/5 p-2.5 flex flex-col justify-between shadow-inner w-full">
                    <div class="flex items-center justify-between">
                      <div class="flex space-x-1.5">
                        <div class="w-2.5 h-2.5 rounded-full bg-[#F85149]"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-[#D29922]"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-[#3FB950]"></div>
                      </div>
                      <span class="text-[9px] font-mono text-[#58A6FF]">#000000</span>
                    </div>
                    <div class="space-y-1.5">
                      <div class="h-2 w-24 bg-[#1C1C1F] rounded border border-white/5"></div>
                      <div class="h-1.5 w-16 bg-[#388BFD]/50 rounded"></div>
                    </div>
                  </div>

                  <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                    True black OLED background with stark high-contrast typography for maximum clarity.
                  </p>
                </button>

                <!-- Light Studio -->
                <button
                  type="button"
                  class="group p-4 rounded-xl border text-left transition-all duration-150 flex flex-col justify-between space-y-3.5 relative overflow-hidden {editTheme === 'light-antd'
                    ? 'border-ant-primary/40 ring-1 ring-ant-primary/30 bg-ant-bg-tertiary/80 shadow-lg shadow-black/20'
                    : 'border-white/5 bg-ant-bg hover:border-white/20 hover:bg-ant-bg-tertiary/40'}"
                  onclick={() => handleThemeChange('light-antd')}
                >
                  <div class="flex items-center justify-between w-full">
                    <div class="flex items-center space-x-2">
                      <span class="w-2.5 h-2.5 rounded-full bg-[#1677FF]"></span>
                      <span class="text-xs font-bold text-ant-text">Clean Light</span>
                    </div>
                    {#if editTheme === 'light-antd'}
                      <CheckCircle2 size={16} class="text-ant-primary" />
                    {/if}
                  </div>

                  <!-- Theme Swatches & Mini Window Mockup -->
                  <div class="h-20 rounded-lg bg-[#F5F5F5] border border-white/10 p-2.5 flex flex-col justify-between shadow-inner w-full">
                    <div class="flex items-center justify-between">
                      <div class="flex space-x-1.5">
                        <div class="w-2.5 h-2.5 rounded-full bg-[#FF4D4F]"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-[#FAAD14]"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-[#52C41A]"></div>
                      </div>
                      <span class="text-[9px] font-mono text-[#0958D9]">#FFFFFF</span>
                    </div>
                    <div class="space-y-1.5">
                      <div class="h-2 w-24 bg-[#FFFFFF] rounded border border-white/10"></div>
                      <div class="h-1.5 w-16 bg-[#1677FF]/30 rounded"></div>
                    </div>
                  </div>

                  <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                    Crisp paper-white theme designed for bright rooms and daytime productivity.
                  </p>
                </button>
              </div>

              <!-- Smooth UI Animations Toggle -->
              <Card>
                <div class="flex items-center justify-between">
                  <div class="space-y-0.5">
                    <div class="text-xs font-semibold text-ant-text flex items-center gap-1.5">
                      <Sparkles size={14} class="text-ant-primary" />
                      Fluid UI Transitions & Animations
                    </div>
                    <div class="text-[11px] text-ant-text-secondary">
                      Enable smooth tab peels, sidebar expansion slides, switch toggles, and action group animations. Disable on low-spec hardware.
                    </div>
                  </div>
                  <Switch
                    bind:checked={editAnimationsEnabled}
                    size="small"
                  />
                </div>
              </Card>
            </div>
          {/if}

          <!-- TAB 5: KEYBOARD SHORTCUTS -->
          {#if activeTab === 'shortcuts'}
            <div class="space-y-6 animate-in fade-in duration-100">
              <div>
                <h3 class="font-serif-display text-base font-semibold text-ant-text">Keyboard Shortcuts Reference</h3>
                <p class="font-serif text-xs text-ant-text-secondary mt-0.5">
                  Accelerate your workflow with dedicated keystroke accelerators.
                </p>
              </div>

              <div class="rounded-xl overflow-hidden bg-ant-bg">
                <table class="w-full text-left border-collapse text-xs">
                  <thead>
                    <tr class="bg-ant-bg-tertiary/40">
                      <th class="px-4 py-2.5 font-semibold text-ant-text">Key Combination</th>
                      <th class="px-4 py-2.5 font-semibold text-ant-text">Action Trigger</th>
                      <th class="px-4 py-2.5 font-semibold text-ant-text-secondary">Scope</th>
                    </tr>
                  </thead>
                  <tbody class="divide-y divide-white/5">
                    {#each keyboardShortcuts as shortcut}
                      <tr class="hover:bg-ant-bg-secondary/40 transition">
                        <td class="px-4 py-2.5">
                          <div class="flex items-center gap-1.5 flex-wrap">
                            {#each shortcut.keys as k, i}
                              <kbd class="px-2 py-0.5 text-xs font-mono font-medium text-ant-primary bg-ant-primary/10 rounded">
                                {k}
                              </kbd>
                              {#if i < shortcut.keys.length - 1}
                                <span class="text-[11px] text-ant-text-muted font-mono font-semibold">+</span>
                              {/if}
                            {/each}
                          </div>
                        </td>
                        <td class="px-4 py-2.5 text-ant-text font-serif">{shortcut.action}</td>
                        <td class="px-4 py-2.5 text-ant-text-secondary">
                          <span class="px-2 py-0.5 text-[10px] font-medium bg-white/[0.04] text-ant-text-secondary rounded">
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

          <!-- TAB 6: ABOUT AETHERGROK -->
          {#if activeTab === 'about'}
            <div class="space-y-6 animate-in fade-in duration-100">
              <div>
                <h3 class="font-serif-display text-base font-semibold text-ant-text flex items-center gap-2">
                  AetherGrok Studio
                  <span class="px-2 py-0.5 text-[10px] font-mono bg-ant-primary/15 text-ant-primary rounded-full">v{__APP_VERSION__}</span>
                </h3>
                <p class="font-serif text-xs text-ant-text-secondary mt-1">
                  A tactile, local-first desktop workstation for autonomous Grok CLI workflows.
                </p>
              </div>

              <!-- Human Crafted Dedication Card -->
              <div class="p-4 rounded-xl bg-ant-bg space-y-2">
                <div class="flex items-center gap-2 text-ant-text text-xs font-serif font-medium">
                  <Heart size={14} class="text-rose-500 fill-rose-500/20" />
                  <span>Made with love by Wiji Fiko Teren</span>
                </div>
                <p class="font-serif text-xs text-ant-text-secondary leading-relaxed">
                  Crafted for developers who care about tactile typography, keyboard-driven navigation, and direct access to native tools without heavy cloud wrappers or slow interfaces.
                </p>
              </div>

              <!-- Open Source & Developer Portfolio Links -->
              <div class="grid grid-cols-2 gap-3.5">
                <!-- GitHub Repository Card -->
                <button
                  type="button"
                  onclick={() => openExternal('https://github.com/fiko942/grok-build')}
                  class="p-4 rounded-xl bg-ant-bg hover:bg-ant-bg-tertiary/40 space-y-3 flex flex-col justify-between transition text-left cursor-pointer group"
                >
                  <div class="space-y-1.5">
                    <div class="flex items-center space-x-2 text-ant-text group-hover:text-ant-primary transition-colors">
                      <Github size={16} class="text-ant-primary" />
                      <span class="text-xs font-semibold">Open-Source Project</span>
                    </div>
                    <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                      Source code, issue tracking, and contributions on GitHub.
                    </p>
                  </div>
                  <div class="inline-flex items-center justify-between w-full px-3 py-2 rounded-lg text-xs font-mono bg-ant-bg-secondary group-hover:bg-white/5 text-ant-text transition">
                    <span class="truncate">github.com/fiko942/grok-build</span>
                    <ExternalLink size={12} class="opacity-60 group-hover:opacity-100 ml-1.5 flex-shrink-0" />
                  </div>
                </button>

                <!-- Developer Portfolio Card -->
                <button
                  type="button"
                  onclick={() => openExternal('https://wijifikoteren.streampeg.com')}
                  class="p-4 rounded-xl bg-ant-bg hover:bg-ant-bg-tertiary/40 space-y-3 flex flex-col justify-between transition text-left cursor-pointer group"
                >
                  <div class="space-y-1.5">
                    <div class="flex items-center space-x-2 text-ant-text group-hover:text-ant-primary transition-colors">
                      <Globe size={16} class="text-ant-primary" />
                      <span class="text-xs font-semibold">Developer Portfolio</span>
                    </div>
                    <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                      Personal portfolio, software projects, design experiments, and writing.
                    </p>
                  </div>
                  <div class="inline-flex items-center justify-between w-full px-3 py-2 rounded-lg text-xs font-mono bg-ant-primary/10 group-hover:bg-ant-primary/20 text-ant-primary transition">
                    <span class="truncate">wijifikoteren.streampeg.com</span>
                    <ExternalLink size={12} class="opacity-80 group-hover:opacity-100 ml-1.5 flex-shrink-0" />
                  </div>
                </button>
              </div>

              <!-- Clean Engineering Overview -->
              <div class="p-3.5 rounded-xl bg-ant-bg space-y-2 text-xs">
                <div class="font-serif font-medium text-ant-text">Core Capabilities</div>
                <div class="grid grid-cols-2 gap-2 text-[11px] text-ant-text-secondary font-serif">
                  <div class="p-2 rounded-lg bg-ant-bg-secondary/60">
                    <div class="font-sans font-semibold text-ant-text mb-0.5">Tactile Desktop Speed</div>
                    Direct Go process management with low-latency streaming and zero telemetry bloat.
                  </div>
                  <div class="p-2 rounded-lg bg-ant-bg-secondary/60">
                    <div class="font-sans font-semibold text-ant-text mb-0.5">Smart Screen Capture</div>
                    Non-intrusive full-display snapshots with auto-window exclusion for visual grounding.
                  </div>
                </div>
              </div>

              <!-- Tech Stack Footer Info -->
              <div class="pt-1 text-center text-[10.5px] text-ant-text-muted font-mono flex items-center justify-center gap-2">
                <span>Wails v2 (Go 1.24)</span>
                <span>•</span>
                <span>Svelte 5</span>
                <span>•</span>
                <span>Anthropic Serif</span>
              </div>
            </div>
          {/if}

        </div>
      </div>

      <!-- Footer Action Toolbar -->
      <div class="px-6 py-3 border-t border-white/5 flex items-center justify-between bg-ant-bg-secondary flex-shrink-0">
        <button
          type="button"
          class="text-xs text-ant-text-secondary hover:text-ant-text flex items-center gap-1.5 transition"
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

  <!-- Key Recorder Modal for Live Input Capture -->
  <KeyRecorderModal
    open={isRecordingShortcut}
    currentShortcut={editSnapshotShortcut}
    onSave={(newKey) => {
      editSnapshotShortcut = newKey;
      isRecordingShortcut = false;
    }}
    onCancel={() => {
      isRecordingShortcut = false;
    }}
  />
{/if}
