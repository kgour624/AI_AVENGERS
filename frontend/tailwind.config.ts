import type { Config } from 'tailwindcss'

export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Surfaces — 2051 depth hierarchy
        'surface-void':         'oklch(11% 0.02 265)'   /* #0B0F19 deep obsidian */,
        'surface-base':         'oklch(13% 0.022 265)',
        'surface-raised':       'oklch(18% 0.025 265)'   /* #151C2C card */,
        'surface-overlay':      'oklch(22% 0.028 265)',
        'surface-float':        'oklch(26% 0.030 265)',
        'surface-border':       'oklch(34% 0.02 265)'    /* #2A3441 hairline */,
        'surface-panel-hover':  'oklch(24% 0.028 265)',

        // Text
        'text-primary':   'oklch(99% 0 0)'      /* pure white headings */,
        'text-secondary': 'oklch(78% 0.012 265)' /* #9CA3AF-ish body, readable */,
        'text-disabled':  'oklch(62% 0.010 265)',

        // Brand — Electric Violet
        'brand':       'oklch(64% 0.22 292)'   /* #8B5CF6 electric purple */,
        'brand-hover': 'oklch(70% 0.23 292)',
        'brand-dim':   'oklch(64% 0.22 292 / 0.18)',
        'accent':      'oklch(72% 0.13 220)'   /* #06B6D4 neon cyan */,
        'accent-dim':  'oklch(72% 0.13 220 / 0.18)',

        // Response modes (functional — hues locked)
        'mode-advise':        'oklch(68% 0.18 145)',
        'mode-advise-hover':  'oklch(60% 0.18 145)',
        'mode-ask':           'oklch(68% 0.18 220)',
        'mode-ask-hover':     'oklch(60% 0.18 220)',
        'mode-warn':          'oklch(75% 0.20 60)',
        'mode-warn-hover':    'oklch(67% 0.20 60)',
        'mode-pushback':      'oklch(72% 0.20 40)',
        'mode-pushback-hover':'oklch(64% 0.20 40)',
        'mode-refuse':        'oklch(62% 0.22 25)',
        'mode-refuse-hover':  'oklch(54% 0.22 25)',

        // Glow palette
        'glow-purple': 'oklch(68% 0.28 295)',
        'glow-violet': 'oklch(65% 0.30 320)',
        'glow-cyan':   'oklch(78% 0.18 200)',
        'glow-amber':  'oklch(80% 0.18 80)',
        'glow-pink':   'oklch(70% 0.25 340)',
        'glow-blue':   'oklch(72% 0.20 240)',

        // Glass
        'glass-border':      'oklch(100% 0 0 / 0.10)',
        'glass-border-glow': 'oklch(68% 0.28 295 / 0.35)',
      },

      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },

      borderRadius: {
        sm:   '6px',
        md:   '10px',
        lg:   '16px',
        xl:   '24px',
        full: '9999px',
      },

      boxShadow: {
        'neon-purple': '0 0 20px oklch(68% 0.28 295 / 0.5), 0 0 60px oklch(68% 0.28 295 / 0.2)',
        'neon-cyan':   '0 0 20px oklch(78% 0.18 200 / 0.5), 0 0 60px oklch(78% 0.18 200 / 0.2)',
        'neon-violet': '0 0 20px oklch(65% 0.30 320 / 0.5), 0 0 60px oklch(65% 0.30 320 / 0.2)',
        'card':        '0 4px 24px oklch(0% 0 0 / 0.5), 0 1px 4px oklch(0% 0 0 / 0.3)',
        'float':       '0 8px 40px oklch(0% 0 0 / 0.6), 0 2px 8px oklch(0% 0 0 / 0.4)',
        'glow-sm':     '0 0 8px oklch(68% 0.28 295 / 0.4)',
      },

      transitionTimingFunction: {
        arc:    'cubic-bezier(0.16, 1, 0.3, 1)',
        spring: 'cubic-bezier(0.34, 1.56, 0.64, 1)',
        sharp:  'cubic-bezier(0.4, 0, 0.2, 1)',
      },

      backdropBlur: {
        xs: '4px',
        sm: '8px',
        md: '16px',
        lg: '24px',
        xl: '32px',
      },

      animation: {
        'aurora-drift':   'aurora-drift 20s ease-in-out infinite alternate',
        'scan-sweep':     'scan-sweep 8s linear infinite',
        'neon-pulse':     'neon-pulse 3s ease-in-out infinite',
        'holo-shimmer':   'holo-shimmer 4s ease-in-out infinite',
        'data-stream':    'data-stream 2s linear infinite',
        'thinking':       'pulse-thinking 1.4s ease-in-out infinite',
      },

      keyframes: {
        'aurora-drift': {
          '0%':   { transform: 'translate(0%, 0%) scale(1)' },
          '33%':  { transform: 'translate(3%, -2%) scale(1.05)' },
          '66%':  { transform: 'translate(-2%, 3%) scale(0.97)' },
          '100%': { transform: 'translate(1%, -1%) scale(1.02)' },
        },
        'scan-sweep': {
          '0%':   { transform: 'translateY(-100%)' },
          '100%': { transform: 'translateY(100vh)' },
        },
        'neon-pulse': {
          '0%, 100%': { opacity: '0.8', filter: 'brightness(1)' },
          '50%':      { opacity: '1',   filter: 'brightness(1.3)' },
        },
        'holo-shimmer': {
          '0%':   { backgroundPosition: '0% 50%' },
          '50%':  { backgroundPosition: '100% 50%' },
          '100%': { backgroundPosition: '0% 50%' },
        },
        'data-stream': {
          '0%':   { transform: 'translateY(0)',    opacity: '1' },
          '100%': { transform: 'translateY(-20px)', opacity: '0' },
        },
        'pulse-thinking': {
          '0%, 100%': { opacity: '0.4' },
          '50%':      { opacity: '1' },
        },
      },
    },
  },
  plugins: [],
} satisfies Config
