<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Check, Bot, ListTodo, Zap } from 'lucide-svelte';

  export type AgentModeType = 'agent' | 'plan' | 'yolo';

  interface Props {
    mode?: AgentModeType;
    disabled?: boolean;
    onChange?: (mode: AgentModeType) => void;
  }

  let {
    mode = $bindable('agent'),
    disabled = false,
    onChange
  }: Props = $props();

  let isOpen = $state(false);
  let containerRef = $state<HTMLDivElement | null>(null);

  const modes: { id: AgentModeType; title: string; desc: string; icon: typeof Bot }[] = [
    {
      id: 'agent',
      title: 'Agent mode',
      desc: 'Grok acts directly, asking approval only for changes it judges sensitive',
      icon: Bot
    },
    {
      id: 'plan',
      title: 'Plan mode',
      desc: 'Grok explores and proposes a plan; file writes and commands are blocked until you approve it',
      icon: ListTodo
    },
    {
      id: 'yolo',
      title: 'Auto accept',
      desc: 'Grok automatically approves all permission requests (YOLO)',
      icon: Zap
    }
  ];

  const currentModeObj = $derived.by(() => {
    return modes.find((m) => m.id === mode) || modes[0];
  });

  const CurrentIcon = $derived(currentModeObj.icon);

  function toggleOpen(e: MouseEvent) {
    e.stopPropagation();
    if (disabled) return;
    isOpen = !isOpen;
  }

  function handleSelect(newMode: AgentModeType, e: MouseEvent) {
    e.stopPropagation();
    mode = newMode;
    onChange?.(newMode);
    isOpen = false;
  }

  function handleClickOutside(e: PointerEvent) {
    if (isOpen && containerRef && !containerRef.contains(e.target as Node)) {
      isOpen = false;
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      isOpen = false;
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
  <!-- Pill Button Trigger (Dynamically reflects mode with appropriate icon & color) -->
  <button
    type="button"
    onclick={toggleOpen}
    {disabled}
    class="flex items-center space-x-1.5 px-2.5 py-1 rounded-lg text-xs font-medium transition-all border {mode === 'plan' ? 'bg-violet-500/10 text-violet-300 border-violet-500/30 hover:bg-violet-500/20' : mode === 'yolo' ? 'bg-amber-500/10 text-amber-300 border-amber-500/30 hover:bg-amber-500/20' : 'bg-ant-bg-secondary/40 hover:bg-ant-bg-tertiary text-ant-text border-transparent hover:border-white/5'} {isOpen ? '!ring-1 !ring-ant-primary/40' : ''} {disabled ? 'opacity-50 cursor-not-allowed' : ''}"
    title="Choose Grok Agent Execution Mode"
  >
    <CurrentIcon size={13} class={mode === 'plan' ? 'text-violet-400' : mode === 'yolo' ? 'text-amber-400' : 'text-ant-text-secondary'} />
    <span class="text-[11px] font-medium">{currentModeObj.title}</span>
  </button>

  <!-- Dropdown Menu (Matching screenshot) -->
  {#if isOpen}
    <div
      class="absolute bottom-full right-0 mb-2.5 w-72 bg-ant-bg border border-white/10 rounded-xl shadow-2xl z-50 p-1.5 select-none text-xs animate-in fade-in zoom-in-95 duration-100 divide-y divide-white/5 backdrop-blur-md"
      role="menu"
      aria-label="Agent Mode Options"
    >
      {#each modes as m}
        {@const isSelected = mode === m.id}
        {@const IconComp = m.icon}
        <button
          type="button"
          onclick={(e) => handleSelect(m.id, e)}
          class="w-full flex items-start justify-between p-2.5 rounded-lg text-left transition-colors {isSelected ? 'bg-ant-primary/10' : 'hover:bg-ant-bg-secondary'}"
        >
          <div class="flex items-start space-x-2.5 min-w-0 pr-2">
            <IconComp size={15} class="flex-shrink-0 mt-0.5 {isSelected ? 'text-ant-primary' : 'text-ant-text-secondary'}" />
            <div class="min-w-0 space-y-0.5">
              <div class="text-xs font-semibold {isSelected ? 'text-ant-primary' : 'text-ant-text'}">{m.title}</div>
              <div class="text-[10.5px] text-ant-text-secondary leading-relaxed">{m.desc}</div>
            </div>
          </div>
          {#if isSelected}
            <Check size={14} class="text-ant-primary flex-shrink-0 mt-0.5" />
          {/if}
        </button>
      {/each}
    </div>
  {/if}
</div>
