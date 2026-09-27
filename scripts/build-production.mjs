#!/usr/bin/env node

/**
 * Interactive Production Build Script for AetherGrok (Go + Wails v2)
 * Compiles Go + Svelte frontend into native binaries for macOS (.app) & Windows (.exe).
 */

import { spawnSync } from 'child_process';
import * as fs from 'fs';
import * as path from 'path';
import * as readline from 'readline';

const ROOT_DIR = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..');
const BUILD_DIR = path.join(ROOT_DIR, 'build');
const BUILD_BIN_DIR = path.join(BUILD_DIR, 'bin');
const APP_ASSETS_DIR = path.join(ROOT_DIR, 'resources', 'app-assets');

const COLORS = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  dim: '\x1b[2m',
  cyan: '\x1b[36m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  red: '\x1b[31m',
  magenta: '\x1b[35m',
  blue: '\x1b[34m',
};

function log(msg = '') {
  console.log(msg);
}

function logStep(step, msg) {
  console.log(`${COLORS.bold}${COLORS.cyan}[${step}]${COLORS.reset} ${msg}`);
}

function logSuccess(msg) {
  console.log(`${COLORS.bold}${COLORS.green}✔ ${msg}${COLORS.reset}`);
}

function logError(msg) {
  console.error(`${COLORS.bold}${COLORS.red}✖ ${msg}${COLORS.reset}`);
}

