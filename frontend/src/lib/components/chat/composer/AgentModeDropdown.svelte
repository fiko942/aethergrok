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
  <!-- Pill Button Trigger (Matching screenshot: [Robot Icon] Agent mode) -->
  <button
    type="button"
    onclick={toggleOpen}
    {disabled}
    class="flex items-center space-x-1.5 px-2.5 py-1 rounded-lg text-xs font-medium bg-ant-bg-secondary hover:bg-ant-bg-tertiary text-ant-text transition-colors border border-ant-border-secondary/60 {isOpen ? 'border-ant-primary/50 text-ant-primary' : ''} {disabled ? 'opacity-50 cursor-not-allowed' : ''}"
    title="Choose Grok Agent Execution Mode"
  >
    <Bot size={13} class="text-ant-text-secondary" />
    <span class="text-[11px]">{currentModeObj.title}</span>
  </button>

  <!-- Dropdown Menu (Matching screenshot) -->
  {#if isOpen}
    <div
      class="absolute bottom-full right-0 mb-2.5 w-72 bg-ant-bg border border-ant-border-secondary rounded-xl shadow-2xl z-50 p-1.5 select-none text-xs animate-in fade-in zoom-in-95 duration-100 divide-y divide-ant-border-secondary/40"
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
