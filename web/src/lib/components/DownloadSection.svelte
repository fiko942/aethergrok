<script lang="ts">
  import {
    githubRepo,
    fetchLiveLatestRelease,
    type LatestReleaseInfo
  } from '../data/downloads';
  import {
    Download,
    Check,
    Copy,
    Apple,
    Terminal,
    ExternalLink,
    RefreshCw,
    X,
    ShieldAlert,
    HelpCircle,
    Info,
    ArrowRight
  } from 'lucide-svelte';
  import { onMount } from 'svelte';

  let activePlatform = $state<'macOS' | 'Windows'>('macOS');
  let copiedCommand = $state(false);
  let releaseInfo = $state<LatestReleaseInfo | null>(null);
  let loadingRelease = $state(true);

  // Download Guide Popup Modal State
  let showModal = $state(false);
  let downloadedFileName = $state('');
  let downloadedFileUrl = $state('');
  let modalPlatform = $state<'macOS' | 'Windows'>('macOS');
  let modalCopied = $state(false);

  onMount(async () => {
    loadingRelease = true;
    releaseInfo = await fetchLiveLatestRelease();
    loadingRelease = false;
  });

  const displayVersion = $derived(releaseInfo?.tagName || 'v1.0.4');
  const publishedDate = $derived(
    releaseInfo?.publishedAt
      ? new Date(releaseInfo.publishedAt).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
      : 'Latest Release'
  );

  // Dynamic macOS assets
  const macArmAsset = $derived(
    releaseInfo?.assets.find((a) => a.name.includes('arm64') && a.name.endsWith('.dmg'))
  );
  const macIntelAsset = $derived(
    releaseInfo?.assets.find((a) => a.name.includes('amd64') && a.name.endsWith('.dmg'))
  );

  // Dynamic Windows assets
  const winSetupAsset = $derived(
    releaseInfo?.assets.find((a) => a.name.includes('amd64') && a.name.includes('setup.exe'))
  );
  const winPortableAsset = $derived(
    releaseInfo?.assets.find((a) => a.name.includes('amd64') && a.name.includes('portable.zip'))
  );

  function formatSize(bytes?: number): string {
    if (!bytes) return '6.8 MB';
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }

  const installCurl = `curl -fsSL https://raw.githubusercontent.com/fiko942/aethergrok/main/scripts/install-app.sh | bash`;

  function copyInstallScript() {
    navigator.clipboard.writeText(installCurl);
    copiedCommand = true;
    setTimeout(() => {
      copiedCommand = false;
    }, 2000);
  }

  function handleDownloadClick(platform: 'macOS' | 'Windows', filename: string, url: string) {
    downloadedFileName = filename;
    downloadedFileUrl = url;
    modalPlatform = platform;
    showModal = true;
  }

  const quarantineTerminalCmd = 'xattr -d com.apple.quarantine /Applications/AetherGrok.app';

  function copyModalQuarantine() {
    navigator.clipboard.writeText(quarantineTerminalCmd);
    modalCopied = true;
    setTimeout(() => {
      modalCopied = false;
    }, 2000);
  }
</script>

