<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { Keyboard, X, Check, RotateCcw, Sparkles } from 'lucide-svelte';

  interface Props {
    open: boolean;
    title?: string;
    currentShortcut: string;
    onSave: (newShortcut: string) => void;
    onCancel: () => void;
  }

  let { open = false, title = 'Record Snapshot Shortcut', currentShortcut = 'CmdOrCtrl+Shift+S', onSave, onCancel }: Props = $props();

  let recordedCode = $state<string>('');
  let recordedParts = $state<string[]>([]);
  let isListening = $state<boolean>(true);

  const isMac = typeof navigator !== 'undefined' && /Mac|iPod|iPhone|iPad/.test(navigator.platform || navigator.userAgent);

  // Helper to parse existing string into parts
  $effect(() => {
    if (open) {
      recordedParts = currentShortcut ? currentShortcut.split('+').map(p => p.trim()) : ['CmdOrCtrl', 'Shift', 'S'];
      recordedCode = currentShortcut;
      isListening = true;
    }
  });

  function normalizeKey(e: KeyboardEvent): { code: string; displayParts: string[] } {
    const parts: string[] = [];

    // Distinct standalone modifier handling when pressed individually
    if (e.key === 'Fn' || e.code === 'Fn' || e.key === 'Globe' || e.code === 'Globe') {
      return { code: 'Fn', displayParts: [isMac ? 'Fn / Globe' : 'Fn'] };
    }
    if (e.code === 'MetaLeft') {
      return { code: 'MetaLeft', displayParts: [isMac ? '⌘ Left' : 'Win Left'] };
    }
    if (e.code === 'MetaRight') {
      return { code: 'MetaRight', displayParts: [isMac ? '⌘ Right' : 'Win Right'] };
    }
    if (e.code === 'ShiftLeft') {
      return { code: 'ShiftLeft', displayParts: ['Left Shift'] };
    }
    if (e.code === 'ShiftRight') {
      return { code: 'ShiftRight', displayParts: ['Right Shift'] };
    }
    if (e.code === 'ControlLeft') {
      return { code: 'ControlLeft', displayParts: ['Ctrl Left'] };
    }
    if (e.code === 'ControlRight') {
      return { code: 'ControlRight', displayParts: ['Ctrl Right'] };
    }
    if (e.code === 'AltLeft') {
      return { code: 'AltLeft', displayParts: [isMac ? '⌥ Left' : 'Alt Left'] };
    }
    if (e.code === 'AltRight') {
      return { code: 'AltRight', displayParts: [isMac ? '⌥ Right' : 'Alt Right'] };
    }

    // Held modifier keys for combinations
    if (e.metaKey || e.ctrlKey) {
      parts.push(isMac ? 'Cmd' : 'Ctrl');
    }
    if (e.altKey) {
      parts.push(isMac ? 'Option' : 'Alt');
    }
    if (e.shiftKey) {
      parts.push('Shift');
    }

    // Non-modifier main key
    let mainKey = '';
    const modifierKeys = ['Meta', 'Control', 'Alt', 'Shift', 'OS'];
    if (!modifierKeys.includes(e.key)) {
      if (e.code === 'Space') mainKey = 'Space';
      else if (e.code === 'Slash') mainKey = '/';
      else if (e.code === 'Backslash') mainKey = '\\';
      else if (e.code === 'Delete') mainKey = 'Delete';
      else if (e.code === 'Backspace') mainKey = 'Backspace';
      else if (e.code === 'Escape') mainKey = 'Escape';
      else if (e.code === 'Tab') mainKey = 'Tab';
      else if (e.code === 'Enter') mainKey = 'Enter';
      else if (e.key && e.key.length === 1) mainKey = e.key.toUpperCase();
      else if (e.code) mainKey = e.code.replace('Key', '').replace('Digit', '');

      if (mainKey) {
        parts.push(mainKey);
      }
    }

    let codeString = '';
    if (parts.length > 0) {
      // Create unified accelerator format
      const isCmdCtrl = parts.includes('Cmd') || parts.includes('Ctrl');
      const hasShift = parts.includes('Shift');
      const hasAlt = parts.includes('Option') || parts.includes('Alt');
      const comboMain = parts.find(p => !['Cmd', 'Ctrl', 'Option', 'Alt', 'Shift'].includes(p));

      const standardParts: string[] = [];
      if (isCmdCtrl) standardParts.push('CmdOrCtrl');
      if (hasAlt) standardParts.push('Alt');
      if (hasShift) standardParts.push('Shift');
      if (comboMain) standardParts.push(comboMain);

      codeString = standardParts.join('+');
    } else {
      codeString = e.code || e.key;
    }

    return { code: codeString, displayParts: parts };
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (!open) return;

    // Prevent default system / browser shortcuts while recording
    e.preventDefault();
    e.stopPropagation();

    // If Escape pressed alone without modifiers, cancel dialog
    if (e.code === 'Escape' && !e.metaKey && !e.ctrlKey && !e.altKey && !e.shiftKey) {
      onCancel();
      return;
    }

    const { code, displayParts } = normalizeKey(e);
    if (displayParts.length > 0) {
      recordedParts = displayParts;
      recordedCode = code;
    }
  }

  // Use capture phase on window to guarantee intercepting keystrokes before any child or parent can block
  $effect(() => {
    if (open) {
      const listener = (e: KeyboardEvent) => handleKeyDown(e);
      window.addEventListener('keydown', listener, true);
      return () => {
        window.removeEventListener('keydown', listener, true);
      };
    }
  });

  function handleSave() {
    if (recordedCode) {
      onSave(recordedCode);
    } else {
      onSave(currentShortcut);
    }
  }

  function handleResetDefault() {
    recordedCode = 'CmdOrCtrl+Shift+S';
    recordedParts = ['CmdOrCtrl', 'Shift', 'S'];
  }
