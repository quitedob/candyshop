import type { Config } from 'tailwindcss'

export default <Partial<Config>>{
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        'alex-ink': '#000000',
        'alex-burnt': '#944a00',
        'alex-secondary-container': '#fd933d',
        'alex-primary-container': '#1e1b19',
        'alex-surface': '#fdf8f7',
        'alex-surface-lowest': '#ffffff',
        'alex-outline': '#7e7570',
        'alex-outline-variant': '#d0c4be',
        'alex-on-surface': '#1c1b1b',
        'alex-on-surface-variant': '#4d4540',
      },
      borderRadius: {
        'alex-sm': '0.125rem',
        'alex-md': '0.25rem',
        'alex-lg': '0.5rem',
        'alex-full': '0.75rem',
      },
      fontFamily: {
        display: ['Noto Serif', 'Noto Serif SC', 'Georgia', 'serif'],
        body: ['Work Sans', 'Noto Sans SC', 'system-ui', 'sans-serif'],
      },
      spacing: {
        'margin-mobile': '16px',
        'gutter': '24px',
        'unit': '4px',
        'margin-desktop': '64px',
        'admin-gap': '12px',
      },
    },
  },
}
