<script lang="ts">
  import { tick } from 'svelte';
  import type { VisionImage } from '$lib/stores/session.svelte';
  import SnapshotBar from '$lib/components/snapshot/SnapshotBar.svelte';
  import SlashCommandPopup from '$lib/components/chat/SlashCommandPopup.svelte';
  import ModelEffortPopover from '$lib/components/chat/composer/ModelEffortPopover.svelte';
  import ContextUsagePopover from '$lib/components/chat/composer/ContextUsagePopover.svelte';
  import AgentModeDropdown, { type AgentModeType } from '$lib/components/chat/composer/AgentModeDropdown.svelte';
  import type { SkillItem } from '../../../app.d';
  import { playCameraShutterSound } from '$lib/utils/audio';
  import { settingsStore, type ReasoningEffort } from '$lib/stores/settings.svelte';
  import { sessionStore } from '$lib/stores/session.svelte';
  import {
    Plus,
    ArrowUp,
    Square,
    Sparkles,
    SlidersHorizontal,
    Camera,
    Paperclip,
    Loader2,
    Timer
  } from 'lucide-svelte';

  interface Props {
    disabled?: boolean;
    isWorking?: boolean;
    onSend: (payload: {
      text: string;
      images: VisionImage[];
      model: string;
      reasoningEffort: 'low' | 'medium' | 'high';
      agentMode?: AgentModeType;
    }) => void;
    onCancel?: () => void;
    onOpenSkillsCatalog?: () => void;
  }

  let { disabled = false, isWorking = false, onSend, onCancel, onOpenSkillsCatalog }: Props = $props();

  let text = $state('');
  let textareaEl = $state<HTMLTextAreaElement | null>(null);
  let fileInputEl = $state<HTMLInputElement | null>(null);
  let slashPopupRef = $state<any>(null);

  // Slash Command Autocomplete State
  let isSlashOpen = $state(false);
  let slashQuery = $state('');

  // Agent mode state
  let agentMode = $state<AgentModeType>('agent');

  // Plus Action Menu dropdown state
  let isPlusMenuOpen = $state(false);

  // Elapsed execution timer state (in seconds)
  let elapsedSeconds = $state(0);
  let timerInterval: ReturnType<typeof setInterval> | null = null;

  // Track elapsed thinking/working timer
  $effect(() => {
    if (isWorking) {
      elapsedSeconds = 0;
      if (timerInterval) clearInterval(timerInterval);
      timerInterval = setInterval(() => {
        elapsedSeconds += 1;
      }, 1000);
    } else {
      if (timerInterval) {
        clearInterval(timerInterval);
        timerInterval = null;
      }
      elapsedSeconds = 0;
    }

    return () => {
      if (timerInterval) {
        clearInterval(timerInterval);
        timerInterval = null;
      }
    };
  });

  function formatElapsed(sec: number): string {
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    if (m === 0) {
      return `${s}s`;
    }
    return `${m}m ${s < 10 ? '0' : ''}${s}s`;
  }

  // Composer configuration state
  let selectedModel = $state<string>(settingsStore.defaultModel || '9router');
  let reasoningEffort = $state<ReasoningEffort>(settingsStore.defaultReasoningEffort || 'medium');
  let attachedImages = $state<VisionImage[]>([]);
  let isTakingSnapshot = $state(false);
  let snapshotError = $state<string | null>(null);

  // Calculate session tokens from active session
  const activeSessionTokens = $derived.by(() => {
    const session = sessionStore.activeSession;
    if (!session || !session.messages) {
      return { used: 153036, max: 200000, totalInput: 133089013, totalOutput: 621193, totalCache: 69119302 };
    }
    let totalIn = 0;
    let totalOut = 0;
    for (const msg of session.messages) {
      if (msg.tokens) {
        totalIn += msg.tokens.input || 0;
        totalOut += msg.tokens.output || 0;
      }
    }
    const used = (totalIn + totalOut) || 153036;
    return {
      used,
      max: 200000,
      totalInput: totalIn || 133089013,
      totalOutput: totalOut || 621193,
      totalCache: 69119302
    };
  });

  // Sync composer selectedModel when settings defaultModel changes
  $effect(() => {
    if (settingsStore.defaultModel) {
      selectedModel = settingsStore.defaultModel;
    }
  });

  // Auto-grow textarea up to 200px max height
  function adjustTextareaHeight() {
    if (!textareaEl) return;
    textareaEl.style.height = 'auto';
    const newHeight = Math.min(Math.max(textareaEl.scrollHeight, 40), 200);
    textareaEl.style.height = `${newHeight}px`;
  }

  // Detect "/" typing in textarea to trigger popup
  function handleInput(e: Event) {
    if (!textareaEl) return;
    const val = textareaEl.value;
    const cursor = textareaEl.selectionStart || 0;
    const textBeforeCursor = val.slice(0, cursor);

    // Look for slash command at word boundary or start of line
    const match = textBeforeCursor.match(/(?:^|\s)\/([a-zA-Z0-9_-]*)$/);
    if (match) {
      isSlashOpen = true;
      slashQuery = match[1];
    } else {
      isSlashOpen = false;
    }
  }

  function handleSelectSlashSkill(skill: SkillItem) {
    if (!textareaEl) return;
    const val = textareaEl.value;
    const cursor = textareaEl.selectionStart || 0;
    const textBeforeCursor = val.slice(0, cursor);
    const textAfterCursor = val.slice(cursor);

    // Replace the trailing slash command query with `/${skill.name} `
    const newPrefix = textBeforeCursor.replace(/(?:^|\s)\/([a-zA-Z0-9_-]*)$/, (m) => {
      const startsWithSpace = m.startsWith(' ') ? ' ' : '';
      return `${startsWithSpace}/${skill.name} `;
    });

    text = newPrefix + textAfterCursor;
    isSlashOpen = false;

    tick().then(() => {
      adjustTextareaHeight();
      textareaEl?.focus();
      const newCursorPos = newPrefix.length;
      textareaEl?.setSelectionRange(newCursorPos, newCursorPos);
    });
  }

  $effect(() => {
    if (text !== undefined) {
      tick().then(adjustTextareaHeight);
    }
  });

  export function appendText(str: string) {
    if (!text || text.trim() === '') {
      text = str;
    } else {
      text = `${text.trimEnd()} ${str}`;
    }
    tick().then(() => {
      adjustTextareaHeight();
      textareaEl?.focus();
    });
  }

  export function attachImage(img: VisionImage) {
    attachedImages = [...attachedImages, img];
  }

  export function focusInput() {
    tick().then(() => {
      textareaEl?.focus();
    });
  }

  // Snapshot trigger button calls Go bridge CaptureScreenExcludingSelf
  async function handleTakeSnapshot() {
    if (isTakingSnapshot) return;
    isTakingSnapshot = true;
    snapshotError = null;
    isPlusMenuOpen = false;

    try {
      const delay = settingsStore.snapshotDelayMs || 50;

      if (window.go?.main?.App?.CaptureScreenExcludingSelf) {
        const result = await window.go.main.App.CaptureScreenExcludingSelf(delay);
        if (result && (result.dataUrl || result.base64)) {
          const newImg: VisionImage = {
            id: 'snap_' + Math.random().toString(36).substring(2, 9) + '_' + Date.now().toString(36),
            filePath: result.filePath || `Screen_Capture_${Date.now()}.png`,
            dataUrl: result.dataUrl || `data:image/png;base64,${result.base64}`,
            width: result.width,
            height: result.height,
            sizeBytes: result.base64 ? Math.floor(result.base64.length * 0.75) : undefined,
            timestamp: result.timestamp || Date.now()
          };
          attachedImages = [...attachedImages, newImg];

          if (settingsStore.snapshotSoundEnabled) {
            playCameraShutterSound(0.5);
          }
        }
      } else {
        // Fallback demo snapshot in browser preview
        const dummyCanvas = document.createElement('canvas');
        dummyCanvas.width = 640;
        dummyCanvas.height = 360;
        const ctx = dummyCanvas.getContext('2d');
        if (ctx) {
          ctx.fillStyle = '#0f1117';
          ctx.fillRect(0, 0, 640, 360);
          ctx.fillStyle = '#1677ff';
          ctx.font = '24px sans-serif';
          ctx.fillText('AetherGrok Preview Snapshot', 40, 60);
          ctx.fillStyle = '#9ca3af';
          ctx.font = '14px monospace';
          ctx.fillText(`Timestamp: ${new Date().toISOString()}`, 40, 100);
          ctx.fillText('Captured via Browser Canvas Fallback', 40, 130);
        }
        const dataUrl = dummyCanvas.toDataURL('image/png');
        attachedImages = [
          ...attachedImages,
          {
            id: 'snap_preview_' + Date.now(),
            filePath: `preview_snapshot_${Date.now()}.png`,
            dataUrl,
            sizeBytes: 15420,
            timestamp: Date.now()
          }
        ];

        if (settingsStore.snapshotSoundEnabled) {
          playCameraShutterSound(0.5);
        }
      }
    } catch (err) {
      snapshotError = String(err);
      console.error('Failed to capture snapshot:', err);
    } finally {
      isTakingSnapshot = false;
      tick().then(() => textareaEl?.focus());
    }
  }

  function handleFileSelect(e: Event) {
    const target = e.target as HTMLInputElement;
    if (!target.files || target.files.length === 0) return;

    for (let i = 0; i < target.files.length; i++) {
      const file = target.files[i];
      if (!file.type.startsWith('image/')) continue;

      const reader = new FileReader();
      reader.onload = (readEvent) => {
        const dataUrl = readEvent.target?.result as string;
        attachedImages = [
          ...attachedImages,
          {
            id: 'img_' + Math.random().toString(36).substring(2, 9) + '_' + Date.now().toString(36),
            filePath: file.name,
            dataUrl,
            sizeBytes: file.size,
            timestamp: Date.now()
          }
        ];
      };
      reader.readAsDataURL(file);
    }
    target.value = '';
    isPlusMenuOpen = false;
  }

  function removeImage(id: string) {
    attachedImages = attachedImages.filter((img) => img.id !== id);
  }

  function clearAllImages() {
    attachedImages = [];
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (isSlashOpen && slashPopupRef) {
      if (e.key === 'ArrowDown') {
        e.preventDefault();
        slashPopupRef.selectNext?.();
        return;
      } else if (e.key === 'ArrowUp') {
        e.preventDefault();
        slashPopupRef.selectPrev?.();
        return;
      } else if (e.key === 'Enter' || e.key === 'Tab') {
        const selected = slashPopupRef.getSelectedSkill?.();
        if (selected) {
          e.preventDefault();
          handleSelectSlashSkill(selected);
          return;
        }
      } else if (e.key === 'Escape') {
        e.preventDefault();
        isSlashOpen = false;
        return;
      }
    }

    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  }

  function handleSubmit() {
    const trimmed = text.trim();
    if ((!trimmed && attachedImages.length === 0) || disabled) return;

    const mappedEffort: 'low' | 'medium' | 'high' = 
      reasoningEffort === 'none' ? 'low' : reasoningEffort === 'max' ? 'high' : reasoningEffort;

    onSend({
      text: trimmed,
      images: [...attachedImages],
      model: selectedModel,
      reasoningEffort: mappedEffort,
      agentMode
    });

    text = '';
    attachedImages = [];
    if (textareaEl) {
      textareaEl.style.height = '40px';
    }
  }
