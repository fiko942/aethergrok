<script lang="ts">
  import type { Snippet } from 'svelte';

  interface Props {
    checked?: boolean;
    disabled?: boolean;
    size?: 'small' | 'default';
    checkedChildren?: Snippet | string;
    unCheckedChildren?: Snippet | string;
    class?: string;
    onchange?: (checked: boolean) => void;
  }

  let {
    checked = $bindable(false),
    disabled = false,
    size = 'default',
    checkedChildren,
    unCheckedChildren,
    class: className = '',
    onchange
  }: Props = $props();

  function toggle() {
    if (disabled) return;
    checked = !checked;
    onchange?.(checked);
  }
</script>

<button
  type="button"
  role="switch"
  aria-checked={checked}
  {disabled}
  onclick={toggle}
  class="relative inline-flex items-center flex-shrink-0 cursor-pointer transition-colors duration-200 ease-in-out border border-transparent rounded-full focus:outline-none focus-visible:ring-2 focus-visible:ring-ant-primary/50 disabled:opacity-40 disabled:cursor-not-allowed select-none
    {size === 'small' ? 'h-4 w-7' : 'h-5 w-10'}
    {checked ? 'bg-ant-primary' : 'bg-ant-border'}
    {className}"
>
  <span
    aria-hidden="true"
    class="pointer-events-none inline-block rounded-full bg-white shadow transform ring-0 transition duration-200 ease-in-out
      {size === 'small' ? 'h-3 w-3' : 'h-4 w-4'}
      {checked
        ? size === 'small' ? 'translate-x-3.5' : 'translate-x-5'
        : 'translate-x-0.5'}"
  ></span>

  {#if checked && checkedChildren}
    <span class="absolute left-1.5 text-[9px] font-semibold text-white select-none">
      {#if typeof checkedChildren === 'string'}
        {checkedChildren}
      {:else}
        {@render checkedChildren()}
      {/if}
    </span>
  {:else if !checked && unCheckedChildren}
    <span class="absolute right-1.5 text-[9px] font-semibold text-ant-text-muted select-none">
      {#if typeof unCheckedChildren === 'string'}
        {unCheckedChildren}
      {:else}
        {@render unCheckedChildren()}
      {/if}
    </span>
  {/if}
</button>
