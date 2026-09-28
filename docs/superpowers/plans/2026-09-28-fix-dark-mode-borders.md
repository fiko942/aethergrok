# Design Audit & Implementation Plan: Dark Mode High-Contrast Border Reduction

> **Goal:** Eliminate harsh and excessive white border lines in dark mode across all modal windows, cards, list items, badges, accordion headers, and settings panels as reported in screenshots, ensuring a sleek, seamless, and authentic dark UI aesthetic aligned with `/writing-plans`, `/design-taste-frontend`, and `/ui-ux-pro-max`.

---

### Root Cause Analysis from User Screenshots
In Image #1, #2, and #3 of the `Settings & Workspace Preferences` and `Updates & Releases` tab:
1. **UpdatesTab Accordion and Card Headers**:
   - `border-ant-border-secondary` and explicit `border border-amber-500/30` / `border border-ant-primary/30` with `border-t border-ant-border-secondary/40`, `border-t border-ant-border-secondary/30`, `border-l-2 border-ant-border-secondary`, and asset download buttons `border border-ant-border-secondary`.
   - In Dark Studio, CSS variable `--ant-border-secondary` fell back or contrasted too harshly against `#141414` / `#1a1a1a` surfaces, creating heavy boxy borders around every single release row, key highlights block, and package button.
2. **SettingsModal and Tabs Navigation**:
   - Card containers (`<Card>` component and raw `div` cards) used explicit `border-ant-border-secondary dark:border-white/5` stacked inside modals that already have borders, causing nested double borders.
   - Status badges and current version indicators used bright contrasting borders.
3. **Global Theme Tokens & CSS Variables (`tokens.ts`, `app.css`)**:
   - `--ant-border` and `--ant-border-secondary` in `app.css` and `tokens.ts` for `dark-studio` were defined with opacity values that, when layered over stacked dark backgrounds with CSS backdrop filters, rendered as noticeable white framing outlines.
   - Using subtle border colors (`rgba(255, 255, 255, 0.03)` / `rgba(255, 255, 255, 0.06)`) or borderless elevated backgrounds creates a much more refined, premium Apple/Ant Design feel.

---

### Implementation Tasks

#### Task 1: Refine Global CSS Border Variables (`frontend/src/app.css` & `frontend/src/lib/antd/tokens.ts`)
- Tone down `--ant-border` and `--ant-border-secondary` in `dark-studio` and `dark-high-contrast` to soft, refined thresholds.
- Update `tokens.ts` so that dark theme border tokens match sleek dark canvas values without harsh outlines.

#### Task 2: Polish `UpdatesTab.svelte` UI/UX
- Remove harsh full-box borders on release rows and replace with subtle background contrast (`bg-ant-bg-secondary/40 hover:bg-ant-bg-secondary/80`).
- Remove redundant internal dividers (`border-t`, `border-l-2`) inside the changelog accordion; use clean typographic hierarchy and comfortable spacing instead.
- Clean up the hero banner (`bg-gradient-to-br from-ant-primary/10 to-transparent`) to use subtle border accents (`border-ant-primary/20`).
- Style package asset buttons with seamless dark pill/card styling (`bg-ant-bg-tertiary/40 hover:bg-ant-bg-tertiary text-ant-text border border-white/[0.04]`).

#### Task 3: Polish `SettingsModal.svelte` & `<Card>` Components
- Audit and refine all tab panels (General, Models, Permissions, Theme, Updates, About) to remove harsh inner white borders and replace with smooth dark surface elevation.
- Ensure selection states (e.g. Model items, Permission mode cards) have refined accent colors without white outlines.

#### Task 4: Verification & Build
- Run `pnpm run build` in `frontend/` to confirm zero errors.
- Run `go test -v ./...` to verify backend integrity.
