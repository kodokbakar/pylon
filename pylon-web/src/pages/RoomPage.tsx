import { Link, useParams } from 'react-router-dom'

import { RoomList } from '../components/room/RoomList'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

export function RoomPage() {
  const { roomId } = useParams<{ roomId: string }>()

  useDocumentTitle(`Pylon Chat ${roomId ?? ''}`.trim())

  return (
    <main className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
        <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
          <Link
            className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)] hover:text-[var(--color-accent)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
            to={`/rooms/${roomId ?? ''}`}
          >
            ← Back to room detail
          </Link>
        </header>

        <div className="col-span-12 grid grid-cols-12 gap-x-4 pt-10 lg:pt-16">
          <aside className="col-span-12 lg:col-span-4">
            <RoomList />
          </aside>

          <section className="col-span-12 mt-12 lg:col-span-8 lg:mt-0 lg:pl-6">
            <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
              Chat shell
            </p>

            <h1 className="text-[clamp(3.8rem,13vw,9rem)] font-black uppercase leading-[0.86] tracking-[-0.08em]">
              Room {roomId}
            </h1>

            <p className="mt-8 max-w-2xl border-t border-[var(--color-line)] pt-5 text-xl font-semibold leading-tight tracking-[-0.03em]">
              Chat streaming will attach here in the next issue. This route is ready at
              `/rooms/:roomId/chat`.
            </p>
          </section>
        </div>
      </section>
    </main>
  )
}
