import { logger } from './logger.svelte';

export type ThemeMode = 'dark-studio' | 'dark-high-contrast' | 'light-antd';
export type DefaultModel = '9router' | '9router-general-purpose' | '9router-explore' | '9router-plan' | 'custom' | string;
export type ReasoningEffort = 'none' | 'low' | 'medium' | 'high' | 'max';
export type PermissionMode = 'default' | 'acceptEdits' | 'auto' | 'plan' | 'bypassPermissions';
export type PlanGateMode = 'active' | 'bypass';

export interface AppSettings {
  theme: ThemeMode;
  defaultModel: DefaultModel;
  defaultReasoningEffort: ReasoningEffort;
  permissionMode: PermissionMode;
  planGateMode: PlanGateMode;
  animationsEnabled: boolean;
  grokBinaryPath: string;
  snapshotShortcut: string;
  snapshotDelayMs: number;
  snapshotAutoHideWindow: boolean;
  snapshotSoundEnabled: boolean;
  snapshotFlashEnabled: boolean;
  snapshotAutoAttach: boolean;
  activeWindowTurnCount: number;
  maxContextTokens: number;
  sidebarWidth: number;
  sidebarCollapsed: boolean;
  selectedMicrophoneDeviceId: string;
  dictationShortcut: string;
  dictationMuteSystemAudio: boolean;
  dictationHoldThresholdMs: number;
}

const STORAGE_KEY = 'aethergrok_settings_v1';

export const DEFAULT_SETTINGS: AppSettings = {
  theme: 'dark-studio',
  defaultModel: '9router',
  defaultReasoningEffort: 'medium',
  permissionMode: 'default',
  planGateMode: 'active',
  animationsEnabled: true,
  grokBinaryPath: '',
  snapshotShortcut: 'CmdOrCtrl+Shift+S',
  snapshotDelayMs: 50,
  snapshotAutoHideWindow: true,
  snapshotSoundEnabled: true,
  snapshotFlashEnabled: true,
  snapshotAutoAttach: true,
  activeWindowTurnCount: 10,
  maxContextTokens: 200000,
  sidebarWidth: 288,
  sidebarCollapsed: false,
  selectedMicrophoneDeviceId: '',
  dictationShortcut: '\\',
  dictationMuteSystemAudio: true,
  dictationHoldThresholdMs: 200
};

export class SettingsStore {
  isHydrated = $state<boolean>(false);
  theme = $state<ThemeMode>(DEFAULT_SETTINGS.theme);
  defaultModel = $state<DefaultModel>(DEFAULT_SETTINGS.defaultModel);
  defaultReasoningEffort = $state<ReasoningEffort>(DEFAULT_SETTINGS.defaultReasoningEffort);
  permissionMode = $state<PermissionMode>(DEFAULT_SETTINGS.permissionMode);
  planGateMode = $state<PlanGateMode>(DEFAULT_SETTINGS.planGateMode);
  animationsEnabled = $state<boolean>(DEFAULT_SETTINGS.animationsEnabled);
  grokBinaryPath = $state<string>(DEFAULT_SETTINGS.grokBinaryPath);
  snapshotShortcut = $state<string>(DEFAULT_SETTINGS.snapshotShortcut);
  snapshotDelayMs = $state<number>(DEFAULT_SETTINGS.snapshotDelayMs);
  snapshotAutoHideWindow = $state<boolean>(DEFAULT_SETTINGS.snapshotAutoHideWindow);
  snapshotSoundEnabled = $state<boolean>(DEFAULT_SETTINGS.snapshotSoundEnabled);
  snapshotFlashEnabled = $state<boolean>(DEFAULT_SETTINGS.snapshotFlashEnabled);
  snapshotAutoAttach = $state<boolean>(DEFAULT_SETTINGS.snapshotAutoAttach);
  activeWindowTurnCount = $state<number>(DEFAULT_SETTINGS.activeWindowTurnCount);
  maxContextTokens = $state<number>(DEFAULT_SETTINGS.maxContextTokens);
  sidebarWidth = $state<number>(DEFAULT_SETTINGS.sidebarWidth);
  sidebarCollapsed = $state<boolean>(DEFAULT_SETTINGS.sidebarCollapsed);
  selectedMicrophoneDeviceId = $state<string>(DEFAULT_SETTINGS.selectedMicrophoneDeviceId);
  dictationShortcut = $state<string>(DEFAULT_SETTINGS.dictationShortcut);
  dictationMuteSystemAudio = $state<boolean>(DEFAULT_SETTINGS.dictationMuteSystemAudio);
  dictationHoldThresholdMs = $state<number>(DEFAULT_SETTINGS.dictationHoldThresholdMs);

