import type { Config } from 'tailwindcss'

export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        // Surfaces — 2051 depth hierarchy
        'surface-void':         'oklch(6%  0.015 285)',
        'surface-base':         'oklch(9%  0.018 285)',
        'surface-raised':       'oklch(13% 0.020 285)',
        'surface-overlay':      'oklch(17% 0.022 285)',
        'surface-float':        'oklch(21% 0.024 285)',
        'surface-border':       'oklch(28% 0.025 285)',
        'surface-panel-hover':  'oklch(18% 0.022 285)',

        // Text
        'text-primary':   'oklch(96% 0.008 285)',
        'text-secondary': 'oklch(62% 0.012 285)',
        'text-disabled':  'oklch(38% 0.008 285)',

        // Brand — Electric Violet
        'brand':       'oklch(68% 0.28 295)',
        'brand-hover': 'oklch(73% 0.28 295)',
        'brand-dim':   'oklch(68% 0.28 295 / 0.15)',

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
