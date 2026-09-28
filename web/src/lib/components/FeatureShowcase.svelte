<script lang="ts">
  import { featuresData } from '../data/features';
  import { CheckCircle2, ChevronRight } from 'lucide-svelte';

  let activeTabId = $state(featuresData[0].id);

  let activeFeature = $derived(
    featuresData.find((f) => f.id === activeTabId) || featuresData[0]
  );
</script>

<section id="features" class="py-20 md:py-32 bg-white">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="text-center max-w-3xl mx-auto mb-16">
      <h2 class="text-xs font-bold uppercase tracking-widest text-blue-600 mb-2">
        Feature Breakdown
      </h2>
      <p class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
        Everything Built for High-Stakes Development
      </p>
      <p class="mt-4 text-base sm:text-lg text-slate-600">
        Explore the deep capabilities that make AetherGrok a distinct, production-grade desktop client.
      </p>
    </div>

    <!-- Feature Navigation Tabs -->
    <div class="flex items-center justify-center gap-2 overflow-x-auto pb-4 mb-12 scrollbar-none">
      <div class="inline-flex p-1.5 bg-slate-100/90 rounded-2xl border border-slate-200/80">
        {#each featuresData as feature}
          <button
            onclick={() => (activeTabId = feature.id)}
            class="px-4 py-2.5 rounded-xl text-xs sm:text-sm font-semibold transition-all duration-150 whitespace-nowrap flex items-center gap-2 {activeTabId === feature.id
              ? 'bg-white text-blue-700 shadow-sm border border-slate-200/60'
              : 'text-slate-600 hover:text-slate-900 hover:bg-slate-200/50'}"
          >
            <span>{feature.title}</span>
          </button>
        {/each}
      </div>
    </div>

    <!-- Active Feature Display Card -->
    <div class="bg-slate-50 border border-slate-200/80 rounded-3xl p-6 sm:p-10 lg:p-12 shadow-sm">
      <div class="grid grid-cols-1 lg:grid-cols-12 gap-10 lg:gap-12 items-center">
        <!-- Left: Text Content -->
        <div class="lg:col-span-5 space-y-6">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-blue-100/80 text-blue-800 border border-blue-200">
            {activeFeature.badge}
          </div>

          <h3 class="text-2xl sm:text-3xl font-extrabold text-slate-900 tracking-tight">
            {activeFeature.tagline}
          </h3>

          <p class="text-slate-600 text-sm sm:text-base leading-relaxed">
            {activeFeature.description}
          </p>

          <!-- Bullet Points -->
          <div class="pt-2 space-y-3">
            {#each activeFeature.bullets as bullet}
              <div class="flex items-start gap-3">
                <CheckCircle2 class="w-5 h-5 text-emerald-600 flex-shrink-0 mt-0.5" />
                <span class="text-sm text-slate-700 font-medium">{bullet}</span>
              </div>
            {/each}
          </div>
        </div>

        <!-- Right: Screenshot Preview -->
        <div class="lg:col-span-7">
          <div class="relative rounded-2xl overflow-hidden bg-slate-900 border border-slate-300 shadow-xl">
            <div class="h-8 bg-slate-800/90 px-3.5 flex items-center gap-1.5 border-b border-slate-700">
              <div class="w-2.5 h-2.5 rounded-full bg-slate-600"></div>
              <div class="w-2.5 h-2.5 rounded-full bg-slate-600"></div>
              <div class="w-2.5 h-2.5 rounded-full bg-slate-600"></div>
              <span class="text-[11px] font-mono text-slate-400 ml-2 truncate">
                {activeFeature.title} — AetherGrok Live View
              </span>
            </div>

            <img
              src="{activeFeature.screenshot}"
              alt="{activeFeature.imageAlt}"
              class="w-full h-auto object-cover max-h-[440px] block"
              loading="lazy"
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</section>
