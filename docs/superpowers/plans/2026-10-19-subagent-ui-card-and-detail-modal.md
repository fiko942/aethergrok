# Subagent UI Card, Live Timer & Process Detail Modal Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform autonomous subagent executions (`spawn_subagent`, `subagent`, `delegate`) into a dedicated violet-themed Subagent Card with a live-updating running timer (`Xs` / `Xm Ys`), clear human-readable task description, and an interactive process detail modal showing full instructions, elapsed execution time, and clean formatted results.

**Architecture:**
1. In `frontend/src/lib/utils/toolUtils.ts`, implement `isSubagentToolCall(toolCall)` and `cleanSubagentOutput(rawResult)`, and parse polymorphic subagent tool payloads (`description`, `prompt`, `subagent_type`, `model`, `isolation`, `background`).
2. In `frontend/src/lib/components/chat/ToolCallCard.svelte`, introduce a dedicated `subagent` card variant with violet styling, an active 1-second reactive ticking timer during `status === 'running'`, clear task description in the header, and an interactive trigger to open the detail popup.
3. In `frontend/src/lib/components/chat/SubagentDetailModal.svelte`, create a modern modal popup displaying:
   - Header with Bot icon, task description, status pill (`Running`, `Completed`, `Failed`), and running timer (`Xm Ys`).
   - Quick metadata chips (Agent Type, Model, Isolation Mode, Execution Mode).
   - Task Instructions (Prompt) section with a one-click copy button.
   - Live Output / Final Results section formatted with markdown and syntax highlighting with a copy button.
   - Collapsible raw payload section for debugging.

**Tech Stack:** Svelte 5 (Runes), TypeScript, Tailwind CSS, Lucide icons, Vitest.

## Global Constraints

- Follow strict TDD: write failing unit tests first, verify failure, implement minimal code, verify pass.
- Do not alter or break existing tool call card implementations (`terminal`, `read`, `write`, `edit`, `search`, `list`, `plan`).
- Ensure normal keyboard shortcuts (`Escape` to close modal, standard copy/paste functionality) remain 100% intact.
- The live duration timer must update every second while the subagent is in `running` state without causing memory leaks (clear intervals on destroy/completion).
- All unit and integration tests must pass cleanly.

---

### Task 1: Subagent Tool Classifier, Parameter Extraction & Output Cleaning

**Files:**
- Modify: `frontend/src/lib/utils/toolUtils.ts`
- Test: `test/subagent-card-parser.test.ts`

**Interfaces:**
- Consumes: `ToolCall` object from `session.svelte.ts`
- Produces:
  - `isSubagentToolCall(toolName?: string, params?: unknown): boolean`
  - `extractSubagentMeta(params?: unknown): SubagentParsedMeta`
  - `cleanSubagentOutput(text?: string): string`

- [ ] **Step 1: Write failing unit tests in `test/subagent-card-parser.test.ts`**

```ts
import { describe, it, expect } from 'vitest';
import {
  isSubagentToolCall,
  extractSubagentMeta,
  cleanSubagentOutput,
  formatSubagentDuration
} from '../frontend/src/lib/utils/toolUtils';

describe('Subagent Tool Classifier & Output Parsing', () => {
  it('correctly identifies spawn_subagent and subagent tool names', () => {
    expect(isSubagentToolCall('spawn_subagent')).toBe(true);
    expect(isSubagentToolCall('subagent')).toBe(true);
    expect(isSubagentToolCall('delegate_task')).toBe(true);
    expect(isSubagentToolCall('run_terminal_command', { is_background: true, description: 'Spawn background subagent' })).toBe(true);
    expect(isSubagentToolCall('run_terminal_command', { command: 'ls -la' })).toBe(false);
    expect(isSubagentToolCall('read_file')).toBe(false);
  });

  it('extracts human task description and prompt cleanly', () => {
    const params = {
      description: 'Investigate slow database queries',
      prompt: 'Check slow query logs in /var/log/postgres and report top 3 bottlenecks',
      subagent_type: 'general-purpose',
      model: 'grok-beta',
      background: true
    };
    const meta = extractSubagentMeta(params);
    expect(meta.description).toBe('Investigate slow database queries');
    expect(meta.prompt).toBe('Check slow query logs in /var/log/postgres and report top 3 bottlenecks');
    expect(meta.subagentType).toBe('general-purpose');
    expect(meta.model).toBe('grok-beta');
    expect(meta.isBackground).toBe(true);
  });

  it('falls back to the first line of prompt if description is missing', () => {
    const params = {
      prompt: 'Refactor authentication middleware\nEnsure JWT validation passes tests'
    };
    const meta = extractSubagentMeta(params);
    expect(meta.description).toBe('Refactor authentication middleware');
    expect(meta.subagentType).toBe('general-purpose');
  });

  it('cleans internal XML tags and CLI boilerplate from subagent output', () => {
    const raw = `<subagent_meta>{"id":"sub-1"}</subagent_meta>
