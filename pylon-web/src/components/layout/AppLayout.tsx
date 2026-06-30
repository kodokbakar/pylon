import { Outlet } from 'react-router-dom'

import { useSidebar } from '../../hooks/useSidebar'
import { Sidebar } from './Sidebar'

export function AppLayout() {
  const sidebar = useSidebar()

  return (
    <div className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      {sidebar.isOpen ? (
        <button
          aria-label="Close sidebar backdrop"
          className="fixed inset-0 z-40 bg-black/50 md:hidden"
          type="button"
          onClick={sidebar.close}
        />
      ) : null}

      <Sidebar
        isOpen={sidebar.isOpen}
        onClose={sidebar.close}
        onRoomSelect={sidebar.closeOnMobile}
      />

      <div className="min-h-screen md:pl-[280px]">
        <header className="sticky top-0 z-30 flex items-center justify-between border-b-2 border-[var(--color-ink)] bg-[var(--color-paper)] px-4 py-3 md:hidden">
          <button
            aria-expanded={sidebar.isOpen}
            aria-label="Open sidebar"
            className="border-2 border-[var(--color-ink)] px-3 py-2 font-mono text-xs uppercase tracking-[0.16em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
            type="button"
            onClick={sidebar.toggle}
          >
            ☰
          </button>

          <p className="font-mono text-[0.65rem] uppercase tracking-[0.22em] text-[var(--color-muted)]">
            Pylon
          </p>
        </header>

        <main className="mx-auto min-h-screen w-full max-w-7xl px-5 py-8 sm:px-8 lg:px-10">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
