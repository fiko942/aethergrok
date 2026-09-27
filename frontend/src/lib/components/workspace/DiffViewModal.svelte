<script lang="ts">
  import { X, Copy, Check, Plus, Minus, FileCode } from 'lucide-svelte';

  let {
    isOpen = false,
    filePath = '',
    workspacePath = '',
    onClose,
  }: {
    isOpen: boolean;
    filePath: string;
    workspacePath: string;
    onClose: () => void;
  } = $props();

  let diffText = $state('');
  let isLoading = $state(false);
  let errorMsg = $state('');
  let copied = $state(false);

  const fileName = $derived(filePath ? filePath.split('/').pop() || filePath : '');

  async function loadDiff() {
    if (!filePath || !workspacePath) return;
    isLoading = true;
    errorMsg = '';
    diffText = '';

    try {
      const win = window as any;
      if (win.go?.main?.App?.GetWorkspaceFileDiff) {
        const text = await win.go.main.App.GetWorkspaceFileDiff(workspacePath, filePath);
        diffText = text;
      } else {
        errorMsg = 'Backend bridge not available';
      }
    } catch (err: any) {
      errorMsg = err?.toString() || 'Failed to load diff';
    } finally {
      isLoading = false;
    }
  }

  $effect(() => {
    if (isOpen && filePath) {
      loadDiff();
    }
  });

  function copyDiff() {
    navigator.clipboard.writeText(diffText);
    copied = true;
    setTimeout(() => (copied = false), 1500);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onClose();
    }
  }

  interface DiffLine {
    type: 'add' | 'del' | 'header' | 'hunk' | 'normal' | 'binary';
    content: string;
    oldLine?: number;
    newLine?: number;
  }

  const parsedLines = $derived.by<DiffLine[]>(() => {
    if (!diffText) return [];
    const rawLines = diffText.split('\n');
    const result: DiffLine[] = [];
    let oldNum = 0;
    let newNum = 0;

    for (const line of rawLines) {
      if (
        line.startsWith('diff --git') ||
        line.startsWith('index ') ||
        line.startsWith('--- ') ||
        line.startsWith('+++ ') ||
        line.startsWith('new file mode') ||
        line.startsWith('deleted file mode') ||
        line.startsWith('similarity index') ||
        line.startsWith('rename from') ||
        line.startsWith('rename to')
      ) {
        result.push({ type: 'header', content: line });
      } else if (line.startsWith('Binary files ') || line.includes('differ')) {
        result.push({ type: 'binary', content: line });
      } else if (line.startsWith('@@')) {
        result.push({ type: 'hunk', content: line });
        const match = line.match(/@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/);
        if (match) {
          oldNum = parseInt(match[1], 10);
          newNum = parseInt(match[2], 10);
        }
      } else if (line.startsWith('+')) {
        // Strip the leading '+' character for clean code display without '+' prefix
        const cleanContent = line.slice(1);
        result.push({ type: 'add', content: cleanContent, newLine: newNum++ });
      } else if (line.startsWith('-')) {
        // Strip the leading '-' character for clean code display without '-' prefix
        const cleanContent = line.slice(1);
        result.push({ type: 'del', content: cleanContent, oldLine: oldNum++ });
      } else {
        if (line.length > 0) {
          // In standard diff, context lines start with a single space ' '
          const cleanContent = line.startsWith(' ') ? line.slice(1) : line;
          result.push({ type: 'normal', content: cleanContent, oldLine: oldNum++, newLine: newNum++ });
        }
      }
    }
    return result;
  });

  const stats = $derived.by(() => {
    let add = 0;
    let del = 0;
    for (const l of parsedLines) {
      if (l.type === 'add') add++;
      if (l.type === 'del') del++;
    }
    return { add, del };
  });
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
  <div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 dark:bg-black/80 backdrop-blur-md animate-in fade-in duration-150">
    <div class="bg-ant-bg border border-ant-border rounded-xl shadow-2xl w-full max-w-4xl h-[85vh] flex flex-col overflow-hidden text-ant-text relative z-[101]">
      <!-- Modal Header -->
      <div class="px-4 py-3 border-b border-ant-border flex items-center justify-between bg-ant-bg-secondary shrink-0">
        <div class="flex items-center gap-2 min-w-0">
          <FileCode class="w-4 h-4 text-ant-primary" />
          <span class="font-medium text-sm text-ant-text truncate">{fileName}</span>
          <span class="text-xs text-ant-text-muted font-mono truncate hidden sm:inline">{filePath}</span>
          <div class="flex items-center gap-1.5 ml-2 text-xs font-mono font-medium">
            <span class="text-emerald-600 dark:text-emerald-400 flex items-center"><Plus class="w-3 h-3" />{stats.add}</span>
            <span class="text-rose-600 dark:text-rose-400 flex items-center"><Minus class="w-3 h-3" />{stats.del}</span>
          </div>
        </div>

        <div class="flex items-center gap-2 shrink-0">
          <button
            onclick={copyDiff}
            title="Copy Diff"
            class="flex items-center gap-1.5 px-2.5 py-1 text-xs rounded-lg border border-ant-border bg-ant-bg-tertiary text-ant-text hover:border-ant-primary/40 hover:text-ant-primary transition-colors"
          >
            {#if copied}
              <Check class="w-3.5 h-3.5 text-emerald-500" />
              <span>Copied</span>
            {:else}
              <Copy class="w-3.5 h-3.5 text-ant-text-muted" />
              <span>Copy</span>
            {/if}
          </button>

          <button
            onclick={onClose}
            title="Close (Esc)"
            class="p-1 rounded-lg text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-tertiary transition-colors"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Diff Container -->
      <div class="flex-1 overflow-auto bg-ant-bg custom-scrollbar font-mono text-xs leading-5">
        {#if isLoading}
          <div class="flex items-center justify-center h-full text-ant-text-muted">
            <span class="animate-pulse">Loading diff...</span>
          </div>
        {:else if errorMsg}
          <div class="flex items-center justify-center h-full text-rose-500 text-xs font-mono p-4">
            {errorMsg}
          </div>
        {:else if parsedLines.length === 0}
          <div class="flex items-center justify-center h-full text-ant-text-muted italic">
            No uncommitted changes detected for this file.
          </div>
        {:else}
          <table class="w-full border-collapse">
            <tbody>
              {#each parsedLines as line}
                {#if line.type === 'header'}
                  <tr class="bg-ant-bg-tertiary/60 text-ant-text-muted select-none">
                    <td colspan="3" class="px-3 py-0.5 text-[11px] font-mono">{line.content}</td>
                  </tr>
                {:else if line.type === 'hunk'}
                  <tr class="bg-sky-500/10 text-sky-700 dark:text-sky-300 select-none border-y border-sky-500/20">
                    <td colspan="3" class="px-3 py-1 font-mono text-[11px] font-semibold">{line.content}</td>
                  </tr>
                {:else if line.type === 'binary'}
                  <tr class="bg-amber-500/10 text-amber-700 dark:text-amber-300 select-none border-y border-amber-500/20">
                    <td colspan="3" class="px-4 py-3 font-mono text-xs">
                      <div class="flex items-center gap-2">
                        <span class="px-1.5 py-0.5 rounded bg-amber-500/20 text-amber-600 dark:text-amber-400 font-semibold text-[10px]">BINARY</span>
                        <span>{line.content}</span>
                      </div>
                    </td>
                  </tr>
                {:else if line.type === 'add'}
                  <tr class="bg-emerald-500/10 hover:bg-emerald-500/15 text-emerald-800 dark:text-emerald-200 transition-colors">
                    <td class="w-10 px-2 py-0 text-right select-none text-ant-text-muted/40 border-r border-ant-border/40"></td>
                    <td class="w-10 px-2 py-0 text-right select-none text-emerald-600 dark:text-emerald-400 font-semibold border-r border-ant-border/40">{line.newLine}</td>
                    <td class="px-3 py-0 whitespace-pre">{line.content}</td>
                  </tr>
                {:else if line.type === 'del'}
                  <tr class="bg-rose-500/10 hover:bg-rose-500/15 text-rose-800 dark:text-rose-200 transition-colors">
                    <td class="w-10 px-2 py-0 text-right select-none text-rose-600 dark:text-rose-400 font-semibold border-r border-ant-border/40">{line.oldLine}</td>
                    <td class="w-10 px-2 py-0 text-right select-none text-ant-text-muted/40 border-r border-ant-border/40"></td>
                    <td class="px-3 py-0 whitespace-pre">{line.content}</td>
                  </tr>
                {:else}
                  <tr class="text-ant-text-secondary hover:bg-ant-bg-secondary/60 transition-colors">
                    <td class="w-10 px-2 py-0 text-right select-none text-ant-text-muted/60 border-r border-ant-border/40">{line.oldLine || ''}</td>
                    <td class="w-10 px-2 py-0 text-right select-none text-ant-text-muted/60 border-r border-ant-border/40">{line.newLine || ''}</td>
                    <td class="px-3 py-0 whitespace-pre">{line.content}</td>
                  </tr>
                {/if}
              {/each}
            </tbody>
          </table>
        {/if}
      </div>

      <!-- Footer -->
      <div class="px-4 py-2 border-t border-ant-border bg-ant-bg-secondary flex items-center justify-between text-[11px] text-ant-text-muted font-mono shrink-0">
        <span>Git Diff Inspector</span>
        <span>Green = Additions &bull; Red = Deletions</span>
      </div>
    </div>
  </div>
{/if}

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 6px;
    height: 6px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: var(--ant-border, #27272a);
    border-radius: 3px;
  }
</style>
