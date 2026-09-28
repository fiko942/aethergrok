export interface ComparisonRow {
  metric: string;
  aethergrok: string;
  claudeCode: string;
  codexCli: string;
  antigravityCua: string;
  highlight?: boolean;
}

export const comparisonData: ComparisonRow[] = [
  {
    metric: 'Primary Interaction Mode',
    aethergrok: 'Native Desktop Studio (Wails v2 + Svelte 5)',
    claudeCode: 'Terminal REPL (Node.js CLI)',
    codexCli: 'Terminal CLI / ACP Wrapper',
    antigravityCua: 'Desktop & Web Automation Agent',
    highlight: true,
  },
  {
    metric: 'Interface Style',
    aethergrok: 'Rich Graphical UI with Visual Diffs & Controls',
    claudeCode: 'Terminal Text Stream',
    codexCli: 'Terminal Text Stream',
    antigravityCua: 'Visual Action Sandbox & Screen Capture',
  },
  {
    metric: 'Runtime Memory Footprint',
    aethergrok: '~35 MB RAM (Go 1.24 + OS Webview)',
    claudeCode: '~180 MB RAM (Node.js runtime)',
    codexCli: '~140 MB RAM (Python / Node.js)',
    antigravityCua: '~450+ MB RAM (Electron / PyAutoGUI)',
    highlight: true,
  },
  {
    metric: 'Long-Session Responsiveness',
    aethergrok: '10-Turn Virtual DOM Windowing (Zero Lag)',
    claudeCode: 'Terminal Scrollback Buffer Limit',
    codexCli: 'Terminal Buffer Management',
    antigravityCua: 'DOM & Action Graph Re-render Load',
    highlight: true,
  },
  {
    metric: 'Mid-Turn Steerability',
    aethergrok: 'Live In-Flight Message Injection & Steer Queue',
    claudeCode: 'Requires Turn Interrupt / Re-prompt',
    codexCli: 'Supported via ACP Protocol',
    antigravityCua: 'Action Queue Adjustment',
  },
  {
    metric: 'Subprocess Lifecycle Management',
    aethergrok: 'POSIX setpgid & Windows Job Objects Isolation',
    claudeCode: 'Standard Node.js child_process Tree',
    codexCli: 'System Subprocess Calls',
    antigravityCua: 'OS Accessibility & Automation APIs',
  },
  {
    metric: 'Visual Screen Context (Vision)',
    aethergrok: 'Non-Intrusive Compositor-Synced Screen Capture',
    claudeCode: 'Manual File Path Attachment',
    codexCli: 'Manual File Path Attachment',
    antigravityCua: 'Continuous Screen Frame Buffering',
  },
  {
    metric: 'Push-to-Talk Voice Dictation',
    aethergrok: 'Integrated Equalizer & Auto-Mute System Audio',
    claudeCode: 'Host Terminal Audio Dependent',
    codexCli: 'Host Terminal Audio Dependent',
    antigravityCua: 'Host Environment Dependent',
  },
  {
    metric: 'Diff Presentation',
    aethergrok: 'Syntax-Highlighted Side-by-Side & Unified Diffs',
    claudeCode: 'Terminal ANSI Color Diffs',
    codexCli: 'Terminal ANSI Color Diffs',
    antigravityCua: 'Visual Screenshot & File Inspection',
  },
  {
    metric: 'Ecosystem & Skills Registry',
    aethergrok: 'Visual Skills Hub (~/.grok/skills & ~/.agents)',
    claudeCode: 'CLI Slash Commands & Markdown Rules',
    codexCli: 'Configuration File Hooks & APIs',
    antigravityCua: 'Pre-recorded Automation DAGs',
  },
  {
    metric: 'Installation & Toolchain Gate',
    aethergrok: '1-Click Built-in Setup & Auto-Detection',
    claudeCode: 'Manual npm install -g @anthropic-ai/claude-code',
    codexCli: 'Manual Package Setup / Python env',
    antigravityCua: 'Manual Environment Setup / Docker',
  },
];
