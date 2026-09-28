<script lang="ts">
  import {
    Activity,
    Eye,
    Cpu,
    Sparkles,
    Check
  } from 'lucide-svelte';
</script>

<section class="py-16 md:py-24 bg-slate-50/50 border-y border-slate-200/80">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="text-center max-w-3xl mx-auto mb-16">
      <h2 class="text-xs font-bold uppercase tracking-widest text-blue-600 mb-2">
        Architectural Precision
      </h2>
      <p class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
        Engineered for Sustained Developer Velocity
      </p>
      <p class="mt-4 text-base sm:text-lg text-slate-600">
        AI coding tools typically degrade as sessions grow long. AetherGrok was architected from the ground up to solve memory pressure and visual friction.
      </p>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
      <!-- Card 1: Performance -->
      <div class="bg-white rounded-2xl p-8 border border-slate-200 shadow-sm hover:shadow-md transition-shadow">
        <div class="w-12 h-12 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center mb-6">
          <Activity class="w-6 h-6" />
        </div>
        <h3 class="text-xl font-bold text-slate-900 mb-3">
          10-Turn DOM Windowing
        </h3>
        <p class="text-slate-600 text-sm leading-relaxed mb-6">
          Standard webviews slow down significantly after 20+ turns with dozens of diff blocks. AetherGrok dynamically virtualizes the DOM tree, keeping only 10 turns active while preserving scroll locks.
        </p>
        <ul class="space-y-2 text-xs text-slate-500">
          <li class="flex items-center gap-2">
            <Check class="w-4 h-4 text-emerald-600 flex-shrink-0" />
            <span>Constant ~35 MB baseline RAM usage</span>
          </li>
          <li class="flex items-center gap-2">
            <Check class="w-4 h-4 text-emerald-600 flex-shrink-0" />
            <span>Smooth 60 FPS viewport scrolling</span>
          </li>
        </ul>
      </div>

      <!-- Card 2: Visual Ergonomics -->
      <div class="bg-white rounded-2xl p-8 border border-slate-200 shadow-sm hover:shadow-md transition-shadow">
        <div class="w-12 h-12 rounded-xl bg-indigo-50 text-indigo-600 flex items-center justify-center mb-6">
          <Eye class="w-6 h-6" />
        </div>
        <h3 class="text-xl font-bold text-slate-900 mb-3">
          Rich Visual Ergonomics
        </h3>
        <p class="text-slate-600 text-sm leading-relaxed mb-6">
          CLI terminals hide context inside scrolling ANSI text. AetherGrok provides side-by-side syntax-highlighted diffs, push-to-talk voice equalizer, and non-intrusive full-screen snapshot capture.
        </p>
        <ul class="space-y-2 text-xs text-slate-500">
          <li class="flex items-center gap-2">
            <Check class="w-4 h-4 text-emerald-600 flex-shrink-0" />
            <span>Compositor-synced OS screen grab</span>
          </li>
          <li class="flex items-center gap-2">
            <Check class="w-4 h-4 text-emerald-600 flex-shrink-0" />
            <span>Auto audio ducking during voice speech</span>
          </li>
        </ul>
      </div>

      <!-- Card 3: Process Reliability -->
      <div class="bg-white rounded-2xl p-8 border border-slate-200 shadow-sm hover:shadow-md transition-shadow">
        <div class="w-12 h-12 rounded-xl bg-emerald-50 text-emerald-600 flex items-center justify-center mb-6">
          <Cpu class="w-6 h-6" />
        </div>
        <h3 class="text-xl font-bold text-slate-900 mb-3">
          Process Group Supervision
        </h3>
        <p class="text-slate-600 text-sm leading-relaxed mb-6">
          Subprocess runners frequently orphan background tasks when cancelled. AetherGrok binds CLI workers with POSIX process groups (<code class="bg-slate-100 px-1 py-0.5 rounded text-[11px] text-slate-700">setpgid</code>) and Windows Job Objects.
        </p>
        <ul class="space-y-2 text-xs text-slate-500">
          <li class="flex items-center gap-2">
            <Check class="w-4 h-4 text-emerald-600 flex-shrink-0" />
            <span>Zero orphaned background processes</span>
          </li>
          <li class="flex items-center gap-2">
            <Check class="w-4 h-4 text-emerald-600 flex-shrink-0" />
            <span>Mid-turn live prompt steering</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</section>
