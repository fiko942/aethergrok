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

  it('determines whether task actions container should auto-scroll down', () => {
    function computeShouldAutoScroll(params: {
      scrollTop: number;
      scrollHeight: number;
      clientHeight: number;
      userHasScrolled: boolean;
      isRunning: boolean;
      isNewItem: boolean;
    }): boolean {
      const distanceFromBottom = params.scrollHeight - (params.scrollTop + params.clientHeight);
      const isNearBottom = distanceFromBottom <= 35;
      if (!params.userHasScrolled || isNearBottom || params.isRunning) {
        return true;
      }
      return false;
    }

    // When near bottom (distance <= 35px), should auto-scroll
    expect(
      computeShouldAutoScroll({
        scrollTop: 980,
        scrollHeight: 1200,
        clientHeight: 200,
        userHasScrolled: false,
        isRunning: false,
        isNewItem: true
      })
    ).toBe(true);

    // When user scrolled up (distance = 400px) and subagent is finished, do not yank scroll
    expect(
      computeShouldAutoScroll({
        scrollTop: 600,
        scrollHeight: 1200,
        clientHeight: 200,
        userHasScrolled: true,
        isRunning: false,
        isNewItem: false
      })
    ).toBe(false);

    // When subagent is actively running and emits a new item, always auto-scroll to latest
    expect(
      computeShouldAutoScroll({
        scrollTop: 500,
        scrollHeight: 1200,
        clientHeight: 200,
        userHasScrolled: true,
        isRunning: true,
        isNewItem: true
      })
    ).toBe(true);
  });

  it('calculates live subagent duration against now while running, and freezes at endTime when stopped', () => {
    function calculateSubagentDuration(params: {
      startTime: number;
      endTime?: number;
      isRunning: boolean;
      now: number;
      childCalls?: Array<{ startTime?: number; endTime?: number }>;
    }): string {
      let end: number;
      if (params.isRunning) {
        end = params.now;
      } else {
        const lastChildEnd = (params.childCalls && params.childCalls.length > 0)
          ? Math.max(0, ...params.childCalls.map(c => c.endTime || c.startTime || 0))
          : 0;
        end = Math.max(params.endTime || 0, lastChildEnd) || params.now;
      }
      return formatSubagentDuration(Math.max(0, end - params.startTime));
    }

    const startTime = 1000000;
    // Initial spawn completed after 1s, giving a premature endTime of 1001000
    const prematureEndTime = 1001000;

    // While running: even with premature endTime present, duration must use `now` and advance
    expect(calculateSubagentDuration({
      startTime,
      endTime: prematureEndTime,
      isRunning: true,
      now: startTime + 5000 // 5 seconds in
    })).toBe('5s');

    expect(calculateSubagentDuration({
      startTime,
      endTime: prematureEndTime,
      isRunning: true,
      now: startTime + 356000 // 5m 56s in
    })).toBe('5m 56s');

    // When stopped/completed: duration freezes at the finalized endTime or child end time
    expect(calculateSubagentDuration({
      startTime,
      endTime: startTime + 356000,
      isRunning: false,
      now: startTime + 999999 // now continues moving forward
    })).toBe('5m 56s');

    // When stopped without explicit endTime, falls back to latest child tool call timestamp
    expect(calculateSubagentDuration({
      startTime,
      endTime: prematureEndTime,
      isRunning: false,
      now: startTime + 999999,
      childCalls: [
        { startTime: startTime + 1000, endTime: startTime + 120000 },
        { startTime: startTime + 121000, endTime: startTime + 180000 }
      ]
    })).toBe('3m 0s');
  });
});
