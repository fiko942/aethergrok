export interface DiffStatResult {
  added: number;
  removed: number;
  oldPath?: string;
  newPath?: string;
  filePath?: string;
  diffUnified?: string;
}

export interface DiffParsedLine {
  type: 'add' | 'del' | 'normal';
  content: string;
}

export function parseUnifiedDiffStats(rawPatch: string): { added: number; removed: number } {
  let added = 0;
  let removed = 0;
  const lines = rawPatch.split('\n');
  for (const l of lines) {
    if (l.startsWith('+') && !l.startsWith('+++')) {
      added++;
    } else if (l.startsWith('-') && !l.startsWith('---')) {
      removed++;
    }
  }
  return { added, removed };
}

/**
 * Computes standard Myers/LCS line diff between oldText and newText
 */
export function computeStringDiff(oldText: string, newText: string): {
  addedCount: number;
  removedCount: number;
  lines: DiffParsedLine[];
} {
  const oldLines = oldText ? oldText.split('\n') : [];
  const newLines = newText ? newText.split('\n') : [];

  // Compute Longest Common Subsequence (LCS) matrix
  const m = oldLines.length;
  const n = newLines.length;
  const dp: number[][] = Array.from({ length: m + 1 }, () => new Array(n + 1).fill(0));

  for (let i = 0; i < m; i++) {
    for (let j = 0; j < n; j++) {
      if (oldLines[i] === newLines[j]) {
        dp[i + 1][j + 1] = dp[i][j] + 1;
      } else {
        dp[i + 1][j + 1] = Math.max(dp[i + 1][j], dp[i][j + 1]);
      }
    }
  }

  // Backtrack to build diff lines
  const lines: DiffParsedLine[] = [];
  let i = m;
  let j = n;
  let addedCount = 0;
  let removedCount = 0;

  while (i > 0 || j > 0) {
    if (i > 0 && j > 0 && oldLines[i - 1] === newLines[j - 1]) {
      lines.unshift({ type: 'normal', content: oldLines[i - 1] });
      i--;
      j--;
    } else if (j > 0 && (i === 0 || dp[i][j - 1] >= dp[i - 1][j])) {
      lines.unshift({ type: 'add', content: newLines[j - 1] });
      addedCount++;
      j--;
    } else if (i > 0 && (j === 0 || dp[i][j - 1] < dp[i - 1][j])) {
      lines.unshift({ type: 'del', content: oldLines[i - 1] });
      removedCount++;
      i--;
    }
  }

  return { addedCount, removedCount, lines };
}

export function calculateDiffStat(toolCall: {
  tool?: string;
  params?: Record<string, unknown> | string;
  diff?: {
    oldPath?: string;
    newPath?: string;
    addedCount?: number;
    removedCount?: number;
    diffUnified?: string;
  };
}): DiffStatResult {
  if (toolCall.diff) {
    let added = toolCall.diff.addedCount || 0;
    let removed = toolCall.diff.removedCount || 0;
    if (added === 0 && removed === 0 && toolCall.diff.diffUnified) {
      const stats = parseUnifiedDiffStats(toolCall.diff.diffUnified);
      added = stats.added;
      removed = stats.removed;
    }
    const path = toolCall.diff.newPath || toolCall.diff.oldPath;
    return {
      added,
      removed,
      oldPath: toolCall.diff.oldPath,
      newPath: toolCall.diff.newPath,
      filePath: path,
      diffUnified: toolCall.diff.diffUnified
    };
  }

  const p = (typeof toolCall.params === 'object' && toolCall.params !== null)
    ? (toolCall.params as Record<string, unknown>)
    : {};

  const toolName = (toolCall.tool || '').toLowerCase();
  const filePath = String(p.file_path || p.target_file || p.path || p.filePath || '');

  if (toolName.includes('search_replace') || toolName.includes('edit') || toolName.includes('str_replace')) {
    const oldStr = String(p.old_string || p.find || '');
    const newStr = String(p.new_string || p.replace || '');
    if (oldStr || newStr) {
      const { addedCount, removedCount } = computeStringDiff(oldStr, newStr);
      return {
        added: addedCount,
        removed: removedCount,
        oldPath: filePath,
        newPath: filePath,
        filePath
      };
    }
  }

  if (toolName.includes('write') || toolName.includes('save') || toolName.includes('create_file')) {
    const content = String(p.content || '');
    const lineCount = content ? content.split('\n').length : 0;
    return {
      added: lineCount,
      removed: 0,
      oldPath: filePath,
      newPath: filePath,
      filePath
    };
  }

  return { added: 0, removed: 0, filePath: filePath || undefined };
}
