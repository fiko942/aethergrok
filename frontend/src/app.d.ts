/// <reference types="svelte" />
/// <reference types="vite/client" />

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
        };
      };
    };
  }
}

export {};
