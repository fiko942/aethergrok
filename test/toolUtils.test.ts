import { describe, it, expect } from 'vitest';
import { cleanANSI, formatToolResult, formatActionSummary, extractSubagentIdFromText, extractSubagentMeta } from '../frontend/src/lib/utils/toolUtils';

describe('toolUtils', () => {
  it('cleanANSI removes escape codes', () => {
    const raw = '\u001b[32m✓\u001b[39m Lockfile passes \u001b[2m$ vitest run\u001b[22m';
    expect(cleanANSI(raw)).toBe('✓ Lockfile passes $ vitest run');
  });

  it('formatToolResult handles standard string', () => {
    const res = formatToolResult('hello world');
    expect(res.output).toBe('hello world');
    expect(res.isJsonTaskOutput).toBe(false);
  });

  it('formatToolResult unwraps TaskOutput with Result object and cleans ANSI codes', () => {
    const raw = JSON.stringify({
      Result: {
        command: 'pnpm test',
        duration_secs: 27.015507,
        ended: '2026-09-27T04:56:36Z',
        exit_code: 1,
        output: '\u001b[32m✓\u001b[39m Lockfile passes supply-chain policies\n(node:32498) Warning: The NO_COLOR env is ignored',
        output_file: '/tmp/test.log',
        status: 'failed',
        truncated: true
      },
      type: 'TaskOutput'
    });

    const parsed = formatToolResult(raw);
    expect(parsed.isJsonTaskOutput).toBe(true);
    expect(parsed.command).toBe('pnpm test');
    expect(parsed.exitCode).toBe(1);
    expect(parsed.durationSecs).toBeCloseTo(27.015, 2);
    expect(parsed.status).toBe('failed');
    expect(parsed.truncated).toBe(true);
    expect(parsed.outputFile).toBe('/tmp/test.log');
    expect(parsed.output).toContain('✓ Lockfile passes supply-chain policies');
    expect(parsed.output).not.toContain('\u001b[32m');
  });

  it('correctly derives description over raw command for terminal tool calls', () => {
    function parseTerminalCall(params: { command?: string; description?: string }) {
      const cmd = String(params.command || '');
      const desc = String(params.description || '').trim();
      return {
        verb: 'Run',
        target: desc || (cmd ? `$ ${cmd}` : 'command'),
        cmd,
        description: desc
      };
    }

    const withDesc = parseTerminalCall({
      command: 'find docs -name "*.md" | sort',
      description: 'extract task brief for Task 2'
    });
    expect(withDesc.verb).toBe('Run');
    expect(withDesc.target).toBe('extract task brief for Task 2');
    expect(withDesc.cmd).toBe('find docs -name "*.md" | sort');
    expect(withDesc.description).toBe('extract task brief for Task 2');

    const withoutDesc = parseTerminalCall({
      command: 'git status'
    });
    expect(withoutDesc.verb).toBe('Run');
    expect(withoutDesc.target).toBe('$ git status');
    expect(withoutDesc.cmd).toBe('git status');
    expect(withoutDesc.description).toBe('');
  });

  describe('formatActionSummary', () => {
    it('formats read_file actions with target and line range', () => {
      const summary = formatActionSummary('read_file', {
        target_file: 'src/main.ts',
        offset: 10,
        limit: 50
      });
      expect(summary.verb).toBe('Read');
      expect(summary.target).toBe('src/main.ts (10-59)');
      expect(summary.iconType).toBe('read');
    });

    it('formats search_replace as Edit action', () => {
      const summary = formatActionSummary('search_replace', {
        file_path: 'pkg/runner.go'
      });
      expect(summary.verb).toBe('Edit');
      expect(summary.target).toBe('pkg/runner.go');
      expect(summary.iconType).toBe('edit');
    });

    it('formats write as Write action', () => {
      const summary = formatActionSummary('write', {
        file_path: 'README.md'
      });
      expect(summary.verb).toBe('Write');
      expect(summary.target).toBe('README.md');
      expect(summary.iconType).toBe('write');
    });

    it('formats grep and search actions', () => {
      const summary = formatActionSummary('grep', {
        pattern: 'func Start',
        path: 'pkg'
      });
      expect(summary.verb).toBe('Search');
      expect(summary.target).toBe('"func Start" in pkg');
      expect(summary.iconType).toBe('search');
    });

    it('formats list_dir as List action', () => {
      const summary = formatActionSummary('list_dir', {
        target_directory: 'frontend/src'
      });
      expect(summary.verb).toBe('List');
      expect(summary.target).toBe('frontend/src');
      expect(summary.iconType).toBe('list');
    });

    it('formats run_terminal_command with description or command', () => {
      const withDesc = formatActionSummary('run_terminal_command', {
        command: 'go build -o app',
        description: 'Compile application binary'
      });
      expect(withDesc.verb).toBe('Run');
      expect(withDesc.target).toBe('Compile application binary');
      expect(withDesc.iconType).toBe('terminal');

      const withoutDesc = formatActionSummary('run_terminal_command', {
        command: 'npm run check'
      });
      expect(withoutDesc.verb).toBe('Run');
      expect(withoutDesc.target).toBe('npm run check');
      expect(withoutDesc.iconType).toBe('terminal');
    });
  });

  describe('extractSubagentIdFromText', () => {
    it('extracts bare subagent_id from acknowledgment text', () => {
      const text = 'Subagent started in background.\nsubagent_id: 01a15226-fcb4-7973-a534-2dc31fbb1a8b\ndescription: Analyze docs';
      expect(extractSubagentIdFromText(text)).toBe('01a15226-fcb4-7973-a534-2dc31fbb1a8b');
    });

    it('extracts backtick quoted subagent_id', () => {
      const text = 'Subagent started in background.\nsubagent_id: `01a15226-fcb4-7973-a534-2dc31fbb1a8b`';
      expect(extractSubagentIdFromText(text)).toBe('01a15226-fcb4-7973-a534-2dc31fbb1a8b');
    });

    it('extracts double and single quoted subagent_id', () => {
      expect(extractSubagentIdFromText('subagent_id: "01a15226-fcb4-7973"')).toBe('01a15226-fcb4-7973');
      expect(extractSubagentIdFromText("subagent_id: '01a15226-fcb4-7973'")).toBe('01a15226-fcb4-7973');
    });

    it('extracts case-insensitively and with varying spaces', () => {
      expect(extractSubagentIdFromText('Subagent_ID:   01a15226-fcb4')).toBe('01a15226-fcb4');
    });

    it('returns undefined when no subagent_id is present', () => {
      expect(extractSubagentIdFromText('Some random output without id')).toBeUndefined();
    });
  });
});
