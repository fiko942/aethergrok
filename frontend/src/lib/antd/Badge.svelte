<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    count?: number | string;
    status?: 'success' | 'processing' | 'default' | 'error' | 'warning';
    color?: string;
    text?: string;
    dot?: boolean;
    overflowCount?: number;
    class?: string;
    children?: Snippet;
  }

  let {
    count,
    status,
    color,
    text = '',
    dot = false,
    overflowCount = 99,
    class: className = '',
    children
  }: Props = $props();

  const statusColors: Record<string, string> = {
    success: 'bg-ant-success',
    processing: 'bg-ant-primary animate-pulse',
    default: 'bg-ant-text-muted',
    error: 'bg-ant-error',
    warning: 'bg-ant-warning'
  };

  const displayCount = $derived.by(() => {
    if (typeof count === 'number' && count > overflowCount) {
      return `${overflowCount}+`;
    }
    return count;
  });
</script>

<div class="relative inline-flex items-center align-middle {className}">
  {#if children}
    {@render children()}
  {/if}

  {#if status || color}
    <div class="inline-flex items-center space-x-1.5 {children ? 'ml-2' : ''}">
      <span
        class="inline-block w-2 h-2 rounded-full {status ? statusColors[status] : ''}"
        style={color ? `background-color: ${color}` : ''}
      ></span>
      {#if text}
        <span class="text-xs text-ant-text-secondary">{text}</span>
      {/if}
    </div>
  {:else if dot}
    <span
      class="absolute -top-1 -right-1 w-2.5 h-2.5 bg-ant-error rounded-full border-2 border-ant-bg ring-1 ring-ant-bg"
    ></span>
  {:else if displayCount !== undefined && displayCount !== null}
    <span
      class="{children
        ? 'absolute -top-1.5 -right-2'
        : 'relative'} inline-flex items-center justify-center min-w-[18px] h-[18px] px-1 text-[10px] font-bold leading-none text-white bg-ant-error rounded-full border border-ant-bg shadow-sm"
    >
      {displayCount}
    </span>
  {/if}
</div>
