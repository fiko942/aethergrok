<script lang="ts">
  import { onMount, tick } from 'svelte';
  import type { SkillItem } from '../../../app.d';
  import Button from '$lib/antd/Button.svelte';
  import SkillImporterModal from '$lib/components/skills/SkillImporterModal.svelte';
  import {
    Sparkles,
    Search,
    X,
    FolderKanban,
    RefreshCw,
    Tag,
    ArrowUpRight,
    Terminal,
    Bot,
    Palette,
    Server,
    Layout,
    Check,
    FileText,
    ExternalLink,
    FolderGit2,
    Plus
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
  let selectedSkillId = $state<string | null>(null);
  let importerModalVisible = $state(false);

  // Tab button references and sliding pill geometry
  let tabContainerEl = $state<HTMLDivElement | null>(null);
  let tabButtonEls = $state<Record<string, HTMLButtonElement | null>>({});
  let pillStyle = $state({ left: 4, width: 40, opacity: 0 });

  function updatePillPosition() {
    const activeBtn = tabButtonEls[activeCategory];
    const container = tabContainerEl;
    if (activeBtn && container) {
      const containerRect = container.getBoundingClientRect();
      const btnRect = activeBtn.getBoundingClientRect();
      pillStyle = {
        left: btnRect.left - containerRect.left,
        width: btnRect.width,
        opacity: 1
      };
    }
  }

  $effect(() => {
    // Recompute pill position whenever activeCategory or visible changes
    if (visible && activeCategory) {
      // Use tick or requestAnimationFrame to ensure DOM is ready
      requestAnimationFrame(updatePillPosition);
    }
  });

  const categories: Array<'All' | 'Frontend' | 'Backend' | 'Design' | 'Agents' | 'Tools'> = [
    'All',
    'Frontend',
    'Backend',
    'Design',
    'Agents',
    'Tools'
  ];

  // Category Styling Tokens
  const categoryBadges: Record<string, { bg: string; text: string; border: string }> = {
    Frontend: { bg: 'bg-blue-500/15', text: 'text-blue-700 dark:text-blue-400', border: 'border-blue-500/30' },
    Backend: { bg: 'bg-emerald-500/15', text: 'text-emerald-700 dark:text-emerald-400', border: 'border-emerald-500/30' },
    Design: { bg: 'bg-purple-500/15', text: 'text-purple-700 dark:text-purple-400', border: 'border-purple-500/30' },
    Agents: { bg: 'bg-amber-500/15', text: 'text-amber-800 dark:text-amber-300', border: 'border-amber-500/30' },
    Tools: { bg: 'bg-cyan-500/15', text: 'text-cyan-800 dark:text-cyan-300', border: 'border-cyan-500/30' }
  };

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
      console.warn('Failed to load skills from Go bridge:', err);
      skills = previewSkills;
    } finally {
      isLoading = false;
    }
  }

  $effect(() => {
    if (visible) {
      loadSkills();
      tick().then(() => {
        searchInputEl?.focus();
      });
    }
  });

  const filteredSkills = $derived(
    skills.filter((skill) => {
      const matchesCategory =
        activeCategory === 'All' || skill.category?.toLowerCase() === activeCategory.toLowerCase();
      const q = searchQuery.toLowerCase().trim();
      if (!q) return matchesCategory;

      const matchesQuery =
        skill.name.toLowerCase().includes(q) ||
        (skill.description && skill.description.toLowerCase().includes(q)) ||
        (skill.tags && skill.tags.some((t) => t.toLowerCase().includes(q))) ||
        (skill.scope && skill.scope.toLowerCase().includes(q));

      return matchesCategory && matchesQuery;
    })
  );

  function handleSelect(skill: SkillItem) {
    selectedSkillId = skill.id;
    onSelectSkill(skill);
    onClose();
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      onClose();
    }
  }

  function handleBackdropClick(e: MouseEvent) {
    if (e.target === e.currentTarget) {
      onClose();
    }
  }
</script>

<svelte:window onkeydown={handleKeyDown} />

