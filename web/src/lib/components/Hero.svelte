<script lang="ts">
  import {
    Download,
    Github,
    Sparkles,
    Zap,
    Layers,
    ShieldCheck,
    CheckCircle2,
    ArrowRight
  } from 'lucide-svelte';
  import { releaseVersion, githubRepo } from '../data/downloads';

  const defaultMacUrl = `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-macOS-arm64.dmg`;
  const defaultWinUrl = `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-windows-amd64-setup.exe`;
</script>

<section class="relative pt-12 pb-20 md:pt-20 md:pb-32 overflow-hidden bg-gradient-to-b from-slate-50 via-white to-white">
  <!-- Subtle Grid Background -->
  <div class="absolute inset-0 bg-[linear-gradient(to_right,#f1f5f9_1px,transparent_1px),linear-gradient(to_bottom,#f1f5f9_1px,transparent_1px)] bg-[size:4rem_4rem] [mask-image:radial-gradient(ellipse_60%_50%_at_50%_0%,#000_70%,transparent_100%)] pointer-events-none"></div>

  <div class="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="text-center max-w-4xl mx-auto">
      <!-- Announcement Pill -->
      <div class="inline-flex items-center gap-2 px-3 py-1.5 rounded-full text-xs sm:text-sm font-medium bg-blue-50 text-blue-800 border border-blue-200/80 mb-8 shadow-sm">
        <Sparkles class="w-4 h-4 text-blue-600" />
        <span>AetherGrok v{releaseVersion} is live with Full Repo Auto-Updates</span>
        <a href="#downloads" class="font-semibold text-blue-600 hover:text-blue-700 flex items-center gap-0.5 ml-1">
          Get it <ArrowRight class="w-3.5 h-3.5" />
        </a>
      </div>

      <!-- Main Headline -->
      <h1 class="text-4xl sm:text-6xl lg:text-7xl font-extrabold tracking-tight text-slate-900 leading-[1.1] mb-6">
        The High-Performance <br class="hidden sm:inline" />
        <span class="bg-gradient-to-r from-blue-600 via-indigo-600 to-blue-700 bg-clip-text text-transparent">
          Desktop Studio
        </span>
        for Grok Build.
      </h1>

      <!-- Subtitle -->
      <p class="text-lg sm:text-xl text-slate-600 max-w-2xl mx-auto mb-10 leading-relaxed">
        Autonomous coding without terminal friction or memory bloat. Engineered with Go 1.24 and Svelte 5, delivering 10-turn DOM windowing, live mid-turn steering, visual code reviews, and instant vision snapshots.
      </p>

      <!-- CTA Buttons -->
      <div class="flex flex-col sm:flex-row items-center justify-center gap-4 mb-10">
        <a
          href="{defaultMacUrl}"
          class="w-full sm:w-auto inline-flex items-center justify-center gap-2.5 px-6 py-3.5 text-base font-semibold text-white bg-blue-600 hover:bg-blue-700 active:scale-[0.99] rounded-xl shadow-md hover:shadow-lg transition-all duration-150 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
        >
          <Download class="w-5 h-5" />
          <span>Download for macOS (Apple Silicon)</span>
        </a>

        <a
          href="{defaultWinUrl}"
          class="w-full sm:w-auto inline-flex items-center justify-center gap-2.5 px-6 py-3.5 text-base font-semibold text-slate-700 bg-white hover:bg-slate-50 border border-slate-300 rounded-xl shadow-sm hover:shadow transition-all duration-150 focus:ring-2 focus:ring-slate-400 focus:ring-offset-2"
        >
          <Download class="w-5 h-5 text-slate-500" />
          <span>Windows Setup (.exe)</span>
        </a>

        <a
          href="{githubRepo}"
          target="_blank"
          rel="noopener noreferrer"
          class="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-5 py-3.5 text-base font-semibold text-slate-600 hover:text-slate-900 bg-slate-100 hover:bg-slate-200/80 rounded-xl transition-colors"
        >
          <Github class="w-5 h-5" />
          <span>GitHub</span>
        </a>
      </div>

      <!-- Trust & Spec Badges -->
      <div class="flex flex-wrap items-center justify-center gap-y-3 gap-x-6 text-xs sm:text-sm text-slate-500">
        <div class="flex items-center gap-1.5">
          <CheckCircle2 class="w-4 h-4 text-emerald-600" />
          <span>MIT Licensed & Open Source</span>
        </div>
        <div class="flex items-center gap-1.5">
          <Zap class="w-4 h-4 text-amber-500" />
          <span>~35 MB Native RAM Footprint</span>
        </div>
        <div class="flex items-center gap-1.5">
          <Layers class="w-4 h-4 text-blue-500" />
          <span>10-Turn Windowing (Zero Lag)</span>
        </div>
        <div class="flex items-center gap-1.5">
          <ShieldCheck class="w-4 h-4 text-indigo-500" />
          <span>Signed macOS & Windows Binaries</span>
        </div>
      </div>
    </div>

    <!-- Product Desktop Screenshot Mockup -->
    <div class="mt-16 sm:mt-20 max-w-5xl mx-auto">
      <div class="relative rounded-2xl p-2 bg-gradient-to-b from-slate-200/80 to-slate-300/40 shadow-2xl ring-1 ring-slate-900/10">
        <!-- Window Chrome / Titlebar -->
        <div class="h-9 bg-slate-100 rounded-t-xl px-4 flex items-center justify-between border-b border-slate-200">
          <div class="flex items-center gap-2">
            <div class="w-3 h-3 rounded-full bg-red-400"></div>
            <div class="w-3 h-3 rounded-full bg-amber-400"></div>
            <div class="w-3 h-3 rounded-full bg-emerald-400"></div>
          </div>
          <span class="text-xs font-mono text-slate-500 font-medium tracking-tight">
            AetherGrok Desktop — Autonomous Engineering Studio
          </span>
          <div class="w-10"></div>
        </div>

        <!-- App Screenshot -->
        <div class="overflow-hidden rounded-b-xl bg-slate-900">
          <img
            src="./screenshots/desktop-hero.webp"
            alt="AetherGrok Desktop Application Workspace"
            class="w-full h-auto object-cover block shadow-inner"
            loading="eager"
          />
        </div>
      </div>
    </div>
  </div>
</section>
