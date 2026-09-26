import Prism from 'prismjs';
import 'prismjs/components/prism-typescript.js';
import 'prismjs/components/prism-javascript.js';
import 'prismjs/components/prism-jsx.js';
import 'prismjs/components/prism-tsx.js';
import 'prismjs/components/prism-go.js';
import 'prismjs/components/prism-python.js';
import 'prismjs/components/prism-bash.js';
import 'prismjs/components/prism-json.js';
import 'prismjs/components/prism-yaml.js';
import 'prismjs/components/prism-markdown.js';
import 'prismjs/components/prism-css.js';
import 'prismjs/components/prism-sql.js';
import 'prismjs/components/prism-rust.js';
import 'prismjs/components/prism-diff.js';
import 'prismjs/components/prism-c.js';
import 'prismjs/components/prism-cpp.js';
import 'prismjs/components/prism-csharp.js';
import 'prismjs/components/prism-java.js';
import 'prismjs/components/prism-kotlin.js';
import 'prismjs/components/prism-swift.js';
import 'prismjs/components/prism-ruby.js';
import 'prismjs/components/prism-php.js';
import 'prismjs/components/prism-toml.js';
import 'prismjs/components/prism-docker.js';
import 'prismjs/components/prism-graphql.js';

export function highlightCode(code: string, language?: string): string {
  if (!code) return '';
  const lang = (language || '').toLowerCase().trim();

  // Handle diagram / ascii flow blocks cleanly
  if (lang === 'diagram' || lang === 'ascii' || lang === 'tree' || lang === 'txt' || lang === 'text') {
    return escapeHtml(code);
  }

  // Map common language aliases
  const langMap: Record<string, string> = {
    js: 'javascript',
    cjs: 'javascript',
    mjs: 'javascript',
    ts: 'typescript',
    cts: 'typescript',
    mts: 'typescript',
    jsx: 'jsx',
    tsx: 'tsx',
    sh: 'bash',
    shell: 'bash',
    zsh: 'bash',
    py: 'python',
    python3: 'python',
    golang: 'go',
    rs: 'rust',
    yml: 'yaml',
    rb: 'ruby',
    cs: 'csharp',
    kt: 'kotlin',
    dockerfile: 'docker',
    gql: 'graphql',
    md: 'markdown'
  };

  const targetLang = langMap[lang] || lang;

  if (targetLang && Prism.languages[targetLang]) {
    try {
      return Prism.highlight(code, Prism.languages[targetLang], targetLang);
    } catch {
      return escapeHtml(code);
    }
  }

  return escapeHtml(code);
}

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}
