<script lang="ts">
  import {
    Download,
    Github,
    Sparkles,
    Menu,
    X,
    ExternalLink,
    Coffee
  } from 'lucide-svelte';
  import { githubRepo, fetchLiveLatestRelease, type LatestReleaseInfo } from '../data/downloads';
  import { onMount } from 'svelte';

  let releaseInfo = $state<LatestReleaseInfo | null>(null);
  let mobileMenuOpen = $state(false);

  onMount(async () => {
    releaseInfo = await fetchLiveLatestRelease();
  });

  const displayVersion = $derived(releaseInfo?.tagName || 'v1.0.3');

  function toggleMobileMenu() {
    mobileMenuOpen = !mobileMenuOpen;
  }

  function closeMobileMenu() {
    mobileMenuOpen = false;
  }
</script>

<header class="sticky top-0 z-50 glass-nav border-b border-slate-200/80 transition-all duration-200">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 h-16 flex items-center justify-between">
    <!-- Brand / Logo -->
    <a href="/" class="flex items-center gap-3 group focus:outline-none focus:ring-2 focus:ring-blue-500 rounded-lg p-1">
      <img
        src="./app-icon.png"
        alt="AetherGrok Logo"
        class="w-9 h-9 rounded-xl shadow-subtle group-hover:scale-105 transition-transform duration-200 object-contain"
      />
      <div class="flex items-center gap-2">
        <span class="text-xl font-bold tracking-tight text-slate-900 group-hover:text-blue-600 transition-colors">
          AetherGrok
        </span>
        <span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-semibold bg-blue-50 text-blue-700 border border-blue-200">
          {displayVersion}
        </span>
      </div>
    </a>

    <!-- Desktop Nav Links -->
    <nav class="hidden md:flex items-center gap-8">
      <a href="#features" class="text-sm font-medium text-slate-600 hover:text-slate-900 transition-colors">
        Features
      </a>
      <a href="#comparison" class="text-sm font-medium text-slate-600 hover:text-slate-900 transition-colors">
        Comparison Matrix
      </a>
      <a href="#downloads" class="text-sm font-medium text-slate-600 hover:text-slate-900 transition-colors">
        Downloads
      </a>
      <a href="#sponsors" class="text-sm font-medium text-slate-600 hover:text-slate-900 flex items-center gap-1.5 transition-colors">
        <Coffee class="w-4 h-4 text-amber-500" />
        <span>Donasi (Saweria)</span>
      </a>
      <a
        href="{githubRepo}"
        target="_blank"
        rel="noopener noreferrer"
        class="text-sm font-medium text-slate-600 hover:text-slate-900 flex items-center gap-1.5 transition-colors"
      >
        <Github class="w-4 h-4" />
        <span>GitHub</span>
        <ExternalLink class="w-3 h-3 text-slate-400" />
      </a>
    </nav>

    <!-- Right Actions -->
    <div class="hidden md:flex items-center gap-3">
      <a
        href="#downloads"
        class="inline-flex items-center gap-2 px-4 py-2 text-sm font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm hover:shadow transition-all duration-150 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
      >
        <Download class="w-4 h-4" />
        <span>Download {displayVersion}</span>
      </a>
    </div>

    <!-- Mobile Menu Button -->
    <div class="flex md:hidden">
      <button
        onclick={toggleMobileMenu}
        aria-label="Toggle menu"
        class="p-2 text-slate-600 hover:text-slate-900 hover:bg-slate-100 rounded-lg transition-colors"
      >
        {#if mobileMenuOpen}
          <X class="w-6 h-6" />
        {:else}
          <Menu class="w-6 h-6" />
        {/if}
      </button>
    </div>
  </div>

  <!-- Mobile Dropdown Menu -->
  {#if mobileMenuOpen}
    <div class="md:hidden border-b border-slate-200 bg-white px-4 pt-3 pb-5 space-y-3 shadow-lg">
      <a
        href="#features"
        onclick={closeMobileMenu}
        class="block px-3 py-2 rounded-md text-base font-medium text-slate-700 hover:text-blue-600 hover:bg-slate-50"
      >
        Features
      </a>
      <a
        href="#comparison"
        onclick={closeMobileMenu}
        class="block px-3 py-2 rounded-md text-base font-medium text-slate-700 hover:text-blue-600 hover:bg-slate-50"
      >
        Comparison Matrix
      </a>
      <a
        href="#downloads"
        onclick={closeMobileMenu}
        class="block px-3 py-2 rounded-md text-base font-medium text-slate-700 hover:text-blue-600 hover:bg-slate-50"
      >
        Downloads ({displayVersion})
      </a>
      <a
        href="#sponsors"
        onclick={closeMobileMenu}
        class="block px-3 py-2 rounded-md text-base font-medium text-amber-700 hover:text-amber-800 hover:bg-amber-50"
      >
        Donasi Saweria
      </a>
      <a
        href="{githubRepo}"
        target="_blank"
        rel="noopener noreferrer"
        class="flex items-center gap-2 px-3 py-2 rounded-md text-base font-medium text-slate-700 hover:text-blue-600 hover:bg-slate-50"
      >
        <Github class="w-5 h-5" />
        <span>GitHub Repository</span>
      </a>
      <div class="pt-2">
        <a
          href="#downloads"
          onclick={closeMobileMenu}
          class="w-full flex items-center justify-center gap-2 px-4 py-2.5 text-base font-semibold text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm"
        >
          <Download class="w-5 h-5" />
          <span>Download App ({displayVersion})</span>
        </a>
      </div>
    </div>
  {/if}
</header>
