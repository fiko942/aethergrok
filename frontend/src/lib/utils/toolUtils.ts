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
