<script lang="ts">
  import { slide } from 'svelte/transition';
  import type { QueuedPrompt } from '$lib/stores/session.svelte';
  import {
    ChevronUp,
    ChevronDown,
    Trash2,
    Edit3,
    Zap,
    GripVertical,
    FileText,
    FileCode,
    Image as ImageIcon,
    Play,
    AlertTriangle,
    PauseCircle
  } from 'lucide-svelte';

  interface Props {
    queue: QueuedPrompt[];
    isPaused?: boolean;
    onResumeQueue?: () => void;
    onClearQueue?: () => void;
    onSteer: (prompt: QueuedPrompt) => void;
    onReorder: (fromIdx: number, toIdx: number) => void;
    onRemove: (id: string) => void;
    onEdit: (prompt: QueuedPrompt) => void;
  }

  let {
    queue = [],
    isPaused = false,
    onResumeQueue,
    onClearQueue,
    onSteer,
    onReorder,
    onRemove,
    onEdit
  }: Props = $props();

  let isExpanded = $state(true);
  let draggedIdx = $state<number | null>(null);
  let dragOverIdx = $state<number | null>(null);

  function handleDragStart(e: DragEvent, idx: number) {
    draggedIdx = idx;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', String(idx));
    }
  }

  function handleDragOver(e: DragEvent, idx: number) {
    e.preventDefault();
    if (e.dataTransfer) {
      e.dataTransfer.dropEffect = 'move';
    }
    dragOverIdx = idx;
  }

  function handleDragLeave() {
    dragOverIdx = null;
  }

  function handleDrop(e: DragEvent, dropIdx: number) {
    e.preventDefault();
    if (draggedIdx !== null && draggedIdx !== dropIdx) {
      onReorder(draggedIdx, dropIdx);
    }
    draggedIdx = null;
    dragOverIdx = null;
  }

  function handleDragEnd() {
    draggedIdx = null;
    dragOverIdx = null;
  }

  function isCodeFile(filename: string): boolean {
    return filename.match(/\.(ts|js|jsx|tsx|go|py|rs|html|css|json|yaml|yml|sh|md)$/i) !== null;
  }
</script>

