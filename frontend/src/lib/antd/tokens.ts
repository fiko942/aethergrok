import { generate } from '@ant-design/colors';

/**
 * Ant Design 10-level base color scales
 */
export const baseColors = {
  blue: '#1677ff',
  green: '#52c41a',
  gold: '#faad14',
  red: '#ff4d4f',
  cyan: '#13c2c2',
  purple: '#722ed1',
  magenta: '#eb2f96'
} as const;

export const darkPalettes = {
  blue: generate(baseColors.blue, { theme: 'dark', backgroundColor: '#0f1117' }),
  green: generate(baseColors.green, { theme: 'dark', backgroundColor: '#0f1117' }),
  gold: generate(baseColors.gold, { theme: 'dark', backgroundColor: '#0f1117' }),
  red: generate(baseColors.red, { theme: 'dark', backgroundColor: '#0f1117' }),
  cyan: generate(baseColors.cyan, { theme: 'dark', backgroundColor: '#0f1117' }),
  purple: generate(baseColors.purple, { theme: 'dark', backgroundColor: '#0f1117' })
};

export const themeDefinitions = {
  'dark-studio': {
    name: 'Dark Studio',
    description: 'Authentic Ant Design Dark palette with dark blue accents',
    colors: {
      primary: '#177ddc',
      primaryHover: '#4096ff',
      primaryActive: '#0958d9',
      primaryBg: '#111a2c',
      primaryBorder: '#15325b',
      bgLayout: '#000000',
      bgContainer: '#141414',
      bgElevated: '#1f1f1f',
      bgSpotlight: '#262626',
      bgHover: '#262626',
      border: 'rgba(255, 255, 255, 0.08)',
      borderSecondary: 'rgba(255, 255, 255, 0.04)',
      text: 'rgba(255, 255, 255, 0.85)',
      textSecondary: 'rgba(255, 255, 255, 0.45)',
      textTertiary: 'rgba(255, 255, 255, 0.30)',
      textQuaternary: 'rgba(255, 255, 255, 0.20)',
      textDisabled: 'rgba(255, 255, 255, 0.15)',
      success: '#52c41a',
      warning: '#faad14',
      error: '#ff4d4f',
      info: '#177ddc'
    }
  },
  'dark-high-contrast': {
    name: 'High Contrast Dark',
    description: 'Pitch black background with stark borders and vivid colors',
    colors: {
      primary: '#388bfd',
      primaryHover: '#58a6ff',
      primaryActive: '#1f6feb',
      primaryBg: '#0d1d30',
      primaryBorder: '#388bfd',
      bgLayout: '#000000',
      bgContainer: '#050505',
      bgElevated: '#0f0f10',
      bgSpotlight: '#1c1c1f',
      bgHover: '#1a1a1e',
      border: 'rgba(255, 255, 255, 0.14)',
      borderSecondary: 'rgba(255, 255, 255, 0.08)',
      text: '#ffffff',
      textSecondary: '#c0c0c8',
      textTertiary: '#90909a',
      textQuaternary: '#707078',
      textDisabled: '#4e4e56',
      success: '#3fb950',
      warning: '#d29922',
      error: '#f85149',
      info: '#58a6ff'
    }
  },
  'light-antd': {
    name: 'Light Ant Design',
    description: 'Clean official Ant Design light theme aesthetic',
    colors: {
      primary: '#1677ff',
      primaryHover: '#4096ff',
      primaryActive: '#0958d9',
      primaryBg: '#e6f4ff',
      primaryBorder: '#91caff',
      bgLayout: '#f5f5f5',
      bgContainer: '#ffffff',
      bgElevated: '#fcfcfd',
      bgSpotlight: '#e8edf3',
      bgHover: '#f0f2f5',
      border: '#d9d9d9',
      borderSecondary: '#f0f0f0',
      text: '#1f1f1f',
      textSecondary: '#595959',
      textTertiary: '#8c8c8c',
      textQuaternary: '#bfbfbf',
      textDisabled: '#d9d9d9',
      success: '#52c41a',
      warning: '#faad14',
      error: '#ff4d4f',
      info: '#1677ff'
    }
  }
} as const;

/**
 * Ant Design Semantic Dark Theme Design Tokens
 */
export const antdTokens = {
  color: themeDefinitions['dark-studio'].colors,
  fontSize: {
    xs: '11px',
    sm: '12px',
    base: '13px',
    md: '14px',
    lg: '16px',
    xl: '20px'
  },
  borderRadius: {
    sm: '4px',
    base: '6px',
    md: '8px',
    lg: '12px',
    full: '9999px'
  },
  shadow: {
    card: '0 1px 3px 0 rgba(0, 0, 0, 0.4), 0 1px 2px -1px rgba(0, 0, 0, 0.4)',
    elevated: '0 10px 15px -3px rgba(0, 0, 0, 0.6), 0 4px 6px -4px rgba(0, 0, 0, 0.6)',
    glowPrimary: '0 0 16px rgba(22, 119, 255, 0.35)'
  }
} as const;

export type AntdTokens = typeof antdTokens;
