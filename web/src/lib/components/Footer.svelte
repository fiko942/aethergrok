<script lang="ts">
  import { githubRepo, fetchLiveLatestRelease, type LatestReleaseInfo } from '../data/downloads';
  import { Github, Heart, Shield, Terminal, Coffee } from 'lucide-svelte';
  import { onMount } from 'svelte';

  let releaseInfo = $state<LatestReleaseInfo | null>(null);

  onMount(async () => {
    releaseInfo = await fetchLiveLatestRelease();
  });

  const displayVersion = $derived(releaseInfo?.tagName || 'v1.0.3');
</script>

<footer class="bg-slate-900 text-slate-400 py-12 sm:py-16 border-t border-slate-800">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="grid grid-cols-1 md:grid-cols-4 gap-8 mb-12">
      <!-- Col 1: Brand & Tagline -->
      <div class="md:col-span-2 space-y-4">
        <div class="flex items-center gap-2.5">
          <img
            src="./app-icon.png"
            alt="AetherGrok Logo"
            class="w-8 h-8 rounded-lg shadow-sm object-contain"
          />
          <span class="text-xl font-bold text-white tracking-tight">
            AetherGrok Desktop
          </span>
        </div>
        <p class="text-sm text-slate-400 max-w-sm leading-relaxed">
          High-performance desktop studio for Grok Build CLI. Engineered with Go 1.24 and Svelte 5 for autonomous software engineering.
        </p>
        <div class="text-xs text-slate-500">
          Licensed under the <a href="{githubRepo}/blob/main/LICENSE" target="_blank" class="text-slate-400 hover:text-white underline">MIT License</a>.
        </div>
      </div>

      <!-- Col 2: Navigation -->
      <div class="space-y-3">
        <h4 class="text-xs font-bold uppercase tracking-wider text-slate-200">
          Product & Tech
        </h4>
        <ul class="space-y-2 text-sm">
          <li><a href="#features" class="hover:text-white transition-colors">10-Turn Windowing</a></li>
          <li><a href="#features" class="hover:text-white transition-colors">Vision Screen Capture</a></li>
          <li><a href="#features" class="hover:text-white transition-colors">Voice Push-to-Talk</a></li>
          <li><a href="#comparison" class="hover:text-white transition-colors">Comparison Matrix</a></li>
          <li><a href="#downloads" class="hover:text-white transition-colors">Latest Release ({displayVersion})</a></li>
        </ul>
      </div>

      <!-- Col 3: Resources & Links -->
      <div class="space-y-3">
        <h4 class="text-xs font-bold uppercase tracking-wider text-slate-200">
          Community & Code
        </h4>
        <ul class="space-y-2 text-sm">
          <li>
            <a
              href="{githubRepo}"
              target="_blank"
              rel="noopener noreferrer"
              class="hover:text-white flex items-center gap-1.5 transition-colors"
            >
              <Github class="w-4 h-4" />
              <span>GitHub Repository</span>
            </a>
          </li>
          <li>
            <a
              href="{githubRepo}/releases/latest"
              target="_blank"
              rel="noopener noreferrer"
              class="hover:text-white transition-colors"
            >
              Latest Release ({displayVersion})
            </a>
          </li>
          <li>
            <a
              href="https://saweria.co/wijifikoteren"
              target="_blank"
              rel="noopener noreferrer"
              class="hover:text-amber-400 flex items-center gap-1 text-amber-400 font-medium transition-colors"
            >
              <Coffee class="w-4 h-4" />
              <span>Donasi Saweria</span>
            </a>
          </li>
          <li>
            <a
              href="{githubRepo}/issues"
              target="_blank"
              rel="noopener noreferrer"
              class="hover:text-white transition-colors"
            >
              Issue Tracker
            </a>
          </li>
        </ul>
      </div>
    </div>

    <!-- Bottom Bar -->
    <div class="pt-8 border-t border-slate-800/80 flex flex-col sm:flex-row items-center justify-between text-xs text-slate-500 gap-4">
      <p>© 2026 AetherGrok Contributors. Not affiliated with or endorsed by SpaceXAI.</p>
      <div class="flex items-center gap-4">
        <span>macOS & Windows Desktop</span>
        <span>•</span>
        <span>Pure Svelte 5 + Go 1.24</span>
      </div>
    </div>
  </div>
</footer>
