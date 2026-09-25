<script lang="ts">
  import {
    Folder,
    FolderOpen,
    File,
    FileCode,
    FileText,
    FileJson,
    Terminal,
    Image as ImageIcon,
    Database,
    Lock,
    Settings,
  } from 'lucide-svelte';

  let {
    isDir = false,
    isOpen = false,
    ext = '',
    name = '',
    class: className = 'w-4 h-4 shrink-0',
  }: {
    isDir?: boolean;
    isOpen?: boolean;
    ext?: string;
    name?: string;
    class?: string;
  } = $props();

  const lowerName = $derived(name.toLowerCase());
  const cleanExt = $derived(ext.startsWith('.') ? ext.slice(1).toLowerCase() : ext.toLowerCase());
</script>

{#if isDir}
  {#if isOpen}
    <FolderOpen class="{className} text-amber-400/90" />
  {:else}
    <Folder class="{className} text-amber-400/80" />
  {/if}
{:else}
  {#if lowerName === 'package.json' || cleanExt === 'json'}
    <FileJson class="{className} text-yellow-400/90" />
  {:else}
    {#if cleanExt === 'ts' || cleanExt === 'tsx'}
      <FileCode class="{className} text-blue-400" />
    {:else}
      {#if cleanExt === 'js' || cleanExt === 'jsx'}
        <FileCode class="{className} text-yellow-300" />
      {:else}
        {#if cleanExt === 'svelte'}
          <FileCode class="{className} text-orange-400" />
        {:else}
          {#if cleanExt === 'go'}
            <FileCode class="{className} text-cyan-400" />
          {:else}
            {#if cleanExt === 'sh' || cleanExt === 'bash' || cleanExt === 'zsh'}
              <Terminal class="{className} text-emerald-400" />
            {:else}
              {#if cleanExt === 'md' || cleanExt === 'markdown' || cleanExt === 'txt'}
                <FileText class="{className} text-zinc-300" />
              {:else}
                {#if cleanExt === 'png' || cleanExt === 'jpg' || cleanExt === 'jpeg' || cleanExt === 'svg' || cleanExt === 'webp' || cleanExt === 'gif'}
                  <ImageIcon class="{className} text-purple-400" />
                {:else}
                  {#if cleanExt === 'sql' || cleanExt === 'db' || cleanExt === 'sqlite'}
                    <Database class="{className} text-indigo-400" />
                  {:else}
                    {#if cleanExt === 'lock' || lowerName.includes('lock')}
                      <Lock class="{className} text-zinc-400" />
                    {:else}
                      {#if lowerName.startsWith('.env') || cleanExt === 'yaml' || cleanExt === 'yml' || cleanExt === 'toml'}
                        <Settings class="{className} text-amber-300" />
                      {:else}
                        <File class="{className} text-zinc-400" />
                      {/if}
                    {/if}
                  {/if}
                {/if}
              {/if}
            {/if}
          {/if}
        {/if}
      {/if}
    {/if}
  {/if}
{/if}
