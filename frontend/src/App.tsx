import { createBrowserRouter, RouterProvider } from 'react-router-dom'
import { AuthGuard } from '@/components/layout/AuthGuard'
import { AdminGuard } from '@/components/layout/AdminGuard'
import { AppShell } from '@/components/layout/AppShell'
import { RouteError } from '@/components/layout/RouteError'

import LoginPage from '@/pages/auth/LoginPage'
import RegisterPage from '@/pages/auth/RegisterPage'
import { projectsRoute } from '@/pages/projects/ProjectsPage'
import { projectRoute } from '@/pages/projects/ProjectPage'
import { chatRoute } from '@/pages/chat/ChatPage'
import { expertsRoute } from '@/pages/experts/ExpertsPage'

/**
 * Route tree. Source: FRONTEND_SYSTEM_DESIGN.md section 5.
 *
 * Cross-question applied before finalizing: does every `errorElement`
 * actually need to be set at every level, or just the root? Traced
 * through - a loader throwing inside the /admin lazy subtree would
 * bubble up past AdminGuard to the nearest ANCESTOR errorElement if
 * the lazy route itself doesn't define one. Setting errorElement only
 * on the root route covers every case (React Router bubbles unhandled
 * loader/render errors up the route tree), so a single root-level
 * errorElement is correct and sufficient here - not an oversight.
 */
const router = createBrowserRouter([
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },

  {
    element: <AuthGuard />,
    errorElement: <RouteError />,
    children: [
      {
        element: <AppShell />,
        children: [
          { path: '/', ...projectsRoute },
          { path: '/projects/:projectId', ...projectRoute },
          { path: '/projects/:projectId/chats/:chatId', ...chatRoute },
          { path: '/experts', ...expertsRoute },
          // Workflow routes — were unregistered (404). Now wired.
          {
            path: '/workflows',
            lazy: () => import('@/pages/workflows/WorkflowsPage'),
          },
          {
            path: '/workflows/:id/kanban',
            lazy: () => import('@/pages/workflows/KanbanPage'),
          },
        ],
      },

      // Admin subtree - guarded by role AND code-split. WHY nested
      // under AuthGuard rather than a sibling top-level route: an
      // unauthenticated visitor hitting /admin directly must still go
      // through the login-redirect logic in AuthGuard first, before
      // AdminGuard even gets a chance to check the role claim.
      {
        path: '/admin',
        element: <AdminGuard />,
        children: [
          {
            lazy: () => import('@/pages/admin/AdminLayout'),
            children: [
              { index: true, lazy: () => import('@/pages/admin/AdminDashboard') },
              { path: 'experts', lazy: () => import('@/pages/admin/AdminExperts') },
              { path: 'categories', lazy: () => import('@/pages/admin/AdminCategories') },
              { path: 'domain-profiles', lazy: () => import('@/pages/admin/AdminDomainProfiles') },
              { path: 'clients', lazy: () => import('@/pages/admin/AdminClients') },
              { path: 'stats', lazy: () => import('@/pages/admin/AdminStats') },
              { path: 'settings', lazy: () => import('@/pages/admin/AdminSettings') },
              { path: 'llm-settings', lazy: () => import('@/pages/admin/AdminLLMSettings') },
            ],
          },
        ],
      },
    ],
  },
])

export function App() {
  return <RouterProvider router={router} />
}
