<script lang="ts">
  import { onMount, tick } from 'svelte';
  import {
    SlidersHorizontal,
    Check,
    ChevronDown,
    Zap,
    Scale,
    BrainCircuit
  } from 'lucide-svelte';

  export type ReasoningEffortLevel = 'low' | 'medium' | 'high';

  interface Props {
    value: ReasoningEffortLevel;
    disabled?: boolean;
    onChange?: (val: ReasoningEffortLevel) => void;
  }

  let {
    value = $bindable('medium'),
    disabled = false,
    onChange
  }: Props = $props();

  let isOpen = $state(false);
  let dropdownRef = $state<HTMLDivElement | null>(null);

  interface EffortOption {
    id: ReasoningEffortLevel;
    label: string;
    description: string;
    turns: string;
    icon: typeof Zap;
    color: string;
  }

  const options: EffortOption[] = [
    {
      id: 'low',
      label: 'Low Effort',
      description: 'Fast, direct responses with minimal chain-of-thought tokens',
      turns: 'Fastest',
      icon: Zap,
      color: 'text-ant-success'
    },
    {
      id: 'medium',
      label: 'Medium Effort',
      description: 'Balanced multi-step reasoning and precise tool planning',
      turns: 'Recommended',
      icon: Scale,
      color: 'text-ant-primary'
    },
    {
      id: 'high',
      label: 'High Effort',
      description: 'Deep exhaustive analysis, rigorous edge-case verification',
      turns: 'Thorough',
      icon: BrainCircuit,
      color: 'text-purple-400'
    }
  ];

  onMount(() => {
    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
        isOpen = false;
      }
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape' && isOpen) {
        isOpen = false;
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  });

  const selectedOption = $derived(
    options.find((opt) => opt.id === value) || options[1]
  );

  function toggleDropdown() {
    if (disabled) return;
    isOpen = !isOpen;
  }

  function handleSelect(effort: ReasoningEffortLevel) {
    value = effort;
    onChange?.(effort);
    isOpen = false;
  }
</script>

<div class="relative inline-block font-sans select-none text-left" bind:this={dropdownRef}>
  <!-- Trigger Pill Button -->
  <button
    type="button"
    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium bg-ant-bg border border-ant-border hover:border-ant-primary/60 text-ant-text hover:bg-ant-bg-tertiary transition-all shadow-sm focus:outline-none focus:ring-1 focus:ring-ant-primary/40 disabled:opacity-50 disabled:cursor-not-allowed group"
    onclick={toggleDropdown}
    {disabled}
    title="Select reasoning effort depth"
  >
    <div class="flex items-center justify-center w-4 h-4 rounded bg-ant-primary/10 text-ant-primary">
      <SlidersHorizontal size={11} />
    </div>
    <span class="text-[11px] font-semibold tracking-wide text-ant-primary uppercase font-mono">
      {selectedOption.id}
    </span>
    <ChevronDown
      size={12}
      class="text-ant-text-muted transition-transform duration-200 group-hover:text-ant-text {isOpen ? 'rotate-180 text-ant-primary' : ''}"
    />
  </button>

  <!-- Dropdown Menu / Popover -->
  {#if isOpen}
    <div
      class="absolute bottom-full left-0 mb-3 w-64 max-w-[90vw] bg-[#ffffff] dark:bg-[#181b26] border border-ant-border rounded-xl shadow-2xl z-[100] flex flex-col animate-in fade-in zoom-in-95 duration-150 ring-1 ring-black/10 dark:ring-white/10"
      style="box-shadow: 0 20px 40px -4px rgba(0, 0, 0, 0.45), 0 8px 16px -4px rgba(0, 0, 0, 0.25);"
    >
      <!-- Header -->
      <div class="px-3.5 py-2.5 border-b border-ant-border bg-[#fafafa] dark:bg-[#181b26] rounded-t-xl flex items-center justify-between">
        <div class="flex items-center gap-1.5 text-[11px] font-semibold text-ant-text">
          <SlidersHorizontal size={12} class="text-ant-primary" />
          <span>Reasoning Effort</span>
        </div>
        <span class="text-[10px] text-ant-text-muted">Grok Thinking</span>
      </div>

      <!-- Options List -->
      <div class="p-1.5 space-y-1 bg-[#ffffff] dark:bg-[#0f1117]">
        {#each options as opt}
          {@const isSelected = opt.id === value}
          {@const IconComponent = opt.icon}
          <button
            type="button"
            class="w-full text-left p-2 rounded-lg text-xs transition-all flex items-start justify-between group {isSelected
              ? 'bg-ant-primary/15 border border-ant-primary/40 text-ant-text'
              : 'hover:bg-ant-bg-tertiary border border-transparent text-ant-text-secondary hover:text-ant-text'}"
            onclick={() => handleSelect(opt.id)}
          >
            <div class="flex items-start gap-2.5 min-w-0 pr-2">
              <div class="mt-0.5 p-1 rounded-md bg-ant-bg border border-ant-border {opt.color}">
                <IconComponent size={13} />
              </div>
              <div class="flex flex-col min-w-0">
                <div class="flex items-center gap-1.5">
                  <span class="font-semibold text-[11px] {isSelected ? 'text-ant-primary' : 'text-ant-text'}">
                    {opt.label}
                  </span>
                  <span class="px-1 py-0.2 text-[9px] font-medium bg-ant-bg-tertiary text-ant-text-muted rounded border border-ant-border-secondary">
                    {opt.turns}
                  </span>
                </div>
                <span class="text-[10px] text-ant-text-muted mt-0.5 leading-tight">
                  {opt.description}
                </span>
              </div>
            </div>

            {#if isSelected}
              <div class="flex-shrink-0 mt-1 text-ant-primary">
                <Check size={14} />
              </div>
            {/if}
          </button>
        {/each}
      </div>

      <!-- Footer Quick Info -->
      <div class="px-3 py-2 bg-[#fafafa] dark:bg-[#181b26] border-t border-ant-border rounded-b-xl text-[10px] text-ant-text-muted">
        Controls budget for Grok agentic thought turns.
      </div>
    </div>
  {/if}
</div>
