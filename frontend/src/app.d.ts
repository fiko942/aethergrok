/// <reference types="svelte" />
/// <reference types="vite/client" />

export interface SnapshotResult {
  filePath: string;
  dataUrl: string;
  base64: string;
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
      WindowHide: () => void;
      WindowShow: () => void;
      WindowClose: () => void;
      Quit: () => void;
    };
    go?: {
      main?: {
        App?: {
          Greet: (name: string) => Promise<string>;
          RunPromptStream: (req: PromptRequestPayload) => Promise<void>;
          RespondPermission: (resp: PermissionResponsePayload) => Promise<void>;
          CancelSession: (sessionId: string) => Promise<void>;
          SetGrokBinaryPath: (path: string) => Promise<void>;
          CaptureScreenExcludingSelf: (delayMs: number) => Promise<SnapshotResult>;
          GetInstalledSkills: () => Promise<SkillItem[]>;
          SearchSkills: (query: string, category: string) => Promise<SkillItem[]>;
        };
      };
    };
  }
}

export {};