This is the output of the subagent:
<response>
Found 2 memory leaks in connection pool.
</response>
Agent ID: sub-1 (resume from here)`;
    const cleaned = cleanSubagentOutput(raw);
    expect(cleaned).toBe('Found 2 memory leaks in connection pool.');
  });

  it('formats running and completed durations cleanly matching Grok CLI rules', () => {
    expect(formatSubagentDuration(500)).toBe('1s');
    expect(formatSubagentDuration(42000)).toBe('42s');
    expect(formatSubagentDuration(75000)).toBe('1m 15s');
    expect(formatSubagentDuration(3665000)).toBe('1h 1m 5s');
  });
});
```

- [ ] **Step 2: Run test to verify failure**

Run: `bun test test/subagent-card-parser.test.ts`
Expected: FAIL (functions not exported from `toolUtils.ts`)

- [ ] **Step 3: Implement subagent parsing functions in `frontend/src/lib/utils/toolUtils.ts`**

Add the following types and functions to `frontend/src/lib/utils/toolUtils.ts`:

```ts
export interface SubagentParsedMeta {
  description: string;
  prompt: string;
  subagentType: string;
  model?: string;
  isolation?: string;
  isBackground: boolean;
  taskId?: string;
}

/**
 * Returns true if the tool call represents an autonomous subagent delegation
 */
export function isSubagentToolCall(toolName?: string, params?: unknown): boolean {
  if (!toolName) return false;
  const lower = toolName.toLowerCase().replace(/[_\s-]/g, '');
  if (lower === 'spawnsubagent' || lower === 'subagent' || lower === 'delegatetask' || lower === 'runagent') {
    return true;
  }

  if (params && typeof params === 'object') {
    const p = params as Record<string, unknown>;
    if (p.subagent_type || p.subagentType || p.subagent || p.agent_type) {
      return true;
    }
    if ((p.is_background === true || p.background === true) && typeof p.description === 'string' && /subagent/i.test(p.description)) {
      return true;
    }
  }

  return false;
}

/**
 * Extracts structured metadata from subagent parameters
 */
export function extractSubagentMeta(params?: unknown): SubagentParsedMeta {
  let p: Record<string, unknown> = {};
  if (typeof params === 'string') {
    try {
      p = JSON.parse(params);
    } catch {
      p = { prompt: params };
    }
  } else if (params && typeof params === 'object') {
    p = params as Record<string, unknown>;
  }

  let description = typeof p.description === 'string' ? p.description.trim() : '';
  const prompt = typeof p.prompt === 'string' ? p.prompt.trim() : (typeof p.instruction === 'string' ? p.instruction.trim() : '');

  if (!description && prompt) {
    description = prompt.split('\n')[0].replace(/^[#\-*\s]+/, '').trim();
    if (description.length > 80) {
      description = description.slice(0, 77) + '...';
    }
  }

  if (!description) {
    description = 'Autonomous subagent task';
  }

  const subagentType = typeof p.subagent_type === 'string'
    ? p.subagent_type
    : (typeof p.subagentType === 'string' ? p.subagentType : (typeof p.type === 'string' ? p.type : 'general-purpose'));

  const model = typeof p.model === 'string' ? p.model : undefined;
  const isolation = typeof p.isolation === 'string' ? p.isolation : undefined;
  const isBackground = p.background === true || p.is_background === true;
  const taskId = typeof p.task_id === 'string' ? p.task_id : (typeof p.subagent_id === 'string' ? p.subagent_id : undefined);

  return {
    description,
    prompt,
    subagentType,
    model,
    isolation,
    isBackground,
    taskId
  };
}

/**
 * Strips internal XML plumbing, lead-in boilerplate, and trailing agent hints
 */
export function cleanSubagentOutput(text?: string): string {
  if (!text || typeof text !== 'string') return '';

  let s = text
    .replace(/<subagent_(meta|result)>[\s\S]*?<\/subagent_\1>/g, '')
    .replace(/<\/?subagent_(meta|result)>/g, '')
    .trim();

  // Strip poller divider if present
  const outDivider = /^[\s\S]*?^===\s*Output\s*===\s*\n?/im;
  if (/^===\s*Task\s+/im.test(s) && outDivider.test(s)) {
    s = s.replace(outDivider, '').trim();
  }

  // Strip common CLI lead-in phrases
  s = s.replace(/^\[subagent:[^\]]*\]\s*/i, '');
  s = s.replace(/^this is the output of the subagent:\s*/i, '');
  s = s.replace(/^response:\s*/i, '');
  s = s.replace(/\n\s*agent id:\s*[0-9a-f][0-9a-f-]*\s*(\([^)]*\))?\s*$/i, '').trim();

  const wrapped = /^<response>\s*([\s\S]*?)\s*<\/response>$/i.exec(s);
  if (wrapped) {
    s = wrapped[1];
  }

  return s.trim();
}

/**
 * Formats duration milliseconds into human-readable stopwatch text (e.g. "45s", "1m 15s", "1h 2m 10s")
 */
export function formatSubagentDuration(ms: number): string {
  if (!Number.isFinite(ms) || ms < 0) return '0s';
  const totalSeconds = Math.max(1, Math.round(ms / 1000));
  if (totalSeconds < 60) {
    return `${totalSeconds}s`;
  }
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`;
  }
  return `${minutes}m ${seconds}s`;
}
```

- [ ] **Step 4: Run unit tests to verify pass**

Run: `bun test test/subagent-card-parser.test.ts`
Expected: PASS (5/5 tests passing)

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/utils/toolUtils.ts test/subagent-card-parser.test.ts
git commit -m "feat(subagent): add subagent tool classifier, meta extractor, and output cleaner"
```

