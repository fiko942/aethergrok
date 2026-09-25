<script lang="ts">
  import { X, Code, Eye, Copy, Check, FileText } from 'lucide-svelte';
  import FileIcon from './FileIcon.svelte';

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

  let content = $state('');
  let isLoading = $state(false);
  let errorMsg = $state('');
  let viewMode = $state<'code' | 'preview'>('code'); // for markdown
  let copied = $state(false);

  const isMarkdown = $derived(filePath.toLowerCase().endsWith('.md') || filePath.toLowerCase().endsWith('.markdown'));
  const fileName = $derived(filePath ? filePath.split('/').pop() || filePath : '');
  const ext = $derived(fileName.includes('.') ? '.' + fileName.split('.').pop() : '');

  async function loadFileContent() {
    if (!filePath || !workspacePath) return;
    isLoading = true;
    errorMsg = '';
    content = '';

    try {
      const win = window as any;
      if (win.go?.main?.App?.ReadWorkspaceFileContent) {
        const text = await win.go.main.App.ReadWorkspaceFileContent(workspacePath, filePath);
        content = text;
      } else {
        errorMsg = 'Backend bridge not available';
      }
    } catch (err: any) {
      errorMsg = err?.toString() || 'Failed to read file';
    } finally {
      isLoading = false;
    }
  }

  $effect(() => {
    if (isOpen && filePath) {
      loadFileContent();
      if (isMarkdown) {
        viewMode = 'preview';
      } else {
        viewMode = 'code';
      }
    }
  });

  function copyContent() {
    navigator.clipboard.writeText(content);
    copied = true;
    setTimeout(() => (copied = false), 1500);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === 'Escape') {
      onClose();
    }
  }

  // Basic syntax highlighter helper for read-only view
  function highlightLine(line: string, fileExt: string): string {
    const escaped = line
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;');

    if (fileExt === '.json') {
      return escaped
        .replace(/"(.*?)":/g, '<span class="text-sky-300 font-semibold">"$1"</span>:')
        .replace(/:\s*"(.*?)"/g, ': <span class="text-amber-200">"$1"</span>')
        .replace(/:\s*(\d+(\.\d+)?)/g, ': <span class="text-emerald-300">$1</span>')
        .replace(/:\s*(true|false|null)/g, ': <span class="text-purple-300 font-medium">$1</span>');
    }

    if (fileExt === '.ts' || fileExt === '.js' || fileExt === '.svelte' || fileExt === '.go') {
      return escaped
        .replace(/\b(const|let|var|function|import|export|from|return|if|else|switch|case|break|for|while|type|interface|class|struct|package|func)\b/g, '<span class="text-purple-400 font-medium">$1</span>')
        .replace(/\b(true|false|null|undefined|nil)\b/g, '<span class="text-amber-400">$1</span>')
        .replace(/\b(string|number|boolean|any|int|error)\b/g, '<span class="text-emerald-400">$1</span>')
        .replace(/(\/\/.*$)/g, '<span class="text-zinc-500 italic">$1</span>');
    }

    if (fileExt === '.sh' || fileExt === '.bash') {
      return escaped
        .replace(/(#.*$)/g, '<span class="text-zinc-500 italic">$1</span>')
        .replace(/\b(echo|export|if|then|else|fi|for|do|done|exit|cd|mkdir|rm)\b/g, '<span class="text-emerald-400 font-medium">$1</span>');
    }

    return escaped;
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm animate-in fade-in duration-150">
    <div
      class="bg-[#18181b] border border-[#27272a] rounded-xl shadow-2xl w-full max-w-4xl h-[85vh] flex flex-col overflow-hidden text-zinc-200"
    >
      <!-- Header -->
      <div class="px-4 py-3 border-b border-[#27272a] flex items-center justify-between bg-[#141416] shrink-0">
        <div class="flex items-center gap-2 min-w-0">
          <FileIcon ext={ext} name={fileName} class="w-4 h-4" />
          <span class="font-medium text-sm text-zinc-100 truncate">{fileName}</span>
          <span class="text-xs text-zinc-500 font-mono truncate hidden sm:inline">{filePath}</span>
        </div>

        <div class="flex items-center gap-2 shrink-0">
          {#if isMarkdown}
            <div class="flex items-center bg-[#1e1e22] border border-[#27272a] rounded-lg p-0.5 text-xs">
              <button
                onclick={() => (viewMode = 'preview')}
                class="flex items-center gap-1 px-2.5 py-1 rounded-md transition-colors {viewMode === 'preview' ? 'bg-zinc-800 text-zinc-100 shadow-sm' : 'text-zinc-400 hover:text-zinc-200'}"
              >
                <Eye class="w-3.5 h-3.5" />
                <span>Preview</span>
              </button>
              <button
                onclick={() => (viewMode = 'code')}
                class="flex items-center gap-1 px-2.5 py-1 rounded-md transition-colors {viewMode === 'code' ? 'bg-zinc-800 text-zinc-100 shadow-sm' : 'text-zinc-400 hover:text-zinc-200'}"
              >
                <Code class="w-3.5 h-3.5" />
                <span>Raw</span>
              </button>
            </div>
          {/if}

          <button
            onclick={copyContent}
            title="Copy Content"
            class="flex items-center gap-1 px-2.5 py-1 text-xs rounded-lg border border-[#27272a] bg-[#1e1e22] text-zinc-300 hover:bg-zinc-800 hover:text-white transition-colors"
          >
            {#if copied}
              <Check class="w-3.5 h-3.5 text-emerald-400" />
              <span>Copied</span>
            {:else}
              <Copy class="w-3.5 h-3.5 text-zinc-400" />
              <span>Copy</span>
            {/if}
          </button>

          <button
            onclick={onClose}
            title="Close (Esc)"
            class="p-1 rounded-lg text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800 transition-colors"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Content View -->
      <div class="flex-1 overflow-auto bg-[#0f0f11] custom-scrollbar p-0">
        {#if isLoading}
          <div class="flex items-center justify-center h-full text-zinc-500">
            <span class="animate-pulse">Loading file content...</span>
          </div>
        {:else if errorMsg}
          <div class="flex items-center justify-center h-full text-rose-400 text-xs font-mono p-4">
            {errorMsg}
          </div>
        {:else if isMarkdown && viewMode === 'preview'}
          <div class="p-6 max-w-3xl mx-auto prose prose-invert font-serif text-zinc-300 leading-relaxed">
            {#each content.split('\n\n') as block}
              {#if block.startsWith('# ')}
                <h1 class="text-2xl font-bold text-zinc-100 mt-4 mb-2">{block.replace('# ', '')}</h1>
              {:else if block.startsWith('## ')}
                <h2 class="text-xl font-semibold text-zinc-100 mt-3 mb-1.5">{block.replace('## ', '')}</h2>
              {:else if block.startsWith('### ')}
                <h3 class="text-lg font-medium text-zinc-200 mt-2 mb-1">{block.replace('### ', '')}</h3>
              {:else if block.startsWith('- ') || block.startsWith('* ')}
                <ul class="list-disc list-inside space-y-1 my-2">
                  {#each block.split('\n') as li}
                    <li>{li.replace(/^[-*]\s+/, '')}</li>
                  {/each}
                </ul>
              {:else}
                <p class="my-2">{block}</p>
              {/if}
            {/each}
          </div>
        {:else}
          <!-- Code View with Line Numbers -->
          <div class="flex font-mono text-xs leading-5">
            <!-- Line numbers -->
            <div class="py-3 px-3 select-none text-zinc-600 text-right bg-[#121214] border-r border-[#222226] shrink-0 font-mono">
              {#each content.split('\n') as _, i}
                <div>{i + 1}</div>
              {/each}
            </div>

            <!-- Code Lines -->
            <div class="py-3 px-4 overflow-x-auto whitespace-pre font-mono text-zinc-200 flex-1">
              {#each content.split('\n') as line}
                <div>{@html highlightLine(line, ext) || '&nbsp;'}</div>
              {/each}
            </div>
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="px-4 py-2 border-t border-[#27272a] bg-[#141416] flex items-center justify-between text-[11px] text-zinc-500 font-mono shrink-0">
        <span>Read-Only Mode</span>
        <span>{content.split('\n').length} lines &bull; {content.length} bytes</span>
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
    background: #27272a;
    border-radius: 3px;
  }
</style>
