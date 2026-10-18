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
 * Extracts structured metadata from subagent parameters
 */
export function extractSubagentMeta(params?: unknown): SubagentParsedMeta {
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
  const taskId = typeof p.task_id === 'string' ? p.task_id : (typeof p.subagent_id === 'string' ? p.subagent_id : undefined);

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
 * Strips internal XML plumbing, lead-in boilerplate, and trailing agent hints
 */
export function cleanSubagentOutput(text?: string): string {
  if (!text || typeof text !== 'string') return '';

  let s = text
    .replace(/<subagent_(meta|result)>[\s\S]*?<\/subagent_\1>/g, '')
    .replace(/<\/?subagent_(meta|result)>/g, '')
    .trim();

  // Strip poller divider if present
  const outDivider = /^[\s\S]*?^===\s*Output\s*===\s*\n?/im;
  if (/^===\s*Task\s+/im.test(s) && outDivider.test(s)) {
    s = s.replace(outDivider, '').trim();
  }

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

