<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    title = '',
    shortcut = '',
    placement = 'top',
    delay = 180,
    children
  }: {
    title: string;
    shortcut?: string;
    placement?: 'top' | 'bottom' | 'left' | 'right';
    delay?: number;
    children?: Snippet;
  } = $props();

  let visible = $state(false);
  let timeoutId: ReturnType<typeof setTimeout> | null = null;

  function show() {
    if (!title && !shortcut) return;
    if (timeoutId) clearTimeout(timeoutId);
    timeoutId = setTimeout(() => {
      visible = true;
    }, delay);
  }

  function hide() {
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
    visible = false;
  }

  const placementClasses = {
    top: 'bottom-full left-1/2 -translate-x-1/2 mb-2',
    bottom: 'top-full left-1/2 -translate-x-1/2 mt-2',
    left: 'right-full top-1/2 -translate-y-1/2 mr-2',
    right: 'left-full top-1/2 -translate-y-1/2 ml-2'
  };

  const arrowClasses = {
    top: 'top-full left-1/2 -translate-x-1/2 border-t-[#18181c]/95 border-x-transparent border-b-transparent border-[5px]',
    bottom: 'bottom-full left-1/2 -translate-x-1/2 border-b-[#18181c]/95 border-x-transparent border-t-transparent border-[5px]',
    left: 'left-full top-1/2 -translate-y-1/2 border-l-[#18181c]/95 border-y-transparent border-r-transparent border-[5px]',
    right: 'right-full top-1/2 -translate-y-1/2 border-r-[#18181c]/95 border-y-transparent border-l-transparent border-[5px]'
  };
</script>

<div
  role="presentation"
  class="relative inline-flex w-full"
  onmouseenter={show}
  onmouseleave={hide}
  onfocusin={show}
  onfocusout={hide}
>
  {@render children?.()}

  {#if visible}
    <div
      role="tooltip"
      class="absolute z-50 pointer-events-none whitespace-nowrap {placementClasses[placement]} animate-in fade-in zoom-in-95 duration-150"
    >
      <div class="relative px-2.5 py-1 rounded-md bg-[#18181c]/95 border border-white/10 shadow-2xl backdrop-blur-md flex items-center gap-1.5 text-xs font-serif text-white tracking-wide">
        <span>{title}</span>
        {#if shortcut}
          <kbd class="px-1.5 py-0.5 text-[10px] font-mono text-ant-text-muted bg-white/10 rounded border border-white/10 leading-none">
            {shortcut}
          </kbd>
        {/if}
        <!-- Decorative subtle arrow -->
        <div class="absolute w-0 h-0 {arrowClasses[placement]} pointer-events-none"></div>
      </div>
    </div>
  {/if}
</div>