  private saveTimeout: any = null;

  constructor() {
    this.loadFromStorage();
    if (typeof window !== 'undefined') {
      this.hydrateFromBackend();
      // Retry once after Wails IPC bindings are fully registered on window
      setTimeout(() => {
        if (!this.isHydrated) {
          this.hydrateFromBackend();
        }
      }, 100);
    }
  }

  async hydrateFromBackend(): Promise<void> {
    const win = typeof window !== 'undefined' ? (window as any) : null;
    if (win?.go?.main?.App?.GetAppSettings) {
      try {
        const backendSettings = await win.go.main.App.GetAppSettings();
        if (backendSettings && typeof backendSettings === 'object') {
          // If backend has settings, apply them
          this.applySettings(backendSettings);

          // Auto-detect Grok binary if empty or still pointing to invalid cross-platform path
          const isWindows = typeof navigator !== 'undefined' && /Win/.test(navigator.platform || navigator.userAgent);
          if (!this.grokBinaryPath || (isWindows && this.grokBinaryPath.startsWith('/Users/'))) {
            if (win.go?.main?.App?.AutoDetectGrokBinaryPath) {
              const detected = await win.go.main.App.AutoDetectGrokBinaryPath();
              if (detected) {
                this.grokBinaryPath = detected;
              }
            } else if (win.go?.main?.App?.CheckGrokInstallation) {
              const status = await win.go.main.App.CheckGrokInstallation();
              if (status?.binaryPath) {
                this.grokBinaryPath = status.binaryPath;
              }
            }
          }

          this.isHydrated = true;
          logger.info('SETTINGS', 'Hydrated settings from persistent backend storage');

          // Sync hydrated shortcuts to OS hooks
          if (win?.go?.main?.App?.RegisterGlobalSnapshotShortcut && this.snapshotShortcut) {
            win.go.main.App.RegisterGlobalSnapshotShortcut(this.snapshotShortcut).catch(() => {});
          }
          if (win?.go?.main?.App?.RegisterGlobalDictationShortcut && this.dictationShortcut) {
            win.go.main.App.RegisterGlobalDictationShortcut(this.dictationShortcut).catch(() => {});
          }
          return;
        }
      } catch (err) {
        logger.error('SETTINGS', 'Failed to hydrate settings from backend, using local fallback', err);
      }
    }

    // Auto-migrate from localStorage if backend was empty
    this.saveToStorage();
    this.isHydrated = true;
  }

  private applySettings(parsed: Partial<AppSettings>): void {
    const isWindows = typeof navigator !== 'undefined' && /Win/.test(navigator.platform || navigator.userAgent);
    if (parsed.theme && ['dark-studio', 'dark-high-contrast', 'light-antd'].includes(parsed.theme)) {
      this.theme = parsed.theme;
    }
    if (parsed.defaultModel) {
      this.defaultModel = parsed.defaultModel;
    }
    if (parsed.defaultReasoningEffort && ['none', 'low', 'medium', 'high', 'max'].includes(parsed.defaultReasoningEffort)) {
      this.defaultReasoningEffort = parsed.defaultReasoningEffort;
    }
    if (parsed.permissionMode && ['default', 'acceptEdits', 'auto', 'plan', 'bypassPermissions'].includes(parsed.permissionMode)) {
      this.permissionMode = parsed.permissionMode;
    }
    if (parsed.planGateMode && ['active', 'bypass'].includes(parsed.planGateMode)) {
      this.planGateMode = parsed.planGateMode;
    }
    if (typeof parsed.animationsEnabled === 'boolean') {
      this.animationsEnabled = parsed.animationsEnabled;
    }
    if (typeof parsed.grokBinaryPath === 'string') {
      if (isWindows && parsed.grokBinaryPath.startsWith('/Users/')) {
        this.grokBinaryPath = '';
      } else {
        this.grokBinaryPath = parsed.grokBinaryPath;
      }
    }
    if (typeof parsed.snapshotShortcut === 'string' && parsed.snapshotShortcut.trim().length > 0) {
      this.snapshotShortcut = parsed.snapshotShortcut;
    }
    if (typeof parsed.snapshotDelayMs === 'number' && Number.isFinite(parsed.snapshotDelayMs)) {
      this.snapshotDelayMs = parsed.snapshotDelayMs;
    }
    if (typeof parsed.snapshotAutoHideWindow === 'boolean') {
      this.snapshotAutoHideWindow = parsed.snapshotAutoHideWindow;
    }
    if (typeof parsed.snapshotSoundEnabled === 'boolean') {
      this.snapshotSoundEnabled = parsed.snapshotSoundEnabled;
    }
    if (typeof parsed.snapshotFlashEnabled === 'boolean') {
      this.snapshotFlashEnabled = parsed.snapshotFlashEnabled;
    }
    if (typeof parsed.snapshotAutoAttach === 'boolean') {
      this.snapshotAutoAttach = parsed.snapshotAutoAttach;
    }
    if (typeof parsed.activeWindowTurnCount === 'number' && Number.isFinite(parsed.activeWindowTurnCount)) {
      this.activeWindowTurnCount = parsed.activeWindowTurnCount;
    }
    if (typeof parsed.maxContextTokens === 'number' && Number.isFinite(parsed.maxContextTokens)) {
      this.maxContextTokens = parsed.maxContextTokens;
    }
    if (typeof parsed.sidebarWidth === 'number' && Number.isFinite(parsed.sidebarWidth)) {
      this.sidebarWidth = Math.min(480, Math.max(220, parsed.sidebarWidth));
    }
    if (typeof parsed.sidebarCollapsed === 'boolean') {
      this.sidebarCollapsed = parsed.sidebarCollapsed;
    }
    if (typeof parsed.selectedMicrophoneDeviceId === 'string') {
      this.selectedMicrophoneDeviceId = parsed.selectedMicrophoneDeviceId;
    }
    if (typeof parsed.dictationShortcut === 'string' && parsed.dictationShortcut.trim().length > 0) {
      this.dictationShortcut = parsed.dictationShortcut;
    }
    if (typeof parsed.dictationMuteSystemAudio === 'boolean') {
      this.dictationMuteSystemAudio = parsed.dictationMuteSystemAudio;
    }
    if (typeof parsed.dictationHoldThresholdMs === 'number' && Number.isFinite(parsed.dictationHoldThresholdMs) && parsed.dictationHoldThresholdMs > 0) {
      this.dictationHoldThresholdMs = parsed.dictationHoldThresholdMs;
    }
  }

