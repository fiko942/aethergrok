// Utility functions for parsing, cleaning ANSI escapes, and formatting tool calls

const ANSI_REGEX = /\x1b\[[0-9;]*[a-zA-Z]/g;

/**
 * Clean ANSI terminal escape codes from text
 */
export function cleanANSI(text: string): string {
  if (!text) return '';
  return text.replace(ANSI_REGEX, '');
}

export interface ParsedToolOutput {
  rawText: string;
  output: string;
  command?: string;
  exitCode?: number;
  durationSecs?: number;
  status?: string;
  isJsonTaskOutput: boolean;
  truncated?: boolean;
  outputFile?: string;
}

/**
 * Parses and unwraps polymorphic tool output.
 * Handles JSON TaskOutput payloads (such as get_command_or_subagent_output / run_terminal_command results)
 * and strips ANSI codes for clean UI presentation.
 */
export function formatToolResult(result?: string): ParsedToolOutput {
  if (!result || typeof result !== 'string') {
    return {
      rawText: '',
      output: '',
      isJsonTaskOutput: false
    };
  }

  const trimmed = result.trim();

  // Try parsing JSON if it looks like a JSON object
  if (trimmed.startsWith('{') && trimmed.endsWith('}')) {
    try {
      const parsed = JSON.parse(trimmed);

      // Handle TaskOutput wrapper { "Result": { ... }, "type": "TaskOutput" }
      const resObj = parsed.Result || parsed.result || (parsed.type === 'TaskOutput' ? parsed : null);

      if (resObj && typeof resObj === 'object') {
        const cmd = typeof resObj.command === 'string' ? resObj.command : undefined;
        const exitCode = typeof resObj.exit_code === 'number' ? resObj.exit_code : (typeof resObj.exitCode === 'number' ? resObj.exitCode : undefined);
        const durationSecs = typeof resObj.duration_secs === 'number' ? resObj.duration_secs : undefined;
        const status = typeof resObj.status === 'string' ? resObj.status : undefined;
        const rawOut = typeof resObj.output === 'string' ? resObj.output : '';
        const truncated = Boolean(resObj.truncated);
        const outputFile = typeof resObj.output_file === 'string' ? resObj.output_file : undefined;

        if (rawOut || cmd || status !== undefined || exitCode !== undefined) {
          return {
            rawText: result,
            output: cleanANSI(rawOut),
            command: cmd,
            exitCode,
            durationSecs,
            status,
            isJsonTaskOutput: true,
            truncated,
            outputFile
          };
        }
      }
    } catch {
      // Ignore JSON parse errors and fall through to string cleaning
    }
  }

  return {
    rawText: result,
    output: cleanANSI(result),
    isJsonTaskOutput: false
  };
}

export interface SubagentParsedMeta {
  description: string;
  prompt: string;
  subagentType: string;
  model?: string;
  isolation?: string;
  isBackground: boolean;
  taskId?: string;
}

/**
 * Returns true if the tool call represents an autonomous subagent delegation
 */
export function isSubagentToolCall(toolName?: string, params?: unknown): boolean {
  if (!toolName) return false;
  const lower = toolName.toLowerCase().replace(/[_\s-]/g, '');
  if (lower === 'spawnsubagent' || lower === 'subagent' || lower === 'delegatetask' || lower === 'runagent') {
    return true;
  }

  if (params && typeof params === 'object') {
    const p = params as Record<string, unknown>;
    if (p.subagent_type || p.subagentType || p.subagent || p.agent_type) {
      return true;
    }
    if ((p.is_background === true || p.background === true) && typeof p.description === 'string' && /subagent/i.test(p.description)) {
      return true;
    }
  }

  return false;
}

/**
 * Detects whether the result is the standard Grok CLI background subagent launch acknowledgment
 * e.g. "Subagent started in background.\nsubagent_id: 01a15062-...\ndescription: ...\nWhen you need its result, use get_command_or_subagent_output..."
 */
export function isSubagentBackgroundAck(text?: string): boolean {
  if (!text || typeof text !== 'string') return false;
  return /subagent started in background\b/i.test(text) || /use get_command_or_subagent_output\b/i.test(text);
}

/**
 * Extracts subagent_id from background ack text if available
 */
export function extractSubagentIdFromText(text?: string): string | undefined {
  if (!text || typeof text !== 'string') return undefined;
  const match = /subagent_id:\s*[`"']?([0-9a-fA-F-]+)[`"']?/i.exec(text);
  return match ? match[1] : undefined;
}

/**
 * Extracts structured metadata from subagent parameters
 */
export function extractSubagentMeta(params?: unknown, result?: string): SubagentParsedMeta {
  let p: Record<string, unknown> = {};
  if (typeof params === 'string') {
    try {
      p = JSON.parse(params);
    } catch {
      p = { prompt: params };
    }
  } else if (params && typeof params === 'object') {
    p = params as Record<string, unknown>;
  }

  let description = typeof p.description === 'string' ? p.description.trim() : '';
  const prompt = typeof p.prompt === 'string' ? p.prompt.trim() : (typeof p.instruction === 'string' ? p.instruction.trim() : '');

  if (!description && prompt) {
    description = prompt.split('\n')[0].replace(/^[#\-*\s]+/, '').trim();
    if (description.length > 80) {
      description = description.slice(0, 77) + '...';
    }
  }

  if (!description) {
    description = 'Autonomous subagent task';
  }

  const subagentType = typeof p.subagent_type === 'string'
    ? p.subagent_type
    : (typeof p.subagentType === 'string' ? p.subagentType : (typeof p.type === 'string' ? p.type : 'general-purpose'));

  const model = typeof p.model === 'string' ? p.model : undefined;
  const isolation = typeof p.isolation === 'string' ? p.isolation : undefined;
  const isBackground = p.background === true || p.is_background === true;
  let taskId = typeof p.task_id === 'string' ? p.task_id : (typeof p.subagent_id === 'string' ? p.subagent_id : undefined);

  if (!taskId && result) {
    taskId = extractSubagentIdFromText(result);
  }

  return {
    description,
    prompt,
    subagentType,
    model,
    isolation,
    isBackground,
    taskId
  };
}

/**
 * Strips internal XML plumbing, lead-in boilerplate, and trailing agent hints.
 * Returns empty string if the text is merely the background launch acknowledgment.
 */
export function cleanSubagentOutput(text?: string): string {
  if (!text || typeof text !== 'string') return '';

  if (isSubagentBackgroundAck(text)) {
    return '';
  }

  let s = text
    .replace(/<subagent_(meta|result)>[\s\S]*?<\/subagent_\1>/g, '')
    .replace(/<\/?subagent_(meta|result)>/g, '')
    .trim();

  // Strip poller divider if present
  const outDivider = /^[\s\S]*?^===\s*Output\s*===\s*\n?/im;
  if (/^===\s*Task\s+/im.test(s) && outDivider.test(s)) {
    s = s.replace(outDivider, '').trim();
  }

  // Strip multi-wait header line if present (e.g. --- Task <id> [<status>] ---)
  s = s.replace(/^(?:---|===)\s*Task\s+[^\n]+(?:---|===)\s*\n?/i, '').trim();

  // Strip common CLI headers like Command: ..., Duration: ..., Exit Code: ...
  s = s.replace(/^(?:Command:\s*\[subagent:[^\]]+\][^\n]*\n)?(?:Duration:\s*[\d.]+\s*s\n)?(?:Exit Code:\s*\d+\n)?/i, '').trim();

  // Strip common CLI lead-in phrases
  s = s.replace(/^\[subagent:[^\]]*\]\s*/i, '');
  s = s.replace(/^this is the output of the subagent:\s*/i, '');
  s = s.replace(/^response:\s*/i, '');
  s = s.replace(/(?:\n|^)\s*agent id:\s*[\w.-]+\s*(\([^)]*\))?\s*$/i, '').trim();

  const wrapped = /^<response>\s*([\s\S]*?)\s*<\/response>$/i.exec(s);
  if (wrapped) {
    s = wrapped[1];
  }

  return s.trim();
}

/**
 * Formats duration milliseconds into human-readable stopwatch text (e.g. "45s", "1m 15s", "1h 2m 10s")
 */
export function formatSubagentDuration(ms: number): string {
  if (!Number.isFinite(ms) || ms < 0) return '0s';
  const totalSeconds = Math.max(1, Math.round(ms / 1000));
  if (totalSeconds < 60) {
    return `${totalSeconds}s`;
  }
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`;
  }
  return `${minutes}m ${seconds}s`;
}

export interface ExtractedTaskResult {
  taskId: string;
  output: string;
  status: 'completed' | 'failed' | 'running';
  durationMs?: number;
}

/**
 * Returns true if the tool call represents a task/subagent output reader (e.g. get_command_or_subagent_output)
 */
export function isSubagentOutputPoller(toolName?: string, params?: unknown): boolean {
  if (toolName) {
    const lower = toolName.toLowerCase().replace(/[\s-]/g, '_');
    if (
      lower.includes('get_command_or_subagent_output') ||
      lower.includes('command_output') ||
      lower.includes('subagent_output') ||
      lower.includes('task_output') ||
      lower.startsWith('output_') ||
      /^output\b/i.test(toolName)
    ) {
      return true;
    }
  }

  let p: Record<string, unknown> = {};
  if (typeof params === 'string') {
    try {
      p = JSON.parse(params);
    } catch {
      // not json
    }
  } else if (params && typeof params === 'object') {
    p = params as Record<string, unknown>;
  }

  if (p.variant === 'TaskOutput' || Array.isArray(p.task_ids) || typeof p.task_id === 'string' || typeof p.taskId === 'string') {
    return true;
  }

  return false;
}

/**
 * Extracts task IDs list from poller tool parameters
 */
export function extractTaskIdsFromParams(params?: unknown): string[] {
  let p: Record<string, unknown> = {};
  if (typeof params === 'string') {
    try {
      p = JSON.parse(params);
    } catch {
      // not json
    }
  } else if (params && typeof params === 'object') {
    p = params as Record<string, unknown>;
  }

  const ids: string[] = [];
  if (Array.isArray(p.task_ids)) {
    for (const id of p.task_ids) {
      if (typeof id === 'string' && id.trim()) ids.push(id.trim());
    }
  }
  if (typeof p.task_id === 'string' && p.task_id.trim()) {
    ids.push(p.task_id.trim());
  }
  if (typeof p.taskId === 'string' && p.taskId.trim()) {
    ids.push(p.taskId.trim());
  }
  return [...new Set(ids)];
}

/**
 * Extracts individual subagent results from a task poller output payload
 */
export function extractTaskPollerResults(
  toolName?: string,
  params?: unknown,
  result?: string
): ExtractedTaskResult[] {
  if (!result || typeof result !== 'string') return [];
  const trimmed = result.trim();
  const results: ExtractedTaskResult[] = [];

  // 1. Try parsing JSON format (TaskOutput with Result or MultiResult)
  if (trimmed.startsWith('{') && trimmed.endsWith('}')) {
    try {
      const parsed = JSON.parse(trimmed);
      const rawResults: Array<Record<string, unknown>> = [];

      if (parsed.MultiResult && Array.isArray(parsed.MultiResult.results)) {
        rawResults.push(...parsed.MultiResult.results);
      } else if (Array.isArray(parsed.results)) {
        rawResults.push(...parsed.results);
      } else if (parsed.Result && typeof parsed.Result === 'object') {
        rawResults.push(parsed.Result as Record<string, unknown>);
      } else if (parsed.result && typeof parsed.result === 'object') {
        rawResults.push(parsed.result as Record<string, unknown>);
      } else if (typeof parsed.task_id === 'string' || typeof parsed.taskId === 'string') {
        rawResults.push(parsed as Record<string, unknown>);
      }

      for (const item of rawResults) {
        const tid = String(item.task_id || item.taskId || '').trim();
        if (!tid) continue;
        const rawStatus = String(item.status || 'completed').toLowerCase();
        let status: 'completed' | 'failed' | 'running' = 'completed';
        if (rawStatus === 'running') {
          status = 'running';
        } else if (rawStatus === 'failed' || rawStatus === 'error') {
          status = 'failed';
        }

        let durationMs: number | undefined;
        if (typeof item.duration_secs === 'number') {
          durationMs = Math.round(item.duration_secs * 1000);
        } else if (typeof item.duration_ms === 'number') {
          durationMs = item.duration_ms;
        }

        const out = typeof item.output === 'string' ? cleanANSI(item.output) : '';
        // If output indicates still waiting/running, treat as running
        if (/task is still running\b/i.test(out) || /wait returned early\b/i.test(out)) {
          status = 'running';
        }

        results.push({
          taskId: tid,
          output: cleanSubagentOutput(out),
          status,
          durationMs
        });
      }

      if (results.length > 0) return results;
    } catch {
      // ignore JSON parse error
    }
  }

  // 2. Try parsing text blocks format:
  // Can be single: === Task <taskId> === ... === Output ===
  // Or multi-wait: --- Task <taskId> [<status>] ---
  // Or bracketed: === Task <taskId> [<status>] ===
  const taskBlockRegex = /(?:---|===)\s*Task\s+(\S+?)(?:\s*\[(\w+)\])?\s*(?:---|===)([\s\S]*?)(?=(?:(?:---|===)\s*Task\s+\S+)|$)/gi;
  let match: RegExpExecArray | null;
  while ((match = taskBlockRegex.exec(result)) !== null) {
    const tid = match[1].trim();
    const inlineStatus = match[2] ? match[2].toLowerCase() : '';
    const body = match[3];

    const statusM = /^\s*Status:\s*(\w+)/im.exec(body);
    const rawStatus = inlineStatus || (statusM ? statusM[1].toLowerCase() : 'completed');
    let status: 'completed' | 'failed' | 'running' = 'completed';
    if (rawStatus === 'running') status = 'running';
    else if (rawStatus === 'failed' || rawStatus === 'error') status = 'failed';

    // Check if body explicitly states still running or wait returned early
    if (/task is still running\b/i.test(body) || /subagent is still running\b/i.test(body) || /wait returned early\b/i.test(body)) {
      status = 'running';
    }

    let durationMs: number | undefined;
    const metaMs = /duration_ms\s*=\s*(\d+)/i.exec(body);
    const durSecs = /^\s*Duration:\s*([\d.]+)\s*s/im.exec(body);
    if (metaMs) durationMs = parseInt(metaMs[1], 10);
    else if (durSecs) durationMs = Math.round(parseFloat(durSecs[1]) * 1000);

    const outM = /^===\s*Output\s*===\s*\n?([\s\S]*)$/im.exec(body);
    let out = outM ? outM[1].trim() : body.trim();

    // Strip Command / Duration / Exit Code lines from out
    out = out.replace(/^(?:Command:\s*\[subagent:[^\]]+\][^\n]*\n)?(?:(?:Status|Duration|Exit Code):\s*[^\n]*\n)*/i, '').trim();

    results.push({
      taskId: tid,
      output: status === 'running' ? '' : cleanSubagentOutput(cleanANSI(out)),
      status,
      durationMs
    });
  }

  if (results.length > 0) return results;

  // 3. Fallback: single task ID from params if output is direct plain text
  const taskIds = extractTaskIdsFromParams(params);
  if (taskIds.length === 1) {
    const tid = taskIds[0];
    const isStillRunning = /task is still running\b/i.test(result) ||
      /wait returned early\b/i.test(result) ||
      /waited the requested\b/i.test(result);

    results.push({
      taskId: tid,
      output: isStillRunning ? '' : cleanSubagentOutput(cleanANSI(result)),
      status: isStillRunning ? 'running' : 'completed'
    });
  }

  return results;
}

/**
 * Checks if a tool call is a redundant subagent poller whose output belongs inside a subagent card
 */
export function isRedundantSubagentPoller(
  toolCall: { tool?: string; params?: unknown; result?: string; id?: string },
  allToolCalls?: Array<{ tool?: string; params?: unknown; result?: string; id?: string }>
): boolean {
  if (!isSubagentOutputPoller(toolCall.tool, toolCall.params)) {
    return false;
  }

  const pollerTaskIds = extractTaskIdsFromParams(toolCall.params);
  if (pollerTaskIds.length === 0) {
    // If tool name is explicitly get_command_or_subagent_output or starts with output
    const lower = (toolCall.tool || '').toLowerCase();
    return lower.includes('get_command_or_subagent_output') || lower.includes('subagent_output');
  }

  if (allToolCalls && allToolCalls.length > 0) {
    // Check if any subagent in the list matches one of these taskIds
    const hasMatchingSubagent = allToolCalls.some((tc) => {
      if (!isSubagentToolCall(tc.tool, tc.params)) return false;
      const meta = extractSubagentMeta(tc.params, tc.result);
      return meta.taskId && pollerTaskIds.includes(meta.taskId);
    });
    if (hasMatchingSubagent) return true;
  }

  // If tool name is explicitly get_command_or_subagent_output, suppress from main list
  const lower = (toolCall.tool || '').toLowerCase();
  return lower.includes('get_command_or_subagent_output') || lower.includes('subagent_output');
}

export interface ChildActionSummary {
  verb: string;
  target: string;
  iconType: 'read' | 'write' | 'edit' | 'search' | 'terminal' | 'list' | 'tool';
}

/**
 * Extracts a clean, concise single-line label for a subagent's child tool call
 */
export function formatActionSummary(toolName?: string, params?: unknown): ChildActionSummary {
  const rawTool = (toolName || '').toLowerCase().replace(/[\s-]/g, '_');
  let p: Record<string, unknown> = {};
  if (typeof params === 'string') {
    try {
      p = JSON.parse(params);
    } catch {
      p = {};
    }
  } else if (params && typeof params === 'object') {
    p = params as Record<string, unknown>;
  }

  // 1. Read
  if (rawTool.includes('read_file') || rawTool.includes('file_read') || rawTool.includes('cat')) {
    const filePath = String(p.target_file || p.file_path || p.path || p.filePath || '');
    let lineRange = '';
    if (p.offset && p.limit) {
      const startLine = Number(p.offset);
      const endLine = startLine + Number(p.limit) - 1;
      lineRange = ` (${startLine}-${endLine})`;
    } else if (p.offset) {
      lineRange = ` (${p.offset}+)`;
    } else if (p.limit) {
      lineRange = ` (1-${p.limit})`;
    }
    return {
      verb: 'Read',
      target: filePath ? `${filePath}${lineRange}` : 'file',
      iconType: 'read'
    };
  }

  // 2. Edit
  if (rawTool.includes('search_replace') || rawTool.includes('edit') || rawTool.includes('str_replace')) {
    const filePath = String(p.file_path || p.target_file || p.path || p.filePath || '');
    return {
      verb: 'Edit',
      target: filePath || 'file',
      iconType: 'edit'
    };
  }

  // 3. Write
  if (rawTool.includes('write') || rawTool.includes('save') || rawTool.includes('create_file')) {
    const filePath = String(p.file_path || p.path || p.target_file || p.filePath || '');
    return {
      verb: 'Write',
      target: filePath || 'file',
      iconType: 'write'
    };
  }

  // 4. Search
  if (rawTool.includes('grep') || rawTool.includes('search') || rawTool.includes('find')) {
    const query = String(p.pattern || p.query || p.text || p.regex || p.glob_pattern || '');
    const path = String(p.path || p.target_directory || p.directory || '');
    const target = query ? `"${query}"${path && path !== '.' ? ` in ${path}` : ''}` : path || 'files';
    return {
      verb: 'Search',
      target,
      iconType: 'search'
    };
  }

  // 5. List
  if (rawTool.includes('list_dir') || rawTool.includes('dir_list') || rawTool.includes('ls') || rawTool.includes('list')) {
    const dir = String(p.target_directory || p.directory || p.path || p.dir || '');
    return {
      verb: 'List',
      target: dir || 'directory',
      iconType: 'list'
    };
  }

  // 6. Terminal / Bash
  if (rawTool.includes('run_terminal') || rawTool.includes('terminal') || rawTool.includes('bash') || rawTool.includes('execute') || rawTool.includes('sh')) {
    const cmd = String(p.command || p.cmd || '');
    const desc = typeof p.description === 'string' && p.description.trim() ? p.description.trim() : cmd;
    return {
      verb: 'Run',
      target: desc || 'command',
      iconType: 'terminal'
    };
  }

  return {
    verb: 'Tool',
    target: toolName || 'tool',
    iconType: 'tool'
  };
}


