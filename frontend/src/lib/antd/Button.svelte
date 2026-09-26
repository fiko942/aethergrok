<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    type?: 'primary' | 'default' | 'dashed' | 'text' | 'link';
    danger?: boolean;
    size?: 'small' | 'middle' | 'large';
    disabled?: boolean;
    loading?: boolean;
    block?: boolean;
    class?: string;
    onclick?: (e: MouseEvent) => void;
    children?: Snippet;
  }

  let {
    type = 'default',
    danger = false,
    size = 'middle',
    disabled = false,
    loading = false,
    block = false,
    class: className = '',
    onclick,
    children
  }: Props = $props();

  const sizeClasses = {
    small: 'px-2.5 py-1 text-xs h-7 rounded',
    middle: 'px-3.5 py-1.5 text-sm h-8 rounded-md',
    large: 'px-4 py-2 text-base h-10 rounded-lg'
  };

  const baseClasses = 'inline-flex items-center justify-center font-serif font-medium transition-all duration-150 select-none outline-none focus-visible:ring-2 focus-visible:ring-ant-primary/50 disabled:opacity-45 disabled:cursor-not-allowed cursor-pointer whitespace-nowrap shrink-0';

  function getTypeClasses(): string {
    if (danger) {
      if (type === 'primary') {
        return 'bg-ant-error hover:bg-[#ff7875] active:bg-[#d9363e] text-white border border-ant-error shadow-sm';
      }
      return 'bg-transparent hover:bg-ant-error/10 text-ant-error border border-ant-error/40 hover:border-ant-error';
    }

    switch (type) {
      case 'primary':
        return 'bg-ant-primary hover:bg-ant-primary-hover active:bg-ant-primary-active text-white border border-ant-primary shadow-sm hover:shadow-[0_0_12px_rgba(22,119,255,0.35)]';
      case 'dashed':
        return 'bg-ant-bg-secondary hover:bg-ant-bg-tertiary text-ant-text hover:text-ant-primary-hover border border-dashed border-ant-border hover:border-ant-primary';
      case 'text':
        return 'bg-transparent hover:bg-ant-bg-tertiary active:bg-ant-border text-ant-text-secondary hover:text-ant-text border-transparent';
      case 'link':
        return 'bg-transparent hover:underline text-ant-primary hover:text-ant-primary-hover p-0 h-auto border-transparent';
      case 'default':
      default:
        return 'bg-ant-bg-secondary hover:bg-ant-bg-tertiary active:bg-ant-bg-tertiary text-ant-text hover:text-ant-primary-hover border border-ant-border-secondary hover:border-ant-border';
    }
  }
</script>

<button
  type="button"
  class="{baseClasses} {sizeClasses[size]} {getTypeClasses()} {block ? 'w-full' : ''} {className}"
  {disabled}
  {onclick}
>
  {#if loading}
    <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-current" fill="none" viewBox="0 0 24 24">
      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
    </svg>
  {/if}
  {#if children}
    {@render children()}
  {/if}
</button>