---

### Task 2: Subagent Detail Modal Component

**Files:**
- Create: `frontend/src/lib/components/chat/SubagentDetailModal.svelte`
- Test: `test/subagent-modal.test.ts`

**Interfaces:**
- Consumes:
  - `visible: boolean`
  - `toolCall: ToolCall`
  - `meta: SubagentParsedMeta`
  - `onClose: () => void`
- Produces: Full interactive modal popup with live timer, prompt preview, clean markdown output, and quick copy triggers.

- [ ] **Step 1: Write unit test in `test/subagent-modal.test.ts`**

```ts
import { describe, it, expect } from 'vitest';
import { extractSubagentMeta, cleanSubagentOutput } from '../frontend/src/lib/utils/toolUtils';

describe('Subagent Detail Modal Data Flow', () => {
  it('prepares complete modal state from toolCall parameters and result', () => {
    const toolCall = {
      id: 'tc-sub-1',
      tool: 'spawn_subagent',
      status: 'completed' as const,
      startTime: 1000,
      endTime: 85000,
      params: {
        description: 'Profile frontend bundle sizes',
        prompt: 'Run vite build and analyze chunk size distribution',
        subagent_type: 'performance-audit',
        model: 'grok-beta',
        isolation: 'none'
      },
      result: 'Bundle analysis complete. Total vendor size: 412kB.'
    };

    const meta = extractSubagentMeta(toolCall.params);
    const cleanedOutput = cleanSubagentOutput(toolCall.result);

    expect(meta.description).toBe('Profile frontend bundle sizes');
    expect(meta.prompt).toBe('Run vite build and analyze chunk size distribution');
    expect(meta.subagentType).toBe('performance-audit');
    expect(cleanedOutput).toBe('Bundle analysis complete. Total vendor size: 412kB.');
  });
});
```

