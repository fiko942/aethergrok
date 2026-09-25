<script lang="ts">
  import { Check } from 'lucide-svelte';

  let {
    checked = $bindable(false),
    disabled = false,
    label = '',
    description = '',
    onchange
  }: {
    checked?: boolean;
    disabled?: boolean;
    label?: string;
    description?: string;
    onchange?: (val: boolean) => void;
  } = $props();

  function toggle() {
    if (disabled) return;
    checked = !checked;
    onchange?.(checked);
  }
</script>

<label class="inline-flex items-start gap-2.5 cursor-pointer select-none group {disabled ? 'opacity-50 cursor-not-allowed' : ''}">
  <button
    type="button"
    role="checkbox"
    aria-checked={checked}
    {disabled}
    onclick={toggle}
    class="w-4 h-4 mt-0.5 rounded flex items-center justify-center transition-all duration-150 flex-shrink-0 {checked
      ? 'bg-ant-primary border border-ant-primary text-white shadow-sm'
      : 'bg-ant-bg-tertiary border border-zinc-700/80 group-hover:border-zinc-500 text-transparent'}"
  >
    <Check size={11} class="stroke-[3] transition-transform duration-150 {checked ? 'scale-100 opacity-100' : 'scale-75 opacity-0'}" />
  </button>

  {#if label || description}
    <div class="flex flex-col">
      {#if label}
        <span class="text-xs font-medium text-zinc-200 group-hover:text-white transition">{label}</span>
      {/if}
      {#if description}
        <span class="text-[10.5px] text-zinc-400 mt-0.5">{description}</span>
      {/if}
    </div>
  {/if}
</label>
