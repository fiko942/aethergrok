import { describe, it, expect } from 'vitest';
import {
  renderMarkdown,
  stripUnsupportedTex,
  isLikelyCurrency,
  renderKatex
} from './markdownRenderer';

describe('markdownRenderer - LaTeX Math Formatting', () => {
  describe('stripUnsupportedTex', () => {
    it('strips \\label{...} commands', () => {
      const input = '\\begin{align}\\label{eq:einstein} E &= mc^2 \\end{align}';
      const output = stripUnsupportedTex(input);
      expect(output).not.toContain('\\label');
      expect(output).toContain('E &= mc^2');
    });

    it('returns empty string on empty input', () => {
      expect(stripUnsupportedTex('')).toBe('');
    });
  });

  describe('isLikelyCurrency', () => {
    it('identifies bare prices as currency', () => {
      expect(isLikelyCurrency('5')).toBe(true);
      expect(isLikelyCurrency('10.50')).toBe(true);
      expect(isLikelyCurrency('100k')).toBe(true);
      expect(isLikelyCurrency('2.5m')).toBe(true);
      expect(isLikelyCurrency('50 million')).toBe(true);
      expect(isLikelyCurrency('5 - 10')).toBe(true);
      expect(isLikelyCurrency('10-20')).toBe(true);
    });

    it('identifies mathematical expressions as not currency', () => {
      expect(isLikelyCurrency('x = 5')).toBe(false);
      expect(isLikelyCurrency('1 + 1 = 2')).toBe(false);
      expect(isLikelyCurrency('x^2 + y^2')).toBe(false);
      expect(isLikelyCurrency('\\alpha + \\beta')).toBe(false);
      expect(isLikelyCurrency('x_1')).toBe(false);
    });
  });

  describe('renderKatex', () => {
    it('renders inline math to KaTeX HTML without error', () => {
      const html = renderKatex('x^2 + y^2 = z^2', false);
      expect(html).toContain('class="katex"');
      expect(html).toContain('<math');
    });

    it('renders display math to KaTeX HTML', () => {
      const html = renderKatex('E = mc^2', true);
      expect(html).toContain('class="katex-display"');
    });

    it('handles malformed LaTeX gracefully with throwOnError: false', () => {
      const html = renderKatex('\\nonExistentLatexCommand{123}', false);
      expect(html).toContain('color:#cc0000');
      expect(html).toContain('\\nonExistentLatexCommand');
      expect(typeof html).toBe('string');
    });
  });

  describe('renderMarkdown - end-to-end math integration', () => {
    it('renders inline \\(...\\) formulas', () => {
      const markdown = 'The area is \\(\\pi r^2\\) exactly.';
      const result = renderMarkdown(markdown);
      expect(result).toContain('class="katex"');
      expect(result).not.toContain('\\(');
      expect(result).not.toContain('\\)');
    });

    it('renders block \\[...\\] display formulas in a scrollable display container', () => {
      const markdown = 'Formula:\n\\[E = mc^2\\]\nDone.';
      const result = renderMarkdown(markdown);
      expect(result).toContain('katex-display-container');
      expect(result).toContain('katex-display');
      expect(result).not.toContain('\\[');
      expect(result).not.toContain('\\]');
    });

    it('renders $$...$$ block formulas across multiple lines', () => {
      const markdown = 'Pythagoras:\n$$\na^2 + b^2 = c^2\n$$\nEnd.';
      const result = renderMarkdown(markdown);
      expect(result).toContain('katex-display-container');
      expect(result).toContain('katex-display');
      expect(result).not.toContain('$$');
    });

    it('renders standalone \\begin{matrix}...\\end{matrix} and environments', () => {
      const markdown = `
\\begin{pmatrix}
1 & 2 \\\\
3 & 4
\\end{pmatrix}
`;
      const result = renderMarkdown(markdown);
      expect(result).toContain('katex-display-container');
      expect(result).toContain('katex-display');
    });

    it('renders single dollar $...$ inline math while ignoring currency', () => {
      const markdown = 'Given $f(x) = x^2 + 1$, the price is $5 and then $10.';
      const result = renderMarkdown(markdown);
      // Math should be formatted
      expect(result).toContain('class="katex"');
      // Currency should remain plain text
      expect(result).toContain('$5 and then $10');
    });

    it('leaves code blocks and inline code spans with dollar signs untouched', () => {
      const markdown = `
Here is a bash script:
\`\`\`bash
echo "$MY_VAR"
price=$50
\`\`\`
And inline variable \`$USER\`.
Also math $x > 0$.
`;
      const result = renderMarkdown(markdown);
      expect(result).toContain('price=$50');
      expect(result).toContain('$USER');
      expect(result).toContain('class="katex"');
    });

    it('preserves underscores inside LaTeX so they do not become markdown italics', () => {
      const markdown = 'Subscripts: $x_1 + x_2 = x_{total}$.';
      const result = renderMarkdown(markdown);
      expect(result).not.toContain('<em>');
      expect(result).toContain('class="katex"');
    });

    it('preserves asterisks inside LaTeX so they do not become markdown bold/emphasis', () => {
      const markdown = 'Convolution: $f * g$.';
      const result = renderMarkdown(markdown);
      expect(result).not.toContain('<em>');
      expect(result).toContain('class="katex"');
    });
  });
});
