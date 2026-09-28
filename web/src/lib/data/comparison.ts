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
  'Arsitektur & Runtime',
  'Interaksi & Ergonomi',
  'Kontrol & Lifecycle',
  'Instalasi & Toolchain',
];

export const comparisonData: ComparisonRow[] = [
  {
    metric: 'Primary Interaction Mode',
    category: 'Arsitektur & Runtime',
    aethergrok: {
      title: 'Native Desktop Studio',
      sub: 'Go 1.24 Core + Wails v2 + Svelte 5',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Terminal REPL',
      sub: 'Node.js CLI via stdout',
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
    category: 'Arsitektur & Runtime',
    aethergrok: {
      title: '~35 MB Baseline RAM',
      sub: 'Native WebKit/WebView2 (tanpa Chromium bloat)',
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
    category: 'Arsitektur & Runtime',
    aethergrok: {
      title: '10-Turn DOM Virtualization',
      sub: 'Memory konstan, fluid 60 FPS di 100+ turns',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Scrollback Buffer Limit',
      sub: 'Tergantung buffer memori terminal emulator',
    },
    codexCli: {
      title: 'Stdout Terminal Stream',
      sub: 'Buffer teks terminal standar',
    },
    antigravityCua: {
      title: 'Heavy Canvas Re-renders',
      sub: 'Beban DOM & canvas meningkat seiring waktu',
    },
  },
  {
    metric: 'Live Mid-Turn Steering',
    category: 'Interaksi & Ergonomi',
    aethergrok: {
      title: 'In-Flight Prompt Injection',
      sub: 'Kirim prompt arah baru tanpa stop atau kill turn',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Interrupt / Re-prompt',
      sub: 'Harus batalkan turn aktif terlebih dahulu',
    },
    codexCli: {
      title: 'Metode ACP Protokol',
      sub: 'Dukungan terbatas via RPC cancel/re-turn',
    },
    antigravityCua: {
      title: 'Queue Re-ordering',
      sub: 'Penyesuaian antrean aksi UI tertunda',
    },
  },
  {
    metric: 'Visual Screen Context (Vision)',
    category: 'Interaksi & Ergonomi',
    aethergrok: {
      title: 'Compositor-Synced Snap',
      sub: 'Auto-hide window 50ms, tangkap layar OS, attach otomatis',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Manual Path Input',
      sub: 'Ketik path file gambar lokal manual',
    },
    codexCli: {
      title: 'Manual Attachment',
      sub: 'Referensi file path lokal',
    },
    antigravityCua: {
      title: 'Full Continuous Screen Grab',
      sub: 'Stream frame pixel layar berkelanjutan',
    },
  },
  {
    metric: 'Voice Dictation & Media Ducking',
    category: 'Interaksi & Ergonomi',
    aethergrok: {
      title: 'Push-to-Talk + Auto Mute',
      sub: 'Equalizer dinamis 4 bar + volume sistem auto-mute ke 0%',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Tidak Ada Fitur Suara',
      sub: 'Tergantung host terminal',
    },
    codexCli: {
      title: 'Tidak Ada Fitur Suara',
      sub: 'Tergantung host terminal',
    },
    antigravityCua: {
      title: 'Tergantung Host OS',
      sub: 'Tidak terintegrasi bawaan',
    },
  },
  {
    metric: 'Subprocess Group Management',
    category: 'Kontrol & Lifecycle',
    aethergrok: {
      title: 'POSIX setpgid & Win Job Objects',
      sub: 'Nol background process tertinggal, sinyal SIGTERM/KILL',
      badge: 'featured',
    },
    claudeCode: {
      title: 'Standard child_process',
      sub: 'Tree proses standar Node.js',
    },
    codexCli: {
      title: 'Subprocess System Call',
      sub: 'Pemanggilan subprocess standar OS',
    },
    antigravityCua: {
      title: 'OS Accessibility & Automation',
      sub: 'Injeksi event keyboard & mouse OS',
    },
  },
  {
    metric: 'Distribusi & Setup Installation',
    category: 'Instalasi & Toolchain',
    aethergrok: {
      title: 'Native DMG / Setup + 1-Liner',
      sub: 'Installer visual mandiri + skrip curl 1 baris terverifikasi',
      badge: 'featured',
    },
    claudeCode: {
      title: 'npm install -g',
      sub: 'Memerlukan runtime Node.js di sistem',
    },
    codexCli: {
      title: 'Package CLI Binary',
      sub: 'Setup interpreter Python atau CLI terpisah',
    },
    antigravityCua: {
      title: 'Container / Environment Setup',
      sub: 'Konfigurasi sandbox OS & dependensi Python',
    },
  },
];