  loadFromStorage(): void {
    if (typeof window === 'undefined' || !window.localStorage) return;
    try {
      const raw = window.localStorage.getItem(STORAGE_KEY);
      if (!raw) return;
      const parsed = JSON.parse(raw) as Partial<AppSettings>;
      this.applySettings(parsed);
    } catch (err) {
      console.warn('Failed to load settings from localStorage:', err);
    }
  }

  saveToStorage(): void {
    const data: AppSettings = {
      theme: this.theme,
      defaultModel: this.defaultModel,
      defaultReasoningEffort: this.defaultReasoningEffort,
      permissionMode: this.permissionMode,
      planGateMode: this.planGateMode,
      animationsEnabled: this.animationsEnabled,
      grokBinaryPath: this.grokBinaryPath,
      snapshotShortcut: this.snapshotShortcut,
      snapshotDelayMs: this.snapshotDelayMs,
      snapshotAutoHideWindow: this.snapshotAutoHideWindow,
      snapshotSoundEnabled: this.snapshotSoundEnabled,
      snapshotFlashEnabled: this.snapshotFlashEnabled,
      snapshotAutoAttach: this.snapshotAutoAttach,
      activeWindowTurnCount: this.activeWindowTurnCount,
      maxContextTokens: this.maxContextTokens,
      sidebarWidth: this.sidebarWidth,
      sidebarCollapsed: this.sidebarCollapsed,
      selectedMicrophoneDeviceId: this.selectedMicrophoneDeviceId,
      dictationShortcut: this.dictationShortcut,
      dictationMuteSystemAudio: this.dictationMuteSystemAudio,
      dictationHoldThresholdMs: this.dictationHoldThresholdMs
    };

    // 1. Fallback save to localStorage for offline cache
    if (typeof window !== 'undefined' && window.localStorage) {
      try {
        window.localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
      } catch (e) {}
    }

    // 2. Persist to Go backend storage
    const win = typeof window !== 'undefined' ? (window as any) : null;
    if (win?.go?.main?.App?.SaveAppSettings) {
      clearTimeout(this.saveTimeout);
      this.saveTimeout = setTimeout(() => {
        win.go.main.App.SaveAppSettings(data).catch((err: any) => {
          logger.error('SETTINGS', 'Failed to save settings to backend storage', err);
        });
      }, 50);
    }
  }

