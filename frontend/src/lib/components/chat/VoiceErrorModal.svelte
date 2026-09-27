<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { AlertCircle, X, Copy, Check, Terminal } from 'lucide-svelte';

  interface Props {
    visible: boolean;
    errorMessage: string;
    errorDetails?: string;
    onClose: () => void;
  }

  let { visible = false, errorMessage, errorDetails = '', onClose }: Props = $props();

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

  async function handleCopyLog() {
    const fullLog = errorDetails ? `${errorMessage}\n\n${errorDetails}` : errorMessage;
    try {
      await navigator.clipboard.writeText(fullLog);
      copied = true;
      setTimeout(() => (copied = false), 2000);
    } catch (err) {
      console.error('Failed to copy error log:', err);
    }
  }
</script>

{#if visible}
  <!-- Backdrop -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fade-in"
    onclick={onClose}
    role="dialog"
    aria-modal="true"
  >
    <!-- Modal Card -->
    <div
      class="relative w-full max-w-lg bg-ant-bg-secondary border border-ant-border-secondary dark:border-white/5 rounded-xl shadow-2xl overflow-hidden flex flex-col animate-scale-in"
      onclick={(e) => e.stopPropagation()}
      role="presentation"
    >
      <!-- Modal Header -->
      <div class="flex items-center justify-between px-5 py-3.5 border-b border-ant-border-secondary bg-ant-bg">
        <div class="flex items-center space-x-2.5">
          <div class="w-8 h-8 rounded-lg bg-rose-500/15 border border-rose-500/30 flex items-center justify-center text-rose-400">
            <AlertCircle size={18} />
          </div>
          <div>
            <h3 class="font-serif-display text-sm font-semibold text-ant-text">Voice Transcription Error</h3>
            <p class="font-serif text-[11px] text-ant-text-muted">Transcription failed to complete</p>
          </div>
        </div>
        <button
          type="button"
          onclick={onClose}
          class="p-1.5 text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-hover rounded-lg transition"
          title="Close dialog"
        >
          <X size={16} />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-5 space-y-4 max-h-[65vh] overflow-y-auto font-serif">
        <div class="space-y-1.5">
          <p class="text-xs text-ant-text-secondary leading-relaxed">
            {errorMessage || 'An error occurred while processing the audio transcription.'}
          </p>
        </div>

        {#if errorDetails}
          <div class="space-y-1.5">
            <div class="flex items-center justify-between">
              <span class="text-[11px] font-mono font-medium text-ant-text-muted flex items-center gap-1.5">
                <Terminal size={12} class="text-ant-text-muted" />
                Error Diagnostics & Logs:
              </span>
              <button
                type="button"
                onclick={handleCopyLog}
                class="flex items-center space-x-1 text-[11px] font-mono text-ant-primary hover:text-ant-primary-hover px-1.5 py-0.5 rounded hover:bg-ant-primary/10 transition"
              >
                {#if copied}
                  <Check size={12} class="text-emerald-400" />
                  <span class="text-emerald-400">Copied!</span>
                {:else}
                  <Copy size={12} />
                  <span>Copy Log</span>
                {/if}
              </button>
            </div>

            <div class="p-3 bg-ant-bg border border-ant-border-secondary dark:border-white/5 rounded-lg text-[11px] font-mono text-rose-300/90 whitespace-pre-wrap break-words max-h-48 overflow-y-auto scrollbar-thin select-text leading-relaxed">
              {errorDetails}
            </div>
          </div>
        {/if}
      </div>

      <!-- Modal Footer -->
      <div class="flex items-center justify-end px-5 py-3 border-t border-ant-border-secondary bg-ant-bg space-x-2">
        {#if errorDetails}
          <button
            type="button"
            onclick={handleCopyLog}
            class="px-3 py-1.5 text-xs font-serif rounded-lg border border-ant-border-secondary dark:border-white/5 text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-hover transition flex items-center space-x-1.5"
          >
            {#if copied}
              <Check size={13} class="text-emerald-400" />
              <span>Copied</span>
            {:else}
              <Copy size={13} />
              <span>Copy Error Log</span>
            {/if}
          </button>
        {/if}
        <button
          type="button"
          onclick={onClose}
          class="px-4 py-1.5 text-xs font-serif font-medium rounded-lg bg-ant-primary text-white hover:bg-ant-primary-hover transition"
        >
          Close
        </button>
      </div>
    </div>
  </div>
{/if}
