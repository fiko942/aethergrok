<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { inputShieldStore } from '$lib/stores/inputShield.svelte';
  import {
    Search,
    Check,
    Cpu,
    ChevronDown,
    Sparkles,
    SlidersHorizontal,
    Info,
    RefreshCw
  } from 'lucide-svelte';

  export interface ModelOption {
    id: string;
    name: string;
    description?: string;
    isDefault?: boolean;
  }

  interface Props {
    value: string;
    options?: ModelOption[];
    disabled?: boolean;
    onChange?: (val: string) => void;
  }

  let {
    value = $bindable('9router'),
    options = [],
    disabled = false,
    onChange
  }: Props = $props();

  let isOpen = $state(false);
  let searchQuery = $state('');
  let dropdownRef = $state<HTMLDivElement | null>(null);
  let searchInputRef = $state<HTMLInputElement | null>(null);
  let internalModels = $state<ModelOption[]>([]);
  let isLoading = $state(false);

  // Default fallback models if none provided
  const defaultModelsList: ModelOption[] = [
    { id: '9router', name: '9router', description: 'Default 9Router gateway model', isDefault: true },
    { id: '9router-general-purpose', name: '9router-general-purpose', description: 'General agentic tasks & code edit' },
    { id: '9router-explore', name: '9router-explore', description: 'Fast code search & exploration' },
    { id: '9router-plan', name: '9router-plan', description: 'High-depth reasoning & planning' },
  ];

  async function loadModels() {
    isLoading = true;
    try {
      const app = (window as any).go?.main?.App;
      if (typeof window !== 'undefined' && app?.GetAvailableModels) {
        const fetched: Array<{ id: string; name?: string; description?: string; isDefault?: boolean }> = await app.GetAvailableModels();
        if (fetched && fetched.length > 0) {
          internalModels = fetched.map((m) => ({
            id: m.id,
            name: m.name || m.id,
            description: m.description,
            isDefault: m.isDefault
          }));
          return;
        }
      }
    } catch (e) {
      console.warn('Failed to fetch models from bridge, using fallback:', e);
    } finally {
      isLoading = false;
    }

    if (options && options.length > 0) {
      internalModels = options;
    } else {
      internalModels = defaultModelsList;
    }
  }

  $effect(() => {
    if (options && options.length > 0) {
      internalModels = options;
    }
  });

  onMount(() => {
    loadModels();

    function handleClickOutside(event: MouseEvent) {
      if (dropdownRef && !dropdownRef.contains(event.target as Node)) {
        isOpen = false;
      }
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape' && isOpen) {
        isOpen = false;
      }
    }

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleKeyDown);

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  });

  const displayList = $derived(
    internalModels.length > 0 ? internalModels : defaultModelsList
  );

  const filteredOptions = $derived(
    displayList.filter(item => {
      const q = searchQuery.toLowerCase().trim();
      if (!q) return true;
      return (
        item.id.toLowerCase().includes(q) ||
        item.name.toLowerCase().includes(q) ||
        (item.description && item.description.toLowerCase().includes(q))
      );
    })
  );

  const selectedItem = $derived(
    displayList.find(item => item.id === value) || {
      id: value,
      name: value,
      description: 'Custom selected model'
    }
  );

  function toggleDropdown() {
    if (disabled) return;
    isOpen = !isOpen;
    if (isOpen) {
      searchQuery = '';
      tick().then(() => {
        searchInputRef?.focus();
      });
    }
  }

  function handleSelect(modelId: string) {
    value = modelId;
    onChange?.(modelId);
    isOpen = false;
  }
</script>

