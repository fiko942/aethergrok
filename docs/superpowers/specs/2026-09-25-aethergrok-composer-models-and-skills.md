# Specification: AetherGrok Adaptive Controls, Model Discovery, and Skill System

**Date**: 2026-09-25  
**Version**: 1.1.0  
**Status**: Implemented & Verified  
**Component**: Prompt Composer, Model Selector, Reasoning Effort, Skill Catalog Table, Slash Autocomplete  
**Target Repository**: `fiko942/grok-build`

---

## 1. Overview & Context

AetherGrok Studio is a native developer GUI environment for Grok Agentic AI CLI built with **Go 1.24**, **Wails v2**, **Svelte 5 Runes**, and **Ant Design Design Tokens**. This specification documents the design and architecture of the refined prompt composer controls, dynamic CLI model discovery, structured tabular skills catalog, and inline slash command autocompletion.

---

## 2. Core Architectural Pillars

### 2.1 Theme-Adaptive Typography & Contrast Engine
- **Ant Design Semantic Tokens**: Full dynamic adherence to `--ant-text`, `--ant-text-secondary`, `--ant-text-muted`, `--ant-bg`, and `--ant-bg-secondary`.
- **Elimination of Hardcoded Overrides**: Replaced static `text-white` styles on `<textarea>`, message header badges (`You`, `Grok`), and markdown `<strong>` tags with reactive design tokens.
- **Bi-Directional Contrast**:
  - *Clean Light Theme (`light-antd`)*: Text renders in high-contrast charcoal (`#1f1f1f`).
  - *Dark Studio (`dark-studio`)*: Text renders in soft pearl slate (`#f3f4f6`).
  - *OLED High Contrast (`dark-high-contrast`)*: Text renders in pure white (`#ffffff`).

### 2.2 Dynamic Model Discovery & ModelSelectDropdown
- **Backend Model Discovery Engine** (`pkg/grokrunner/models.go`):
  - Queries local CLI via `grok models` subprocess with 3-second timeout fallback.
  - Automatically parses standard models (`9router`, `9router-general-purpose`, `9router-explore`, `9router-plan`) along with default flags.
  - Exposes Go method `GetAvailableModels()` via Wails desktop bridge.
- **Frontend Popover Selector** (`frontend/src/lib/components/chat/ModelSelectDropdown.svelte`):
  - Replaces cyclic click toggles with an upward popover modal (`bottom-full`) equipped with a **Search Input Box**.
  - Real-time client-side substring search against model IDs, display titles, and capability descriptions.
  - Integrated "Sync" action triggering on-demand rescan from local Grok CLI.
  - Persists active selection directly to `settingsStore.defaultModel`.

### 2.3 Granular Reasoning Effort Controls (`ReasoningEffortDropdown.svelte`)
- **Direct Choice Popover**: Provides single-click access to thinking depth levels:
  - **⚡ Low Effort (`Fastest`)**: Minimal chain-of-thought token overhead for rapid answers.
  - **⚖️ Medium Effort (`Recommended`)**: Balanced multi-step tool execution and code review.
  - **🧠 High Effort (`Thorough`)**: Deep exhaustive planning, edge-case analysis, and verification.
- **Z-Index & Clipping Prevention**: Raised z-index to `z-[100]` and removed container `overflow-hidden` constraints to guarantee popover rendering visibility.
- **Store Sync**: Binds with `settingsStore.defaultReasoningEffort`.

### 2.4 Tabular Skill Hub Catalog (`SkillCatalog.svelte`)
- **Structured Data Table View**: Replaced card grid with an informative Ant Design table.
- **Table Columns**:
  1. *Skill / Command*: Skill name, category icon, trigger badge `/{name}`, and scope (`grok` vs `agents`).
  2. *Category*: Color-coded badge (`Frontend`, `Backend`, `Design`, `Agents`, `Tools`).
  3. *Description & Triggers*: Concise functional explanation.
  4. *Tags*: Extensible tag chips.
  5. *Action*: Quick `Use` button to insert directly into the active prompt composer.

### 2.5 Inline Slash Command Autocomplete (`SlashCommandPopup.svelte`)
- **Instant Trigger**: Typing `/` at word boundaries inside the prompt composer triggers the autocomplete overlay.
- **Real-Time Substring Filtering**: Dynamically filters installed skills matching characters after `/`.
- **Keyboard Navigation**:
  - `ArrowUp` / `ArrowDown`: Moves active selection through results.
  - `Enter` / `Tab`: Inserts `/{skill.name} ` at cursor position.
  - `Escape`: Dismisses autocomplete popup without affecting prompt text.

---

## 3. Data Contracts & Structures

### 3.1 Model Info Interface (Go & TypeScript)
```go
type ModelInfo struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Description string `json:"description"`
    IsDefault   bool   `json:"isDefault"`
}
```

```typescript
export interface ModelOption {
    id: string;
    name: string;
    description?: string;
    isDefault?: boolean;
}
```

### 3.2 Reasoning Effort Type
```typescript
export type ReasoningEffortLevel = 'low' | 'medium' | 'high';
```

---

## 4. Verification & Validation Metrics

| Checkpoint | Target | Result |
| :--- | :--- | :--- |
| **Go Test Suite** | `go test ./...` | Pass (0 failures) |
| **Svelte 5 Diagnostics** | `svelte-check --tsconfig ./tsconfig.json` | 0 errors, 0 warnings |
| **Frontend Build** | `vite build` | Clean production bundle |
| **Popover Overflow** | Popovers render unclipped over chat viewport | Confirmed via z-[100] & container fix |
| **Theme Contrast** | Light mode (`light-antd`) and Dark mode readability | Confirmed with semantic tokens |
