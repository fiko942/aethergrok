import { marked } from 'marked';
import katex from 'katex';
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

  // 2. Plain text path candidates (e.g. .github/workflows/build-windows.yml, /Users/..., C:\..., ~/.grok/config.toml)
  const pathRegex = new RegExp(
    `(?:^|[\\s(\`"'\\[<])((?:(?:\\/|[a-zA-Z]:[\\\\/]|~[\\\\/]|\\.\\.?[\\\\/]|\\.[\\w-]+\\/)[\\w.\\-\\/\\\\]+|[\\w.\\-\\/\\\\]+\\.(?:${FILE_EXTS})))(?=[.,;:\\s\`"'\\]>]|$)`,
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
 * Renders an interactive badge chip for existing local files or folders
 */
function renderFileChip(fullPath: string, label: string, isDir = false): string {
  const encoded = encodeURIComponent(fullPath);
  const iconSvg = isDir
    ? `<svg class="w-3 h-3 text-sky-800 dark:text-sky-400 inline shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"></path></svg>`
    : `<svg class="w-3 h-3 text-sky-800 dark:text-sky-400 inline shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path></svg>`;

  return `<button type="button" class="inline-file-chip inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[11px] leading-tight font-mono font-semibold text-sky-950 dark:text-sky-300 bg-sky-100 hover:bg-sky-200/90 dark:bg-sky-500/10 dark:hover:bg-sky-500/20 border border-sky-300/80 dark:border-sky-500/25 hover:border-sky-400 dark:hover:border-sky-500/50 transition cursor-pointer align-middle max-w-full truncate whitespace-nowrap shadow-2xs" data-file-path="${encoded}" data-is-dir="${isDir ? 'true' : 'false'}" title="${isDir ? 'Click to reveal folder in Finder/Explorer' : 'Click to preview file'}: ${label}">${iconSvg}<span class="truncate">${label}</span></button>`;
}

/**
 * Normalizes and repairs nested markdown code fences.
 * When an assistant outputs ```markdown containing nested ```bash or other blocks,
 * using 3 backticks for both outer and inner blocks causes standard CommonMark parsers
 * to close the outer block prematurely, leaving the rest of the message as a raw code block.
 * This lifts outer ```markdown / ```md blocks to 4 backticks (````markdown ... ````).
 */
export function repairMarkdownFences(markdown: string): string {
  if (!markdown) return '';

  const lines = markdown.split('\n');
  const result: string[] = [];
  
  let inMarkdownBlock = false;
  let nestedOpen = false;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const match = line.match(/^([ \t]*)(`{3,}|~{3,})(.*)$/);

    if (match) {
      const fenceLen = match[2].length;
      const info = match[3].trim();
      const lang = info.split(/\s+/)[0].toLowerCase();

      if (!inMarkdownBlock) {
        if (['markdown', 'md'].includes(lang) && fenceLen === 3) {
          inMarkdownBlock = true;
          result.push('````' + (info || 'markdown'));
          continue;
        } else {
          result.push(line);
        }
      } else {
        if (info !== '') {
          // Inner opening fence (e.g. ```bash)
          nestedOpen = true;
          result.push(line);
        } else {
          if (nestedOpen) {
            nestedOpen = false;
            result.push(line);
          } else {
            inMarkdownBlock = false;
            result.push('````');
          }
        }
      }
    } else {
      result.push(line);
    }
  }

  if (inMarkdownBlock) {
    result.push('````');
  }

  return result.join('\n');
}

/**
 * Pre-processes text to detect un-backticked ASCII architecture diagrams,
 * flow charts, directory trees, and box drawings so they render as monospace blocks.
 */
