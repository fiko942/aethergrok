<script lang="ts">
  import {
    Heart,
    Coffee,
    Sparkles,
    Wallet,
    Copy,
    Check,
    ExternalLink,
    QrCode,
    Zap
  } from 'lucide-svelte';
  import { githubRepo } from '../data/downloads';

  const saweriaUrl = 'https://saweria.co/wijifikoteren';
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
      <div class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-semibold bg-amber-50 text-amber-800 border border-amber-200 mb-3">
        <Sparkles class="w-3.5 h-3.5 text-amber-600" />
        <span>Dukungan & Sponsorship Terbuka</span>
      </div>
      <h2 class="text-3xl sm:text-4xl font-extrabold text-slate-900 tracking-tight">
        Dukung Pengembangan AetherGrok
      </h2>
      <p class="mt-4 text-base text-slate-600 leading-relaxed">
        AetherGrok dibangun dan dikembangkan secara independen sebagai proyek open source berperforma tinggi. Dukungan Anda langsung mendanai pemeliharaan native binary, sertifikasi, serta pengembangan fitur selanjutnya.
      </p>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-8 max-w-4xl mx-auto">
      <!-- Saweria Card (Primary Donation Gateway) -->
      <div class="bg-gradient-to-br from-amber-500/10 via-white to-amber-50/30 rounded-2xl p-8 border-2 border-amber-400/80 shadow-md hover:shadow-lg transition-all flex flex-col justify-between relative overflow-hidden">
        <div class="absolute -top-3 right-6 px-3 py-0.5 rounded-full text-[11px] font-bold uppercase tracking-wider bg-amber-500 text-slate-950 shadow-sm flex items-center gap-1">
          <Zap class="w-3 h-3 fill-slate-950" />
          <span>Direkomendasikan (Indonesia)</span>
        </div>

        <div>
          <div class="w-14 h-14 rounded-2xl bg-amber-100 text-amber-700 flex items-center justify-center mb-6 shadow-sm border border-amber-200">
            <Coffee class="w-7 h-7" />
          </div>

          <h3 class="text-2xl font-black text-slate-900 mb-2">
            Saweria (QRIS, GoPay, OVO, Dana)
          </h3>

          <p class="text-slate-600 text-sm leading-relaxed mb-6">
            Dukung langsung melalui portal Saweria dengan metode pembayaran instan: QRIS untuk semua bank dan e-wallet (GoPay, OVO, Dana, LinkAja, ShopeePay).
          </p>

          <div class="bg-white/90 border border-amber-200 rounded-xl p-3 mb-6 flex items-center justify-between">
            <span class="text-xs font-mono font-medium text-slate-700">saweria.co/wijifikoteren</span>
            <span class="text-[11px] font-semibold text-amber-700 bg-amber-50 px-2 py-0.5 rounded">Instan & Terverifikasi</span>
          </div>
        </div>

        <a
          href="{saweriaUrl}"
          target="_blank"
          rel="noopener noreferrer"
          class="w-full inline-flex items-center justify-center gap-2.5 px-6 py-3.5 rounded-xl font-bold text-base text-slate-950 bg-amber-400 hover:bg-amber-300 active:scale-[0.99] shadow-md hover:shadow-lg transition-all"
        >
          <Coffee class="w-5 h-5 fill-slate-950" />
          <span>Donasi via Saweria</span>
          <ExternalLink class="w-4 h-4 ml-1" />
        </a>
      </div>

      <!-- GitHub Sponsors & Global Community Card -->
      <div class="bg-white rounded-2xl p-8 border border-slate-200 shadow-sm flex flex-col justify-between hover:border-slate-300 transition-all">
        <div>
          <div class="w-14 h-14 rounded-2xl bg-rose-50 text-rose-600 flex items-center justify-center mb-6 shadow-sm border border-rose-100">
            <Heart class="w-7 h-7 fill-rose-500" />
          </div>

          <h3 class="text-2xl font-black text-slate-900 mb-2">
            GitHub Sponsors & International
          </h3>

          <p class="text-slate-600 text-sm leading-relaxed mb-6">
            For international supporters and organizations. Become a monthly backer or make a one-time contribution via GitHub Sponsors.
          </p>

          <div class="space-y-2 mb-6">
            <div class="bg-slate-50 border border-slate-200 rounded-lg p-2.5 flex items-center justify-between text-xs">
              <span class="text-slate-500 font-mono">github.com/sponsors/fiko942</span>
              <span class="text-emerald-700 font-medium">Monthly / 1-time</span>
            </div>
          </div>
        </div>

        <div class="space-y-3">
          <a
            href="https://github.com/sponsors/fiko942"
            target="_blank"
            rel="noopener noreferrer"
            class="w-full inline-flex items-center justify-center gap-2 px-6 py-3.5 rounded-xl font-bold text-sm text-white bg-rose-600 hover:bg-rose-700 shadow-sm hover:shadow transition-all"
          >
            <Heart class="w-4 h-4 fill-white" />
            <span>Sponsor on GitHub</span>
            <ExternalLink class="w-4 h-4" />
          </a>

          <a
            href="{githubRepo}"
            target="_blank"
            rel="noopener noreferrer"
            class="w-full inline-flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl font-medium text-xs text-slate-600 hover:text-slate-900 bg-slate-50 hover:bg-slate-100 transition-colors"
          >
            <span>Star on GitHub Repository</span>
            <ExternalLink class="w-3.5 h-3.5" />
          </a>
        </div>
      </div>
    </div>
  </div>
</section>
