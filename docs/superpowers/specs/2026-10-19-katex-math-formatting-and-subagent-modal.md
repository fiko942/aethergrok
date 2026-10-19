# KaTeX LaTeX Mathematical Rendering, Subagent Scoping & Stop Architecture Specification

**Author:** AetherGrok Engineering
**Date:** 2026-10-19
**Version:** v1.1.16
**Status:** Shipped in v1.1.16

---

## 1. Executive Summary

AetherGrok v1.1.16 delivers three foundational engineering enhancements:
1. **KaTeX-Powered LaTeX Mathematical Formatting:** Full inline and display mathematical rendering across both primary chat messages and autonomous subagent outputs, supporting standard LaTeX delimiters (`$$...$$`, `\[...\]`, `\(...\)`, `$x$`) and advanced environments (`matrix`, `pmatrix`, `bmatrix`, `align`, `cases`, `equation`, `split`).
2. **Deterministic Delimiter Protection & Currency Heuristics:** Zero-collision parsing that prevents common currency expressions (`$100`, `$5.50`, `$10k`, `$50-$100`, `$5 million`) from triggering math rendering, while safeguarding fenced code blocks and inline code spans with pre-marked placeholder tokens.
3. **Subagent Activity Scoping & Streamlined Modal:** Complete isolation of subagent-internal actions (file reads, searches, command executions) from the parent session timeline, coupled with an interactive inspector modal and robust session stop/cancellation.

---

## 2. Technical Architecture & Data Flow

```
[Raw Stream / LLM Output]
          │
          ▼
┌──────────────────────────────────────────────┐
│  Stage 1: Code Block & Span Shielding        │
│  - Replace ```...``` with PREFENCEDCODEBLOCK │
│  - Replace `...` with PREINLINECODESPAN      │
└──────────────────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────────┐
│  Stage 2: LaTeX Environments & Delimiters    │
│  - \begin{...}...\end{...} -> $$...$$        │
│  - \[...\] -> $$...$$ (Display Mode)         │
│  - \(...\) -> $...$ (Inline Mode)            │
└──────────────────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────────┐
│  Stage 3: Currency & Plain Text Protection   │
│  - Evaluate contents against isLikelyCurrency│
│  - Strip unsupported TeX (\label{})          │
│  - Execute KaTeX renderToString              │
│  - Wrap in .katex-display-container (scroll) │
└──────────────────────────────────────────────┘
          │
          ▼
┌──────────────────────────────────────────────┐
│  Stage 4: Marked HTML & Code Restoration     │
│  - Parse standard markdown                   │
│  - Restore code blocks & code spans          │
│  - Render in MessageItem & SubagentModal     │
└──────────────────────────────────────────────┘
```

---

## 3. Core Component Specifications

### 3.1. LaTeX Math Renderer (`frontend/src/lib/utils/markdownRenderer.ts`)

#### Supported Delimiters
- **Display Math:**
  - `$$ ... $$`
  - `\[ ... \]`
  - `\begin{equation} ... \end{equation}`
  - `\begin{align} ... \end{align}`
  - `\begin{gather} ... \end{gather}`
  - `\begin{alignat} ... \end{alignat}`
  - `\begin{multline} ... \end{multline}`
  - `\begin{matrix} ... \end{matrix}`
  - `\begin{pmatrix} ... \end{pmatrix}`
  - `\begin{bmatrix} ... \end{bmatrix}`
  - `\begin{vmatrix} ... \end{vmatrix}`
  - `\begin{Vmatrix} ... \end{Vmatrix}`
  - `\begin{cases} ... \end{cases}`
  - `\begin{array} ... \end{array}`
  - `\begin{split} ... \end{split}`
- **Inline Math:**
  - `\( ... \)`
  - `$ ... $` (single dollar delimiter subjected to currency heuristics)

#### Currency Collision Avoidance Heuristic
```typescript
export function isLikelyCurrency(text: string): boolean {
  const trimmed = text.trim();
  // Matches "$100", "$5.50", "$100,000", "$10k", "$5m", "$2b", "$5 million"
  if (/^\d+([.,]\d+)?\s*(k|m|b|million|billion)?$/i.test(trimmed)) return true;
  // Matches "$10 - $20" or "$10-$20" ranges
  if (/^\d+\s*[-–—]\s*\d+$/.test(trimmed)) return true;
  return false;
}
```

#### Unsupported TeX Filtering
Certain LaTeX macros emitted by models (such as `\label{...}` tags inside equations) are unsupported by KaTeX and trigger noisy errors. The renderer automatically sanitizes these without altering mathematical logic:
```typescript
export function stripUnsupportedTex(tex: string): string {
  if (!tex) return '';
  return tex.replace(/\\label\s*\{[^}]*\}/g, '');
}
```

#### Responsive CSS Display Container
In `frontend/src/app.css`:
```css
@import 'katex/dist/katex.min.css';

.katex-display-container {
  overflow-x: auto;
  overflow-y: hidden;
  max-width: 100%;
  padding: 0.25rem 0;
  margin: 0.5rem 0;
  -webkit-overflow-scrolling: touch;
}
```

---

### 3.2. Subagent Activity Scoping & Detail Inspector

#### Activity Timeline Isolation
- Subagents spawned via `spawn_subagent` emit intermediate progress, tool executions, and file operations.
- Previously, these child operations were attributed to the main conversation turn, causing action list pollution and visual noise.
- **Resolution:**
  - Child tool calls carrying parent task metadata or originating from subagent workers are filtered from the primary `ToolCallCard` list.
  - Subagent cards render a clean, high-level summary with a 1-second live stopwatch timer (`Running · 1m 24s`).
  - Detailed task prompt, model configuration, execution status, and child activity logs are encapsulated inside `SubagentDetailModal.svelte`.

#### Markdown & KaTeX Rendering in Subagent Inspector
Both the subagent prompt, output report, and tool execution logs inside `SubagentDetailModal.svelte` utilize `renderMarkdown()` ensuring math formulas and structured reports display with full fidelity.

---

### 3.3. Process Group Termination & Session Stop

#### Backend Termination (`pkg/grokrunner/runner.go` & `proc_unix.go`)
- Ensures that when `Cancel(sessionID)` is invoked:
  1. The running process group ID (`-pgid`) receives `syscall.SIGTERM`, followed by `syscall.SIGKILL`.
  2. The process handle itself receives immediate `cmd.Process.Kill()`.
  3. The active session map is cleared, releasing process handles and preventing orphaned Grok CLI binaries.

#### Frontend Event Isolation (`frontend/src/lib/components/chat/Composer.svelte`)
- Isolates DOM `PointerEvent` objects from `handleCancelSession` parameter signatures so that accidental event propagation does not override active session identifiers.

---

## 4. Test Verification & Quality Gates

1. **Unit Testing:**
   - `frontend/src/lib/utils/markdownRenderer.test.ts`: 15 comprehensive tests validating all delimiters, matrices, currency heuristics, underscore preservation, and error handling.
   - `test/subagent-card-parser.test.ts` & `test/subagent-modal.test.ts`: 21 tests validating card parsing, state reconciliation, and modal data flow.
   - `test/session-stop-cancel.test.ts`: Tests validating session ID resolution and running tool call finalization upon cancellation.
2. **Full Regression Suite:**
   - 249 test files passed with 5,417 passing unit and integration tests.
3. **Multi-Platform CI Matrix:**
   - Windows amd64 & arm64 NSIS installers built and verified.
   - macOS arm64, amd64, and universal DMGs built and verified.
   - GitHub Release assets published with SHA-256 checksums.
