<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { X } from 'lucide-svelte';

  let {
    open = false,
    title = 'Confirm Action',
    content = 'Are you sure you want to proceed?',
    confirmText = 'Confirm',
    cancelText = 'Cancel',
    type = 'danger',
    onConfirm = () => {},
    onCancel = () => {}
  }: {
    open: boolean;
    title?: string;
    content?: string;
    confirmText?: string;
    cancelText?: string;
    type?: 'danger' | 'warning' | 'info';
    onConfirm: () => void;
    onCancel: () => void;
  } = $props();

  function handleKeydown(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      onCancel();
    } else if (e.key === 'Enter') {
      e.preventDefault();
      e.stopPropagation();
      onConfirm();
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeydown, true);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeydown, true);
  });
</script>

{#if open}
  <!-- Global Root Backdrop -->
  <div
    class="fixed inset-0 z-[99999] bg-black/80 backdrop-blur-md flex items-center justify-center p-4 animate-in fade-in duration-150 select-none"
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-labelledby="confirm-modal-title"
    onclick={onCancel}
    onkeydown={(e) => { if (e.key === 'Escape') onCancel(); }}
  >
    <!-- Modal Card -->
    <div
      role="document"
      tabindex="-1"
      class="w-full max-w-sm bg-ant-bg-secondary border border-white/10 rounded-2xl shadow-2xl p-6 space-y-4 animate-in zoom-in-95 duration-150 outline-none"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <div class="flex items-start justify-between">
        <div class="space-y-1.5 flex-1 pr-3">
          <h3 id="confirm-modal-title" class="font-serif-display text-base font-semibold text-ant-text tracking-tight">
            {title}
          </h3>
          <p class="font-serif text-xs text-ant-text-secondary leading-relaxed">
            {content}
          </p>
        </div>

        <button
          type="button"
          onclick={onCancel}
          class="p-1 text-ant-text-muted hover:text-ant-text rounded-lg hover:bg-ant-bg-tertiary transition"
          aria-label="Close dialog"
        >
          <X size={15} />
        </button>
      </div>

      <!-- Action Buttons with Distinct Separation -->
      <div class="pt-3 border-t border-white/5 flex items-center justify-end space-x-2.5">
        <!-- Cancel: Subtle text button with transparent background -->
        <button
          type="button"
          onclick={onCancel}
          class="px-3.5 py-1.5 rounded-lg text-xs font-serif text-ant-text-secondary hover:text-ant-text hover:bg-white/5 border border-transparent transition"
        >
          {cancelText}
        </button>

        <!-- Destructive Action: Prominent solid danger button -->
        <button
          type="button"
          onclick={onConfirm}
          class="px-4 py-1.5 rounded-lg text-xs font-serif font-medium text-white shadow-sm transition {
            type === 'danger'
              ? 'bg-rose-600 hover:bg-rose-500 active:bg-rose-700 shadow-rose-950/40'
              : 'bg-ant-primary hover:bg-ant-primary/90 active:bg-ant-primary/80'
          }"
        >
          {confirmText}
        </button>
      </div>
    </div>
  </div>
{/if}
