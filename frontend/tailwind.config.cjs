/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        ant: {
          primary: '#1677ff',
          'primary-hover': '#4096ff',
          'primary-active': '#0958d9',
          'primary-bg': '#111a2c',
          bg: '#0f1117',
          'bg-secondary': '#181b26',
          'bg-tertiary': '#222634',
          border: '#2e3446',
          'border-secondary': '#1f2430',
          text: '#f3f4f6',
          'text-secondary': '#9ca3af',
          'text-muted': '#6b7280',
          success: '#52c41a',
          warning: '#faad14',
          error: '#ff4d4f',
          info: '#1677ff'
        }
      }
    }
  },
  plugins: []
};
