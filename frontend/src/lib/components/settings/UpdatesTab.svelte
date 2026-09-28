<script lang="ts">
  import { onMount } from 'svelte';
  import { updaterStore } from '$lib/stores/updater.svelte';
  import { parsePackageInfo } from '$lib/utils/packageFormatter';
  import {
    RotateCw,
    CheckCircle2,
    Sparkles,
    DownloadCloud,
    Download,
    ExternalLink,
    ChevronDown,
    ChevronUp,
    Calendar,
    AlertCircle,
    Package,
    Tag,
    Check,
    ArrowRight,
    Laptop,
    FileCode2,
    RefreshCw,
    ShieldCheck,
    XCircle,
    Play,
    Zap,
    HardDriveDownload,
    Monitor,
    TerminalSquare
  } from 'lucide-svelte';
  import Button from '$lib/antd/Button.svelte';
  import Tooltip from '$lib/antd/Tooltip.svelte';
  import Badge from '$lib/antd/Badge.svelte';

  // Expanded release item IDs
  let expandedReleases = $state<Record<string, boolean>>({});

  function toggleExpand(version: string) {
    expandedReleases[version] = !expandedReleases[version];
  }

  function formatBytes(bytes: number): string {
    if (!bytes || bytes <= 0) return '0 B';
    const mb = bytes / (1024 * 1024);
    if (mb >= 1) return `${mb.toFixed(1)} MB`;
    const kb = bytes / 1024;
    return `${kb.toFixed(0)} KB`;
  }

  function formatDate(dateStr: string): string {
    if (!dateStr) return '';
    try {
      const d = new Date(dateStr);
      if (isNaN(d.getTime())) return dateStr;
      return d.toLocaleDateString(undefined, {
        year: 'numeric',
        month: 'short',
        day: 'numeric'
      });
    } catch {
      return dateStr;
    }
  }

  function formatTimestamp(isoStr: string | null): string {
    if (!isoStr) return 'Never';
    try {
      const d = new Date(isoStr);
      if (isNaN(d.getTime())) return isoStr;
      return d.toLocaleTimeString(undefined, {
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      }) + ' (' + d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' }) + ')';
    } catch {
      return isoStr;
    }
  }

  onMount(() => {
    // If no releases loaded yet, fetch changelog history and check updates
    if (updaterStore.allReleases.length === 0) {
      updaterStore.fetchChangelogHistory();
    }
    // Automatically expand the first release
    if (updaterStore.allReleases.length > 0 && updaterStore.allReleases[0]) {
      expandedReleases[updaterStore.allReleases[0].version] = true;
    }
  });

  $effect(() => {
    // Auto-expand latest release when releases update if none expanded
    if (updaterStore.allReleases.length > 0 && Object.keys(expandedReleases).length === 0) {
      const first = updaterStore.allReleases[0];
      if (first) {
        expandedReleases[first.version] = true;
      }
    }
  });
</script>

