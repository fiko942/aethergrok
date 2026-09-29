<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import type { GrokInstallStatus, GrokInstallProgress } from '../../../app';
  import {
    Sparkles,
    Download,
    Terminal,
    CheckCircle2,
    AlertCircle,
    RotateCw,
    Copy,
    Check,
    ChevronDown,
    ChevronUp,
    ExternalLink,
    Loader2,
    Cpu,
    ArrowRight
  } from 'lucide-svelte';

  interface Props {
    onComplete?: () => void;
  }

  let { onComplete }: Props = $props();

  type InstallState = 'not_installed' | 'installing' | 'success' | 'error';

  let installState = $state<InstallState>('not_installed');
  let status = $state<GrokInstallStatus | null>(null);
  let progress = $state<GrokInstallProgress>({
    stage: 'preparing',
    percent: 0,
    message: 'Preparing Grok CLI installation...',
    logLine: ''
  });
  let logs = $state<string[]>([]);
  let showLogs = $state<boolean>(false);
  let copied = $state<boolean>(false);
  let logContainer = $state<HTMLDivElement | null>(null);
  let unlistenProgress: (() => void) | null = null;

  function scrollToBottom() {
    if (logContainer) {
      logContainer.scrollTop = logContainer.scrollHeight;
    }
  }

  async function checkStatus() {
    try {
      if (window.go?.main?.App?.CheckGrokInstallation) {
        const res = await window.go.main.App.CheckGrokInstallation();
        status = res;
        if (res.installed) {
          installState = 'success';
        }
      }
    } catch (err) {
      console.error('Failed to check Grok installation status:', err);
    }
  }

  async function startInstallation() {
    installState = 'installing';
    logs = ['[INFO] Initiating Grok CLI automated installer...'];
    progress = {
      stage: 'preparing',
      percent: 5,
      message: 'Initializing installer package...',
      logLine: 'Initializing installer package...'
    };

    try {
      if (window.go?.main?.App?.InstallGrokCLI) {
        const res = await window.go.main.App.InstallGrokCLI();
        status = res;
        if (res.installed) {
          installState = 'success';
          progress = {
            stage: 'completed',
            percent: 100,
            message: 'Grok CLI installed successfully!',
            logLine: `Verified binary at ${res.binaryPath}`
          };
          logs = [...logs, `[SUCCESS] Grok CLI ${res.version || ''} verified at ${res.binaryPath}`];
        } else {
          installState = 'error';
          const errMsg = res.error || 'Installation encountered an unknown issue.';
          logs = [...logs, `[ERROR] ${errMsg}`];
        }
      } else {
        // Fallback for mock/test environments
        installState = 'error';
        const errMsg = 'InstallGrokCLI backend binding is unavailable.';
        logs = [...logs, `[ERROR] ${errMsg}`];
      }
    } catch (err: any) {
      installState = 'error';
      const errMsg = err?.message || String(err) || 'Failed to execute Grok installer.';
      logs = [...logs, `[ERROR] ${errMsg}`];
    }
  }

  function handleCopyLogs() {
    const textToCopy = [
      `Grok CLI Installation Diagnostic Logs`,
      `Platform: ${status?.platform || navigator.platform}`,
      `State: ${installState}`,
      `Progress: ${progress.percent}% - ${progress.message}`,
      `Error: ${status?.error || 'N/A'}`,
      `--- Logs ---`,
      ...logs
    ].join('\n');

    navigator.clipboard.writeText(textToCopy);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  }

  function handleComplete() {
    if (onComplete) {
      onComplete();
    }
  }

  onMount(() => {
    checkStatus();

    if (window.runtime?.EventsOn) {
      unlistenProgress = window.runtime.EventsOn('grok:install_progress', (data: GrokInstallProgress) => {
        if (data) {
          progress = data;
          if (data.logLine) {
            logs = [...logs, data.logLine];
            setTimeout(scrollToBottom, 50);
          }
          if (data.stage === 'completed') {
            installState = 'success';
          } else if (data.stage === 'error') {
            installState = 'error';
          } else if (installState !== 'installing') {
            installState = 'installing';
          }
        }
      });
    }
  });

  onDestroy(() => {
    if (unlistenProgress) {
      unlistenProgress();
      unlistenProgress = null;
    }
  });
</script>

<div
  class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-200 select-none"
  role="dialog"
  aria-modal="true"
  aria-labelledby="grok-install-modal-title"
  tabindex="-1"
