import { useDocumentTitle } from '../hooks/useDocumentTitle'

const systemNotes = ['Authenticated rooms', 'Connect service', 'Token refresh', 'React Query cache']

export function HomePage() {
  useDocumentTitle('Pylon Web')

  return (
    <section className="pt-8 lg:pt-14" aria-labelledby="page-title">
      <header className="border-b border-[var(--color-line)] pb-4">
        <div className="grid grid-cols-12 gap-x-4">
          <p className="col-span-12 font-mono text-xs uppercase tracking-[0.34em] text-[var(--color-muted)] sm:col-span-5">
            Pylon / realtime chat infrastructure
          </p>

          <p className="col-span-12 mt-3 font-mono text-xs uppercase tracking-[0.24em] text-[var(--color-muted)] sm:col-span-7 sm:mt-0 sm:text-right">
            Protected room navigation
          </p>
        </div>
      </header>

      <section className="pt-10 lg:pt-16">
        <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
          Sprint 3 / layout
        </p>

        <h1
          id="page-title"
          className="max-w-5xl text-[clamp(4rem,15vw,12rem)] font-black uppercase leading-[0.82] tracking-[-0.09em]"
        >
          Select Room
        </h1>

        <div className="mt-8 grid max-w-4xl grid-cols-12 gap-4 border-t border-[var(--color-line)] pt-5">
          <p className="col-span-12 text-xl font-semibold leading-tight tracking-[-0.03em] sm:col-span-5 sm:text-2xl">
            Your room list is now the persistent navigation rail.
          </p>

          <p className="col-span-12 max-w-2xl text-base leading-7 text-[var(--color-muted)] sm:col-span-7">
            Use the sidebar to jump between rooms. On mobile, open it with the menu button and it
            slides in as a full-width overlay.
          </p>
        </div>

        <div className="mt-10 grid grid-cols-2 border border-[var(--color-ink)]">
          {systemNotes.map((note) => (
            <div
              className="border-b border-r border-[var(--color-line)] px-4 py-5 font-mono text-xs uppercase tracking-[0.18em] text-[var(--color-muted)]"
              key={note}
            >
              {note}
            </div>
          ))}
        </div>
      </section>
    </section>
  )
}
