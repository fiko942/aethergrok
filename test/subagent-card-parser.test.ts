import { describe, it, expect } from 'vitest';
import {
  isSubagentToolCall,
  extractSubagentMeta,
  cleanSubagentOutput,
  formatSubagentDuration,
  isSubagentBackgroundAck,
  extractSubagentIdFromText,
  isSubagentOutputPoller,
  extractTaskIdsFromParams,
  extractTaskPollerResults,
  isRedundantSubagentPoller
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

  it('detects CLI background subagent launch acknowledgments and extracts task ID', () => {
    const ackText = `Subagent started in background.
subagent_id: 01a15062-c047-7d91-a7e1-10e862e35c64
description: Subagent 1 Academic WS3 WS4
When you need its result, use get_command_or_subagent_output with task_ids=["01a15062-c047-7d91-a7e1-10e862e35c64"] and a positive timeout_ms.`;

    expect(isSubagentBackgroundAck(ackText)).toBe(true);
    expect(extractSubagentIdFromText(ackText)).toBe('01a15062-c047-7d91-a7e1-10e862e35c64');
    // cleanSubagentOutput should return empty string so the launch notice is not displayed as output
    expect(cleanSubagentOutput(ackText)).toBe('');

    const meta = extractSubagentMeta({ description: 'Subagent 1 Academic' }, ackText);
    expect(meta.taskId).toBe('01a15062-c047-7d91-a7e1-10e862e35c64');
  });

  it('preserves genuine subagent output when not a background launch notice', () => {
    const realResult = 'Completed analysis report:\n- Layout checks: Pass\n- Margin rules: Compliant';
    expect(isSubagentBackgroundAck(realResult)).toBe(false);
    expect(cleanSubagentOutput(realResult)).toBe(realResult);
  });

  describe('Subagent Poller Detection and Output Extraction', () => {
    it('identifies get_command_or_subagent_output and variants as pollers', () => {
      expect(isSubagentOutputPoller('get_command_or_subagent_output')).toBe(true);
      expect(isSubagentOutputPoller('Get task output: 2 tasks')).toBe(true);
      expect(isSubagentOutputPoller('Output 01a15062-c047-7d91-a7e1-10e862e35c64')).toBe(true);
      expect(isSubagentOutputPoller('command', { variant: 'TaskOutput' })).toBe(true);
      expect(isSubagentOutputPoller('custom_tool', { task_ids: ['t1', 't2'] })).toBe(true);
      expect(isSubagentOutputPoller('spawn_subagent')).toBe(false);
      expect(isSubagentOutputPoller('read_file')).toBe(false);
    });

    it('extracts task IDs from parameters correctly', () => {
      expect(extractTaskIdsFromParams({ task_ids: ['id-1', 'id-2'] })).toEqual(['id-1', 'id-2']);
      expect(extractTaskIdsFromParams({ task_id: 'id-solo' })).toEqual(['id-solo']);
      expect(extractTaskIdsFromParams(JSON.stringify({ task_ids: ['id-json'] }))).toEqual(['id-json']);
      expect(extractTaskIdsFromParams(null)).toEqual([]);
    });

    it('extracts structured results from JSON TaskOutput Result', () => {
      const jsonResult = JSON.stringify({
        type: 'TaskOutput',
        Result: {
          task_id: '01a15062-c047-7d91-a7e1-10e862e35c64',
          status: 'completed',
          duration_secs: 15.4,
          output: 'Analysis report for WS3 and WS4 completed successfully.'
        }
      });
      const results = extractTaskPollerResults('get_command_or_subagent_output', { task_ids: ['01a15062-c047-7d91-a7e1-10e862e35c64'] }, jsonResult);
      expect(results).toHaveLength(1);
      expect(results[0].taskId).toBe('01a15062-c047-7d91-a7e1-10e862e35c64');
      expect(results[0].status).toBe('completed');
      expect(results[0].durationMs).toBe(15400);
      expect(results[0].output).toBe('Analysis report for WS3 and WS4 completed successfully.');
    });

    it('extracts multi-task results from MultiResult JSON', () => {
      const multiJson = JSON.stringify({
        type: 'TaskOutput',
        MultiResult: {
          mode: 'wait_all',
          results: [
            {
              task_id: 'sub-1',
              status: 'completed',
              duration_secs: 10.0,
              output: 'Report 1 done.'
            },
            {
              task_id: 'sub-2',
              status: 'running',
              output: 'Wait returned early because another finished; this task is still running.'
            }
          ]
        }
      });
      const results = extractTaskPollerResults('get_command_or_subagent_output', {}, multiJson);
      expect(results).toHaveLength(2);
      expect(results[0].taskId).toBe('sub-1');
      expect(results[0].status).toBe('completed');
      expect(results[0].output).toBe('Report 1 done.');

      expect(results[1].taskId).toBe('sub-2');
      expect(results[1].status).toBe('running');
    });

    it('extracts results from text block divider format', () => {
      const textBlock = `=== Task 01a15062-dd04-7a61-85eb-ed3c0f16c040 ===
Command: [subagent:general-purpose] Subagent 2 Empirical WS5 WS6
Status: completed
Duration: 22.5s

=== Output ===
Empirical analysis of Worksheet 5 and 6 is complete. All 12 criteria verified.`;

      const results = extractTaskPollerResults('get_command_or_subagent_output', {}, textBlock);
      expect(results).toHaveLength(1);
      expect(results[0].taskId).toBe('01a15062-dd04-7a61-85eb-ed3c0f16c040');
      expect(results[0].status).toBe('completed');
      expect(results[0].durationMs).toBe(22500);
      expect(results[0].output).toBe('Empirical analysis of Worksheet 5 and 6 is complete. All 12 criteria verified.');
    });

    it('identifies redundant subagent pollers to filter out of the main action list', () => {
      const subagentCall = {
        id: 'call-1',
        tool: 'spawn_subagent',
        params: { description: 'Subagent 1' },
        result: 'Subagent started in background.\nsubagent_id: 01a15062-c047-7d91-a7e1-10e862e35c64'
      };

      const matchingPoller = {
        id: 'call-2',
        tool: 'get_command_or_subagent_output',
        params: { task_ids: ['01a15062-c047-7d91-a7e1-10e862e35c64'] }
      };

      const normalTool = {
        id: 'call-3',
        tool: 'read_file',
        params: { target_file: 'README.md' }
      };

      expect(isRedundantSubagentPoller(matchingPoller, [subagentCall, matchingPoller, normalTool])).toBe(true);
      expect(isRedundantSubagentPoller(normalTool, [subagentCall, matchingPoller, normalTool])).toBe(false);
      expect(isRedundantSubagentPoller(subagentCall, [subagentCall, matchingPoller, normalTool])).toBe(false);
    });
  });
});
