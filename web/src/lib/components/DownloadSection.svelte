<script lang="ts">
  import {
    downloadsData,
    releaseVersion,
    releaseDate,
    githubRepo
  } from '../data/downloads';
  import {
    Download,
    Check,
    Copy,
    Apple,
    Terminal,
    FileCode,
    ExternalLink
  } from 'lucide-svelte';

  let activePlatform = $state<'macOS' | 'Windows'>('macOS');
  let copiedCommand = $state(false);

  const filteredDownloads = $derived(
    downloadsData.filter((d) => d.platform === activePlatform)
  );

  const installCurl = `curl -fsSL https://raw.githubusercontent.com/fiko942/aethergrok/main/scripts/install.sh | bash`;

  function copyInstallScript() {
    navigator.clipboard.writeText(installCurl);
    copiedCommand = true;
    setTimeout(() => {
      copiedCommand = false;
    }, 2000);
  }
</script>

<section id="downloads" class="py-20 md:py-32 bg-white">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="text-center max-w-3xl mx-auto mb-16">
      <h2 class="text-xs font-bold uppercase tracking-widest text-blue-600 mb-2">
        Production Distribution
      </h2>
      <p class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
        Download AetherGrok v{releaseVersion}
      </p>
      <p class="mt-4 text-base sm:text-lg text-slate-600">
        Free, self-contained, ad-hoc codesigned packages built natively for macOS and Windows.
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
      {#each filteredDownloads as pkg}
        <div class="relative bg-slate-50/70 border border-slate-200 rounded-2xl p-6 sm:p-7 flex flex-col justify-between hover:border-blue-300 hover:shadow-md transition-all">
          {#if pkg.recommended}
            <div class="absolute -top-3 right-6 px-3 py-0.5 rounded-full text-[11px] font-bold uppercase tracking-wider bg-blue-600 text-white shadow-sm">
              Recommended
            </div>
          {/if}

          <div>
            <div class="flex items-center justify-between mb-3">
              <span class="text-xs font-bold uppercase tracking-wider text-slate-500">
                {pkg.type}
              </span>
              <span class="text-xs font-mono font-medium text-slate-400 bg-white px-2 py-0.5 rounded border border-slate-200">
                {pkg.size}
              </span>
            </div>

            <h3 class="text-lg font-bold text-slate-900 mb-1">
              {pkg.arch}
            </h3>

            <p class="text-xs font-mono text-slate-500 mb-6 truncate">
              {pkg.filename}
            </p>
          </div>

          <div class="space-y-2">
            <a
              href="{pkg.url}"
              class="w-full inline-flex items-center justify-center gap-2 px-5 py-3 rounded-xl font-bold text-sm text-white bg-blue-600 hover:bg-blue-700 shadow-sm hover:shadow active:scale-[0.99] transition-all"
            >
              <Download class="w-4 h-4" />
              <span>Download Package</span>
            </a>

            {#if pkg.checksumUrl}
              <div class="text-center">
                <a
                  href="{pkg.checksumUrl}"
                  target="_blank"
                  class="text-[11px] font-mono text-slate-400 hover:text-blue-600 underline transition-colors"
                >
                  Verify SHA256 Checksum
                </a>
              </div>
            {/if}
          </div>
        </div>
      {/each}
    </div>

    <!-- Quick Terminal One-Liner Box -->
    <div class="max-w-2xl mx-auto bg-slate-900 rounded-2xl p-4 sm:p-5 text-slate-200 shadow-lg border border-slate-800">
      <div class="flex items-center justify-between mb-2">
        <span class="text-xs font-mono text-slate-400 font-semibold flex items-center gap-1.5">
          <Terminal class="w-3.5 h-3.5 text-blue-400" />
          <span>Quick Install via Terminal (macOS / Linux)</span>
        </span>
        <button
          onclick={copyInstallScript}
          class="flex items-center gap-1 text-xs text-slate-300 hover:text-white px-2 py-1 rounded bg-slate-800 hover:bg-slate-700 transition-colors"
        >
          {#if copiedCommand}
            <Check class="w-3.5 h-3.5 text-emerald-400" />
            <span class="text-emerald-400">Copied</span>
          {:else}
            <Copy class="w-3.5 h-3.5" />
            <span>Copy</span>
          {/if}
        </button>
      </div>
      <div class="font-mono text-xs sm:text-sm text-slate-300 select-all overflow-x-auto py-1">
        {installCurl}
      </div>
    </div>
  </div>
</section>