- [ ] **Step 2: Run test to verify pass**

Run: `bun test test/subagent-modal.test.ts`
Expected: PASS

- [ ] **Step 3: Create `frontend/src/lib/components/chat/SubagentDetailModal.svelte`**

Create the modern modal component:

```svelte
<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import type { ToolCall } from '$lib/stores/session.svelte';
  import {
    type SubagentParsedMeta,
    cleanSubagentOutput,
    formatSubagentDuration
  } from '$lib/utils/toolUtils';
  import { renderMarkdown } from '$lib/utils/markdownRenderer';
  import {
    Bot,
    X,
    Clock,
    Copy,
    Check,
    Loader2,
    CheckCircle2,
    XCircle,
    Cpu,
    Shield,
    ExternalLink,
    Code,
    FileText
  } from 'lucide-svelte';

  interface Props {
    visible: boolean;
    toolCall: ToolCall;
    meta: SubagentParsedMeta;
    onClose: () => void;
  }

  let { visible = false, toolCall, meta, onClose }: Props = $props();

  let copiedPrompt = $state(false);
  let copiedOutput = $state(false);
  let now = $state(Date.now());
  let showRawPayload = $state(false);

  // Active ticking timer for running subagents
  $effect(() => {
    if (!visible || toolCall.status !== 'running') return;
    const interval = setInterval(() => {
      now = Date.now();
    }, 1000);
    return () => clearInterval(interval);
  });

  const elapsedDuration = $derived.by(() => {
    if (!toolCall.startTime) return null;
    const end = toolCall.endTime || (toolCall.status === 'running' ? now : toolCall.startTime);
    return formatSubagentDuration(Math.max(0, end - toolCall.startTime));
  });

  const cleanedOutput = $derived.by(() => {
    return cleanSubagentOutput(toolCall.result);
  });

  async function copyPrompt() {
    if (!meta.prompt) return;
    try {
      await navigator.clipboard.writeText(meta.prompt);
      copiedPrompt = true;
      setTimeout(() => (copiedPrompt = false), 2000);
    } catch {
      // ignore
    }
  }

  async function copyOutput() {
    if (!cleanedOutput) return;
    try {
      await navigator.clipboard.writeText(cleanedOutput);
      copiedOutput = true;
      setTimeout(() => (copiedOutput = false), 2000);
    } catch {
      // ignore
    }
  }

  function handleKeyDown(e: KeyboardEvent) {
    if (!visible) return;
    if (e.key === 'Escape') {
      e.preventDefault();
      onClose();
    }
  }

  onMount(() => {
    window.addEventListener('keydown', handleKeyDown);
  });

  onDestroy(() => {
    window.removeEventListener('keydown', handleKeyDown);
  });
</script>

{#if visible}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm animate-in fade-in duration-150 select-text font-sans"
    onclick={(e) => {
      if (e.target === e.currentTarget) onClose();
    }}
    onkeydown={handleKeyDown}
  >
    <!-- Modal Card Container -->
    <div
      class="w-full max-w-3xl max-h-[85vh] flex flex-col bg-ant-bg border border-violet-500/30 rounded-xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150"
    >
      <!-- Header -->
      <div class="flex items-center justify-between px-5 py-3.5 bg-ant-bg-secondary/70 border-b border-white/5 select-none">
        <div class="flex items-center space-x-3 min-w-0">
          <div class="w-8 h-8 rounded-lg bg-violet-500/15 border border-violet-500/30 flex items-center justify-center flex-shrink-0 text-violet-400">
            {#if toolCall.status === 'running'}
              <Loader2 size={18} class="animate-spin text-violet-400" />
            {:else}
              <Bot size={18} />
            {/if}
          </div>
          <div class="min-w-0">
            <div class="flex items-center space-x-2">
              <span class="text-xs font-semibold uppercase tracking-wider text-violet-400 font-mono">Subagent Process</span>
              {#if elapsedDuration !== null}
                <span class="inline-flex items-center text-[11px] font-mono font-medium px-2 py-0.5 rounded-full bg-violet-500/10 text-violet-300 border border-violet-500/20">
                  <Clock size={11} class="mr-1" />
                  {elapsedDuration}
                </span>
              {/if}
            </div>
            <h2 class="text-sm font-semibold text-ant-text truncate mt-0.5">
              {meta.description}
            </h2>
          </div>
        </div>

        <!-- Right: Status Badge & Close Button -->
        <div class="flex items-center space-x-3 flex-shrink-0">
          {#if toolCall.status === 'running'}
            <span class="inline-flex items-center text-xs font-medium text-violet-400 animate-pulse bg-violet-500/10 px-2.5 py-1 rounded-md border border-violet-500/20">
              <Loader2 size={12} class="animate-spin mr-1.5" />
              Running
            </span>
          {:else if toolCall.status === 'completed'}
            <span class="inline-flex items-center text-xs font-medium text-emerald-400 bg-emerald-500/10 px-2.5 py-1 rounded-md border border-emerald-500/20">
              <CheckCircle2 size={12} class="mr-1.5" />
              Completed
            </span>
          {:else if toolCall.status === 'error'}
            <span class="inline-flex items-center text-xs font-medium text-rose-400 bg-rose-500/10 px-2.5 py-1 rounded-md border border-rose-500/20">
              <XCircle size={12} class="mr-1.5" />
              Failed
            </span>
          {/if}

          <button
            type="button"
            onclick={onClose}
            class="text-ant-text-muted hover:text-ant-text hover:bg-white/5 p-1 rounded-md transition-colors"
            title="Close (Escape)"
          >
            <X size={16} />
          </button>
        </div>
      </div>

      <!-- Quick Metadata Bar -->
      <div class="flex items-center flex-wrap gap-2 px-5 py-2.5 bg-ant-bg-tertiary/40 border-b border-white/5 text-xs text-ant-text-secondary font-mono select-none">
        <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
          <Bot size={12} class="text-violet-400" />
          <span>Role: {meta.subagentType}</span>
        </span>
        {#if meta.model}
          <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
            <Cpu size={12} class="text-indigo-400" />
            <span>Model: {meta.model}</span>
          </span>
        {/if}
        {#if meta.isolation}
          <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
            <Shield size={12} class="text-amber-400" />
            <span>Isolation: {meta.isolation}</span>
          </span>
        {/if}
        <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded bg-white/5 text-ant-text">
          <span>Execution: {meta.isBackground ? 'Background' : 'Foreground'}</span>
        </span>
      </div>

      <!-- Modal Body (Scrollable) -->
      <div class="flex-1 overflow-y-auto p-5 space-y-4 scrollbar-thin">
        <!-- 1. Task Instructions (Prompt) -->
        <div class="rounded-lg border border-white/10 bg-ant-bg-secondary/40 overflow-hidden">
          <div class="flex items-center justify-between px-3.5 py-2 bg-ant-bg-tertiary/60 border-b border-white/5 text-xs select-none">
            <span class="font-semibold text-ant-text flex items-center gap-1.5">
              <FileText size={13} class="text-violet-400" />
              Task Prompt
            </span>
            <button
              type="button"
              onclick={copyPrompt}
              class="flex items-center space-x-1 text-xs text-ant-text-secondary hover:text-ant-text px-1.5 py-0.5 rounded hover:bg-white/5 transition-colors"
            >
              {#if copiedPrompt}
                <Check size={12} class="text-emerald-400" />
                <span class="text-emerald-400">Copied</span>
              {:else}
                <Copy size={12} />
                <span>Copy</span>
              {/if}
            </button>
          </div>
          <div class="p-3.5 text-xs font-mono text-ant-text leading-relaxed whitespace-pre-wrap max-h-48 overflow-y-auto scrollbar-thin select-text">
            {meta.prompt || 'No prompt instructions recorded.'}
          </div>
        </div>

        <!-- 2. Subagent Result / Live Activity -->
        <div class="rounded-lg border border-white/10 bg-ant-bg-secondary/40 overflow-hidden">
          <div class="flex items-center justify-between px-3.5 py-2 bg-ant-bg-tertiary/60 border-b border-white/5 text-xs select-none">
            <span class="font-semibold text-ant-text flex items-center gap-1.5">
              <Code size={13} class="text-emerald-400" />
              Output of Subagent
            </span>
            {#if cleanedOutput}
              <button
                type="button"
                onclick={copyOutput}
                class="flex items-center space-x-1 text-xs text-ant-text-secondary hover:text-ant-text px-1.5 py-0.5 rounded hover:bg-white/5 transition-colors"
              >
                {#if copiedOutput}
                  <Check size={12} class="text-emerald-400" />
                  <span class="text-emerald-400">Copied</span>
                {:else}
                  <Copy size={12} />
                  <span>Copy</span>
                {/if}
              </button>
            {/if}
          </div>

          <div class="p-4 text-xs leading-relaxed max-h-80 overflow-y-auto scrollbar-thin select-text">
            {#if toolCall.status === 'running' && !cleanedOutput}
              <div class="flex items-center justify-center py-8 space-x-3 text-ant-text-muted font-mono">
                <Loader2 size={16} class="animate-spin text-violet-400" />
                <span>Subagent is actively executing tasks in the background...</span>
              </div>
            {:else if cleanedOutput}
              <div class="prose prose-invert prose-xs max-w-none">
                {@html renderMarkdown(cleanedOutput)}
              </div>
            {:else}
              <div class="py-4 text-center text-ant-text-muted font-mono">
                No output returned from subagent.
              </div>
            {/if}
          </div>
        </div>

        <!-- 3. Collapsible Raw Arguments -->
        <div class="rounded-lg border border-white/5 bg-ant-bg-tertiary/20 overflow-hidden text-xs">
          <button
            type="button"
            onclick={() => (showRawPayload = !showRawPayload)}
            class="w-full flex items-center justify-between px-3.5 py-2 hover:bg-white/5 text-ant-text-muted hover:text-ant-text font-mono text-left transition-colors"
          >
            <span>Raw Parameters & Config</span>
            <span class="text-[10px]">{showRawPayload ? '▲ Hide' : '▼ View'}</span>
          </button>
          {#if showRawPayload}
            <div class="p-3 border-t border-white/5 bg-black/40 overflow-x-auto max-h-40 scrollbar-thin">
              <pre class="font-mono text-[11px] text-ant-text-muted">{JSON.stringify(toolCall.params, null, 2)}</pre>
            </div>
          {/if}
        </div>
      </div>

      <!-- Footer -->
      <div class="flex items-center justify-between px-5 py-3 bg-ant-bg-secondary/40 border-t border-white/5 text-xs text-ant-text-muted select-none">
        <span class="font-mono">Press Escape to close</span>
        <button
          type="button"
          onclick={onClose}
          class="px-3 py-1.5 rounded-md bg-white/10 hover:bg-white/15 text-ant-text font-medium transition-colors"
        >
          Close
        </button>
      </div>
    </div>
  </div>
{/if}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/components/chat/SubagentDetailModal.svelte test/subagent-modal.test.ts
git commit -m "feat(ui): add SubagentDetailModal component with live timer and formatted markdown output"
```