function formatBytes(bytes) {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`;
}

// Ensure Go and Wails are found in PATH
function getEnhancedEnv() {
  const home = process.env.HOME || '';
  const extraPaths = [
    '/opt/homebrew/bin',
    '/usr/local/bin',
    path.join(home, 'go', 'bin'),
    path.join(home, '.local', 'bin'),
  ];
  const currentPath = process.env.PATH || '';
  return {
    ...process.env,
    PATH: `${extraPaths.join(':')}:${currentPath}`,
  };
}

/**
 * Clean existing build binaries and temporary build output
 */
function cleanBuildOutput() {
  if (fs.existsSync(BUILD_BIN_DIR)) {
    logStep('Clean', `Membersihkan folder output build: ${BUILD_BIN_DIR}`);
    fs.rmSync(BUILD_BIN_DIR, { recursive: true, force: true });
  }
}

/**
 * Synchronize canonical assets from resources/app-assets into build/
 * to ensure custom icons and platform configs are preserved when build/ is empty or deleted.
 */
function syncAppAssets() {
  logStep('Assets', 'Memastikan file asset dan konfigurasi build kustom terpasang...');

  if (!fs.existsSync(APP_ASSETS_DIR)) {
    return;
  }

  if (!fs.existsSync(BUILD_DIR)) {
    fs.mkdirSync(BUILD_DIR, { recursive: true });
  }

  // Copy appicon.png
  const sourceIcon = path.join(APP_ASSETS_DIR, 'appicon.png');
  const targetIcon = path.join(BUILD_DIR, 'appicon.png');
  if (fs.existsSync(sourceIcon)) {
    fs.copyFileSync(sourceIcon, targetIcon);
  }

  // Copy darwin directory (Info.plist, etc.)
  const sourceDarwin = path.join(APP_ASSETS_DIR, 'darwin');
  const targetDarwin = path.join(BUILD_DIR, 'darwin');
  if (fs.existsSync(sourceDarwin)) {
    if (!fs.existsSync(targetDarwin)) {
      fs.mkdirSync(targetDarwin, { recursive: true });
    }
    const darwinFiles = fs.readdirSync(sourceDarwin);
    for (const file of darwinFiles) {
      fs.copyFileSync(path.join(sourceDarwin, file), path.join(targetDarwin, file));
    }
  }

  // Copy windows directory (icon.ico, info.json, wails.exe.manifest, etc.)
  const sourceWindows = path.join(APP_ASSETS_DIR, 'windows');
  const targetWindows = path.join(BUILD_DIR, 'windows');
  if (fs.existsSync(sourceWindows)) {
    if (!fs.existsSync(targetWindows)) {
      fs.mkdirSync(targetWindows, { recursive: true });
    }
    const winFiles = fs.readdirSync(sourceWindows);
    for (const file of winFiles) {
      fs.copyFileSync(path.join(sourceWindows, file), path.join(targetWindows, file));
    }
  }

  logSuccess('Asset dan konfigurasi build berhasil disinkronisasi.');
}

/**
 * Post-build step for macOS: Inject permission usage descriptions into Info.plist and re-sign bundle
 */
function postProcessMacBundle() {
  const appBundle = path.join(BUILD_BIN_DIR, 'aethergrok.app');
  const appPlist = path.join(appBundle, 'Contents', 'Info.plist');
  if (fs.existsSync(appPlist) && process.platform === 'darwin') {
    logStep('Post-Build', 'Menambahkan permission string macOS ke Info.plist dan re-sign bundle...');
    try {
      spawnSync(
        '/usr/libexec/PlistBuddy',
        ['-c', "Add :NSMicrophoneUsageDescription string 'AetherGrok requires microphone access for voice dictation and speech-to-text input.'", appPlist],
        { stdio: 'ignore' }
      );
      spawnSync(
        '/usr/libexec/PlistBuddy',
        ['-c', "Set :NSMicrophoneUsageDescription 'AetherGrok requires microphone access for voice dictation and speech-to-text input.'", appPlist],
        { stdio: 'ignore' }
      );
      spawnSync(
        '/usr/libexec/PlistBuddy',
        ['-c', "Add :NSSpeechRecognitionUsageDescription string 'AetherGrok uses speech recognition to convert dictated voice prompts into text.'", appPlist],
        { stdio: 'ignore' }
      );
      spawnSync(
        '/usr/libexec/PlistBuddy',
        ['-c', "Set :NSSpeechRecognitionUsageDescription 'AetherGrok uses speech recognition to convert dictated voice prompts into text.'", appPlist],
        { stdio: 'ignore' }
      );

      spawnSync(
        '/usr/libexec/PlistBuddy',
        ['-c', "Add :NSAppleEventsUsageDescription string 'AetherGrok requires automation access to coordinate workflow tasks.'", appPlist],
        { stdio: 'ignore' }
      );
      spawnSync(
        '/usr/libexec/PlistBuddy',
        ['-c', "Set :NSAppleEventsUsageDescription 'AetherGrok requires automation access to coordinate workflow tasks.'", appPlist],
        { stdio: 'ignore' }
      );

      // Re-sign the app bundle with a persistent designated requirement (DR) tied to identifier "com.wails.aethergrok"
      // This ensures macOS TCC preserves granted permissions across app updates and rebuilds
      spawnSync(
        'codesign',
        ['--force', '--deep', '--sign', '-', '--requirements', '= designated => identifier "com.wails.aethergrok"', appBundle],
        { stdio: 'ignore' }
      );
    } catch {
      // PlistBuddy / codesign fallback
    }
  }
}

async function runInteractiveSelection() {
  const options = [
    { label: 'macOS (Go/Wails .app: Apple Silicon arm64)', id: 'mac', checked: true },
    { label: 'Windows (Go/Wails .exe: Windows x64)', id: 'win', checked: true },
  ];

  if (!process.stdin.isTTY) {
    return { mac: true, win: true };
  }

  return new Promise((resolve) => {
    readline.emitKeypressEvents(process.stdin);
    process.stdin.setRawMode(true);

    let cursor = 0;
    let linesRendered = 0;

    function render(initial = false) {
      if (!initial && linesRendered > 0) {
        readline.cursorTo(process.stdout, 0);
        readline.moveCursor(process.stdout, 0, -linesRendered);
        readline.clearScreenDown(process.stdout);
      }

      let output = '';
      output += `${COLORS.bold}${COLORS.magenta}╔══════════════════════════════════════════════════════════════════╗${COLORS.reset}\n`;
      output += `${COLORS.bold}${COLORS.magenta}║          ⚡ AetherGrok (Go + Wails) — Production Builder         ║${COLORS.reset}\n`;
      output += `${COLORS.bold}${COLORS.magenta}╚══════════════════════════════════════════════════════════════════╝${COLORS.reset}\n\n`;

      output += `${COLORS.bold}Gunakan tombol [Panah Atas/Bawah] untuk navigasi, [SPASI] untuk centang/uncheck, [ENTER] untuk mulai:${COLORS.reset}\n\n`;

      options.forEach((opt, idx) => {
        const isCurrent = idx === cursor;
        const pointer = isCurrent ? `${COLORS.bold}${COLORS.cyan}❯${COLORS.reset}` : ' ';
        const checkbox = opt.checked
          ? `${COLORS.bold}${COLORS.green}[✔]${COLORS.reset}`
          : `${COLORS.dim}[ ]${COLORS.reset}`;
        const labelStyle = isCurrent ? `${COLORS.bold}${COLORS.cyan}${opt.label}${COLORS.reset}` : opt.label;
        output += ` ${pointer} ${checkbox} ${labelStyle}\n`;
      });

      output += `\n${COLORS.dim}Tombol pintas: [A] Centang Semua, [N] Bersihkan Centang, [Q] / [Ctrl+C] Keluar${COLORS.reset}\n`;

      process.stdout.write(output);
      linesRendered = output.split('\n').length - 1;
    }

    render(true);

    function onKeypress(str, key) {
      if (key.ctrl && key.name === 'c' || key.name === 'q') {
        process.stdin.setRawMode(false);
        process.stdin.pause();
        process.stdin.removeListener('keypress', onKeypress);
        log(`\n${COLORS.yellow}Build dibatalkan.${COLORS.reset}`);
        process.exit(0);
      }

      if (key.name === 'up') {
        cursor = (cursor - 1 + options.length) % options.length;
        render();
      } else if (key.name === 'down') {
        cursor = (cursor + 1) % options.length;
        render();
      } else if (key.name === 'space') {
        options[cursor].checked = !options[cursor].checked;
        render();
      } else if (key.name === 'a') {
        options.forEach((o) => (o.checked = true));
        render();
      } else if (key.name === 'n') {
        options.forEach((o) => (o.checked = false));
        render();
      } else if (key.name === 'return' || key.name === 'enter') {
        const selected = {
          mac: options.find((o) => o.id === 'mac')?.checked || false,
          win: options.find((o) => o.id === 'win')?.checked || false,
        };

        if (!selected.mac && !selected.win) {
          render();
          log(`${COLORS.bold}${COLORS.red}Peringatan: Pilih minimal satu target build (spasi untuk mencentang).${COLORS.reset}`);
          return;
        }

        process.stdin.setRawMode(false);
        process.stdin.pause();
        process.stdin.removeListener('keypress', onKeypress);
        process.stdout.write('\n');
        resolve(selected);
      }
    }

    process.stdin.on('keypress', onKeypress);
  });
}

function parseCliArgs() {
  const args = process.argv.slice(2);
  const targets = { mac: false, win: false, isExplicit: false };

  for (const arg of args) {
    if (arg === '--mac' || arg === '-m') {
      targets.mac = true;
      targets.isExplicit = true;
    } else if (arg === '--win' || arg === '--windows' || arg === '-w') {
      targets.win = true;
      targets.isExplicit = true;
    } else if (arg === '--all' || arg === '--both' || arg === '-a') {
      targets.mac = true;
      targets.win = true;
      targets.isExplicit = true;
    }
  }

  return targets;
}

function runWailsBuild(platform, stepName) {
  logStep(stepName, `wails build -platform ${platform}`);
  const env = getEnhancedEnv();
  const result = spawnSync('wails', ['build', '-platform', platform], {
    cwd: ROOT_DIR,
    stdio: 'inherit',
    env,
  });

  if (result.status !== 0) {
    logError(`Build Wails gagal untuk platform ${platform} (exit code: ${result.status})`);
    process.exit(result.status || 1);
  }
}

async function main() {
  const cliTargets = parseCliArgs();
  let selected = { mac: cliTargets.mac, win: cliTargets.win };

  if (!cliTargets.isExplicit) {
    selected = await runInteractiveSelection();
  }

  log(`\n${COLORS.bold}${COLORS.magenta}=== Memulai Build Produksi AetherGrok (Go + Wails) ===${COLORS.reset}`);
  log(`Target: ${[selected.mac && 'macOS (darwin/arm64)', selected.win && 'Windows (windows/amd64)'].filter(Boolean).join(' & ')}\n`);

  // 1. Clean previous build artifacts
  cleanBuildOutput();

  // 2. Synchronize master custom assets into build/
  syncAppAssets();

  let currentStep = 1;
  const totalSteps = (selected.mac ? 1 : 0) + (selected.win ? 1 : 0);

  if (selected.mac) {
    runWailsBuild('darwin/arm64', `${currentStep++}/${totalSteps} Building macOS App`);
    postProcessMacBundle();
  }

  if (selected.win) {
    runWailsBuild('windows/amd64', `${currentStep++}/${totalSteps} Building Windows Executable`);
  }

  // Summary
  log(`\n${COLORS.bold}${COLORS.green}══════════════════════════════════════════════════════════════════${COLORS.reset}`);
  logSuccess('Build produksi Wails (Go) berhasil diselesaikan!\n');

  if (fs.existsSync(BUILD_BIN_DIR)) {
    log(`${COLORS.bold}Daftar File Output di ${BUILD_BIN_DIR}:${COLORS.reset}`);
    const files = fs.readdirSync(BUILD_BIN_DIR);
    for (const f of files) {
      const fullPath = path.join(BUILD_BIN_DIR, f);
      const stat = fs.statSync(fullPath);
      if (stat.isDirectory() && f.endsWith('.app')) {
        log(` • ${COLORS.cyan}${f}${COLORS.reset} (macOS App Bundle)`);
      } else if (f.endsWith('.exe')) {
        const sizeFormatted = formatBytes(stat.size).padStart(10);
        log(` • ${COLORS.cyan}${f}${COLORS.reset} (${COLORS.yellow}${sizeFormatted}${COLORS.reset})`);
      }
    }
  }
  log(`${COLORS.bold}${COLORS.green}══════════════════════════════════════════════════════════════════${COLORS.reset}\n`);
}

main().catch((err) => {
  logError(err.message);
  process.exit(1);
});
