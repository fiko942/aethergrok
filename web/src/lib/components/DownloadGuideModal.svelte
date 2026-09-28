<script lang="ts">
  import { downloadModalStore, closeDownloadModal } from '../data/downloadModal';
  import {
    Download,
    Check,
    Copy,
    Apple,
    Terminal,
    ExternalLink,
    X,
    ShieldAlert,
    HelpCircle,
    ArrowRight
  } from 'lucide-svelte';

  let modalCopied = $state(false);

  const quarantineTerminalCmd = 'xattr -d com.apple.quarantine /Applications/AetherGrok.app';

  function copyQuarantineCommand() {
    navigator.clipboard.writeText(quarantineTerminalCmd);
    modalCopied = true;
    setTimeout(() => {
      modalCopied = false;
    }, 2000);
  }

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      closeDownloadModal();
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      closeDownloadModal();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if $downloadModalStore.isOpen}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-900/60 backdrop-blur-sm animate-fade-in"
    onclick={handleBackdropClick}
    role="presentation"
  >
    <div
      class="bg-white rounded-2xl shadow-2xl border border-slate-200 max-w-lg w-full overflow-hidden transition-all transform animate-scale-up"
      role="dialog"
      aria-modal="true"
    >
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-slate-200 flex items-center justify-between bg-slate-50/80">
        <div class="flex items-center gap-2">
          {#if $downloadModalStore.platform === 'macOS'}
            <div class="w-8 h-8 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center">
              <Apple class="w-4 h-4" />
            </div>
            <div>
              <h3 class="text-base font-bold text-slate-900">macOS Installation Guide</h3>
              <p class="text-xs text-slate-500">Your download has started</p>
            </div>
          {:else}
            <div class="w-8 h-8 rounded-lg bg-blue-100 text-blue-600 flex items-center justify-center">
              <Download class="w-4 h-4" />
            </div>
            <div>
              <h3 class="text-base font-bold text-slate-900">Windows Installation Guide</h3>
              <p class="text-xs text-slate-500">Your download has started</p>
            </div>
          {/if}
        </div>
        <button
          onclick={closeDownloadModal}
          class="p-1.5 text-slate-400 hover:text-slate-700 hover:bg-slate-200/60 rounded-lg transition-colors"
          aria-label="Close modal"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Modal Content -->
      <div class="p-6 space-y-4">
        {#if $downloadModalStore.platform === 'macOS'}
          <!-- macOS Gatekeeper Notice -->
          <div class="p-3.5 rounded-xl bg-amber-50/90 border border-amber-200 text-amber-950 text-xs sm:text-sm flex gap-3 items-start">
            <ShieldAlert class="w-5 h-5 text-amber-600 flex-shrink-0 mt-0.5" />
            <div class="space-y-1">
              <p class="font-bold text-amber-900">"AetherGrok is damaged and can't be opened"?</p>
              <p class="text-amber-800 leading-relaxed text-xs">
                macOS Gatekeeper blocks ad-hoc signed open-source applications by default. Run this 1-line command in Terminal after dragging AetherGrok to <span class="font-mono font-semibold">/Applications</span>:
              </p>
            </div>
          </div>

          <!-- Terminal Command Box -->
          <div class="space-y-1.5">
            <div class="flex items-center justify-between text-xs font-semibold text-slate-700">
              <span class="flex items-center gap-1.5">
                <Terminal class="w-3.5 h-3.5 text-blue-600" />
                Terminal Fix Command:
              </span>
              <span class="text-[11px] text-slate-400 font-normal">Instant bypass</span>
            </div>
            <div class="relative bg-slate-900 rounded-xl p-3 border border-slate-800 text-left font-mono text-xs text-slate-200 flex items-center justify-between gap-3">
              <code class="break-all select-all text-emerald-400 font-medium">
                {quarantineTerminalCmd}
              </code>
              <button
                onclick={copyQuarantineCommand}
                class="flex-shrink-0 inline-flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-white transition-colors"
              >
                {#if modalCopied}
                  <Check class="w-3.5 h-3.5 text-emerald-400" />
                  <span class="text-emerald-400">Copied!</span>
                {:else}
                  <Copy class="w-3.5 h-3.5 text-slate-300" />
                  <span>Copy</span>
                {/if}
              </button>
            </div>
          </div>

          <!-- Alternative GUI Steps -->
          <div class="p-3 bg-slate-50 rounded-xl border border-slate-200 text-xs text-slate-600 space-y-1.5">
            <p class="font-semibold text-slate-800 flex items-center gap-1.5">
              <HelpCircle class="w-3.5 h-3.5 text-slate-500" />
              Alternative via System Settings:
            </p>
            <ol class="list-decimal list-inside space-y-1 pl-1 text-slate-600">
              <li>Open <strong>System Settings &rarr; Privacy & Security</strong></li>
              <li>Scroll down to the <strong>Security</strong> section</li>
              <li>Click <strong>"Open Anyway"</strong> next to AetherGrok</li>
            </ol>
          </div>
        {:else}
          <!-- Windows SmartScreen Notice -->
          <div class="p-4 rounded-xl bg-blue-50/80 border border-blue-200 text-blue-950 text-xs sm:text-sm flex gap-3 items-start">
            <ShieldAlert class="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
            <div class="space-y-1">
              <p class="font-bold text-blue-900">Windows Protected Your PC?</p>
              <p class="text-blue-800 leading-relaxed text-xs">
                If Windows SmartScreen appears during setup, click <span class="font-semibold">"More info"</span> and then select <span class="font-semibold">"Run anyway"</span> to proceed with the installation.
              </p>
            </div>
          </div>
        {/if}

        {#if $downloadModalStore.downloadUrl}
          <div class="pt-2 text-center">
            <a
              href={$downloadModalStore.downloadUrl}
              class="inline-flex items-center gap-1.5 text-xs font-semibold text-blue-600 hover:text-blue-700 hover:underline"
            >
              <span>Download didn't start automatically? Click here to restart</span>
              <ExternalLink class="w-3 h-3" />
            </a>
          </div>
        {/if}
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-3.5 bg-slate-50 border-t border-slate-200 flex items-center justify-between">
        <a
          href="#gatekeeper"
          onclick={closeDownloadModal}
          class="text-xs font-semibold text-slate-600 hover:text-blue-600 flex items-center gap-1"
        >
          <span>View Detailed macOS Guide</span>
          <ArrowRight class="w-3.5 h-3.5" />
        </a>
        <button
          onclick={closeDownloadModal}
          class="px-4 py-2 text-xs font-bold text-white bg-blue-600 hover:bg-blue-700 rounded-lg shadow-sm transition-colors"
        >
          Got it, proceed
        </button>
      </div>
    </div>
  </div>
{/if}
