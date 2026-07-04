import { Link } from 'react-router-dom'

import { useDocumentTitle } from '../hooks/useDocumentTitle'

const features = [
  {
    label: '01',
    title: 'Realtime Chat',
    description:
      'Low-latency room messaging over WebSocket, with optimistic sends and resilient reconnect behavior.',
  },
  {
    label: '02',
    title: 'Rooms',
    description:
      'Persistent room navigation keeps team conversations organized without losing operational context.',
  },
  {
    label: '03',
    title: 'Presence',
    description:
      'Online, offline, and typing states make every room feel alive without noisy interface chrome.',
  },
  {
    label: '04',
    title: 'Microservices',
    description:
      'Built around Go services, Connect RPC, PostgreSQL, Redis, Kafka, and observable deployment primitives.',
  },
]

const stats = ['WebSocket', 'Connect RPC', 'Presence stream', 'Token refresh']

export function LandingPage() {
  useDocumentTitle('Pylon / Realtime Chat Infrastructure')

  return (
    <main className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
        <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
          <nav className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <Link
              className="font-mono text-xs uppercase tracking-[0.34em] text-[var(--color-muted)] hover:text-[var(--color-accent)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
              to="/"
            >
              Pylon Web / realtime chat infrastructure
            </Link>

            <div className="flex flex-col gap-3 sm:flex-row">
              <Link
                className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] px-4 py-3 font-mono text-xs uppercase tracking-[0.18em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                to="/login"
              >
                Login
                <span className="ml-8">→</span>
              </Link>

              <Link
                className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-4 py-3 font-mono text-xs uppercase tracking-[0.18em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                to="/register"
              >
                Register
                <span className="ml-8">↗</span>
              </Link>
            </div>
          </nav>
        </header>

        <section className="col-span-12 grid grid-cols-12 gap-4 pt-12 lg:pt-20">
          <div className="col-span-12 lg:col-span-9">
            <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
              Distributed chat control plane
            </p>

            <h1 className="max-w-6xl break-words text-[clamp(4rem,18vw,15rem)] font-black uppercase leading-[0.78] tracking-[-0.1em]">
              Pylon
            </h1>

            <p className="mt-8 max-w-3xl border-t border-[var(--color-line)] pt-5 text-2xl font-semibold leading-tight tracking-[-0.04em] sm:text-3xl">
              Realtime rooms, presence, and WebSocket messaging for teams that need fast
              coordination without a bloated interface.
            </p>
          </div>

          <aside className="col-span-12 border-2 border-[var(--color-ink)] lg:col-span-3 lg:self-end">
            {stats.map((stat) => (
              <div className="border-b border-[var(--color-line)] px-4 py-5" key={stat}>
                <p className="font-mono text-[0.65rem] uppercase tracking-[0.18em] text-[var(--color-muted)]">
                  Stack signal
                </p>
                <p className="mt-2 text-lg font-black uppercase tracking-[-0.03em]">{stat}</p>
              </div>
            ))}
          </aside>
        </section>

        <section className="col-span-12 pt-14 lg:pt-20" aria-labelledby="features-title">
          <div className="border-y-2 border-[var(--color-ink)] py-4">
            <p className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)]">
              Feature map
            </p>
            <h2
              className="mt-2 text-4xl font-black uppercase leading-none tracking-[-0.06em] sm:text-6xl"
              id="features-title"
            >
              Built for live rooms
            </h2>
          </div>

          <div className="grid grid-cols-1 border-l border-[var(--color-line)] md:grid-cols-2">
            {features.map((feature) => (
              <article
                className="border-b border-r border-[var(--color-line)] px-4 py-8 sm:px-6"
                key={feature.title}
              >
                <p className="font-mono text-xs uppercase tracking-[0.24em] text-[var(--color-accent)]">
                  {feature.label}
                </p>
                <h3 className="mt-4 text-3xl font-black uppercase leading-none tracking-[-0.05em]">
                  {feature.title}
                </h3>
                <p className="mt-4 max-w-xl text-base font-semibold leading-7 text-[var(--color-muted)]">
                  {feature.description}
                </p>
              </article>
            ))}
          </div>
        </section>

        <section className="col-span-12 py-14 lg:py-20" aria-labelledby="cta-title">
          <div className="grid grid-cols-12 gap-4 border-2 border-[var(--color-ink)] px-5 py-6 shadow-[6px_6px_0_var(--color-ink)] sm:shadow-[10px_10px_0_var(--color-ink)]">
            <div className="col-span-12 lg:col-span-7">
              <p className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)]">
                Access terminal
              </p>
              <h2
                className="mt-3 text-4xl font-black uppercase leading-none tracking-[-0.06em] sm:text-6xl"
                id="cta-title"
              >
                Open the room grid
              </h2>
              <p className="mt-5 max-w-2xl text-lg font-semibold leading-7 text-[var(--color-muted)]">
                Sign in to continue an existing session, or create an operator account to start
                testing realtime collaboration flows.
              </p>
            </div>

            <div className="col-span-12 flex flex-col justify-end gap-3 lg:col-span-5">
              <Link
                className="inline-flex w-full items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                to="/login"
              >
                Login
                <span className="ml-8">→</span>
              </Link>

              <Link
                className="inline-flex w-full items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                to="/register"
              >
                Register
                <span className="ml-8">↗</span>
              </Link>
            </div>
          </div>
        </section>
      </section>
    </main>
  )
}
