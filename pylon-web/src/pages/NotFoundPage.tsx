import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <main className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
        <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
          <p className="font-mono text-xs uppercase tracking-[0.34em] text-[var(--color-muted)]">
            Pylon Web / route not found
          </p>
        </header>

        <section className="col-span-12 pt-12 lg:col-span-9 lg:pt-20">
          <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
            404
          </p>

          <h1 className="max-w-5xl text-[clamp(4rem,15vw,11rem)] font-black uppercase leading-[0.82] tracking-[-0.09em]">
            No route
          </h1>

          <p className="mt-8 max-w-2xl border-t border-[var(--color-line)] pt-5 text-xl font-semibold leading-tight tracking-[-0.03em]">
            The requested route does not exist in the Pylon Web route table.
          </p>

          <Link
            className="mt-10 inline-flex items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
            to="/"
          >
            Return to app
            <span className="ml-8">→</span>
          </Link>
        </section>
      </section>
    </main>
  )
}
