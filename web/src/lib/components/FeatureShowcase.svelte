<script lang="ts">
  import { featuresData } from '../data/features';
  import {
    Activity,
    Layers,
    Eye,
    Mic,
    FileCode,
    Sparkles,
    CheckCircle2,
    Cpu,
    ArrowUpRight
  } from 'lucide-svelte';

  const iconMap: Record<string, any> = {
    windowing: Layers,
    'process-group': Cpu,
    vision: Eye,
    voice: Mic,
    diffs: FileCode,
    toolchain: Sparkles,
  };
</script>

<section id="features" class="py-20 md:py-32 bg-white">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="text-center max-w-3xl mx-auto mb-16">
      <h2 class="text-xs font-bold uppercase tracking-widest text-blue-600 mb-2">
        Core Capabilities
      </h2>
      <p class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
        Engineered for Sustained Developer Velocity
      </p>
      <p class="mt-4 text-base sm:text-lg text-slate-600">
        AetherGrok replaces terminal friction with high-performance desktop engineering. Every component is optimized for speed, precision, and reliability.
      </p>
    </div>

    <!-- 6-Card Feature Grid (Pure Code, Typography, & Architecture) -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
      {#each featuresData as feature}
        {@const IconComponent = iconMap[feature.id] || Layers}
        <div class="bg-slate-50/60 border border-slate-200/90 rounded-2xl p-7 flex flex-col justify-between hover:border-blue-300 hover:shadow-md hover:bg-white transition-all duration-200 group">
          <div>
            <div class="flex items-center justify-between mb-5">
              <div class="w-12 h-12 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center group-hover:scale-105 group-hover:bg-blue-600 group-hover:text-white transition-all duration-200">
                <IconComponent class="w-6 h-6" />
              </div>
              <span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-semibold bg-white text-slate-700 border border-slate-200 shadow-2xs">
                {feature.category}
              </span>
            </div>

            <h3 class="text-xl font-bold text-slate-900 mb-2 group-hover:text-blue-600 transition-colors">
              {feature.title}
            </h3>

            <p class="text-slate-600 text-sm leading-relaxed mb-6">
              {feature.description}
            </p>

            <div class="space-y-2.5 border-t border-slate-200/80 pt-5 mb-6">
              {#each feature.bullets as bullet}
                <div class="flex items-start gap-2.5 text-xs text-slate-700">
                  <CheckCircle2 class="w-4 h-4 text-emerald-600 flex-shrink-0 mt-0.5" />
                  <span class="font-medium leading-relaxed">{bullet}</span>
                </div>
              {/each}
            </div>
          </div>

          <div class="pt-4 border-t border-slate-200/60 flex items-center justify-between text-[11px] font-mono text-slate-500">
            <span class="truncate">{feature.techSpec}</span>
            <span class="inline-block w-1.5 h-1.5 rounded-full bg-blue-500"></span>
          </div>
        </div>
      {/each}
    </div>
  </div>
</section>
