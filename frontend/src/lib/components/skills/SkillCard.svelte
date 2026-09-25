<script lang="ts">
  import type { SkillItem } from '../../../app.d';
  import Button from '$lib/antd/Button.svelte';
  import {
    Sparkles,
    Tag,
    FolderGit2,
    ArrowUpRight,
    Terminal,
    Layers,
    Bot,
    Palette,
    Server,
    Layout
  } from 'lucide-svelte';

  interface Props {
    skill: SkillItem;
    onUse: (skill: SkillItem) => void;
  }

  let { skill, onUse }: Props = $props();

  // Color mapping based on category
  const categoryStyles: Record<string, { bg: string; text: string; border: string }> = {
    Frontend: { bg: 'bg-blue-500/10', text: 'text-blue-400', border: 'border-blue-500/30' },
    Backend: { bg: 'bg-emerald-500/10', text: 'text-emerald-400', border: 'border-emerald-500/30' },
    Design: { bg: 'bg-purple-500/10', text: 'text-purple-400', border: 'border-purple-500/30' },
    Agents: { bg: 'bg-amber-500/10', text: 'text-amber-400', border: 'border-amber-500/30' },
    Tools: { bg: 'bg-cyan-500/10', text: 'text-cyan-400', border: 'border-cyan-500/30' }
  };

  const currentCategoryStyle = $derived(
    categoryStyles[skill.category] || {
      bg: 'bg-ant-primary/10',
      text: 'text-ant-primary',
      border: 'border-ant-primary/30'
    }
  );
</script>

<div
  class="group relative flex flex-col justify-between p-3.5 rounded-xl bg-ant-bg border border-ant-border-secondary hover:border-ant-primary/50 transition-all duration-200 shadow-sm hover:shadow-md hover:shadow-ant-primary/5 hover:-translate-y-0.5"
>
  <div>
    <!-- Top row: Scope Badge + Category Tag -->
    <div class="flex items-center justify-between gap-2 mb-2">
      <div class="flex items-center gap-1.5 min-w-0">
        <span
          class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-semibold border {currentCategoryStyle.bg} {currentCategoryStyle.text} {currentCategoryStyle.border}"
        >
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

        <span
          class="px-1.5 py-0.5 rounded text-[9px] font-mono tracking-wider uppercase bg-ant-bg-tertiary text-ant-text-muted border border-ant-border-secondary"
        >
          {skill.scope}
        </span>
      </div>

      <div class="text-[10px] text-ant-text-muted font-mono opacity-0 group-hover:opacity-100 transition-opacity">
        /{skill.name}
      </div>
    </div>

    <!-- Skill Name & Title with Editorial Heading -->
    <h4 class="font-serif-display text-[13px] font-semibold text-ant-text group-hover:text-ant-primary transition-colors flex items-center gap-1.5 truncate">
      <Sparkles size={12} class="text-ant-primary flex-shrink-0" />
      <span class="truncate">{skill.name}</span>
    </h4>

    <!-- Skill Description -->
    <p class="text-[11px] text-ant-text-secondary mt-1 line-clamp-2 leading-relaxed min-h-[32px]">
      {skill.description || 'No description provided for this skill.'}
    </p>

    <!-- Tag Badges -->
    {#if skill.tags && skill.tags.length > 0}
      <div class="flex flex-wrap gap-1 mt-2.5 max-h-11 overflow-hidden">
        {#each skill.tags.slice(0, 3) as tag}
          <span
            class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] bg-ant-bg-secondary text-ant-text-secondary border border-ant-border-secondary/60"
          >
            <Tag size={9} class="mr-1 text-ant-text-muted" />
            <span class="truncate max-w-[90px]">{tag}</span>
          </span>
        {/each}
        {#if skill.tags.length > 3}
          <span class="text-[9px] text-ant-text-muted self-center">
            +{skill.tags.length - 3}
          </span>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Bottom Action Footer -->
  <div class="pt-3 mt-3 border-t border-ant-border-secondary/50 flex items-center justify-between">
    <div
      class="text-[10px] text-ant-text-muted font-mono truncate max-w-[130px]"
      title={skill.path}
    >
      <span class="truncate">{skill.directory.split('/').pop() || skill.name}</span>
    </div>

    <Button
      type="primary"
      size="small"
      onclick={() => onUse(skill)}
      class="!px-2.5 !py-1 !text-[11px] shadow-sm flex items-center"
    >
      <span>Insert / Use</span>
      <ArrowUpRight size={11} class="ml-1" />
    </Button>
  </div>
</div>
