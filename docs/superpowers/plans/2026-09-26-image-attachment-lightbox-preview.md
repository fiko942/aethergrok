# Image Attachment Lightbox Preview Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Enable instant enlarged image lightbox pop-up preview with zoom, copy, and download capabilities when clicking attached image thumbnails in the Composer prompt box.

**Architecture:** Integrate the existing `ImageLightboxModal.svelte` into `Composer.svelte`, wire the `onPreview` event callback from `AttachmentChip.svelte` to open the modal with the image's `dataUrl` and title, and refine non-image attachment chip affordances.

**Tech Stack:** Svelte 5 (Runes `$state`, `$derived`, `$props`), Tailwind CSS, Lucide Svelte, TypeScript.

## Global Constraints

- Use Svelte 5 Runes exclusively (`$state`, `$derived`, `$props`).
- Modal must support Esc key dismissal, backdrop click closing, zoom in/out, and reset zoom.
- Non-image files must not trigger the image lightbox modal.
- Zero TypeScript and Svelte check compilation errors (`pnpm --prefix frontend run check`).

---

### Task 1: Refine `AttachmentChip.svelte` Preview Affordances

**Files:**
- Modify: `frontend/src/lib/components/chat/composer/AttachmentChip.svelte`

**Interfaces:**
- Consumes: `attachment: AttachedFile`, `onRemove: (id: string) => void`, `onPreview?: (attachment: AttachedFile) => void`
- Produces: Visual zoom affordance for images with active click trigger, and clean non-interactive icon container for non-image files.

- [ ] **Step 1: Update `AttachmentChip.svelte` to specialize image thumbnail trigger**

Modify `frontend/src/lib/components/chat/composer/AttachmentChip.svelte`:
```svelte
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
    <!-- Thumbnail for Image with Clickable Preview -->
    <button
      type="button"
      onclick={() => onPreview?.(attachment)}
      class="relative w-7 h-7 rounded overflow-hidden bg-black/40 flex-shrink-0 cursor-pointer flex items-center justify-center border border-white/5 transition hover:opacity-90 group/thumb"
      title="Preview image"
    >
      <img src={attachment.dataUrl} alt={attachment.name} class="w-full h-full object-cover" />
      <div class="absolute inset-0 bg-black/40 opacity-0 group-hover/thumb:opacity-100 flex items-center justify-center transition-opacity">
        <ZoomIn size={12} class="text-white drop-shadow" />
      </div>
    </button>
  {:else}
    <!-- Static Typed File Icon Badge (Non-Image) -->
    <div
      class="w-7 h-7 rounded {iconMeta.bgClass} flex items-center justify-center {iconMeta.colorClass} flex-shrink-0 select-none"
      title={attachment.name}
    >
      <IconComponent size={14} />
    </div>
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
```

- [ ] **Step 2: Verify Svelte check passes**

Run: `pnpm --prefix frontend run check`
Expected: PASS with 0 errors.

- [ ] **Step 3: Commit Task 1 changes**

```bash
git add frontend/src/lib/components/chat/composer/AttachmentChip.svelte
git commit -m "feat(composer): refine attachment chip preview trigger for images"
```

---

### Task 2: Integrate `ImageLightboxModal` in `Composer.svelte`

**Files:**
- Modify: `frontend/src/lib/components/chat/Composer.svelte`

**Interfaces:**
- Consumes: `ImageLightboxModal.svelte`, `AttachedFile`
- Produces: Lightbox overlay rendering when an attached image is clicked.

- [ ] **Step 1: Update `Composer.svelte` with Lightbox import, state, handler, and modal mount**

1. Import `ImageLightboxModal`:
```typescript
import ImageLightboxModal from './ImageLightboxModal.svelte';
```

2. Add reactive state and preview handler:
```typescript
let lightboxVisible = $state(false);
let lightboxSrc = $state('');
let lightboxTitle = $state('');

function handleAttachmentPreview(att: AttachedFile) {
  if (att.isImage && att.dataUrl) {
    lightboxSrc = att.dataUrl;
    lightboxTitle = att.name;
    lightboxVisible = true;
  }
}
```

3. Update `<AttachmentChip>` invocation:
```svelte
{#each attachedFiles as att (att.id)}
  <AttachmentChip
    attachment={att}
    onRemove={removeAttachment}
    onPreview={handleAttachmentPreview}
  />
{/each}
```

4. Mount `<ImageLightboxModal>` at the bottom of the template:
```svelte
<!-- Fullscreen Image Lightbox Modal for Attachments -->
<ImageLightboxModal
  visible={lightboxVisible}
  imageSrc={lightboxSrc}
  imageTitle={lightboxTitle}
  onClose={() => { lightboxVisible = false; }}
/>
```

- [ ] **Step 2: Run typecheck and frontend build**

Run: `pnpm --prefix frontend run check && pnpm --prefix frontend run build`
Expected: PASS with 0 errors.

- [ ] **Step 3: Commit Task 2 changes**

```bash
git add frontend/src/lib/components/chat/Composer.svelte
git commit -m "feat(composer): mount ImageLightboxModal for attachment preview"
```

---

### Task 3: Verification & Edge Case Validation

**Files:**
- Verify across all frontend components

- [ ] **Step 1: Verify Go and Frontend Tests**

Run: `go test ./test/... -v` and `pnpm --prefix frontend run check`
Expected: All tests PASS.

- [ ] **Step 2: Commit any cleanups if needed**
