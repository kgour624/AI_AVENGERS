import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { motion, useReducedMotion } from 'framer-motion'
import { resetPassword } from '@/api/auth'
import { handleAPIError } from '@/utils/errors'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ARC_MOTION } from '@/design-system/motion'

export default function ResetPasswordPage() {
  const navigate = useNavigate()
  const reduceMotion = useReducedMotion()
  const [searchParams] = useSearchParams()

  // Token comes from the reset link: /reset-password?token=<rawToken>
  const token = searchParams.get('token') ?? ''

  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(e: FormEvent) {
    e.preventDefault()
    setError(null)

    if (!token) {
      setError('Reset token is missing. Please use the link from your email.')
      return
    }
    if (newPassword.length < 8) {
      setError('Password must be at least 8 characters.')
      return
    }
    if (newPassword !== confirmPassword) {
      setError('Passwords do not match.')
      return
    }

    setIsSubmitting(true)
    try {
      await resetPassword(token, newPassword)
      // Redirect to login with a success flash.
      // WHY location state: LoginPage can read it and show a one-time banner
      // without a separate query param that would stay in the URL.
      navigate('/login', {
        replace: true,
        state: { resetSuccess: true },
      })
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="relative flex h-screen items-center justify-center overflow-hidden bg-surface-void">

      {/* Aurora orbs — same as LoginPage */}
      <div aria-hidden="true"
        className="aurora-orb pointer-events-none absolute -left-1/3 -top-1/3 h-[80vh] w-[80vh] rounded-full"
        style={{ background: 'radial-gradient(circle, oklch(68% 0.28 295 / 0.12) 0%, transparent 70%)' }}
      />
      <div aria-hidden="true"
        className="aurora-orb-slow pointer-events-none absolute -bottom-1/3 -right-1/3 h-[70vh] w-[70vh] rounded-full"
        style={{ background: 'radial-gradient(circle, oklch(78% 0.18 200 / 0.10) 0%, transparent 70%)' }}
      />
      <div aria-hidden="true" className="scan-line" />

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
          <div className="mb-6 text-center">
            <h1 className="holo-text mb-1 text-2xl font-bold tracking-[0.1em] uppercase">
              New Password
            </h1>
            <p className="text-xs font-medium uppercase tracking-[0.2em] text-text-disabled">
              AI Avengers
            </p>
            <div className="mx-auto mt-3 h-px w-24"
              style={{ background: 'linear-gradient(90deg, transparent, oklch(68% 0.28 295 / 0.6), transparent)' }}
            />
          </div>

          {!token ? (
            <div className="rounded-md border border-mode-refuse/30 bg-mode-refuse/10 px-4 py-3 text-center">
              <p className="text-sm font-medium text-mode-refuse">Invalid reset link</p>
              <p className="mt-1 text-xs text-text-secondary">
                The link is missing a token. Please request a new one.
              </p>
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="flex flex-col gap-4">
              <Input
                label="New Password"
                type="password"
                name="new_password"
                autoComplete="new-password"
                placeholder="At least 8 characters"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
                disabled={isSubmitting}
                required
              />
              <Input
                label="Confirm Password"
                type="password"
                name="confirm_password"
                autoComplete="new-password"
                placeholder="Repeat your new password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                disabled={isSubmitting}
                required
              />

              {error && (
                <p className="rounded-md border border-mode-refuse/30 bg-mode-refuse/10 px-3 py-2 text-xs text-mode-refuse">
                  {error}
                </p>
              )}

              <Button type="submit" isLoading={isSubmitting} className="mt-1 w-full">
                Set New Password
              </Button>
            </form>
          )}

          <p className="mt-5 text-center text-xs text-text-disabled">
            <Link to="/login" className="text-glow-purple/80 transition-colors hover:text-glow-purple">
              {'\u2190'} Back to Login
            </Link>
          </p>
        </div>
      </motion.div>
    </div>
  )
}
