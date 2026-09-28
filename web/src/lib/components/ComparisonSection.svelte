<script lang="ts">
  import { comparisonData, comparisonCategories } from '../data/comparison';
  import { ShieldCheck, Info, Sparkles, Check, Layers, Cpu, Eye, Terminal } from 'lucide-svelte';

  let activeCategory = $state<string>('Semua');

  const filteredRows = $derived(
    activeCategory === 'Semua'
      ? comparisonData
      : comparisonData.filter((r) => r.category === activeCategory)
  );
</script>

<section id="comparison" class="py-20 md:py-32 bg-slate-50/70 border-y border-slate-200/80">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="text-center max-w-3xl mx-auto mb-12">
      <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-blue-50 text-blue-800 border border-blue-200 mb-3">
        <Sparkles class="w-3.5 h-3.5 text-blue-600" />
        <span>Komparasi Arsitektural Berbasis Fakta & Data Nyata</span>
      </div>
      <h2 class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
        Tabel Komparasi Fitur & Performa
      </h2>
      <p class="mt-4 text-base sm:text-lg text-slate-600">
        Perbandingan mendalam antara AetherGrok Desktop, Anthropic Claude Code, OpenAI Codex CLI, dan Antigravity / CUA.
      </p>
    </div>

    <!-- Category Filter Pills -->
    <div class="flex items-center justify-center gap-2 overflow-x-auto pb-4 mb-8">
      <button
        onclick={() => (activeCategory = 'Semua')}
        class="px-4 py-2 rounded-xl text-xs sm:text-sm font-semibold transition-all {activeCategory === 'Semua'
          ? 'bg-blue-600 text-white shadow-sm'
          : 'bg-white text-slate-600 hover:text-slate-900 border border-slate-200'}"
      >
        Semua Dimensi ({comparisonData.length})
      </button>
      {#each comparisonCategories as cat}
        <button
          onclick={() => (activeCategory = cat)}
          class="px-4 py-2 rounded-xl text-xs sm:text-sm font-semibold transition-all whitespace-nowrap {activeCategory === cat
            ? 'bg-blue-600 text-white shadow-sm'
            : 'bg-white text-slate-600 hover:text-slate-900 border border-slate-200'}"
        >
          {cat}
        </button>
      {/each}
    </div>

    <!-- Redesigned High-Contrast Comparison Table -->
    <div class="overflow-x-auto rounded-2xl border-2 border-slate-200/90 bg-white shadow-lg">
      <table class="w-full text-left text-sm border-collapse min-w-[840px]">
        <thead>
          <tr class="border-b-2 border-slate-200 bg-slate-100/90 text-slate-800">
            <th class="py-4 px-6 font-bold text-xs uppercase tracking-wider text-slate-600 w-1/4">
              Dimensi / Spesifikasi
            </th>
            <th class="py-4 px-6 font-extrabold text-xs uppercase tracking-wider text-blue-900 bg-blue-100/60 w-[30%] border-x-2 border-blue-200 shadow-inner">
              <div class="flex items-center gap-2">
                <span class="inline-block w-2.5 h-2.5 rounded-full bg-blue-600"></span>
                <span>AetherGrok Desktop</span>
                <span class="px-2 py-0.5 rounded text-[10px] font-bold bg-blue-600 text-white uppercase tracking-wider">
                  Unggulan
                </span>
              </div>
            </th>
            <th class="py-4 px-5 font-bold text-xs uppercase tracking-wider text-slate-700 w-[15%]">
              Anthropic Claude Code
            </th>
            <th class="py-4 px-5 font-bold text-xs uppercase tracking-wider text-slate-700 w-[15%]">
              OpenAI Codex CLI
            </th>
            <th class="py-4 px-5 font-bold text-xs uppercase tracking-wider text-slate-700 w-[15%]">
              Antigravity / CUA
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          {#each filteredRows as row}
            <tr class="hover:bg-slate-50/90 transition-colors">
              <!-- Metric Column -->
              <td class="py-4 px-6 font-semibold text-slate-900">
                <div class="text-sm font-bold text-slate-900">{row.metric}</div>
                <div class="text-[11px] font-mono font-medium text-slate-600 mt-0.5">{row.category}</div>
              </td>

              <!-- AetherGrok Column (High-Contrast Highlighted) -->
              <td class="py-4 px-6 bg-blue-50/50 border-x-2 border-blue-200">
                <div class="flex items-center gap-1.5 text-blue-950 font-bold text-sm">
                  <Check class="w-4 h-4 text-blue-600 flex-shrink-0" />
                  <span>{row.aethergrok.title}</span>
                </div>
                <p class="text-xs text-blue-800 font-medium mt-1 leading-relaxed">
                  {row.aethergrok.sub}
                </p>
              </td>

              <!-- Claude Code Column -->
              <td class="py-4 px-5">
                <div class="text-slate-900 font-semibold text-xs">{row.claudeCode.title}</div>
                <p class="text-[11px] text-slate-700 mt-0.5 leading-snug">{row.claudeCode.sub}</p>
              </td>

              <!-- Codex CLI Column -->
              <td class="py-4 px-5">
                <div class="text-slate-900 font-semibold text-xs">{row.codexCli.title}</div>
                <p class="text-[11px] text-slate-700 mt-0.5 leading-snug">{row.codexCli.sub}</p>
              </td>

              <!-- Antigravity / CUA Column -->
              <td class="py-4 px-5">
                <div class="text-slate-900 font-semibold text-xs">{row.antigravityCua.title}</div>
                <p class="text-[11px] text-slate-700 mt-0.5 leading-snug">{row.antigravityCua.sub}</p>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Evidence Note -->
    <div class="mt-6 flex items-start gap-2.5 text-xs text-slate-600 max-w-3xl">
      <ShieldCheck class="w-4 h-4 text-slate-500 flex-shrink-0 mt-0.5" />
      <p>
        Data spesifikasi diukur berdasarkan pengujian lokal di lingkungan produksi macOS & Windows, benchmarking runtime, serta dokumentasi arsitektural resmi dari masing-masing alat.
      </p>
    </div>
  </div>
</section>
