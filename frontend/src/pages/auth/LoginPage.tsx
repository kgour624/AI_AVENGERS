import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { motion, useReducedMotion } from 'framer-motion'
import { adminLogin, login } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import { handleAPIError } from '@/utils/errors'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ExpertAvatar } from '@/components/expert/ExpertAvatar'
import { ARC_MOTION } from '@/design-system/motion'

const CONSTELLATION_DOMAINS = ['system_design', 'database', 'security', 'architecture']

export default function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const setAuth = useAuthStore((s) => s.setAuth)
  const reduceMotion = useReducedMotion()

  // Show a one-time success banner when redirected from ResetPasswordPage.
  // WHY location.state: avoids a query param that would stay in the URL
  // and re-show the banner on every reload.
  const resetSuccess = (location.state as { resetSuccess?: boolean } | null)?.resetSuccess ?? false

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [adminMode, setAdminMode] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)
    const trimmedEmail = email.trim()
    if (!trimmedEmail || !password) { setError('Email and password are required.'); return }
    setIsSubmitting(true)
    try {
      if (adminMode) {
        const { user, tokenPair } = await adminLogin({
          email: trimmedEmail,
          password,
          totpCode: totpCode.trim(),
        })
        setAuth(user, tokenPair.accessToken)
        navigate('/admin', { replace: true })
      } else {
        const { user, tokenPair } = await login({ email: trimmedEmail, password })
        setAuth(user, tokenPair.accessToken)
        navigate(user.role === 'admin' ? '/admin' : '/', { replace: true })
      }
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="relative flex h-screen items-center justify-center overflow-hidden bg-surface-void">

      {/* Aurora orbs — decorative, aria-hidden */}
      <div aria-hidden="true"
        className="aurora-orb pointer-events-none absolute -left-1/3 -top-1/3 h-[80vh] w-[80vh] rounded-full"
        style={{ background: 'radial-gradient(circle, oklch(68% 0.28 295 / 0.12) 0%, transparent 70%)' }}
      />
      <div aria-hidden="true"
        className="aurora-orb-slow pointer-events-none absolute -bottom-1/3 -right-1/3 h-[70vh] w-[70vh] rounded-full"
        style={{ background: 'radial-gradient(circle, oklch(78% 0.18 200 / 0.10) 0%, transparent 70%)' }}
      />
      <div aria-hidden="true"
        className="pointer-events-none absolute left-1/2 top-1/2 h-[40vh] w-[40vh] -translate-x-1/2 -translate-y-1/2 rounded-full"
        style={{ background: 'radial-gradient(circle, oklch(70% 0.25 340 / 0.05) 0%, transparent 70%)' }}
      />

      {/* Scan-line */}
      <div aria-hidden="true" className="scan-line" />

      {/* Expert constellation */}
      <div aria-hidden="true" className="pointer-events-none absolute inset-0">
        {CONSTELLATION_DOMAINS.map((domain, i) => (
          <div key={domain} className="absolute opacity-30"
            style={{
              top: `${12 + i * 22}%`,
              left: i % 2 === 0 ? '6%' : undefined,
              right: i % 2 === 1 ? '6%' : undefined,
              animation: `pulse-thinking ${4 + i * 0.8}s ease-in-out infinite`,
            }}
          >
            <ExpertAvatar domain={domain} status="idle" size="lg" />
          </div>
        ))}
      </div>

      {/* Login card */}
      <motion.div
        initial={reduceMotion ? undefined : { opacity: 0, scale: 0.94, y: 12 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        transition={{ duration: ARC_MOTION.panel, ease: ARC_MOTION.ease }}
        className="relative w-full max-w-sm"
      >
        {/* Holographic border glow */}
        <div aria-hidden="true"
          className="pointer-events-none absolute -inset-px rounded-xl"
          style={{
            background: 'linear-gradient(135deg, oklch(68% 0.28 295 / 0.4), oklch(78% 0.18 200 / 0.3), oklch(65% 0.30 320 / 0.4))',
            borderRadius: 'inherit',
          }}
        />

        <div className="relative rounded-xl border border-glass-border bg-surface-raised/85 p-8 backdrop-blur-2xl shadow-float">
          {/* Brand */}
          <div className="mb-6 text-center">
            <h1 className="holo-text mb-1 text-3xl font-bold tracking-[0.12em] uppercase">
              {'\u26A1'} AI Avengers
            </h1>
            <p className="text-xs font-medium uppercase tracking-[0.2em] text-text-disabled">
              Neural Command Center
            </p>
            <div className="mx-auto mt-3 h-px w-24"
              style={{ background: 'linear-gradient(90deg, transparent, oklch(68% 0.28 295 / 0.6), transparent)' }}
            />
          </div>

          {/* Password-reset success banner */}
          {resetSuccess && (
            <div className="mb-4 rounded-md border border-mode-advise/30 bg-mode-advise/10 px-3 py-2 text-center">
              <p className="text-xs text-mode-advise">
                {'\u2705'} Password reset successful. Please sign in.
              </p>
            </div>
          )}

          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <Input
              label="Email"
              type="email" name="email" autoComplete="email"
              placeholder="you@company.com"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isSubmitting} required
            />
            <Input
              label="Password"
              type="password" name="password" autoComplete="current-password"
              placeholder="{'\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022'}"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isSubmitting} required
            />

            {adminMode && (
              <Input
                label="Authenticator code"
                name="totp"
                autoComplete="one-time-code"
                inputMode="numeric"
                placeholder="6-digit code"
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
                disabled={isSubmitting}
              />
            )}

            <label className="flex items-center gap-2 text-xs text-text-secondary">
              <input
                type="checkbox"
                checked={adminMode}
                onChange={(e) => setAdminMode(e.target.checked)}
                disabled={isSubmitting}
              />
              Admin sign-in (password + authenticator)
            </label>

            {error && (
              <p className="rounded-md border border-mode-refuse/30 bg-mode-refuse/10 px-3 py-2 text-xs text-mode-refuse">
                {error}
              </p>
            )}

            <Button type="submit" isLoading={isSubmitting} className="mt-1 w-full">
              Sign In
            </Button>
          </form>

          {/* Forgot Password */}
          <p className="mt-4 text-center text-xs text-text-disabled">
            <Link
              to="/forgot-password"
              className="text-text-disabled transition-colors hover:text-text-secondary"
            >
              Forgot Password?
            </Link>
          </p>

          <p className="mt-3 text-center text-xs text-text-disabled">
            No account?{' '}
            <Link to="/register" className="text-glow-purple/80 transition-colors hover:text-glow-purple">
              Register
            </Link>
          </p>
        </div>
      </motion.div>
    </div>
  )
}
