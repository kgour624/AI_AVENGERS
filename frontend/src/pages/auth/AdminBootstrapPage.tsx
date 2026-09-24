import { useState, type FormEvent } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { bootstrapAdminComplete, bootstrapAdminStart } from '@/api/auth'
import { useAuthStore } from '@/stores/authStore'
import { handleAPIError } from '@/utils/errors'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'

/**
 * Hidden first-admin bootstrap. Route is /bootstrap/:token — not linked
 * from login/register. Token is the unguessable one-time secret issued
 * by POST /admin/bootstrap-tokens (or ops tooling).
 *
 * Flow: credentials → QR/secret → OTP verify → admin session.
 */
export default function AdminBootstrapPage() {
  const { token: pathToken } = useParams<{ token: string }>()
  const navigate = useNavigate()
  const setAuth = useAuthStore((s) => s.setAuth)

  const [token, setToken] = useState(pathToken ?? '')
  const [fullName, setFullName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [totpCode, setTotpCode] = useState('')
  const [secret, setSecret] = useState<string | null>(null)
  const [qrUrl, setQrUrl] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleStart(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!token.trim() || !fullName.trim() || !email.trim() || password.length < 8) {
      setError('All fields are required. Password must be at least 8 characters.')
      return
    }
    setIsSubmitting(true)
    try {
      const result = await bootstrapAdminStart({
        token: token.trim(),
        email: email.trim(),
        password,
        fullName: fullName.trim(),
      })
      setSecret(result.secret)
      setQrUrl(result.qrUrl)
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  async function handleComplete(e: FormEvent) {
    e.preventDefault()
    setError(null)
    if (!totpCode.trim()) {
      setError('Enter the 6-digit code from Google Authenticator.')
      return
    }
    setIsSubmitting(true)
    try {
      const { user, tokenPair } = await bootstrapAdminComplete({
        token: token.trim(),
        totpCode: totpCode.trim(),
      })
      setAuth(user, tokenPair.accessToken)
      navigate('/admin', { replace: true })
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-surface-void p-4">
      <div className="w-full max-w-md rounded-xl border border-glass-border bg-surface-raised/90 p-6 backdrop-blur-xl">
        <h1 className="mb-1 text-xl font-semibold text-text-primary">Admin Bootstrap</h1>
        <p className="mb-4 text-xs text-text-secondary">
          One-time secure admin enrollment. Scan the authenticator QR before the account is created.
        </p>

        {!secret ? (
          <form onSubmit={handleStart} className="flex flex-col gap-3">
            {!pathToken && (
              <Input
                label="Bootstrap token"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                disabled={isSubmitting}
                required
              />
            )}
            <Input
              label="Full name"
              value={fullName}
              onChange={(e) => setFullName(e.target.value)}
              disabled={isSubmitting}
              required
            />
            <Input
              label="Email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isSubmitting}
              required
            />
            <Input
              label="Password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isSubmitting}
              required
            />
            {error && (
              <p className="rounded-md border border-mode-refuse/30 bg-mode-refuse/10 px-3 py-2 text-xs text-mode-refuse">
                {error}
              </p>
            )}
            <Button type="submit" isLoading={isSubmitting} className="w-full">
              Continue to authenticator
            </Button>
          </form>
        ) : (
          <form onSubmit={handleComplete} className="flex flex-col gap-3">
            <p className="text-sm text-text-secondary">
              Add this account in Google Authenticator, then enter a 6-digit code.
            </p>
            {qrUrl && (
              <a
                href={qrUrl}
                className="break-all rounded-md border border-glass-border bg-surface-base/60 p-2 text-xs text-glow-cyan"
                target="_blank"
                rel="noreferrer"
              >
                Open otpauth URL / scan via authenticator
              </a>
            )}
            <Input label="Manual secret" value={secret} readOnly />
            <Input
              label="Authenticator code"
              value={totpCode}
              onChange={(e) => setTotpCode(e.target.value)}
              disabled={isSubmitting}
              inputMode="numeric"
              autoComplete="one-time-code"
              required
            />
            {error && (
              <p className="rounded-md border border-mode-refuse/30 bg-mode-refuse/10 px-3 py-2 text-xs text-mode-refuse">
                {error}
              </p>
            )}
            <Button type="submit" isLoading={isSubmitting} className="w-full">
              Verify and create admin
            </Button>
          </form>
        )}
      </div>
    </div>
  )
}
