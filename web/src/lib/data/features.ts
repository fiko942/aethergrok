export interface FeatureItem {
  id: string;
  title: string;
  badge: string;
  category: string;
  tagline: string;
  description: string;
  bullets: string[];
  techSpec: string;
}

export const featuresData: FeatureItem[] = [
  {
    id: 'windowing',
    title: '10-Turn DOM Windowing',
    badge: 'Virtualization Engine',
    category: 'Performance',
    tagline: 'Infinite session length with zero memory bloat or UI lag.',
    description:
      'Long autonomous workflows create hundreds of turns with complex syntax-highlighted diffs and terminal outputs. AetherGrok dynamically virtualizes the DOM tree, retaining only the active 10 conversational turns in memory while preserving fluid scroll anchor locks.',
    bullets: [
      'Constant ~35 MB baseline memory footprint across 100+ turns',
      'Smooth 60 FPS viewport scrolling without frame skips or GC pauses',
      'Frame-aligned 16ms NDJSON token batching eliminates streaming micro-stutters',
    ],
    techSpec: 'Go 1.24 Concurrent Parser • Svelte 5 Virtual Windowing',
  },
  {
    id: 'process-group',
    title: 'Process Group Supervision',
    badge: 'Process Isolation',
    category: 'Reliability',
    tagline: 'Zero orphaned background tasks and mid-turn live steering.',
    description:
      'Standard subprocess runners frequently leave background bash commands orphaned when turns are cancelled. AetherGrok binds CLI workers with POSIX process groups (setpgid) and Windows Job Objects, ensuring complete sub-process lifecycle control and instantaneous steering injection.',
    bullets: [
      'POSIX setpgid and Windows Job Objects process tree containment',
      'Live mid-turn prompt steering without abrupt process termination',
      'Robust graceful kill handling with escalating signals (SIGTERM / SIGKILL)',
    ],
    techSpec: 'POSIX setpgid • Windows Job Objects • SysProcAttr',
  },
  {
    id: 'vision',
    title: 'Compositor-Synced Vision Capture',
    badge: 'Multimodal Input',
    category: 'Ergonomics',
    tagline: 'Non-intrusive full-screen capture without capturing itself.',
    description:
      'Capture any window, UI glitch, or browser layout instantly. AetherGrok temporarily yields to the OS compositor (50ms on macOS, 80ms on Windows) to snap your entire desktop display natively, then refocuses and attaches the screenshot directly into your prompt chip bar.',
    bullets: [
      'Global shortcut Cmd+Alt+S / Ctrl+Alt+S with instant composer chip injection',
      'Guaranteed panic-safe window restoration with deferred OS handlers',
      'Direct multimodal vision support for Grok 4.6 and image-to-code workflows',
    ],
    techSpec: 'macOS screencapture • Windows GDI32 • CoreGraphics',
  },
  {
    id: 'voice',
    title: 'Push-to-Talk Voice & Audio Ducking',
    badge: 'Voice Dictation',
    category: 'Input Audio',
    tagline: 'Speak directly to your agent while music automatically fades.',
    description:
      'Engineered with a real-time 4-bar dynamic audio volume equalizer. Hold down your shortcut or double-tap to speak. Active system audio playback is automatically muted to 0% during speech recording and seamlessly restored upon completion.',
    bullets: [
      'Hardware-accelerated native audio worklet with zero transcription lag',
      'Configurable push-to-talk hold and double-tap toggle keyboard shortcuts',
      'Automatic system volume ducking to eliminate background audio interference',
    ],
    techSpec: 'CoreAudio / AppleScript • Windows CoreAudio API',
  },
  {
    id: 'diffs',
    title: 'Rich Syntax-Highlighted Diffs',
    badge: 'Code Review Studio',
    category: 'Code Quality',
    tagline: 'Audit agent modifications with precision side-by-side reviews.',
    description:
      'Say goodbye to flat ANSI terminal text. Review file additions, deletions, and inline modifications with side-by-side diff views, syntax coloring, line numbering, and one-click accept/rejection controls.',
    bullets: [
      'Interactive Diff2Html engine with side-by-side and unified viewing modes',
      'Collapsible diff cards with file mutation statistics (+N / -N lines)',
      'Direct integration with git worktrees and multi-session isolation',
    ],
    techSpec: 'Diff2Html • PrismJS Syntax Tokenizer • KaTeX Expressions',
  },
  {
    id: 'toolchain',
    title: '1-Click Toolchain Setup Gate',
    badge: 'Environment Setup',
    category: 'Bootstrapping',
    tagline: 'Automatic Grok CLI detection and background installation.',
    description:
      'Never suffer from missing CLI toolchain errors. On initial launch, AetherGrok inspects your system PATH and local binaries. If Grok Build is missing, it provides a 1-click installer gate executing Homebrew or native shell scripts directly from the UI.',
    bullets: [
      'Automatic PATH and ~/.grok/bin binary verification on startup',
      'Integrated 1-click Homebrew and curl installation pipeline with live log output',
      'Zero manual terminal bootstrapping required for new developers',
    ],
    techSpec: 'Homebrew CLI API • Shell Subprocess Runner • Live Diagnostics',
  },
];
