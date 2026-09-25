<script lang="ts">
  import { onMount } from 'svelte';
  import { Search, RefreshCw, ChevronRight, ChevronDown, Loader2 } from 'lucide-svelte';
  import FileIcon from './FileIcon.svelte';

  interface FileNode {
    name: string;
    path: string;
    fullPath: string;
    isDir: boolean;
    sizeBytes: number;
    ext: string;
    children?: FileNode[];
    isExpanded?: boolean;
    isLoading?: boolean;
  }

  let {
    workspacePath,
    onSelectFile,
  }: {
    workspacePath: string;
    onSelectFile: (relPath: string) => void;
  } = $props();

  let rootNodes = $state<FileNode[]>([]);
  let isLoadingRoot = $state(false);
  let searchQuery = $state('');
  let debouncedSearch = $state('');
  let searchTimeout: any;

  function handleSearchInput(e: Event) {
    const val = (e.target as HTMLInputElement).value;
    searchQuery = val;
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(() => {
      debouncedSearch = val.trim().toLowerCase();
    }, 200);
  }

  // Wails binding caller
  async function fetchDirectory(relativeDir: string): Promise<FileNode[]> {
    try {
      const win = window as any;
      if (win.go?.main?.App?.ReadWorkspaceDirectory) {
        const res = await win.go.main.App.ReadWorkspaceDirectory(workspacePath, relativeDir);
        return res || [];
      }
      return [];
    } catch (err) {
      console.error('Failed to read workspace directory:', err);
      return [];
    }
  }

  export async function loadRoot() {
    if (!workspacePath) return;
    isLoadingRoot = true;
    try {
      const items = await fetchDirectory('');
      rootNodes = items.map((i) => ({
        ...i,
        children: i.isDir ? [] : undefined,
        isExpanded: false,
        isLoading: false,
      }));
    } finally {
      isLoadingRoot = false;
    }
  }

  async function toggleNode(node: FileNode) {
    if (!node.isDir) {
      onSelectFile(node.path);
      return;
    }

    if (node.isExpanded) {
      node.isExpanded = false;
      return;
    }

    // Lazy load children
    if (!node.children || node.children.length === 0) {
      node.isLoading = true;
      try {
        const children = await fetchDirectory(node.path);
        node.children = children.map((c) => ({
          ...c,
          children: c.isDir ? [] : undefined,
          isExpanded: false,
          isLoading: false,
        }));
      } finally {
        node.isLoading = false;
      }
    }

    node.isExpanded = true;
  }

  $effect(() => {
    if (workspacePath) {
      loadRoot();
    }
  });

  // Flat filtering for search queries
  function filterNode(node: FileNode, query: string): boolean {
    if (!query) return true;
    if (node.name.toLowerCase().includes(query)) return true;
    if (node.children && node.children.some((c) => filterNode(c, query))) {
      return true;
    }
    return false;
  }
</script>

<div class="flex flex-col h-full bg-ant-bg text-ant-text font-sans text-xs select-none">
  <!-- Search & Toolbar -->
  <div class="p-2 border-b border-ant-border flex items-center gap-1.5 shrink-0 bg-ant-bg-secondary">
    <div class="relative flex-1">
      <Search class="w-3.5 h-3.5 absolute left-2 top-1/2 -translate-y-1/2 text-ant-text-muted pointer-events-none" />
      <input
        type="text"
        placeholder="Filter files & folders..."
        value={searchQuery}
        oninput={handleSearchInput}
        class="w-full bg-ant-bg-tertiary text-ant-text placeholder-ant-text-muted rounded px-2.5 py-1 pl-7 text-xs border border-transparent focus:border-ant-primary/40 focus:outline-none transition-colors"
      />
    </div>

    <button
      onclick={() => loadRoot()}
      title="Refresh Workspace"
      class="p-1 rounded text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition-colors shrink-0"
    >
      <RefreshCw class="w-3.5 h-3.5 {isLoadingRoot ? 'animate-spin' : ''}" />
    </button>
  </div>

  <!-- Explorer Tree Content -->
  <div class="flex-1 overflow-y-auto p-1.5 space-y-0.5 custom-scrollbar">
    {#if isLoadingRoot}
      <div class="flex items-center justify-center py-8 text-zinc-500 gap-2">
        <Loader2 class="w-4 h-4 animate-spin text-zinc-400" />
        <span>Loading workspace...</span>
      </div>
    {:else if rootNodes.length === 0}
      <div class="text-center py-8 text-zinc-500 italic">
        Workspace empty or unreadable
      </div>
    {:else}
      {#snippet renderTree(nodes: FileNode[], depth: number)}
        {#each nodes as node}
          {#if filterNode(node, debouncedSearch)}
            <div>
              <div
                role="button"
                tabindex="0"
                onclick={() => toggleNode(node)}
                onkeydown={(e) => (e.key === 'Enter' || e.key === ' ') && toggleNode(node)}
                style="padding-left: {depth * 14 + 6}px"
                class="flex items-center gap-1.5 py-1 px-1.5 rounded hover:bg-ant-bg-tertiary cursor-pointer group text-ant-text transition-colors"
              >
                <!-- Chevron or Indent -->
                {#if node.isDir}
                  <div class="w-3.5 h-3.5 flex items-center justify-center text-ant-text-muted group-hover:text-ant-text">
                    {#if node.isLoading}
                      <Loader2 class="w-3 h-3 animate-spin text-amber-500" />
                    {:else if node.isExpanded}
                      <ChevronDown class="w-3 h-3" />
                    {:else}
                      <ChevronRight class="w-3 h-3" />
                    {/if}
                  </div>
                {:else}
                  <div class="w-3.5 h-3.5"></div>
                {/if}

                <!-- File/Folder Icon -->
                <FileIcon
                  isDir={node.isDir}
                  isOpen={node.isExpanded}
                  name={node.name}
                  ext={node.ext}
                  class="w-3.5 h-3.5"
                />

                <!-- Name -->
                <span class="truncate font-mono text-[11px] leading-tight">
                  {node.name}
                </span>
              </div>

              <!-- Children Subtree -->
              {#if node.isDir && node.isExpanded && node.children}
                {@render renderTree(node.children, depth + 1)}
              {/if}
            </div>
          {/if}
        {/each}
      {/snippet}

      {@render renderTree(rootNodes, 0)}
    {/if}
  </div>
</div>

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 5px;
    height: 5px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: #27272a;
    border-radius: 3px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: #3f3f46;
  }
</style>
