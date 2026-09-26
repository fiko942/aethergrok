import { marked } from 'marked';
import { highlightCode } from './codeHighlighter';

/**
 * Pre-processes text to detect un-backticked ASCII architecture diagrams,
 * flow charts, directory trees, and box drawings so they render as monospace blocks.
 */
export function preprocessMarkdownDiagrams(raw: string): string {
  if (!raw) return '';

  // 1. Temporarily protect already backticked code blocks
  const codeBlocks: string[] = [];
  const protectedContent = raw.replace(/```[\s\S]*?```/g, (match) => {
    codeBlocks.push(match);
    return `___PRE_SAVED_CODE_BLOCK_${codeBlocks.length - 1}___`;
  });

  // 2. Identify and group ASCII diagram blocks
  const paragraphs = protectedContent.split(/\n\s*\n/);
  const processedParagraphs: string[] = [];

  const diagramIndicators = [
    /^\s*\|/,                               // Lone pipe or starting pipe '|'
    /\+->|\+-->|-->|<--|<->|==>/,          // Arrow connectors
    /├──|└──|│|┌──|└──|┐|┘|└|┌|├|┤|┬|┴|┼/,  // Unicode box/tree connectors
    /^\s*\[[A-Za-z0-9\s/_\-.:]+\]\s*$/,     // Standalone [Box Name] header
    /^\s*\+[-=+]+\+\s*$/,                  // +------+ box boundary
    /^\s*\|\s+.*\s+\|\s*$/,                // | text | box row
    /^\s*\|\s*[A-Za-z0-9_]+\s*=\s*/        // Assignment indented under pipe '| Data = ...'
  ];

  let currentDiagramAccumulator: string[] = [];

  const flushDiagram = () => {
    if (currentDiagramAccumulator.length > 0) {
      processedParagraphs.push(`\`\`\`diagram\n${currentDiagramAccumulator.join('\n\n')}\n\`\`\``);
      currentDiagramAccumulator = [];
    }
  };

  for (const para of paragraphs) {
    const lines = para.split('\n');
    let isDiagram = false;

    if (lines.length >= 2) {
      let matchCount = 0;
      for (const line of lines) {
        if (diagramIndicators.some((regex) => regex.test(line))) {
          matchCount++;
        }
      }
      if (matchCount >= 2 && matchCount / lines.length >= 0.28) {
        isDiagram = true;
      }
    }

    if (isDiagram) {
      currentDiagramAccumulator.push(para);
    } else {
      flushDiagram();
      processedParagraphs.push(para);
    }
  }

  flushDiagram();

  let finalMarkdown = processedParagraphs.join('\n\n');

  // 3. Restore protected code blocks
  codeBlocks.forEach((block, idx) => {
    finalMarkdown = finalMarkdown.replace(`___PRE_SAVED_CODE_BLOCK_${idx}___`, block);
  });

  return finalMarkdown;
}

/**
 * Configure marked renderer with custom syntax highlighting,
 * language badges, copy buttons, and refined typography.
 */
const customRenderer = new marked.Renderer();

// Custom code block renderer with card header, language badge, and copy button
customRenderer.code = function ({ text, lang }: { text: string; lang?: string }) {
  const language = (lang || 'text').toLowerCase().trim();
  const highlighted = highlightCode(text, language);
  const displayLang = language === 'diagram' ? 'Architecture Flow / Diagram' : language.toUpperCase();

  // Escape text for data-code attribute to enable instant client copy
  const encodedCode = encodeURIComponent(text);

  return `
    <div class="my-3 rounded-xl overflow-hidden border border-white/10 bg-ant-bg-secondary/95 shadow-md group/codeblock not-prose">
      <div class="flex items-center justify-between px-3.5 py-1.5 bg-ant-bg-tertiary/60 border-b border-white/5 text-[11px] font-mono text-ant-text-muted select-none">
        <span class="font-semibold text-ant-primary/90 flex items-center gap-1.5">
          ${language === 'diagram' ? '<span class="w-2 h-2 rounded-full bg-cyan-400/80 animate-pulse"></span>' : ''}
          ${displayLang}
        </span>
        <button
          type="button"
          class="copy-code-btn px-2 py-0.5 rounded text-[10.5px] hover:text-ant-text hover:bg-white/5 transition flex items-center gap-1 cursor-pointer"
          data-raw-code="${encodedCode}"
        >
          <svg class="w-3 h-3 copy-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"></path></svg>
          <span class="copy-label">Copy</span>
        </button>
      </div>
      <div class="p-3.5 overflow-x-auto text-[12.5px] font-mono leading-relaxed select-text ${language === 'diagram' ? 'text-cyan-200/90 whitespace-pre font-mono tracking-tight' : 'text-ant-text'}">
        <code>${highlighted}</code>
      </div>
    </div>
  `;
};

marked.setOptions({
  gfm: true,
  breaks: true,
  renderer: customRenderer
});

export function renderMarkdown(content: string): string {
  if (!content) return '';
  try {
    const preprocessed = preprocessMarkdownDiagrams(content);
    return marked.parse(preprocessed) as string;
  } catch {
    return content;
  }
}
