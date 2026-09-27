# Implementation Plan - Dark Mode Harsh Border Softening

Harmonize dark theme border tokens across the desktop interface to soften stark white outlines while preserving tactile structural definition.

## Proposed Changes

### Global Theme Tokens
#### [frontend/src/app.css](../../frontend/src/app.css)
- Reduce default dark mode border opacity from `0.06` to `0.04` for `--ant-border` and `0.03` to `0.02` for `--ant-border-secondary`.
- Soften dark high contrast mode border opacity from `0.14` to `0.07` for `--ant-border` and `0.08` to `0.04` for `--ant-border-secondary`.

## Verification Plan
### Automated Build & Test
- Run `cd frontend && npm run build` to ensure all CSS variables and Svelte components compile cleanly.
- Verify assets in `frontend/dist`.
