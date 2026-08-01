import { defineConfig, presetWind4 } from 'unocss'

// Every value points at a CSS variable from static/tokens.css - the contract.
// Colour names follow shadcn/ui exactly, so a theme written for shadcn themes
// this too. Structure (radius, control padding, elevation) is tokenised as well,
// which is what lets a whole design system swap by replacing one stylesheet.
export default defineConfig({
  presets: [presetWind4()],
  theme: {
    colors: {
      background: 'var(--background)',
      foreground: 'var(--foreground)',
      'site-foreground': 'var(--site-foreground)',
      'site-foreground-muted': 'var(--site-foreground-muted)',
      card: {
        DEFAULT: 'var(--card)',
        foreground: 'var(--card-foreground)',
      },
      popover: {
        DEFAULT: 'var(--popover)',
        foreground: 'var(--popover-foreground)',
      },
      primary: {
        DEFAULT: 'var(--primary)',
        foreground: 'var(--primary-foreground)',
      },
      secondary: {
        DEFAULT: 'var(--secondary)',
        foreground: 'var(--secondary-foreground)',
      },
      muted: {
        DEFAULT: 'var(--muted)',
        foreground: 'var(--muted-foreground)',
      },
      accent: {
        DEFAULT: 'var(--accent)',
        foreground: 'var(--accent-foreground)',
      },
      destructive: {
        DEFAULT: 'var(--destructive)',
        foreground: 'var(--destructive-foreground)',
      },
      border: 'var(--border)',
      input: 'var(--input)',
      ring: 'var(--ring)',
      code: {
        DEFAULT: 'var(--code)',
        foreground: 'var(--code-foreground)',
      },
    },
    // Derived from a single --radius, so a design system sets one number.
    radius: {
      xs: 'var(--r-xs)',
      sm: 'var(--r-sm)',
      md: 'var(--r-md)',
      lg: 'var(--r-lg)',
      xl: 'var(--r-xl)',
      pill: 'var(--r-pill)',
    },
    // Density. Controls pad with px-control-x / py-control-y and nothing else,
    // so one token pair resizes every control in the set.
    spacing: {
      'control-x': 'var(--control-px)',
      'control-y': 'var(--control-py)',
      'row-x': 'var(--row-px)',
      'row-y': 'var(--row-py)',
      card: 'var(--pad-card)',
    },
    shadow: {
      e1: 'var(--shadow-1)',
      e2: 'var(--shadow-2)',
    },
    font: {
      display: 'var(--type-display)',
      sans: 'var(--type-sans)',
      mono: 'var(--type-mono)',
    },
    // Wind4 wants objects with camelCase keys here - the Wind3 [size, {...}]
    // tuple and kebab-case keys both fail silently, emitting a utility that
    // references an undefined variable.
    text: {
      'display-xl': { fontSize: '64px', lineHeight: '1.05', letterSpacing: '-1.5px' },
      'display-lg': { fontSize: '48px', lineHeight: '1.1', letterSpacing: '-1px' },
      'display-md': { fontSize: '36px', lineHeight: '1.15', letterSpacing: '-0.5px' },
      'display-sm': { fontSize: '28px', lineHeight: '1.2', letterSpacing: '-0.3px' },
      'title-lg': { fontSize: '22px', lineHeight: '1.3' },
      'title-md': { fontSize: '18px', lineHeight: '1.4' },
      'title-sm': { fontSize: '16px', lineHeight: '1.4' },
      'body-md': { fontSize: '16px', lineHeight: '1.55' },
      'body-sm': { fontSize: '14px', lineHeight: '1.55' },
      caption: { fontSize: '13px', lineHeight: '1.4' },
      'caption-up': { fontSize: '12px', lineHeight: '1.4', letterSpacing: '1.5px' },
      mono: { fontSize: '14px', lineHeight: '1.6' },
      button: { fontSize: '14px', lineHeight: '1' },
    },
  },
})
