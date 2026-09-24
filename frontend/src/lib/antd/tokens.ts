import { generate } from '@ant-design/colors';

/**
 * Ant Design 10-level base color scales for Dark Theme
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

/**
 * Ant Design Semantic Dark Theme Design Tokens
 */
export const antdTokens = {
  color: {
    primary: '#1677ff',
    primaryHover: '#4096ff',
    primaryActive: '#0958d9',
    primaryBg: '#111a2c',
    primaryBorder: '#15325b',

    success: '#52c41a',
    successHover: '#73d13d',
    successActive: '#389e0d',
    successBg: '#112117',
    successBorder: '#1e4620',

    warning: '#faad14',
    warningHover: '#ffc53d',
    warningActive: '#d48806',
    warningBg: '#2b2111',
    warningBorder: '#5b4515',

    error: '#ff4d4f',
    errorHover: '#ff7875',
    errorActive: '#d9363e',
    errorBg: '#2a1215',
    errorBorder: '#58181c',

    info: '#1677ff',
    infoBg: '#111a2c',

    bgLayout: '#0a0c10',
    bgContainer: '#0f1117',
    bgElevated: '#181b26',
    bgSpotlight: '#222634',
    bgHover: '#1f2433',

    border: '#2e3446',
    borderSecondary: '#1f2430',

    text: '#f3f4f6',
    textSecondary: '#9ca3af',
    textTertiary: '#6b7280',
    textQuaternary: '#4b5563',
    textDisabled: '#374151'
  },
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
