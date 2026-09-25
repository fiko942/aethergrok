import {
  FileText,
  FileCode,
  FileArchive,
  FileSpreadsheet,
  FileJson,
  FileVideo,
  FileAudio,
  FileImage,
  FileTerminal,
  Settings,
  Database,
  FileCog,
  File
} from 'lucide-svelte';

export interface FileIconMeta {
  component: any;
  colorClass: string;
  bgClass: string;
  badge: string;
}

export function resolveFileIcon(fileName: string, isImage = false): FileIconMeta {
  if (isImage) {
    return {
      component: FileImage,
      colorClass: 'text-amber-400',
      bgClass: 'bg-amber-400/10',
      badge: 'IMG'
    };
  }

  const clean = (fileName || '').toLowerCase().trim();
  const ext = clean.split('.').pop() || '';

  // Archives & Compressed files
  if (['zip', 'tar', 'gz', 'tgz', '7z', 'rar', 'bz2', 'xz', 'iso', 'dmg'].includes(ext)) {
    return {
      component: FileArchive,
      colorClass: 'text-amber-500',
      bgClass: 'bg-amber-500/10',
      badge: 'ZIP'
    };
  }

  // JSON & Data files
  if (['json', 'jsonl', 'ndjson', 'geojson'].includes(ext)) {
    return {
      component: FileJson,
      colorClass: 'text-yellow-400',
      bgClass: 'bg-yellow-400/10',
      badge: 'JSON'
    };
  }

  // Database files
  if (['sql', 'sqlite', 'db', 'prisma', 'schema'].includes(ext)) {
    return {
      component: Database,
      colorClass: 'text-cyan-400',
      bgClass: 'bg-cyan-400/10',
      badge: 'SQL'
    };
  }

  // Config & Env files
  if (['env', 'yaml', 'yml', 'toml', 'ini', 'conf', 'config', 'properties', 'dockerignore', 'gitignore'].includes(ext) || clean.startsWith('.env') || clean === 'dockerfile') {
    return {
      component: Settings,
      colorClass: 'text-zinc-400',
      bgClass: 'bg-zinc-400/10',
      badge: 'CFG'
    };
  }

  // Build & CMake / Make
  if (['cmake', 'makefile', 'mk', 'gradle', 'bazel'].includes(ext) || clean === 'cmakelists.txt' || clean === 'makefile') {
    return {
      component: FileCog,
      colorClass: 'text-orange-400',
      bgClass: 'bg-orange-400/10',
      badge: 'MAKE'
    };
  }

  // Shell & Scripts
  if (['sh', 'bash', 'zsh', 'fish', 'ps1', 'bat', 'cmd'].includes(ext)) {
    return {
      component: FileTerminal,
      colorClass: 'text-emerald-400',
      bgClass: 'bg-emerald-400/10',
      badge: 'SH'
    };
  }

  // Code files
  if (['ts', 'tsx', 'js', 'jsx', 'svelte', 'vue', 'py', 'go', 'rs', 'c', 'cpp', 'h', 'hpp', 'java', 'kt', 'swift', 'rb', 'php', 'html', 'css', 'scss', 'wasm'].includes(ext)) {
    return {
      component: FileCode,
      colorClass: 'text-blue-400',
      bgClass: 'bg-blue-400/10',
      badge: ext.toUpperCase()
    };
  }

  // Spreadsheets
  if (['csv', 'tsv', 'xlsx', 'xls'].includes(ext)) {
    return {
      component: FileSpreadsheet,
      colorClass: 'text-emerald-500',
      bgClass: 'bg-emerald-500/10',
      badge: 'SHEET'
    };
  }

  // Media
  if (['mp4', 'mov', 'mkv', 'webm', 'avi'].includes(ext)) {
    return {
      component: FileVideo,
      colorClass: 'text-purple-400',
      bgClass: 'bg-purple-400/10',
      badge: 'VID'
    };
  }
  if (['mp3', 'wav', 'ogg', 'm4a', 'flac', 'aac'].includes(ext)) {
    return {
      component: FileAudio,
      colorClass: 'text-pink-400',
      bgClass: 'bg-pink-400/10',
      badge: 'AUD'
    };
  }

  // Documents & Markdown
  if (['md', 'markdown', 'txt', 'pdf', 'doc', 'docx', 'rtf'].includes(ext)) {
    return {
      component: FileText,
      colorClass: 'text-indigo-400',
      bgClass: 'bg-indigo-400/10',
      badge: ext.toUpperCase()
    };
  }

  return {
    component: File,
    colorClass: 'text-zinc-400',
    bgClass: 'bg-zinc-400/10',
    badge: ext ? ext.slice(0, 4).toUpperCase() : 'FILE'
  };
}
