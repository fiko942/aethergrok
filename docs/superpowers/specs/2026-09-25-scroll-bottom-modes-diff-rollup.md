# Superpowers Spec: Scroll to Bottom, Agent/Plan Modes, and Diff Summary Rollup

**Date**: 2026-09-25  
**Version**: 1.0.0  
**Scope**: Chat Interface, Agent Modes, Diff Visualization  

---

## Technical Specifications

### 1. Scroll-To-Bottom State Machine
- `showScrollToBottom = distanceFromBottom > 160`
- `unreadActivityBelow`: Activated when new messages or streaming chunks arrive while `autoScrollToBottom === false`. Cleared upon reaching bottom or clicking the button.
- Smooth transition: `containerEl.scrollTo({ top: containerEl.scrollHeight, behavior: 'smooth' })`.

### 2. Plan Mode Review Flow
- Trigger: Assistant response detected as plan (`### Execution Plan`, `## Plan`, or `exit_plan_mode`).
- Interactions:
  - `handlePlanAction('approve')`: Dispatches "The plan is approved. Please implement the changes step by step now."
  - `handlePlanAction('reject')`: Dispatches "I reject the proposed plan. Please stop or suggest an alternative approach."
  - `handlePlanAction('custom', feedback)`: Dispatches customized user feedback for plan revision.

### 3. Turn Diff Rollup
- Aggregation across all tool calls in a turn (`tc.diff`).
- Computes `totalAdded` and `totalRemoved`.
- Renders expandable tree showing full file paths and individual visual diff cards.
