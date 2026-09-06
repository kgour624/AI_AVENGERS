import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion, useReducedMotion } from 'framer-motion'
import { login } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import { handleAPIError } from '@/utils/errors'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ExpertAvatar } from '@/components/expert/ExpertAvatar'
import { ARC_MOTION } from '@/design-system/motion'

/**
 * ARC-51 decorative expert constellation - purely presentational,
 * aria-hidden, NO API call (see docs/ARC51_UI_CONTRACT.md §6: /experts
 * requires a JWT this unauthenticated page doesn't have - fetching a
 * real list here would be a scope violation, not a nice-to-have).
 * Domain strings are real ones this product actually has (system
 * design/database/security/architecture), not invented placeholders.
 */
const CONSTELLATION_DOMAINS = ['system_design', 'database', 'security', 'architecture']

/**
 * Source: FRONTEND_SYSTEM_DESIGN.md section 10 wireframe ("Login Page").
 *
 * Cross-questioned before writing:
 * - What if login succeeds but the user navigates away before the
 *   promise resolves (e.g. double-clicked back button)? navigate()
 *   after an unmount is a no-op in React Router, not a crash - safe.
 * - What if the email field has leading/trailing whitespace? Trimmed
 *   before submit - a pasted email with a trailing newline is a real,
 *   common failure mode, not a hypothetical.
 * - What happens to the password field's value if login fails? Left
 *   as-is (NOT cleared) - clearing it on every failed attempt would
 *   punish a user who made a one-character typo by making them retype
 *   the whole password. Email is also left as-is for the same reason.
 */
export default function LoginPage() {
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)
  const reduceMotion = useReducedMotion()

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)

    const trimmedEmail = email.trim()
    if (!trimmedEmail || !password) {
      setError('Email and password are required.')
      return
    }

    setIsSubmitting(true)
    try {
      const { user, tokenPair } = await login({ email: trimmedEmail, password })
      setAuth(user, tokenPair.accessToken)
      navigate('/', { replace: true })
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="relative flex h-screen items-center justify-center overflow-hidden bg-surface-void">
      {/* Ambient drifting glow - CSS keyframes, NOT mouse-tracked, per
          contract §6 (avoids input-lag risk on low-end hardware). */}
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -left-1/4 -top-1/4 h-[70vh] w-[70vh] rounded-full bg-glow-purple/10 blur-3xl"
        style={{ animation: 'arc-ring-rotate 40s linear infinite' }}
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -bottom-1/4 -right-1/4 h-[60vh] w-[60vh] rounded-full bg-glow-cyan/10 blur-3xl"
        style={{ animation: 'arc-ring-rotate 55s linear infinite reverse' }}
      />

      {/* Decorative expert constellation - see header comment. */}
      <div aria-hidden="true" className="pointer-events-none absolute inset-0">
        {CONSTELLATION_DOMAINS.map((domain, i) => (
          <div
            key={domain}
            className="absolute opacity-40"
            style={{
              top: `${15 + i * 20}%`,
              left: i % 2 === 0 ? '8%' : undefined,
              right: i % 2 === 1 ? '8%' : undefined,
              animation: `pulse-thinking ${4 + i}s ease-in-out infinite`,
            }}
          >
            <ExpertAvatar domain={domain} status="idle" size="lg" />
          </div>
        ))}
      </div>

      <motion.div
        initial={reduceMotion ? undefined : { opacity: 0, scale: 0.96, y: 8 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        transition={{ duration: ARC_MOTION.panel, ease: ARC_MOTION.ease }}
        className="relative w-full max-w-sm rounded-lg border border-glass-border bg-surface-raised/80 p-8 backdrop-blur-xl"
      >
        <h1 className="mb-1 text-center text-2xl font-semibold tracking-wide text-text-primary [text-shadow:0_0_20px_var(--glow-purple)]">
          {'\u26A1'} AI AVENGERS
        </h1>
        <p className="mb-6 text-center text-sm text-text-secondary">
          Neural Command Center {'\u2014'} Multi-Agent Domain Expert Simulator
        </p>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <Input
            label="Email"
            type="email"
            name="email"
            autoComplete="email"
            placeholder="you@company.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={isSubmitting}
            required
          />
          <Input
            label="Password"
            type="password"
            name="password"
            autoComplete="current-password"
            placeholder="\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={isSubmitting}
            required
          />

          {error && <p className="text-sm text-mode-refuse">{error}</p>}

          <Button type="submit" isLoading={isSubmitting} className="mt-2 w-full">
            Sign In
          </Button>
        </form>

        <p className="mt-6 text-center text-sm text-text-secondary">
          Don't have an account?{' '}
          <Link to="/register" className="text-brand hover:text-brand-hover">
            Register
          </Link>
        </p>
      </motion.div>
    </div>
  )
}
