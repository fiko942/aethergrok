import { describe, it, expect } from 'vitest';

// Define Svelte 5 runes mock on globalThis for node test environment
if (typeof (globalThis as any).$state === 'undefined') {
  (globalThis as any).$state = (v: any) => v;
}
if (typeof (globalThis as any).$derived === 'undefined') {
  (globalThis as any).$derived = (v: any) => v;
}

import {
  extractSubagentMeta,
  cleanSubagentOutput,
  formatSubagentDuration,
  extractTaskPollerResults
} from '../frontend/src/lib/utils/toolUtils';

const { dialogStore } = await import('../frontend/src/lib/stores/dialog.svelte');

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

  it('manages subagent modal open and close state via dialogStore correctly', () => {
    const mockToolCall = {
      id: 'tc-sub-3',
      tool: 'spawn_subagent',
      status: 'running' as const,
      params: { description: 'Audit dependencies' }
    };
    const mockMeta = extractSubagentMeta(mockToolCall.params);

    expect(dialogStore.subagentModalState.open).toBe(false);

    dialogStore.openSubagentModal(mockToolCall, mockMeta);
    expect(dialogStore.subagentModalState.open).toBe(true);
    expect(dialogStore.subagentModalState.toolCall?.id).toBe('tc-sub-3');
    expect(dialogStore.subagentModalState.meta?.description).toBe('Audit dependencies');

    dialogStore.closeSubagentModal();
    expect(dialogStore.subagentModalState.open).toBe(false);
    expect(dialogStore.subagentModalState.toolCall).toBeNull();
    expect(dialogStore.subagentModalState.meta).toBeNull();
  });

  it('reconciles subagent outputs from task pollers across session messages', async () => {
    const { reconcileSessionSubagents } = await import('../frontend/src/lib/stores/session.svelte');

    const subagentCall = {
      id: 'sub-call-1',
      tool: 'spawn_subagent',
      status: 'running' as const,
      startTime: 1000,
      params: { description: 'Subagent 1 Academic WS3 WS4' },
      result: 'Subagent started in background.\nsubagent_id: 01a15062-c047-7d91-a7e1-10e862e35c64'
    };

    const pollerCall = {
      id: 'poller-call-1',
      tool: 'get_command_or_subagent_output',
      status: 'completed' as const,
      params: { task_ids: ['01a15062-c047-7d91-a7e1-10e862e35c64'] },
      result: JSON.stringify({
        type: 'TaskOutput',
        Result: {
          task_id: '01a15062-c047-7d91-a7e1-10e862e35c64',
          status: 'completed',
          duration_secs: 14.5,
          output: 'Full Academic analysis for WS3 and WS4 completed.'
        }
      })
    };

    const mockSession: any = {
      id: 'sess-1',
      messages: [
        {
          id: 'msg-1',
          role: 'assistant',
          toolCalls: [subagentCall, pollerCall]
        }
      ]
    };

    reconcileSessionSubagents(mockSession);

    // The subagent card must receive the completed output from the poller
    expect(subagentCall.status).toBe('completed');
    expect(subagentCall.result).toBe('Full Academic analysis for WS3 and WS4 completed.');
    expect(subagentCall.endTime).toBe(1000 + 14500);
  });

  it('retrieves subagent-specific execution plan tasks and computes progress', async () => {
    const { PlanStore } = await import('../frontend/src/lib/stores/plan.svelte');
    const planStore = new PlanStore();

    // Ingest output for subagent-task-1
    planStore.inspectToolForPlan(
      'sess-test',
      'sub-task-1',
      'spawn_subagent',
      {
        description: 'TDD Engine Builder',
        prompt: 'Build test engine'
      },
      '- [x] Step 1: Write unit tests\n- [x] Step 2: Run tests\n- [ ] Step 3: Implement core engine\n- [ ] Step 4: Re-test',
      { id: 'sub-task-1', name: 'TDD Engine Builder' }
    );

    const subPlan = planStore.getSubagentPlan('sess-test', 'sub-task-1');
    expect(subPlan).not.toBeNull();
    expect(subPlan?.subagentName).toBe('TDD Engine Builder');
    expect(subPlan?.totalCount).toBe(4);
    expect(subPlan?.completedCount).toBe(2);
    expect(subPlan?.pendingCount).toBe(2);
    expect(subPlan?.progressPercent).toBe(50);
    expect(subPlan?.todos[0].content).toBe('Step 1: Write unit tests');
    expect(subPlan?.todos[0].status).toBe('completed');
    expect(subPlan?.todos[2].status).toBe('pending');
  });

  it('tracks childToolCalls inside parent subagent ToolCall correctly', () => {
    const parentSubagent: any = {
      id: 'subagent-root',
      tool: 'spawn_subagent',
      status: 'running',
      childToolCalls: []
    };

    const childToolCall1 = {
      id: 'child-tc-1',
      tool: 'read_file',
      params: { target_file: 'src/main.go', offset: 1, limit: 30 },
      status: 'completed',
      result: 'package main...'
    };

    const childToolCall2 = {
      id: 'child-tc-2',
      tool: 'run_terminal_command',
      params: { command: 'go test ./...', description: 'Run test suite' },
      status: 'running'
    };

    parentSubagent.childToolCalls.push(childToolCall1);
    parentSubagent.childToolCalls.push(childToolCall2);

    expect(parentSubagent.childToolCalls).toHaveLength(2);
    expect(parentSubagent.childToolCalls[0].tool).toBe('read_file');
    expect(parentSubagent.childToolCalls[1].status).toBe('running');

    // Updating existing child tool call
    const targetChild = parentSubagent.childToolCalls.find((c: any) => c.id === 'child-tc-2');
    if (targetChild) {
      targetChild.status = 'completed';
      targetChild.result = 'PASS\nok';
    }

    expect(parentSubagent.childToolCalls[1].status).toBe('completed');
    expect(parentSubagent.childToolCalls[1].result).toBe('PASS\nok');
  });

  it('correctly parses multi-wait task poller text blocks and reconciles session subagents', async () => {
    const rawMultiWait = `=== Multi-wait (wait_all) ===
--- Task 01a15062-c047-7d91-a7e1-10e862e35c64 [completed] ---
Command: [subagent:general-purpose] Subagent 1 Academic WS3 WS4
Duration: 94.19s
Exit Code: 0
Laporan analisis mendalam untuk **Cluster 1** telah berhasil disusun.

### Ringkasan Eksekutif Analisis Cluster 1:
- Parameter Pengguna: 10
- Metodologi: Think-Aloud & In-Depth Interview

<subagent_meta>id=01a15062-c047-7d91-a7e1-10e862e35c64, tool_calls=13, turns=1, duration_ms=94188</subagent_meta>
<subagent_result>
subagent_id: 01a15062-c047-7d91-a7e1-10e862e35c64
</subagent_result>
--- Task 01a15062-dd04-7a61-85eb-ed3c0f16c040 [running] ---
Command: [subagent:general-purpose] Subagent 2 Empirical WS5 WS6
Duration: 71.31s
Subagent is still running.
Type: general-purpose
Progress: turn 1, 15 tool calls

1/2 tasks completed (wait_all)`;

    const results = extractTaskPollerResults('get_command_or_subagent_output', {
      task_ids: ['01a15062-c047-7d91-a7e1-10e862e35c64', '01a15062-dd04-7a61-85eb-ed3c0f16c040']
    }, rawMultiWait);

    expect(results).toHaveLength(2);

    // First task should be completed with extracted duration and clean output
    const task1 = results.find(r => r.taskId === '01a15062-c047-7d91-a7e1-10e862e35c64');
    expect(task1).toBeDefined();
    expect(task1?.status).toBe('completed');
    expect(task1?.durationMs).toBe(94188);
    expect(task1?.output).toContain('Laporan analisis mendalam untuk **Cluster 1**');
    expect(task1?.output).not.toContain('Command: [subagent:');
    expect(task1?.output).not.toContain('<subagent_meta>');

    // Second task should be running
    const task2 = results.find(r => r.taskId === '01a15062-dd04-7a61-85eb-ed3c0f16c040');
    expect(task2).toBeDefined();
    expect(task2?.status).toBe('running');
    expect(task2?.output).toBe('');

    // Reconcile across session messages
    const { reconcileSessionSubagents } = await import('../frontend/src/lib/stores/session.svelte');

    const sub1 = {
      id: 'tc-spawn-1',
      tool: 'spawn_subagent',
      status: 'running' as const,
      startTime: 1000,
      params: { description: 'Subagent 1 Academic' },
      result: 'Subagent started in background.\nsubagent_id: 01a15062-c047-7d91-a7e1-10e862e35c64'
    };

    const sub2 = {
      id: 'tc-spawn-2',
      tool: 'spawn_subagent',
      status: 'running' as const,
      startTime: 2000,
      params: { description: 'Subagent 2 Empirical' },
      result: 'Subagent started in background.\nsubagent_id: 01a15062-dd04-7a61-85eb-ed3c0f16c040'
    };

    const poller = {
      id: 'tc-poll-1',
      tool: 'get_command_or_subagent_output',
      status: 'completed' as const,
      params: { task_ids: ['01a15062-c047-7d91-a7e1-10e862e35c64', '01a15062-dd04-7a61-85eb-ed3c0f16c040'] },
      result: rawMultiWait
    };

    const mockSession: any = {
      id: 'sess-reconcile-test',
      messages: [{ id: 'm-1', role: 'assistant', toolCalls: [sub1, sub2, poller] }]
    };

    reconcileSessionSubagents(mockSession);

    expect(sub1.status).toBe('completed');
    expect(sub1.result).toContain('Laporan analisis mendalam untuk **Cluster 1**');
    expect(sub1.endTime).toBe(1000 + 94188);

    // Subagent 2 should still be running
    expect(sub2.status).toBe('running');
  });

  it('synchronizes subagent status, output, and childToolCalls into message parts', async () => {
    const { syncMessageToolCallParts } = await import('../frontend/src/lib/stores/session.svelte');

    const subTc = {
      id: 'tc-sub-sync-1',
      tool: 'spawn_subagent',
      status: 'completed' as const,
      result: 'Detailed execution result from subagent',
      endTime: 12500,
      childSessionId: 'sub-sess-999',
      childToolCalls: [
        { id: 'c1', tool: 'read_file', status: 'completed' as const },
        { id: 'c2', tool: 'write', status: 'completed' as const }
      ]
    };

    const mockMessage: any = {
      id: 'msg-sync-1',
      role: 'assistant',
      toolCalls: [subTc],
      parts: [
        {
          type: 'tool_calls',
          toolCalls: [
            {
              id: 'tc-sub-sync-1',
              tool: 'spawn_subagent',
              status: 'running',
              result: 'Started...'
            }
          ]
        }
      ]
    };

    syncMessageToolCallParts(mockMessage);

    const partTool = mockMessage.parts[0].toolCalls[0];
    expect(partTool.status).toBe('completed');
    expect(partTool.result).toBe('Detailed execution result from subagent');
    expect(partTool.endTime).toBe(12500);
    expect(partTool.childSessionId).toBe('sub-sess-999');
    expect(partTool.childToolCalls).toHaveLength(2);
    expect(partTool.childToolCalls[0].tool).toBe('read_file');
    expect(partTool.childToolCalls[1].tool).toBe('write');
  });

  it('syncToolCallUpdate propagates subagent updates to store sessions and messages', async () => {
    const { sessionStore } = await import('../frontend/src/lib/stores/session.svelte');

    const testSessionId = 'sess-sync-update-test';
    const subToolId = 'tc-sub-store-sync';

    // Inject session into sessionStore
    const testSession: any = {
      id: testSessionId,
      workspaceId: 'ws-test',
      title: 'Subagent Persistence Test',
      status: 'idle',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      messages: [
        {
          id: 'msg-store-1',
          role: 'assistant',
          toolCalls: [
            {
              id: subToolId,
              tool: 'spawn_subagent',
              status: 'running',
              params: { description: 'Subagent Background Worker' },
              result: 'Subagent started in background.\nsubagent_id: 01a99999-sub'
            }
          ],
          parts: [
            {
              type: 'tool_calls',
              toolCalls: [
                {
                  id: subToolId,
                  tool: 'spawn_subagent',
                  status: 'running',
                  params: { description: 'Subagent Background Worker' }
                }
              ]
            }
          ]
        }
      ]
    };

    sessionStore.sessions.push(testSession);

    // Update tool call via syncToolCallUpdate
    const updatedToolCall: any = {
      id: subToolId,
      tool: 'spawn_subagent',
      status: 'completed',
      result: 'Worker completed all assigned operations successfully.',
      childSessionId: '01a99999-sub',
      endTime: 98765,
      childToolCalls: [
        { id: 'ch-1', tool: 'run_terminal_command', status: 'completed', result: 'All tests passed' }
      ]
    };

    sessionStore.syncToolCallUpdate(updatedToolCall);

    const sessionInStore = sessionStore.sessions.find(s => s.id === testSessionId);
    expect(sessionInStore).toBeDefined();

    const storedMsg = sessionInStore?.messages[0];
    expect(storedMsg?.toolCalls?.[0].status).toBe('completed');
    expect(storedMsg?.toolCalls?.[0].result).toBe('Worker completed all assigned operations successfully.');
    expect(storedMsg?.toolCalls?.[0].childToolCalls).toHaveLength(1);
    expect(storedMsg?.toolCalls?.[0].childSessionId).toBe('01a99999-sub');

    // Also verify parts
    const partTool = (storedMsg?.parts?.[0] as any)?.toolCalls?.[0];
    expect(partTool.status).toBe('completed');
    expect(partTool.result).toBe('Worker completed all assigned operations successfully.');
    expect(partTool.childToolCalls).toHaveLength(1);
  });
});
