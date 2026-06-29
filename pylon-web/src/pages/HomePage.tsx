import { RoomList } from '../components/room/RoomList'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

const systemNotes = ['Authenticated rooms', 'Connect service', 'Token refresh', 'React Query cache']

export function HomePage() {
  useDocumentTitle('Pylon Web')

  return (
    <main className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
        <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
          <div className="grid grid-cols-12 gap-x-4">
            <p className="col-span-12 font-mono text-xs uppercase tracking-[0.34em] text-[var(--color-muted)] sm:col-span-5">
              Pylon / realtime chat infrastructure
            </p>

            <p className="col-span-12 mt-3 font-mono text-xs uppercase tracking-[0.24em] text-[var(--color-muted)] sm:col-span-7 sm:mt-0 sm:text-right">
              Protected room navigation
            </p>
          </div>
        </header>

        <div className="col-span-12 grid grid-cols-12 gap-x-4 pt-10 lg:pt-16">
          <aside className="col-span-12 lg:col-span-4">
            <RoomList />
          </aside>

          <section
            className="col-span-12 mt-12 lg:col-span-8 lg:mt-0 lg:pl-6"
            aria-labelledby="page-title"
          >
            <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
              Sprint 2 / rooms
            </p>

            <h1
              id="page-title"
              className="max-w-5xl text-[clamp(4rem,15vw,12rem)] font-black uppercase leading-[0.82] tracking-[-0.09em]"
            >
              Select Room
            </h1>

            <div className="mt-8 grid max-w-4xl grid-cols-12 gap-4 border-t border-[var(--color-line)] pt-5">
              <p className="col-span-12 text-xl font-semibold leading-tight tracking-[-0.03em] sm:col-span-5 sm:text-2xl">
                Your room list is now the main entry point into chat.
              </p>

              <p className="col-span-12 max-w-2xl text-base leading-7 text-[var(--color-muted)] sm:col-span-7">
                The sidebar fetches joined rooms through the authenticated RoomService client,
                attaches the JWT automatically, and keeps the list cached for fast navigation.
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
        </div>
      </section>
    </main>
  )
}
