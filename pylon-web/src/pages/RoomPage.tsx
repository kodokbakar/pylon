import { Link, useParams } from 'react-router-dom'

import { useDocumentTitle } from '../hooks/useDocumentTitle'

export function RoomPage() {
  const { id } = useParams<{ id: string }>()

  useDocumentTitle(`Pylon Room ${id ?? ''}`.trim())

  return (
    <main className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
        <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
          <Link
            className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)] hover:text-[var(--color-accent)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
            to="/"
          >
            ← Back to app
          </Link>
        </header>

        <section className="col-span-12 pt-12 lg:col-span-8 lg:pt-20">
          <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
            Protected room route
          </p>

          <h1 className="text-[clamp(3.8rem,13vw,9rem)] font-black uppercase leading-[0.86] tracking-[-0.08em]">
            Room {id}
          </h1>

          <p className="mt-8 max-w-2xl border-t border-[var(--color-line)] pt-5 text-xl font-semibold leading-tight tracking-[-0.03em]">
            `/rooms/:id` is now routed behind the same auth token guard as the main app.
          </p>
        </section>
      </section>
    </main>
  )
}
