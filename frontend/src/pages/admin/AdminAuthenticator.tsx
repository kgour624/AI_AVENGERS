import { useState, type FormEvent } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { disableTotp, enableTotp, getTotpStatus, setupTotp } from '@/api/auth'
import { Card } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Skeleton } from '@/components/ui/Skeleton'
import { handleAPIError } from '@/utils/errors'

/**
 * Admin self-service Google Authenticator settings.
 * Wires existing backend SetupTOTP / VerifyAndEnableTOTP / DisableTOTP.
 */
function AdminAuthenticator() {
  const queryClient = useQueryClient()
  const { data, isLoading } = useQuery({
    queryKey: ['auth', 'totp'],
    queryFn: getTotpStatus,
  })

  const [secret, setSecret] = useState<string | null>(null)
  const [qrUrl, setQrUrl] = useState<string | null>(null)
  const [code, setCode] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [message, setMessage] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)

  async function handleSetup() {
    setError(null)
    setMessage(null)
    setBusy(true)
    try {
      const result = await setupTotp()
      setSecret(result.secret)
      setQrUrl(result.qrUrl)
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setBusy(false)
    }
  }

  async function handleEnable(e: FormEvent) {
    e.preventDefault()
    setError(null)
    setMessage(null)
    setBusy(true)
    try {
      await enableTotp(code.trim())
      setSecret(null)
      setQrUrl(null)
      setCode('')
      setMessage('Authenticator enabled.')
      await queryClient.invalidateQueries({ queryKey: ['auth', 'totp'] })
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setBusy(false)
    }
  }

  async function handleDisable() {
    setError(null)
    setMessage(null)
    setBusy(true)
    try {
      await disableTotp()
      setMessage('Authenticator disabled. Set up again before requiring OTP on login.')
      await queryClient.invalidateQueries({ queryKey: ['auth', 'totp'] })
    } catch (err) {
      setError(handleAPIError(err))
    } finally {
      setBusy(false)
    }
  }

  if (isLoading) {
    return (
      <div className="p-6">
        <Skeleton className="h-32" />
      </div>
    )
  }

  return (
    <div className="p-6">
      <h1 className="mb-2 text-xl font-semibold">Authenticator</h1>
      <p className="mb-4 text-xs text-text-secondary">
        Google Authenticator (TOTP) for this admin account. Required path for secure admin login.
      </p>

      <Card className="max-w-lg space-y-3 p-4">
        <p className="text-sm text-text-primary">
          Status:{' '}
          <span className={data?.totpEnabled ? 'text-mode-advise' : 'text-mode-refuse'}>
            {data?.totpEnabled ? 'Enabled' : 'Not enabled'}
          </span>
        </p>

        {!data?.totpEnabled && !secret && (
          <Button onClick={handleSetup} isLoading={busy}>
            Set up authenticator
          </Button>
        )}

        {secret && (
          <form onSubmit={handleEnable} className="space-y-3">
            <p className="text-xs text-text-secondary">
              Scan or enter the secret in Google Authenticator, then verify with a code.
            </p>
            {qrUrl && (
              <a
                href={qrUrl}
                target="_blank"
                rel="noreferrer"
                className="block break-all text-xs text-glow-cyan"
              >
                otpauth link
              </a>
            )}
            <Input label="Secret" value={secret} readOnly />
            <Input
              label="6-digit code"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              inputMode="numeric"
              autoComplete="one-time-code"
              required
            />
            <Button type="submit" isLoading={busy}>
              Verify and enable
            </Button>
          </form>
        )}

        {data?.totpEnabled && (
          <Button variant="secondary" onClick={handleDisable} isLoading={busy}>
            Disable authenticator
          </Button>
        )}

        {error && (
          <p className="rounded-md border border-mode-refuse/30 bg-mode-refuse/10 px-3 py-2 text-xs text-mode-refuse">
            {error}
          </p>
        )}
        {message && (
          <p className="rounded-md border border-mode-advise/30 bg-mode-advise/10 px-3 py-2 text-xs text-mode-advise">
            {message}
          </p>
        )}
      </Card>
    </div>
  )
}

export const Component = AdminAuthenticator
