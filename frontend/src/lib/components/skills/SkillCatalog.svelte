<script lang="ts">
  import { onMount, tick } from 'svelte';
  import type { SkillItem } from '../../../app.d';
  import SkillCard from './SkillCard.svelte';
  import Button from '$lib/antd/Button.svelte';
  import {
    Sparkles,
    Search,
    X,
    Filter,
    FolderKanban,
    RefreshCw,
    Wrench,
    CheckCircle2,
    Code2,
    Palette,
    Server,
    Bot,
    Terminal
  } from 'lucide-svelte';

  interface Props {
    visible: boolean;
    onClose: () => void;
    onSelectSkill: (skill: SkillItem) => void;
  }

  let { visible = $bindable(false), onClose, onSelectSkill }: Props = $props();

  let searchQuery = $state('');
  let activeCategory = $state<'All' | 'Frontend' | 'Backend' | 'Design' | 'Agents' | 'Tools'>('All');
  let skills = $state<SkillItem[]>([]);
  let isLoading = $state(false);
  let searchInputEl = $state<HTMLInputElement | null>(null);

  const categories: Array<'All' | 'Frontend' | 'Backend' | 'Design' | 'Agents' | 'Tools'> = [
    'All',
    'Frontend',
    'Backend',
    'Design',
    'Agents',
    'Tools'
  ];

  // Mock skills fallback for preview mode
  const previewSkills: SkillItem[] = [
    {
      id: 'grok:design-taste-frontend',
      name: 'design-taste-frontend',
      description: 'Anti-slop frontend skill for landing pages, portfolios, and redesigns. Infers the right design direction and ships interfaces that do not look templated.',
      category: 'Design',
      tags: ['design', 'frontend', 'anti-slop', 'ui'],
      path: '~/.grok/skills/design-taste-frontend/SKILL.md',
      directory: '~/.grok/skills/design-taste-frontend',
      scope: 'grok'
    },
    {
      id: 'grok:code-review',
      name: 'code-review',
      description: 'Review the current diff for bugs, regressions, security, and missing tests. Concrete fixes with exact file:line references.',
      category: 'Tools',
      tags: ['review', 'diff', 'security', 'git'],
      path: '~/.grok/skills/code-review/SKILL.md',
      directory: '~/.grok/skills/code-review',
      scope: 'grok'
    },
    {
      id: 'agents:build-mcp-server',
      name: 'build-mcp-server',
      description: 'Guidance and scaffolding for Model Context Protocol servers connecting tools, APIs, and headless adapters.',
      category: 'Backend',
      tags: ['mcp', 'server', 'protocol', 'api'],
      path: '~/.agents/skills/build-mcp-server/SKILL.md',
      directory: '~/.agents/skills/build-mcp-server',
      scope: 'agents'
    },
    {
      id: 'grok:browser-skill',
      name: 'browser-skill',
      description: 'Autonomous browser automation and web scraping actions for AI agent workflows.',
      category: 'Agents',
      tags: ['agent', 'browser', 'automation', 'scraping'],
      path: '~/.grok/skills/browser-skill/SKILL.md',
      directory: '~/.grok/skills/browser-skill',
      scope: 'grok'
    },
    {
      id: 'agents:frontend-design',
      name: 'frontend-design',
      description: 'Guidance for distinctive visual design when building new UI. Helps with aesthetic direction, typography, and layout.',
      category: 'Frontend',
      tags: ['frontend', 'css', 'layout', 'typography'],
      path: '~/.agents/skills/frontend-design/SKILL.md',
      directory: '~/.agents/skills/frontend-design',
      scope: 'agents'
    },
    {
      id: 'grok:commit-message',
      name: 'commit-message',
      description: 'Generates concise conventional commit messages based on staged workspace diffs.',
      category: 'Tools',
      tags: ['git', 'commit', 'conventional'],
      path: '~/.grok/skills/commit-message/SKILL.md',
      directory: '~/.grok/skills/commit-message',
      scope: 'grok'
    }
  ];

  async function loadSkills() {
    isLoading = true;
    try {
      if (window.go?.main?.App?.GetInstalledSkills) {
        const result = await window.go.main.App.GetInstalledSkills();
        if (result && Array.isArray(result) && result.length > 0) {
          skills = result;
        } else {
          skills = previewSkills;
        }
      } else {
        skills = previewSkills;
      }
    } catch (err) {
      console.error('Failed to load skills:', err);
      skills = previewSkills;
    } finally {
      isLoading = false;
    }
  }

  // Filter skills based on query and active category
  const filteredSkills = $derived(
    skills.filter((skill) => {
      const matchCat =
        activeCategory === 'All' ||
        skill.category.toLowerCase() === activeCategory.toLowerCase();

      if (!matchCat) return false;

      const q = searchQuery.toLowerCase().trim();
      if (!q) return true;

      const nameMatch = skill.name.toLowerCase().includes(q);
      const descMatch = skill.description?.toLowerCase().includes(q) || false;
      const tagMatch = skill.tags?.some((t) => t.toLowerCase().includes(q)) || false;
      const actMatch = skill.actions?.some((a) => a.toLowerCase().includes(q)) || false;

      return nameMatch || descMatch || tagMatch || actMatch;
    })
  );

  function handleSelect(skill: SkillItem) {
    onSelectSkill(skill);
    onClose();
  }

  $effect(() => {
    if (visible) {
      loadSkills();
      tick().then(() => {
        searchInputEl?.focus();
      });
    }
  });

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      onClose();
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (e.key === 'Escape' && visible) {
      onClose();
    }
  }
