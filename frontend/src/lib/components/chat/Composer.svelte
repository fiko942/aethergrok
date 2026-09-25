<script lang="ts">
  import { tick, onMount, onDestroy } from 'svelte';
  import type { VisionImage, AttachedFile, QueuedPrompt } from '$lib/stores/session.svelte';
  import SlashCommandPopup from '$lib/components/chat/SlashCommandPopup.svelte';
  import ModelEffortPopover from '$lib/components/chat/composer/ModelEffortPopover.svelte';
  import ContextUsagePopover from '$lib/components/chat/composer/ContextUsagePopover.svelte';
  import AgentModeDropdown, { type AgentModeType } from '$lib/components/chat/composer/AgentModeDropdown.svelte';
  import AttachmentChip from '$lib/components/chat/composer/AttachmentChip.svelte';
  import QueueStackBar from '$lib/components/chat/composer/QueueStackBar.svelte';
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
    FileText,
    FileCode,
    Loader2,
    Timer,
    X,
    Eye,
    Upload
  } from 'lucide-svelte';

  interface Props {
    disabled?: boolean;
    isWorking?: boolean;
    onSend: (payload: {
      text: string;
      images: VisionImage[];
      attachments?: AttachedFile[];
      model: string;
      reasoningEffort: 'low' | 'medium' | 'high';
      agentMode?: AgentModeType;
    }) => void;
    onSteer?: (prompt: QueuedPrompt) => void;
    onCancel?: () => void;
    onOpenSkillsCatalog?: () => void;
  }

  let { disabled = false, isWorking = false, onSend, onSteer, onCancel, onOpenSkillsCatalog }: Props = $props();

  let text = $state('');
  let textareaEl = $state<HTMLTextAreaElement | null>(null);
  let fileInputEl = $state<HTMLInputElement | null>(null);
  let plusMenuContainerEl = $state<HTMLDivElement | null>(null);
  let slashPopupRef = $state<any>(null);

  // Slash Command Autocomplete State
  let isSlashOpen = $state(false);
  let slashQuery = $state('');

  // Agent mode state
  let agentMode = $state<AgentModeType>('agent');

  // Plus Action Menu dropdown state
  let isPlusMenuOpen = $state(false);

  // Full file and preview modal state
  let attachedFiles = $state<AttachedFile[]>([]);
  let previewModalItem = $state<AttachedFile | null>(null);

  // Active session queue
  const currentQueue = $derived(sessionStore.activeSession?.queuedPrompts || []);

  // Click outside listener for plus menu
  function handleWindowClick(e: MouseEvent) {
    if (isPlusMenuOpen && plusMenuContainerEl && !plusMenuContainerEl.contains(e.target as Node)) {
      isPlusMenuOpen = false;
    }
  }

  onMount(() => {
    window.addEventListener('click', handleWindowClick);
  });

  onDestroy(() => {
    window.removeEventListener('click', handleWindowClick);
  });

  // Elapsed execution timer state (in milliseconds and seconds)
  let elapsedMs = $state(0);
  let timerInterval: ReturnType<typeof setInterval> | null = null;
  let timerStart = 0;

  // Track elapsed thinking/working timer
  $effect(() => {
    if (isWorking) {
      elapsedMs = 0;
      timerStart = Date.now();
      if (timerInterval) clearInterval(timerInterval);
      timerInterval = setInterval(() => {
        elapsedMs = Date.now() - timerStart;
      }, 100);
    } else {
      if (timerInterval) {
        clearInterval(timerInterval);
        timerInterval = null;
      }
      elapsedMs = 0;
    }

    return () => {
      if (timerInterval) {
        clearInterval(timerInterval);
        timerInterval = null;
      }
    };
  });

  function formatElapsed(ms: number): string {
    if (ms < 1000) {
      return `${ms}ms`;
    }
    const sec = Math.floor(ms / 1000);
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    if (m === 0) {
      return `${(ms / 1000).toFixed(1)}s`;
    }
    return `${m}m ${s < 10 ? '0' : ''}${s}s`;
  }

  // Composer configuration state
  let selectedModel = $state<string>(settingsStore.defaultModel || '9router');
  let reasoningEffort = $state<ReasoningEffort>(settingsStore.defaultReasoningEffort || 'medium');
  let attachedImages = $state<VisionImage[]>([]);
  let isTakingSnapshot = $state(false);
  let snapshotError = $state<string | null>(null);

  // Drag and drop overlay state
  let isDragOver = $state(false);
  let dragCounter = 0;

  // Calculate session tokens from active session
  const activeSessionTokens = $derived.by(() => {
    const session = sessionStore.activeSession;
    if (session?.usage) {
      return {
        used: session.usage.usedTokens,
        max: session.usage.maxTokens,
        lastTurnInput: session.usage.lastTurnInput,
        lastTurnOutput: session.usage.lastTurnOutput,
        lastTurnCacheRead: session.usage.lastTurnCacheRead,
        lastTurnReasoning: session.usage.lastTurnReasoning,
        lastTurnModelCalls: session.usage.lastTurnModelCalls,
        totalInput: session.usage.totalInput,
        totalOutput: session.usage.totalOutput,
        totalCache: session.usage.totalCacheRead
      };
    }

    if (!session || !session.messages || session.messages.length === 0) {
      return {
        used: 0,
        max: settingsStore.maxContextTokens || 200000,
        lastTurnInput: 0,
        lastTurnOutput: 0,
        lastTurnCacheRead: 0,
        lastTurnReasoning: 0,
        lastTurnModelCalls: 0,
        totalInput: 0,
        totalOutput: 0,
        totalCache: 0
      };
    }

    let totalIn = 0;
    let totalOut = 0;
    for (const msg of session.messages) {
      if (msg.tokens) {
        totalIn += msg.tokens.input || 0;
        totalOut += msg.tokens.output || 0;
      }
    }
    const used = totalIn + totalOut;
    return {
      used,
      max: settingsStore.maxContextTokens || 200000,
      lastTurnInput: 0,
      lastTurnOutput: 0,
      lastTurnCacheRead: 0,
      lastTurnReasoning: 0,
      lastTurnModelCalls: 0,
      totalInput: totalIn,
      totalOutput: totalOut,
      totalCache: 0
    };
  });

  // Sync composer selectedModel when settings defaultModel changes
  $effect(() => {
    if (settingsStore.defaultModel) {
      selectedModel = settingsStore.defaultModel;
    }
  });

  // Per-session prompt draft isolation
  let activeSessionId = $derived(sessionStore.activeSessionId);
  let trackedSessionId: string | null = null;

  $effect(() => {
    const curId = activeSessionId;
    if (curId !== trackedSessionId) {
      // 1. Save draft of the previous session if any
      if (trackedSessionId) {
        const prevSession = sessionStore.sessions.find((s) => s.id === trackedSessionId);
        if (prevSession) {
          prevSession.draft = {
            text,
            images: [...attachedImages],
            attachments: [...attachedFiles]
          };
        }
      }

      // 2. Load draft of the newly selected session
      if (curId) {
        const newSession = sessionStore.sessions.find((s) => s.id === curId);
        if (newSession && newSession.draft) {
          text = newSession.draft.text || '';
          attachedImages = [...(newSession.draft.images || [])];
          attachedFiles = [...(newSession.draft.attachments || [])];
        } else {
          text = '';
          attachedImages = [];
          attachedFiles = [];
        }
      } else {
        text = '';
        attachedImages = [];
        attachedFiles = [];
      }

      trackedSessionId = curId;
      tick().then(() => adjustTextareaHeight());
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
    attachedFiles = [
      ...attachedFiles,
      {
        id: img.id,
        name: img.filePath.split(/[/\\]/).pop() || 'Snapshot.jpg',
        filePath: img.filePath,
        sizeBytes: img.sizeBytes,
        dataUrl: img.dataUrl,
        isImage: true,
        timestamp: img.timestamp
      }
    ];
  }

  export function focusInput() {
    tick().then(() => {
      textareaEl?.focus();
    });
  }

  // Restore prompt text, images, and attachments from a rolled back turn or queue
  export function restorePrompt(payload: { text: string; images?: VisionImage[]; attachments?: AttachedFile[] }) {
    text = payload.text || '';
    attachedImages = payload.images ? [...payload.images] : [];
    attachedFiles = payload.attachments ? [...payload.attachments] : [];

    tick().then(() => {
      adjustTextareaHeight();
      textareaEl?.focus();
    });
  }

  // Load a queued prompt back into composer for editing, removing it from queue
  function handleEditQueuedPrompt(promptItem: QueuedPrompt) {
    text = promptItem.text || '';
    attachedImages = [...(promptItem.images || [])];
    attachedFiles = [...(promptItem.attachments || [])];

    // Remove from queue
    if (sessionStore.activeSessionId) {
      sessionStore.removeQueuedPrompt(sessionStore.activeSessionId, promptItem.id);
    }

    tick().then(() => {
      adjustTextareaHeight();
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
          const calculatedSize = result.sizeBytes || (result.base64 ? Math.floor(result.base64.length * 0.75) : undefined);
          const newImg: VisionImage = {
            id: 'snap_' + Math.random().toString(36).substring(2, 9) + '_' + Date.now().toString(36),
            filePath: result.filePath || `Screen_Capture_${Date.now()}.jpg`,
            dataUrl: result.dataUrl || `data:image/jpeg;base64,${result.base64}`,
            width: result.width,
            height: result.height,
            sizeBytes: calculatedSize,
            timestamp: result.timestamp || Date.now()
          };
          attachedImages = [...attachedImages, newImg];
          attachedFiles = [
            ...attachedFiles,
            {
              id: newImg.id,
              name: newImg.filePath.split(/[/\\]/).pop() || 'Screen_Capture.jpg',
              filePath: newImg.filePath,
              sizeBytes: newImg.sizeBytes,
              dataUrl: newImg.dataUrl,
              isImage: true,
              timestamp: newImg.timestamp
            }
          ];

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
        const newImg: VisionImage = {
          id: 'snap_preview_' + Date.now(),
          filePath: `preview_snapshot_${Date.now()}.png`,
          dataUrl,
          sizeBytes: 15420,
          timestamp: Date.now()
        };
        attachedImages = [...attachedImages, newImg];
        attachedFiles = [
          ...attachedFiles,
          {
            id: newImg.id,
            name: newImg.filePath,
            filePath: newImg.filePath,
            sizeBytes: newImg.sizeBytes,
            dataUrl: newImg.dataUrl,
            isImage: true,
            timestamp: newImg.timestamp
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

  // Centralized file processor for input selection, drag-and-drop, and paste
  function processFiles(files: FileList | File[]) {
    for (let i = 0; i < files.length; i++) {
      const file = files[i];
      const isImg = file.type.startsWith('image/');
      const fileId = 'file_' + Math.random().toString(36).substring(2, 9) + '_' + Date.now().toString(36);

      if (isImg) {
        const reader = new FileReader();
        reader.onload = (readEvent) => {
          const dataUrl = readEvent.target?.result as string;
          const newImg: VisionImage = {
            id: fileId,
            filePath: file.name,
            dataUrl,
            sizeBytes: file.size,
            timestamp: Date.now()
          };
          attachedImages = [...attachedImages, newImg];
          attachedFiles = [
            ...attachedFiles,
            {
              id: fileId,
              name: file.name,
              filePath: file.name,
              sizeBytes: file.size,
              mimeType: file.type,
              dataUrl,
              isImage: true,
              timestamp: Date.now()
            }
          ];
        };
        reader.readAsDataURL(file);
      } else {
        // Read text/markdown/code/json files as text for LLM injection
        const reader = new FileReader();
        reader.onload = (readEvent) => {
          const content = readEvent.target?.result as string;
          attachedFiles = [
            ...attachedFiles,
            {
              id: fileId,
              name: file.name,
              filePath: file.name,
              sizeBytes: file.size,
              mimeType: file.type || 'text/plain',
              content,
              isImage: false,
              timestamp: Date.now()
            }
          ];
        };
        reader.readAsText(file);
      }
    }
  }

  function handleFileSelect(e: Event) {
    const target = e.target as HTMLInputElement;
    if (!target.files || target.files.length === 0) return;
    processFiles(target.files);
    target.value = '';
    isPlusMenuOpen = false;
  }

  // Drag and Drop handlers for file drop directly into prompt box
  function handleContainerDragEnter(e: DragEvent) {
    if (e.dataTransfer?.types?.includes('Files')) {
      e.preventDefault();
      dragCounter++;
      isDragOver = true;
    }
  }

  export function handleExternalFiles(files: FileList | File[]) {
    processFiles(files);
  }

  function handleContainerDragOver(e: DragEvent) {
    if (e.dataTransfer?.types?.includes('Files')) {
      e.preventDefault();
      if (e.dataTransfer) e.dataTransfer.dropEffect = 'copy';
      isDragOver = true;
    }
  }

  function handleContainerDragLeave(e: DragEvent) {
    e.preventDefault();
    dragCounter--;
    if (dragCounter <= 0) {
      dragCounter = 0;
      isDragOver = false;
    }
  }

  function handleContainerDrop(e: DragEvent) {
    e.preventDefault();
    dragCounter = 0;
    isDragOver = false;
    if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
      processFiles(e.dataTransfer.files);
    }
  }

  // Paste handler: if images or files are in clipboard, intercept them as attachments
  function handlePaste(e: ClipboardEvent) {
    if (e.clipboardData?.files && e.clipboardData.files.length > 0) {
      e.preventDefault();
      processFiles(e.clipboardData.files);
    }
  }

  function removeAttachment(id: string) {
    attachedImages = attachedImages.filter((img) => img.id !== id);
    attachedFiles = attachedFiles.filter((att) => att.id !== id);
  }

  function clearAllAttachments() {
    attachedImages = [];
    attachedFiles = [];
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
    if ((!trimmed && attachedImages.length === 0 && attachedFiles.length === 0) || disabled) return;

    // Intelligent auto-detection of plan mode from prompt keywords
    let effectiveMode: AgentModeType = agentMode;
    const lower = trimmed.toLowerCase();
    if (
      lower.startsWith('coba bikin plan') ||
      lower.startsWith('bikin plan') ||
      lower.startsWith('buat plan') ||
      lower.startsWith('buat perencanaan') ||
      lower.startsWith('create a plan') ||
      lower.startsWith('plan for') ||
      lower.startsWith('/plan') ||
      lower.includes('mode plan') ||
      lower.includes('planning dulu')
    ) {
      effectiveMode = 'plan';
    }

    const mappedEffort: 'low' | 'medium' | 'high' = 
      reasoningEffort === 'none' ? 'low' : reasoningEffort === 'max' ? 'high' : reasoningEffort;

    onSend({
      text: trimmed,
      images: [...attachedImages],
      attachments: [...attachedFiles],
      model: selectedModel,
      reasoningEffort: mappedEffort,
      agentMode: effectiveMode
    });

    text = '';
    attachedImages = [];
    attachedFiles = [];
    if (sessionStore.activeSessionId) {
      const curSession = sessionStore.sessions.find((s) => s.id === sessionStore.activeSessionId);
      if (curSession) {
        curSession.draft = undefined;
      }
    }
    if (textareaEl) {
      textareaEl.style.height = '40px';
    }
  }
</script>

<div class="flex flex-col w-full bg-ant-bg-secondary border-t border-white/5 flex-shrink-0 relative z-30">
  <!-- Interactive Queue Stack Bar -->
  <div class="px-3 pt-2">
    <QueueStackBar
      queue={currentQueue}
      onSteer={(prompt) => {
        if (onSteer) {
          onSteer(prompt);
        } else {
          sessionStore.removeQueuedPrompt(sessionStore.activeSessionId || '', prompt.id);
          handleSubmit();
        }
      }}
      onReorder={(fromIdx, toIdx) => {
        sessionStore.reorderQueuedPrompt(sessionStore.activeSessionId || '', fromIdx, toIdx);
      }}
      onRemove={(id) => {
        sessionStore.removeQueuedPrompt(sessionStore.activeSessionId || '', id);
      }}
      onEdit={(prompt) => {
        handleEditQueuedPrompt(prompt);
      }}
    />
  </div>

  {#if snapshotError}
    <div class="px-3 py-1 bg-ant-error/15 text-ant-error text-[11px] border-b border-ant-error/20 flex items-center justify-between">
      <span>Snapshot error: {snapshotError}</span>
      <button
        type="button"
        onclick={() => snapshotError = null}
        class="text-ant-text-muted hover:text-ant-text ml-2 text-xs"
      >
        ×
      </button>
    </div>
  {/if}

  <!-- Main Multi-line Input Area with IN-BOX Attachment Chips -->
  <div class="p-3 relative z-30">
    <!-- Slash Command Autocomplete Popover -->
    <SlashCommandPopup
      bind:this={slashPopupRef}
      visible={isSlashOpen}
      query={slashQuery}
      onSelect={handleSelectSlashSkill}
      onClose={() => isSlashOpen = false}
    />

    <div
      class="relative bg-ant-bg border border-white/[0.04] hover:border-white/[0.08] focus-within:!border-ant-primary/30 rounded-xl transition-all shadow-sm flex flex-col"
      ondragenter={handleContainerDragEnter}
      ondragover={handleContainerDragOver}
      ondragleave={handleContainerDragLeave}
      ondrop={handleContainerDrop}
      role="region"
      aria-label="Prompt and attachment drop zone"
    >
      <!-- Visual Drag-and-Drop Active Overlay -->
      {#if isDragOver}
        <div class="absolute inset-0 z-50 bg-[#121316]/95 border-2 border-dashed border-blue-500/50 rounded-xl flex flex-col items-center justify-center space-y-1.5 backdrop-blur-md pointer-events-none animate-in fade-in duration-100 select-none">
          <div class="p-2 rounded-full bg-blue-500/10 text-blue-400 border border-blue-500/20 shadow-sm">
            <Upload size={18} />
          </div>
          <p class="text-xs font-serif text-ant-text font-medium">
            Drop files or images to attach
          </p>
          <p class="text-[10px] font-mono text-ant-text-muted">
            Images, Markdown, PDF, Code
          </p>
        </div>
      {/if}

      <!-- In-Box Attachment Chips Strip (Directly inside prompt box) -->
      {#if attachedFiles.length > 0}
        <div class="px-3 pt-2.5 pb-1 flex items-center gap-1.5 flex-wrap border-b border-white/5 bg-ant-bg/80 select-none rounded-t-xl">
          <div class="text-[10px] font-mono text-ant-text-muted flex items-center gap-1 mr-1">
            <Paperclip size={11} class="text-ant-primary" />
            <span>Files ({attachedFiles.length})</span>
          </div>

          {#each attachedFiles as att (att.id)}
            <AttachmentChip
              attachment={att}
              onRemove={removeAttachment}
              onPreview={(item) => previewModalItem = item}
            />
          {/each}

          <button
            type="button"
            onclick={clearAllAttachments}
            class="text-[10px] text-ant-text-muted hover:text-rose-400 font-serif px-1.5 py-0.5 rounded hover:bg-white/5 transition ml-auto"
          >
            Clear all
          </button>
        </div>
      {/if}

      <textarea
        bind:this={textareaEl}
        bind:value={text}
        oninput={handleInput}
        onkeydown={handleKeyDown}
        onpaste={handlePaste}
        placeholder={isWorking ? "Grok is executing... (type to queue or steer)" : "Ask Grok anything, command tools, or inspect code... (Enter to send, Shift+Enter for newline)"}
        rows={1}
        class="w-full bg-transparent text-[13.5px] text-ant-text placeholder:text-ant-text-muted placeholder:font-serif placeholder:text-xs px-3.5 pt-3 pb-2 outline-none resize-none min-h-[44px] max-h-[200px] leading-relaxed block scrollbar-thin font-serif"
      ></textarea>

      <!-- Compact Reference-Style Prompt Box Bottom Bar -->
      <div class="flex items-center justify-between px-2.5 py-1.5 border-t border-white/5 bg-ant-bg/60 text-xs select-none relative z-40 rounded-b-xl">
        <!-- Left Action Cluster -->
        <div class="flex items-center space-x-1.5">
          <!-- Plus (+) Attachment Trigger Menu with Click Outside Support -->
          <div class="relative" bind:this={plusMenuContainerEl}>
            <button
              type="button"
              class="w-6 h-6 rounded-md flex items-center justify-center text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-secondary transition border border-transparent hover:border-white/5"
              onclick={(e) => { e.stopPropagation(); isPlusMenuOpen = !isPlusMenuOpen; }}
              title="Add files, images, or snapshot"
            >
              <Plus size={15} />
            </button>

            {#if isPlusMenuOpen}
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div
                class="absolute bottom-full left-0 mb-1.5 min-w-[220px] w-max bg-ant-bg border border-white/10 rounded-lg shadow-xl py-1 z-50 text-xs backdrop-blur-md animate-in fade-in zoom-in-95 duration-100"
                onclick={(e) => e.stopPropagation()}
                onkeydown={(e) => e.key === 'Escape' && (isPlusMenuOpen = false)}
              >
                <!-- File / Document Picker Trigger (Images, MD, PDF, Code) -->
                <button
                  type="button"
                  class="w-full px-3 py-1.5 flex items-center space-x-2 text-left text-ant-text hover:bg-ant-primary/10 hover:text-ant-primary transition whitespace-nowrap"
                  onclick={() => { fileInputEl?.click(); isPlusMenuOpen = false; }}
                >
                  <Paperclip size={13} class="text-ant-text-muted shrink-0" />
                  <span class="whitespace-nowrap">Attach File (Image, MD, PDF)</span>
                </button>

                <!-- Snapshot Screen Trigger -->
                <button
                  type="button"
                  class="w-full px-3 py-1.5 flex items-center space-x-2 text-left text-ant-text hover:bg-ant-primary/10 hover:text-ant-primary transition whitespace-nowrap"
                  onclick={handleTakeSnapshot}
                  disabled={isTakingSnapshot}
                >
                  {#if isTakingSnapshot}
                    <Loader2 size={13} class="animate-spin text-ant-primary shrink-0" />
                    <span class="whitespace-nowrap">Capturing...</span>
                  {:else}
                    <Camera size={13} class="text-ant-primary shrink-0" />
                    <span class="whitespace-nowrap">Take Screen Snapshot</span>
                  {/if}
                </button>

                {#if onOpenSkillsCatalog}
                  <div class="h-px bg-ant-border-secondary/50 my-1"></div>
                  <button
                    type="button"
                    class="w-full px-3 py-1.5 flex items-center space-x-2 text-left text-ant-text hover:bg-ant-primary/10 hover:text-ant-primary transition whitespace-nowrap"
                    onclick={() => { onOpenSkillsCatalog(); isPlusMenuOpen = false; }}
                  >
                    <Sparkles size={13} class="text-ant-warning shrink-0" />
                    <span class="whitespace-nowrap">Browse Skills Hub</span>
                  </button>
                {/if}
              </div>
            {/if}
          </div>

          <!-- Hidden File Input (Accepts Images, Markdown, PDFs, Code Files, Text) -->
          <input
            bind:this={fileInputEl}
            type="file"
            accept="image/*,.md,.markdown,.txt,.pdf,.json,.ts,.js,.go,.py,.rs,.html,.css,.yaml,.yml"
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
            lastTurnInput={activeSessionTokens.lastTurnInput}
            lastTurnOutput={activeSessionTokens.lastTurnOutput}
            lastTurnCacheRead={activeSessionTokens.lastTurnCacheRead}
            lastTurnReasoning={activeSessionTokens.lastTurnReasoning}
            lastTurnModelCalls={activeSessionTokens.lastTurnModelCalls}
            totalInput={activeSessionTokens.totalInput}
            totalOutput={activeSessionTokens.totalOutput}
            totalCacheRead={activeSessionTokens.totalCache}
            isCompacting={sessionStore.isCompacting}
            onCompact={() => {
              sessionStore.compactActiveSession();
            }}
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
            <div class="flex items-center space-x-1.5 px-2.5 py-1 rounded-lg bg-[#1a1a1d] text-zinc-300 text-[11px] font-mono shadow-inner">
              <Timer size={11} class="animate-spin text-indigo-400" />
              <span class="font-medium">{formatElapsed(elapsedMs)}</span>
            </div>

            <!-- Stop/Cancel Execution Button -->
            <button
              type="button"
              class="w-7 h-7 rounded-lg flex items-center justify-center bg-rose-600/90 text-white hover:bg-rose-500 transition shadow-sm"
              onclick={onCancel}
              title="Stop turn execution"
            >
              <Square size={11} class="fill-current" />
            </button>
          {:else}
            <!-- Send Button (Blue circle / rounded up arrow) -->
            <button
              type="button"
              class="w-7 h-7 rounded-lg flex items-center justify-center bg-ant-primary text-white hover:bg-ant-primary-hover disabled:opacity-30 disabled:hover:bg-ant-primary disabled:cursor-not-allowed transition shadow-sm"
              disabled={(!text.trim() && attachedImages.length === 0 && attachedFiles.length === 0) || disabled}
              onclick={handleSubmit}
              title="Send to Grok (Enter)"
            >
              <ArrowUp size={14} stroke-width="2.5" />
            </button>
          {/if}
        </div>
      </div>
    </div>
  </div>
</div>

