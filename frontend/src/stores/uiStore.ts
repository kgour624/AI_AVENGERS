import { create } from 'zustand'

interface UIState {
  sidebarOpen: boolean
  theme: 'dark' | 'light'
  toggleSidebar: () => void
  setSidebarOpen: (open: boolean) => void
  setTheme: (theme: 'dark' | 'light') => void
}

// WHY default theme is 'dark': FRONTEND_SYSTEM_DESIGN.md section 2 -
// "Dark theme by default", and index.html already sets class="dark"
// on <html> to match, avoiding a flash-of-light-theme on first paint.
export const useUIStore = create<UIState>((set) => ({
  sidebarOpen: true,
  theme: 'dark',

  toggleSidebar: () => set((state) => ({ sidebarOpen: !state.sidebarOpen })),
  setSidebarOpen: (open) => set({ sidebarOpen: open }),

  setTheme: (theme) => {
    document.documentElement.classList.toggle('dark', theme === 'dark')
    set({ theme })
  },
}))
