<script lang="ts">
  import { onMount, tick } from 'svelte';
  import type { SkillItem } from '../../../app.d';
  import {
    Sparkles,
    Terminal,
    Bot,
    Palette,
    Server,
    Layout,
    Tag,
    CornerDownLeft,
    Search
  } from 'lucide-svelte';

  interface Props {
    visible: boolean;
    query: string;
    onSelect: (skill: SkillItem) => void;
    onClose: () => void;
  }

  let { visible = false, query = '', onSelect, onClose }: Props = $props();

  let skills = $state<SkillItem[]>([]);
  let selectedIndex = $state(0);
  let listContainerRef = $state<HTMLDivElement | null>(null);

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
    try {
      if (typeof window !== 'undefined' && window.go?.main?.App?.GetInstalledSkills) {
        const result = await window.go.main.App.GetInstalledSkills();
        if (result && Array.isArray(result) && result.length > 0) {
          skills = result;
          return;
        }
      }
    } catch (e) {
      console.warn('Slash command autocomplete skill load failed:', e);
    }
    skills = previewSkills;
  }

  onMount(() => {
    loadSkills();
  });

  const filteredSkills = $derived.by(() => {
    const cleanQuery = query.toLowerCase().replace(/^\//, '').trim();
    if (!cleanQuery) return skills.slice(0, 8);

    return skills
      .filter((s) => {
        return (
          s.name.toLowerCase().includes(cleanQuery) ||
          s.category?.toLowerCase().includes(cleanQuery) ||
          (s.description && s.description.toLowerCase().includes(cleanQuery)) ||
          (s.tags && s.tags.some((t) => t.toLowerCase().includes(cleanQuery)))
        );
      })
      .slice(0, 10);
  });

  $effect(() => {
    // Reset selection index when query changes or bounds change
    if (selectedIndex >= filteredSkills.length) {
      selectedIndex = 0;
    }
  });

  export function selectNext() {
    if (filteredSkills.length === 0) return;
    selectedIndex = (selectedIndex + 1) % filteredSkills.length;
    scrollSelectedIntoView();
  }

  export function selectPrev() {
    if (filteredSkills.length === 0) return;
    selectedIndex = (selectedIndex - 1 + filteredSkills.length) % filteredSkills.length;
    scrollSelectedIntoView();
  }

  export function getSelectedSkill(): SkillItem | null {
    if (filteredSkills.length > 0 && selectedIndex >= 0 && selectedIndex < filteredSkills.length) {
      return filteredSkills[selectedIndex];
    }
    return null;
  }

  function scrollSelectedIntoView() {
    tick().then(() => {
      const activeEl = listContainerRef?.querySelector(`[data-index="${selectedIndex}"]`);
      activeEl?.scrollIntoView({ block: 'nearest' });
    });
  }
</script>

{#if visible && filteredSkills.length > 0}
  <div
    class="absolute bottom-full left-0 mb-3 w-96 max-w-[95vw] bg-ant-bg border border-ant-border rounded-xl shadow-2xl z-[100] flex flex-col animate-in fade-in zoom-in-95 duration-100 text-ant-text"
    style="box-shadow: 0 20px 40px -4px rgba(0, 0, 0, 0.45), 0 8px 16px -4px rgba(0, 0, 0, 0.25);"
  >
    <!-- Autocomplete Header -->
    <div class="px-3.5 py-2.5 border-b border-ant-border bg-ant-bg-secondary rounded-t-xl flex items-center justify-between">
      <div class="flex items-center gap-1.5 text-[11px] font-semibold text-ant-text">
        <Sparkles size={12} class="text-ant-primary" />
        <span>Matching Skills ({filteredSkills.length})</span>
      </div>
      <div class="flex items-center gap-1.5 text-[10px] text-ant-text-muted">
        <span class="px-1 py-0.2 rounded bg-ant-bg-tertiary border border-ant-border font-mono">↑↓</span>
        <span>navigate</span>
        <span class="px-1 py-0.2 rounded bg-ant-bg-tertiary border border-ant-border font-mono ml-1">Tab / ↵</span>
        <span>select</span>
      </div>
    </div>

    <!-- Autocomplete Items List -->
    <div
      bind:this={listContainerRef}
      class="max-h-64 overflow-y-auto p-1.5 space-y-1 scrollbar-thin bg-ant-bg"
    >
      {#each filteredSkills as skill, idx (skill.id)}
        {@const isSelected = idx === selectedIndex}
        <button
          type="button"
          data-index={idx}
          class="w-full text-left p-2 rounded-lg text-xs transition-all flex items-start justify-between group {isSelected
            ? 'bg-ant-primary/15 border border-ant-primary/40 text-ant-text'
            : 'hover:bg-ant-bg-tertiary border border-transparent text-ant-text-secondary hover:text-ant-text'}"
          onclick={() => onSelect(skill)}
          onmouseenter={() => selectedIndex = idx}
        >
          <div class="flex items-start gap-2.5 min-w-0 pr-2">
            <!-- Scope / Category Avatar -->
            <div class="mt-0.5 w-6 h-6 rounded flex items-center justify-center flex-shrink-0 {isSelected ? 'bg-ant-primary text-white' : 'bg-ant-bg-tertiary text-ant-primary border border-ant-border'}">
              {#if skill.category === 'Frontend'}
                <Layout size={12} />
              {:else if skill.category === 'Backend'}
                <Server size={12} />
              {:else if skill.category === 'Design'}
                <Palette size={12} />
              {:else if skill.category === 'Agents'}
                <Bot size={12} />
              {:else}
                <Terminal size={12} />
              {/if}
            </div>

            <!-- Details -->
            <div class="flex flex-col min-w-0">
              <div class="flex items-center gap-1.5">
                <span class="font-mono font-semibold text-[11px] {isSelected ? 'text-ant-primary' : 'text-ant-text'}">
                  /{skill.name}
                </span>
                <span class="px-1 py-0.2 text-[9px] uppercase tracking-wider bg-ant-bg-secondary text-ant-text-muted rounded border border-ant-border-secondary font-mono">
                  {skill.scope}
                </span>
              </div>
              <p class="text-[10px] text-ant-text-muted mt-0.5 line-clamp-1 leading-normal">
                {skill.description || 'No description available'}
              </p>
            </div>
          </div>

          {#if isSelected}
            <div class="flex-shrink-0 mt-1 text-ant-primary flex items-center gap-1 text-[10px] font-mono">
              <CornerDownLeft size={11} />
            </div>
          {/if}
        </button>
      {/each}
    </div>

    <!-- Quick Footer -->
    <div class="px-3 py-2 bg-ant-bg-secondary border-t border-ant-border rounded-b-xl flex items-center justify-between text-[10px] text-ant-text-muted">
      <span>Type skill name or press <strong class="text-ant-text">Esc</strong> to dismiss</span>
      <span class="font-mono text-ant-primary">/{query.replace(/^\//, '')}</span>
    </div>
  </div>
{/if}
