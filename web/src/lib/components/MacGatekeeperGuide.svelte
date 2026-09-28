<script lang="ts">
  import { ShieldAlert, Terminal, CheckCircle2, ChevronRight, Apple, AlertTriangle, Key } from 'lucide-svelte';
  import { onMount } from 'svelte';

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
          Panduan Pengguna macOS
        </span>
        <h2 class="text-2xl sm:text-3xl font-black text-slate-900 tracking-tight">
          Cara Membuka Aplikasi di macOS (Gatekeeper & Quarantine)
        </h2>
      </div>
    </div>

    <!-- Explanation Box -->
    <div class="bg-white rounded-2xl p-6 sm:p-8 border border-amber-200 shadow-sm mb-8 space-y-4 text-slate-700 leading-relaxed text-sm sm:text-base">
      <div class="flex items-start gap-3">
        <AlertTriangle class="w-5 h-5 text-amber-600 flex-shrink-0 mt-1" />
        <p>
          Karena <strong>AetherGrok</strong> merupakan proyek open-source independen yang didistribusikan tanpa sertifikat berbayar Apple Developer ID (ditandatangani secara <em>ad-hoc</em>), fitur keamanan <strong>macOS Gatekeeper</strong> mungkin akan menampilkan pesan peringatan:
        </p>
      </div>

      <div class="p-3 bg-amber-50 border-l-4 border-amber-500 rounded-r-lg font-mono text-xs text-amber-900">
        “AetherGrok.app tidak dapat dibuka karena Apple tidak dapat memeriksa perangkat lunak berbahaya dari pengembang yang belum terverifikasi.”
      </div>

      <p class="text-xs text-slate-600">
        Ini adalah perilaku standar macOS untuk semua aplikasi open source gratis yang diunduh di luar App Store. Anda dapat membukanya dengan mudah menggunakan salah satu dari 2 cara di bawah ini:
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
        <span>Cara 1: Lewat System Settings (UI)</span>
      </button>

      <button
        onclick={() => (activeTab = 'terminal')}
        class="flex items-center gap-2 px-5 py-2.5 rounded-xl font-bold text-xs sm:text-sm transition-all {activeTab === 'terminal'
          ? 'bg-amber-500 text-slate-950 shadow-sm'
          : 'bg-white text-slate-600 hover:text-slate-900 border border-slate-200'}"
      >
        <Terminal class="w-4 h-4" />
        <span>Cara 2: Lewat Terminal (1 Detik)</span>
      </button>
    </div>

    <!-- Content Card -->
    <div class="bg-white rounded-2xl p-6 sm:p-8 border border-slate-200 shadow-sm">
      {#if activeTab === 'settings'}
        <div class="space-y-4">
          <h3 class="text-lg font-bold text-slate-900">
            Langkah Persetujuan melalui Pengaturan Sistem (System Settings):
          </h3>
          <ol class="space-y-3 text-sm text-slate-700">
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-blue-100 text-blue-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">1</span>
              <span>Buka <strong>System Settings</strong> (Pengaturan Sistem) pada Mac Anda.</span>
            </li>
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-blue-100 text-blue-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">2</span>
              <span>Pilih menu <strong>Privacy & Security</strong> (Privasi & Keamanan) dan gulir ke bagian <strong>Security</strong> (Keamanan).</span>
            </li>
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-blue-100 text-blue-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">3</span>
              <span>Akan muncul keterangan: <em>“AetherGrok.app diblokir agar tidak digunakan karena bukan dari pengembang yang teridentifikasi”</em>.</span>
            </li>
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-blue-100 text-blue-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">4</span>
              <span>Klik tombol <strong>Open Anyway</strong> (Tetap Buka), lalu masukkan password / Touch ID Mac Anda.</span>
            </li>
            <li class="flex items-start gap-3">
              <span class="w-6 h-6 rounded-full bg-emerald-100 text-emerald-800 font-bold text-xs flex items-center justify-center flex-shrink-0 mt-0.5">5</span>
              <span>Klik <strong>Open</strong> pada jendela konfirmasi. Aplikasi kini siap digunakan untuk seterusnya!</span>
            </li>
          </ol>
        </div>
      {:else}
        <div class="space-y-4">
          <h3 class="text-lg font-bold text-slate-900">
            Hapus Atribut Karantina macOS via Terminal:
          </h3>
          <p class="text-sm text-slate-600">
            Jalankan perintah berikut di Terminal untuk menghapus atribut karantina Gatekeeper pada AetherGrok secara instan:
          </p>

          <div class="bg-slate-900 rounded-xl p-4 text-slate-200 border border-slate-800 space-y-2">
            <div class="flex items-center justify-between text-xs text-slate-400">
              <span class="font-mono">Terminal Command (Quarantine Removal):</span>
              <button
                onclick={() => copyQuarantine(quarantineCmd)}
                class="text-blue-400 hover:text-blue-300 font-bold"
              >
                {copiedCmd ? 'Tersalin!' : 'Salin Perintah'}
              </button>
            </div>
            <div class="font-mono text-xs sm:text-sm text-emerald-300 select-all overflow-x-auto">
              {quarantineCmd}
            </div>
          </div>

          <p class="text-xs text-slate-500">
            Atau jika aplikasi masih berada di folder Downloads / direktori kustom:
            <code class="bg-slate-100 px-1 py-0.5 rounded text-slate-800 font-mono text-[11px] ml-1">{quarantineRecursiveCmd}</code>
          </p>
        </div>
      {/if}
    </div>
  </div>
</section>
