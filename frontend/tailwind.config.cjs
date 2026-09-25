/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      fontFamily: {
        sans: [
          '"Anthropic Serif Text"',
          '"Anthropic Serif"',
          'Newsreader',
          'Source Serif 4',
          'Charter',
          'Iowan Old Style',
          'Palatino',
          'Georgia',
          'serif'
        ],
        serif: [
          '"Anthropic Serif Text"',
          '"Anthropic Serif"',
          'Newsreader',
          'Source Serif 4',
          'Charter',
          'Iowan Old Style',
          'Palatino',
          'Georgia',
          'serif'
        ],
        'serif-display': [
          '"Anthropic Serif Display"',
          '"Anthropic Serif"',
          'Newsreader',
          'Source Serif 4',
          'Charter',
          'Georgia',
          'serif'
        ],
        mono: [
          'Fira Code',
          'JetBrains Mono',
          'ui-monospace',
          'SFMono-Regular',
          'Menlo',
          'Monaco',
          'Consolas',
          'monospace'
        ]
      },
      colors: {
        ant: {
          primary: 'var(--ant-primary, #1677ff)',
          'primary-hover': 'var(--ant-primary-hover, #4096ff)',
          'primary-active': 'var(--ant-primary-active, #0958d9)',
          'primary-bg': 'var(--ant-primary-bg, #111a2c)',
          bg: 'var(--ant-bg, #0f1117)',
          'bg-secondary': 'var(--ant-bg-secondary, #181b26)',
          'bg-tertiary': 'var(--ant-bg-tertiary, #222634)',
          border: 'var(--ant-border, #2e3446)',
          'border-secondary': 'var(--ant-border-secondary, #1f2430)',
          text: 'var(--ant-text, #f3f4f6)',
          'text-secondary': 'var(--ant-text-secondary, #9ca3af)',
          'text-muted': 'var(--ant-text-muted, #6b7280)',
          success: 'var(--ant-success, #52c41a)',
          warning: 'var(--ant-warning, #faad14)',
          error: 'var(--ant-error, #ff4d4f)',
          info: 'var(--ant-info, #1677ff)'
        }
      }
    }
  },
  plugins: []
};
