<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    title?: string;
    extra?: Snippet;
    bordered?: boolean;
    hoverable?: boolean;
    size?: 'small' | 'default';
    class?: string;
    bodyClass?: string;
    children?: Snippet;
  }

  let {
    title = '',
    extra,
    bordered = true,
    hoverable = false,
    size = 'default',
    class: className = '',
    bodyClass = '',
    children
  }: Props = $props();
</script>

<div
  class="bg-ant-bg-secondary text-ant-text rounded-lg transition-all duration-200 {bordered ? 'border border-ant-border' : ''} {hoverable ? 'hover:border-ant-primary/50 hover:shadow-lg hover:-translate-y-0.5' : ''} {className}"
>
  {#if title || extra}
    <div
      class="flex items-center justify-between border-b border-ant-border-secondary {size === 'small' ? 'px-3 py-2 text-xs font-semibold' : 'px-4 py-3 text-sm font-semibold'}"
    >
      <div class="flex items-center space-x-2">
        <span>{title}</span>
      </div>
      {#if extra}
        <div class="text-xs text-ant-text-secondary">
          {@render extra()}
        </div>
      {/if}
    </div>
  {/if}

  <div class="{size === 'small' ? 'p-3' : 'p-4'} {bodyClass}">
    {#if children}
      {@render children()}
    {/if}
  </div>
</div>
