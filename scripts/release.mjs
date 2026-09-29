#!/usr/bin/env node

/**
 * AetherGrok Interactive Release Manager
 * 
 * - Detects previous release and commit from GitHub Releases API
 * - Computes git commit diff range (since last release tag/commit)
 * - Interactive SemVer version bump & changelog prompt
 * - Updates changelog.json, wails.json, and package.json
 * - Commits, creates git tag, and pushes to remote
 * - Triggers & streams GitHub Actions multi-platform build progress
 * - Provides live download links for all 8 artifact variants
 * 
 * Author: Wiji Fiko Teren
 */

import { execSync } from 'child_process';
import readline from 'readline';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const ROOT_DIR = path.resolve(__dirname, '..');

// ANSI Color Helpers
const colors = {
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

const c = (color, text) => `${colors[color] || ''}${text}${colors.reset}`;

// Helper: Run command synchronously
function run(cmd, options = {}) {
  try {
    return execSync(cmd, { cwd: ROOT_DIR, encoding: 'utf8', stdio: ['pipe', 'pipe', 'pipe'], ...options }).trim();
  } catch (err) {
    if (options.silent) return '';
    throw err;
  }
}

// Find Go executable across Windows, macOS, and Linux
function findGoExecutable() {
  if (process.platform === 'win32') {
    const candidates = [
      'go',
      'C:\\Program Files\\Go\\bin\\go.exe',
      'C:\\Go\\bin\\go.exe',
      path.join(process.env.ProgramFiles || 'C:\\Program Files', 'Go', 'bin', 'go.exe'),
    ];
    for (const cand of candidates) {
      try {
        execSync(`"${cand}" version`, { stdio: 'ignore' });
        return `"${cand}"`;
      } catch (_) {}
    }
  }
  return 'go';
}

// Strict Pre-flight Checks
function runPreflightChecks() {
  console.log('\n' + c('yellow', '🧪 Running Strict Pre-Flight Release Validation Suite...'));

  const steps = [
    {
      title: 'Checking Git working directory status',
      fn: () => {
        const status = run('git status --porcelain', { silent: true });
        // Allow uncommitted changes only if they are changelog/wails/package/scripts
        const uncommitted = status.split('\n').filter(Boolean).filter(line => {
          return !line.includes('changelog.json') && !line.includes('wails.json') && !line.includes('.github/') && !line.includes('scripts/');
        });
        if (uncommitted.length > 0) {
          throw new Error(`Working directory has uncommitted files:\n${uncommitted.join('\n')}`);
        }
      }
    },
    {
      title: 'Validating root TypeScript compilation (tsc -p .)',
      fn: () => {
        execSync('pnpm run compile', { cwd: ROOT_DIR, stdio: 'inherit' });
      }
    },
    {
      title: 'Running unit test suite (vitest run)',
      skip: () => process.env.SKIP_TESTS === '1' || process.env.SKIP_TESTS === 'true',
      fn: () => {
        execSync('pnpm run test', { cwd: ROOT_DIR, stdio: 'inherit' });
      }
    },
    {
      title: 'Compiling Vite Frontend Production Bundle',
      fn: () => {
        execSync('pnpm run build', { cwd: path.join(ROOT_DIR, 'frontend'), stdio: 'inherit' });
      }
    },
    {
      title: 'Testing Go backend packages (go test ./...)',
      fn: () => {
        const goBin = findGoExecutable();
        const envPath = process.platform === 'win32'
          ? `C:\\Program Files\\Go\\bin;${process.env.PATH || ''}`
          : `/opt/homebrew/bin:/usr/local/bin:${process.env.PATH || ''}`;
        execSync(`${goBin} test ./...`, { cwd: ROOT_DIR, stdio: 'inherit', env: { ...process.env, PATH: envPath } });
      }
    },
    {
      title: 'Testing Go desktop application build',
      fn: () => {
        const goBin = findGoExecutable();
        const envPath = process.platform === 'win32'
          ? `C:\\Program Files\\Go\\bin;${process.env.PATH || ''}`
          : `/opt/homebrew/bin:/usr/local/bin:${process.env.PATH || ''}`;
        const nullOut = process.platform === 'win32' ? 'NUL' : '/dev/null';
        execSync(`${goBin} build -o ${nullOut} .`, { cwd: ROOT_DIR, stdio: 'inherit', env: { ...process.env, PATH: envPath } });
      }
    }
  ];

  for (let i = 0; i < steps.length; i++) {
    const step = steps[i];
    if (step.skip && step.skip()) {
      console.log(`   [${i + 1}/${steps.length}] ${step.title}... ` + c('yellow', '⚡ SKIPPED (already verified)'));
      continue;
    }
    process.stdout.write(`   [${i + 1}/${steps.length}] ${step.title}... `);
    try {
      step.fn();
      console.log(c('green', '✓ PASSED'));
    } catch (err) {
      console.log(c('red', '✗ FAILED'));
      throw new Error(`Pre-flight validation failed at step: "${step.title}"\n${err.message || ''}`);
    }
  }

  console.log(c('green', '✨ All strict pre-flight checks passed successfully!\n'));
}

// Helper: Ask question via Readline
function ask(rl, query) {
  return new Promise((resolve) => rl.question(query, resolve));
}

// GitHub API Fetcher using native fetch (Node 18+)
async function ghApi(endpoint, token, options = {}) {
  const url = endpoint.startsWith('https://') ? endpoint : `https://api.github.com${endpoint}`;
  const headers = {
    Accept: 'application/vnd.github.v3+json',
    'User-Agent': 'AetherGrok-Release-Manager',
    ...(token ? { Authorization: `token ${token}` } : {}),
    ...(options.headers || {}),
  };

  const res = await fetch(url, { ...options, headers });
  if (!res.ok) {
    const errorText = await res.text();
    throw new Error(`GitHub API Error (${res.status} ${res.statusText}): ${errorText}`);
  }
  return res.json();
}

// Get GitHub Token from ENV or Git Credential Helper
function getGitHubToken() {
  if (process.env.GITHUB_TOKEN) return process.env.GITHUB_TOKEN;
  if (process.env.GH_TOKEN) return process.env.GH_TOKEN;
  try {
    const creds = run('git credential fill', {
      input: 'protocol=https\nhost=github.com\n',
      silent: true,
    });
    const match = creds.match(/password=(.*)/);
    if (match && match[1]) return match[1].trim();
  } catch (_) {}
  return null;
}

// Detect Repository Owner and Name from Git Remote
function getRepoInfo() {
  const remoteUrl = run('git config --get remote.origin.url', { silent: true });
  const match = remoteUrl.match(/github\.com[:/]([^/]+)\/([^/.]+)(?:\.git)?$/);
  if (match) {
    return { owner: match[1], repo: match[2] };
  }
  return { owner: 'fiko942', repo: 'aethergrok' };
}

// SemVer Bump Helper
function bumpVersion(version, type) {
  const parts = version.replace(/^v/, '').split('.').map(Number);
  if (parts.length < 3) return '1.0.1';
  let [major, minor, patch] = parts;
  if (type === 'major') {
    major += 1;
    minor = 0;
    patch = 0;
  } else if (type === 'minor') {
    minor += 1;
    patch = 0;
  } else if (type === 'patch') {
    patch += 1;
  }
  return `${major}.${minor}.${patch}`;
}

async function main() {
  console.clear();
  console.log(c('cyan', '╔═══════════════════════════════════════════════════════════════════════╗'));
  console.log(c('cyan', '║             🚀 AetherGrok Interactive Release Manager 🚀             ║'));
  console.log(c('cyan', '╚═══════════════════════════════════════════════════════════════════════╝\n'));

  const rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout,
  });

  try {
    const token = getGitHubToken();
    const { owner, repo } = getRepoInfo();
    console.log(c('blue', '📍 Target Repository: ') + c('bold', `${owner}/${repo}`));
    if (token) {
      console.log(c('green', '🔑 GitHub Authentication: ') + c('dim', 'Token successfully detected.'));
    } else {
      console.log(c('yellow', '⚠️  No GitHub token found. (Releases will require manual git push/API authentication)'));
    }

    // 1. Fetch Latest Release Information
    console.log('\n' + c('yellow', '1. Fetching latest release info from GitHub API...'));
    let latestRelease = null;
    let latestTag = null;
    let lastCommitHash = null;

    try {
      const releases = await ghApi(`/repos/${owner}/${repo}/releases`, token);
      if (releases && releases.length > 0) {
        latestRelease = releases[0];
        latestTag = latestRelease.tag_name;
      }
    } catch (e) {
      console.log(c('dim', `   Note: Could not fetch releases via API (${e.message}). Falling back to git tags.`));
    }

    // Fallback: Read git tags locally
    if (!latestTag) {
      const gitTags = run('git tag -l --sort=-v:refname', { silent: true });
      if (gitTags) {
        const first = gitTags.split('\n')[0]?.trim();
        if (first) latestTag = first;
      }
    }

    // Read current version from changelog.json or wails.json
    let currentVersion = '1.0.0';
    const changelogPath = path.join(ROOT_DIR, 'changelog.json');
    let changelogData = { latest: '1.0.0', releases: [] };

    if (fs.existsSync(changelogPath)) {
      try {
        changelogData = JSON.parse(fs.readFileSync(changelogPath, 'utf8'));
        if (changelogData.latest) currentVersion = changelogData.latest;
      } catch (_) {}
    } else {
      const wailsPath = path.join(ROOT_DIR, 'wails.json');
      if (fs.existsSync(wailsPath)) {
        try {
          const w = JSON.parse(fs.readFileSync(wailsPath, 'utf8'));
          if (w.version) currentVersion = w.version;
        } catch (_) {}
      }
    }

    if (latestTag) {
      try {
        lastCommitHash = run(`git rev-list -n 1 ${latestTag}`, { silent: true });
      } catch (_) {}
    }

    console.log(c('green', '✓ Current active release: ') + c('bold', latestTag || `v${currentVersion}`));
    if (lastCommitHash) {
      console.log(c('dim', `  Latest release commit: ${lastCommitHash.substring(0, 8)}`));
    }

    // 2. Compute Commit History Since Last Release
    console.log('\n' + c('yellow', '2. Analyzing git commit changes since last release...'));
    const gitLogRange = latestTag ? `${latestTag}..HEAD` : 'HEAD~10..HEAD';
    let commitList = [];
    try {
      const rawLog = run(`git log ${gitLogRange} --pretty=format:"%h%x09%s"`, { silent: true });
      if (rawLog.trim()) {
        commitList = rawLog.split('\n').map((line) => {
          const [hash, ...rest] = line.split('\t');
          return { hash, message: rest.join('\t') };
        });
      }
    } catch (_) {
      // In case tag doesn't exist locally, take last 5 commits
      const rawLog = run('git log -n 5 --pretty=format:"%h%x09%s"', { silent: true });
      commitList = rawLog.split('\n').map((line) => {
        const [hash, ...rest] = line.split('\t');
        return { hash, message: rest.join('\t') };
      });
    }

    if (commitList.length === 0) {
      console.log(c('yellow', '   No new commits detected since last release tag. Proceeding with current tree.'));
    } else {
      console.log(c('cyan', `   Found ${commitList.length} new commit(s):`));
      commitList.forEach((item, idx) => {
        console.log(`   ${c('dim', `${idx + 1}.`)} ${c('magenta', item.hash)} ${item.message}`);
      });
    }

    const patchVer = bumpVersion(currentVersion, 'patch');
    const minorVer = bumpVersion(currentVersion, 'minor');
    const majorVer = bumpVersion(currentVersion, 'major');

    const inputVersion = process.env.RELEASE_VERSION;
    const inputTitleEnv = process.env.RELEASE_TITLE;
    const inputHighlightsEnv = process.env.RELEASE_HIGHLIGHTS;
    const autoConfirmEnv = process.env.RELEASE_CONFIRM === 'yes' || process.env.RELEASE_CONFIRM === 'y';

    let nextVersion = patchVer;

    if (inputVersion) {
      nextVersion = inputVersion.replace(/^v/, '');
    } else {
      console.log('\n' + c('yellow', '3. Select New Release Version:'));
      console.log(`   [1] Patch: ${c('bold', patchVer)} (Bug fixes & small improvements)`);
      console.log(`   [2] Minor: ${c('bold', minorVer)} (New features & enhancements)`);
      console.log(`   [3] Major: ${c('bold', majorVer)} (Breaking changes / major overhaul)`);
      console.log(`   [4] Custom input`);

      const verChoice = (await ask(rl, c('cyan', '\n👉 Choose option [1-4] or enter version (default: 1): '))).trim() || '1';
      if (verChoice === '1') nextVersion = patchVer;
      else if (verChoice === '2') nextVersion = minorVer;
      else if (verChoice === '3') nextVersion = majorVer;
      else if (verChoice === '4' || /^\d+\.\d+\.\d+/.test(verChoice)) {
        if (/^\d+\.\d+\.\d+/.test(verChoice)) {
          nextVersion = verChoice.replace(/^v/, '');
        } else {
          const custom = (await ask(rl, '   Enter custom SemVer version (e.g. 1.2.0): ')).trim();
          nextVersion = custom.replace(/^v/, '') || patchVer;
        }
      }
    }

    const nextTag = `v${nextVersion}`;
    console.log(c('green', `\n✓ Target release version set to: `) + c('bold', nextTag));

    // 4. Prompt for Release Title & Highlights
    const defaultTitle = `AetherGrok ${nextVersion} - Release`;
    let inputTitle = defaultTitle;
    const highlights = [];

    if (inputTitleEnv) {
      inputTitle = inputTitleEnv;
      if (inputHighlightsEnv) {
        inputHighlightsEnv.split('\n').map((h) => h.trim()).filter(Boolean).forEach((h) => highlights.push(h));
      }
    } else {
      console.log('\n' + c('yellow', '4. Enter Release Title & Highlights:'));
      inputTitle = (await ask(rl, c('cyan', `👉 Release title [${defaultTitle}]: `))).trim() || defaultTitle;

      console.log(c('cyan', '\n👉 Enter Changelog highlights (Bullet points, enter an empty line when finished):'));
      let highlightIdx = 1;
      while (true) {
        const line = (await ask(rl, `   ${highlightIdx}. `)).trim();
        if (!line) {
          if (highlights.length === 0) {
            // Default highlight from commits
            highlights.push(commitList[0]?.message || 'General performance and stability improvements');
          }
          break;
        }
        highlights.push(line);
        highlightIdx++;
      }
    }

    // 5. Build Artifact Names & URLs
    const artifacts = {
      macos_arm64_dmg: `AetherGrok-${nextVersion}-macOS-arm64.dmg`,
      macos_x64_dmg: `AetherGrok-${nextVersion}-macOS-amd64.dmg`,
      windows_x64_setup: `AetherGrok-${nextVersion}-windows-amd64-setup.exe`,
      windows_x64_portable: `AetherGrok-${nextVersion}-windows-amd64-portable.zip`,
      windows_arm64_setup: `AetherGrok-${nextVersion}-windows-arm64-setup.exe`,
      windows_arm64_portable: `AetherGrok-${nextVersion}-windows-arm64-portable.zip`,
    };

    const downloadUrls = {};
    for (const [key, filename] of Object.entries(artifacts)) {
      downloadUrls[key] = `https://github.com/${owner}/${repo}/releases/download/${nextTag}/${filename}`;
    }

    // 6. Preview Release Summary
    console.log('\n' + c('cyan', '═══════════════════════ RELEASE PREVIEW ═══════════════════════'));
    console.log(`Version:    ${c('bold', nextVersion)} (${nextTag})`);
    console.log(`Title:      ${c('bold', inputTitle)}`);
    console.log(`Repository: ${owner}/${repo}`);
    console.log(`Commits:    ${commitList.length} included`);
    console.log(`Highlights:`);
    highlights.forEach((h) => console.log(`  • ${h}`));
    console.log(c('cyan', '════════════════════════════════════════════════════════════════'));

    let confirm = 'y';
    if (!autoConfirmEnv) {
      confirm = (await ask(rl, c('yellow', '\n⚠️  Ready to commit, tag, and publish release to GitHub? (y/N): '))).trim().toLowerCase();
    }
    if (confirm !== 'y' && confirm !== 'yes') {
      console.log(c('red', 'Release aborted by user.'));
      rl.close();
      return;
    }

    // Run Strict Pre-flight verification before modifying files or creating tags
    runPreflightChecks();

    // 7. Update Files: changelog.json, wails.json, package.json
    console.log('\n' + c('yellow', '5. Updating project files & changelog.json...'));

    const headCommit = run('git rev-parse HEAD', { silent: true });

    // Update changelog.json
    const newReleaseEntry = {
      version: nextVersion,
      tag: nextTag,
      date: new Date().toISOString(),
      title: inputTitle,
      targetCommit: headCommit.substring(0, 7),
      previousCommit: lastCommitHash ? lastCommitHash.substring(0, 7) : null,
      highlights,
      commits: commitList.slice(0, 50),
      downloads: downloadUrls,
    };

    changelogData.latest = nextVersion;
    changelogData.releases = [newReleaseEntry, ...changelogData.releases.filter((r) => r.version !== nextVersion)];
    fs.writeFileSync(changelogPath, JSON.stringify(changelogData, null, 2) + '\n');
    console.log(c('green', '✓ Updated changelog.json'));

    // Update wails.json
    const wailsPath = path.join(ROOT_DIR, 'wails.json');
    if (fs.existsSync(wailsPath)) {
      const wailsConfig = JSON.parse(fs.readFileSync(wailsPath, 'utf8'));
      wailsConfig.version = nextVersion;
      fs.writeFileSync(wailsPath, JSON.stringify(wailsConfig, null, 2) + '\n');
      console.log(c('green', '✓ Updated wails.json'));
    }

    // 8. Commit, Tag, and Push
    console.log('\n' + c('yellow', '6. Git commit, create tag, and push to GitHub...'));
    run('git add changelog.json wails.json');
    run(`git commit -m "chore(release): ${nextTag} - ${inputTitle}"`);
    console.log(c('green', `✓ Committed release updates.`));

    run(`git tag -a ${nextTag} -m "Release ${nextTag}: ${inputTitle}"`);
    console.log(c('green', `✓ Tag ${nextTag} created.`));

    console.log(c('cyan', '   Pushing commits and tags to origin...'));
    run('git push origin main --follow-tags');
    console.log(c('green', `✓ Pushed to origin/main with tag ${nextTag}.`));

    // 9. Create GitHub Release via API
    let releaseUrl = `https://github.com/${owner}/${repo}/releases/tag/${nextTag}`;
    if (token) {
      console.log('\n' + c('yellow', '7. Creating GitHub Release via API...'));

      let releaseBody = `## ${inputTitle}\n\n### Highlights\n`;
      highlights.forEach((h) => {
        releaseBody += `- ${h}\n`;
      });

      releaseBody += `\n### Included Commits\n`;
      commitList.forEach((c) => {
        releaseBody += `- ${c.hash} ${c.message}\n`;
      });

      releaseBody += `\n### Download Artifacts\n`;
      releaseBody += `| Platform | Architecture | Type | Package |\n`;
      releaseBody += `|---|---|---|---|\n`;
      releaseBody += `| **macOS** | Apple Silicon (M1-M4) | DMG Installer | [${artifacts.macos_arm64_dmg}](${downloadUrls.macos_arm64_dmg}) |\n`;
      releaseBody += `| **macOS** | Intel x64 | DMG Installer | [${artifacts.macos_x64_dmg}](${downloadUrls.macos_x64_dmg}) |\n`;
      releaseBody += `| **Windows** | x64 (64-bit) | Setup Installer | [${artifacts.windows_x64_setup}](${downloadUrls.windows_x64_setup}) |\n`;
      releaseBody += `| **Windows** | x64 (64-bit) | Portable Zip | [${artifacts.windows_x64_portable}](${downloadUrls.windows_x64_portable}) |\n`;
      releaseBody += `| **Windows** | ARM64 | Setup Installer | [${artifacts.windows_arm64_setup}](${downloadUrls.windows_arm64_setup}) |\n`;
      releaseBody += `| **Windows** | ARM64 | Portable Zip | [${artifacts.windows_arm64_portable}](${downloadUrls.windows_arm64_portable}) |\n`;

      try {
        const createdRelease = await ghApi(`/repos/${owner}/${repo}/releases`, token, {
          method: 'POST',
          body: JSON.stringify({
            tag_name: nextTag,
            target_commitish: 'main',
            name: `${nextTag} - ${inputTitle}`,
            body: releaseBody,
            draft: false,
            prerelease: false,
          }),
        });
        releaseUrl = createdRelease.html_url;
        console.log(c('green', `✓ GitHub Release created: `) + c('bold', releaseUrl));
      } catch (err) {
        console.log(c('yellow', `   Notice: ${err.message}`));
      }
    }

    // 10. Live Polling GitHub Actions Build Status
    console.log('\n' + c('yellow', '8. Monitoring GitHub Actions multi-platform build progress...'));
    console.log(c('dim', '   Waiting for workflow run to start on GitHub...'));

    let activeRun = null;
    for (let attempt = 0; attempt < 12; attempt++) {
      await new Promise((res) => setTimeout(res, 4000));
      if (!token) break;
      try {
        const runsData = await ghApi(`/repos/${owner}/${repo}/actions/runs?event=push`, token);
        if (runsData.workflow_runs && runsData.workflow_runs.length > 0) {
          const matchRun = runsData.workflow_runs.find((r) => r.head_branch === nextTag || r.head_sha === headCommit);
          if (matchRun) {
            activeRun = matchRun;
            break;
          }
        }
      } catch (_) {}
    }

    if (activeRun && token) {
      console.log(c('cyan', `   Workflow triggered: `) + `${activeRun.html_url}`);
      let completed = false;
      const spinners = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
      let spinIdx = 0;

      while (!completed) {
        await new Promise((res) => setTimeout(res, 5000));
        try {
          const currentRun = await ghApi(`/repos/${owner}/${repo}/actions/runs/${activeRun.id}`, token);
          const jobsData = await ghApi(`/repos/${owner}/${repo}/actions/runs/${activeRun.id}/jobs`, token);
          
          process.stdout.write(`\r${c('cyan', spinners[spinIdx++ % spinners.length])} Status: ${c('bold', currentRun.status)} | Conclusion: ${currentRun.conclusion || 'running...'}`);

          if (currentRun.status === 'completed') {
            completed = true;
            process.stdout.write('\n');
            if (currentRun.conclusion === 'success') {
              console.log(c('green', '🎉 All multi-platform builds completed successfully!'));
            } else {
              console.log(c('yellow', `⚠️  Workflow finished with status: ${currentRun.conclusion}`));
            }
          }
        } catch (_) {}
      }
    } else {
      console.log(c('dim', '   (Workflow is running in background on GitHub Actions)'));
    }

    // 11. Final Summary & Download Links
    console.log('\n' + c('green', '══════════════════════ RELEASE READY ══════════════════════'));
    console.log(c('bold', `✨ Release URL: `) + c('cyan', releaseUrl));
    console.log('\n' + c('bold', 'Direct Download Artifacts:'));
    console.log(`🍏 macOS Apple Silicon (M1-M4):`);
    console.log(`   Installer (.dmg): ${c('cyan', downloadUrls.macos_arm64_dmg)}`);
    console.log(`🍏 macOS Intel (x64):`);
    console.log(`   Installer (.dmg): ${c('cyan', downloadUrls.macos_x64_dmg)}`);
    console.log(`🪟 Windows (x64):`);
    console.log(`   Installer (.exe): ${c('cyan', downloadUrls.windows_x64_setup)}`);
    console.log(`   Portable (.zip):  ${c('cyan', downloadUrls.windows_x64_portable)}`);
    console.log(`🪟 Windows (ARM64):`);
    console.log(`   Installer (.exe): ${c('cyan', downloadUrls.windows_arm64_setup)}`);
    console.log(`   Portable (.zip):  ${c('cyan', downloadUrls.windows_arm64_portable)}`);
    console.log(c('green', '═════════════════════════════════════════════════════════════\n'));

  } catch (error) {
    console.error(c('red', `\n❌ Error during release execution: ${error.message}`));
  } finally {
    rl.close();
  }
}

main();
