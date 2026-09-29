# Complete Svelte 5 Architecture & UI Modernization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Transform and standardize the entire desktop frontend architecture into pure Svelte 5 (Runes), eliminating all React dependencies and virtual DOM runtime overhead, while redesigning modals, popups, and controls with zero latency and 100% functional parity.

**Architecture:** Compile-time fine-grained reactive architecture using Svelte 5 Runes (`$state`, `$derived`, `$effect`), pure CSS variables/Tailwind tokens, and Wails Go IPC bridge. Build output consists of direct, lean HTML, CSS, and vanilla JS without framework runtime bloat.

**Tech Stack:** Svelte 5, Vite 6, TypeScript 5, Tailwind CSS, Lucide Svelte, Wails v2.8+ Go Backend.

**Spec:** `docs/superpowers/specs/2026-09-29-svelte-architecture-migration.md`

## Global Constraints

- Zero React runtime dependencies (`react`, `react-dom`, `@types/react`, React JSX transformers).
- Compile output must be pure vanilla JS/CSS bundles optimized for production desktop WebView.
- 100% functional parity across all desktop features: ACP streaming, terminal docking, hotkeys, settings, updater, audio/voice visualizer, and file attachments.
- Svelte 5 Runes reactivity standard: strictly use `$state`, `$derived`, `$effect`, and `.svelte.ts` modules.
- Zero type errors on `svelte-check` and zero compilation warnings on `vite build`.

---

### Task 1: React Dependency & Tooling Purge

**Files:**
- Modify: `package.json`
- Modify: `frontend/package.json`
- Modify: `web/package.json`

**Interfaces:**
- Consumes: Existing package configurations.
- Produces: Clean dependency manifests free of React, `@babel/*-react*`, and JSX transformers.

- [ ] **Step 1: Inspect and identify React remnants in dependencies**

Check root, `frontend/`, and `web/` packages for any `@babel/preset-react`, `@babel/plugin-transform-react-jsx`, `eslint-plugin-react`, `react-is`, or `@types/react`.

- [ ] **Step 2: Update `frontend/package.json` to lock pure Svelte 5 dependencies**

```json
{
  "name": "aethergrok-frontend",
  "private": true,
  "version": "1.0.6",
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "check": "svelte-check --tsconfig ./tsconfig.json"
  },
  "dependencies": {
    "@ant-design/colors": "^7.2.0",
    "@types/prismjs": "^1.26.6",
    "@xterm/addon-fit": "^0.11.0",
    "@xterm/xterm": "^6.0.0",
    "diff2html": "^3.4.48",
    "katex": "^0.16.21",
    "lucide-svelte": "^0.475.0",
    "marked": "^18.0.14",
    "prismjs": "^1.30.0"
  },
  "devDependencies": {
    "@sveltejs/vite-plugin-svelte": "^5.0.3",
    "@types/katex": "^0.16.7",
    "@types/marked": "^5.0.2",
    "@types/node": "^22.13.5",
    "autoprefixer": "^10.4.20",
    "postcss": "^8.5.3",
    "svelte": "^5.20.2",
    "svelte-check": "^4.1.4",
    "tailwindcss": "^3.4.17",
    "typescript": "^5.7.3",
    "vite": "^6.1.1"
  }
}
```

- [ ] **Step 3: Run package lock verification**

Run: `npm --prefix frontend install --dry-run`
Expected: 0 react packages resolved in direct dependencies.

- [ ] **Step 4: Commit**

```bash
git add frontend/package.json
git commit -m "chore(frontend): purge legacy react dependencies and lock pure svelte 5 stack"
```

---

### Task 2: Vite & Svelte 5 Compiler Optimization for Production Desktop

**Files:**
- Modify: `frontend/vite.config.ts`
- Modify: `frontend/svelte.config.js`
- Modify: `frontend/tsconfig.json`

**Interfaces:**
- Consumes: Svelte Vite plugin and TypeScript configurations.
- Produces: Optimized tree-shaking, CSS code splitting, and zero-overhead compilation.

- [ ] **Step 1: Configure `frontend/svelte.config.js` for Svelte 5 runes mode**

```javascript
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

export default {
  preprocess: vitePreprocess(),
  compilerOptions: {
    runes: true
  }
};
```

- [ ] **Step 2: Optimize `frontend/vite.config.ts` for lean desktop builds**

```typescript
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

export default defineConfig({
  plugins: [svelte()],
  build: {
    target: 'esnext',
    minify: 'esbuild',
    cssMinify: true,
    rollupOptions: {
      output: {
        manualChunks: undefined
      }
    }
  },
  server: {
    port: 5173,
    strictPort: true
  }
});
```

- [ ] **Step 3: Verify TypeScript configuration for `.svelte.ts` modules**

Ensure `frontend/tsconfig.json` includes `moduleResolution: "bundler"`, `strict: true`, and covers `**/*.svelte.ts`.

- [ ] **Step 4: Run typecheck to verify compiler setup**

Run: `npm --prefix frontend run check`
Expected: PASS with 0 errors.

- [ ] **Step 5: Commit**

```bash
git add frontend/vite.config.ts frontend/svelte.config.js frontend/tsconfig.json
git commit -m "build(frontend): optimize svelte 5 runes compiler and vite bundle pipeline"
```

---

### Task 3: Reactive State Engine Architecture (Runes-Only Stores)

**Files:**
- Modify: `frontend/src/lib/stores/settings.svelte.ts`
- Modify: `frontend/src/lib/stores/session.svelte.ts`
- Modify: `frontend/src/lib/stores/updater.svelte.ts`
- Modify: `frontend/src/lib/stores/permissions.svelte.ts`