---

### Task 3: Integrate Subagent Card & Live Ticking Timer into `ToolCallCard.svelte`

**Files:**
- Modify: `frontend/src/lib/components/chat/ToolCallCard.svelte`
- Test: `test/subagent-card-render.test.ts`

**Interfaces:**
- Consumes:
  - `isSubagentToolCall`, `extractSubagentMeta`, `formatSubagentDuration` from `toolUtils.ts`
  - `SubagentDetailModal.svelte`
- Produces:
  - Dedicated `subagent` branch in `toolParsed`
  - Live 1-second reactive ticking interval during `toolCall.status === 'running'`
  - Click-to-open and explicit Inspect pill button opening `SubagentDetailModal`

- [ ] **Step 1: Write unit test in `test/subagent-card-render.test.ts`**

```ts
import { describe, it, expect } from 'vitest';
import { isSubagentToolCall, extractSubagentMeta } from '../frontend/src/lib/utils/toolUtils';

describe('ToolCallCard Subagent Variant Derivation', () => {
  it('correctly derives subagent card attributes for spawn_subagent tool call', () => {
    const rawTool = 'spawn_subagent';
    const params = {
      description: 'Find memory leaks in cache',
      prompt: 'Inspect heap snapshot',
      subagent_type: 'explore'
    };

    expect(isSubagentToolCall(rawTool, params)).toBe(true);
    const meta = extractSubagentMeta(params);

    expect(meta.description).toBe('Find memory leaks in cache');
    expect(meta.subagentType).toBe('explore');
  });
});
```

