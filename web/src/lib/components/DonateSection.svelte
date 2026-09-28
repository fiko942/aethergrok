<script lang="ts">
  import { Heart, Coffee, Wallet, Copy, Check, ExternalLink } from 'lucide-svelte';
  import { githubRepo } from '../data/downloads';

  let copiedBtc = $state(false);
  let copiedEth = $state(false);

  const btcAddress = 'bc1q9fiko42exampledevelopmentaddressxyz';
  const ethAddress = '0x42fiko942developmentaddressforethereumxyz';

  function copyText(text: string, type: 'btc' | 'eth') {
    navigator.clipboard.writeText(text);
    if (type === 'btc') {
      copiedBtc = true;
      setTimeout(() => (copiedBtc = false), 2000);
    } else {
      copiedEth = true;
      setTimeout(() => (copiedEth = false), 2000);
    }
  }
</script>

<section id="sponsors" class="py-20 md:py-28 bg-slate-50/70 border-t border-slate-200/80">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
    <div class="text-center max-w-3xl mx-auto mb-14">
      <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-rose-50 text-rose-700 border border-rose-200 mb-3">
        <Heart class="w-3.5 h-3.5 fill-rose-500 text-rose-500" />
        <span>Support Independent Open Source</span>
      </div>
      <h2 class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
        Sponsor AetherGrok Development
      </h2>
      <p class="mt-4 text-base text-slate-600">
        AetherGrok is built and maintained as a high-performance open-source project. Your sponsorship directly funds native platform maintenance, automated test infrastructure, and continuous feature releases.
      </p>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl mx-auto">
      <!-- GitHub Sponsors & Community Card -->
      <div class="bg-white rounded-2xl p-8 border border-slate-200 shadow-sm flex flex-col justify-between">
        <div>
          <div class="w-12 h-12 rounded-xl bg-rose-50 text-rose-600 flex items-center justify-center mb-5">
            <Heart class="w-6 h-6 fill-rose-500" />
          </div>
          <h3 class="text-xl font-bold text-slate-900 mb-2">
            GitHub Sponsors & Backers
          </h3>
          <p class="text-slate-600 text-sm leading-relaxed mb-6">
            Become a monthly or one-time backer on GitHub Sponsors to help support ongoing engineering, macOS notarization, and continuous updates.
          </p>
        </div>

        <a
          href="https://github.com/sponsors/fiko942"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-xl font-bold text-sm text-white bg-rose-600 hover:bg-rose-700 shadow-sm hover:shadow transition-all"
        >
          <Heart class="w-4 h-4 fill-white" />
          <span>Sponsor on GitHub</span>
          <ExternalLink class="w-3.5 h-3.5" />
        </a>
      </div>

      <!-- Crypto & Direct Donation Card -->
      <div class="bg-white rounded-2xl p-8 border border-slate-200 shadow-sm flex flex-col justify-between">
        <div>
          <div class="w-12 h-12 rounded-xl bg-amber-50 text-amber-600 flex items-center justify-center mb-5">
            <Wallet class="w-6 h-6" />
          </div>
          <h3 class="text-xl font-bold text-slate-900 mb-2">
            Direct & Crypto Donations
          </h3>
          <p class="text-slate-600 text-sm leading-relaxed mb-4">
            Support developer tool maintenance directly via Bitcoin or Ethereum addresses.
          </p>

          <div class="space-y-3 pt-2 mb-6">
            <div class="bg-slate-50 border border-slate-200 rounded-lg p-2.5 flex items-center justify-between">
              <div class="truncate mr-2">
                <span class="text-[10px] font-bold uppercase tracking-wider text-slate-400 block">ETH / ERC20</span>
                <span class="font-mono text-xs text-slate-700 truncate">{ethAddress}</span>
              </div>
              <button
                onclick={() => copyText(ethAddress, 'eth')}
                class="p-1.5 text-slate-400 hover:text-slate-700 rounded bg-white border border-slate-200 shadow-2xs"
                title="Copy Ethereum Address"
              >
                {#if copiedEth}
                  <Check class="w-4 h-4 text-emerald-600" />
                {:else}
                  <Copy class="w-4 h-4" />
                {/if}
              </button>
            </div>
          </div>
        </div>

        <a
          href="{githubRepo}"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-xl font-bold text-sm text-slate-700 bg-slate-100 hover:bg-slate-200 transition-colors"
        >
          <span>Star on GitHub Repository</span>
          <ExternalLink class="w-3.5 h-3.5" />
        </a>
      </div>
    </div>
  </div>
</section>
