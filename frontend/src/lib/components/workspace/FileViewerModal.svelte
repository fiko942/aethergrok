<script lang="ts">
  import { X, Code, Eye, Copy, Check, FileText, AlertTriangle, FileWarning, ArrowRight } from 'lucide-svelte';
  import { marked } from 'marked';
  import FileIcon from './FileIcon.svelte';
  import { highlightCode } from '$lib/utils/codeHighlighter';

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
  let isLargeFilePending = $state(false);
  let largeFileSize = $state(0);
  let viewMode = $state<'code' | 'preview'>('code'); // for markdown
  let copied = $state(false);

  const isMarkdown = $derived(filePath.toLowerCase().endsWith('.md') || filePath.toLowerCase().endsWith('.markdown'));
  const fileName = $derived(filePath ? filePath.split(/[/\\]/).pop() || filePath : '');
  const ext = $derived(fileName.includes('.') ? '.' + fileName.split('.').pop() : '');

  // Configure marked for full GitHub Flavored Markdown (tables, lists, breaks, codeblocks, checklists)
  marked.setOptions({
    gfm: true,
    breaks: true,
  });

  const renderedMarkdownHtml = $derived.by(() => {
    if (!content) return '';
    try {
      return marked.parse(content) as string;
    } catch {
      return content;
    }
  });

  function formatBytes(bytes: number): string {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  async function loadFileContent(allowLarge: boolean = false) {
    if (!filePath) return;
    isLoading = true;
    errorMsg = '';
    if (!allowLarge) {
      content = '';
      isLargeFilePending = false;
      largeFileSize = 0;
    }

    try {
      const win = window as any;
      if (win.go?.main?.App?.ReadWorkspaceFileContent) {
        const text = await win.go.main.App.ReadWorkspaceFileContent(workspacePath || '', filePath, allowLarge);
        content = text;
        isLargeFilePending = false;
      } else {
        errorMsg = 'Backend bridge not available';
      }
    } catch (err: any) {
      const errStr = err?.toString() || '';
      if (errStr.includes('LARGE_FILE_CONFIRM_REQUIRED:')) {
        const match = errStr.match(/LARGE_FILE_CONFIRM_REQUIRED:(\d+)/);
        if (match) {
          largeFileSize = parseInt(match[1], 10);
        }
        isLargeFilePending = true;
      } else {
        errorMsg = errStr || 'Failed to read file';
      }
    } finally {
      isLoading = false;
    }
  }

  $effect(() => {
    if (isOpen && filePath) {
      isLargeFilePending = false;
      largeFileSize = 0;
      loadFileContent(false);
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

  // Professional syntax highlighter powered by PrismJS with extension mapping
  const highlightedCodeHtml = $derived.by(() => {
    if (!content) return '';
    const cleanExt = (ext.startsWith('.') ? ext.slice(1) : ext).toLowerCase();
    return highlightCode(content, cleanExt);
  });
</script>

<svelte:window onkeydown={handleKeydown} />

{#if isOpen}
  <div class="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-black/60 dark:bg-black/80 backdrop-blur-md animate-in fade-in duration-150">
    <div
      class="bg-ant-bg border border-ant-border-secondary dark:border-white/5 rounded-xl shadow-2xl w-full max-w-4xl h-[85vh] flex flex-col overflow-hidden text-ant-text relative z-[101]"
    >
      <!-- Header -->
      <div class="px-4 py-3 border-b border-ant-border-secondary dark:border-white/5 flex items-center justify-between bg-ant-bg-secondary shrink-0">
        <div class="flex items-center gap-2 min-w-0">
          <FileIcon ext={ext} name={fileName} class="w-4 h-4" />
          <span class="font-medium text-sm text-ant-text truncate">{fileName}</span>
          <span class="text-xs text-ant-text-muted font-mono truncate hidden sm:inline">{filePath}</span>
        </div>

        <div class="flex items-center gap-2 shrink-0">
          {#if isMarkdown && !isLargeFilePending && !errorMsg}
            <div class="flex items-center bg-ant-bg-tertiary border border-ant-border-secondary dark:border-white/5 rounded-lg p-0.5 text-xs">
              <button
                onclick={() => (viewMode = 'preview')}
                class="flex items-center gap-1 px-2.5 py-1 rounded-md transition-colors {viewMode === 'preview' ? 'bg-ant-bg text-ant-text shadow-sm' : 'text-ant-text-muted hover:text-ant-text'}"
              >
                <Eye class="w-3.5 h-3.5" />
                <span>Preview</span>
              </button>
              <button
                onclick={() => (viewMode = 'code')}
                class="flex items-center gap-1 px-2.5 py-1 rounded-md transition-colors {viewMode === 'code' ? 'bg-ant-bg text-ant-text shadow-sm' : 'text-ant-text-muted hover:text-ant-text'}"
              >
                <Code class="w-3.5 h-3.5" />
                <span>Raw</span>
              </button>
            </div>
          {/if}

          {#if !isLargeFilePending && content}
            <button
              onclick={copyContent}
              title="Copy Content"
              class="flex items-center gap-1.5 px-2.5 py-1 text-xs rounded-lg border border-ant-border-secondary dark:border-white/5 bg-ant-bg-tertiary text-ant-text hover:border-ant-primary/40 hover:text-ant-primary transition-colors"
            >
              {#if copied}
                <Check class="w-3.5 h-3.5 text-emerald-500" />
                <span>Copied</span>
              {:else}
                <Copy class="w-3.5 h-3.5 text-ant-text-muted" />
                <span>Copy</span>
              {/if}
            </button>
          {/if}

          <button
            onclick={onClose}
            title="Close (Esc)"
            class="p-1 rounded-lg text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-tertiary transition-colors"
          >
            <X class="w-4 h-4" />
          </button>
        </div>
      </div>

      <!-- Content View -->
      <div class="flex-1 overflow-auto bg-ant-bg custom-scrollbar p-0">
        {#if isLoading}
          <div class="flex items-center justify-center h-full text-ant-text-muted">
            <span class="animate-pulse">Loading file content...</span>
          </div>
        {:else if isLargeFilePending}
          <!-- Large File Confirmation Box -->
          <div class="flex flex-col items-center justify-center h-full p-6 text-center max-w-lg mx-auto animate-in fade-in duration-200">
            <div class="w-12 h-12 rounded-2xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center mb-4 text-amber-500 shadow-lg shadow-amber-500/5">
              <FileWarning class="w-6 h-6" />
            </div>
            
            <h3 class="text-base font-semibold text-ant-text mb-1.5 font-sans">
              Ukuran Berkas Cukup Besar
            </h3>
            
            <p class="text-xs text-ant-text-secondary mb-6 leading-relaxed">
              Berkas <span class="font-mono text-ant-text font-semibold">{fileName}</span> berukuran <span class="text-amber-500 font-mono font-semibold">{formatBytes(largeFileSize)}</span>. Membuka berkas berukuran besar dapat membutuhkan waktu render lebih lama.
            </p>

            <div class="flex items-center gap-3">
              <button
                onclick={onClose}
                class="px-4 py-2 text-xs font-medium rounded-lg border border-ant-border-secondary dark:border-white/5 bg-ant-bg-secondary text-ant-text hover:bg-ant-bg-tertiary transition-colors"
              >
                Batal
              </button>

              <button
                onclick={() => loadFileContent(true)}
                class="px-4 py-2 text-xs font-medium rounded-lg bg-amber-600 hover:bg-amber-500 text-white shadow-lg shadow-amber-600/20 transition-all flex items-center gap-1.5"
              >
                <span>Tetap Buka Berkas</span>
                <ArrowRight class="w-3.5 h-3.5" />
              </button>
            </div>
          </div>
        {:else if errorMsg}
          <div class="flex items-center justify-center h-full text-rose-500 text-xs font-mono p-4">
            {errorMsg}
          </div>
        {:else if isMarkdown && viewMode === 'preview'}
          <div class="p-6 md:p-8 max-w-4xl mx-auto markdown-rendered-body select-text">
            {@html renderedMarkdownHtml}
          </div>
        {:else}
          <!-- Code View with Line Numbers & PrismJS Syntax Highlighting -->
          <div class="flex font-mono text-xs leading-5">
            <!-- Line numbers -->
            <div class="py-3 px-3 select-none text-ant-text-muted/60 text-right bg-ant-bg-secondary border-r border-ant-border-secondary dark:border-white/5 shrink-0 font-mono">
              {#each content.split('\n') as _, i}
                <div>{i + 1}</div>
              {/each}
            </div>

            <!-- PrismJS Highlighted Code Container -->
            <div class="py-3 px-4 overflow-x-auto whitespace-pre font-mono text-ant-text flex-1 leading-5">
              <code>{@html highlightedCodeHtml}</code>
            </div>
          </div>
        {/if}
      </div>

      <!-- Footer -->
      <div class="px-4 py-2 border-t border-ant-border-secondary dark:border-white/5 bg-ant-bg-secondary flex items-center justify-between text-[11px] text-ant-text-muted font-mono shrink-0">
        <span>Read-Only Mode</span>
        {#if !isLargeFilePending && content}
          <span>{content.split('\n').length} lines &bull; {formatBytes(content.length)}</span>
        {:else if isLargeFilePending}
          <span class="text-amber-500">{formatBytes(largeFileSize)}</span>
        {:else}
          <span>0 lines</span>
        {/if}
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

  /* Full GitHub-Flavored Markdown Typography Styling matching Theme */
  :global(.markdown-rendered-body) {
    color: var(--ant-text, #e4e4e7);
    font-family: ui-serif, Georgia, Cambria, "Times New Roman", Times, serif;
    font-size: 14px;
    line-height: 1.75;
  }

  :global(.markdown-rendered-body h1) {
    font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    font-size: 1.75rem;
    font-weight: 700;
    color: var(--ant-text, #fafafa);
    margin-top: 1.5rem;
    margin-bottom: 0.75rem;
    padding-bottom: 0.35rem;
    border-bottom: 1px solid var(--ant-border, rgba(255, 255, 255, 0.1));
  }

  :global(.markdown-rendered-body h2) {
    font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    font-size: 1.35rem;
    font-weight: 600;
    color: var(--ant-text, #fafafa);
    margin-top: 1.35rem;
    margin-bottom: 0.5rem;
    padding-bottom: 0.25rem;
    border-bottom: 1px solid var(--ant-border, rgba(255, 255, 255, 0.07));
  }

  :global(.markdown-rendered-body h3) {
    font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    font-size: 1.15rem;
    font-weight: 600;
    color: var(--ant-text, #f4f4f5);
    margin-top: 1.15rem;
    margin-bottom: 0.4rem;
  }

  :global(.markdown-rendered-body h4),
  :global(.markdown-rendered-body h5),
  :global(.markdown-rendered-body h6) {
    font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    font-size: 1rem;
    font-weight: 600;
    color: var(--ant-text, #e4e4e7);
    margin-top: 1rem;
    margin-bottom: 0.3rem;
  }

  :global(.markdown-rendered-body p) {
    margin-top: 0.6rem;
    margin-bottom: 0.85rem;
  }

  :global(.markdown-rendered-body strong) {
    color: var(--ant-text, #ffffff);
    font-weight: 600;
  }

  :global(.markdown-rendered-body em) {
    color: var(--ant-text-secondary, #d4d4d8);
  }

  :global(.markdown-rendered-body ul) {
    list-style-type: disc;
    padding-left: 1.5rem;
    margin-top: 0.4rem;
    margin-bottom: 0.85rem;
  }

  :global(.markdown-rendered-body ol) {
    list-style-type: decimal;
    padding-left: 1.5rem;
    margin-top: 0.4rem;
    margin-bottom: 0.85rem;
  }

  :global(.markdown-rendered-body li) {
    margin-top: 0.25rem;
    margin-bottom: 0.25rem;
  }

  :global(.markdown-rendered-body hr) {
    border: 0;
    border-top: 1px solid var(--ant-border, rgba(255, 255, 255, 0.12));
    margin-top: 1.5rem;
    margin-bottom: 1.5rem;
  }

  :global(.markdown-rendered-body blockquote) {
    border-left: 3px solid var(--ant-primary, #6366f1);
    padding-left: 1rem;
    margin-left: 0;
    margin-right: 0;
    margin-top: 0.75rem;
    margin-bottom: 0.75rem;
    color: var(--ant-text-secondary, #a1a1aa);
    background: var(--ant-primary-bg, rgba(99, 102, 241, 0.05));
    border-radius: 0 0.5rem 0.5rem 0;
    padding-top: 0.35rem;
    padding-bottom: 0.35rem;
  }

  :global(.markdown-rendered-body code) {
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
    font-size: 0.85em;
    background-color: var(--ant-bg-tertiary, rgba(255, 255, 255, 0.08));
    color: var(--ant-primary, #38bdf8);
    padding: 0.15rem 0.35rem;
    border-radius: 0.35rem;
    border: 1px solid var(--ant-border, rgba(255, 255, 255, 0.08));
  }

  :global(.markdown-rendered-body pre) {
    background-color: var(--ant-bg-secondary, #121215);
    border: 1px solid var(--ant-border, #27272a);
    border-radius: 0.5rem;
    padding: 0.85rem 1rem;
    overflow-x: auto;
    margin-top: 0.85rem;
    margin-bottom: 1rem;
  }

  :global(.markdown-rendered-body pre code) {
    background-color: transparent;
    border: none;
    padding: 0;
    color: var(--ant-text, #e4e4e7);
    font-size: 0.825rem;
    line-height: 1.6;
  }

  :global(.markdown-rendered-body table) {
    width: 100%;
    border-collapse: collapse;
    margin-top: 1rem;
    margin-bottom: 1.25rem;
    font-size: 0.85rem;
    font-family: ui-sans-serif, system-ui, -apple-system, sans-serif;
  }

  :global(.markdown-rendered-body th) {
    background-color: var(--ant-bg-secondary, #18181c);
    border: 1px solid var(--ant-border, #27272a);
    padding: 0.5rem 0.75rem;
    text-align: left;
    font-weight: 600;
    color: var(--ant-text, #f4f4f5);
  }

  :global(.markdown-rendered-body td) {
    border: 1px solid var(--ant-border, #27272a);
    padding: 0.45rem 0.75rem;
    color: var(--ant-text-secondary, #d4d4d8);
  }

  :global(.markdown-rendered-body tr:nth-child(even)) {
    background-color: var(--ant-bg-tertiary, rgba(255, 255, 255, 0.02));
  }

  :global(.markdown-rendered-body a) {
    color: var(--ant-primary, #60a5fa);
    text-decoration: underline;
    text-underline-offset: 3px;
  }

  :global(.markdown-rendered-body a:hover) {
    color: var(--ant-primary-hover, #93c5fd);
  }
</style>
