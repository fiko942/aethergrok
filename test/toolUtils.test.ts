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
});