export function preprocessMarkdownDiagrams(raw: string): string {
  if (!raw) return '';

  // 1. Temporarily protect already backticked code blocks (support 3, 4, or more backticks)
  const codeBlocks: string[] = [];
  const protectedContent = raw.replace(/(`{3,}|~{3,})[\s\S]*?\1/g, (match) => {
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
    const trimmed = para.trim();
    const lines = trimmed.split('\n');
    let isDiagram = false;

    // Check if paragraph is a standard Markdown GFM Table
    // If line 2 is a table header separator (| :--- | :--- |), keep it as markdown table, not diagram!
    const isTable = lines.length >= 2 && /^\s*\|?\s*:?-{2,}:?\s*(\|\s*:?-{2,}:?\s*)+\|?\s*$/.test(lines[1]);

    if (!isTable && lines.length >= 2) {
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
 * Strips common leading whitespace / indentation from code blocks
 * (e.g. when a code fence is nested inside markdown numbered or bullet lists).
 */
export function dedentCode(text: string): string {
  if (!text) return '';
  const lines = text.split('\n');
  let minIndent = Infinity;

  for (const line of lines) {
    if (line.trim().length === 0) continue;
    const match = line.match(/^[ \t]*/);
    if (match) {
      minIndent = Math.min(minIndent, match[0].length);
    }
  }

  if (minIndent > 0 && minIndent !== Infinity) {
    return lines.map(line => (line.trim().length === 0 ? '' : line.slice(minIndent))).join('\n');
  }

  return text;
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
    const cleanText = dedentCode(text);
    const highlighted = highlightCode(cleanText, language);
    const displayLang = language === 'diagram' ? 'Architecture Flow / Diagram' : language.toUpperCase();

    // Escape text for data-code attribute to enable instant client copy
    const encodedCode = encodeURIComponent(cleanText);

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
        <div class="p-3.5 overflow-x-auto text-[12.5px] font-mono leading-relaxed select-text whitespace-pre ${language === 'diagram' ? 'text-cyan-200/90 tracking-tight' : 'text-ant-text'}">
          <code>${highlighted}</code>
        </div>
      </div>
    `;
  };

  // Custom link renderer: ensure target="_blank" and rel="noopener noreferrer"
  customRenderer.link = function ({ href, title, text }: { href: string; title?: string | null; text: string }) {
    const titleAttr = title ? ` title="${title}"` : '';
    return `<a href="${href}" target="_blank" rel="noopener noreferrer"${titleAttr} class="text-ant-primary hover:text-blue-400 underline underline-offset-2 transition">${text}</a>`;
  };

  // Custom codespan renderer: turns verified file paths in `...` into clickable chips
  customRenderer.codespan = function ({ text }: { text: string }) {
    const clean = cleanCandidate(text);
    if (existingFilesMap[clean] && existingFilesMap[clean].exists) {
      return renderFileChip(existingFilesMap[clean].fullPath, clean, existingFilesMap[clean].isDir);
    }
    return `<code class="font-mono text-[11.5px] bg-ant-bg-tertiary px-1.5 py-0.5 rounded text-blue-900 dark:text-blue-300 border border-ant-border-secondary dark:border-white/5 font-semibold">${text}</code>`;
  };

  // Custom text token renderer: handles marked v18 nested tokens and turns verified file paths in prose into clickable blue chips
  customRenderer.text = function (token: any) {
    // If marked v18 passes a token with child tokens (e.g. bold, codespan, links inside list items), parse children
    if (token && typeof token === 'object' && 'tokens' in token && token.tokens && (this as any).parser) {
      return (this as any).parser.parseInline(token.tokens);
    }

    let res = typeof token === 'string' ? token : token?.text || '';
    const sortedPaths = Object.keys(existingFilesMap).sort((a, b) => b.length - a.length);
    for (const cand of sortedPaths) {
      const info = existingFilesMap[cand];
      if (info && info.exists) {
        const escaped = cand.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        const reg = new RegExp(`(?<=^|[\\s(\`"'\\[<])${escaped}(?=[.,;:\\s\`"'\\]>]|$)`, 'g');
        res = res.replace(reg, renderFileChip(info.fullPath, cand, info.isDir));
      }
    }
    return res;
  };

  return customRenderer;
}

/**
 * Strips unsupported TeX commands (like \label{...}) that Grok often emits
 * inside equations or align blocks for cross-referencing, but which KaTeX
 * has no ref system for and would otherwise render as red error tokens.
 */
export function stripUnsupportedTex(tex: string): string {
  if (!tex) return '';
  return tex.replace(/\\label\s*\{[^}]*\}/g, '');
}

/**
 * Checks whether an inline $...$ match looks like plain currency (e.g. $5, $10.50, $100k, $5 - $10)
 * rather than a mathematical formula.
 */
export function isLikelyCurrency(text: string): boolean {
  const trimmed = text.trim();
  // Pure numbers or quantities like 5, 10.50, 100k, 2.5m, 50 million
  if (/^\d+([.,]\d+)?\s*(k|m|b|million|billion)?$/i.test(trimmed)) {
    return true;
  }
  // Price ranges like 5-10, 5 - 10, 10 - 20
  if (/^\d+\s*[-–—]\s*\d+$/.test(trimmed)) {
    return true;
  }
  return false;
}