<div class="flex flex-col h-full space-y-5 font-serif">
  <!-- Top Bar: Title & Live Header Actions -->
  <div class="flex items-center justify-between gap-4 pb-1">
    <div class="min-w-0 flex-1">
      <h3 class="font-serif-display text-base font-semibold text-ant-text flex items-center gap-2.5">
        <span>Updates & Release Manager</span>
        <span class="px-2 py-0.5 text-[10.5px] font-mono font-medium bg-ant-primary/15 text-ant-primary rounded-full">
          Current: v{updaterStore.currentVersion}
        </span>
      </h3>
      <p class="font-serif text-xs text-ant-text-secondary mt-0.5">
        1-Click in-app background download, integrity verification, and instant self-installation.
      </p>
    </div>

    <div class="flex items-center space-x-2 flex-shrink-0">
      <button
        type="button"
        onclick={() => updaterStore.checkForUpdates(false)}
        disabled={updaterStore.checking || updaterStore.isInstalling}
        class="px-3.5 py-1.5 rounded-lg text-xs font-serif font-medium bg-ant-primary text-white hover:bg-ant-primary-hover active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed transition flex items-center gap-2 shadow-xs cursor-pointer whitespace-nowrap flex-shrink-0"
      >
        <RotateCw size={13} class={updaterStore.checking ? 'animate-spin' : ''} />
        <span class="whitespace-nowrap">{updaterStore.checking ? 'Checking...' : 'Check for Updates'}</span>
      </button>

      <button
        type="button"
        onclick={() => updaterStore.openDownload('https://github.com/fiko942/aethergrok/releases')}
        class="px-2.5 py-1.5 rounded-lg text-xs font-serif text-ant-text-secondary hover:text-ant-text bg-ant-bg-secondary hover:bg-ant-bg-tertiary transition flex items-center gap-1.5 cursor-pointer whitespace-nowrap flex-shrink-0"
        title="View repository releases on GitHub"
      >
        <ExternalLink size={12} class="text-ant-primary" />
        <span>GitHub</span>
      </button>
    </div>
  </div>

  <!-- Update Available Hero Card with 1-Click Action & Telemetry -->
  {#if updaterStore.updateAvailable && updaterStore.latestRelease}
    <div class="p-5 rounded-2xl bg-gradient-to-br from-ant-primary/15 via-ant-bg-secondary to-ant-bg-secondary shadow-lg space-y-4 animate-in fade-in zoom-in-95 duration-200">
      <div class="flex items-start justify-between gap-4">
        <div class="space-y-1.5 flex-1">
          <!-- Update Available Header Banner -->
          <div class="flex items-center gap-2 flex-wrap">
            <span class="px-2.5 py-0.5 text-xs font-mono font-bold bg-amber-500 text-white rounded-md shadow-xs flex items-center gap-1">
              <Sparkles size={12} />
              v{updaterStore.latestRelease.version} Ready
            </span>
            <span class="text-xs text-ant-text-secondary font-mono flex items-center gap-1.5">
              <span>Installed: <strong>v{updaterStore.currentVersion}</strong></span>
              <ArrowRight size={11} class="text-ant-primary" />
              <span class="font-bold text-ant-primary">v{updaterStore.latestRelease.version}</span>
            </span>
            {#if updaterStore.latestRelease.publishedAt}
              <span class="text-[11px] text-ant-text-muted flex items-center gap-1">
                <Calendar size={11} /> {formatDate(updaterStore.latestRelease.publishedAt)}
              </span>
            {/if}
          </div>

          <h4 class="font-serif-display text-sm font-bold text-ant-text pt-1">
            {updaterStore.latestRelease.title || `AetherGrok v${updaterStore.latestRelease.version} Released`}
          </h4>

          {#if updaterStore.latestRelease.highlights && updaterStore.latestRelease.highlights.length > 0}
            <ul class="space-y-1 pt-1 pl-0.5">
              {#each updaterStore.latestRelease.highlights.slice(0, 3) as item}
                <li class="text-xs text-ant-text-secondary flex items-start gap-2 leading-relaxed">
                  <span class="text-ant-primary font-bold leading-none mt-1">✦</span>
                  <span>{item}</span>
                </li>
              {/each}
            </ul>
          {/if}
        </div>
      </div>

      <!-- Live Interactive Download / Installation Progress Bar -->
      {#if updaterStore.isInstalling}
        <div class="p-4 rounded-xl bg-ant-bg/95 space-y-3 shadow-inner">
          <div class="flex items-center justify-between text-xs">
            <div class="flex items-center gap-2">
              {#if updaterStore.installProgress.stage === 'downloading'}
                <HardDriveDownload size={15} class="text-ant-primary animate-bounce" />
              {:else if updaterStore.installProgress.stage === 'verifying'}
                <ShieldCheck size={15} class="text-indigo-400 animate-pulse" />
              {:else if updaterStore.installProgress.stage === 'installing'}
                <RotateCw size={15} class="text-amber-400 animate-spin" />
              {:else if updaterStore.installProgress.stage === 'ready'}
                <CheckCircle2 size={15} class="text-emerald-400" />
              {/if}
              <span class="font-semibold text-ant-text">
                {updaterStore.installProgress.message || 'Processing update...'}
              </span>
            </div>

            <div class="flex items-center gap-3 font-mono text-[11px] text-ant-text-secondary">
              {#if updaterStore.installProgress.stage === 'downloading'}
                <span class="text-ant-primary font-bold">{updaterStore.installProgress.speedFormatted}</span>
                <span>
                  {formatBytes(updaterStore.installProgress.downloadedBytes)}
                  {#if updaterStore.installProgress.totalBytes > 0}
                    / {formatBytes(updaterStore.installProgress.totalBytes)}
                  {/if}
                </span>
              {/if}
              <span class="font-bold text-ant-text px-1.5 py-0.5 rounded bg-ant-bg-tertiary">
                {updaterStore.installProgress.percent.toFixed(1)}%
              </span>
            </div>
          </div>

          <!-- Progress track and filled bar -->
          <div class="w-full h-2 rounded-full bg-ant-bg-tertiary overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-300 ease-out {
                updaterStore.installProgress.stage === 'verifying'
                  ? 'bg-gradient-to-r from-indigo-500 to-purple-500 animate-pulse'
                  : updaterStore.installProgress.stage === 'installing' || updaterStore.installProgress.stage === 'ready'
                  ? 'bg-gradient-to-r from-amber-500 to-emerald-500'
                  : 'bg-gradient-to-r from-ant-primary via-emerald-400 to-ant-primary'
              }"
              style="width: {Math.max(3, updaterStore.installProgress.percent)}%"
            ></div>
          </div>

          <div class="flex items-center justify-between text-[11px] text-ant-text-secondary pt-0.5">
            <span class="capitalize flex items-center gap-1.5">
              <span class="w-2 h-2 rounded-full {
                updaterStore.installProgress.stage === 'ready' ? 'bg-emerald-500' : 'bg-ant-primary animate-ping'
              }"></span>
              Stage: <strong>{updaterStore.installProgress.stage}</strong>
            </span>

            {#if updaterStore.installProgress.stage === 'downloading'}
              <button
                type="button"
                onclick={() => updaterStore.cancelDownload()}
                class="text-rose-500 hover:text-rose-600 font-sans hover:underline flex items-center gap-1 cursor-pointer"
              >
                <XCircle size={12} /> Cancel Download
              </button>
            {/if}
          </div>
        </div>
      {/if}

      <!-- Diagnostic Error Banner if installation failed -->
      {#if updaterStore.installError}
        <div class="p-3 rounded-xl bg-rose-500/10 text-rose-600 dark:text-rose-400 text-xs flex items-center justify-between gap-2">
          <div class="flex items-center gap-2">
            <AlertCircle size={15} class="shrink-0" />
            <span>{updaterStore.installError}</span>
          </div>
          <button
            type="button"
            onclick={() => updaterStore.startDownloadAndInstall()}
            class="px-2.5 py-1 rounded-lg bg-rose-500/20 hover:bg-rose-500/30 text-rose-600 dark:text-rose-300 text-[11px] font-sans font-semibold transition cursor-pointer"
          >
            Retry
          </button>
        </div>
      {/if}

      <!-- Hero Banner Action Buttons -->
      <div class="pt-3 flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-2.5 flex-wrap">
          <!-- 1-Click Automatic In-App Download & Self-Install -->
          <button
            type="button"
            onclick={() => updaterStore.startDownloadAndInstall()}
            disabled={updaterStore.isInstalling}
            class="px-4 py-2 rounded-xl text-xs font-serif font-semibold bg-ant-primary hover:bg-ant-primary-hover text-white shadow-md shadow-ant-primary/25 disabled:opacity-50 disabled:cursor-not-allowed transition flex items-center gap-2 cursor-pointer active:scale-95"
          >
            {#if updaterStore.isInstalling}
              <RotateCw size={15} class="animate-spin" />
              <span>Installing Update...</span>
            {:else}
              <Zap size={15} class="fill-current text-amber-300" />
              <span>Update Now (Auto-Install)</span>
              {#if updaterStore.matchedAsset && updaterStore.matchedAsset.size > 0}
                <span class="opacity-80 font-mono text-[10.5px]">({formatBytes(updaterStore.matchedAsset.size)})</span>
              {/if}
            {/if}
          </button>

          <!-- Manual Download fallback button -->
          {#if updaterStore.matchedAsset}
            <button
              type="button"
              onclick={() => updaterStore.openDownload(updaterStore.matchedAsset?.downloadUrl)}
              disabled={updaterStore.isInstalling}
              class="px-3 py-2 rounded-xl text-xs font-serif text-ant-text-secondary hover:text-ant-text bg-ant-bg-tertiary/60 hover:bg-ant-bg-tertiary transition flex items-center gap-1.5 cursor-pointer"
              title="Download binary package manually via browser"
            >
              <Download size={13} />
              <span>Manual Download</span>
            </button>
          {/if}
        </div>

        <button
          type="button"
          onclick={() => updaterStore.openDownload(`https://github.com/fiko942/aethergrok/releases/tag/${updaterStore.latestRelease?.tagName}`)}
          class="text-xs text-ant-text-secondary hover:text-ant-primary transition flex items-center gap-1 cursor-pointer font-sans"
        >
          <span>Full Release Notes</span>
          <ExternalLink size={11} />
        </button>
      </div>
    </div>
  {:else}
    <!-- Up-to-Date State Summary Card -->
    <div class="p-4 rounded-xl bg-ant-bg-secondary flex items-center justify-between flex-wrap gap-3">
      <div class="flex items-center space-x-3">
        <div class="w-9 h-9 rounded-xl flex items-center justify-center bg-emerald-500/15 text-emerald-500 dark:bg-emerald-500/10">
          <CheckCircle2 size={18} />
        </div>
        <div>
          <div class="flex items-center gap-2">
            <span class="text-xs font-semibold text-ant-text">AetherGrok is up-to-date</span>
            <span class="px-2 py-0.2 text-[10px] font-mono font-semibold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 rounded-full flex items-center gap-1">
              <Check size={10} /> v{updaterStore.currentVersion}
            </span>
          </div>
          <p class="text-[11px] text-ant-text-secondary font-serif mt-0.5">
            Last checked: <strong class="font-mono text-ant-text">{formatTimestamp(updaterStore.lastChecked)}</strong>
          </p>
        </div>
      </div>

      {#if updaterStore.error}
        <span class="text-xs text-rose-500 flex items-center gap-1">
          <AlertCircle size={13} /> {updaterStore.error}
        </span>
      {/if}
    </div>
  {/if}

  <!-- Version Changelog History List -->
  <div class="space-y-3 pt-1">
    <div class="flex items-center justify-between">
      <h4 class="font-serif-display text-sm font-semibold text-ant-text flex items-center gap-2">
        <Package size={15} class="text-ant-primary" />
        <span>Version History & Changelogs</span>
      </h4>
      <span class="text-xs text-ant-text-secondary font-mono">
        {updaterStore.allReleases.length} {updaterStore.allReleases.length === 1 ? 'release' : 'releases'} recorded
      </span>
    </div>

    {#if updaterStore.allReleases.length === 0}
      <div class="p-8 rounded-xl bg-ant-bg-secondary text-center space-y-2">
        <Package size={28} class="mx-auto text-ant-text-secondary opacity-40" />
        <p class="text-xs text-ant-text-secondary">No release history found.</p>
        <button
          type="button"
          onclick={() => updaterStore.fetchChangelogHistory()}
          class="px-3 py-1.5 rounded-lg text-xs font-serif bg-ant-bg-tertiary hover:bg-ant-bg-secondary text-ant-text transition cursor-pointer"
        >
          Load Releases
        </button>
      </div>
    {:else}
      <div class="space-y-2.5">
        {#each updaterStore.allReleases as release (release.version)}
          {@const isExpanded = !!expandedReleases[release.version]}
          {@const isCurrent = release.version === updaterStore.currentVersion}
          {@const isNewer = release.isLatest && updaterStore.updateAvailable}

          <div
            class="rounded-xl transition-all duration-200 overflow-hidden {
              isNewer
                ? 'bg-amber-500/10'
                : isCurrent
                ? 'bg-ant-primary/10'
                : 'bg-ant-bg-secondary hover:bg-ant-bg-tertiary/40'
            }"
          >
            <!-- Release Card Header -->
            <button
              type="button"
              onclick={() => toggleExpand(release.version)}
              class="w-full px-4 py-3 flex items-center justify-between gap-4 text-left cursor-pointer select-none transition"
            >
              <div class="flex items-center gap-2.5 min-w-0 flex-1">
                <span class="font-mono text-xs font-bold px-2 py-0.5 rounded-md flex-shrink-0 {
                  isNewer
                    ? 'bg-amber-500 text-white'
                    : isCurrent
                    ? 'bg-ant-primary text-white'
                    : 'bg-ant-bg-tertiary text-ant-text'
                }">
                  v{release.version}
                </span>

                <span class="text-xs font-semibold text-ant-text font-serif truncate">
                  {release.title || `Release v${release.version}`}
                </span>

                {#if isCurrent}
                  <span class="px-2 py-0.5 text-[10px] font-mono bg-ant-primary/20 text-ant-primary rounded-full flex items-center gap-1 flex-shrink-0">
                    <Check size={9} /> Installed
                  </span>
                {/if}

                {#if isNewer}
                  <span class="px-2 py-0.5 text-[10px] font-mono font-medium tracking-wide bg-amber-500/20 text-amber-500 border border-amber-500/30 rounded-full flex-shrink-0">
                    Latest
                  </span>
                {/if}
              </div>

              <div class="flex items-center space-x-3 flex-shrink-0">
                {#if release.publishedAt}
                  <span class="text-[11px] text-ant-text-secondary font-mono whitespace-nowrap flex items-center gap-1.5">
                    <Calendar size={12} class="flex-shrink-0 opacity-70" />
                    <span>{formatDate(release.publishedAt)}</span>
                  </span>
                {/if}

                <div class="text-ant-text-secondary flex-shrink-0">
                  {#if isExpanded}
                    <ChevronUp size={15} />
                  {:else}
                    <ChevronDown size={15} />
                  {/if}
                </div>
              </div>
            </button>

            <!-- Release Details Accordion Body -->
            {#if isExpanded}
              <div class="px-4 pb-4 pt-1 space-y-3 font-serif">
                <!-- Highlights / Changelog bullets -->
                {#if release.highlights && release.highlights.length > 0}
                  <div class="space-y-1.5 pt-1">
                    <span class="text-[11px] font-semibold text-ant-text-secondary uppercase tracking-wider">
                      Key Highlights & Changes:
                    </span>
                    <ul class="space-y-1 pl-1">
                      {#each release.highlights as highlight}
                        <li class="text-xs text-ant-text-secondary flex items-start gap-2 leading-relaxed">
                          <span class="text-ant-primary font-bold leading-none mt-1.5">✦</span>
                          <span>{highlight}</span>
                        </li>
                      {/each}
                    </ul>
                  </div>
                {/if}

                <!-- Full release body if available and no highlights -->
                {#if (!release.highlights || release.highlights.length === 0) && release.body}
                  <div class="text-xs text-ant-text-secondary whitespace-pre-line leading-relaxed max-h-48 overflow-y-auto pr-2 pl-2">
                    {release.body}
                  </div>
                {/if}

                <!-- Downloadable Assets List with Humanized Labels -->
                {#if release.assets && release.assets.length > 0}
                  <div class="pt-2 space-y-2">
                    <span class="text-[11px] font-semibold text-ant-text-secondary uppercase tracking-wider flex items-center gap-1">
                      <Download size={11} /> Available Installers & Packages:
                    </span>

                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      {#each release.assets as asset}
                        {@const info = parsePackageInfo(asset.name)}
                        <button
                          type="button"
                          onclick={() => updaterStore.openDownload(asset.downloadUrl)}
                          class="p-2.5 rounded-lg bg-ant-bg-tertiary/60 hover:bg-ant-bg-tertiary text-left transition flex items-center justify-between gap-2 text-xs group cursor-pointer"
                        >
                          <div class="flex items-center space-x-2.5 min-w-0">
                            {#if info.osName === 'macOS'}
                              <Laptop size={14} class="text-ant-primary shrink-0 group-hover:scale-110 transition" />
                            {:else if info.osName === 'Windows'}
                              <Monitor size={14} class="text-blue-400 shrink-0 group-hover:scale-110 transition" />
                            {:else}
                              <Package size={14} class="text-ant-primary shrink-0 group-hover:scale-110 transition" />
                            {/if}

                            <div class="min-w-0">
                              <div class="text-[11.5px] font-medium text-ant-text truncate">
                                {info.label}
                              </div>
                              <div class="text-[10px] text-ant-text-muted font-mono truncate">
                                {asset.name}
                              </div>
                            </div>
                          </div>

                          <div class="flex items-center space-x-1.5 shrink-0 text-ant-text-secondary text-[10.5px] font-mono">
                            {#if asset.size > 0}
                              <span>{formatBytes(asset.size)}</span>
                            {/if}
                            <Download size={11} class="group-hover:text-ant-primary transition" />
                          </div>
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
