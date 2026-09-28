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
    id: 'studio',
    title: 'Native Studio Workspace',
    badge: 'Core Environment',
    tagline: 'Multi-session project management with zero memory bloat.',
    description:
      'Manage multiple active workspaces, project trees, and live sessions simultaneously. Built with Go 1.24 and Svelte 5 to guarantee constant memory usage and instantaneous tab switching.',
    bullets: [
      'Multi-project sidebar navigation with automatic workspace discovery',
      'Unified conversation view with 10-turn progressive DOM windowing',
      'Real-time token streaming with 16ms frame-rate batching',
    ],
    screenshot: './screenshots/desktop-hero.webp',
    imageAlt: 'AetherGrok Native Studio Workspace and multi-session interface',
  },
  {
    id: 'voice',
    title: 'Push-to-Talk Voice & Equalizer',
    badge: 'Audio Ergonomics',
    tagline: 'Custom keyboard shortcuts with system audio ducking.',
    description:
      'Configure custom push-to-talk hold and double-tap shortcuts. Features a real-time 4-bar dynamic audio volume equalizer that automatically mutes system audio playback during active speech recording.',
    bullets: [
      'Configurable custom keyboard shortcut trigger with double-tap support',
      'Real-time 4-bar dynamic audio volume visualizer',
      'Automatic system audio ducking to 0% during voice recording',
    ],
    screenshot: './screenshots/voice_mode.png',
    imageAlt: 'AetherGrok Push-to-Talk Voice Dictation and Shortcut Settings',
  },
  {
    id: 'permissions',
    title: 'macOS Security & Permissions Gate',
    badge: 'System Integration',
    tagline: 'Automated 1-click permission auditing and status monitoring.',
    description:
      'AetherGrok automatically audits required macOS TCC permissions (Screen Recording, Accessibility, and Microphone) with live 3-second status refreshes and 1-click system settings deep links.',
    bullets: [
      'Real-time system permission status detection with auto-refresh polling',
      'One-click direct links to macOS Privacy & Security Settings panels',
      'Clean light/dark mode UI tokens styled with Ant Design precision',
    ],
    screenshot: './screenshots/permission_diff.png',
    imageAlt: 'AetherGrok System Permissions and Security Gate Modal',
  },
  {
    id: 'updates',
    title: 'In-App Updates & Changelog',
    badge: 'Lifecycle Management',
    tagline: 'Automatic GitHub Release synchronization and SemVer detection.',
    description:
      'Stay current with in-app update checks connected directly to the official GitHub repository (`fiko942/aethergrok`). Inspect formatted release notes and download installers directly from the app.',
    bullets: [
      'Direct synchronization with official fiko942/aethergrok GitHub Releases',
      'In-app Markdown release notes and version history viewer',
      'One-click direct platform package downloads for macOS and Windows',
    ],
    screenshot: './screenshots/agent_modes.png',
    imageAlt: 'AetherGrok In-App Check for Updates and Version History Modal',
  },
];