<section id="downloads" class="py-20 md:py-32 bg-white">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="text-center max-w-3xl mx-auto mb-16">
      <h2 class="text-xs font-bold uppercase tracking-widest text-blue-600 mb-2">
        Direct Binary Releases
      </h2>
      <p class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
        Download AetherGrok ({displayVersion})
      </p>
      <p class="mt-4 text-base sm:text-lg text-slate-600">
        Standalone native desktop packages for macOS and Windows, synchronized live with GitHub Releases ({publishedDate}).
      </p>
    </div>

    <!-- Platform Selector Tabs -->
    <div class="flex justify-center mb-10">
      <div class="inline-flex p-1.5 bg-slate-100 rounded-xl border border-slate-200">
        <button
          onclick={() => (activePlatform = 'macOS')}
          class="flex items-center gap-2 px-6 py-2.5 rounded-lg text-sm font-bold transition-all {activePlatform === 'macOS'
            ? 'bg-white text-blue-700 shadow-sm border border-slate-200/80'
            : 'text-slate-600 hover:text-slate-900'}"
        >
          <Apple class="w-4 h-4" />
          <span>macOS (Apple Silicon & Intel)</span>
        </button>
        <button
          onclick={() => (activePlatform = 'Windows')}
          class="flex items-center gap-2 px-6 py-2.5 rounded-lg text-sm font-bold transition-all {activePlatform === 'Windows'
            ? 'bg-white text-blue-700 shadow-sm border border-slate-200/80'
            : 'text-slate-600 hover:text-slate-900'}"
        >
          <Terminal class="w-4 h-4" />
          <span>Windows (x64 & ARM64)</span>
        </button>
      </div>
    </div>

    <!-- Download Cards Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-6 max-w-4xl mx-auto mb-12">
      {#if activePlatform === 'macOS'}
        <!-- macOS Apple Silicon -->
        <div class="relative bg-slate-50/70 border border-slate-200 rounded-2xl p-6 sm:p-7 flex flex-col justify-between hover:border-blue-300 hover:shadow-md transition-all">
          <div class="absolute -top-3 right-6 px-3 py-0.5 rounded-full text-[11px] font-bold uppercase tracking-wider bg-blue-600 text-white shadow-sm">
            Recommended
          </div>
          <div>
            <div class="flex items-center justify-between mb-3">
              <span class="text-xs font-bold uppercase tracking-wider text-slate-500">
                macOS Styled .dmg
              </span>
              <span class="text-xs font-mono font-medium text-slate-400 bg-white px-2 py-0.5 rounded border border-slate-200">
                {formatSize(macArmAsset?.size)}
              </span>
            </div>
            <h3 class="text-lg font-bold text-slate-900 mb-1">
              Apple Silicon (M1 / M2 / M3 / M4)
            </h3>
            <p class="text-xs font-mono text-slate-500 mb-6 truncate">
              {macArmAsset?.name || `AetherGrok-${displayVersion.replace('v', '')}-macOS-arm64.dmg`}
            </p>
          </div>
          <div class="space-y-2">
            <a
              href="{macArmAsset?.browser_download_url || `${githubRepo}/releases/latest`}"
              onclick={() => handleDownloadClick('macOS', macArmAsset?.name || `AetherGrok-${displayVersion.replace('v', '')}-macOS-arm64.dmg`, macArmAsset?.browser_download_url || `${githubRepo}/releases/latest`)}
              class="w-full inline-flex items-center justify-center gap-2 px-5 py-3 rounded-xl font-bold text-sm text-white bg-blue-600 hover:bg-blue-700 shadow-sm hover:shadow active:scale-[0.99] transition-all"
            >
              <Download class="w-4 h-4" />
              <span>Download Apple Silicon DMG</span>
            </a>
          </div>
        </div>

        <!-- macOS Intel -->
        <div class="relative bg-slate-50/70 border border-slate-200 rounded-2xl p-6 sm:p-7 flex flex-col justify-between hover:border-blue-300 hover:shadow-md transition-all">
          <div>
            <div class="flex items-center justify-between mb-3">
              <span class="text-xs font-bold uppercase tracking-wider text-slate-500">
                macOS Styled .dmg
              </span>
              <span class="text-xs font-mono font-medium text-slate-400 bg-white px-2 py-0.5 rounded border border-slate-200">
                {formatSize(macIntelAsset?.size)}
              </span>
            </div>
            <h3 class="text-lg font-bold text-slate-900 mb-1">
              Intel x86_64
            </h3>
            <p class="text-xs font-mono text-slate-500 mb-6 truncate">
              {macIntelAsset?.name || `AetherGrok-${displayVersion.replace('v', '')}-macOS-amd64.dmg`}
            </p>
          </div>
          <div class="space-y-2">
            <a
              href="{macIntelAsset?.browser_download_url || `${githubRepo}/releases/latest`}"
              onclick={() => handleDownloadClick('macOS', macIntelAsset?.name || `AetherGrok-${displayVersion.replace('v', '')}-macOS-amd64.dmg`, macIntelAsset?.browser_download_url || `${githubRepo}/releases/latest`)}
              class="w-full inline-flex items-center justify-center gap-2 px-5 py-3 rounded-xl font-bold text-sm text-slate-700 bg-white hover:bg-slate-100 border border-slate-300 shadow-sm hover:shadow active:scale-[0.99] transition-all"
            >
              <Download class="w-4 h-4 text-slate-500" />
              <span>Download Intel Mac DMG</span>
            </a>
          </div>
        </div>
      {:else}
        <!-- Windows Latest Release Card -->
        <div class="relative bg-slate-50/70 border border-slate-200 rounded-2xl p-6 sm:p-7 flex flex-col justify-between hover:border-blue-300 hover:shadow-md transition-all">
          <div class="absolute -top-3 right-6 px-3 py-0.5 rounded-full text-[11px] font-bold uppercase tracking-wider bg-blue-600 text-white shadow-sm">
            GitHub Releases
          </div>
          <div>
            <div class="flex items-center justify-between mb-3">
              <span class="text-xs font-bold uppercase tracking-wider text-slate-500">
                Windows Setup Installer (.exe)
              </span>
              <span class="text-xs font-mono font-medium text-slate-400 bg-white px-2 py-0.5 rounded border border-slate-200">
                {formatSize(winSetupAsset?.size)}
              </span>
            </div>
            <h3 class="text-lg font-bold text-slate-900 mb-1">
              Windows x64 / ARM64 Setup
            </h3>
            <p class="text-xs font-mono text-slate-500 mb-6 truncate">
              {winSetupAsset?.name || 'AetherGrok-windows-setup.exe'}
            </p>
          </div>
          <div class="space-y-2">
            <a
              href="{winSetupAsset?.browser_download_url || `${githubRepo}/releases/latest`}"
              onclick={() => handleDownloadClick('Windows', winSetupAsset?.name || 'AetherGrok-windows-setup.exe', winSetupAsset?.browser_download_url || `${githubRepo}/releases/latest`)}
              class="w-full inline-flex items-center justify-center gap-2 px-5 py-3 rounded-xl font-bold text-sm text-white bg-blue-600 hover:bg-blue-700 shadow-sm hover:shadow active:scale-[0.99] transition-all"
            >
              <Download class="w-4 h-4" />
              <span>Download Windows Setup (Latest)</span>
            </a>
          </div>
        </div>

        <!-- Windows All Assets Directory Card -->
        <div class="relative bg-slate-50/70 border border-slate-200 rounded-2xl p-6 sm:p-7 flex flex-col justify-between hover:border-blue-300 hover:shadow-md transition-all">
          <div>
            <div class="flex items-center justify-between mb-3">
              <span class="text-xs font-bold uppercase tracking-wider text-slate-500">
                Release Assets Hub
              </span>
              <span class="text-xs font-mono font-medium text-slate-400 bg-white px-2 py-0.5 rounded border border-slate-200">
                All Formats
              </span>
            </div>
            <h3 class="text-lg font-bold text-slate-900 mb-1">
              All Packages & Portable Zip
            </h3>
            <p class="text-xs font-mono text-slate-500 mb-6 truncate">
              github.com/fiko942/aethergrok/releases/latest
            </p>
          </div>
          <div class="space-y-2">
            <a
              href="{githubRepo}/releases/latest"
              target="_blank"
              rel="noopener noreferrer"
              class="w-full inline-flex items-center justify-center gap-2 px-5 py-3 rounded-xl font-bold text-sm text-slate-700 bg-white hover:bg-slate-100 border border-slate-300 shadow-sm hover:shadow active:scale-[0.99] transition-all"
            >
              <span>Explore GitHub Release Assets</span>
              <ExternalLink class="w-4 h-4 text-slate-500" />
            </a>
          </div>
        </div>
      {/if}
    </div>

    <!-- Verified Live Terminal 1-Liner Installer -->
    <div class="max-w-3xl mx-auto bg-slate-900 rounded-2xl p-5 text-slate-200 shadow-lg border border-slate-800">
      <div class="flex items-center justify-between mb-2">
        <span class="text-xs font-mono text-slate-400 font-semibold flex items-center gap-1.5">
          <Terminal class="w-4 h-4 text-blue-400" />
          <span>Quick Install via Terminal (macOS / Linux) — Automated & Verified</span>
        </span>
        <button
          onclick={copyInstallScript}
          class="flex items-center gap-1 text-xs text-slate-300 hover:text-white px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 transition-colors"
        >
          {#if copiedCommand}
            <Check class="w-3.5 h-3.5 text-emerald-400" />
            <span class="text-emerald-400 font-bold">Copied</span>
          {:else}
            <Copy class="w-3.5 h-3.5" />
            <span>Copy Command</span>
          {/if}
        </button>
      </div>
      <div class="font-mono text-xs sm:text-sm text-emerald-300 select-all overflow-x-auto py-1">
        {installCurl}
      </div>
      <p class="text-[11px] text-slate-400 mt-2">
        The script automatically detects your Mac architecture (Apple Silicon or Intel), downloads the latest release DMG from GitHub, installs to <code>/Applications/AetherGrok.app</code>, and clears Apple Quarantine attributes for seamless instant launching.
      </p>
    </div>
  </div>
</section>

<!-- Post-Download Instructions Modal Popup -->
{#if showModal}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/75 backdrop-blur-sm animate-fade-in"
    role="dialog"
    aria-modal="true"
    tabindex="-1"
    onclick={() => (showModal = false)}
  >
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div
      class="bg-white rounded-2xl max-w-xl w-full border border-slate-200 shadow-2xl overflow-hidden flex flex-col my-auto"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Modal Header -->
      <div class="px-6 py-4 bg-slate-50 border-b border-slate-200 flex items-center justify-between">
        <div class="flex items-center gap-2.5">
          <div class="w-8 h-8 rounded-lg bg-blue-100 text-blue-700 flex items-center justify-center">
            <Download class="w-4 h-4" />
          </div>
          <div>
            <h3 class="text-base font-bold text-slate-900 leading-none">
              Download Started!
            </h3>
            <p class="text-xs text-slate-500 mt-1 truncate max-w-xs sm:max-w-md">
              {downloadedFileName}
            </p>
          </div>
        </div>
        <button
          onclick={() => (showModal = false)}
          class="p-1.5 text-slate-400 hover:text-slate-700 rounded-lg hover:bg-slate-200/60 transition-colors"
          title="Close dialog"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 space-y-5 text-sm text-slate-700 max-h-[75vh] overflow-y-auto">
        {#if modalPlatform === 'macOS'}
          <!-- macOS First-Launch Guidance -->
          <div class="p-4 rounded-xl bg-amber-50 border border-amber-200/90 text-amber-950 space-y-2">
            <div class="flex items-center gap-2 font-bold text-xs uppercase tracking-wider text-amber-800">
              <ShieldAlert class="w-4 h-4 text-amber-600" />
              <span>macOS Gatekeeper First-Launch Notice</span>
            </div>
            <p class="text-xs leading-relaxed text-amber-900">
              Because AetherGrok is free open-source software signed ad-hoc, macOS may display a warning: <em>“AetherGrok.app cannot be opened because Apple cannot check it for malicious software”</em>.
            </p>
          </div>

          <div class="space-y-3">
            <h4 class="font-bold text-slate-900 text-xs uppercase tracking-wider">
              Quick 1-Second Terminal Fix:
            </h4>
            <div class="bg-slate-900 rounded-xl p-3.5 text-slate-200 border border-slate-800">
              <div class="flex items-center justify-between text-xs text-slate-400 mb-1.5">
                <span class="font-mono text-[11px]">Copy to Terminal:</span>
                <button
                  onclick={copyModalQuarantine}
                  class="text-blue-400 hover:text-blue-300 font-bold flex items-center gap-1 text-xs"
                >
                  {#if modalCopied}
                    <Check class="w-3.5 h-3.5 text-emerald-400" />
                    <span class="text-emerald-400">Copied!</span>
                  {:else}
                    <Copy class="w-3.5 h-3.5" />
                    <span>Copy Command</span>
                  {/if}
                </button>
              </div>
              <div class="font-mono text-xs text-emerald-300 select-all overflow-x-auto py-0.5">
                {quarantineTerminalCmd}
              </div>
            </div>
          </div>

          <div class="space-y-2">
            <h4 class="font-bold text-slate-900 text-xs uppercase tracking-wider">
              Or via macOS System Settings:
            </h4>
            <ol class="space-y-1.5 text-xs text-slate-600 list-decimal list-inside leading-relaxed">
              <li>Open <strong>System Settings</strong> &rarr; <strong>Privacy & Security</strong></li>
              <li>Scroll down to <strong>Security</strong> section</li>
              <li>Click <strong>Open Anyway</strong> next to the AetherGrok notice</li>
            </ol>
          </div>
        {:else}
          <!-- Windows Guidance -->
          <div class="space-y-3 text-xs leading-relaxed text-slate-600">
            <p>
              1. Run the downloaded <code>{downloadedFileName}</code> installer.
            </p>
            <p>
              2. If Windows SmartScreen prompts <em>"Windows protected your PC"</em>, click <strong>More info</strong> &rarr; <strong>Run anyway</strong>.
            </p>
            <p>
              3. AetherGrok will automatically configure environment paths and launch smoothly.
            </p>
          </div>
        {/if}
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-3.5 bg-slate-50 border-t border-slate-200 flex items-center justify-between">
        <a
          href="#gatekeeper"
          onclick={() => (showModal = false)}
          class="text-xs text-blue-600 hover:text-blue-800 font-semibold flex items-center gap-1"
        >
          <span>View Full Gatekeeper Guide</span>
          <ArrowRight class="w-3.5 h-3.5" />
        </a>
        <button
          onclick={() => (showModal = false)}
          class="px-4 py-2 rounded-xl text-xs font-bold text-white bg-blue-600 hover:bg-blue-700 transition-colors shadow-sm"
        >
          Got it!
        </button>
      </div>
    </div>
  </div>
{/if}
