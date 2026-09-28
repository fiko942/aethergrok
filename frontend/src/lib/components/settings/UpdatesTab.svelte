<script lang="ts">
  import { onMount } from 'svelte';
  import { updaterStore } from '$lib/stores/updater.svelte';
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
    RefreshCw
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

<div class="flex flex-col h-full space-y-6 font-serif">
  <!-- Header Section -->
  <div class="flex items-start justify-between">
    <div>
      <h3 class="font-serif-display text-base font-semibold text-ant-text flex items-center gap-2">
        <span>Updates & Release Changelog</span>
        <span class="px-2 py-0.5 text-[10px] font-mono bg-ant-primary/15 text-ant-primary rounded-full">
          v{updaterStore.currentVersion}
        </span>
      </h3>
      <p class="font-serif text-xs text-ant-text-secondary mt-0.5">
        Live GitHub release integration, version change history, and native installers.
      </p>
    </div>

    <button
      type="button"
      onclick={() => updaterStore.openDownload('https://github.com/fiko942/grok-build/releases')}
      class="px-2.5 py-1.5 rounded-lg text-xs font-serif text-ant-text-secondary hover:text-ant-text bg-ant-bg border border-ant-border-secondary dark:border-white/5 hover:bg-ant-bg-tertiary transition flex items-center gap-1.5 shadow-2xs cursor-pointer"
      title="View all releases on GitHub"
    >
      <ExternalLink size={12} class="text-ant-primary" />
      <span>GitHub Releases</span>
    </button>
  </div>

  <!-- Live Status Card -->
  <div class="p-4 rounded-xl bg-ant-bg border border-ant-border-secondary dark:border-white/5 shadow-2xs space-y-3">
    <div class="flex items-center justify-between flex-wrap gap-3">
      <div class="flex items-center space-x-3">
        <!-- Status Icon Avatar -->
        <div class="w-10 h-10 rounded-xl flex items-center justify-center shadow-xs {
          updaterStore.updateAvailable
            ? 'bg-amber-500/15 text-amber-500 border border-amber-500/30 dark:bg-amber-500/10'
            : 'bg-emerald-500/15 text-emerald-500 border border-emerald-500/30 dark:bg-emerald-500/10'
        }">
          {#if updaterStore.updateAvailable}
            <Sparkles size={20} class="animate-pulse" />
          {:else}
            <CheckCircle2 size={20} />
          {/if}
        </div>

        <div class="space-y-0.5">
          <div class="flex items-center gap-2">
            <span class="text-xs font-semibold text-ant-text">
              Installed Version:
            </span>
            <span class="px-2 py-0.5 text-[11px] font-mono font-bold bg-ant-bg-secondary border border-ant-border-secondary dark:border-white/5 text-ant-text rounded-md">
              v{updaterStore.currentVersion}
            </span>

            {#if updaterStore.updateAvailable}
              <span class="px-2 py-0.5 text-[10.5px] font-mono font-semibold bg-amber-500/20 text-amber-600 dark:text-amber-400 border border-amber-500/30 rounded-full animate-pulse flex items-center gap-1">
                <span class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
                Update Available (v{updaterStore.latestVersion})
              </span>
            {:else}
              <span class="px-2 py-0.5 text-[10.5px] font-mono font-semibold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30 rounded-full flex items-center gap-1">
                <Check size={11} />
                Up to date
              </span>
            {/if}
          </div>

          <div class="text-[11px] text-ant-text-secondary font-serif flex items-center gap-2">
            <span>Last checked: <strong class="font-mono text-ant-text">{formatTimestamp(updaterStore.lastChecked)}</strong></span>
            {#if updaterStore.error}
              <span class="text-rose-500 flex items-center gap-1">
                <AlertCircle size={12} /> {updaterStore.error}
              </span>
            {/if}
          </div>
        </div>
      </div>

      <!-- Action Button -->
      <button
        type="button"
        onclick={() => updaterStore.checkForUpdates(false)}
        disabled={updaterStore.checking}
        class="px-3.5 py-1.5 rounded-lg text-xs font-serif font-medium bg-ant-primary text-white hover:bg-ant-primary-hover active:scale-95 disabled:opacity-50 disabled:cursor-not-allowed transition flex items-center gap-2 shadow-xs cursor-pointer"
      >
        <RotateCw size={13} class={updaterStore.checking ? 'animate-spin' : ''} />
        <span>{updaterStore.checking ? 'Checking for updates...' : 'Check for Updates'}</span>
      </button>
    </div>
  </div>

  <!-- Update Available Hero Banner -->
  {#if updaterStore.updateAvailable && updaterStore.latestRelease}
    <div class="p-5 rounded-2xl bg-gradient-to-r from-ant-primary/10 via-ant-primary/5 to-transparent border border-ant-primary/30 shadow-lg space-y-4 animate-in fade-in zoom-in-95 duration-200">
      <div class="flex items-start justify-between gap-4">
        <div class="space-y-1.5">
          <div class="flex items-center gap-2 flex-wrap">
            <span class="px-2.5 py-0.5 text-xs font-mono font-bold bg-ant-primary text-white rounded-md shadow-xs">
              v{updaterStore.latestRelease.version}
            </span>
            <span class="text-xs text-ant-text-secondary font-mono flex items-center gap-1">
              <span>(Current: v{updaterStore.currentVersion})</span>
              <ArrowRight size={12} class="text-ant-primary" />
              <span class="font-bold text-ant-primary">v{updaterStore.latestRelease.version}</span>
            </span>
            {#if updaterStore.latestRelease.publishedAt}
              <span class="text-[11px] text-ant-text-muted flex items-center gap-1">
                <Calendar size={11} /> {formatDate(updaterStore.latestRelease.publishedAt)}
              </span>
            {/if}
          </div>

          <h4 class="font-serif-display text-sm font-bold text-ant-text">
            {updaterStore.latestRelease.title || `AetherGrok v${updaterStore.latestRelease.version} Released`}
          </h4>

          {#if updaterStore.latestRelease.highlights && updaterStore.latestRelease.highlights.length > 0}
            <ul class="space-y-1 pt-1">
              {#each updaterStore.latestRelease.highlights.slice(0, 3) as item}
                <li class="text-xs text-ant-text-secondary flex items-start gap-2">
                  <span class="text-ant-primary font-bold leading-none mt-1">•</span>
                  <span>{item}</span>
                </li>
              {/each}
            </ul>
          {/if}
        </div>
      </div>

      <!-- Hero Banner Action Buttons -->
      <div class="pt-2 border-t border-ant-primary/20 flex items-center justify-between flex-wrap gap-3">
        <div class="flex items-center gap-2">
          {#if updaterStore.matchedAsset}
            <button
              type="button"
              onclick={() => updaterStore.openDownload(updaterStore.matchedAsset?.downloadUrl)}
              class="px-4 py-2 rounded-xl text-xs font-serif font-semibold bg-ant-primary hover:bg-ant-primary-hover text-white shadow-md shadow-ant-primary/20 transition flex items-center gap-2 cursor-pointer active:scale-95"
            >
              <DownloadCloud size={15} />
              <span>Download Installer ({updaterStore.matchedAsset.name})</span>
              {#if updaterStore.matchedAsset.size > 0}
                <span class="opacity-80 font-mono text-[10.5px]">({formatBytes(updaterStore.matchedAsset.size)})</span>
              {/if}
            </button>
          {:else}
            <button
              type="button"
              onclick={() => updaterStore.openDownload(`https://github.com/fiko942/grok-build/releases/tag/${updaterStore.latestRelease?.tagName}`)}
              class="px-4 py-2 rounded-xl text-xs font-serif font-semibold bg-ant-primary hover:bg-ant-primary-hover text-white shadow-md shadow-ant-primary/20 transition flex items-center gap-2 cursor-pointer active:scale-95"
            >
              <DownloadCloud size={15} />
              <span>Download Release Assets</span>
            </button>
          {/if}

          <button
            type="button"
            onclick={() => updaterStore.openDownload(`https://github.com/fiko942/grok-build/releases/tag/${updaterStore.latestRelease?.tagName}`)}
            class="px-3 py-2 rounded-xl text-xs font-serif text-ant-text hover:text-ant-primary bg-ant-bg border border-ant-border-secondary dark:border-white/5 hover:bg-ant-bg-tertiary transition flex items-center gap-1.5 cursor-pointer shadow-2xs"
          >
            <ExternalLink size={13} />
            <span>View on GitHub</span>
          </button>
        </div>

        <span class="text-[11px] text-ant-text-muted font-mono">
          Ad-hoc signed installer package
        </span>
      </div>
    </div>
  {/if}

  <!-- Version History & Changelog Timeline -->
  <div class="space-y-3">
    <div class="flex items-center justify-between">
      <div class="text-xs font-semibold text-ant-text flex items-center gap-1.5">
        <Tag size={13} class="text-ant-primary" />
        <span>Version History & Release Notes</span>
      </div>
      <span class="text-[11px] text-ant-text-muted font-mono">
        {updaterStore.allReleases.length} {updaterStore.allReleases.length === 1 ? 'release' : 'releases'} recorded
      </span>
    </div>

    {#if updaterStore.checking && updaterStore.allReleases.length === 0}
      <div class="p-8 rounded-xl bg-ant-bg border border-ant-border-secondary dark:border-white/5 flex flex-col items-center justify-center text-center space-y-2">
        <RotateCw size={24} class="text-ant-primary animate-spin" />
        <p class="text-xs font-serif text-ant-text-secondary">Fetching releases from GitHub...</p>
      </div>
    {:else if updaterStore.allReleases.length === 0}
      <div class="p-8 rounded-xl bg-ant-bg border border-ant-border-secondary dark:border-white/5 flex flex-col items-center justify-center text-center space-y-3">
        <Package size={28} class="text-ant-text-muted opacity-50" />
        <div class="space-y-1">
          <p class="text-xs font-serif font-medium text-ant-text">No Release History Found</p>
          <p class="text-[11px] font-serif text-ant-text-secondary max-w-sm">
            Check your network connection or verify that GitHub releases are published on the repository.
          </p>
        </div>
        <button
          type="button"
          onclick={() => updaterStore.fetchChangelogHistory()}
          class="px-3 py-1.5 rounded-lg text-xs font-serif bg-ant-bg-tertiary hover:bg-ant-primary hover:text-white text-ant-text border border-ant-border-secondary dark:border-white/5 transition flex items-center gap-1.5"
        >
          <RefreshCw size={12} />
          <span>Retry Loading Changelog</span>
        </button>
      </div>
    {:else}
      <div class="space-y-3">
        {#each updaterStore.allReleases as release (release.version)}
          {@const isExpanded = !!expandedReleases[release.version]}
          {@const isCurrent = release.version === updaterStore.currentVersion}
          {@const isLatest = release.isLatest || release.version === updaterStore.latestVersion}

          <div class="rounded-xl bg-ant-bg border {
            isCurrent
              ? 'border-ant-primary/40 shadow-xs'
              : 'border-ant-border-secondary dark:border-white/5'
          } overflow-hidden transition-all duration-150">
            <!-- Release Card Header Accordion Trigger -->
            <button
              type="button"
              onclick={() => toggleExpand(release.version)}
              class="w-full p-3.5 flex items-center justify-between text-left hover:bg-ant-bg-secondary/60 transition cursor-pointer"
            >
              <div class="flex items-center space-x-3 min-w-0">
                <!-- Version Pill -->
                <div class="flex items-center gap-2">
                  <span class="px-2 py-0.5 text-xs font-mono font-bold {
                    isCurrent
                      ? 'bg-ant-primary text-white'
                      : 'bg-ant-bg-secondary border border-ant-border-secondary dark:border-white/5 text-ant-text'
                  } rounded-md">
                    v{release.version}
                  </span>

                  {#if isCurrent}
                    <span class="px-1.5 py-0.2 text-[10px] font-mono font-semibold bg-emerald-500/15 text-emerald-600 dark:text-emerald-400 border border-emerald-500/30 rounded">
                      Current
                    </span>
                  {/if}

                  {#if isLatest && !isCurrent}
                    <span class="px-1.5 py-0.2 text-[10px] font-mono font-semibold bg-amber-500/15 text-amber-600 dark:text-amber-400 border border-amber-500/30 rounded">
                      Latest
                    </span>
                  {/if}
                </div>

                <!-- Release Title -->
                <span class="text-xs font-serif font-medium text-ant-text truncate">
                  {release.title || `Release v${release.version}`}
                </span>
              </div>

              <div class="flex items-center space-x-3 shrink-0">
                {#if release.publishedAt}
                  <span class="text-[11px] text-ant-text-muted font-serif flex items-center gap-1">
                    <Calendar size={11} /> {formatDate(release.publishedAt)}
                  </span>
                {/if}

                <div class="p-1 rounded text-ant-text-muted hover:text-ant-text">
                  {#if isExpanded}
                    <ChevronUp size={14} />
                  {:else}
                    <ChevronDown size={14} />
                  {/if}
                </div>
              </div>
            </button>

            <!-- Expanded Accordion Content -->
            {#if isExpanded}
              <div class="p-4 pt-2 border-t border-ant-border-secondary dark:border-white/5 space-y-4 animate-in fade-in duration-100 bg-ant-bg-secondary/30">
                
                <!-- Highlights List -->
                {#if release.highlights && release.highlights.length > 0}
                  <div class="space-y-1.5">
                    <div class="text-[11px] font-semibold text-ant-text uppercase tracking-wider">
                      Key Highlights & Features
                    </div>
                    <ul class="space-y-1">
                      {#each release.highlights as highlight}
                        <li class="text-xs text-ant-text-secondary flex items-start gap-2">
                          <span class="text-ant-primary font-bold leading-none mt-1">•</span>
                          <span>{highlight}</span>
                        </li>
                      {/each}
                    </ul>
                  </div>
                {/if}

                <!-- Formatted Release Body / Notes -->
                {#if release.body}
                  <div class="space-y-1.5">
                    <div class="text-[11px] font-semibold text-ant-text uppercase tracking-wider">
                      Release Notes
                    </div>
                    <div class="p-3 rounded-lg bg-ant-bg border border-ant-border-secondary dark:border-white/5 text-xs text-ant-text font-serif whitespace-pre-wrap leading-relaxed max-h-48 overflow-y-auto custom-scrollbar select-text">
                      {release.body}
                    </div>
                  </div>
                {/if}

                <!-- Platform Assets Downloads Table / Chips -->
                {#if release.assets && release.assets.length > 0}
                  <div class="space-y-2 pt-1 border-t border-ant-border-secondary dark:border-white/5">
                    <div class="text-[11px] font-semibold text-ant-text uppercase tracking-wider flex items-center gap-1.5">
                      <Download size={12} class="text-ant-primary" />
                      <span>Available Platform Binaries & Installers</span>
                    </div>

                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                      {#each release.assets as asset}
                        {@const isDmg = asset.name.endsWith('.dmg')}
                        {@const isExe = asset.name.endsWith('.exe')}
                        {@const isZip = asset.name.endsWith('.zip')}

                        <button
                          type="button"
                          onclick={() => updaterStore.openDownload(asset.downloadUrl)}
                          class="p-2.5 rounded-lg bg-ant-bg border border-ant-border-secondary dark:border-white/5 hover:border-ant-primary/40 hover:bg-ant-bg-tertiary transition flex items-center justify-between text-left group cursor-pointer shadow-2xs"
                        >
                          <div class="flex items-center space-x-2 min-w-0">
                            <div class="p-1.5 rounded-md bg-ant-primary/10 text-ant-primary group-hover:scale-105 transition-transform shrink-0">
                              {#if isDmg || isExe}
                                <Laptop size={13} />
                              {:else}
                                <Package size={13} />
                              {/if}
                            </div>
                            <div class="min-w-0">
                              <div class="text-xs font-mono font-medium text-ant-text truncate group-hover:text-ant-primary transition-colors">
                                {asset.name}
                              </div>
                              <div class="text-[10px] text-ant-text-muted font-mono">
                                {formatBytes(asset.size)}
                              </div>
                            </div>
                          </div>

                          <Download size={13} class="text-ant-text-muted group-hover:text-ant-primary shrink-0 ml-2" />
                        </button>
                      {/each}
                    </div>
                  </div>
                {/if}

                <!-- Direct GitHub Release Link -->
                <div class="pt-1 flex items-center justify-end">
                  <button
                    type="button"
                    onclick={() => updaterStore.openDownload(`https://github.com/fiko942/grok-build/releases/tag/${release.tagName}`)}
                    class="text-[11px] font-serif text-ant-primary hover:underline flex items-center gap-1 cursor-pointer"
                  >
                    <span>View v{release.version} Release Page on GitHub</span>
                    <ExternalLink size={10} />
                  </button>
                </div>
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
