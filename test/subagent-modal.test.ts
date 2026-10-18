import { describe, it, expect } from 'vitest';
import {
  extractSubagentMeta,
  cleanSubagentOutput,
  formatSubagentDuration
} from '../frontend/src/lib/utils/toolUtils';

describe('Subagent Detail Modal Data Flow', () => {
  it('prepares complete modal state from toolCall parameters and result', () => {
    const toolCall = {
      id: 'tc-sub-1',
      tool: 'spawn_subagent',
      status: 'completed' as const,
      startTime: 1000,
      endTime: 85000,
      params: {
        description: 'Profile frontend bundle sizes',
        prompt: 'Run vite build and analyze chunk size distribution',
        subagent_type: 'performance-audit',
        model: 'grok-beta',
        isolation: 'none'
      },
      result: 'Bundle analysis complete. Total vendor size: 412kB.'
    };

    const meta = extractSubagentMeta(toolCall.params);
    const cleanedOutput = cleanSubagentOutput(toolCall.result);

    expect(meta.description).toBe('Profile frontend bundle sizes');
    expect(meta.prompt).toBe('Run vite build and analyze chunk size distribution');
    expect(meta.subagentType).toBe('performance-audit');
    expect(cleanedOutput).toBe('Bundle analysis complete. Total vendor size: 412kB.');
  });

  it('computes and formats active and finished elapsed durations correctly', () => {
    const startTime = 10000;
    const endTime = 75000;
    const duration = formatSubagentDuration(endTime - startTime);
    expect(duration).toBe('1m 5s');

    const runningDuration = formatSubagentDuration(5000);
    expect(runningDuration).toBe('5s');
  });

  it('extracts fallback description and handles raw parameters formatting', () => {
    const toolCall = {
      id: 'tc-sub-2',
      tool: 'subagent',
      status: 'running' as const,
      params: {
        prompt: 'Inspect memory leaks in websocket connections\nSecond line'
      }
    };

    const meta = extractSubagentMeta(toolCall.params);
    expect(meta.description).toBe('Inspect memory leaks in websocket connections');
    expect(meta.subagentType).toBe('general-purpose');
    expect(meta.isBackground).toBe(false);
  });
});