{#if queue.length > 0}
  <div class="mb-2 bg-ant-bg-secondary dark:bg-[#121316] border border-ant-border-secondary dark:border-white/[0.07] rounded-xl overflow-hidden shadow-md dark:shadow-2xl transition-all">
    <!-- Header Summary (No redundant global steer button, pure English, subtle pill) -->
    <div
      class="px-3.5 py-2.5 flex items-center justify-between bg-ant-bg/60 dark:bg-white/[0.015] hover:bg-ant-bg dark:hover:bg-white/[0.03] select-none cursor-pointer transition border-b border-ant-border-secondary dark:border-white/[0.04]"
      onclick={() => (isExpanded = !isExpanded)}
      onkeydown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          e.preventDefault();
          isExpanded = !isExpanded;
        }
      }}
      role="button"
      tabindex="0"
    >
      <div class="flex items-center space-x-2.5">
        <div class="flex items-center justify-center w-5 h-5 rounded-full bg-ant-bg-tertiary dark:bg-white/[0.08] text-ant-text font-mono text-[10.5px] font-semibold border border-ant-border-secondary dark:border-white/10">
          {queue.length}
        </div>
        <div class="flex items-baseline space-x-2">
          <span class="text-xs font-serif font-medium text-ant-text tracking-tight">
            {queue.length === 1 ? 'Queued Prompt' : `${queue.length} Queued Prompts`}
          </span>
          {#if isPaused}
            <span class="text-[11px] text-amber-500 font-sans font-medium flex items-center gap-1">
              <PauseCircle size={11} />
              Antrean Dijeda
            </span>
          {:else}
            <span class="text-[11px] text-ant-text-muted font-sans hidden sm:inline">
              (Executes automatically once current turn finishes)
            </span>
          {/if}
        </div>
      </div>

      <div class="flex items-center space-x-1.5 text-ant-text-muted hover:text-ant-text transition">
        <span class="text-[11px] font-sans text-ant-text-muted">
          {isExpanded ? 'Hide' : 'Show list'}
        </span>
        {#if isExpanded}
          <ChevronUp size={14} />
        {:else}
          <ChevronDown size={14} />
        {/if}
      </div>
    </div>

    {#if isPaused}
      <div class="px-3.5 py-2 bg-amber-500/10 dark:bg-amber-950/40 border-b border-amber-500/20 flex flex-wrap items-center justify-between gap-2 text-xs">
        <div class="flex items-center space-x-2 text-amber-600 dark:text-amber-400 font-sans">
          <AlertTriangle size={13} class="text-amber-500 flex-shrink-0" />
          <span class="font-medium text-[11.5px]">Antrean dijeda demi keamanan. Prompt tersimpan aman dan tidak dieksekusi otomatis.</span>
        </div>
        <div class="flex items-center space-x-2">
          {#if onResumeQueue}
            <button
              type="button"
              onclick={(e) => { e.stopPropagation(); onResumeQueue(); }}
              class="flex items-center space-x-1 px-2.5 py-1 rounded bg-amber-500 hover:bg-amber-600 text-white font-medium text-[11px] transition shadow-sm"
              title="Lanjutkan eksekusi antrean prompt"
            >
              <Play size={10} class="fill-current" />
              <span>Lanjutkan Antrean</span>
            </button>
          {/if}
          {#if onClearQueue}
            <button
              type="button"
              onclick={(e) => { e.stopPropagation(); onClearQueue(); }}
              class="flex items-center space-x-1 px-2 py-1 rounded text-ant-text-muted hover:text-rose-500 hover:bg-rose-500/10 text-[11px] transition"
              title="Hapus semua prompt di antrean"
            >
              <Trash2 size={11} />
              <span>Hapus Semua</span>
            </button>
          {/if}
        </div>
      </div>
    {/if}

    <!-- Draggable Queue List -->
    {#if isExpanded}
      <div
        transition:slide={{ duration: 160 }}
        class="p-2 space-y-1.5 max-h-64 overflow-y-auto ant-scrollbar"
      >
        {#each queue as item, idx (item.id)}
          <div
            draggable="true"
            ondragstart={(e) => handleDragStart(e, idx)}
            ondragover={(e) => handleDragOver(e, idx)}
            ondragleave={handleDragLeave}
            ondrop={(e) => handleDrop(e, idx)}
            ondragend={handleDragEnd}
            class="flex flex-col p-2.5 rounded-lg bg-ant-bg dark:bg-white/[0.02] border transition-all group relative {
              draggedIdx === idx
                ? 'opacity-30 border-dashed border-ant-border-secondary dark:border-white/20'
                : dragOverIdx === idx
                  ? 'border-blue-500/60 bg-blue-500/[0.05]'
                  : 'border-ant-border-secondary dark:border-white/[0.05] hover:border-blue-500/30 dark:hover:border-white/[0.12] hover:bg-ant-bg-tertiary/40 dark:hover:bg-white/[0.035]'
            }"
          >
            <!-- Card Top: Drag Handle, Number, Model Badge, Individual Steer, Edit, Delete -->
            <div class="flex items-center justify-between pb-1.5 border-b border-ant-border-secondary dark:border-white/[0.03]">
              <div class="flex items-center space-x-2">
                <!-- Drag Grip Handle -->
                <div
                  class="cursor-grab active:cursor-grabbing text-ant-text-muted/60 group-hover:text-ant-text-secondary p-0.5 rounded hover:bg-ant-bg-tertiary dark:hover:bg-white/5 transition"
                  title="Drag to reorder"
                >
                  <GripVertical size={13} />
                </div>
                <span class="text-[10px] font-mono text-ant-text-muted">#{idx + 1}</span>
                {#if item.model}
                  <span class="px-1.5 py-0.2 rounded bg-ant-bg-tertiary dark:bg-white/[0.04] text-ant-text-secondary text-[10px] font-mono border border-ant-border-secondary dark:border-white/5">
                    {item.model}
                  </span>
                {/if}
              </div>

              <!-- Action Buttons -->
              <div class="flex items-center space-x-1.5">
                <!-- Steer button (Subtle & elegant, cancels active execution and runs this item immediately) -->
                <button
                  type="button"
                  onclick={() => onSteer(item)}
                  class="flex items-center space-x-1 px-2 py-0.5 rounded bg-amber-500/15 hover:bg-amber-500/25 text-amber-700 dark:text-amber-300 border border-amber-500/30 text-[10.5px] font-serif transition"
                  title="Interrupt current execution and run this prompt immediately"
                >
                  <Zap size={11} class="fill-current text-amber-600 dark:text-amber-400" />
                  <span>Steer</span>
                </button>

                <!-- Edit Button: Re-loads item and attachments back into Prompt Box -->
                <button
                  type="button"
                  onclick={() => onEdit(item)}
                  class="flex items-center space-x-1 px-1.5 py-0.5 rounded text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-tertiary dark:hover:bg-white/5 text-[10.5px] font-serif transition"
                  title="Move prompt back to input box for editing"
                >
                  <Edit3 size={12} />
                  <span class="hidden sm:inline">Edit</span>
                </button>

                <!-- Delete Button -->
                <button
                  type="button"
                  onclick={() => onRemove(item.id)}
                  class="p-1 rounded text-ant-text-muted hover:text-rose-500 hover:bg-rose-500/10 transition"
                  title="Delete from queue"
                >
                  <Trash2 size={12} />
                </button>
              </div>
            </div>

            <!-- Prompt Text -->
            <div class="py-1.5 text-xs font-serif text-ant-text leading-relaxed select-text line-clamp-3">
              {item.text || '(Attachment only prompt)'}
            </div>

            <!-- Attachment Thumbnails & Badges -->
            {#if (item.images && item.images.length > 0) || (item.attachments && item.attachments.length > 0)}
              <div class="pt-1.5 flex items-center gap-1.5 flex-wrap border-t border-ant-border-secondary dark:border-white/[0.03]">
                <!-- Image Thumbnails -->
                {#if item.images}
                  {#each item.images as img}
                    <div class="relative group/thumb inline-flex items-center rounded overflow-hidden border border-ant-border-secondary dark:border-white/10 bg-ant-bg-tertiary dark:bg-black/40">
                      {#if img.dataUrl}
                        <img
                          src={img.dataUrl}
                          alt="Queued reference"
                          class="w-7 h-7 object-cover hover:scale-105 transition-transform"
                        />
                      {:else}
                        <div class="w-7 h-7 flex items-center justify-center bg-ant-bg-tertiary dark:bg-white/5">
                          <ImageIcon size={12} class="text-blue-500 dark:text-blue-400" />
                        </div>
                      {/if}
                    </div>
                  {/each}
                {/if}

                <!-- File Chips -->
                {#if item.attachments}
                  {#each item.attachments as att}
                    <div class="inline-flex items-center space-x-1 px-1.5 py-0.5 rounded bg-ant-bg-tertiary/70 dark:bg-white/[0.03] border border-ant-border-secondary dark:border-white/5 text-[10px] text-ant-text-secondary font-mono max-w-[150px]">
                      {#if att.isImage}
                        {#if att.dataUrl}
                          <img src={att.dataUrl} alt="" class="w-3.5 h-3.5 rounded object-cover" />
                        {:else}
                          <ImageIcon size={10} class="text-blue-500 dark:text-blue-400" />
                        {/if}
                      {:else if isCodeFile(att.name)}
                        <FileCode size={10} class="text-amber-600 dark:text-amber-400/80" />
                      {:else}
                        <FileText size={10} class="text-purple-600 dark:text-purple-400/80" />
                      {/if}
                      <span class="truncate">{att.name}</span>
                    </div>
                  {/each}
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
{/if}
