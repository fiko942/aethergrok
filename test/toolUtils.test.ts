import { describe, it, expect } from 'vitest';
import { cleanANSI, formatToolResult } from '../frontend/src/lib/utils/toolUtils';

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
});
