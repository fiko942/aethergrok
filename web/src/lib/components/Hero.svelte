<script lang="ts">
  import {
    Download,
    Github,
    Sparkles,
    Zap,
    Layers,
    ShieldCheck,
    CheckCircle2,
    ArrowRight,
    Terminal,
    Code,
    Cpu,
    Play,
    Check
  } from 'lucide-svelte';
  import { releaseVersion, githubRepo } from '../data/downloads';

  const defaultMacUrl = `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-macOS-arm64.dmg`;
  const defaultWinUrl = `${githubRepo}/releases/download/v${releaseVersion}/AetherGrok-${releaseVersion}-windows-amd64-setup.exe`;
</script>

<section class="relative pt-12 pb-20 md:pt-20 md:pb-28 overflow-hidden bg-gradient-to-b from-slate-50 via-white to-white">
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
        Autonomous software engineering without terminal friction or memory bloat. Engineered with Go 1.24 and Svelte 5, delivering 10-turn DOM windowing, live mid-turn steering, visual code reviews, and instant vision snapshots.
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

    <!-- Clean Interactive Code & Architecture Studio Preview (No Screenshot) -->
    <div class="mt-16 sm:mt-20 max-w-4xl mx-auto">
      <div class="rounded-2xl border border-slate-200 bg-white shadow-xl overflow-hidden">
        <!-- Studio Chrome / Title Bar -->
        <div class="bg-slate-50 px-4 py-3 border-b border-slate-200 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <div class="w-3 h-3 rounded-full bg-slate-300"></div>
            <div class="w-3 h-3 rounded-full bg-slate-300"></div>
            <div class="w-3 h-3 rounded-full bg-slate-300"></div>
            <span class="text-xs font-mono font-medium text-slate-500 ml-2">
              AetherGrok Core Studio • Wails v2 / Go 1.24 Runtime
            </span>
          </div>

          <div class="flex items-center gap-2">
            <span class="inline-flex items-center gap-1 text-[11px] font-mono font-semibold px-2 py-0.5 rounded bg-emerald-50 text-emerald-700 border border-emerald-200">
              <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
              Virtual DOM: 10 Active Turns
            </span>
          </div>
        </div>

        <!-- Terminal & Code Stream Preview -->
        <div class="p-6 font-mono text-xs sm:text-sm bg-slate-900 text-slate-200 leading-relaxed overflow-x-auto space-y-3">
          <div class="flex items-center gap-2 text-slate-400 border-b border-slate-800 pb-2">
            <Terminal class="w-4 h-4 text-blue-400" />
            <span>Process Group Supervisor: [pgid: 48921 | status: running | memory: 34.2 MB]</span>
          </div>

          <div class="space-y-1.5">
            <div class="text-slate-400">
              <span class="text-emerald-400 font-bold">&gt; grok</span> run --stream ndjson --batch 16ms
            </div>
            <div class="text-slate-300">
              <span class="text-blue-400 font-bold">[Turn 42]</span> Analyzing codebase AST across 18 worktree files...
            </div>
            <div class="text-slate-400 pl-4 border-l border-slate-700 space-y-1 text-xs">
              <div class="text-emerald-300 flex items-center gap-1.5">
                <Check class="w-3.5 h-3.5" /> Checked pkg/grokrunner/stream_parser.go (0 errors)
              </div>
              <div class="text-emerald-300 flex items-center gap-1.5">
                <Check class="w-3.5 h-3.5" /> Virtualized DOM turns [1..32] pruned from layout tree (RAM locked)
              </div>
              <div class="text-amber-300 flex items-center gap-1.5">
                <Cpu class="w-3.5 h-3.5" /> Compositor snapshot synced: [Display: 2560x1440 @ 120Hz | latency: 48ms]
              </div>
            </div>
            <div class="text-slate-200 pt-1 flex items-center gap-2">
              <span class="inline-block w-2 h-4 bg-blue-500 animate-pulse"></span>
              <span class="text-slate-400 text-xs font-sans italic">Awaiting prompt steering or background worker completion...</span>
            </div>
          </div>
        </div>

        <!-- Footer Stats Bar -->
        <div class="bg-slate-50 px-6 py-3 border-t border-slate-200 flex flex-wrap items-center justify-between text-xs text-slate-600 gap-4">
          <div class="flex items-center gap-4">
            <span><strong>Runtime:</strong> Go 1.24 + Svelte 5</span>
            <span>•</span>
            <span><strong>IPC Latency:</strong> &lt; 1ms</span>
          </div>
          <div class="flex items-center gap-2 font-mono text-[11px] text-slate-500">
            <span>setpgid: active</span>
            <span>•</span>
            <span>ducking: enabled</span>
            <span>•</span>
            <span>diff-engine: diff2html</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
