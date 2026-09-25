<script lang="ts">
  import type { AttachedFile } from '$lib/stores/session.svelte';
  import { FileText, FileCode, File, Image as ImageIcon, X, ZoomIn, Eye } from 'lucide-svelte';

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

  function getFileIcon(name: string, isImage: boolean) {
    if (isImage) return ImageIcon;
    const ext = name.split('.').pop()?.toLowerCase();
    if (['ts', 'js', 'json', 'py', 'go', 'rs', 'html', 'css', 'svelte'].includes(ext || '')) {
      return FileCode;
    }
    if (['md', 'markdown', 'txt'].includes(ext || '')) {
      return FileText;
    }
    return File;
  }

  const IconComponent = $derived(getFileIcon(attachment.name, attachment.isImage));
</script>

<div
  class="group relative inline-flex items-center gap-1.5 bg-ant-bg-secondary/90 hover:bg-ant-bg-tertiary border border-white/10 hover:border-white/20 rounded-lg pl-1.5 pr-2 py-1 text-xs text-ant-text transition-all duration-150 shadow-sm select-none max-w-[200px]"
>
  {#if attachment.isImage && attachment.dataUrl}
    <!-- Thumbnail for Image -->
    <button
      type="button"
      onclick={() => onPreview?.(attachment)}
      class="relative w-6 h-6 rounded overflow-hidden border border-white/10 bg-black/40 flex-shrink-0 cursor-pointer flex items-center justify-center group-hover:border-ant-primary/50 transition"
      title="View Image"
    >
      <img src={attachment.dataUrl} alt={attachment.name} class="w-full h-full object-cover" />
      <div class="absolute inset-0 bg-black/30 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity">
        <ZoomIn size={11} class="text-white drop-shadow" />
      </div>
    </button>
  {:else}
    <!-- Icon for Document/Code -->
    <button
      type="button"
      onclick={() => onPreview?.(attachment)}
      class="w-6 h-6 rounded bg-ant-primary/10 border border-ant-primary/20 flex items-center justify-center text-ant-primary flex-shrink-0 cursor-pointer"
      title="Inspect File"
    >
      <IconComponent size={13} />
    </button>
  {/if}

  <div class="flex flex-col justify-center min-w-0 flex-1 pr-1">
    <span class="text-[11.5px] font-mono font-medium text-ant-text truncate leading-tight" title={attachment.name}>
      {attachment.name}
    </span>
    {#if attachment.sizeBytes}
      <span class="text-[9px] text-ant-text-muted font-mono leading-none mt-0.5">
        {formatBytes(attachment.sizeBytes)}
      </span>
    {/if}
  </div>

  <!-- Remove Button -->
  <button
    type="button"
    onclick={() => onRemove(attachment.id)}
    class="w-4 h-4 rounded hover:bg-white/10 flex items-center justify-center text-ant-text-muted hover:text-ant-text transition flex-shrink-0"
    title="Remove attachment"
  >
    <X size={11} />
  </button>
</div>
