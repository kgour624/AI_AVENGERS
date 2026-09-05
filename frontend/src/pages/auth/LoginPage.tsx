export default function LoginPage() {
  // Full form implementation (api/auth.ts login(), authStore.setAuth())
  // is a Phase 2 concern - this is the routing-layer placeholder so
  // App.tsx's route tree resolves to a real component, not a TODO.
  return (
    <div className="flex h-screen items-center justify-center bg-surface-base">
      <div className="w-full max-w-sm rounded-lg border border-surface-border bg-surface-raised p-8">
        <h1 className="mb-6 text-center text-xl font-semibold text-text-primary">
          \u26A1 AI Avengers
        </h1>
        <p className="text-center text-sm text-text-secondary">Sign in - form pending Phase 2</p>
      </div>
    </div>
  )
}
