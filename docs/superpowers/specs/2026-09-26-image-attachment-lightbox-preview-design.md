# Specification: Image Attachment Lightbox Preview in Composer

**Date:** 2026-09-26  
**Status:** Approved by User  
**Target Components:**
- `frontend/src/lib/components/chat/Composer.svelte`
- `frontend/src/lib/components/chat/composer/AttachmentChip.svelte`
- `frontend/src/lib/components/chat/ImageLightboxModal.svelte`

---

## 1. Overview & Problem Statement

Users attaching images to the Composer prompt box (via file upload, clipboard paste, or screen snapshot) see an attachment chip thumbnail inside the input container. Clicking the image thumbnail currently has no visible effect because the `previewModalItem` state in `Composer.svelte` is not linked to any modal renderer.

This specification defines the integration of the full-featured, semi-transparent fullscreen `ImageLightboxModal` directly into the Composer workflow, providing instant enlarged previews with zoom, copy, and download actions.

---

## 2. User Experience & Interactions

1. **Thumbnail Click Trigger:**
   - When a user clicks on an image thumbnail within an `AttachmentChip` in the Composer, `ImageLightboxModal` opens immediately.
   - The modal presents the enlarged image centered over a darkened, blurred backdrop (`bg-black/90 backdrop-blur-md z-50`).

2. **Modal Capabilities:**
   - **Zoom Controls:** Zoom in (+25%), zoom out (-25%), and reset zoom (100%).
   - **Export Controls:** Copy image to clipboard and download image file locally.
   - **Keyboard & Click Navigation:** Pressing `Escape` or clicking anywhere on the dimmed background closes the lightbox.
   - **Header Bar:** Shows image name title and current zoom percentage.

3. **Non-Image Attachment Scope:**
   - Non-image files (such as `.md`, `.json`, `.py`, `.txt`, `.pdf`) retain their colored type badge and metadata.
   - Non-image icon buttons do not trigger the image lightbox, ensuring clear affordance focused on image previews.

---

## 3. Architecture & State Management

### 3.1 `Composer.svelte`
- **Import:** Import `ImageLightboxModal` from `./ImageLightboxModal.svelte`.
- **State Definition:**
  ```typescript
  let lightboxVisible = $state(false);
  let lightboxSrc = $state('');
  let lightboxTitle = $state('');
  ```
- **Handler Method:**
  ```typescript
  function handleAttachmentPreview(att: AttachedFile) {
    if (att.isImage && att.dataUrl) {
      lightboxSrc = att.dataUrl;
      lightboxTitle = att.name;
      lightboxVisible = true;
    }
  }
  ```
- **Template Integration:**
  Pass `handleAttachmentPreview` to `<AttachmentChip onPreview={handleAttachmentPreview} />`.
  Mount `<ImageLightboxModal />` at the bottom of `Composer.svelte` bound to the reactive states.

### 3.2 `AttachmentChip.svelte`
- Refine the thumbnail button hover effects with clear tooltip (`"Preview image"`).
- For non-image files, display the typed file icon cleanly without an active preview action.

---

## 4. Verification & Testing Strategy

1. **Type & Component Check:** Run `pnpm --prefix frontend run check` to ensure zero Svelte 5 and TypeScript errors.
2. **Interactive UI Verification:**
   - Attach an image file through file picker, clipboard paste, and screenshot snapshot.
   - Click the image thumbnail in the composer chip and verify the lightbox renders centered with dark backdrop.
   - Test zoom in, zoom out, reset zoom, and ESC key dismissal.
   - Verify non-image files do not trigger the image modal.
