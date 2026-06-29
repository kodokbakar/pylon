import { Link } from 'react-router-dom'

import { getApiBaseUrl } from '../api/config'
import { SignalBadge } from '../components/SignalBadge'
import { useDocumentTitle } from '../hooks/useDocumentTitle'

const apiBaseUrl = getApiBaseUrl()

const serviceSignals = [
  { label: 'API Gateway', value: '8080', status: 'online' },
  { label: 'Chat Service', value: '9001', status: 'ready' },
  { label: 'Presence', value: '9002', status: 'synced' },
  { label: 'Room Service', value: '9003', status: 'ready' },
]

const systemNotes = ['Bun runtime', 'React + TypeScript', 'Tailwind CSS v4', 'Vite dev server']

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
              Protected app route
            </p>
          </div>
        </header>

        <div className="col-span-12 grid grid-cols-12 gap-x-4 pt-10 lg:pt-16">
          <section className="col-span-12 lg:col-span-8" aria-labelledby="page-title">
            <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
              Bun + Vite + TypeScript
            </p>

            <h1
              id="page-title"
              className="max-w-5xl text-[clamp(4rem,15vw,12rem)] font-black uppercase leading-[0.82] tracking-[-0.09em]"
            >
              Pylon Web
            </h1>

            <div className="mt-8 grid max-w-4xl grid-cols-12 gap-4 border-t border-[var(--color-line)] pt-5">
              <p className="col-span-12 text-xl font-semibold leading-tight tracking-[-0.03em] sm:col-span-5 sm:text-2xl">
                A React control surface for the distributed chat system.
              </p>

              <p className="col-span-12 max-w-2xl text-base leading-7 text-[var(--color-muted)] sm:col-span-7">
                This protected route renders only when the auth session exists. The auth layer now
                owns token persistence, guarded routes, refresh, and API request authorization.
              </p>
            </div>

            <div className="mt-10 flex flex-col gap-3 sm:flex-row">
              <a
                className="group inline-flex items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                href={`${apiBaseUrl}/health`}
              >
                Check API health
                <span className="ml-8 transition-transform duration-200 group-hover:translate-x-1">
                  ↗
                </span>
              </a>

              <a
                className="group inline-flex items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                href={`${apiBaseUrl}/metrics`}
              >
                Inspect metrics
                <span className="ml-8 transition-transform duration-200 group-hover:translate-x-1">
                  ↗
                </span>
              </a>

              <Link
                className="group inline-flex items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                to="/rooms/general"
              >
                Open demo room
                <span className="ml-8 transition-transform duration-200 group-hover:translate-x-1">
                  →
                </span>
              </Link>
            </div>
          </section>

          <aside className="col-span-12 mt-12 border-[var(--color-ink)] lg:col-span-4 lg:mt-0 lg:border-l lg:pl-6">
            <div className="border-y-2 border-[var(--color-ink)] py-4">
              <p className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)]">
                Service register
              </p>
            </div>

            <div className="divide-y divide-[var(--color-line)]">
              {serviceSignals.map((signal) => (
                <SignalBadge key={signal.label} signal={signal} />
              ))}
            </div>

            <div className="mt-8 grid grid-cols-2 border border-[var(--color-ink)]">
              {systemNotes.map((note) => (
                <div
                  className="border-b border-r border-[var(--color-line)] px-4 py-5 font-mono text-xs uppercase tracking-[0.18em] text-[var(--color-muted)]"
                  key={note}
                >
                  {note}
                </div>
              ))}
            </div>
          </aside>
        </div>
      </section>
    </main>
  )
}
