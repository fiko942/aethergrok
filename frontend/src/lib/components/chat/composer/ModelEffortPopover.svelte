<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Check, ChevronDown, Sparkles, Cpu, Brain, Zap } from 'lucide-svelte';
  import type { ReasoningEffort } from '$lib/stores/settings.svelte';

  interface Props {
    model: string;
    reasoningEffort: ReasoningEffort;
    disabled?: boolean;
    onModelChange: (model: string) => void;
    onEffortChange: (effort: ReasoningEffort) => void;
  }

  let {
    model = $bindable('9router'),
    reasoningEffort = $bindable('medium'),
    disabled = false,
    onModelChange,
    onEffortChange
  }: Props = $props();

  let isOpen = $state(false);
  let isModelPickerOpen = $state(false);
  let containerRef = $state<HTMLDivElement | null>(null);

  const availableModels = [
    { id: '9router', label: '9Router', desc: 'Default unified router gateway', icon: Sparkles },
    { id: '9router-general-purpose', label: '9Router General', desc: 'Balanced world knowledge', icon: Brain },
    { id: '9router-plan', label: '9Router Plan', desc: 'Architectural planning engine', icon: Cpu },
    { id: '9router-explore', label: '9Router Explore', desc: 'Fast file scan & discovery', icon: Zap }
  ];

  const effortLevels: { id: ReasoningEffort; label: string; level: number }[] = [
    { id: 'none', label: 'None', level: 0 },
    { id: 'low', label: 'Low', level: 1 },
    { id: 'medium', label: 'Medium', level: 2 },
    { id: 'high', label: 'High', level: 3 },
    { id: 'max', label: 'Max', level: 4 }
  ];

  const currentLevelIndex = $derived.by(() => {
    const found = effortLevels.findIndex((e) => e.id === reasoningEffort);
    return found !== -1 ? found : 2;
  });

  const activeModelObj = $derived.by(() => {
    return availableModels.find((m) => m.id === model) || {
      id: model,
      label: model.charAt(0).toUpperCase() + model.slice(1),
      desc: 'Custom model engine',
      icon: Sparkles
    };
  });

  function toggleOpen(e: MouseEvent) {
    e.stopPropagation();
    if (disabled) return;
    isOpen = !isOpen;
    if (!isOpen) isModelPickerOpen = false;
  }

  function handleSelectEffort(level: (typeof effortLevels)[0], e: MouseEvent) {
    e.stopPropagation();
    reasoningEffort = level.id;
    onEffortChange(level.id);
  }

  function handleSelectModel(modelId: string, e: MouseEvent) {
    e.stopPropagation();
    model = modelId;
    onModelChange(modelId);
    isModelPickerOpen = false;
  }

  function handleClickOutside(e: PointerEvent) {
    if (isOpen && containerRef && !containerRef.contains(e.target as Node)) {
      isOpen = false;
      isModelPickerOpen = false;
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      isOpen = false;
      isModelPickerOpen = false;
    }
  }

  onMount(() => {
    window.addEventListener('pointerdown', handleClickOutside);
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('pointerdown', handleClickOutside);
      window.removeEventListener('keydown', handleKeyDown);
    }
  });
</script>

