import { describe, it, expect } from 'vitest';
import * as fs from 'fs';
import * as path from 'path';

describe('GPU Layer Styling Audit', () => {
  it('ensures MessageList status banners do not use heavy backdrop-blur-md', () => {
    const file = fs.readFileSync(path.resolve(__dirname, '../frontend/src/lib/components/chat/MessageList.svelte'), 'utf-8');
    // Banners inside MessageList should use solid/semi-translucent backgrounds instead of backdrop-blur-md
    expect(file).not.toMatch(/my-2\.5 p-3 rounded-xl border border-zinc-800 bg-zinc-900\/90 backdrop-blur-md/);
    expect(file).not.toMatch(/backdrop-blur-md/);
  });

  it('ensures FloatingPlanTracker does not use continuous backdrop-blur', () => {
    const file = fs.readFileSync(path.resolve(__dirname, '../frontend/src/lib/components/chat/FloatingPlanTracker.svelte'), 'utf-8');
    expect(file).not.toMatch(/backdrop-blur/);
  });

  it('ensures PlanReviewCard does not use continuous backdrop-blur', () => {
    const file = fs.readFileSync(path.resolve(__dirname, '../frontend/src/lib/components/chat/PlanReviewCard.svelte'), 'utf-8');
    expect(file).not.toMatch(/backdrop-blur/);
  });

  it('ensures terminal scrollback is bounded to 2000 lines', () => {
    const file = fs.readFileSync(path.resolve(__dirname, '../frontend/src/lib/components/terminal/TerminalPanel.svelte'), 'utf-8');
    expect(file).toMatch(/scrollback:\s*2000/);
  });
});
