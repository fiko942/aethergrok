<script lang="ts">
  import { tick } from 'svelte';
  import type { VisionImage } from '$lib/stores/session.svelte';
  import SnapshotBar from '$lib/components/snapshot/SnapshotBar.svelte';
  import Button from '$lib/antd/Button.svelte';
  import { playCameraShutterSound } from '$lib/utils/audio';
  import { settingsStore } from '$lib/stores/settings.svelte';
  import {
    Camera,
    Send,
    Square,
    Sparkles,
    SlidersHorizontal,
    Cpu,
    Loader2,
    ImagePlus,
    Paperclip
  } from 'lucide-svelte';

  interface Props {
    disabled?: boolean;
    isWorking?: boolean;
    onSend: (payload: {
      text: string;
      images: VisionImage[];
      model: string;
      reasoningEffort: 'low' | 'medium' | 'high';
    }) => void;
    onCancel?: () => void;
    onOpenSkillsCatalog?: () => void;
  }

  let { disabled = false, isWorking = false, onSend, onCancel, onOpenSkillsCatalog }: Props = $props();

  let text = $state('');
  let textareaEl = $state<HTMLTextAreaElement | null>(null);
  let fileInputEl = $state<HTMLInputElement | null>(null);

  // Composer configuration state
  let selectedModel = $state<string>(settingsStore.defaultModel || '9router');
  let reasoningEffort = $state<'low' | 'medium' | 'high'>('medium');
  let attachedImages = $state<VisionImage[]>([]);
  let isTakingSnapshot = $state(false);
  let snapshotError = $state<string | null>(null);

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
  }

  function removeImage(id: string) {
    attachedImages = attachedImages.filter((img) => img.id !== id);
  }

  function clearAllImages() {
    attachedImages = [];
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit();
    }
  }

  function handleSubmit() {
    const trimmed = text.trim();
    if ((!trimmed && attachedImages.length === 0) || disabled) return;

    onSend({
      text: trimmed,
      images: [...attachedImages],
      model: selectedModel,
      reasoningEffort
    });

    text = '';
    attachedImages = [];
    if (textareaEl) {
      textareaEl.style.height = '40px';
    }
  }

  function cycleEffort() {
    if (reasoningEffort === 'low') reasoningEffort = 'medium';
    else if (reasoningEffort === 'medium') reasoningEffort = 'high';
    else reasoningEffort = 'low';
  }
</script>

<div class="flex flex-col w-full bg-ant-bg-secondary border-t border-ant-border flex-shrink-0">
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
  <div class="p-3">
    <div class="relative bg-ant-bg border border-ant-border focus-within:border-ant-primary rounded-xl transition-all shadow-inner overflow-hidden">
      <textarea
        bind:this={textareaEl}
        bind:value={text}
        onkeydown={handleKeyDown}
        placeholder={isWorking ? "Grok is executing... (type to queue or steer)" : "Ask Grok anything, command tools, or inspect code... (Enter to send, Shift+Enter for newline)"}
        rows={1}
        class="w-full bg-transparent text-xs text-ant-text placeholder:text-ant-text-muted px-3.5 pt-3 pb-2 outline-none resize-none min-h-[42px] max-h-[200px] leading-relaxed block scrollbar-thin font-sans"
      ></textarea>

      <!-- Composer Bottom Control Bar -->
      <div class="flex items-center justify-between px-3 py-2 border-t border-ant-border-secondary/60 bg-ant-bg/70 text-xs select-none">
        <div class="flex items-center space-x-1.5">
          <!-- Non-Intrusive Snapshot Trigger -->
          <button
            type="button"
            class="inline-flex items-center px-2 py-1 rounded-md text-[11px] font-medium transition {isTakingSnapshot
              ? 'bg-ant-primary/20 text-ant-primary cursor-wait'
              : 'bg-ant-bg-secondary hover:bg-ant-primary/10 hover:text-ant-primary text-ant-text-secondary border border-ant-border'}"
            onclick={handleTakeSnapshot}
            disabled={isTakingSnapshot || disabled}
            title="Capture Active Screen excluding AetherGrok window (Cmd/Ctrl+Shift+S)"
          >
            {#if isTakingSnapshot}
              <Loader2 size={13} class="animate-spin mr-1" />
              <span>Capturing...</span>
            {:else}
              <Camera size={13} class="mr-1 text-ant-primary" />
              <span>Snapshot</span>
            {/if}
          </button>

          <!-- File / Image Picker -->
          <input
            bind:this={fileInputEl}
            type="file"
            accept="image/*"
            multiple
            class="hidden"
            onchange={handleFileSelect}
          />
          <button
            type="button"
            class="p-1 rounded-md text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-secondary transition border border-transparent"
            onclick={() => fileInputEl?.click()}
            title="Attach vision reference image"
          >
            <Paperclip size={14} />
          </button>

          <div class="h-3 w-px bg-ant-border mx-1"></div>

          <!-- Skills Catalog Quick Trigger -->
          {#if onOpenSkillsCatalog}
            <button
              type="button"
              class="inline-flex items-center px-1.5 py-1 rounded text-[11px] text-ant-text-muted hover:text-ant-primary transition"
              onclick={onOpenSkillsCatalog}
              title="Browse and insert skills (Cmd/Ctrl + K)"
            >
              <Sparkles size={13} class="mr-1" />
              <span>Skills Hub</span>
            </button>
          {/if}

          <!-- Model Selector Pill (cycles available CLI models) -->
          <button
            type="button"
            class="inline-flex items-center px-2 py-0.5 rounded text-[11px] font-mono text-ant-text-secondary bg-ant-bg-secondary border border-ant-border hover:border-ant-primary/40 transition"
            onclick={() => {
              const models = ['9router', '9router-general-purpose', '9router-explore', '9router-plan'];
              const currIdx = models.indexOf(selectedModel);
              const nextIdx = (currIdx + 1) % models.length;
              selectedModel = models[nextIdx];
            }}
            title="Toggle active model engine"
          >
            <Cpu size={12} class="mr-1 text-ant-primary" />
            <span>{selectedModel}</span>
          </button>

          <!-- Reasoning Effort Indicator -->
          <button
            type="button"
            class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold uppercase tracking-wider text-ant-text-muted hover:text-ant-text transition"
            onclick={cycleEffort}
            title="Click to cycle reasoning effort stop"
          >
            <SlidersHorizontal size={11} class="mr-1" />
            <span class="text-ant-primary">{reasoningEffort}</span>
          </button>
        </div>

        <!-- Submit / Cancel Action -->
        <div class="flex items-center space-x-2">
          {#if isWorking}
            <button
              type="button"
              class="inline-flex items-center px-3 py-1.5 rounded-lg text-xs font-medium bg-ant-error text-white hover:bg-ant-error-hover transition shadow-sm"
              onclick={onCancel}
              title="Stop turn execution"
            >
              <Square size={12} class="mr-1.5 fill-current" />
              <span>Cancel</span>
            </button>
          {:else}
            <button
              type="button"
              class="inline-flex items-center px-3 py-1.5 rounded-lg text-xs font-semibold bg-ant-primary text-white hover:bg-ant-primary-hover disabled:opacity-40 disabled:cursor-not-allowed transition shadow-sm"
              disabled={(!text.trim() && attachedImages.length === 0) || disabled}
              onclick={handleSubmit}
              title="Send to Grok (Enter)"
            >
              <Send size={12} class="mr-1.5" />
              <span>Send</span>
            </button>
          {/if}
        </div>
      </div>
    </div>
  </div>
</div>