{#if visible}
  <div
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 select-none animate-in fade-in duration-150"
    onclick={handleBackdropClick}
    role="presentation"
  >
    <div
      class="w-full max-w-5xl max-h-[88vh] flex flex-col bg-ant-bg-secondary border border-ant-border-secondary rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150 text-ant-text"
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-ant-border-secondary flex items-center justify-between bg-ant-bg-secondary flex-shrink-0">
        <div class="flex items-center space-x-3">
          <div class="w-8 h-8 rounded-lg bg-ant-primary/10 flex items-center justify-center text-ant-primary shadow-sm">
            <Sparkles size={18} />
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="font-serif-display text-base font-semibold text-ant-text tracking-tight">Skills & MCP Discovery Catalog</h2>
              <span class="px-2.5 py-0.5 text-[10px] font-serif font-semibold bg-blue-500/10 text-blue-400 rounded-full border border-blue-500/20 shadow-xs">
                {skills.length} Installed
              </span>
            </div>
            <p class="text-xs text-ant-text-secondary mt-0.5">
              Explore skills from <code class="text-ant-primary text-[11px] font-mono">~/.grok/skills/</code> and <code class="text-ant-primary text-[11px] font-mono">~/.agents/skills/</code>
            </p>
          </div>
        </div>

        <div class="flex items-center space-x-2">
          <Button
            size="small"
            type="primary"
            onclick={() => importerModalVisible = true}
            class="!px-2.5 !py-1 text-xs flex items-center gap-1.5"
          >
            <FolderGit2 size={13} />
            <span>Add Skills from GitHub</span>
          </Button>

          <button
            type="button"
            onclick={loadSkills}
            class="p-1.5 rounded-lg text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition"
            title="Rescan Skill Directories"
          >
            <RefreshCw size={15} class={isLoading ? 'animate-spin text-ant-primary' : ''} />
          </button>
          <button
            type="button"
            onclick={onClose}
            class="p-1.5 rounded-lg text-ant-text-secondary hover:text-ant-text hover:bg-ant-bg-tertiary transition"
            title="Close"
          >
            <X size={17} />
          </button>
        </div>
      </div>

      <!-- Controls: Search Input & Category Filter Tabs -->
      <div class="px-6 py-3.5 border-b border-ant-border-secondary bg-ant-bg/80 flex flex-col sm:flex-row gap-3 items-center justify-between flex-shrink-0">
        <!-- Search Field -->
        <div class="relative w-full sm:w-80">
          <Search size={14} class="absolute left-3 top-1/2 -translate-y-1/2 text-ant-text-muted" />
          <input
            bind:this={searchInputEl}
            bind:value={searchQuery}
            type="text"
            placeholder="Search skills, triggers, keywords..."
            class="w-full pl-9 pr-8 py-1.5 text-xs bg-ant-bg border border-ant-border-secondary focus:border-ant-primary rounded-lg text-ant-text placeholder:text-ant-text-muted outline-none transition shadow-inner"
          />
          {#if searchQuery}
            <button
              type="button"
              onclick={() => searchQuery = ''}
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-ant-text-muted hover:text-ant-text text-xs"
            >
              <X size={13} />
            </button>
          {/if}
        </div>

        <!-- Category Tabs with Animated Sliding Pill -->
        <div
          bind:this={tabContainerEl}
          class="relative flex items-center bg-ant-bg-secondary p-1 rounded-xl border border-white/5 overflow-x-auto max-w-full"
        >
          <!-- Sliding Pill Background Indicator -->
          <div
            class="absolute top-1 bottom-1 bg-ant-primary rounded-lg shadow-sm transition-all duration-200 ease-out pointer-events-none"
            style="left: {pillStyle.left}px; width: {pillStyle.width}px; opacity: {pillStyle.opacity};"
          ></div>

          {#each categories as cat}
            <button
              bind:this={tabButtonEls[cat]}
              type="button"
              onclick={() => {
                activeCategory = cat;
                requestAnimationFrame(updatePillPosition);
              }}
              class="relative z-10 px-3 py-1 text-xs font-medium rounded-lg transition-colors whitespace-nowrap {activeCategory === cat ? 'text-white' : 'text-ant-text-secondary hover:text-ant-text'}"
            >
              {cat}
            </button>
          {/each}
        </div>
      </div>

      <!-- Clean Table View Layout -->
      <div class="flex-1 overflow-y-auto p-4 scrollbar-thin">
        {#if isLoading}
          <div class="flex flex-col items-center justify-center py-20 text-ant-text-secondary">
            <RefreshCw size={24} class="animate-spin text-ant-primary mb-3" />
            <span class="text-xs">Scanning local skill directories...</span>
          </div>
        {:else if filteredSkills.length === 0}
          <div class="flex flex-col items-center justify-center py-16 text-center text-ant-text-secondary">
            <FolderKanban size={36} class="text-ant-text-muted mb-3 opacity-40" />
            <div class="text-sm font-semibold text-ant-text">No skills matching query</div>
            <p class="text-xs text-ant-text-muted mt-1 max-w-sm">
              Try adjusting your search query or switching the category tab.
            </p>
            {#if searchQuery}
              <Button size="small" type="default" onclick={() => { searchQuery = ''; activeCategory = 'All'; }} class="mt-4">
                Clear Filters
              </Button>
            {/if}
          </div>
        {:else}
          <div class="border border-ant-border-secondary rounded-xl overflow-hidden bg-ant-bg shadow-sm">
            <table class="w-full text-left border-collapse text-xs">
              <thead>
                <tr class="bg-ant-bg-secondary border-b border-ant-border-secondary text-[11px] font-semibold text-ant-text-secondary select-none">
                  <th class="py-2.5 px-4 w-[220px]">Skill / Command</th>
                  <th class="py-2.5 px-3 w-[110px]">Category</th>
                  <th class="py-2.5 px-3">Description & Triggers</th>
                  <th class="py-2.5 px-3 w-[140px] hidden md:table-cell">Tags</th>
                  <th class="py-2.5 px-4 w-[110px] text-right">Action</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-ant-border-secondary">
                {#each filteredSkills as skill (skill.id)}
                  {@const catStyle = categoryBadges[skill.category] || { bg: 'bg-ant-primary/10', text: 'text-ant-primary', border: 'border-ant-primary/30' }}
                  <tr class="hover:bg-ant-bg-tertiary/60 transition-colors group">
                    <!-- Name & Command -->
                    <td class="py-3 px-4 align-top">
                      <div class="flex flex-col gap-1">
                        <div class="flex items-center gap-1.5 font-semibold text-ant-text group-hover:text-ant-primary transition-colors">
                          <Sparkles size={12} class="text-ant-primary flex-shrink-0" />
                          <span class="truncate">{skill.name}</span>
                        </div>
                        <div class="inline-flex items-center gap-1">
                          <span class="px-1.5 py-0.2 rounded font-mono text-[10px] bg-ant-bg-tertiary text-blue-700 dark:text-blue-400 border border-ant-border-secondary">
                            /{skill.name}
                          </span>
                          <span class="text-[9px] uppercase tracking-wider text-ant-text-muted px-1 font-mono">
                            {skill.scope}
                          </span>
                        </div>
                      </div>
                    </td>

                    <!-- Category -->
                    <td class="py-3 px-3 align-top">
                      <span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium border {catStyle.bg} {catStyle.text} {catStyle.border}">
                        {#if skill.category === 'Frontend'}
                          <Layout size={10} class="mr-1" />
                        {:else if skill.category === 'Backend'}
                          <Server size={10} class="mr-1" />
                        {:else if skill.category === 'Design'}
                          <Palette size={10} class="mr-1" />
                        {:else if skill.category === 'Agents'}
                          <Bot size={10} class="mr-1" />
                        {:else}
                          <Terminal size={10} class="mr-1" />
                        {/if}
                        {skill.category}
                      </span>
                    </td>

                    <!-- Description -->
                    <td class="py-3 px-3 align-top">
                      <p class="text-[11px] text-ant-text-secondary leading-relaxed line-clamp-2">
                        {skill.description || 'No description provided.'}
                      </p>
                    </td>

                    <!-- Tags -->
                    <td class="py-3 px-3 align-top hidden md:table-cell">
                      {#if skill.tags && skill.tags.length > 0}
                        <div class="flex flex-wrap gap-1">
                          {#each skill.tags.slice(0, 2) as tag}
                            <span class="inline-flex items-center px-1.5 py-0.2 rounded text-[10px] bg-ant-bg-tertiary text-ant-text-secondary border border-ant-border-secondary">
                              <Tag size={9} class="mr-1 opacity-70" />
                              <span class="truncate max-w-[70px]">{tag}</span>
                            </span>
                          {/each}
                          {#if skill.tags.length > 2}
                            <span class="text-[9px] text-ant-text-muted self-center font-medium">
                              +{skill.tags.length - 2}
                            </span>
                          {/if}
                        </div>
                      {:else}
                        <span class="text-[10px] text-ant-text-muted italic">-</span>
                      {/if}
                    </td>

                    <!-- Action -->
                    <td class="py-3 px-4 align-top text-right">
                      <Button
                        type="primary"
                        size="small"
                        onclick={() => handleSelect(skill)}
                        class="!px-2.5 !py-1 !text-[11px] shadow-sm flex items-center justify-center ml-auto"
                      >
                        <span>Use</span>
                        <ArrowUpRight size={11} class="ml-1" />
                      </Button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>

      <!-- Footer Bar -->
      <div class="px-6 py-2.5 border-t border-ant-border-secondary bg-ant-bg-secondary flex items-center justify-between text-xs text-ant-text-muted flex-shrink-0">
        <span>Showing <strong class="text-ant-text">{filteredSkills.length}</strong> of {skills.length} skills</span>
        <div class="flex items-center space-x-2">
          <span class="font-mono text-[10px] bg-ant-bg px-1.5 py-0.5 rounded border border-ant-border-secondary">Press Esc to exit</span>
        </div>
      </div>
    </div>
  </div>
{/if}

<!-- Skill Importer Modal (GitHub) -->
<SkillImporterModal
  bind:visible={importerModalVisible}
  onClose={() => importerModalVisible = false}
  onInstalled={() => {
    loadSkills();
  }}
/>
