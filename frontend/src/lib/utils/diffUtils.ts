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
  oldLineNo?: number;
  newLineNo?: number;
}

export interface ToolCallFileChange {
  toolCallId: string;
  path: string;
  addedCount: number;
  removedCount: number;
  oldContent?: string;
  newContent?: string;
  diffUnified?: string;
  diff?: {
    oldPath?: string;
    newPath?: string;
    oldContent?: string;
    newContent?: string;
    diffUnified?: string;
    addedCount?: number;
    removedCount?: number;
  };
}

export function parseToolCallParams(params?: Record<string, unknown> | string): Record<string, unknown> {
  if (typeof params === 'object' && params !== null) {
    return params as Record<string, unknown>;
  }
  if (typeof params === 'string' && params.trim().startsWith('{')) {
    try {
      const parsed = JSON.parse(params);
      if (typeof parsed === 'object' && parsed !== null) {
        return parsed as Record<string, unknown>;
      }
    } catch {
      // ignore JSON parse error
    }
  }
  return {};
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
 * Computes standard Myers/LCS line diff between oldText and newText with line numbers
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

  // Assign sequential line numbers
  let oldLine = 1;
  let newLine = 1;
  for (const line of lines) {
    if (line.type === 'normal') {
      line.oldLineNo = oldLine++;
      line.newLineNo = newLine++;
    } else if (line.type === 'del') {
      line.oldLineNo = oldLine++;
    } else if (line.type === 'add') {
      line.newLineNo = newLine++;
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

  const p = parseToolCallParams(toolCall.params);
  const toolName = (toolCall.tool || '').toLowerCase();
  const filePath = String(p.file_path || p.target_file || p.path || p.filePath || p.file || '');

  const oldStr = typeof p.old_string === 'string' ? p.old_string :
    typeof p.oldStr === 'string' ? p.oldStr :
    typeof p.find === 'string' ? p.find :
    typeof p.search === 'string' ? p.search :
    typeof p.oldText === 'string' ? p.oldText :
    typeof p.old_content === 'string' ? p.old_content : '';

  const newStr = typeof p.new_string === 'string' ? p.new_string :
    typeof p.newStr === 'string' ? p.newStr :
    typeof p.replace === 'string' ? p.replace :
    typeof p.newText === 'string' ? p.newText :
    typeof p.new_content === 'string' ? p.new_content : '';

  if (oldStr || newStr || toolName.includes('search_replace') || toolName.includes('edit') || toolName.includes('str_replace')) {
    if (oldStr || newStr) {
      const { addedCount, removedCount } = computeStringDiff(oldStr, newStr);
      return {
        added: addedCount,
        removed: removedCount,
        oldPath: filePath,
        newPath: filePath,
        filePath: filePath || undefined
      };
    }
  }

  if (toolName.includes('write') || toolName.includes('save') || toolName.includes('create_file') || p.content !== undefined || p.contents !== undefined) {
    const content = String(p.content ?? p.contents ?? p.text ?? '');
    const lineCount = content ? content.split('\n').length : 0;
    return {
      added: lineCount,
      removed: 0,
      oldPath: filePath,
      newPath: filePath,
      filePath: filePath || undefined
    };
  }

  return { added: 0, removed: 0, filePath: filePath || undefined };
}

export function extractToolCallFileChange(toolCall: {
  id?: string;
  tool?: string;
  params?: Record<string, unknown> | string;
  diff?: {
    oldPath?: string;
    newPath?: string;
    oldContent?: string;
    newContent?: string;
    diffUnified?: string;
    addedCount?: number;
    removedCount?: number;
  };
}): ToolCallFileChange {
  const stat = calculateDiffStat(toolCall);
  const p = parseToolCallParams(toolCall.params);
  const toolName = (toolCall.tool || '').toLowerCase();

  const path = stat.filePath ||
    (toolCall.diff && (toolCall.diff.newPath || toolCall.diff.oldPath)) ||
    String(p.file_path || p.target_file || p.path || p.filePath || p.file || '') ||
    'modified_file';

  const oldContent = toolCall.diff?.oldContent ??
    (typeof p.old_string === 'string' ? p.old_string :
     typeof p.oldStr === 'string' ? p.oldStr :
     typeof p.find === 'string' ? p.find :
     typeof p.search === 'string' ? p.search :
     typeof p.oldText === 'string' ? p.oldText :
     typeof p.old_content === 'string' ? p.old_content :
     (toolName.includes('write') || toolName.includes('save') || toolName.includes('create_file') ? '' : undefined));

  const newContent = toolCall.diff?.newContent ??
    (typeof p.new_string === 'string' ? p.new_string :
     typeof p.newStr === 'string' ? p.newStr :
     typeof p.replace === 'string' ? p.replace :
     typeof p.newText === 'string' ? p.newText :
     typeof p.new_content === 'string' ? p.new_content :
     typeof p.content === 'string' ? p.content :
     typeof p.contents === 'string' ? p.contents :
     typeof p.text === 'string' ? p.text : undefined);

  return {
    toolCallId: toolCall.id || '',
    path,
    addedCount: stat.added,
    removedCount: stat.removed,
    oldContent,
    newContent,
    diffUnified: stat.diffUnified || toolCall.diff?.diffUnified,
    diff: toolCall.diff
  };
}
