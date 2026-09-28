<script lang="ts">
  import { ShieldAlert, Terminal, CheckCircle2, ChevronRight, Apple, AlertTriangle, Key, Copy, Check } from 'lucide-svelte';

  let activeTab = $state<'settings' | 'terminal'>('settings');
  let copiedCmd = $state(false);

  const quarantineCmd = 'xattr -d com.apple.quarantine /Applications/AetherGrok.app';
  const quarantineRecursiveCmd = 'xattr -cr /Applications/AetherGrok.app';

  function copyQuarantine(cmd: string) {
    navigator.clipboard.writeText(cmd);
    copiedCmd = true;
    setTimeout(() => (copiedCmd = false), 2000);
  }
</script>

<section id="gatekeeper" class="py-16 md:py-24 bg-amber-50/40 border-y border-amber-200/80">
  <div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="flex items-center gap-3 mb-6">
      <div class="w-10 h-10 rounded-xl bg-amber-100 text-amber-800 flex items-center justify-center border border-amber-300">
        <Apple class="w-6 h-6" />
      </div>
      <div>
        <span class="text-xs font-bold uppercase tracking-wider text-amber-800 font-mono">
          macOS Security & Gatekeeper Guide
        </span>
        <h2 class="text-2xl sm:text-3xl font-black text-slate-900 tracking-tight">
          Running AetherGrok on macOS (Gatekeeper & Quarantine)
        </h2>
      </div>
    </div>

    <!-- Explanation Box -->
    <div class="bg-white rounded-2xl p-6 sm:p-8 border border-amber-200 shadow-sm mb-8 space-y-4 text-slate-700 leading-relaxed text-sm sm:text-base">
      <div class="flex items-start gap-3">
        <AlertTriangle class="w-5 h-5 text-amber-600 flex-shrink-0 mt-1" />
        <p>
          Because <strong>AetherGrok</strong> is an open-source tool distributed without an annual paid Apple Developer certificate (it is ad-hoc signed), <strong>macOS Gatekeeper</strong> may show a security dialog when you launch it for the first time:
        </p>
      </div>

      <div class="p-3 bg-amber-50 border-l-4 border-amber-500 rounded-r-lg font-mono text-xs text-amber-900">
        “AetherGrok.app cannot be opened because Apple cannot check it for malicious software.”
      </div>

      <p class="text-xs text-slate-600">
        This is standard macOS protection for non-notarized open-source applications downloaded outside the App Store. You can launch it using either method below:
      </p>
    </div>

    <!-- Toggle Selector -->
    <div class="flex gap-2 mb-6">
      <button
        onclick={() => (activeTab = 'settings')}
        class="flex items-center gap-2 px-5 py-2.5 rounded-xl font-bold text-xs sm:text-sm transition-all {activeTab === 'settings'
          ? 'bg-amber-500 text-slate-950 shadow-sm'
          : 'bg-white text-slate-600 hover:text-slate-900 border border-slate-200'}"
      >
        <Key class="w-4 h-4" />
        <span>Option 1: System Settings (GUI)</span>
      </button>

      <button
        onclick={() => (activeTab = 'terminal')}
        class="flex items-center gap-2 px-5 py-2.5 rounded-xl font-bold text-xs sm:text-sm transition-all {activeTab === 'terminal'
          ? 'bg-amber-500 text-slate-950 shadow-sm'
          : 'bg-white text-slate-600 hover:text-slate-900 border border-slate-200'}"
      >
        <Terminal class="w-4 h-4" />
        <span>Option 2: Terminal Command (Fastest)</span>
      </button>
    </div>

    <!-- Content Card -->
    <div class="bg-white rounded-2xl p-6 sm:p-8 border border-slate-200 shadow-sm">
      {#if activeTab === 'settings'}
        <div class="space-y-4">
          <h3 class="text-lg font-bold text-slate-900">
            Approval Steps via macOS System Settings:
          </h3>
          <ol class="space-y-3 text-sm text-slate-700">
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-blue-100 text-blue-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">1</span>
              <span>Open <strong>System Settings</strong> on your Mac.</span>
            </li>
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-blue-100 text-blue-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">2</span>
              <span>Navigate to <strong>Privacy & Security</strong> in the sidebar, then scroll down to the <strong>Security</strong> section.</span>
            </li>
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-blue-100 text-blue-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">3</span>
              <span>You will see the notice: <em>“AetherGrok.app was blocked from use because it is not from an identified developer”</em>.</span>
            </li>
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-blue-100 text-blue-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">4</span>
              <span>Click <strong>Open Anyway</strong>, and confirm with your Mac administrator password or Touch ID.</span>
            </li>
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-emerald-100 text-emerald-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">5</span>
              <span>Click <strong>Open</strong> in the prompt. AetherGrok will now open immediately on all future launches.</span>
            </li>
          </ol>
        </div>
      {:else}
        <div class="space-y-4">
          <h3 class="text-lg font-bold text-slate-900">
            Remove macOS Quarantine Flag via Terminal:
          </h3>
          <p class="text-sm text-slate-600">
            Run this command in Terminal to remove the Gatekeeper quarantine attribute instantly:
          </p>

          <div class="bg-slate-900 rounded-xl p-4 text-slate-200 border border-slate-800 space-y-2">
            <div class="flex items-center justify-between text-xs text-slate-400">
              <span class="font-mono">Terminal Command (Quarantine Removal):</span>
              <button
                onclick={() => copyQuarantine(quarantineCmd)}
                class="text-blue-400 hover:text-blue-300 font-bold flex items-center gap-1"
              >
                {#if copiedCmd}
                  <Check class="w-3.5 h-3.5 text-emerald-400" />
                  <span class="text-emerald-400 font-bold">Copied!</span>
                {:else}
                  <Copy class="w-3.5 h-3.5" />
                  <span>Copy Command</span>
                {/if}
              </button>
            </div>
            <div class="font-mono text-xs sm:text-sm text-emerald-300 select-all overflow-x-auto">
              {quarantineCmd}
            </div>
          </div>

          <p class="text-xs text-slate-500">
            For apps in Downloads or custom directories, run recursively:
            <code class="bg-slate-100 px-1 py-0.5 rounded text-slate-800 font-mono text-[11px] ml-1">{quarantineRecursiveCmd}</code>
          </p>
        </div>
      {/if}
    </div>
  </div>
</section>
