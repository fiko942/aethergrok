import { describe, it, expect } from 'vitest';
import * as fs from 'fs';
import * as path from 'path';

describe('Tool Call and Subagent Loading State Audit', () => {
  it('ensures ToolCallCard provides an animated loading indicator for running tool calls when expanded', () => {
    const cardContent = fs.readFileSync(
      path.resolve(__dirname, '../frontend/src/lib/components/chat/ToolCallCard.svelte'),
      'utf-8'
    );

    // Verify running indicator exists for general tools (e.g. Search, Fetch, Grep)
    expect(cardContent).toContain("{:else if toolCall.status === 'running'}");
    expect(cardContent).toContain("toolParsed.verb ? `${toolParsed.verb} in progress...` : 'Executing tool...'");
    expect(cardContent).toContain("Loader2 size={13} class=\"animate-spin text-ant-primary shrink-0\"");

    // Verify running indicator exists in Terminal OUT panel
    expect(cardContent).toContain("Executing command...");
    expect(cardContent).toContain("Loader2 size={12} class=\"animate-spin text-indigo-400 shrink-0\"");

    // Verify subagent running indicator exists for output section
    expect(cardContent).toContain("{:else if subagentIsActuallyRunning}");
    expect(cardContent).toContain("Subagent is executing tasks...");
  });

  it('ensures SubagentDetailModal provides active loading indicators for running child actions and final output', () => {
    const modalContent = fs.readFileSync(
      path.resolve(__dirname, '../frontend/src/lib/components/chat/SubagentDetailModal.svelte'),
      'utf-8'
    );

    // Verify running indicator for child action in progress
    expect(modalContent).toContain("Action in progress... waiting for output");
    expect(modalContent).toContain("Loader2 size={12} class=\"animate-spin text-violet-400 shrink-0\"");

    // Verify running indicator for subagent final output
    expect(modalContent).toContain("{:else if subagentIsActuallyRunning}");
    expect(modalContent).toContain("Subagent is executing tasks...");
    expect(modalContent).toContain("Final report will be displayed here once execution completes.");
  });
});
