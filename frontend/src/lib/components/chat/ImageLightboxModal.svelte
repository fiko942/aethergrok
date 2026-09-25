<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { X, ZoomIn, ZoomOut, RotateCcw, Copy, Check, ExternalLink, Download } from 'lucide-svelte';

  interface Props {
    visible: boolean;
    imageSrc: string;
    imageTitle?: string;
    onClose: () => void;
  }

  let { visible = false, imageSrc, imageTitle = 'Image Preview', onClose }: Props = $props();

  let scale = $state(1);
  let copied = $state(false);

  function handleKeyDown(e: KeyboardEvent) {
    if (!visible) return;
    if (e.key === 'Escape') {
      onClose();
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeyDown);
  });

  function zoomIn() {
    scale = Math.min(scale + 0.25, 4);
  }

  function zoomOut() {
    scale = Math.max(scale - 0.25, 0.5);
  }

  function resetZoom() {
    scale = 1;
  }

  async function copyImage() {
    try {
      if (imageSrc.startsWith('data:image/')) {
        const res = await fetch(imageSrc);
        const blob = await res.blob();
        await navigator.clipboard.write([
          new ClipboardItem({ [blob.type]: blob })
        ]);
      } else {
        await navigator.clipboard.writeText(imageSrc);
      }
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch (err) {
      console.error('Failed to copy image to clipboard:', err);
    }
  }

  function downloadImage() {
    const a = document.createElement('a');
    a.href = imageSrc;
    a.download = imageTitle || 'download.png';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
  }
</script>

{#if visible && imageSrc}
  <!-- Fullscreen Modal Container -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex flex-col bg-black/90 backdrop-blur-md animate-in fade-in duration-150 select-none"
    onclick={onClose}
    onkeydown={(e) => e.key === 'Escape' && onClose()}
  >
    <!-- Top Action Bar -->
    <div
      class="flex items-center justify-between px-6 py-3.5 bg-black/40 border-b border-white/10 z-10"
      onclick={(e) => e.stopPropagation()}
    >
      <div class="flex items-center space-x-3">
        <span class="text-sm font-serif font-medium text-ant-text truncate max-w-md">
          {imageTitle}
        </span>
        <span class="text-[11px] font-mono text-ant-text-muted px-2 py-0.5 rounded bg-white/5 border border-white/5">
          {Math.round(scale * 100)}%
        </span>
      </div>

      <div class="flex items-center space-x-2">
        <button
          type="button"
          onclick={zoomOut}
          class="p-1.5 rounded-lg text-ant-text-muted hover:text-ant-text hover:bg-white/10 transition"
          title="Zoom out"
        >
          <ZoomOut size={16} />
        </button>

        <button
          type="button"
          onclick={zoomIn}
          class="p-1.5 rounded-lg text-ant-text-muted hover:text-ant-text hover:bg-white/10 transition"
          title="Zoom in"
        >
          <ZoomIn size={16} />
        </button>

        <button
          type="button"
          onclick={resetZoom}
          class="p-1.5 rounded-lg text-ant-text-muted hover:text-ant-text hover:bg-white/10 transition"
          title="Reset zoom"
        >
          <RotateCcw size={15} />
        </button>

        <div class="w-px h-4 bg-white/10 mx-1"></div>

        <button
          type="button"
          onclick={copyImage}
          class="p-1.5 rounded-lg text-ant-text-muted hover:text-ant-text hover:bg-white/10 transition flex items-center space-x-1.5 text-xs font-serif"
          title="Copy image"
        >
          {#if copied}
            <Check size={15} class="text-green-400" />
            <span class="text-green-400">Copied</span>
          {:else}
            <Copy size={15} />
            <span>Copy</span>
          {/if}
        </button>

        <button
          type="button"
          onclick={downloadImage}
          class="p-1.5 rounded-lg text-ant-text-muted hover:text-ant-text hover:bg-white/10 transition"
          title="Download image"
        >
          <Download size={16} />
        </button>

        <div class="w-px h-4 bg-white/10 mx-1"></div>

        <button
          type="button"
          onclick={onClose}
          class="p-1.5 rounded-lg text-ant-text-muted hover:text-white hover:bg-white/10 transition"
          title="Close (Esc)"
          aria-label="Close"
        >
          <X size={18} />
        </button>
      </div>
    </div>

    <!-- Center Stage Image Canvas -->
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <div
      class="flex-1 flex items-center justify-center p-6 overflow-auto cursor-zoom-out"
      onclick={onClose}
    >
      <!-- Image Wrapper with pan/scale transform -->
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <img
        src={imageSrc}
        alt={imageTitle}
        onclick={(e) => e.stopPropagation()}
        style="transform: scale({scale}); transition: transform 0.15s ease-out;"
        class="max-w-full max-h-[85vh] object-contain rounded-lg shadow-2xl border border-white/10 cursor-default select-none"
      />
    </div>
  </div>
{/if}