>
  <div
    class="max-w-xl w-full rounded-2xl border border-ant-border-secondary dark:border-white/5 shadow-2xl bg-ant-bg p-6 space-y-6 flex flex-col"
    onclick={(e) => e.stopPropagation()}
    role="presentation"
  >
    <!-- Header -->
    <div class="flex items-start space-x-4">
      <div
        class="w-12 h-12 rounded-xl flex items-center justify-center shrink-0 transition-colors duration-300 {installState === 'success'
          ? 'bg-emerald-500/15 text-emerald-400 shadow-[0_0_20px_rgba(16,185,129,0.15)]'
          : installState === 'error'
          ? 'bg-rose-500/15 text-rose-400 shadow-[0_0_20px_rgba(244,63,94,0.15)]'
          : 'bg-ant-primary/15 text-ant-primary shadow-[0_0_20px_rgba(22,119,255,0.15)]'}"
      >
        {#if installState === 'success'}
          <CheckCircle2 size={28} />
        {:else if installState === 'error'}
          <AlertCircle size={28} />
        {:else if installState === 'installing'}
          <Loader2 size={28} class="animate-spin" />
        {:else}
          <Cpu size={28} />
        {/if}
      </div>

      <div class="flex-1 min-w-0">
        <h2 id="grok-install-modal-title" class="font-serif-display text-lg font-bold text-ant-text tracking-tight">
          {#if installState === 'success'}
            Grok CLI Ready
          {:else if installState === 'installing'}
            Installing Grok CLI
          {:else if installState === 'error'}
            Installation Encountered an Issue
          {:else}
            Grok CLI Required
          {/if}
        </h2>
        <p class="font-serif text-xs text-ant-text-secondary mt-1 leading-relaxed">
          {#if installState === 'success'}
            Grok CLI is verified and connected. AetherGrok Desktop Studio is ready to run sessions, execute tools, and automate workflows.
          {:else if installState === 'installing'}
            Setting up the official Grok CLI runtime engine for desktop reasoning and agent tool execution.
          {:else if installState === 'error'}
            We encountered a problem while attempting to download or verify the Grok CLI binary on your system.
          {:else}
            AetherGrok Desktop Studio relies on the official Grok CLI to power fast multi-model reasoning, interactive coding agents, and terminal tool execution.
          {/if}
        </p>
      </div>
    </div>

    <!-- Main Content Area based on State -->
    {#if installState === 'not_installed'}
      <div class="bg-ant-bg-secondary rounded-xl p-4 space-y-3.5 border border-ant-border-secondary dark:border-white/5">
        <div class="space-y-2 text-xs font-serif text-ant-text">
          <div class="flex items-center gap-2 font-medium text-ant-text">
            <Sparkles size={15} class="text-ant-primary shrink-0" />
            <span>Automated 1-Click Setup</span>
          </div>
          <p class="text-ant-text-secondary leading-relaxed pl-5.5">
            Clicking install will automatically fetch the latest official Grok binary, verify its checksum, configure the execution path, and link it directly to your AetherGrok workspace.
          </p>
        </div>

        <div class="pt-2 border-t border-ant-border-secondary dark:border-white/5 flex items-center justify-between text-[11px] text-ant-text-muted font-mono">
          <span>Platform: {status?.platform || 'Auto-detecting...'}</span>
          <span>Target: ~/.grok/bin/grok</span>
        </div>
      </div>

      <!-- Actions -->
      <div class="flex items-center justify-end space-x-3 pt-2">
        <button
          type="button"
          onclick={startInstallation}
          class="px-5 py-2.5 rounded-xl text-xs font-serif font-semibold bg-ant-primary hover:bg-ant-primary-hover text-white shadow-md shadow-ant-primary/25 active:scale-95 transition-all flex items-center gap-2 cursor-pointer"
        >
          <Download size={15} />
          <span>Install Grok CLI</span>
        </button>
      </div>
    {:else if installState === 'installing'}
      <!-- Progress Bar & Status -->
      <div class="bg-ant-bg-secondary rounded-xl p-4 space-y-3 border border-ant-border-secondary dark:border-white/5">
        <div class="flex items-center justify-between text-xs font-serif">
          <span class="text-ant-text font-medium flex items-center gap-2">
            <Loader2 size={13} class="animate-spin text-ant-primary" />
            {progress.message || 'Processing installation...'}
          </span>
          <span class="font-mono font-semibold text-ant-primary">{progress.percent}%</span>
        </div>

        <div class="w-full h-2 rounded-full bg-ant-bg-tertiary overflow-hidden">
          <div
            class="h-full bg-ant-primary transition-all duration-300 ease-out"
            style="width: {Math.max(5, Math.min(100, progress.percent))}%"
          ></div>
        </div>

        <!-- Collapsible / Live Terminal Logs -->
        <div class="pt-1">
          <button
            type="button"
            onclick={() => (showLogs = !showLogs)}
            class="flex items-center gap-1.5 text-[11px] font-mono text-ant-text-secondary hover:text-ant-text transition cursor-pointer"
          >
            {#if showLogs}
              <ChevronUp size={13} />
              <span>Hide installation log</span>
            {:else}
              <ChevronDown size={13} />
              <span>Show live terminal output ({logs.length} lines)</span>
            {/if}
          </button>

          {#if showLogs}
            <div
              bind:this={logContainer}
              class="mt-2.5 p-3 rounded-lg bg-ant-bg border border-ant-border-secondary dark:border-white/5 font-mono text-[11px] leading-relaxed text-ant-text max-h-48 overflow-y-auto space-y-1 select-text"
            >
              {#each logs as log}
                <div class="break-all whitespace-pre-wrap {log.startsWith('[ERROR]') ? 'text-rose-400' : log.startsWith('[SUCCESS]') ? 'text-emerald-400' : 'text-ant-text-secondary'}">
                  {log}
                </div>
              {/each}
            </div>
          {/if}
        </div>
      </div>
    {:else if installState === 'success'}
      <div class="bg-ant-bg-secondary rounded-xl p-4 space-y-3 border border-ant-border-secondary dark:border-white/5">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <span class="text-xs font-serif text-ant-text-secondary">Detected Binary:</span>
            <span class="text-xs font-mono font-semibold text-ant-text px-2 py-0.5 rounded bg-ant-bg-tertiary border border-ant-border-secondary dark:border-white/5">
              {status?.version || 'v1.0.0+'}
            </span>
          </div>
          <span class="text-[11px] font-mono text-emerald-400 font-medium flex items-center gap-1">
            <Check size={13} class="stroke-[3]" /> Ready
          </span>
        </div>

        {#if status?.binaryPath}
          <div class="p-2.5 rounded-lg bg-ant-bg border border-ant-border-secondary dark:border-white/5 font-mono text-[11px] text-ant-text-secondary break-all">
            {status.binaryPath}
          </div>
        {/if}
      </div>

      <!-- Actions -->
      <div class="flex items-center justify-end space-x-3 pt-2">
        <button
          type="button"
          onclick={handleComplete}
          class="px-5 py-2.5 rounded-xl text-xs font-serif font-semibold bg-ant-primary hover:bg-ant-primary-hover text-white shadow-md shadow-ant-primary/25 active:scale-95 transition-all flex items-center gap-2 cursor-pointer"
        >
          <span>Launch AetherGrok</span>
          <ArrowRight size={15} />
        </button>
      </div>
    {:else if installState === 'error'}
      <div class="bg-ant-bg-secondary rounded-xl p-4 space-y-3 border border-rose-500/20">
        <div class="text-xs font-serif text-rose-400 font-medium">
          {status?.error || progress.message || 'An error occurred during installation.'}
        </div>

        <!-- Terminal Output -->
        <div
          bind:this={logContainer}
          class="p-3 rounded-lg bg-ant-bg border border-ant-border-secondary dark:border-white/5 font-mono text-[11px] leading-relaxed text-ant-text max-h-40 overflow-y-auto space-y-1 select-text"
        >
          {#each logs as log}
            <div class="break-all whitespace-pre-wrap {log.startsWith('[ERROR]') ? 'text-rose-400' : 'text-ant-text-secondary'}">
              {log}
            </div>
          {/each}
        </div>
      </div>

      <!-- Error Actions -->
      <div class="flex items-center justify-between pt-2">
        <button
          type="button"
          onclick={handleCopyLogs}
          class="px-3.5 py-2 rounded-xl text-xs font-serif font-medium bg-ant-bg-secondary hover:bg-ant-bg-tertiary text-ant-text-secondary hover:text-ant-text border border-ant-border-secondary dark:border-white/5 active:scale-95 transition-all flex items-center gap-1.5 cursor-pointer"
        >
          {#if copied}
            <Check size={14} class="text-emerald-400" />
            <span class="text-emerald-400">Copied to Clipboard</span>
          {:else}
            <Copy size={14} />
            <span>Copy Diagnostic Logs</span>
          {/if}
        </button>

        <button
          type="button"
          onclick={startInstallation}
          class="px-4 py-2 rounded-xl text-xs font-serif font-semibold bg-ant-primary hover:bg-ant-primary-hover text-white shadow-md shadow-ant-primary/25 active:scale-95 transition-all flex items-center gap-2 cursor-pointer"
        >
          <RotateCw size={14} />
          <span>Retry Installation</span>
        </button>
      </div>
    {/if}
  </div>
</div>