- [ ] **Step 2: Run test to verify pass**

Run: `bun test test/subagent-card-render.test.ts`
Expected: PASS

- [ ] **Step 3: Update `frontend/src/lib/components/chat/ToolCallCard.svelte`**

1. Import `isSubagentToolCall`, `extractSubagentMeta`, `formatSubagentDuration` and `SubagentDetailModal`.
2. Add state `showSubagentModal = $state(false)` and active `now = $state(Date.now())` with a 1s interval when `toolCall.status === 'running'`.
3. In `toolParsed`, add subagent resolution before fallback:
```ts
    // Subagent Autonomous Delegation
    if (isSubagentToolCall(rawTool, p)) {
      const subagentMeta = extractSubagentMeta(p);
      return {
        verb: 'Subagent',
        type: 'subagent' as const,
        target: subagentMeta.description,
        meta: subagentMeta,
        icon: Bot,
        iconClass: 'text-violet-400'
      };
    }
```
4. If `toolParsed.type === 'subagent'`, format duration using live `now` when running:
```ts
  const formattedDuration = $derived.by(() => {
    if (!toolCall.startTime) return null;
    const end = toolCall.endTime || (toolCall.status === 'running' ? now : toolCall.startTime);
    return formatSubagentDuration(Math.max(0, end - toolCall.startTime));
  });
```
5. On the card header:
   - For subagent cards, add a distinct violet border / accent.
   - Display an `Inspect` / `View Details` button that opens `showSubagentModal = true`.
   - When clicking anywhere on a subagent card, open `showSubagentModal = true`.
