<script lang="ts">
  import { onMount } from 'svelte';
  import { settingsStore } from '$lib/stores/settings.svelte';
  import { inputShieldStore } from '$lib/stores/inputShield.svelte';
  import { Check, X, MessageSquare, ListTodo } from 'lucide-svelte';

  interface Props {
    planTitle?: string;
    onAction: (action: 'approve' | 'reject' | 'custom', feedback?: string) => void;
  }

  let { planTitle = 'Execution Plan Review', onAction }: Props = $props();

  let showCustomInput = $state(false);
  let customFeedback = $state('');

  onMount(() => {
    // If Plan Gate Bypass mode is enabled in Settings, automatically approve and implement
    if (settingsStore.planGateMode === 'bypass') {
      setTimeout(() => {
        onAction('approve');
      }, 300);
    }
  });

  function handleApprove() {
    onAction('approve');
  }

  function handleReject() {
    onAction('reject');
  }

  function handleSubmitCustom() {
    if (customFeedback.trim()) {
      onAction('custom', customFeedback.trim());
      showCustomInput = false;
      customFeedback = '';
    }
  }
</script>

<div class="my-3 rounded-xl border border-ant-border-secondary dark:border-white/[0.08] bg-ant-bg-secondary dark:bg-[#18181b]/95 p-4 shadow-xl backdrop-blur-md space-y-3 font-serif">
  <!-- Card Header -->
  <div class="flex items-center justify-between pb-2 border-b border-ant-border-secondary dark:border-white/5">
    <div class="flex items-center space-x-2.5">
      <div class="w-7 h-7 rounded-lg bg-indigo-500/15 text-indigo-600 dark:text-indigo-400 flex items-center justify-center flex-shrink-0">
        <ListTodo size={16} />
      </div>
      <div>
        <h4 class="font-serif-display text-sm font-semibold text-ant-text tracking-tight flex items-center gap-1.5">
          <span>{planTitle}</span>
          {#if settingsStore.planGateMode === 'bypass'}
            <span class="text-[10px] font-mono px-1.5 py-0.2 rounded bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border border-emerald-500/20">
              Auto-Bypass (Active)
            </span>
          {:else}
            <span class="text-[10px] font-mono px-1.5 py-0.2 rounded bg-indigo-500/10 text-indigo-700 dark:text-indigo-300 border border-indigo-500/30 font-medium">
              Plan Gate Active
            </span>
          {/if}
        </h4>
        <p class="text-[11.5px] text-ant-text-secondary mt-0.5">
          {settingsStore.planGateMode === 'bypass'
            ? 'Plan Gate bypass is active. Actions are approved automatically.'
            : 'Review the generated plan above. Workspace mutations and commands remain blocked until approved.'}
        </p>
      </div>
    </div>
  </div>

  <!-- Inline Custom Feedback Input Area -->
  {#if showCustomInput}
    <div class="space-y-2 pt-1 animate-in fade-in duration-150">
      <textarea
        bind:value={customFeedback}
        readonly={inputShieldStore.isReadOnly}
        placeholder="Describe revisions, additional constraints, or changes to the plan..."
        rows="2"
        class="w-full bg-ant-bg border border-ant-border-secondary dark:border-[#2e2e34] rounded-lg p-2.5 text-xs text-ant-text focus:outline-none focus:border-indigo-500/50 resize-none font-serif leading-relaxed placeholder:text-ant-text-muted"
        onkeydown={(e) => {
          if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
            e.preventDefault();
            handleSubmitCustom();
          }
        }}
      ></textarea>

      <div class="flex items-center justify-end space-x-2">
        <button
          type="button"
          onclick={() => showCustomInput = false}
          class="px-3 py-1 rounded-md text-xs text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition"
        >
          Cancel
        </button>
        <button
          type="button"
          onclick={handleSubmitCustom}
          disabled={!customFeedback.trim()}
          class="px-3.5 py-1 rounded-md text-xs font-medium text-white bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 transition flex items-center gap-1.5"
        >
          <MessageSquare size={12} />
          <span>Send Feedback & Revise</span>
        </button>
      </div>
    </div>
  {/if}

  <!-- Action Buttons Row -->
  <div class="flex items-center justify-between pt-1">
    <div class="flex items-center space-x-2">
      <!-- Approve and execute button -->
      <button
        type="button"
        onclick={handleApprove}
        class="flex items-center space-x-1.5 px-4 py-1.5 rounded-lg text-xs font-serif font-medium bg-indigo-600 hover:bg-indigo-500 text-white shadow-sm transition active:scale-95 cursor-pointer"
      >
        <Check size={13} class="stroke-[2.5]" />
        <span>Approve & Implement</span>
      </button>

      <!-- Custom feedback toggle button -->
      <button
        type="button"
        onclick={() => showCustomInput = !showCustomInput}
        class="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg text-xs font-serif text-ant-text-secondary hover:text-ant-text bg-ant-bg hover:bg-ant-bg-tertiary border border-ant-border-secondary dark:border-white/10 transition cursor-pointer"
      >
        <MessageSquare size={13} class="text-ant-text-secondary" />
        <span>Custom Feedback</span>
      </button>
    </div>

    <!-- Reject button -->
    <button
      type="button"
      onclick={handleReject}
      class="flex items-center space-x-1 px-3 py-1.5 rounded-lg text-xs font-serif text-ant-text-secondary hover:text-rose-500 hover:bg-rose-500/10 border border-transparent hover:border-rose-500/20 transition cursor-pointer"
      title="Reject plan and cancel implementation"
    >
      <X size={13} />
      <span>Reject</span>
    </button>
  </div>
</div>
