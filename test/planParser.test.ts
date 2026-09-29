import { describe, it, expect } from 'vitest';
import {
  parseTodosUpdated,
  normalizeTodoStatus,
  calculatePlanMetrics,
  cleanExtractedPath,
  extractPlanFilePath,
  extractPlanMarkdownFromResult
} from '../frontend/src/lib/utils/planParser';

describe('planParser', () => {
  it('normalizes various status representations', () => {
    expect(normalizeTodoStatus('completed')).toBe('completed');
    expect(normalizeTodoStatus('done')).toBe('completed');
    expect(normalizeTodoStatus('finished')).toBe('completed');
    expect(normalizeTodoStatus('in_progress')).toBe('in_progress');
    expect(normalizeTodoStatus('in-progress')).toBe('in_progress');
    expect(normalizeTodoStatus('running')).toBe('in_progress');
    expect(normalizeTodoStatus('cancelled')).toBe('cancelled');
    expect(normalizeTodoStatus('pending')).toBe('pending');
    expect(normalizeTodoStatus('unknown')).toBe('pending');
    expect(normalizeTodoStatus(undefined)).toBe('pending');
  });

  it('parses TodosUpdated nested dictionary payload (like in screenshot)', () => {
    const rawJson = JSON.stringify({
      TodosUpdated: {
        state: {
          todos: {
            'create-input-shield-store': {
              content: 'Create centralized inputShieldStore for managing temporary read-only state',
              priority: 'medium',
              status: 'completed'
            },
            'test-and-verify': {
              content: 'Add and run comprehensive unit tests for inputShieldStore',
              priority: 'medium',
              status: 'pending'
            }
          }
        }
      }
    });

    const parsed = parseTodosUpdated(rawJson);
    expect(parsed).not.toBeNull();
    expect(parsed?.length).toBe(2);

    expect(parsed?.[0].id).toBe('create-input-shield-store');
    expect(parsed?.[0].content).toContain('Create centralized inputShieldStore');
    expect(parsed?.[0].status).toBe('completed');

    expect(parsed?.[1].id).toBe('test-and-verify');
    expect(parsed?.[1].status).toBe('pending');
  });

  it('parses flat array todos format', () => {
    const raw = {
      todos: [
        { id: 'step-1', content: 'Design architecture', status: 'completed' },
        { id: 'step-2', content: 'Implement UI components', status: 'in_progress' },
        { id: 'step-3', content: 'Write tests', status: 'pending' }
      ]
    };

    const parsed = parseTodosUpdated(raw);
    expect(parsed).toHaveLength(3);
    expect(parsed?.[1].status).toBe('in_progress');
  });

  it('calculates plan metrics accurately', () => {
    const todos = [
      { id: '1', content: 'A', status: 'completed' as const },
      { id: '2', content: 'B', status: 'completed' as const },
      { id: '3', content: 'C', status: 'in_progress' as const },
      { id: '4', content: 'D', status: 'pending' as const }
    ];

    const metrics = calculatePlanMetrics(todos, 'tool-1', 'write');
    expect(metrics.totalCount).toBe(4);
    expect(metrics.completedCount).toBe(2);
    expect(metrics.inProgressCount).toBe(1);
    expect(metrics.pendingCount).toBe(1);
    expect(metrics.progressPercent).toBe(50);
    expect(metrics.sourceToolId).toBe('tool-1');
  });

  it('returns null for non-todo objects', () => {
    expect(parseTodosUpdated('hello world')).toBeNull();
    expect(parseTodosUpdated(JSON.stringify({ diff: 'some text' }))).toBeNull();
    expect(parseTodosUpdated(null)).toBeNull();
  });

  describe('extractPlanFilePath & cleanExtractedPath', () => {
    it('cleans extracted path from sentence markers and trailing punctuation', () => {
      const raw = 'C:\\Users\\Admin\\.grok\\sessions\\xyz\\plan.md. The file exists and is empty.';
      expect(cleanExtractedPath(raw)).toBe('C:\\Users\\Admin\\.grok\\sessions\\xyz\\plan.md');

      const quoted = '"C:\\Users\\Admin\\.grok\\sessions\\xyz\\plan.md."';
      expect(cleanExtractedPath(quoted)).toBe('C:\\Users\\Admin\\.grok\\sessions\\xyz\\plan.md');
    });

    it('extracts plan path from enter_plan_mode text content', () => {
      const enterResult = `You have entered plan mode. You should now focus on exploring the codebase and creating an implementation plan.

Write your plan to C:\\Users\\Administrator\\.grok\\sessions\\C%3A%5CUsers%5CAdministrator%5CDesktop\\01a0ede8-a17d-73f3-aee0-b11380e75870\\plan.md. The file exists and is empty.

In plan mode, you should:
1. Thoroughly explore the codebase`;

      const path = extractPlanFilePath({}, enterResult);
      expect(path).toBe('C:\\Users\\Administrator\\.grok\\sessions\\C%3A%5CUsers%5CAdministrator%5CDesktop\\01a0ede8-a17d-73f3-aee0-b11380e75870\\plan.md');
    });

    it('extracts plan path from exit_plan_mode text content', () => {
      const exitResult = `Your plan has been approved. You can now start coding.

Your plan has been saved at: C:\\Users\\Administrator\\.grok\\sessions\\proj\\plan.md

## Plan:
# Implementation Plan`;

      const path = extractPlanFilePath({}, exitResult);
      expect(path).toBe('C:\\Users\\Administrator\\.grok\\sessions\\proj\\plan.md');
    });

    it('extracts plan path from params object or structured result', () => {
      expect(extractPlanFilePath({ plan_file_path: '/path/to/plan.md' }, null)).toBe('/path/to/plan.md');
      expect(extractPlanFilePath({}, { Entered: { plan_file_path: '/path/to/plan.md' } })).toBe('/path/to/plan.md');
      expect(extractPlanFilePath({}, 'Plan file: /tmp/plan.md')).toBe('/tmp/plan.md');
    });
  });

  describe('extractPlanMarkdownFromResult', () => {
    it('extracts markdown from exit_plan_mode result containing ## Plan:', () => {
      const result = `Your plan has been approved. You can now start coding.

Your plan has been saved at: C:\\Users\\plan.md

## Plan:
# Implementation Plan: Browser Login

### 1. Overview
Summary text here.`;

      const md = extractPlanMarkdownFromResult(result);
      expect(md).toContain('# Implementation Plan: Browser Login');
      expect(md).toContain('### 1. Overview');
      expect(md).not.toContain('Your plan has been approved');
    });

    it('extracts markdown if result is already formatted markdown', () => {
      const mdDoc = '# Master Plan\n\n- Task 1\n- Task 2';
      expect(extractPlanMarkdownFromResult(mdDoc)).toBe(mdDoc);
    });
  });
});
