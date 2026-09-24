<script lang="ts">
  import type { VisionImage } from '$lib/stores/session.svelte';
  import { Image as ImageIcon, X, ZoomIn, FileText } from 'lucide-svelte';

  interface Props {
    images: VisionImage[];
    onRemove: (id: string) => void;
    onClearAll?: () => void;
  }

  let { images = [], onRemove, onClearAll }: Props = $props();

  let previewModalImage = $state<VisionImage | null>(null);

  function formatBytes(bytes?: number): string {
    if (!bytes || bytes === 0) return 'Vision Chip';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  }

  function getFileName(filePath: string): string {
    if (!filePath) return 'Snapshot';
    const parts = filePath.split(/[/\\]/);
    return parts[parts.length - 1] || filePath;
  }
</script>

{#if images.length > 0}
  <div class="flex items-center gap-2 p-2 bg-ant-bg border-b border-ant-border-secondary overflow-x-auto select-none">
    <div class="flex items-center text-[10px] font-semibold tracking-wider text-ant-text-muted uppercase px-1 flex-shrink-0">
      <ImageIcon size={12} class="mr-1 text-ant-primary" /> Vision Context ({images.length})
    </div>

    <div class="flex items-center gap-2 flex-1 min-w-0">
      {#each images as img (img.id)}
        <div
          class="group relative flex items-center bg-ant-bg-secondary hover:bg-ant-bg-tertiary border border-ant-border rounded-lg pl-1.5 pr-2 py-1 text-xs text-ant-text transition-all duration-150 flex-shrink-0 shadow-sm"
        >
          <!-- Thumbnail Image -->
          <button
            type="button"
            onclick={() => previewModalImage = img}
            class="relative w-7 h-7 rounded overflow-hidden border border-ant-border-secondary bg-black/40 flex-shrink-0 cursor-pointer flex items-center justify-center group-hover:border-ant-primary/50 transition"
            title="Click to view full preview"
          >
            {#if img.dataUrl}
              <img
                src={img.dataUrl}
                alt={getFileName(img.filePath)}
                class="w-full h-full object-cover"
              />
            {:else}
              <FileText size={14} class="text-ant-text-muted" />
            {/if}
            <div class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity">
              <ZoomIn size={12} class="text-white drop-shadow" />
            </div>
          </button>

          <!-- Image Info -->
          <div class="ml-2 flex flex-col justify-center min-w-[70px] max-w-[130px]">
            <span class="text-[11px] font-medium text-white truncate" title={img.filePath}>
              {getFileName(img.filePath)}
            </span>
            <span class="text-[9px] text-ant-text-muted font-mono">
              {formatBytes(img.sizeBytes)}
            </span>
          </div>

          <!-- Quick Remove Button -->
          <button
            type="button"
            onclick={(e) => {
              e.stopPropagation();
              onRemove(img.id);
            }}
            class="ml-1.5 p-1 text-ant-text-muted hover:text-ant-error hover:bg-ant-error/10 rounded transition flex-shrink-0"
            title="Remove attachment"
          >
            <X size={12} />
          </button>
        </div>
      {/each}
    </div>

    {#if onClearAll && images.length > 1}
      <button
        type="button"
        onclick={onClearAll}
        class="text-[10px] text-ant-text-muted hover:text-ant-error px-2 py-1 rounded hover:bg-ant-bg-secondary transition flex-shrink-0"
      >
        Clear all
      </button>
    {/if}
  </div>
{/if}

<!-- Full Size Vision Modal Preview -->
{#if previewModalImage}
  <div
    class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4"
    onclick={() => previewModalImage = null}
    role="presentation"
  >
    <div
      class="bg-ant-bg-secondary border border-ant-border rounded-xl shadow-2xl max-w-4xl max-h-[85vh] flex flex-col overflow-hidden"
      onclick={(e) => e.stopPropagation()}
      role="presentation"
    >
      <div class="flex items-center justify-between px-4 py-2.5 border-b border-ant-border-secondary bg-ant-bg">
        <div class="flex items-center space-x-2 text-xs font-semibold text-white">
          <ImageIcon size={14} class="text-ant-primary" />
          <span class="truncate">{getFileName(previewModalImage.filePath)}</span>
          <span class="text-ant-text-muted text-[10px] font-mono">({formatBytes(previewModalImage.sizeBytes)})</span>
        </div>
        <button
          type="button"
          onclick={() => previewModalImage = null}
          class="p-1 rounded-md text-ant-text-muted hover:text-white hover:bg-ant-bg-secondary transition"
        >
          <X size={15} />
        </button>
      </div>
      <div class="p-4 flex items-center justify-center overflow-auto max-h-[75vh] bg-ant-bg">
        <img
          src={previewModalImage.dataUrl}
          alt="Vision attachment preview"
          class="max-h-full max-w-full rounded border border-ant-border object-contain"
        />
      </div>
    </div>
  </div>
{/if}