<div class="relative inline-block font-serif select-none text-left" bind:this={dropdownRef}>
  <!-- Trigger Pill Button -->
  <button
    type="button"
    class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg text-xs font-medium bg-ant-bg border border-ant-border-secondary dark:border-white/5 hover:border-blue-500/40 text-ant-text hover:bg-ant-bg-tertiary transition-all shadow-sm focus:outline-none focus:ring-1 focus:ring-ant-primary/40 disabled:opacity-50 disabled:cursor-not-allowed group"
    onclick={toggleDropdown}
    {disabled}
    title="Choose active model engine"
  >
    <div class="flex items-center justify-center w-4 h-4 rounded bg-ant-primary/10 text-ant-primary">
      <Cpu size={11} />
    </div>
    <span class="font-mono text-[11px] font-semibold text-ant-text tracking-tight truncate max-w-[160px]">
      {selectedItem.id}
    </span>
    <ChevronDown
      size={12}
      class="text-ant-text-muted transition-transform duration-200 group-hover:text-ant-text {isOpen ? 'rotate-180 text-ant-primary' : ''}"
    />
  </button>

  <!-- Dropdown Menu / Popover with Search & Options -->
  {#if isOpen}
    <div
      class="absolute bottom-full left-0 mb-3 w-80 max-w-[90vw] bg-ant-bg border border-ant-border-secondary dark:border-white/5 rounded-xl shadow-2xl z-[100] flex flex-col animate-in fade-in zoom-in-95 duration-150"
      style="box-shadow: 0 20px 40px -4px rgba(0, 0, 0, 0.45), 0 8px 16px -4px rgba(0, 0, 0, 0.25);"
    >
      <!-- Dropdown Header & Search Box -->
      <div class="p-3 border-b border-ant-border-secondary dark:border-white/5 bg-ant-bg-secondary rounded-t-xl">
        <div class="flex items-center justify-between mb-2 px-0.5">
          <div class="flex items-center gap-1.5 text-[11px] font-semibold text-ant-text">
            <Sparkles size={12} class="text-ant-primary" />
            <span>Select Model Engine</span>
          </div>
          <button
            type="button"
            class="text-[10px] text-ant-text-muted hover:text-ant-primary p-0.5 rounded transition flex items-center gap-1"
            onclick={() => loadModels()}
            title="Refresh models from Grok CLI"
          >
            <RefreshCw size={10} class={isLoading ? 'animate-spin' : ''} />
            <span>Sync</span>
          </button>
        </div>

        <div class="relative">
          <Search size={12} class="absolute left-2.5 top-1/2 -translate-y-1/2 text-ant-text-muted" />
          <input
            bind:this={searchInputRef}
            type="text"
            bind:value={searchQuery}
            readonly={inputShieldStore.isReadOnly}
            placeholder="Search model by name..."
            class="w-full pl-7 pr-3 py-1.5 text-xs bg-ant-bg border border-ant-border-secondary dark:border-white/5 rounded-lg text-ant-text placeholder:text-ant-text-muted outline-none focus:border-ant-primary focus:ring-1 focus:ring-ant-primary/30 transition shadow-inner"
          />
        </div>
      </div>

      <!-- Models List -->
      <div class="max-h-60 overflow-y-auto p-1.5 space-y-1 scrollbar-thin bg-ant-bg">
        {#if filteredOptions.length === 0}
          <div class="py-6 text-center text-xs text-ant-text-muted">
            <Info size={16} class="mx-auto mb-1.5 opacity-50" />
            <span>No models found matching "{searchQuery}"</span>
          </div>
        {:else}
          {#each filteredOptions as item}
            {@const isSelected = item.id === value}
            <button
              type="button"
              class="w-full text-left p-2 rounded-lg text-xs transition-all flex items-start justify-between group {isSelected
                ? 'bg-ant-primary/15 border border-ant-primary/40 text-ant-text'
                : 'hover:bg-ant-bg-tertiary border border-transparent text-ant-text-secondary hover:text-ant-text'}"
              onclick={() => handleSelect(item.id)}
            >
              <div class="flex flex-col min-w-0 pr-2">
                <div class="flex items-center gap-1.5">
                  <span class="font-mono font-medium text-[11px] {isSelected ? 'text-ant-primary font-bold' : 'text-ant-text'}">
                    {item.id}
                  </span>
                  {#if item.isDefault}
                    <span class="px-1 py-0.2 text-[9px] font-semibold bg-ant-primary/20 text-ant-primary rounded border border-ant-primary/30">
                      default
                    </span>
                  {/if}
                </div>
                {#if item.description}
                  <span class="text-[10px] text-ant-text-muted mt-0.5 truncate leading-tight">
                    {item.description}
                  </span>
                {/if}
              </div>

              {#if isSelected}
                <div class="flex-shrink-0 mt-0.5 text-ant-primary">
                  <Check size={14} />
                </div>
              {/if}
            </button>
          {/each}
        {/if}
      </div>

      <!-- Footer Quick Status -->
      <div class="px-3 py-2 bg-ant-bg-secondary border-t border-ant-border-secondary dark:border-white/5 rounded-b-xl flex items-center justify-between text-[10px] text-ant-text-muted">
        <span>{filteredOptions.length} available</span>
        <span class="font-mono text-[9px]">grok CLI</span>
      </div>
    </div>
  {/if}
</div>
