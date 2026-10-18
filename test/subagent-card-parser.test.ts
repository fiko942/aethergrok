import { describe, it, expect } from 'vitest';
import {
  isSubagentToolCall,
  extractSubagentMeta,
  cleanSubagentOutput,
  formatSubagentDuration
} from '../frontend/src/lib/utils/toolUtils';

describe('Subagent Tool Classifier & Output Parsing', () => {
  it('correctly identifies spawn_subagent and subagent tool names', () => {
    expect(isSubagentToolCall('spawn_subagent')).toBe(true);
    expect(isSubagentToolCall('subagent')).toBe(true);
    expect(isSubagentToolCall('delegate_task')).toBe(true);
    expect(isSubagentToolCall('run_terminal_command', { is_background: true, description: 'Spawn background subagent' })).toBe(true);
    expect(isSubagentToolCall('run_terminal_command', { command: 'ls -la' })).toBe(false);
    expect(isSubagentToolCall('read_file')).toBe(false);
  });

  it('extracts human task description and prompt cleanly', () => {
    const params = {
      description: 'Investigate slow database queries',
      prompt: 'Check slow query logs in /var/log/postgres and report top 3 bottlenecks',
      subagent_type: 'general-purpose',
      model: 'grok-beta',
      background: true
    };
    const meta = extractSubagentMeta(params);
    expect(meta.description).toBe('Investigate slow database queries');
    expect(meta.prompt).toBe('Check slow query logs in /var/log/postgres and report top 3 bottlenecks');
    expect(meta.subagentType).toBe('general-purpose');
    expect(meta.model).toBe('grok-beta');
    expect(meta.isBackground).toBe(true);
  });

  it('falls back to the first line of prompt if description is missing', () => {
    const params = {
      prompt: 'Refactor authentication middleware\nEnsure JWT validation passes tests'
    };
    const meta = extractSubagentMeta(params);
    expect(meta.description).toBe('Refactor authentication middleware');
    expect(meta.subagentType).toBe('general-purpose');
  });

  it('cleans internal XML tags and CLI boilerplate from subagent output', () => {
    const raw = `<subagent_meta>{"id":"sub-1"}</subagent_meta>
This is the output of the subagent:
<response>
Found 2 memory leaks in connection pool.
</response>
Agent ID: sub-1 (resume from here)`;
    const cleaned = cleanSubagentOutput(raw);
    expect(cleaned).toBe('Found 2 memory leaks in connection pool.');
  });

  it('formats running and completed durations cleanly matching Grok CLI rules', () => {
    expect(formatSubagentDuration(500)).toBe('1s');
    expect(formatSubagentDuration(42000)).toBe('42s');
    expect(formatSubagentDuration(75000)).toBe('1m 15s');
    expect(formatSubagentDuration(3665000)).toBe('1h 1m 5s');
  });
});
