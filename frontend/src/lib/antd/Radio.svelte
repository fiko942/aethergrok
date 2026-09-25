<script lang="ts">
  let {
    checked = false,
    value = '',
    disabled = false,
    label = '',
    description = '',
    onselect
  }: {
    checked: boolean;
    value?: string;
    disabled?: boolean;
    label?: string;
    description?: string;
    onselect?: (val: string) => void;
  } = $props();

  function select() {
    if (disabled) return;
    onselect?.(value);
  }
</script>

<div
  role="radio"
  aria-checked={checked}
  tabindex={disabled ? -1 : 0}
  onclick={select}
  onkeydown={(e) => {
    if ((e.key === ' ' || e.key === 'Enter') && !disabled) {
      e.preventDefault();
      select();
    }
  }}
  class="inline-flex items-start gap-2.5 cursor-pointer select-none group focus:outline-none {disabled ? 'opacity-50 cursor-not-allowed' : ''}"
>
  <div
    class="w-4 h-4 mt-0.5 rounded-full flex items-center justify-center transition-all duration-150 flex-shrink-0 border {checked
      ? 'border-ant-primary bg-ant-primary/10 shadow-sm'
      : 'border-zinc-700/80 bg-ant-bg-tertiary group-hover:border-zinc-500'}"
  >
    <div
      class="w-2 h-2 rounded-full bg-ant-primary transition-all duration-150 {checked
        ? 'scale-100 opacity-100'
        : 'scale-0 opacity-0'}"
    ></div>
  </div>

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
</div>