</script>

{#if open}
  <!-- Backdrop -->
  <div
    class="fixed inset-0 bg-black/60 backdrop-blur-sm z-[110] flex items-center justify-center p-4 animate-in fade-in duration-150"
    onclick={onCancel}
    role="presentation"
  >
    <!-- Modal Card (Removed harsh borders, stopPropagation only on click) -->
    <div
      role="dialog"
      aria-modal="true"
      tabindex="-1"
      class="w-full max-w-md bg-ant-bg-secondary rounded-2xl shadow-2xl p-6 space-y-5 animate-in zoom-in-95 duration-150 outline-none"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Header -->
      <div class="flex items-start justify-between">
        <div class="space-y-1">
          <div class="flex items-center gap-2">
            <div class="w-7 h-7 rounded-lg bg-ant-primary/10 flex items-center justify-center text-ant-primary">
              <Keyboard size={15} />
            </div>
            <h3 class="font-serif-display text-base font-semibold text-ant-text">
              {title}
            </h3>
          </div>
          <p class="font-serif text-xs text-ant-text-secondary leading-relaxed">
            Press any single key (e.g. <span class="font-mono text-ant-text">/</span>, <span class="font-mono text-ant-text">Delete</span>), modifier key (e.g. <span class="font-mono text-ant-text">Shift Left</span>, <span class="font-mono text-ant-text">⌘ Right</span>), or multi-key combination.
          </p>
        </div>

        <button
          type="button"
          onclick={onCancel}
          class="p-1 text-ant-text-muted hover:text-ant-text rounded-lg hover:bg-white/5 transition"
          aria-label="Close dialog"
        >
          <X size={15} />
        </button>
      </div>

      <!-- Live Keystroke Display Box -->
      <div class="p-6 rounded-xl bg-ant-bg flex flex-col items-center justify-center space-y-3 min-h-[110px] relative overflow-hidden">
        <div class="flex items-center justify-center flex-wrap gap-2">
          {#if recordedParts.length > 0}
            {#each recordedParts as part, i}
              <kbd class="px-3.5 py-1.5 text-sm font-mono font-semibold bg-ant-bg-secondary text-ant-primary rounded-lg shadow-sm">
                {part}
              </kbd>
              {#if i < recordedParts.length - 1}
                <span class="text-xs text-ant-text-muted font-mono font-bold">+</span>
              {/if}
            {/each}
          {:else}
            <span class="text-xs font-serif text-ant-text-muted italic animate-pulse">
              Press any key or combination on your keyboard...
            </span>
          {/if}
        </div>

        <!-- Listening Indicator -->
        <div class="flex items-center gap-1.5 text-[11px] font-mono text-ant-primary">
          <span class="relative flex h-2 w-2">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-ant-primary opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2 w-2 bg-ant-primary"></span>
          </span>
          <span>Listening for keystrokes...</span>
        </div>
      </div>

      <!-- Quick Presets & Tips -->
      <div class="space-y-2">
        <div class="text-[11px] font-medium text-ant-text-secondary flex items-center justify-between">
          <span>Popular Presets:</span>
          <span class="text-[10px] text-ant-text-muted">Click any preset to select</span>
        </div>
        <div class="flex items-center gap-1.5 flex-wrap">
          <button
            type="button"
            onclick={() => { recordedCode = '\\'; recordedParts = ['\\']; }}
            class="px-2.5 py-1 text-[11px] font-mono rounded-lg bg-ant-bg border border-ant-border-secondary hover:border-ant-primary/40 hover:text-ant-primary transition text-ant-text-secondary"
          >
            \ (Backslash)
          </button>
          <button
            type="button"
            onclick={() => { recordedCode = 'RightOption'; recordedParts = [isMac ? 'Right ⌥' : 'Right Alt']; }}
            class="px-2 py-1 text-[11px] font-mono rounded-lg bg-ant-bg border border-ant-border-secondary hover:border-ant-primary/40 hover:text-ant-primary transition text-ant-text-secondary"
          >
            {isMac ? 'Right ⌥' : 'Right Alt'}
          </button>
          <button
            type="button"
            onclick={() => { recordedCode = 'ShiftRight'; recordedParts = ['Right Shift']; }}
            class="px-2 py-1 text-[11px] font-mono rounded-lg bg-ant-bg border border-ant-border-secondary hover:border-ant-primary/40 hover:text-ant-primary transition text-ant-text-secondary"
          >
            Right Shift
          </button>
          <button
            type="button"
            onclick={() => { recordedCode = 'CmdOrCtrl+D'; recordedParts = [isMac ? 'Cmd' : 'Ctrl', 'D']; }}
            class="px-2 py-1 text-[11px] font-mono rounded-lg bg-ant-bg border border-ant-border-secondary hover:border-ant-primary/40 hover:text-ant-primary transition text-ant-text-secondary"
          >
            {isMac ? '⌘D' : 'Ctrl+D'}
          </button>
          <button
            type="button"
            onclick={() => { recordedCode = 'CmdOrCtrl+Shift+D'; recordedParts = [isMac ? 'Cmd' : 'Ctrl', 'Shift', 'D']; }}
            class="px-2 py-1 text-[11px] font-mono rounded-lg bg-ant-bg border border-ant-border-secondary hover:border-ant-primary/40 hover:text-ant-primary transition text-ant-text-secondary"
          >
            {isMac ? '⌘⇧D' : 'Ctrl+Shift+D'}
          </button>
        </div>
      </div>

      <!-- Quick Tips -->
      <div class="grid grid-cols-2 gap-2 text-[10.5px] font-serif text-ant-text-secondary">
        <div class="p-2 rounded-lg bg-ant-bg/60 space-y-0.5">
          <div class="text-ant-text font-sans font-medium">Single Modifier Support</div>
          <span>Supports Shift (Left/Right), Option, or Cmd standalone.</span>
        </div>
        <div class="p-2 rounded-lg bg-ant-bg/60 space-y-0.5">
          <div class="text-ant-text font-sans font-medium">Direct Key Trigger</div>
          <span>Supports single keys like '\', '/', 'Delete', or combo keys.</span>
        </div>
      </div>

      <!-- Actions -->
      <div class="pt-2 border-t border-white/5 flex items-center justify-between">
        <button
          type="button"
          onclick={handleResetDefault}
          class="text-xs text-ant-text-secondary hover:text-ant-text flex items-center gap-1.5 transition"
        >
          <RotateCcw size={12} /> Default (Cmd+Shift+S)
        </button>

        <div class="flex items-center space-x-2">
          <button
            type="button"
            onclick={onCancel}
            class="px-3.5 py-1.5 rounded-lg text-xs font-serif text-ant-text-secondary hover:text-ant-text hover:bg-white/5 transition"
          >
            Cancel
          </button>

          <button
            type="button"
            onclick={handleSave}
            class="px-4 py-1.5 rounded-lg text-xs font-serif font-medium text-white bg-ant-primary hover:bg-ant-primary/90 active:bg-ant-primary/80 shadow-sm transition flex items-center gap-1.5"
          >
            <Check size={13} /> Save Keybinding
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
