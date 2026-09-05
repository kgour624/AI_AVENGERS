import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { login } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import { handleAPIError } from '@/utils/errors'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

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
    <div className="flex h-screen items-center justify-center bg-surface-base">
      <div className="w-full max-w-sm rounded-lg border border-surface-border bg-surface-raised p-8">
        <h1 className="mb-6 text-center text-xl font-semibold text-text-primary">
          \u26A1 AI Avengers
        </h1>
        <p className="mb-6 text-center text-sm text-text-secondary">
          Domain Expert Team Simulator
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
      </div>
    </div>
  )
}
