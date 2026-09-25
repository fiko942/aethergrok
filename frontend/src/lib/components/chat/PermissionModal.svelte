<script lang="ts">
  import type { PermissionRequest } from '$lib/stores/session.svelte';
  import Button from '$lib/antd/Button.svelte';
  import {
    ShieldAlert,
    Terminal,
    FileEdit,
    AlertTriangle,
    X,
    Check,
    CheckCheck,
    Ban
  } from 'lucide-svelte';

  interface Props {
    request: PermissionRequest;
    onDecision: (decision: 'allow_once' | 'allow_always' | 'reject') => void;
    onClose?: () => void;
  }

  let { request, onDecision, onClose }: Props = $props();

  // Extract command or file path details
  const commandOrPath = $derived.by(() => {
    if (!request.details) return null;
    const d = request.details;
    return (
      (d.command as string) ||
      (d.cmd as string) ||
      (d.path as string) ||
      (d.file_path as string) ||
      (d.target_file as string) ||
      null
    );
  });
</script>

<div
  class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-fade-in"
  role="dialog"
  aria-modal="true"
>
  <div
    class="relative w-full max-w-lg bg-ant-bg-secondary border border-ant-border rounded-xl shadow-2xl overflow-hidden flex flex-col"
    onclick={(e) => e.stopPropagation()}
    role="presentation"
  >
    <!-- Modal Header -->
    <div class="flex items-center justify-between px-5 py-3.5 border-b border-ant-border-secondary bg-ant-bg">
      <div class="flex items-center space-x-2.5">
        <div class="w-8 h-8 rounded-lg bg-ant-warning/15 border border-ant-warning/30 flex items-center justify-center text-ant-warning">
          <ShieldAlert size={18} />
        </div>
        <div>
          <h3 class="font-serif-display text-sm font-semibold text-ant-text">Tool Execution Permission</h3>
          <p class="font-serif text-[11px] text-ant-text-muted">AetherGrok Agent requires approval to proceed</p>
        </div>
      </div>
      {#if onClose}
        <button
          type="button"
          onclick={onClose}
          class="p-1 rounded-md text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-tertiary transition"
        >
          <X size={16} />
        </button>
      {/if}
    </div>

    <!-- Modal Content -->
    <div class="p-5 space-y-4 text-xs font-serif">
      <!-- Tool identification chip & description -->
      <div class="space-y-1.5">
        <div class="flex items-center space-x-2">
          <span class="text-ant-text-muted text-[11px] uppercase font-semibold">Target Tool:</span>
          <span class="font-mono text-xs font-semibold px-2 py-0.5 rounded bg-ant-bg text-ant-primary border border-ant-primary/30">
            {request.toolName}
          </span>
        </div>
        <p class="text-ant-text leading-relaxed text-[13px]">
          {request.description || `The agent wants to execute tool "${request.toolName}".`}
        </p>
      </div>

      <!-- Command or path highlight box -->
      {#if commandOrPath}
        <div class="space-y-1">
          <span class="text-ant-text-muted text-[11px] uppercase font-semibold flex items-center gap-1">
            <Terminal size={12} class="text-ant-warning" /> Command / Target Path:
          </span>
          <div class="p-2.5 rounded-lg bg-ant-bg border border-ant-border text-[11px] font-mono text-ant-text overflow-x-auto max-h-32 scrollbar-thin">
            <code>{commandOrPath}</code>
          </div>
        </div>
      {/if}

      <!-- Detailed parameters if present -->
      {#if request.details && Object.keys(request.details).length > 0}
        <div class="space-y-1">
          <span class="text-ant-text-muted text-[11px] uppercase font-semibold">Execution Arguments:</span>
          <pre class="p-2.5 rounded-lg bg-ant-bg border border-ant-border-secondary text-[11px] font-mono text-ant-text-secondary overflow-x-auto max-h-36 scrollbar-thin"><code>{JSON.stringify(request.details, null, 2)}</code></pre>
        </div>
      {/if}

      <!-- Security Notice -->
      <div class="p-2.5 rounded-md bg-ant-warning/10 border border-ant-warning/20 text-ant-warning text-[11px] flex items-start space-x-2">
        <AlertTriangle size={15} class="flex-shrink-0 mt-0.5" />
        <span class="leading-normal">
          Always review shell commands and filesystem write operations before granting permissions.
        </span>
      </div>
    </div>

    <!-- Modal Footer Actions (Allow once, Always allow, Reject) -->
    <div class="flex items-center justify-between px-5 py-3 border-t border-ant-border-secondary bg-ant-bg">
      <Button
        type="default"
        danger
        size="small"
        onclick={() => onDecision('reject')}
        class="flex items-center"
      >
        <Ban size={13} class="mr-1.5" /> Reject
      </Button>

      <div class="flex items-center space-x-2">
        <Button
          type="default"
          size="small"
          onclick={() => onDecision('allow_once')}
          class="flex items-center"
        >
          <Check size={13} class="mr-1.5" /> Allow Once
        </Button>

        <Button
          type="primary"
          size="small"
          onclick={() => onDecision('allow_always')}
          class="flex items-center shadow-md shadow-ant-primary/20"
        >
          <CheckCheck size={13} class="mr-1.5" /> Always Allow
        </Button>
      </div>
    </div>
  </div>
</div>