6. Mount `<SubagentDetailModal>` at the bottom of `ToolCallCard.svelte`.

- [ ] **Step 4: Run all unit and integration tests**

Run: `bun test test/subagent-card-parser.test.ts test/subagent-modal.test.ts test/subagent-card-render.test.ts`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/components/chat/ToolCallCard.svelte test/subagent-card-render.test.ts
git commit -m "feat(ui): render dedicated violet subagent card with live timer and inspection modal trigger"
```

---

### Task 4: End-to-End Build, Verification and Packaging

**Files:**
- All modified files

- [ ] **Step 1: Run frontend typecheck**

Run: `cd frontend && pnpm check && cd ..`
Expected: 0 errors.

- [ ] **Step 2: Run all Go unit and integration tests**

Run: `go test ./...`
Expected: PASS.

- [ ] **Step 3: Run all Bun / TypeScript tests**

Run: `bun test test/subagent-card-parser.test.ts test/subagent-modal.test.ts test/subagent-card-render.test.ts test/quit-shortcut-intent.test.ts test/session-running-guard.test.ts test/dom-window-flush.test.ts test/turn-timer-isolation.test.ts`
Expected: All tests PASS.

- [ ] **Step 4: Build desktop application**

Run: `wails build -platform darwin/arm64 -clean`
Expected: Build succeeds and packages `build/bin/aethergrok.app`.
