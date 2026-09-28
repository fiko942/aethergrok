<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import Button from '$lib/antd/Button.svelte';
  import type { AllPermissionsStatus, SystemPermissionItem } from '$app';
  import {
    ShieldCheck,
    ShieldAlert,
    Keyboard,
    Mic,
    Monitor,
    Check,
    RotateCw,
    ExternalLink,
    AlertCircle,
    ArrowRight
  } from 'lucide-svelte';

  interface Props {
    onComplete?: () => void;
  }

  let { onComplete }: Props = $props();

  let loading = $state(true);
  let actionLoading = $state<string | null>(null); // Permission ID currently running action
  let refreshTimer: ReturnType<typeof setInterval> | null = null;

  let status = $state<AllPermissionsStatus>({
    platform: 'darwin',
    allGranted: false,
    items: [
      {
        id: 'accessibility',
        title: 'Accessibility & Global Shortcuts',
        description: 'Required for global snapshot shortcuts, hotkeys, and automated keyboard interactions.',
        granted: false,
        message: 'Accessibility permissions not granted',
        required: true
      },
      {
        id: 'screen_capture',
        title: 'Screen Recording & Window Capture',
        description: 'Allows taking instant smart visual snapshots of desktop windows for multimodal context.',
        granted: false,
        message: 'Screen capture permission not granted',
        required: true
      },
      {
        id: 'microphone',
        title: 'Microphone & Voice Dictation',
        description: 'Enables high-fidelity voice notes, transcription, and speech-to-text queries in chat.',
        granted: false,
        message: 'Microphone access not granted',
        required: true
      }
    ]
  });

  const totalCount = $derived(status.items.length);
  const grantedCount = $derived(status.items.filter((item) => item.granted).length);
  const allGranted = $derived(status.allGranted || (totalCount > 0 && grantedCount === totalCount));
  const progressPercent = $derived(totalCount > 0 ? Math.round((grantedCount / totalCount) * 100) : 0);

  async function checkAllPermissions(silent = false) {
    if (!silent) {
      loading = true;
    }
    try {
      if (window.go?.main?.App?.CheckAllSystemPermissions) {
        const res = await window.go.main.App.CheckAllSystemPermissions();
        if (res && Array.isArray(res.items)) {
          status = res;
        }
      }
    } catch (err) {
      console.error('Failed to query system permissions:', err);
    } finally {
      loading = false;
    }
  }

  async function handleAction(item: SystemPermissionItem) {
    actionLoading = item.id;
    try {
      const app = window.go?.main?.App;
      if (!app) return;

      if (item.id === 'accessibility') {
        if (app.CheckAndRequestAccessibilityPermissions) {
          const res = await app.CheckAndRequestAccessibilityPermissions();
          if (!res.granted && app.OpenAccessibilitySettings) {
            await app.OpenAccessibilitySettings();
          }
        } else if (app.OpenAccessibilitySettings) {
          await app.OpenAccessibilitySettings();
        }
      } else if (item.id === 'screen_capture') {
        if (app.RequestScreenCapturePermission) {
          const res = await app.RequestScreenCapturePermission();
          if (!res.granted && app.OpenScreenCaptureSettings) {
            await app.OpenScreenCaptureSettings();
          }
        } else if (app.OpenScreenCaptureSettings) {
          await app.OpenScreenCaptureSettings();
        }
      } else if (item.id === 'microphone') {
        if (app.RequestMicrophonePermission) {
          const res = await app.RequestMicrophonePermission();
          if (!res.granted && app.OpenMicrophoneSettings) {
            await app.OpenMicrophoneSettings();
          }
        } else if (app.OpenMicrophoneSettings) {
          await app.OpenMicrophoneSettings();
        }
      }

      // Quick re-check status after initiating action
      setTimeout(() => {
        checkAllPermissions(true);
      }, 400);
    } catch (err) {
      console.error(`Error requesting permission for ${item.id}:`, err);
    } finally {
      actionLoading = null;
    }
  }

  function handleWindowFocus() {
    checkAllPermissions(true);
  }

  function handleContinue() {
    if (onComplete) {
      onComplete();
    }
  }

  onMount(() => {
    checkAllPermissions();
    window.addEventListener('focus', handleWindowFocus);

    // Auto-refresh permission status every 3 seconds
    refreshTimer = setInterval(() => {
      checkAllPermissions(true);
    }, 3000);
  });

  onDestroy(() => {
    window.removeEventListener('focus', handleWindowFocus);
    if (refreshTimer) {
      clearInterval(refreshTimer);
      refreshTimer = null;
    }
  });
</script>

<div
  class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-200 select-none"
  role="dialog"
  aria-modal="true"
  aria-labelledby="permission-modal-title"
