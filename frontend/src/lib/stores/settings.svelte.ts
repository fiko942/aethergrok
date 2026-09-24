export type ThemeMode = 'dark-studio' | 'dark-high-contrast' | 'light-antd';
export type DefaultModel = 'grok-4.6' | 'grok-code' | 'custom' | string;
export type ReasoningEffort = 'none' | 'low' | 'medium' | 'high' | 'max';
export type PermissionMode = 'default' | 'acceptEdits' | 'auto' | 'plan' | 'bypassPermissions';

export interface AppSettings {
  theme: ThemeMode;
  defaultModel: DefaultModel;
  defaultReasoningEffort: ReasoningEffort;
  permissionMode: PermissionMode;
  grokBinaryPath: string;
  snapshotDelayMs: number;
  snapshotSoundEnabled: boolean;
  snapshotFlashEnabled: boolean;
  snapshotAutoAttach: boolean;
  activeWindowTurnCount: number;
}

const STORAGE_KEY = 'aethergrok_settings_v1';

export const DEFAULT_SETTINGS: AppSettings = {
  theme: 'dark-studio',
  defaultModel: '9router',
  defaultReasoningEffort: 'medium',
  permissionMode: 'default',
  grokBinaryPath: '/Users/fiko942/.local/bin/grok',
  snapshotDelayMs: 50,
  snapshotSoundEnabled: true,
  snapshotFlashEnabled: true,
  snapshotAutoAttach: true,
  activeWindowTurnCount: 10
};

export class SettingsStore {
  theme = $state<ThemeMode>(DEFAULT_SETTINGS.theme);
  defaultModel = $state<DefaultModel>(DEFAULT_SETTINGS.defaultModel);
  defaultReasoningEffort = $state<ReasoningEffort>(DEFAULT_SETTINGS.defaultReasoningEffort);
  permissionMode = $state<PermissionMode>(DEFAULT_SETTINGS.permissionMode);
  grokBinaryPath = $state<string>(DEFAULT_SETTINGS.grokBinaryPath);
  snapshotDelayMs = $state<number>(DEFAULT_SETTINGS.snapshotDelayMs);
  snapshotSoundEnabled = $state<boolean>(DEFAULT_SETTINGS.snapshotSoundEnabled);
  snapshotFlashEnabled = $state<boolean>(DEFAULT_SETTINGS.snapshotFlashEnabled);
  snapshotAutoAttach = $state<boolean>(DEFAULT_SETTINGS.snapshotAutoAttach);
  activeWindowTurnCount = $state<number>(DEFAULT_SETTINGS.activeWindowTurnCount);

  constructor() {
    this.loadFromStorage();
  }

  loadFromStorage(): void {
    if (typeof window === 'undefined' || !window.localStorage) return;
    try {
      const raw = window.localStorage.getItem(STORAGE_KEY);
      if (!raw) return;
      const parsed = JSON.parse(raw) as Partial<AppSettings>;

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
      if (typeof parsed.grokBinaryPath === 'string') {
        this.grokBinaryPath = parsed.grokBinaryPath;
      }
      if (typeof parsed.snapshotDelayMs === 'number' && Number.isFinite(parsed.snapshotDelayMs)) {
        this.snapshotDelayMs = parsed.snapshotDelayMs;
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
    } catch (err) {
      console.warn('Failed to load settings from localStorage:', err);
    }
  }

  saveToStorage(): void {
    if (typeof window === 'undefined' || !window.localStorage) return;
    try {
      const data: AppSettings = {
        theme: this.theme,
        defaultModel: this.defaultModel,
        defaultReasoningEffort: this.defaultReasoningEffort,
        permissionMode: this.permissionMode,
        grokBinaryPath: this.grokBinaryPath,
        snapshotDelayMs: this.snapshotDelayMs,
        snapshotSoundEnabled: this.snapshotSoundEnabled,
        snapshotFlashEnabled: this.snapshotFlashEnabled,
        snapshotAutoAttach: this.snapshotAutoAttach,
        activeWindowTurnCount: this.activeWindowTurnCount
      };
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
    } catch (err) {
      console.warn('Failed to save settings to localStorage:', err);
    }
  }

  updateSettings(partial: Partial<AppSettings>): void {
    if (partial.theme !== undefined) this.theme = partial.theme;
    if (partial.defaultModel !== undefined) this.defaultModel = partial.defaultModel;
    if (partial.defaultReasoningEffort !== undefined) this.defaultReasoningEffort = partial.defaultReasoningEffort;
    if (partial.permissionMode !== undefined) this.permissionMode = partial.permissionMode;
    if (partial.grokBinaryPath !== undefined) this.grokBinaryPath = partial.grokBinaryPath;
    if (partial.snapshotDelayMs !== undefined) this.snapshotDelayMs = partial.snapshotDelayMs;
    if (partial.snapshotSoundEnabled !== undefined) this.snapshotSoundEnabled = partial.snapshotSoundEnabled;
    if (partial.snapshotFlashEnabled !== undefined) this.snapshotFlashEnabled = partial.snapshotFlashEnabled;
    if (partial.snapshotAutoAttach !== undefined) this.snapshotAutoAttach = partial.snapshotAutoAttach;
    if (partial.activeWindowTurnCount !== undefined) this.activeWindowTurnCount = partial.activeWindowTurnCount;
    this.saveToStorage();
  }

  resetToDefaults(): void {
    this.theme = DEFAULT_SETTINGS.theme;
    this.defaultModel = DEFAULT_SETTINGS.defaultModel;
    this.defaultReasoningEffort = DEFAULT_SETTINGS.defaultReasoningEffort;
    this.permissionMode = DEFAULT_SETTINGS.permissionMode;
    this.grokBinaryPath = DEFAULT_SETTINGS.grokBinaryPath;
    this.snapshotDelayMs = DEFAULT_SETTINGS.snapshotDelayMs;
    this.snapshotSoundEnabled = DEFAULT_SETTINGS.snapshotSoundEnabled;
    this.snapshotFlashEnabled = DEFAULT_SETTINGS.snapshotFlashEnabled;
    this.snapshotAutoAttach = DEFAULT_SETTINGS.snapshotAutoAttach;
    this.activeWindowTurnCount = DEFAULT_SETTINGS.activeWindowTurnCount;
    this.saveToStorage();
  }
}

export const settingsStore = new SettingsStore();
