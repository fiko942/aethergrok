import { describe, it, expect } from 'vitest';
import { extractToolCallFileChange, computeStringDiff, calculateDiffStat } from '../frontend/src/lib/utils/diffUtils';
import type { ToolCall } from '../frontend/src/lib/stores/session.svelte';

describe('TurnDiffSummary and DiffCard diff extraction & computation', () => {
  it('correctly extracts file changes and computes diff for search_replace tool call', () => {
    const shortcutDetectorOld = `export function detectShortcut(e: KeyboardEvent) {
  if (e.key === 'k' && e.ctrlKey) {
    return 'search';
  }
  return null;
}`;

    const shortcutDetectorNew = `export function detectShortcut(e: KeyboardEvent, isMac: boolean) {
  const mod = isMac ? e.metaKey : e.ctrlKey;
  if (e.key.toLowerCase() === 'k' && mod) {
    return 'search';
  }
  if (e.key.toLowerCase() === 'b' && mod) {
    return 'sidebar';
  }
  if (e.key.toLowerCase() === 'j' && mod) {
    return 'terminal';
  }
  if (e.key.toLowerCase() === 'p' && mod && e.shiftKey) {
    return 'command_palette';
  }
  return null;
}`;

    const toolCall: ToolCall = {
      id: 'call_665367',
      tool: 'search_replace',
      status: 'completed',
      params: {
        file_path: 'frontend/src/lib/utils/shortcutDetector.ts',
        old_string: shortcutDetectorOld,
        new_string: shortcutDetectorNew
      },
      startTime: 1000,
      endTime: 1050
    };

    // 1. Stat calculation
    const stat = calculateDiffStat(toolCall);
    expect(stat.filePath).toBe('frontend/src/lib/utils/shortcutDetector.ts');
    expect(stat.removed).toBe(2);
    expect(stat.added).toBe(12);

    // 2. File change extraction
    const change = extractToolCallFileChange(toolCall);
    expect(change.toolCallId).toBe('call_665367');
    expect(change.path).toBe('frontend/src/lib/utils/shortcutDetector.ts');
    expect(change.oldContent).toBe(shortcutDetectorOld);
    expect(change.newContent).toBe(shortcutDetectorNew);
    expect(change.addedCount).toBe(12);
    expect(change.removedCount).toBe(2);

    // 3. DiffCard computation logic
    const diff = computeStringDiff(change.oldContent || '', change.newContent || '');
    expect(diff.lines.length).toBeGreaterThan(0);
    expect(diff.lines.some(l => l.type === 'del' && l.content.includes("e.key === 'k'"))).toBe(true);
    expect(diff.lines.some(l => l.type === 'add' && l.content.includes("const mod = isMac"))).toBe(true);
    expect(diff.lines.some(l => l.type === 'normal' && l.content.includes("return null;"))).toBe(true);
  });

  it('correctly handles multi-tool turn with both write and search_replace', () => {
    const inputShieldContent = Array.from({ length: 94 }, (_, i) => `// inputShield line ${i + 1}`).join('\n');
    const toolCalls: ToolCall[] = [
      {
        id: 'call_write_1',
        tool: 'write',
        status: 'completed',
        params: {
          file_path: 'C:\\Users\\Administrator\\Desktop\\project pribadi\\aethergrok\\frontend\\src\\lib\\stores\\inputShield.svelte.ts',
          content: inputShieldContent
        }
      },
      {
        id: 'call_665367',
        tool: 'search_replace',
        status: 'completed',
        params: {
          file_path: 'frontend/src/lib/utils/shortcutDetector.ts',
          old_string: 'line A\nline B\nline C\nline D\nline E',
          new_string: Array.from({ length: 16 }, (_, i) => `replacement line ${i + 1}`).join('\n')
        }
      }
    ];

    const changes = toolCalls.map(tc => extractToolCallFileChange(tc));
    expect(changes).toHaveLength(2);

    // First file: write
    expect(changes[0].path).toBe('C:\\Users\\Administrator\\Desktop\\project pribadi\\aethergrok\\frontend\\src\\lib\\stores\\inputShield.svelte.ts');
    expect(changes[0].addedCount).toBe(94);
    expect(changes[0].removedCount).toBe(0);
    expect(changes[0].newContent).toBe(inputShieldContent);

    // Second file: search_replace (exact case from user screenshot)
    expect(changes[1].path).toBe('frontend/src/lib/utils/shortcutDetector.ts');
    expect(changes[1].addedCount).toBe(16);
    expect(changes[1].removedCount).toBe(5);
    expect(changes[1].oldContent).toBeDefined();
    expect(changes[1].newContent).toBeDefined();

    // Sum totals match +110 -5 in user screenshot
    const totalAdded = changes.reduce((sum, c) => sum + c.addedCount, 0);
    const totalRemoved = changes.reduce((sum, c) => sum + c.removedCount, 0);
    expect(totalAdded).toBe(110);
    expect(totalRemoved).toBe(5);
  });
});
