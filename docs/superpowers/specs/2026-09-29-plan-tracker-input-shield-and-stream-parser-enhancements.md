# Specification: Plan Tracker, Input Shielding & ACP Stream Parser (v1.0.8)

**Date**: 2026-09-29  
**Version**: 1.0.8  
**Scope**: Frontend UI & Store Architecture, Backend ACP Stream Parser, and Input Safety

## 1. System Components

### Frontend Components
- `FloatingPlanTracker.svelte`: Collapsible, floating on-canvas HUD showing the current active plan, active step with spinner/progress ring, completed count, and quick toggle to open full plan panel.
- `PlanPanel.svelte`: Full-height sidebar tab detailing all planned steps, execution state, status badges (`pending`, `in_progress`, `completed`, `cancelled`), and markdown descriptions.
- `DiffCard.svelte` & `TurnDiffSummary.svelte`: Detailed file-by-file diff cards with visual additions/deletions counts, unified diff formatting, and syntax highlighting.
- `ToolCallCard.svelte`: Comprehensive tool execution view supporting nested parameters, expandable standard output/error, duration tracking, and status chips.

### Frontend Stores & Utilities
- `plan.svelte.ts`: Global reactive Svelte 5 store maintaining the active plan hierarchy, step mutations, and persistence across session navigation.
- `planParser.ts`: High-performance parser extracting todo checklists and structured plan tokens from agent responses.
- `inputShield.svelte.ts`: Context-aware input barrier ensuring global shortcuts do not collide with active terminal typing, dialog inputs, or code editors.

### Backend Engine
- `pkg/grokrunner/stream_parser.go`: High-throughput NDJSON stream parser with resilient handling of truncated tokens, partial JSON structures, and metadata events.
- `pkg/grokrunner/session_scanner.go`: Multi-workspace session indexer with disk cache validation and instant metadata hydration.

## 2. Behavioral Guarantees
- Plan updates immediately reflect in UI within 16ms of agent stream emission.
- Keyboard navigation inside terminal panels does not propagate to application-level hotkeys.
- Streaming parser operates without memory leaks or unhandled JSON syntax panics on malformed lines.