**Interfaces:**
- Consumes: Wails IPC runtime (`window.go.main.App`).
- Produces: Reactive state classes using `$state` and `$derived` with automatic persistence.

- [ ] **Step 1: Implement Svelte 5 reactive settings store**

Verify and ensure `settings.svelte.ts` manages all configuration parameters (delay, timeouts, threads, theme, audio devices, shortcuts) using `$state`:

```typescript
export class SettingsStore {
  delayUpdate = $state(1000);
  threadUpdate = $state(5);
  loginTimeout = $state(60);
  zoomLevel = $state(100);
  theme = $state<'dark' | 'light'>('dark');
  
  setDelay(val: number) { this.delayUpdate = val; }
  setThreads(val: number) { this.threadUpdate = val; }
  setTimeout(val: number) { this.loginTimeout = val; }
  setZoom(val: number) { this.zoomLevel = val; }
}

export const settingsStore = new SettingsStore();
```

- [ ] **Step 2: Verify state hydration and storage synchronization**

Run test suites for store persistence and reactive bindings.

- [ ] **Step 3: Commit**

```bash
git add frontend/src/lib/stores/settings.svelte.ts
git commit -m "refactor(frontend): solidify svelte 5 runes reactive store engine"
```

---

### Task 4: High-Performance Modal & Dialog Component Overhaul

**Files:**
- Modify: `frontend/src/lib/components/layout/SettingsModal.svelte`
- Create: `frontend/src/lib/components/ui/NumericSlider.svelte`
- Modify: `frontend/src/lib/components/ui/ConfirmModal.svelte`
- Modify: `frontend/src/lib/components/ui/KeyRecorderModal.svelte`

**Interfaces:**
- Consumes: `settingsStore`, `lucide-svelte` icons, Ant Design tokens.
- Produces: Zero-lag modal dialogs, smooth keyboard trap, and accessible numeric slider controls.

- [ ] **Step 1: Create reusable `NumericSlider.svelte` component**

```svelte
<script lang="ts">
  interface Props {
    label: string;
    description?: string;
    value: number;
    min: number;
    max: number;
    step?: number;
    unit?: string;
    marks?: Array<{ val: number; label: string }>;
    onchange?: (val: number) => void;
  }

  let {
    label,
    description,
    value = $bindable(),
    min,
    max,
    step = 1,
    unit = '',
    marks = [],
    onchange
  }: Props = $props();

  function handleInput(e: Event) {
    const val = Number((e.target as HTMLInputElement).value);
    value = Math.min(Math.max(val, min), max);
    onchange?.(value);
  }
</script>

<div class="rounded-xl border border-ant-border-secondary/70 bg-ant-bg-elevated/40 p-4 transition-colors hover:border-ant-primary/30">
  <div class="flex items-center justify-between mb-3">
    <div>
      <span class="text-sm font-medium text-ant-text-primary">{label}</span>
      {#if description}
        <p class="text-xs text-ant-text-tertiary mt-0.5">{description}</p>
      {/if}
    </div>
    <div class="flex items-center gap-1.5 bg-ant-bg-container px-2.5 py-1 rounded-lg border border-ant-border-secondary">
      <input
        type="number"
        {min}
        {max}
        {step}
        bind:value
        oninput={handleInput}
        class="w-14 text-right text-xs font-semibold text-ant-primary bg-transparent focus:outline-none"
      />
      {#if unit}
        <span class="text-xs text-ant-text-quaternary">{unit}</span>
      {/if}
    </div>
  </div>
  
  <input
    type="range"
    {min}
    {max}
    {step}
    bind:value
    oninput={handleInput}
    class="w-full accent-ant-primary h-1.5 bg-ant-bg-spotlight rounded-lg cursor-pointer"
  />

  {#if marks.length > 0}
    <div class="flex justify-between text-[11px] text-ant-text-quaternary mt-1.5">
      {#each marks as mark}
        <span>{mark.label}</span>
      {/each}
    </div>
  {/if}
</div>
```

- [ ] **Step 2: Modernize `SettingsModal.svelte` with refined layouts and controls**

Integrate the `NumericSlider` components, Lucide icons, seamless tab switching, and clean action buttons.

- [ ] **Step 3: Run component checks**

Run: `npm --prefix frontend run check`
Expected: 0 errors across Svelte components.

- [ ] **Step 4: Commit**

```bash
git add frontend/src/lib/components/ui/NumericSlider.svelte frontend/src/lib/components/layout/SettingsModal.svelte
git commit -m "feat(frontend): implement lightweight numeric slider and modern settings modal"
```

---

### Task 5: Full Application Build & Functional Parity Verification

**Files:**
- Test: `frontend/src/App.svelte`
- Build: `frontend/dist/`

**Interfaces:**
- Consumes: Complete Svelte component tree.
- Produces: Production build artifacts (`dist/index.html`, `dist/assets/*.js`, `dist/assets/*.css`).

- [ ] **Step 1: Execute production build**

Run: `npm --prefix frontend run build`
Expected: Build succeeds in under 3 seconds with compact asset sizes (< 1.5MB total bundle including KaTeX and Xterm).

- [ ] **Step 2: Verify Wails backend compilation**

Run: `go build -o build/bin/aethergrok.exe .`
Expected: Executable compiles cleanly with embedded Svelte static assets.

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore(release): complete pure svelte 5 architecture and production verification"
```
