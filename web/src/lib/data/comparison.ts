export interface ComparisonRow {
  metric: string;
  category: string;
  aethergrok: {
    title: string;
    sub: string;
    badge: 'native' | 'featured' | 'standard';
  };
  claudeCode: {
    title: string;
    sub: string;
  };
  codexCli: {
    title: string;
    sub: string;
  };
  antigravityCua: {
    title: string;
    sub: string;
  };
}

export const comparisonCategories = [
  'Architecture & Runtime',
  'Interaction & Ergonomics',
  'Control & Lifecycle',
  'Installation & Toolchain',
];

export const comparisonData: ComparisonRow[] = [
  {
    metric: 'Primary Interaction Mode',
    category: 'Architecture & Runtime',
    aethergrok: {
      title: 'Native Desktop Studio',
      sub: 'Go 1.24 Core + Wails v2 + Svelte 5',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Terminal REPL',
      sub: 'Node.js CLI via stdout stream',
    },
    codexCli: {
      title: 'Terminal CLI / ACP',
      sub: 'Subprocess wrapper / JSON-RPC',
    },
    antigravityCua: {
      title: 'Desktop Automation Sandbox',
      sub: 'Electron / Python browser & OS agent',
    },
  },
  {
    metric: 'Runtime Memory Footprint',
    category: 'Architecture & Runtime',
    aethergrok: {
      title: '~35 MB Baseline RAM',
      sub: 'Native WebKit/WebView2 (zero Chromium overhead)',
      badge: 'featured',
    },
    claudeCode: {
      title: '~180 MB RAM',
      sub: 'V8 Node.js runtime process',
    },
    codexCli: {
      title: '~140 MB RAM',
      sub: 'Python runtime + interpreter',
    },
    antigravityCua: {
      title: '~450+ MB RAM',
      sub: 'Full Chromium / Electron browser process',
    },
  },
  {
    metric: 'Long-Session Performance',
    category: 'Architecture & Runtime',
    aethergrok: {
      title: '10-Turn DOM Virtualization',
      sub: 'Constant memory, fluid 60 FPS across 100+ turns',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Scrollback Buffer Bound',
      sub: 'Limited by terminal emulator memory buffer',
    },
    codexCli: {
      title: 'Stdout Terminal Stream',
      sub: 'Standard terminal text scrollback buffer',
    },
    antigravityCua: {
      title: 'Heavy Canvas Re-renders',
      sub: 'DOM & canvas overhead increases over time',
    },
  },
  {
    metric: 'Live Mid-Turn Steering',
    category: 'Interaction & Ergonomics',
    aethergrok: {
      title: 'In-Flight Prompt Injection',
      sub: 'Redirect execution mid-turn without killing workers',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Interrupt / Re-prompt',
      sub: 'Must cancel active turn prior to input',
    },
    codexCli: {
      title: 'ACP Protocol Method',
      sub: 'Limited support via RPC cancel/re-turn',
    },
    antigravityCua: {
      title: 'Queue Re-ordering',
      sub: 'Reorder pending UI automation action queue',
    },
  },
  {
    metric: 'Visual Screen Context (Vision)',
    category: 'Interaction & Ergonomics',
    aethergrok: {
      title: 'Compositor-Synced Snap',
      sub: '50ms auto-hide, full OS capture, auto prompt chip',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Manual Path Input',
      sub: 'Type local image path manually in terminal',
    },
    codexCli: {
      title: 'Manual Attachment',
      sub: 'Provide local file path reference',
    },
    antigravityCua: {
      title: 'Continuous Screen Stream',
      sub: 'Constant pixel streaming across frames',
    },
  },
  {
    metric: 'Voice Dictation & Media Ducking',
    category: 'Interaction & Ergonomics',
    aethergrok: {
      title: 'Push-to-Talk + Auto Mute',
      sub: 'Dynamic 4-bar equalizer + system audio mute to 0%',
      badge: 'featured',
    },
    claudeCode: {
      title: 'No Audio Features',
      sub: 'Relies on host terminal capabilities',
    },
    codexCli: {
      title: 'No Audio Features',
      sub: 'Relies on host terminal capabilities',
    },
    antigravityCua: {
      title: 'Host OS Dependent',
      sub: 'No integrated voice workflow',
    },
  },
  {
    metric: 'Subprocess Group Management',
    category: 'Control & Lifecycle',
    aethergrok: {
      title: 'POSIX setpgid & Win Job Objects',
      sub: 'Zero orphaned background tasks, SIGTERM/SIGKILL escalation',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Standard child_process',
      sub: 'Node.js standard process tree',
    },
    codexCli: {
      title: 'Subprocess System Call',
      sub: 'OS standard subprocess invocation',
    },
    antigravityCua: {
      title: 'OS Accessibility & Automation',
      sub: 'Direct OS mouse & keyboard event injection',
    },
  },
  {
    metric: 'Distribution & Setup Installation',
    category: 'Installation & Toolchain',
    aethergrok: {
      title: 'Native DMG / Setup + 1-Liner',
      sub: 'Self-contained installer + verified 1-line curl script',
      badge: 'featured',
    },
    claudeCode: {
      title: 'npm install -g',
      sub: 'Requires Node.js runtime on host machine',
    },
    codexCli: {
      title: 'Package CLI Binary',
      sub: 'Separate Python or CLI interpreter setup',
    },
    antigravityCua: {
      title: 'Container / Environment Setup',
      sub: 'Custom sandbox config & Python dependencies',
    },
  },
];
