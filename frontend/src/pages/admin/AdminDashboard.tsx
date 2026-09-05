function AdminDashboard() {
  return <div className="p-6">Admin Dashboard - stats wiring pending Phase 5</div>
}

// React Router's `lazy: () => import(...)` convention requires a named
// `Component` export (or `loader`/`action`/`ErrorBoundary`) - a bare
// default export is NOT picked up by the router and would render
// nothing with no error, which is worse than a crash because it fails
// silently. Verified this against the same pattern already used in
// AdminLayout.tsx and applied consistently to every admin leaf page.
export const Component = AdminDashboard
