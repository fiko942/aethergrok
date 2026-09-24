<script lang="ts">
  import { tick } from 'svelte';
  import type { VisionImage } from '$lib/stores/session.svelte';
  import SnapshotBar from '$lib/components/snapshot/SnapshotBar.svelte';
  import Button from '$lib/antd/Button.svelte';
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
  }

  let { disabled = false, isWorking = false, onSend, onCancel }: Props = $props();

  let text = $state('');
  let textareaEl = $state<HTMLTextAreaElement | null>(null);
  let fileInputEl = $state<HTMLInputElement | null>(null);

  // Composer configuration state
  let selectedModel = $state<'grok-4.6' | 'grok-code'>('grok-4.6');
  let reasoningEffort = $state<'low' | 'medium' | 'high'>('medium');
  let attachedImages = $state<VisionImage[]>([]);
  let isTakingSnapshot = $state(false);
  let snapshotError = $state<string | null>(null);

  // Auto-grow textarea up to 200px max height
  function adjustTextareaHeight() {
    if (!textareaEl) return;
    textareaEl.style.height = 'auto';
    const newHeight = Math.min(Math.max(textareaEl.scrollHeight, 40), 200);
    textareaEl.style.height = `${newHeight}px`;
  }

  $effect(() => {
    // Reactively adjust when text changes
    if (text !== undefined) {
      tick().then(adjustTextareaHeight);
    }
  });

  // Snapshot trigger button calls Go bridge CaptureScreenExcludingSelf
  async function handleTakeSnapshot() {
    if (isTakingSnapshot) return;
    isTakingSnapshot = true;
    snapshotError = null;

    try {
      if (window.go?.main?.App?.CaptureScreenExcludingSelf) {
        // Platform auto compositor delay is handled in Go, pass 0 or 50
        const result = await window.go.main.App.CaptureScreenExcludingSelf(50);
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
        class="w-full bg-transparent text-xs text-white placeholder:text-ant-text-muted px-3.5 pt-3 pb-2 outline-none resize-none min-h-[42px] max-h-[200px] leading-relaxed block scrollbar-thin"
      ></textarea>

      <!-- Composer Bottom Control Bar -->
      <div class="flex items-center justify-between px-3 py-2 border-t border-ant-border-secondary/60 bg-ant-bg/70 text-xs select-none">
        <!-- Left Controls: Snapshot, Attachments, Model Selector, Effort Chips -->
        <div class="flex items-center gap-2 flex-wrap">
          <!-- Non-intrusive Snapshot Trigger Button -->
          <button
            type="button"
            onclick={handleTakeSnapshot}
            disabled={isTakingSnapshot}
            class="flex items-center space-x-1.5 px-2.5 py-1 rounded-md text-[11px] font-medium bg-ant-bg-tertiary hover:bg-ant-primary/20 text-ant-text hover:text-ant-primary border border-ant-border-secondary transition shadow-sm active:scale-95 disabled:opacity-50"
            title="Capture screen snapshot excluding this window (Calls Go bridge CaptureScreenExcludingSelf)"
          >
            {#if isTakingSnapshot}
              <Loader2 size={13} class="animate-spin text-ant-primary" />
              <span>Capturing...</span>
            {:else}
              <Camera size={13} class="text-ant-primary" />
              <span>Snapshot</span>
            {/if}
          </button>

          <!-- Attachment Upload Button -->
          <button
            type="button"
            onclick={() => fileInputEl?.click()}
            class="flex items-center space-x-1 px-2 py-1 rounded-md text-[11px] text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-tertiary border border-transparent hover:border-ant-border-secondary transition"
            title="Attach vision image files"
          >
            <Paperclip size={13} />
            <span class="hidden sm:inline">Attach</span>
          </button>

          <input
            bind:this={fileInputEl}
            type="file"
            accept="image/*"
            multiple
            class="hidden"
            onchange={handleFileSelect}
          />

          <!-- Divider -->
          <div class="w-px h-4 bg-ant-border-secondary hidden sm:block"></div>

          <!-- Model Selector Chip (grok-4.6, grok-code) -->
          <div class="flex items-center bg-ant-bg-tertiary border border-ant-border-secondary rounded-md p-0.5">
            <button
              type="button"
              onclick={() => selectedModel = 'grok-4.6'}
              class="px-2 py-0.5 rounded text-[11px] font-medium transition {selectedModel === 'grok-4.6' ? 'bg-ant-primary text-white shadow-sm' : 'text-ant-text-muted hover:text-ant-text'}"
            >
              grok-4.6
            </button>
            <button
              type="button"
              onclick={() => selectedModel = 'grok-code'}
              class="px-2 py-0.5 rounded text-[11px] font-medium transition {selectedModel === 'grok-code' ? 'bg-ant-primary text-white shadow-sm' : 'text-ant-text-muted hover:text-ant-text'}"
            >
              grok-code
            </button>
          </div>

          <!-- Reasoning Effort Indicator Chip -->
          <button
            type="button"
            onclick={cycleEffort}
            class="flex items-center space-x-1.5 px-2 py-1 rounded-md text-[11px] bg-ant-bg-tertiary hover:bg-ant-bg-secondary border border-ant-border-secondary text-ant-text-secondary transition"
            title="Toggle reasoning effort: Low, Medium, High"
          >
            <SlidersHorizontal size={11} class="text-ant-text-muted" />
            <span class="text-ant-text-muted text-[10px]">Effort:</span>
            <span class="font-semibold uppercase text-[10px] {reasoningEffort === 'high' ? 'text-ant-warning' : reasoningEffort === 'medium' ? 'text-ant-primary' : 'text-ant-text'}">
              {reasoningEffort}
            </span>
          </button>
        </div>

        <!-- Right Controls: Submit or Cancel Button -->
        <div class="flex items-center space-x-2 flex-shrink-0">
          {#if isWorking}
            <Button
              type="default"
              size="small"
              danger
              onclick={onCancel}
              class="flex items-center"
            >
              <Square size={12} class="mr-1 fill-current" /> Stop
            </Button>
          {:else}
            <Button
              type="primary"
              size="small"
              disabled={disabled || (!text.trim() && attachedImages.length === 0)}
              onclick={handleSubmit}
              class="flex items-center shadow-md shadow-ant-primary/20"
            >
              <Send size={12} class="mr-1.5" /> Send
            </Button>
          {/if}
        </div>
      </div>
    </div>
  </div>
</div>
