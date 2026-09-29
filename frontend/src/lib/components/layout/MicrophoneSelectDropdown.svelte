<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Mic, Check, ChevronDown, Radio, Headphones, Sparkles, Volume2 } from 'lucide-svelte';
  import type { AudioInputDevice } from '$lib/utils/voiceRecorder';

  interface Props {
    selectedDeviceId: string;
    devices: AudioInputDevice[];
    disabled?: boolean;
    onselect: (deviceId: string) => void;
  }

  let { selectedDeviceId, devices, disabled = false, onselect }: Props = $props();

  let isOpen = $state(false);
  let dropdownRef: HTMLDivElement | null = $state(null);

  let activeDevice = $derived.by(() => {
    if (!selectedDeviceId) {
      return devices.find((d) => d.isDefault) || devices[0] || null;
    }
    return devices.find((d) => d.deviceId === selectedDeviceId) || devices[0] || null;
  });

  function toggleDropdown() {
    if (disabled) return;
    isOpen = !isOpen;
  }

  function handleSelect(deviceId: string) {
    onselect(deviceId);
    isOpen = false;
  }

  function handleClickOutside(e: MouseEvent) {
    if (dropdownRef && !dropdownRef.contains(e.target as Node)) {
      isOpen = false;
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape' && isOpen) {
      isOpen = false;
    }
  }

  onMount(() => {
    window.addEventListener('click', handleClickOutside);
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    window.removeEventListener('click', handleClickOutside);
    window.removeEventListener('keydown', handleKeyDown);
  });

  function getDeviceIcon(transport?: string) {
    switch (transport) {
      case 'bluetooth':
        return Headphones;
      case 'virtual':
        return Radio;
      case 'continuity':
        return Volume2;
      default:
        return Mic;
    }
  }

  function getTransportLabel(transport?: string): string {
    switch (transport) {
      case 'built-in':
        return 'Built-in Audio';
      case 'bluetooth':
        return 'Bluetooth';
      case 'usb':
        return 'USB Audio';
      case 'virtual':
        return 'Virtual Device';
      case 'continuity':
        return 'Continuity';
      default:
        return 'Microphone';
    }
  }
</script>

