import type { Config } from 'tailwindcss'

// Maps OKLCH tokens from design-system/tokens.css to Tailwind utility classes.
// WHY duplicate values here instead of referencing CSS vars: Tailwind's JIT
// compiler needs literal color values at build time to generate utilities
// (e.g. bg-mode-advise/10) - CSS var opacity modifiers are not supported
// the same way. Source of truth for the *values* remains tokens.css; if you
// change a color there, mirror it here. (FRONTEND_SYSTEM_DESIGN.md section 4)
export default {
  darkMode: 'class',
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        'surface-base': 'oklch(12% 0.01 270)',
        'surface-raised': 'oklch(16% 0.01 270)',
        'surface-overlay': 'oklch(20% 0.01 270)',
        'surface-border': 'oklch(25% 0.01 270)',
        'text-primary': 'oklch(95% 0.01 270)',
        'text-secondary': 'oklch(65% 0.01 270)',
        'text-disabled': 'oklch(40% 0.01 270)',
        brand: 'oklch(65% 0.20 270)',
        'mode-advise': 'oklch(65% 0.15 145)',
        'mode-ask': 'oklch(65% 0.15 220)',
        'mode-warn': 'oklch(70% 0.18 60)',
        'mode-pushback': 'oklch(68% 0.18 40)',
        'mode-refuse': 'oklch(60% 0.20 25)',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
      },
      borderRadius: {
        sm: '4px',
        md: '8px',
        lg: '12px',
      },
    },
  },
  plugins: [],
} satisfies Config
