import { marked } from 'marked';
import { highlightCode } from './codeHighlighter';

/**
 * File extension list used to extract file candidates from prose and code spans
 */
const FILE_EXTS = 'ts|tsx|js|jsx|svelte|vue|py|go|rs|rb|php|java|c|cpp|h|hpp|cs|swift|kt|json|yaml|yml|toml|xml|html|css|scss|sass|less|md|markdown|txt|sh|bash|zsh|sql|lock|plist|conf|ini|log|spec';

export interface FilePathInfo {
  exists: boolean;
  fullPath: string;
  relPath?: string;
  isDir?: boolean;
  sizeBytes?: number;
}

/**
 * Extracts candidate file paths from markdown prose or inline code
 */
export function extractCandidateFilePaths(content: string): string[] {
  if (!content) return [];
  const candidates = new Set<string>();

  // 1. Markdown inline codespans `some/file.py` or `license_manager.py`
  const codeSpanRegex = /`([^`\n]+)`/g;
  let match: RegExpExecArray | null;
  while ((match = codeSpanRegex.exec(content)) !== null) {
    const raw = cleanCandidate(match[1]);
    if (isValidPathCandidate(raw)) {
      candidates.add(raw);
    }
  }

  // 2. Plain text path candidates (e.g. .github/workflows/build-windows.yml, /Users/..., ~/.grok/config.toml)
  const pathRegex = new RegExp(
    `(?:^|[\\s(\`"'\\[<])((?:(?:\\/|[a-zA-Z]:[\\\\/]|~[\\\\/]|\\.\\.?[\\\\/]|\\.[\\w-]+\\/)[\\w.\\-\\/]+|[\\w.\\-\\/]+\\.(?:${FILE_EXTS})))(?=[.,;:\\s\`"'\\]>]|$)`,
    'g'
  );

  while ((match = pathRegex.exec(content)) !== null) {
    const raw = cleanCandidate(match[1]);
    if (isValidPathCandidate(raw)) {
      candidates.add(raw);
    }
  }

  return Array.from(candidates);
}

function cleanCandidate(raw: string): string {
  return raw.replace(/^[`"'(\[{<]+/, '').replace(/[`"')\]}>.,;:!?]+$/, '').trim();
}

function isValidPathCandidate(p: string): boolean {
  if (!p || p.length < 2 || p.length > 250) return false;
  if (p.startsWith('http://') || p.startsWith('https://')) return false;
  if (/^\d+(\.\d+)+$/.test(p)) return false; // Version strings like 1.0.0
  if (/^\d+\.\s*$/.test(p)) return false;     // Numbered list item like 1.
  return true;
}

/**
 * Renders an interactive blue badge chip for existing local files
 */
function renderFileChip(fullPath: string, label: string): string {
  const encoded = encodeURIComponent(fullPath);
  return `<button type="button" class="inline-file-chip inline-flex items-center gap-1 px-1.5 py-0.5 my-0.5 rounded text-[11.5px] font-mono font-medium text-sky-400 bg-sky-500/10 hover:bg-sky-500/20 border border-sky-500/25 hover:border-sky-500/50 transition cursor-pointer" data-file-path="${encoded}" title="Click to open file: ${label}"><svg class="w-3 h-3 text-sky-400 inline shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg><span>${label}</span></button>`;
}

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
export function createMarkdownRenderer(existingFilesMap: Record<string, FilePathInfo> = {}) {
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

  // Custom codespan renderer: turns verified file paths in `...` into clickable chips
  customRenderer.codespan = function ({ text }: { text: string }) {
    const clean = cleanCandidate(text);
    if (existingFilesMap[clean] && existingFilesMap[clean].exists) {
      return renderFileChip(existingFilesMap[clean].fullPath, clean);
    }
    return `<code class="font-mono text-[11.5px] bg-ant-bg-tertiary px-1.5 py-0.5 rounded text-ant-primary">${text}</code>`;
  };

  // Custom text token renderer: turns verified file paths in prose into clickable blue chips
  customRenderer.text = function ({ text }: { text: string }) {
    let res = text;
    const sortedPaths = Object.keys(existingFilesMap).sort((a, b) => b.length - a.length);
    for (const cand of sortedPaths) {
      const info = existingFilesMap[cand];
      if (info && info.exists) {
        const escaped = cand.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        const reg = new RegExp(`(?<=^|[\\s(\`"'\\[<])${escaped}(?=[.,;:\\s\`"'\\]>]|$)`, 'g');
        res = res.replace(reg, renderFileChip(info.fullPath, cand));
      }
    }
    return res;
  };

  return customRenderer;
}

export function renderMarkdown(content: string, existingFilesMap: Record<string, FilePathInfo> = {}): string {
  if (!content) return '';
  try {
    const preprocessed = preprocessMarkdownDiagrams(content);
    const renderer = createMarkdownRenderer(existingFilesMap);
    return marked.parse(preprocessed, {
      gfm: true,
      breaks: true,
      renderer
    }) as string;
  } catch {
    return content;
  }
}
