import { NavLink, Outlet } from 'react-router-dom'

export default function Layout() {
  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-50 border-b bg-background/80 backdrop-blur">
        <div className="mx-auto flex w-full max-w-5xl items-center justify-between px-4 py-3">
          <div className="text-sm font-semibold tracking-tight">Task Manager</div>
          <nav className="flex items-center gap-2 text-sm">
            <NavLink
              className={({ isActive }) =>
                isActive
                  ? 'rounded-md bg-secondary px-3 py-1.5 text-secondary-foreground'
                  : 'rounded-md px-3 py-1.5 text-muted-foreground hover:bg-accent hover:text-accent-foreground'
              }
              to="/"
            >
              Tasks
            </NavLink>
            <NavLink
              className={({ isActive }) =>
                isActive
                  ? 'rounded-md bg-secondary px-3 py-1.5 text-secondary-foreground'
                  : 'rounded-md px-3 py-1.5 text-muted-foreground hover:bg-accent hover:text-accent-foreground'
              }
              to="/stats"
            >
              Stats
            </NavLink>
          </nav>
        </div>
      </header>

      <main className="mx-auto w-full max-w-5xl px-4 py-6">
        <Outlet />
      </main>
    </div>
  )
}
