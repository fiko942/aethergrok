/// <reference types="svelte" />
/// <reference types="vite/client" />

declare global {
  const __APP_VERSION__: string;
}

export interface SnapshotResult {
  filePath: string;
  dataUrl: string;
  base64?: string;
  sizeBytes?: number;
  width?: number;
  height?: number;
  timestamp: number;
}

export interface PromptRequestPayload {
  sessionId: string;
  prompt: string;
  images?: string[];
  options?: {
    model?: string;
    reasoningEffort?: string;
    workingDir?: string;
    skillDirs?: string[];
    temperature?: number;
    disableTools?: boolean;
    systemPrompt?: string;
    customFlags?: string[];
  };
}

export interface PermissionResponsePayload {
  sessionId: string;
  requestId: string;
  decision: 'allow_once' | 'allow_always' | 'reject';
}

export interface SkillItem {
  id: string;
  name: string;
  description: string;
  category: 'All' | 'Frontend' | 'Backend' | 'Design' | 'Agents' | 'Tools';
  tags?: string[];
  actions?: string[];
  path: string;
  directory: string;
  scope: string;
  prompt?: string;
}

export interface DiscoveredSkill {
  name: string;
  description: string;
  category: string;
  tags?: string[];
  relativePath: string;
  skillFile: string;
  prereqs: string[];
  commands: string[];
}

export interface SkillAnalysisResult {
  repoUrl: string;
  repoName: string;
  tempPath: string;
  skills: DiscoveredSkill[];
  globalPrereqs: string[];
  suggestedScripts: string[];
}

export interface SkillInstallPayload {
  tempPath: string;
  skillPaths: string[];
  targetScope: 'grok' | 'agents';
}

export interface SkillInstallResult {
  success: boolean;
  installedCount: number;
  installedPaths: string[];
  errors?: string[];
}

declare global {
  interface Window {
    runtime?: {
      EventsOn: (eventName: string, callback: (...args: any[]) => void) => () => void;
      EventsOnce: (eventName: string, callback: (...args: any[]) => void) => () => void;
      EventsOnMultiple: (eventName: string, callback: (...args: any[]) => void, maxCallbacks: number) => () => void;
      EventsEmit: (eventName: string, ...args: any[]) => void;
      WindowMinimise: () => void;
      WindowMaximise: () => void;
      WindowUnmaximise: () => void;
      WindowToggleMaximise: () => void;
      WindowFullscreen: () => void;
      WindowUnfullscreen: () => void;
      WindowIsFullscreen: () => Promise<boolean>;
      WindowHide: () => void;
      WindowShow: () => void;
      WindowClose: () => void;
      Quit: () => void;
      BrowserOpenURL: (url: string) => void;
    };
    go?: {
      main?: {
        App?: {
          Greet: (name: string) => Promise<string>;
          OpenExternalURL: (targetURL: string) => Promise<void>;
          RunPromptStream: (req: PromptRequestPayload) => Promise<void>;
          RespondPermission: (resp: PermissionResponsePayload) => Promise<void>;
          CancelSession: (sessionId: string) => Promise<void>;
          SetGrokBinaryPath: (path: string) => Promise<void>;
          CaptureScreenExcludingSelf: (delayMs: number) => Promise<SnapshotResult>;
          RegisterGlobalSnapshotShortcut?: (shortcutStr: string) => Promise<void>;
          UnregisterGlobalSnapshotShortcut?: () => Promise<void>;
          GetInstalledSkills: () => Promise<SkillItem[]>;
          SearchSkills: (query: string, category: string) => Promise<SkillItem[]>;
          SelectWorkspaceDirectory: () => Promise<string>;
          CheckDirectoryExists: (dirPath: string) => Promise<boolean>;
          SaveMarkdownExport: (defaultFilename: string, content: string) => Promise<string>;
          GetAvailableModels: () => Promise<Array<{ id: string; name: string; description: string; isDefault: boolean }>>;
          // Terminal & Process Management APIs
          CreateTerminal: (sessionId: string, termId: string, cwd: string, shell: string) => Promise<void>;
          WriteTerminal: (termId: string, data: string) => Promise<void>;
          ResizeTerminal: (termId: string, cols: number, rows: number) => Promise<void>;
          CloseTerminal: (termId: string) => Promise<void>;
          CloseSessionTerminals: (sessionId: string) => Promise<void>;
          GetPlanContent: (planPath: string) => Promise<string>;

          DiscoverGrokSessions: (workspacePath: string) => Promise<Array<{ id: string; title: string; createdAt: number; updatedAt: number }>>;
          LoadGrokSessionHistory: (workspacePath: string, sessionID: string) => Promise<Array<any>>;
          DeleteGrokSession: (workspacePath: string, sessionId: string) => Promise<void>;
          GetSessionUsage: (workspacePath: string, sessionID: string) => Promise<any>;
          CompactSession: (workspacePath: string, sessionID: string) => Promise<any>;
          ScanGitHubSkills: (repoURL: string) => Promise<SkillAnalysisResult>;
          InstallDiscoveredSkills: (payload: SkillInstallPayload) => Promise<SkillInstallResult>;
          CleanupSkillImportTemp: (tempPath: string) => Promise<void>;
          ExecuteSkillSetupCommand: (workDir: string, commandLine: string) => Promise<void>;
          RevertWorkspaceFiles: (workspacePath: string, filePaths: string[]) => Promise<void>;
          CheckAndRequestAccessibilityPermissions: () => Promise<{ granted: boolean; message: string; platform: string }>;
          OpenAccessibilitySettings: () => Promise<void>;
          CheckMicrophonePermission: () => Promise<{ granted: boolean; message: string; platform: string }>;
          RequestMicrophonePermission: () => Promise<{ granted: boolean; message: string; platform: string }>;
          OpenMicrophoneSettings: () => Promise<void>;
          SaveVoiceAudioRecording: (base64Data: string, ext: string) => Promise<string>;
          DeleteVoiceAudioRecording: (filePath: string) => Promise<void>;
          TranscribeAudioWithGrok: (workspacePath: string, audioFilePath: string) => Promise<string>;
          RevealGrokConfigFile?: () => Promise<void>;
          GetSnapshotCacheStats?: () => Promise<{ totalBytes: number; fileCount: number; formattedSize: string }>;
          ClearSnapshotCache?: () => Promise<{ freedBytes: number; deletedCount: number; formattedSize: string }>;
          SaveTemporaryImage?: (base64Data: string, mimeType: string) => Promise<{ filePath: string; dataUrl: string; base64: string; width: number; height: number; sizeBytes: number; timestamp: number }>;
          DeleteSessionTempFiles?: (filePaths: string[]) => Promise<void>;
        };
      };
    };
  }
}

export {};
