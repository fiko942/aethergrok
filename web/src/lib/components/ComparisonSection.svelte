<script lang="ts">
  import { comparisonData } from '../data/comparison';
  import { ShieldCheck, Info } from 'lucide-svelte';
</script>

<section id="comparison" class="py-20 md:py-32 bg-slate-50/60 border-y border-slate-200/80">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="text-center max-w-3xl mx-auto mb-16">
      <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-slate-200/80 text-slate-800 mb-3">
        <Info class="w-3.5 h-3.5 text-slate-600" />
        <span>Evidence-Based Architectural Comparison</span>
      </div>
      <h2 class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
        Technical Comparison Matrix
      </h2>
      <p class="mt-4 text-base sm:text-lg text-slate-600">
        An objective overview of architectural choices, runtime constraints, and interaction models across developer tools.
      </p>
    </div>

    <!-- Responsive Comparison Table -->
    <div class="overflow-x-auto rounded-2xl border border-slate-200 bg-white shadow-sm">
      <table class="w-full text-left text-sm border-collapse min-w-[760px]">
        <thead>
          <tr class="bg-slate-100/80 border-b border-slate-200 text-slate-700">
            <th class="py-4 px-6 font-bold text-xs uppercase tracking-wider text-slate-600 w-1/4">
              Metric / Dimension
            </th>
            <th class="py-4 px-6 font-bold text-xs uppercase tracking-wider text-blue-700 bg-blue-50/50 w-1/4 border-x border-blue-100">
              AetherGrok Desktop
            </th>
            <th class="py-4 px-6 font-semibold text-xs uppercase tracking-wider text-slate-600 w-1/6">
              Anthropic Claude Code
            </th>
            <th class="py-4 px-6 font-semibold text-xs uppercase tracking-wider text-slate-600 w-1/6">
              OpenAI Codex CLI
            </th>
            <th class="py-4 px-6 font-semibold text-xs uppercase tracking-wider text-slate-600 w-1/6">
              Antigravity / CUA
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          {#each comparisonData as row}
            <tr class="hover:bg-slate-50/80 transition-colors {row.highlight ? 'bg-slate-50/30' : ''}">
              <td class="py-4 px-6 font-semibold text-slate-900 flex items-center gap-2">
                <span>{row.metric}</span>
              </td>
              <td class="py-4 px-6 font-semibold text-blue-900 bg-blue-50/30 border-x border-blue-100/60">
                {row.aethergrok}
              </td>
              <td class="py-4 px-6 text-slate-600">
                {row.claudeCode}
              </td>
              <td class="py-4 px-6 text-slate-600">
                {row.codexCli}
              </td>
              <td class="py-4 px-6 text-slate-600">
                {row.antigravityCua}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>

    <!-- Footnote -->
    <div class="mt-6 flex items-start gap-2.5 text-xs text-slate-500 max-w-3xl">
      <ShieldCheck class="w-4 h-4 text-slate-400 flex-shrink-0 mt-0.5" />
      <p>
        Note: Specifications are benchmarked against official documentation, local system metrics, and repository manifests. All trademarks belong to their respective owners.
      </p>
    </div>
  </div>
</section>
