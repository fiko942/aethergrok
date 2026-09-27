# Spec: Dark Mode Harsh Border Softening & Global Theme Harmonization

## Problem Statement
In dark mode (`dark-studio` and `dark-high-contrast`), multiple UI borders appeared excessively bright and stark against dark background surfaces (`#141414`, `#1a1a1a`, `#000000`). This visual friction created jarring separation lines around headers, tab strips, panels, and cards.

## Key Changes
1. **Global CSS Variable Tokens (`frontend/src/app.css`)**:
   - `dark-studio`:
     - `--ant-border`: Softened from `rgba(255, 255, 255, 0.06)` to `rgba(255, 255, 255, 0.04)`.
     - `--ant-border-secondary`: Softened from `rgba(255, 255, 255, 0.03)` to `rgba(255, 255, 255, 0.02)`.
   - `dark-high-contrast`:
     - `--ant-border`: Softened from `rgba(255, 255, 255, 0.14)` to `rgba(255, 255, 255, 0.07)`.
     - `--ant-border-secondary`: Softened from `rgba(255, 255, 255, 0.08)` to `rgba(255, 255, 255, 0.04)`.
2. **Harmonized Component Rendering**:
   - Headers, Session Tabs, Terminal Docks, Message Cards, and Modals leverage `--ant-border` for an elegant, low-contrast visual structure that matches macOS Studio Dark guidelines.

## Verification
- Clean build completed via `cd frontend && npm run build`.
- Frontend assets generated without type or style compilation errors.
