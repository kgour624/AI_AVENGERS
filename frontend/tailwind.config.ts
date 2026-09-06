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
        'brand-hover': 'oklch(from oklch(65% 0.20 270) calc(l - 8%) c h)',
        'mode-advise': 'oklch(65% 0.15 145)',
        'mode-advise-hover': 'oklch(from oklch(65% 0.15 145) calc(l - 8%) c h)',
        'mode-ask': 'oklch(65% 0.15 220)',
        'mode-ask-hover': 'oklch(from oklch(65% 0.15 220) calc(l - 8%) c h)',
        'mode-warn': 'oklch(70% 0.18 60)',
        'mode-warn-hover': 'oklch(from oklch(70% 0.18 60) calc(l - 8%) c h)',
        'mode-pushback': 'oklch(68% 0.18 40)',
        'mode-pushback-hover': 'oklch(from oklch(68% 0.18 40) calc(l - 8%) c h)',
        'mode-refuse': 'oklch(60% 0.20 25)',
        'mode-refuse-hover': 'oklch(from oklch(60% 0.20 25) calc(l - 8%) c h)',
        // ARC-51 additive tokens - see docs/ARC51_UI_CONTRACT.md §2.
        // WHY literal oklch() here too, mirroring tokens.css: same JIT
        // constraint documented in this file's own header comment.
        'surface-void': 'oklch(9% 0.012 270)',
        'surface-panel-hover': 'oklch(18% 0.015 270)',
        'glow-cyan': 'oklch(75% 0.15 200)',
        'glow-purple': 'oklch(65% 0.20 300)',
        'glass-border': 'oklch(100% 0 0 / 0.08)',
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