</script>

<svelte:window onkeydown={handleKeyDown} />

{#if visible}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4 select-none animate-in fade-in duration-150"
    onclick={handleBackdropClick}
    role="presentation"
  >
    <div
      class="w-full max-w-4xl max-h-[85vh] flex flex-col bg-ant-bg-secondary border border-ant-border rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150"
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-ant-border flex items-center justify-between bg-ant-bg-secondary flex-shrink-0">
        <div class="flex items-center space-x-3">
          <div class="w-8 h-8 rounded-lg bg-ant-primary/10 border border-ant-primary/30 flex items-center justify-center text-ant-primary">
            <Sparkles size={18} />
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="text-sm font-bold text-white tracking-tight">Skills & MCP Discovery Catalog</h2>
              <span class="px-2 py-0.5 text-[10px] font-semibold bg-ant-primary/20 text-ant-primary rounded-full border border-ant-primary/30">
                {skills.length} Installed
              </span>
            </div>
            <p class="text-xs text-ant-text-secondary mt-0.5">
              Explore skills from <code class="text-ant-primary text-[11px]">~/.grok/skills/</code> and <code class="text-ant-primary text-[11px]">~/.agents/skills/</code>
            </p>
          </div>
        </div>

        <div class="flex items-center space-x-2">
          <button
            type="button"
            onclick={loadSkills}
            class="p-1.5 rounded-lg text-ant-text-secondary hover:text-white hover:bg-ant-bg-tertiary transition"
            title="Rescan Skill Directories"
          >
            <RefreshCw size={15} class={isLoading ? 'animate-spin text-ant-primary' : ''} />
          </button>
          <button
            type="button"
            onclick={onClose}
            class="p-1.5 rounded-lg text-ant-text-secondary hover:text-white hover:bg-ant-bg-tertiary transition"
            title="Close"
          >
            <X size={17} />
          </button>
        </div>
      </div>

      <!-- Controls: Search Input & Category Tabs -->
      <div class="px-6 py-3.5 border-b border-ant-border-secondary bg-ant-bg/80 flex flex-col sm:flex-row gap-3 items-center justify-between flex-shrink-0">
        <!-- Search Field -->
        <div class="relative w-full sm:w-80">
          <Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-ant-text-muted" />
          <input
            bind:this={searchInputEl}
            bind:value={searchQuery}
            type="text"
            placeholder="Search skills, actions, tags..."
            class="w-full pl-9 pr-8 py-1.5 text-xs bg-ant-bg-secondary border border-ant-border focus:border-ant-primary rounded-lg text-white placeholder:text-ant-text-muted outline-none transition"
          />
          {#if searchQuery}
            <button
              type="button"
              onclick={() => searchQuery = ''}
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-ant-text-muted hover:text-white text-xs"
            >
              <X size={13} />
            </button>
          {/if}
        </div>

        <!-- Category Tabs -->
        <div class="flex items-center space-x-1 bg-ant-bg-secondary p-1 rounded-lg border border-ant-border-secondary overflow-x-auto max-w-full">
          {#each categories as cat}
            <button
              type="button"
              onclick={() => activeCategory = cat}
              class="px-2.5 py-1 text-xs font-medium rounded-md transition whitespace-nowrap {activeCategory === cat ? 'bg-ant-primary text-white shadow-sm' : 'text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary'}"
            >
              {cat}
            </button>
          {/each}
        </div>
      </div>

      <!-- Skills Grid Content -->
      <div class="flex-1 overflow-y-auto p-6 scrollbar-thin">
        {#if isLoading}
          <div class="flex flex-col items-center justify-center py-20 text-ant-text-secondary">
            <RefreshCw size={24} class="animate-spin text-ant-primary mb-3" />
            <span class="text-xs">Scanning local skill directories...</span>
          </div>
        {:else if filteredSkills.length === 0}
          <div class="flex flex-col items-center justify-center py-16 text-center text-ant-text-secondary">
            <FolderKanban size={36} class="text-ant-text-muted mb-3 opacity-40" />
            <div class="text-sm font-semibold text-white">No skills matching query</div>
            <p class="text-xs text-ant-text-muted mt-1 max-w-sm">
              Try adjusting your search terms or selecting the "All" category tab.
            </p>
            {#if searchQuery}
              <Button size="small" type="default" onclick={() => { searchQuery = ''; activeCategory = 'All'; }} class="mt-4">
                Clear Filters
              </Button>
            {/if}
          </div>
        {:else}
          <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3.5">
            {#each filteredSkills as skill (skill.id)}
              <SkillCard
                {skill}
                onUse={handleSelect}
              />
            {/each}
          </div>
        {/if}
      </div>

      <!-- Footer Info -->
      <div class="px-6 py-2.5 bg-ant-bg border-t border-ant-border-secondary/60 flex items-center justify-between text-[11px] text-ant-text-muted flex-shrink-0">
        <div class="flex items-center space-x-2">
          <span>Showing <strong class="text-white">{filteredSkills.length}</strong> of {skills.length} skills</span>
        </div>
        <div>
          <span>Press <kbd class="px-1.5 py-0.5 rounded bg-ant-bg-secondary border border-ant-border text-[10px] text-ant-text font-mono">Esc</kbd> to exit</span>
        </div>
      </div>
    </div>
  </div>
{/if}
