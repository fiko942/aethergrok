import { describe, it, expect } from 'vitest';
import {
  isSubagentToolCall,
  extractSubagentMeta,
  formatSubagentDuration
} from '../frontend/src/lib/utils/toolUtils';

describe('ToolCallCard Subagent Variant Derivation', () => {
  it('correctly derives subagent card attributes for spawn_subagent tool call', () => {
    const rawTool = 'spawn_subagent';
    const params = {
      description: 'Find memory leaks in cache',
      prompt: 'Inspect heap snapshot',
      subagent_type: 'explore'
    };

    expect(isSubagentToolCall(rawTool, params)).toBe(true);
    const meta = extractSubagentMeta(params);

    expect(meta.description).toBe('Find memory leaks in cache');
    expect(meta.subagentType).toBe('explore');
  });

  it('correctly derives subagent card attributes for delegate_task with prompt fallback', () => {
    const rawTool = 'delegate_task';
    const params = {
      prompt: 'Audit package vulnerabilities and fix audit warnings'
    };

    expect(isSubagentToolCall(rawTool, params)).toBe(true);
    const meta = extractSubagentMeta(params);

    expect(meta.description).toBe('Audit package vulnerabilities and fix audit warnings');
    expect(meta.subagentType).toBe('general-purpose');
  });

  it('correctly identifies background run_terminal_command subagent call', () => {
    const rawTool = 'run_terminal_command';
    const params = {
      command: 'grok-subagent --task "refactor components"',
      is_background: true,
      description: 'Spawn background subagent'
    };

    expect(isSubagentToolCall(rawTool, params)).toBe(true);
    const meta = extractSubagentMeta(params);
    expect(meta.description).toBe('Spawn background subagent');
    expect(meta.isBackground).toBe(true);
  });

  it('does not classify standard tool calls as subagents', () => {
    expect(isSubagentToolCall('read_file', { target_file: 'src/main.ts' })).toBe(false);
    expect(isSubagentToolCall('search_replace', { file_path: 'src/app.ts' })).toBe(false);
    expect(isSubagentToolCall('write', { file_path: 'src/index.ts' })).toBe(false);
    expect(isSubagentToolCall('run_terminal_command', { command: 'cargo test' })).toBe(false);
    expect(isSubagentToolCall('list_dir', { target_directory: '.' })).toBe(false);
  });

  it('formats running duration using now and completed duration using toolCall.endTime', () => {
    const startTime = 10000;
    const now = 25000;
    const runningElapsed = now - startTime;
    expect(formatSubagentDuration(runningElapsed)).toBe('15s');

    const endTime = 75000;
    const completedElapsed = endTime - startTime;
    expect(formatSubagentDuration(completedElapsed)).toBe('1m 5s');
  });
});
