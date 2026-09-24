import { Link } from 'react-router-dom'

/**
 * Self-registration is disabled. This page stays reachable (so an old
 * bookmark or a direct hit on /register does not 404) but is intentionally
 * non-functional and explains how to get an account.
 *
 * The previous registration form and the `register` API call are kept in
 * the codebase (see frontend/src/api/auth.ts) so the feature can be
 * re-enabled without re-writing it — flip SELF_REGISTRATION_ENABLED=true
 * on the backend and restore the form here.
 */
export default function RegisterPage() {
  return (
    <div className="flex h-screen items-center justify-center bg-surface-base">
      <div className="w-full max-w-sm rounded-lg border border-surface-border bg-surface-raised p-8 text-center">
        <h1 className="mb-3 text-xl font-semibold text-text-primary">Registration closed</h1>
        <p className="text-sm text-text-secondary">
          Self registration is disabled. Accounts are created by an administrator —
          please contact your admin to get access.
        </p>
        <p className="mt-6 text-center text-sm text-text-secondary">
          Already have an account?{' '}
          <Link to="/login" className="text-brand hover:text-brand-hover">
            Sign in
          </Link>
        </p>
      </div>
    </div>
  )
}