/**
 * Renders a LaTeX string into KaTeX HTML.
 * Uses throwOnError: false so malformed math renders gracefully without crashing the UI.
 */
export function renderKatex(tex: string, displayMode: boolean): string {
  const cleaned = stripUnsupportedTex(tex.trim());
  try {
    return katex.renderToString(cleaned, {
      displayMode,
      throwOnError: false
    });
  } catch (err: any) {
    const escaped = cleaned.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return `<span class="katex-error text-red-400 font-mono text-xs" title="${err?.message || 'Math rendering error'}">${escaped}</span>`;
  }
}

export function renderMarkdown(content: string, existingFilesMap: Record<string, FilePathInfo> = {}): string {
  if (!content) return '';
  try {
    const repaired = repairMarkdownFences(content);
    const preprocessed = preprocessMarkdownDiagrams(repaired);

    // 1. Temporarily protect fenced code blocks and inline codespans from math regex processing
    const codeBlocks: string[] = [];
    let text = preprocessed.replace(/(`{3,}|~{3,})[\s\S]*?\1/g, (match) => {
      codeBlocks.push(match);
      return `PREFENCEDCODEBLOCK${codeBlocks.length - 1}TOKEN`;
    });

    const inlineCodes: string[] = [];
    text = text.replace(/`+[^`\n]+`+/g, (match) => {
      inlineCodes.push(match);
      return `PREINLINECODESPAN${inlineCodes.length - 1}TOKEN`;
    });

    // 2. Extract and pre-render mathematical expressions
    interface MathPlaceholder {
      id: string;
      rendered: string;
      display: boolean;
    }
    const mathItems: MathPlaceholder[] = [];

    const addMath = (rawTex: string, display: boolean): string => {
      const rendered = renderKatex(rawTex, display);
      const id = display ? `KATEXDISP${mathItems.length}TOKEN` : `KATEXINL${mathItems.length}TOKEN`;
      mathItems.push({ id, rendered, display });
      return id;
    };

    // A. Display Math: $$ ... $$
    text = text.replace(/\$\$([\s\S]+?)\$\$/g, (_, tex) => addMath(tex, true));

    // B. Display Math: \[ ... \]
    text = text.replace(/\\\[([\s\S]+?)\\\]/g, (_, tex) => addMath(tex, true));

    // C. Display Math: Standalone LaTeX environments
    const envRegex = /\\begin\{(equation|align|gather|alignat|multline|matrix|pmatrix|bmatrix|vmatrix|Vmatrix|cases|array|split)\*?\}[\s\S]+?\\end\{\1\*?\}/g;
    text = text.replace(envRegex, (match) => addMath(match, true));

    // D. Inline Math: \( ... \)
    text = text.replace(/\\\(([\s\S]+?)\\\)/g, (_, tex) => addMath(tex, false));

    // E. Inline Math: $ ... $ (safe single dollar detection)
    text = text.replace(/(?<!\$)\$(?!\s)([^\$\n]+?)(?<!\s)\$(?!\$)/g, (match, tex) => {
      if (isLikelyCurrency(tex)) {
        return match;
      }
      return addMath(tex, false);
    });

    // 3. Restore protected code blocks and inline code spans before passing to marked
    text = text.replace(/PREINLINECODESPAN(\d+)TOKEN/g, (_, idx) => inlineCodes[Number(idx)]);
    text = text.replace(/PREFENCEDCODEBLOCK(\d+)TOKEN/g, (_, idx) => codeBlocks[Number(idx)]);

    // 4. Parse markdown with marked
    const renderer = createMarkdownRenderer(existingFilesMap);
    let html = marked.parse(text, {
      gfm: true,
      breaks: true,
      renderer
    }) as string;

    // 5. Replace math placeholders with rendered KaTeX HTML
    for (const item of mathItems) {
      if (item.display) {
        // If marked wrapped the placeholder inside a <p>...</p>, replace the whole paragraph to avoid invalid nesting
        const pRegex = new RegExp(`<p>\\s*${item.id}\\s*<\\/p>`, 'g');
        const displayWrapper = `<div class="katex-display-container my-3 overflow-x-auto text-center select-text">${item.rendered}</div>`;
        if (pRegex.test(html)) {
          html = html.replace(pRegex, displayWrapper);
        } else {
          html = html.replaceAll(item.id, displayWrapper);
        }
      } else {
        html = html.replaceAll(item.id, item.rendered);
      }
    }

    return html;
  } catch {
    return content;
  }
}
