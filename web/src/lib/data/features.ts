export interface FeatureItem {
  id: string;
  title: string;
  badge: string;
  tagline: string;
  description: string;
  bullets: string[];
  screenshot: string;
  imageAlt: string;
}

export const featuresData: FeatureItem[] = [
  {
    id: 'windowing',
    title: '10-Turn DOM Windowing',
    badge: 'Performance Engine',
    tagline: 'Infinite scroll with zero memory bloat or UI lag.',
    description:
      'Long autonomous workflows create hundreds of turns with complex syntax-highlighted diffs and stdout traces. AetherGrok dynamically virtualizes the DOM, retaining only the active 10 conversational turns in memory while preserving fluid scroll anchor locks.',
    bullets: [
      'Constant ~35 MB baseline memory footprint regardless of conversation length',
      'Smooth 60 FPS viewport scrolling without frame skips or garbage collection stalls',
      'Frame-aligned 16ms NDJSON token batching eliminates streaming micro-stutters',
    ],
    screenshot: './screenshots/desktop-hero.webp',
    imageAlt: 'AetherGrok Desktop Workspace and 10-turn windowing engine',
  },
  {
    id: 'vision',
    title: 'Compositor-Synced Vision Capture',
    badge: 'Vision & Multimodal',
    tagline: 'Non-intrusive full-screen capture without capturing itself.',
    description:
      'Capture any window, bug, or web design instantly. AetherGrok temporarily yields to the OS compositor (50ms on macOS, 80ms on Windows) to snap your entire desktop display natively, then refocuses and attaches the screenshot directly into your prompt chip bar.',
    bullets: [
      'Global shortcut Cmd+Alt+S / Ctrl+Alt+S with instant composer chip injection',
      'Guaranteed panic-safe window restoration with deferred OS handlers',
      'Zero manual cropping or external screenshot tool steps required',
    ],
    screenshot: './screenshots/tool_calls.png',
    imageAlt: 'AetherGrok Screen Capture and Tool Calls Visualization',
  },
  {
    id: 'voice',
    title: 'Push-to-Talk Voice & Audio Ducking',
    badge: 'Audio Ergonomics',
    tagline: 'Speak directly to your agent while music automatically fades.',
    description:
      'Engineered with a real-time 4-bar dynamic audio volume equalizer. Hold down your shortcut or double-tap to speak. Active system audio playback is automatically muted during speech recognition and seamlessly restored upon completion.',
    bullets: [
      'Hardware-accelerated native audio worklet with zero transcription delay',
      'Push-to-talk hold and double-tap toggle options with custom key binding',
      'Automatic system volume ducking to eliminate background audio interference',
    ],
    screenshot: './screenshots/voice_mode.png',
    imageAlt: 'AetherGrok Push-to-Talk Voice Mode with Dynamic Equalizer',
  },
  {
    id: 'diffs',
    title: 'Rich Syntax-Highlighted Diffs',
    badge: 'Code Review Studio',
    tagline: 'Audit agent modifications with precision side-by-side reviews.',
    description:
      'Say goodbye to flat ANSI terminal text. Review file additions, deletions, and inline modifications with side-by-side diff views, syntax coloring, line numbering, and one-click accept/rejection controls.',
    bullets: [
      'Interactive Diff2Html engine with side-by-side and unified viewing modes',
      'Collapsible diff cards with file mutation statistics (+N / -N lines)',
      'Direct integration with git worktrees and multi-session isolation',
    ],
    screenshot: './screenshots/permission_diff.png',
    imageAlt: 'AetherGrok Side-by-Side Unified Code Diff Viewer',
  },
  {
    id: 'skills',
    title: 'Universal Skills Hub Catalog',
    badge: 'Ecosystem',
    tagline: 'Discover, inspect, and invoke skills with slash commands.',
    description:
      'Automatic scanning of local skill repositories (`~/.grok/skills/` and `~/.agents/skills/`). Browse installed capabilities, read YAML frontmatter definitions, and inject `/skill-name` shortcuts directly into your composer.',
    bullets: [
      'Universal discovery across Grok, Claude Code, and autonomous agent registries',
      'Interactive autocomplete with parameter descriptions and usage tags',
      '1-click skill importer and dependency executor with safe sandboxing',
    ],
    screenshot: './screenshots/agent_modes.png',
    imageAlt: 'AetherGrok Skills Hub and Agent Execution Modes',
  },
];
