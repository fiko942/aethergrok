# Superpowers Plan: Scroll to Bottom, Agent/Plan/Auto-Accept Modes, and Multi-File Diff Summary Rollup

**Date**: 2026-09-25  
**Component**: Chat Feed, Execution Engine, Diff Viewer, Composer  
**Status**: Completed & Verified  

---

## 1. Architectural Overview & Requirements

Based on user requirements and reference architecture from `phuryn/grok-build-vscode`:

1. **Floating Action Scroll to Bottom**:
   - Floating pill button dynamically rendered when the user scrolls up past 160px from the bottom.
   - Real-time unread activity indicator (`New activity`) when messages/tool events stream while scrolled up.
   - Smooth scroll to bottom on click, re-engaging auto-scroll stick-to-bottom.

2. **Agent Execution & Plan Modes**:
   - Three execution modes: **Agent mode** (`agent`), **Plan mode** (`plan`), and **Auto accept / YOLO** (`yolo`).
   - Intelligent keyword auto-detection in Composer (`coba bikin plan...`, `buat perencanan...`, `/plan`).
   - Interactive `PlanReviewCard` presented on plan proposals with:
     - **Approve & Implement**: Unlocks plan gate and begins step-by-step implementation.
     - **Reject**: Rejects proposal and asks for alternative approaches.
     - **Custom Feedback**: Expandable inline textarea to give specific revision instructions to the agent.
   - Permission mode integration: Auto-approves requests when in `bypassPermissions`, `auto`, or `acceptEdits` mode.

3. **Multi-File Diff Summary Rollup**:
   - Turn-level aggregation of all file edits (`Changed X files: +Y lines, -Z lines`).
   - Interactive accordion list displaying each modified file with full path, individual line diff counts, and expandable inline `DiffCard`.

---

## 2. Implementation Files

- `frontend/src/lib/components/chat/MessageList.svelte`: Added Scroll to Bottom FAB with unread activity indicator and `onPlanAction` prop.
- `frontend/src/lib/components/chat/TurnDiffSummary.svelte`: Created turn-level diff summary rollup with per-file expandable diffs.
- `frontend/src/lib/components/chat/PlanReviewCard.svelte`: Created interactive plan review card with Approve, Reject, and Custom feedback.
- `frontend/src/lib/components/chat/MessageItem.svelte`: Integrated `TurnDiffSummary` and `PlanReviewCard`.
- `frontend/src/lib/components/chat/Composer.svelte`: Added plan mode keyword auto-detection.
- `frontend/src/App.svelte`: Added `handlePlanAction` dispatcher and wired automatic permission approval based on `permissionMode`.
