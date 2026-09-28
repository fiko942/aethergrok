<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Heart, X, ExternalLink, Sparkles, Coffee, ShieldCheck, Copy, Check } from 'lucide-svelte';

  let {
    open = false,
    onClose = () => {}
  }: {
    open: boolean;
    onClose: () => void;
  } = $props();

  let copiedLink = $state(false);

  function openExternal(url: string) {
    if (window.runtime?.BrowserOpenURL) {
      window.runtime.BrowserOpenURL(url);
    } else {
      window.open(url, '_blank', 'noopener,noreferrer');
    }
  }

  function copySaweriaLink() {
    navigator.clipboard.writeText('https://saweria.co/wijifikoteren');
    copiedLink = true;
    setTimeout(() => (copiedLink = false), 2000);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === 'Escape') {
      e.preventDefault();
      e.stopPropagation();
      onClose();
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
  <!-- Backdrop -->
  <div
    class="fixed inset-0 z-[9999] bg-black/75 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-150 select-none font-serif"
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-labelledby="donate-modal-title"
    onclick={onClose}
    onkeydown={(e) => { if (e.key === 'Escape') onClose(); }}
  >
    <!-- Modal Dialog Card -->
    <div
      role="document"
      tabindex="-1"
      class="w-full max-w-md bg-ant-bg-secondary border border-ant-border-secondary dark:border-white/10 rounded-2xl shadow-2xl p-6 space-y-5 animate-in zoom-in-95 duration-150 outline-none text-ant-text relative overflow-hidden"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <!-- Top Subtle Background Glow -->
      <div class="absolute -top-12 -right-12 w-36 h-36 bg-rose-500/10 rounded-full blur-2xl pointer-events-none"></div>
      <div class="absolute -bottom-12 -left-12 w-36 h-36 bg-amber-500/10 rounded-full blur-2xl pointer-events-none"></div>

      <!-- Header -->
      <div class="flex items-start justify-between relative z-10">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-tr from-rose-500/20 to-amber-500/20 border border-rose-500/30 flex items-center justify-center text-rose-500 shadow-xs">
            <Heart size={20} class="fill-rose-500/30" />
          </div>
          <div>
            <h3 id="donate-modal-title" class="font-serif-display text-base font-bold text-ant-text tracking-tight flex items-center gap-1.5">
              <span>Support AetherGrok</span>
              <Sparkles size={14} class="text-amber-500" />
            </h3>
            <p class="text-xs text-ant-text-secondary mt-0.5">
              Independent Open-Source Development
            </p>
          </div>
        </div>

        <button
          type="button"
          onclick={onClose}
          class="w-7 h-7 flex items-center justify-center rounded-lg text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-tertiary transition-colors cursor-pointer"
          aria-label="Close dialog"
        >
          <X size={15} />
        </button>
      </div>

      <!-- Main Explanation -->
      <div class="text-xs text-ant-text-secondary leading-relaxed space-y-2 relative z-10">
        <p>
          AetherGrok is built and maintained as free, open-source software for the engineering community. If it boosts your daily productivity, consider supporting its maintenance!
        </p>
      </div>

      <!-- Primary Saweria Donation Card -->
      <div class="p-4 rounded-xl bg-gradient-to-br from-amber-500/15 via-amber-500/5 to-transparent border border-amber-500/30 space-y-3 shadow-xs relative z-10">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <Coffee size={16} class="text-amber-500" />
            <span class="text-xs font-bold text-ant-text">Saweria (QRIS & E-Wallet)</span>
          </div>
          <span class="px-2 py-0.5 rounded text-[10px] font-sans font-bold bg-amber-500/20 text-amber-600 dark:text-amber-400">
            Instant QRIS
          </span>
        </div>

        <p class="text-[11.5px] text-ant-text-secondary leading-normal">
          Supports instant donations via GoPay, OVO, Dana, ShopeePay, LinkAja, and any Indonesian bank QRIS app.
        </p>

        <div class="pt-1 flex items-center gap-2">
          <button
            type="button"
            onclick={() => openExternal('https://saweria.co/wijifikoteren')}
            class="flex-1 py-2 px-3 rounded-lg bg-amber-500 hover:bg-amber-400 active:scale-[0.99] text-slate-950 font-sans font-bold text-xs flex items-center justify-center gap-1.5 transition shadow-sm cursor-pointer"
          >
            <span>Open Saweria Gateway</span>
            <ExternalLink size={12} />
          </button>

          <button
            type="button"
            onclick={copySaweriaLink}
            class="py-2 px-2.5 rounded-lg bg-ant-bg hover:bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-text border border-ant-border-secondary transition text-xs flex items-center gap-1 cursor-pointer"
            title="Copy donation URL"
          >
            {#if copiedLink}
              <Check size={13} class="text-emerald-500" />
              <span class="text-[11px] font-sans text-emerald-500">Copied</span>
            {:else}
              <Copy size={13} />
              <span class="text-[11px] font-sans">Copy</span>
            {/if}
          </button>
        </div>
      </div>

      <!-- Secondary GitHub Sponsors / Open Source Card -->
      <div class="p-3.5 rounded-xl bg-ant-bg border border-ant-border-secondary flex items-center justify-between relative z-10">
        <div class="space-y-0.5">
          <div class="text-xs font-semibold text-ant-text flex items-center gap-1.5">
            <ShieldCheck size={14} class="text-blue-500" />
            <span>GitHub Sponsors & Star</span>
          </div>
          <p class="text-[10.5px] text-ant-text-muted">
            Star the repository or contribute via GitHub.
          </p>
        </div>

        <button
          type="button"
          onclick={() => openExternal('https://github.com/fiko942/aethergrok')}
          class="px-3 py-1.5 rounded-lg bg-ant-bg-tertiary hover:bg-ant-primary/10 hover:text-ant-primary text-ant-text-secondary text-xs font-semibold transition border border-ant-border-secondary flex items-center gap-1.5 cursor-pointer"
        >
          <span>GitHub</span>
          <ExternalLink size={11} />
        </button>
      </div>

      <!-- Footer Action -->
      <div class="flex justify-end pt-1 relative z-10">
        <button
          type="button"
          onclick={onClose}
          class="px-4 py-1.5 rounded-lg text-xs font-medium text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition cursor-pointer"
        >
          Close
        </button>
      </div>
    </div>
  </div>
{/if}