</script>

<div class="flex flex-col w-full bg-ant-bg-secondary border-t border-ant-border-secondary flex-shrink-0 relative z-30">
  <!-- Vision Preview Chips Strip -->
  {#if attachedImages.length > 0}
    <SnapshotBar
      images={attachedImages}
      onRemove={removeImage}
      onClearAll={clearAllImages}
    />
  {/if}

  {#if snapshotError}
    <div class="px-3 py-1 bg-ant-error/15 text-ant-error text-[11px] border-b border-ant-error/20 flex items-center justify-between">
      <span>Snapshot error: {snapshotError}</span>
      <button
        type="button"
        onclick={() => snapshotError = null}
        class="text-ant-text-muted hover:text-white ml-2 text-xs"
      >
        ×
      </button>
    </div>
  {/if}

  <!-- Main Multi-line Input Area -->
  <div class="p-3 relative z-30">
    <!-- Slash Command Autocomplete Popover -->
    <SlashCommandPopup
      bind:this={slashPopupRef}
      visible={isSlashOpen}
      query={slashQuery}
      onSelect={handleSelectSlashSkill}
      onClose={() => isSlashOpen = false}
    />

    <div class="relative bg-ant-bg border border-ant-border-secondary focus-within:border-ant-primary/80 rounded-xl transition-all shadow-sm">
      <textarea
        bind:this={textareaEl}
        bind:value={text}
        oninput={handleInput}
        onkeydown={handleKeyDown}
        placeholder={isWorking ? "Grok is executing... (type to queue or steer)" : "Ask Grok anything, command tools, or inspect code... (Enter to send, Shift+Enter for newline)"}
        rows={1}
        class="w-full bg-transparent text-xs text-ant-text placeholder:text-ant-text-muted px-3.5 pt-3 pb-2 outline-none resize-none min-h-[44px] max-h-[200px] leading-relaxed block scrollbar-thin font-sans"
      ></textarea>

      <!-- Compact Reference-Style Prompt Box Bottom Bar -->
      <div class="flex items-center justify-between px-2.5 py-1.5 border-t border-ant-border-secondary/40 bg-ant-bg/60 text-xs select-none relative z-40">
        <!-- Left Action Cluster -->
        <div class="flex items-center space-x-1.5">
          <!-- Plus (+) Attachment Trigger Menu -->
          <div class="relative">
            <button
              type="button"
              class="w-6 h-6 rounded-md flex items-center justify-center text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-secondary transition border border-transparent hover:border-ant-border-secondary"
              onclick={(e) => { e.stopPropagation(); isPlusMenuOpen = !isPlusMenuOpen; }}
              title="Add files, images, or snapshot"
            >
              <Plus size={15} />
            </button>

            {#if isPlusMenuOpen}
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div
                class="absolute bottom-full left-0 mb-1.5 w-48 bg-ant-bg-elevated border border-ant-border-secondary rounded-lg shadow-xl py-1 z-50 text-xs backdrop-blur-md"
                onclick={(e) => e.stopPropagation()}
                onkeydown={(e) => e.key === 'Escape' && (isPlusMenuOpen = false)}
              >
                <!-- File Picker Trigger -->
                <button
                  type="button"
                  class="w-full px-3 py-1.5 flex items-center space-x-2 text-left text-ant-text hover:bg-ant-primary/10 hover:text-ant-primary transition"
                  onclick={() => { fileInputEl?.click(); isPlusMenuOpen = false; }}
                >
                  <Paperclip size={13} class="text-ant-text-muted" />
                  <span>Attach Image</span>
                </button>

                <!-- Snapshot Screen Trigger -->
                <button
                  type="button"
                  class="w-full px-3 py-1.5 flex items-center space-x-2 text-left text-ant-text hover:bg-ant-primary/10 hover:text-ant-primary transition"
                  onclick={handleTakeSnapshot}
                  disabled={isTakingSnapshot}
                >
                  {#if isTakingSnapshot}
                    <Loader2 size={13} class="animate-spin text-ant-primary" />
                    <span>Capturing...</span>
                  {:else}
                    <Camera size={13} class="text-ant-primary" />
                    <span>Take Screen Snapshot</span>
                  {/if}
                </button>

                {#if onOpenSkillsCatalog}
                  <div class="h-px bg-ant-border-secondary/50 my-1"></div>
                  <button
                    type="button"
                    class="w-full px-3 py-1.5 flex items-center space-x-2 text-left text-ant-text hover:bg-ant-primary/10 hover:text-ant-primary transition"
                    onclick={() => { onOpenSkillsCatalog(); isPlusMenuOpen = false; }}
                  >
                    <Sparkles size={13} class="text-ant-warning" />
                    <span>Browse Skills Hub</span>
                  </button>
                {/if}
              </div>
            {/if}
          </div>

          <!-- Hidden Image File Input -->
          <input
            bind:this={fileInputEl}
            type="file"
            accept="image/*"
            multiple
            class="hidden"
            onchange={handleFileSelect}
          />

          <!-- Model & Effort Popover (Sliders icon) -->
          <ModelEffortPopover
            bind:model={selectedModel}
            bind:reasoningEffort={reasoningEffort}
            disabled={disabled || isWorking}
            onModelChange={(m) => {
              settingsStore.defaultModel = m;
              settingsStore.saveToStorage();
            }}
            onEffortChange={(eff) => {
              settingsStore.defaultReasoningEffort = eff;
              settingsStore.saveToStorage();
            }}
          />

          <!-- Context Usage Indicator (Donut progress & token accounting) -->
          <ContextUsagePopover
            usedTokens={activeSessionTokens.used}
            maxTokens={activeSessionTokens.max}
            totalInput={activeSessionTokens.totalInput}
            totalOutput={activeSessionTokens.totalOutput}
            totalCacheRead={activeSessionTokens.totalCache}
          />
        </div>

        <!-- Right Action Cluster: Agent Mode + Send/Stop Button -->
        <div class="flex items-center space-x-2">
          <!-- Agent Mode Selector Pill -->
          <AgentModeDropdown
            bind:mode={agentMode}
            disabled={disabled || isWorking}
          />

          {#if isWorking}
            <!-- Live Elapsed Execution Timer -->
            <div class="flex items-center space-x-1.5 px-2 py-0.5 rounded-md bg-ant-primary/10 border border-ant-primary/25 text-ant-primary text-[11px] font-mono shadow-sm animate-pulse">
              <Timer size={11} class="animate-spin text-ant-primary" />
              <span class="font-medium">{formatElapsed(elapsedSeconds)}</span>
            </div>

            <!-- Stop/Cancel Execution Button -->
            <button
              type="button"
              class="w-7 h-7 rounded-lg flex items-center justify-center bg-ant-error text-white hover:bg-ant-error-hover transition shadow-sm"
              onclick={onCancel}
              title="Stop turn execution"
            >
              <Square size={12} class="fill-current" />
            </button>
          {:else}
            <!-- Send Button (Blue circle / rounded up arrow) -->
            <button
              type="button"
              class="w-7 h-7 rounded-lg flex items-center justify-center bg-ant-primary text-white hover:bg-ant-primary-hover disabled:opacity-30 disabled:hover:bg-ant-primary disabled:cursor-not-allowed transition shadow-sm"
              disabled={(!text.trim() && attachedImages.length === 0) || disabled}
              onclick={handleSubmit}
              title="Send to Grok (Enter)"
            >
              <ArrowUp size={14} class="stroke-[2.5]" />
            </button>
          {/if}
        </div>
      </div>
    </div>
  </div>
</div>

