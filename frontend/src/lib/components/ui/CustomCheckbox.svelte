<script lang="ts">
  import { Check } from 'lucide-svelte';

  let {
    checked = false,
    indeterminate = false,
    size = 'md', // 'sm' | 'md'
    disabled = false,
    class: className = '',
  }: {
    checked?: boolean;
    indeterminate?: boolean;
    size?: 'sm' | 'md';
    disabled?: boolean;
    class?: string;
  } = $props();

  const sizeClasses = $derived(
    size === 'sm' ? 'w-3.5 h-3.5 rounded' : 'w-4 h-4 rounded-[5px]'
  );
  const iconSize = $derived(size === 'sm' ? 10 : 12);
</script>

<div
  class="relative inline-flex items-center justify-center {sizeClasses} border transition-all duration-200 ease-out select-none flex-shrink-0 {disabled ? 'opacity-40 cursor-not-allowed' : 'cursor-pointer'} {checked || indeterminate
    ? 'bg-ant-primary border-ant-primary text-white shadow-sm shadow-ant-primary/30 ring-2 ring-ant-primary/20'
    : 'bg-ant-bg-tertiary border-ant-border hover:border-ant-primary/50 text-transparent'} {className}"
>
  {#if checked}
    <div class="animate-in zoom-in-50 duration-150 flex items-center justify-center">
      <Check size={iconSize} class="stroke-[3]" />
    </div>
  {:else if indeterminate}
    <div class="w-2 h-0.5 bg-white rounded-full animate-in zoom-in-50 duration-150"></div>
  {/if}
</div>