<div class="relative w-full font-serif select-none text-left" bind:this={dropdownRef}>
  <!-- Custom Themed Trigger Button -->
  <button
    type="button"
    class="w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl border transition-all duration-150 outline-none text-left group {isOpen
      ? 'border-ant-border-secondary dark:border-white/10 bg-ant-bg ring-1 ring-ant-primary/20 shadow-xs'
      : 'border-ant-border-secondary dark:border-white/5 bg-ant-bg hover:border-ant-border-secondary hover:bg-ant-bg-secondary/40'} {disabled
      ? 'opacity-50 cursor-not-allowed'
      : 'cursor-pointer'}"
    onclick={toggleDropdown}
    {disabled}
  >
    <div class="flex items-center space-x-3 min-w-0 pr-2">
      <!-- Icon pill with primary glow when active -->
      <div
        class="w-7 h-7 rounded-lg flex items-center justify-center flex-shrink-0 transition-colors {isOpen
          ? 'bg-ant-primary/15 text-ant-primary'
          : 'bg-ant-bg-tertiary text-ant-text-secondary group-hover:text-ant-text'}"
      >
        <Mic size={14} />
      </div>

      <div class="flex flex-col min-w-0">
        <div class="flex items-center space-x-2">
          <span class="text-xs font-semibold text-ant-text truncate">
            {activeDevice ? activeDevice.label : 'Default System Microphone'}
          </span>
          {#if activeDevice?.isDefault || !selectedDeviceId}
            <span
              class="px-1.5 py-0.5 text-[9px] uppercase tracking-wider font-semibold rounded bg-ant-primary/15 text-ant-primary border-0 flex-shrink-0"
            >
              Default
            </span>
          {/if}
        </div>
        <span class="text-[10px] text-ant-text-muted truncate mt-0.5">
          {activeDevice ? getTransportLabel(activeDevice.transport) : 'System Audio Input'}
        </span>
      </div>
    </div>

    <ChevronDown
      size={14}
      class="text-ant-text-muted transition-transform duration-200 group-hover:text-ant-text flex-shrink-0 {isOpen
        ? 'rotate-180 text-ant-primary'
        : ''}"
    />
  </button>

  <!-- Themed Floating Dropdown Menu -->
  {#if isOpen}
    <div
      class="absolute top-full left-0 mt-2 w-full bg-ant-bg-secondary border border-ant-border-secondary dark:border-white/10 rounded-xl shadow-2xl z-50 overflow-hidden flex flex-col backdrop-blur-xl animate-in fade-in zoom-in-95 duration-100"
    >
      <!-- Menu Header -->
      <div class="px-3.5 py-2 border-b border-ant-border-secondary dark:border-white/5 bg-ant-bg flex items-center justify-between text-ant-text">
        <span class="text-[11px] font-medium text-ant-text-secondary">Connected Hardware Inputs</span>
        <span class="text-[10px] text-ant-text-muted">{devices.length} {devices.length === 1 ? 'device' : 'devices'}</span>
      </div>

      <!-- Device List -->
      <div class="p-1.5 space-y-1 max-h-60 overflow-y-auto custom-scrollbar">
        <!-- Default Option -->
        {#if true}
          {@const isDefaultSelected = !selectedDeviceId}
          <button
            type="button"
            class="w-full text-left p-2 rounded-lg text-xs transition-all flex items-center justify-between group {isDefaultSelected
              ? 'bg-ant-primary/15 text-ant-text'
              : 'hover:bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-text'}"
            onclick={() => handleSelect('')}
          >
            <div class="flex items-center space-x-2.5 min-w-0 pr-2">
              <div class="p-1.5 rounded-md bg-ant-bg text-ant-primary flex-shrink-0 border border-ant-border-secondary dark:border-white/5">
                <Sparkles size={13} />
              </div>
              <div class="flex flex-col min-w-0">
                <div class="flex items-center space-x-1.5">
                  <span class="font-medium text-xs {isDefaultSelected ? 'text-ant-primary' : 'text-ant-text'} truncate">
                    Default System Microphone
                  </span>
                  <span class="px-1 py-0.2 text-[9px] font-medium bg-ant-bg text-ant-text-muted rounded border border-ant-border-secondary dark:border-white/5">
                    System
                  </span>
                </div>
                <span class="text-[10px] text-ant-text-muted mt-0.5">
                  Follows macOS / System Sound Settings dynamically
                </span>
              </div>
            </div>
            {#if isDefaultSelected}
              <Check size={14} class="text-ant-primary flex-shrink-0" />
            {/if}
          </button>
        {/if}

        {#if devices.length > 0}
          <div class="h-px bg-ant-border-secondary dark:bg-white/5 my-1"></div>
        {/if}

        {#each devices as device}
          {@const isSelected = selectedDeviceId === device.deviceId || (!selectedDeviceId && device.isDefault)}
          {@const IconComponent = getDeviceIcon(device.transport)}
          <button
            type="button"
            class="w-full text-left p-2 rounded-lg text-xs transition-all flex items-center justify-between group {isSelected
              ? 'bg-ant-primary/15 text-ant-text'
              : 'hover:bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-text'}"
            onclick={() => handleSelect(device.deviceId)}
          >
            <div class="flex items-center space-x-2.5 min-w-0 pr-2">
              <div class="p-1.5 rounded-md bg-ant-bg text-ant-text-secondary group-hover:text-ant-text flex-shrink-0 border border-ant-border-secondary dark:border-white/5">
                <IconComponent size={13} />
              </div>
              <div class="flex flex-col min-w-0">
                <div class="flex items-center space-x-1.5">
                  <span class="font-medium text-xs {isSelected ? 'text-ant-primary' : 'text-ant-text'} truncate">
                    {device.label}
                  </span>
                  {#if device.isDefault}
                    <span class="px-1 py-0.2 text-[9px] font-semibold bg-ant-primary/20 text-ant-primary rounded">
                      Default
                    </span>
                  {/if}
                </div>
                <div class="flex items-center space-x-2 text-[10px] text-ant-text-muted mt-0.5">
                  <span>{getTransportLabel(device.transport)}</span>
                  {#if device.manufacturer}
                    <span>• {device.manufacturer}</span>
                  {/if}
                </div>
              </div>
            </div>

            {#if isSelected}
              <Check size={14} class="text-ant-primary flex-shrink-0" />
            {/if}
          </button>
        {/each}
      </div>
    </div>
  {/if}
</div>
