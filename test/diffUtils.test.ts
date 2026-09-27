import { describe, it, expect } from 'vitest';
import { calculateDiffStat, computeStringDiff, parseUnifiedDiffStats } from '../frontend/src/lib/utils/diffUtils';

describe('diffUtils', () => {
  it('parses unified diff patch statistics accurately', () => {
    const patch = `--- a/file.ts
+++ b/file.ts
@@ -1,3 +1,4 @@
 line 1
-line 2
+line 2 mod
+line 3 new
 line 4`;
    const stats = parseUnifiedDiffStats(patch);
    expect(stats.removed).toBe(1);
    expect(stats.added).toBe(2);
  });

  it('calculates accurate line counts for multi-line search_replace', () => {
    const oldStr = 'line 1\nline 2\nline 3';
    const newStr = 'line 1\nline 2 modified\nline 3\nline 4\nline 5';
    const result = computeStringDiff(oldStr, newStr);

    expect(result.removedCount).toBe(1);
    expect(result.addedCount).toBe(3);
  });

  it('extracts diff stat from ToolCall with params old_string and new_string', () => {
    const tc = {
      id: 'tc1',
      tool: 'search_replace',
      status: 'completed' as const,
      params: {
        file_path: 'src/main.ts',
        old_string: 'const a = 1;\nconst b = 2;',
        new_string: 'const a = 10;\nconst b = 20;\nconst c = 30;'
      }
    };
    const stat = calculateDiffStat(tc);
    expect(stat.removed).toBe(2);
    expect(stat.added).toBe(3);
    expect(stat.filePath).toBe('src/main.ts');
  });

  it('extracts diff stat from write toolcall', () => {
    const tc = {
      id: 'tc2',
      tool: 'write',
      status: 'completed' as const,
      params: {
        file_path: 'src/newFile.ts',
        content: 'line 1\nline 2\nline 3\nline 4'
      }
    };
    const stat = calculateDiffStat(tc);
    expect(stat.removed).toBe(0);
    expect(stat.added).toBe(4);
    expect(stat.filePath).toBe('src/newFile.ts');
  });

  it('prefers explicit toolCall.diff when present', () => {
    const tc = {
      id: 'tc3',
      tool: 'edit',
      status: 'completed' as const,
      params: {},
      diff: {
        oldPath: 'src/old.ts',
        newPath: 'src/new.ts',
        addedCount: 5,
        removedCount: 2
      }
    };
    const stat = calculateDiffStat(tc);
    expect(stat.removed).toBe(2);
    expect(stat.added).toBe(5);
    expect(stat.filePath).toBe('src/new.ts');
  });
});
