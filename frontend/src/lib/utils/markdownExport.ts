import type { Session, WorkspaceFolder, ToolCall } from '../stores/session.svelte';

export function formatSessionAsMarkdown(session: Session, workspace?: WorkspaceFolder): string {
  const createdDate = new Date(session.createdAt).toLocaleString('id-ID', {
    dateStyle: 'full',
    timeStyle: 'medium'
  });
  const updatedDate = new Date(session.updatedAt).toLocaleString('id-ID', {
    dateStyle: 'full',
    timeStyle: 'medium'
  });

  const lines: string[] = [
    `# Session Log: ${session.title}`,
    ``,
    `| Field | Details |`,
    `| :--- | :--- |`,
    `| **Session ID** | \`${session.id}\` |`,
    `| **Workspace** | **${workspace?.name || 'Default Workspace'}** (\`${workspace?.path || '/'}\`) |`,
    `| **Status** | \`${session.status.toUpperCase()}\` |`,
    `| **Created At** | ${createdDate} |`,
    `| **Updated At** | ${updatedDate} |`,
    `| **Total Messages** | ${session.messages.length} |`,
    ``,
    `---`,
    ``,
    `## Conversation History`,
    ``
  ];

  if (session.messages.length === 0) {
    lines.push(`*(No messages in this session)*`);
    return lines.join('\n');
  }

  session.messages.forEach((msg, idx) => {
    const time = new Date(msg.timestamp).toLocaleTimeString('id-ID');
    const roleIcon = msg.role === 'user' ? '👤' : msg.role === 'assistant' ? '⚡' : '⚙️';
    const roleLabel = msg.role.toUpperCase();

    lines.push(`### [${time}] ${roleIcon} ${roleLabel} (Turn #${idx + 1})`);
    lines.push(``);
    lines.push(msg.content);
    lines.push(``);

    if (msg.images && msg.images.length > 0) {
      lines.push(`**Attached Vision Context (${msg.images.length} frame/s):**`);
      msg.images.forEach((img, imgIdx) => {
        lines.push(`- Frame #${imgIdx + 1}: \`${img.filePath || 'Snapshot'}\` (${img.width || 1920}x${img.height || 1080})`);
      });
      lines.push(``);
    }

    if (msg.toolCalls && msg.toolCalls.length > 0) {
      lines.push(`#### 🛠️ Executed Tools & Actions (${msg.toolCalls.length}):`);
      lines.push(``);
      msg.toolCalls.forEach((tool: ToolCall, tIdx: number) => {
        const duration = tool.startTime && tool.endTime ? `${tool.endTime - tool.startTime}ms` : '—';
        lines.push(`**${tIdx + 1}. Tool: \`${tool.tool}\`** (Status: \`${tool.status}\`, Duration: \`${duration}\`)`);
        
        if (tool.params) {
          const paramsFormatted = typeof tool.params === 'string' ? tool.params : JSON.stringify(tool.params, null, 2);
          lines.push(`- **Input Parameters:**`);
          lines.push('```json');
          lines.push(paramsFormatted);
          lines.push('```');
        }

        if (tool.result) {
          lines.push(`- **Execution Output:**`);
          lines.push('```');
          lines.push(tool.result);
          lines.push('```');
        }

        if (tool.diff && tool.diff.diffUnified) {
          lines.push(`- **Unified Diff Patch:**`);
          lines.push('```diff');
          lines.push(tool.diff.diffUnified);
          lines.push('```');
        }

        lines.push(``);
      });
    }

    if (msg.tokens) {
      lines.push(`*Tokens — In: ${msg.tokens.input || 0} | Out: ${msg.tokens.output || 0} | Total: ${msg.tokens.total || 0}*`);
      lines.push(``);
    }

    lines.push(`---`);
    lines.push(``);
  });

  return lines.join('\n');
}

export function formatMultipleSessionsAsMarkdown(sessions: Session[], workspace?: WorkspaceFolder): string {
  const exportTime = new Date().toLocaleString('id-ID', { dateStyle: 'full', timeStyle: 'medium' });
  const lines: string[] = [
    `# AetherGrok Workspace Export`,
    `**Workspace:** ${workspace?.name || 'Default'} (\`${workspace?.path || '/'}\`)`,
    `**Export Time:** ${exportTime}`,
    `**Total Exported Sessions:** ${sessions.length}`,
    ``,
    `---`,
    ``
  ];

  sessions.forEach((sess, i) => {
    lines.push(`## [${i + 1}/${sessions.length}] ${sess.title}`);
    lines.push(formatSessionAsMarkdown(sess, workspace));
    lines.push(`\n\n================================================================================\n\n`);
  });

  return lines.join('\n');
}

export async function downloadOrSaveMarkdown(
  filename: string,
  content: string
): Promise<{ success: boolean; filePath?: string; error?: string }> {
  // Try native Go Wails Save Dialog if available
  if (window.go?.main?.App?.SaveMarkdownExport) {
    try {
      const filePath = await window.go.main.App.SaveMarkdownExport(filename, content);
      if (!filePath) {
        return { success: false }; // Cancelled by user
      }
      return { success: true, filePath };
    } catch (err) {
      return { success: false, error: String(err) };
    }
  }

  // Browser Fallback (trigger DOM Blob download)
  try {
    const blob = new Blob([content], { type: 'text/markdown;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
    return { success: true, filePath: filename };
  } catch (err) {
    return { success: false, error: String(err) };
  }
}
