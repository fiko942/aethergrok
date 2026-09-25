<script lang="ts">
  import type { AttachedFile } from '$lib/stores/session.svelte';
  import { resolveFileIcon } from '$lib/utils/fileIcons';
  import { X, ZoomIn } from 'lucide-svelte';

  interface Props {
    attachment: AttachedFile;
    onRemove: (id: string) => void;
    onPreview?: (attachment: AttachedFile) => void;
  }

  let { attachment, onRemove, onPreview }: Props = $props();

  function formatBytes(bytes?: number): string {
    if (!bytes || bytes === 0) return '';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  const iconMeta = $derived(resolveFileIcon(attachment.name, attachment.isImage));
  const IconComponent = $derived(iconMeta.component);
</script>

<div
  class="group relative inline-flex items-center gap-2 bg-white/[0.04] hover:bg-white/[0.07] border border-white/[0.08] hover:border-white/[0.14] rounded-lg pl-1.5 pr-2 py-1.5 text-xs text-ant-text transition-colors duration-150 select-none max-w-[220px]"
>
  {#if attachment.isImage && attachment.dataUrl}
    <!-- Thumbnail for Image -->
    <button
      type="button"
      onclick={() => onPreview?.(attachment)}
      class="relative w-7 h-7 rounded overflow-hidden bg-black/40 flex-shrink-0 cursor-pointer flex items-center justify-center border border-white/5 transition hover:opacity-90"
      title="View Image"
    >
      <img src={attachment.dataUrl} alt={attachment.name} class="w-full h-full object-cover" />
      <div class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity">
        <ZoomIn size={11} class="text-white drop-shadow" />
      </div>
    </button>
  {:else}
    <!-- Dynamic Typed File Icon Badge -->
    <button
      type="button"
      onclick={() => onPreview?.(attachment)}
      class="w-7 h-7 rounded {iconMeta.bgClass} flex items-center justify-center {iconMeta.colorClass} flex-shrink-0 cursor-pointer transition-transform group-hover:scale-105"
      title="Inspect File"
    >
      <IconComponent size={14} />
    </button>
  {/if}

  <div class="flex flex-col justify-center min-w-0 flex-1 pr-0.5">
    <div class="flex items-center gap-1 min-w-0">
      <span class="text-[11.5px] font-mono font-medium text-ant-text truncate leading-tight" title={attachment.name}>
        {attachment.name}
      </span>
    </div>
    <div class="flex items-center gap-1.5 mt-0.5">
      <span class="text-[8.5px] font-mono px-1 py-0.2 rounded bg-white/[0.06] text-ant-text-muted uppercase leading-none">
        {iconMeta.badge}
      </span>
      {#if attachment.sizeBytes}
        <span class="text-[9px] text-ant-text-muted font-mono leading-none">
          {formatBytes(attachment.sizeBytes)}
        </span>
      {/if}
    </div>
  </div>

  <!-- Remove Button -->
  <button
    type="button"
    onclick={() => onRemove(attachment.id)}
    class="w-4 h-4 rounded hover:bg-white/10 flex items-center justify-center text-ant-text-muted hover:text-rose-400 transition flex-shrink-0"
    title="Remove attachment"
  >
    <X size={11} />
  </button>
</div>