  updateSettings(partial: Partial<AppSettings>): void {
    if (partial.theme !== undefined) this.theme = partial.theme;
    if (partial.defaultModel !== undefined) this.defaultModel = partial.defaultModel;
    if (partial.defaultReasoningEffort !== undefined) this.defaultReasoningEffort = partial.defaultReasoningEffort;
    if (partial.permissionMode !== undefined) this.permissionMode = partial.permissionMode;
    if (partial.planGateMode !== undefined) this.planGateMode = partial.planGateMode;
    if (partial.animationsEnabled !== undefined) this.animationsEnabled = partial.animationsEnabled;
    if (partial.grokBinaryPath !== undefined) this.grokBinaryPath = partial.grokBinaryPath;
    if (partial.snapshotShortcut !== undefined) this.snapshotShortcut = partial.snapshotShortcut;
    if (partial.snapshotDelayMs !== undefined) this.snapshotDelayMs = partial.snapshotDelayMs;
    if (partial.snapshotAutoHideWindow !== undefined) this.snapshotAutoHideWindow = partial.snapshotAutoHideWindow;
    if (partial.snapshotSoundEnabled !== undefined) this.snapshotSoundEnabled = partial.snapshotSoundEnabled;
    if (partial.snapshotFlashEnabled !== undefined) this.snapshotFlashEnabled = partial.snapshotFlashEnabled;
    if (partial.snapshotAutoAttach !== undefined) this.snapshotAutoAttach = partial.snapshotAutoAttach;
    if (partial.activeWindowTurnCount !== undefined) this.activeWindowTurnCount = partial.activeWindowTurnCount;
    if (partial.maxContextTokens !== undefined) this.maxContextTokens = partial.maxContextTokens;
    if (partial.sidebarWidth !== undefined) this.sidebarWidth = Math.min(480, Math.max(220, partial.sidebarWidth));
    if (partial.sidebarCollapsed !== undefined) this.sidebarCollapsed = partial.sidebarCollapsed;
    if (partial.selectedMicrophoneDeviceId !== undefined) this.selectedMicrophoneDeviceId = partial.selectedMicrophoneDeviceId;
    if (partial.dictationShortcut !== undefined) this.dictationShortcut = partial.dictationShortcut;
    if (partial.dictationMuteSystemAudio !== undefined) this.dictationMuteSystemAudio = partial.dictationMuteSystemAudio;
    if (partial.dictationHoldThresholdMs !== undefined) this.dictationHoldThresholdMs = partial.dictationHoldThresholdMs;
    this.saveToStorage();
  }

  resetToDefaults(): void {
    this.theme = DEFAULT_SETTINGS.theme;
    this.defaultModel = DEFAULT_SETTINGS.defaultModel;
    this.defaultReasoningEffort = DEFAULT_SETTINGS.defaultReasoningEffort;
    this.permissionMode = DEFAULT_SETTINGS.permissionMode;
    this.planGateMode = DEFAULT_SETTINGS.planGateMode;
    this.animationsEnabled = DEFAULT_SETTINGS.animationsEnabled;
    this.grokBinaryPath = DEFAULT_SETTINGS.grokBinaryPath;
    this.snapshotShortcut = DEFAULT_SETTINGS.snapshotShortcut;
    this.snapshotDelayMs = DEFAULT_SETTINGS.snapshotDelayMs;
    this.snapshotAutoHideWindow = DEFAULT_SETTINGS.snapshotAutoHideWindow;
    this.snapshotSoundEnabled = DEFAULT_SETTINGS.snapshotSoundEnabled;
    this.snapshotFlashEnabled = DEFAULT_SETTINGS.snapshotFlashEnabled;
    this.snapshotAutoAttach = DEFAULT_SETTINGS.snapshotAutoAttach;
    this.activeWindowTurnCount = DEFAULT_SETTINGS.activeWindowTurnCount;
    this.maxContextTokens = DEFAULT_SETTINGS.maxContextTokens;
    this.sidebarWidth = DEFAULT_SETTINGS.sidebarWidth;
    this.sidebarCollapsed = DEFAULT_SETTINGS.sidebarCollapsed;
    this.selectedMicrophoneDeviceId = DEFAULT_SETTINGS.selectedMicrophoneDeviceId;
    this.dictationShortcut = DEFAULT_SETTINGS.dictationShortcut;
    this.dictationMuteSystemAudio = DEFAULT_SETTINGS.dictationMuteSystemAudio;
    this.dictationHoldThresholdMs = DEFAULT_SETTINGS.dictationHoldThresholdMs;

    const win = typeof window !== 'undefined' ? (window as any) : null;
    if (win?.go?.main?.App?.AutoDetectGrokBinaryPath) {
      win.go.main.App.AutoDetectGrokBinaryPath().then((detected: string) => {
        if (detected) this.grokBinaryPath = detected;
      }).catch(() => {});
    }

    this.saveToStorage();
  }
}

export const settingsStore = new SettingsStore();