>
  <div
    class="max-w-xl w-full rounded-2xl border border-ant-border-secondary shadow-2xl bg-ant-bg p-6 space-y-6 flex flex-col"
    onclick={(e) => e.stopPropagation()}
    role="presentation"
  >
    <!-- Header -->
    <div class="flex items-start space-x-4">
      <div
        class="w-12 h-12 rounded-xl flex items-center justify-center shrink-0 transition-colors duration-300 {allGranted
          ? 'bg-emerald-500/15 text-emerald-400 shadow-[0_0_20px_rgba(16,185,129,0.15)]'
          : 'bg-ant-primary/15 text-ant-primary shadow-[0_0_20px_rgba(22,119,255,0.15)]'}"
      >
        {#if allGranted}
          <ShieldCheck size={28} />
        {:else}
          <ShieldAlert size={28} />
        {/if}
      </div>
      <div class="flex-1 min-w-0">
        <h2 id="permission-modal-title" class="font-serif-display text-lg font-bold text-ant-text tracking-tight">
          macOS System Permissions Setup
        </h2>
        <p class="font-serif text-xs text-ant-text-secondary mt-1 leading-relaxed">
          AetherGrok requires standard macOS system permissions to enable keyboard shortcuts, voice dictation, and smart visual captures.
        </p>
      </div>
    </div>

    <!-- Progress Status & Bar -->
    <div class="bg-ant-bg-secondary rounded-xl p-3.5 space-y-2">
      <div class="flex items-center justify-between text-xs font-serif">
        <span class="text-ant-text-secondary font-medium">
          {#if allGranted}
            <span class="text-emerald-400 font-semibold flex items-center gap-1.5">
              <Check size={14} class="stroke-[3]" /> All permissions granted! Ready to start.
            </span>
          {:else}
            <span>{grantedCount} of {totalCount} permissions granted</span>
          {/if}
        </span>
        <span class="font-mono text-xs text-ant-text-muted font-medium">{progressPercent}%</span>
      </div>
      <div class="w-full h-2 rounded-full bg-ant-bg-tertiary overflow-hidden">
        <div
          class="h-full transition-all duration-500 ease-out {allGranted ? 'bg-emerald-500' : 'bg-ant-primary'}"
          style="width: {progressPercent}%"
        ></div>
      </div>
    </div>

    <!-- Permissions List -->
    <div class="space-y-3">
      {#each status.items as item (item.id)}
        <div
          class="flex items-center justify-between p-3.5 rounded-xl transition-all duration-200 {item.granted
            ? 'bg-emerald-500/[0.04]'
            : 'bg-ant-bg-secondary'}"
        >
          <div class="flex items-start space-x-3.5 mr-3 min-w-0">
            <div
              class="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 mt-0.5 {item.granted
                ? 'bg-emerald-500/10 text-emerald-400'
                : 'bg-ant-bg-tertiary text-ant-text-secondary'}"
            >
              {#if item.id === 'accessibility'}
                <Keyboard size={18} />
              {:else if item.id === 'microphone'}
                <Mic size={18} />
              {:else}
                <Monitor size={18} />
              {/if}
            </div>
            <div class="min-w-0">
              <div class="flex items-center space-x-2">
                <span class="text-xs font-semibold text-ant-text font-serif">
                  {item.title}
                </span>
                {#if item.granted}
                  <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-emerald-500/15 text-emerald-400">
                    <Check size={10} class="mr-1 stroke-[3]" /> GRANTED
                  </span>
                {:else}
                  <span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-semibold bg-amber-500/15 text-amber-400">
                    ACTION REQUIRED
                  </span>
                {/if}
              </div>
              <p class="font-serif text-[11.5px] text-ant-text-secondary mt-0.5 leading-relaxed line-clamp-2">
                {item.description}
              </p>
            </div>
          </div>

          <div class="shrink-0">
            {#if item.granted}
              <div class="w-8 h-8 rounded-lg bg-emerald-500/10 flex items-center justify-center text-emerald-400">
                <Check size={16} class="stroke-[2.5]" />
              </div>
            {:else}
              <button
                type="button"
                disabled={actionLoading === item.id}
                onclick={() => handleAction(item)}
                class="px-3 py-1.5 rounded-lg text-xs font-serif font-medium bg-ant-bg-tertiary hover:bg-ant-primary text-ant-text hover:text-white transition flex items-center gap-1.5 cursor-pointer disabled:opacity-50"
              >
                <span>Allow</span>
                <ExternalLink size={12} class="opacity-70" />
              </button>
            {/if}
          </div>
        </div>
      {/each}
    </div>

    <!-- Missing Permissions Notice (if not all granted) -->
    {#if !allGranted}
      <div class="flex items-start space-x-2 text-[11px] text-ant-text-muted bg-ant-bg-secondary/60 rounded-lg p-2.5">
        <AlertCircle size={14} class="text-amber-400 shrink-0 mt-0.5" />
        <p class="leading-relaxed">
          Status updates automatically every 3 seconds or when you switch back from macOS System Settings.
        </p>
      </div>
    {/if}

    <!-- Footer Actions -->
    <div class="flex items-center justify-between pt-2 border-t border-ant-border-secondary">
      <Button
        type="text"
        size="middle"
        loading={loading}
        onclick={() => checkAllPermissions()}
        class="text-xs text-ant-text-secondary hover:text-ant-text flex items-center gap-1.5"
      >
        <RotateCw size={13} class={loading ? 'animate-spin' : ''} />
        <span>Re-check Status</span>
      </Button>

      <div class="flex items-center space-x-2.5">
        {#if !allGranted}
          <Button
            type="text"
            size="middle"
            onclick={handleContinue}
            class="text-xs text-ant-text-muted hover:text-ant-text font-serif"
          >
            Skip for now
          </Button>
        {/if}
        <Button
          type="primary"
          size="middle"
          onclick={handleContinue}
          class="font-serif text-xs font-semibold px-4 flex items-center gap-1.5 {allGranted ? 'shadow-[0_0_15px_rgba(16,185,129,0.35)] !bg-emerald-600 hover:!bg-emerald-500 active:!bg-emerald-700 !border-emerald-600' : ''}"
        >
          <span>Continue to AetherGrok</span>
          <ArrowRight size={13} />
        </Button>
      </div>
    </div>
  </div>
</div>