<div class="relative inline-flex items-center" bind:this={containerRef}>
  <!-- Sliders Button Trigger -->
  <button
    type="button"
    onclick={toggleOpen}
    {disabled}
    class="p-1.5 rounded-lg text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition-colors flex items-center justify-center {isOpen ? 'text-ant-primary bg-ant-bg-tertiary' : ''} {disabled ? 'opacity-50 cursor-not-allowed' : ''}"
    title="Model & Reasoning Effort settings"
    aria-label="Model & Reasoning Effort"
  >
    <!-- Custom SVG Sliders Horizontal Icon matching reference -->
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
      <line x1="4" y1="21" x2="4" y2="14"></line>
      <line x1="4" y1="10" x2="4" y2="3"></line>
      <line x1="12" y1="21" x2="12" y2="12"></line>
      <line x1="12" y1="8" x2="12" y2="3"></line>
      <line x1="20" y1="21" x2="20" y2="16"></line>
      <line x1="20" y1="12" x2="20" y2="3"></line>
      <line x1="1" y1="14" x2="7" y2="14"></line>
      <line x1="9" y1="8" x2="15" y2="8"></line>
      <line x1="17" y1="16" x2="23" y2="16"></line>
    </svg>
  </button>

  <!-- Popover Panel -->
  {#if isOpen}
    <div
      class="absolute bottom-full left-0 mb-2.5 w-72 bg-ant-bg border border-ant-border-secondary rounded-xl shadow-2xl z-50 p-3 select-none text-xs animate-in fade-in zoom-in-95 duration-100"
      role="dialog"
      aria-label="Model and Effort"
    >
      <!-- Small Header -->
      <div class="text-[10px] font-semibold text-ant-text-muted uppercase tracking-wider mb-2.5">
        Model and Effort
      </div>

      <!-- Main Row: Model Switcher on Left + Effort Dots on Right -->
      <div class="flex items-center justify-between bg-ant-bg-secondary/70 border border-ant-border-secondary/60 rounded-lg p-2 relative">
        <!-- Model Selector Button -->
        <button
          type="button"
          onclick={(e) => { e.stopPropagation(); isModelPickerOpen = !isModelPickerOpen; }}
          class="flex items-center space-x-1.5 text-xs font-medium text-ant-text hover:text-ant-primary transition-colors py-0.5 px-1 rounded hover:bg-ant-bg-tertiary"
        >
          <span class="truncate max-w-[100px]">{activeModelObj.label}</span>
          <ChevronDown size={12} class="text-ant-text-muted flex-shrink-0 transition-transform {isModelPickerOpen ? 'rotate-180' : ''}" />
        </button>

        <!-- Effort Level Circles (Matching screenshot: ○ ○ ● ○ ○) -->
        <div class="flex items-center space-x-1.5" title={`Effort: ${reasoningEffort.toUpperCase()}`}>
          {#each effortLevels as lvl, idx}
            {@const isFilled = idx <= currentLevelIndex}
            {@const isCurrent = idx === currentLevelIndex}
            <button
              type="button"
              onclick={(e) => handleSelectEffort(lvl, e)}
              class="w-3.5 h-3.5 rounded-full flex items-center justify-center transition-all p-0 border {isCurrent
                ? 'bg-ant-primary border-ant-primary shadow-sm scale-110'
                : isFilled
                ? 'bg-ant-primary/40 border-ant-primary/60 hover:bg-ant-primary/60'
                : 'bg-transparent border-ant-border-secondary hover:border-ant-text-muted'}"
              title={`Effort: ${lvl.label}`}
              aria-label={`Set effort to ${lvl.label}`}
            >
              {#if isCurrent}
                <span class="w-1 h-1 rounded-full bg-white"></span>
              {/if}
            </button>
          {/each}
        </div>
      </div>

      <!-- Nested Model Picker Dropdown -->
      {#if isModelPickerOpen}
        <div class="mt-2 pt-2 border-t border-ant-border-secondary/60 space-y-1 animate-in fade-in duration-100">
          <div class="text-[10px] text-ant-text-muted px-1 pb-1">Select Architecture</div>
          {#each availableModels as m}
            {@const isSelected = model === m.id}
            {@const IconComponent = m.icon}
            <button
              type="button"
              onclick={(e) => handleSelectModel(m.id, e)}
              class="w-full flex items-center justify-between p-1.5 rounded-lg text-left transition-colors {isSelected ? 'bg-ant-primary/15 text-ant-primary font-medium' : 'text-ant-text hover:bg-ant-bg-tertiary'}"
            >
              <div class="flex items-center space-x-2 min-w-0">
                <IconComponent size={13} class={isSelected ? 'text-ant-primary' : 'text-ant-text-muted'} />
                <div class="min-w-0">
                  <div class="text-[11px] truncate leading-tight">{m.label}</div>
                  <div class="text-[9.5px] text-ant-text-muted truncate">{m.desc}</div>
                </div>
              </div>
              {#if isSelected}
                <Check size={12} class="text-ant-primary flex-shrink-0" />
              {/if}
            </button>
          {/each}
        </div>
      {/if}
    </div>
  {/if}
</div>
