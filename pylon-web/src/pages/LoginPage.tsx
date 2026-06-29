import { Link, useLocation, useNavigate } from 'react-router-dom'

import { setAuthToken } from '../utils/authToken'

type LoginLocationState = {
  registrationSuccess?: string
}

export function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const locationState = location.state as LoginLocationState | null
  const registrationSuccess = locationState?.registrationSuccess

  function handleDemoLogin() {
    setAuthToken('demo-auth-token')
    navigate('/', { replace: true })
  }

  return (
    <main className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
        <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
          <p className="font-mono text-xs uppercase tracking-[0.34em] text-[var(--color-muted)]">
            Pylon Web / public route
          </p>
        </header>

        <section className="col-span-12 grid grid-cols-12 gap-x-4 pt-12 lg:pt-20">
          <div className="col-span-12 lg:col-span-7">
            <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
              Authentication
            </p>

            <h1 className="max-w-4xl text-[clamp(4rem,14vw,10rem)] font-black uppercase leading-[0.84] tracking-[-0.08em]">
              Login
            </h1>
          </div>

          <div className="col-span-12 mt-10 border-y-2 border-[var(--color-ink)] py-6 lg:col-span-5 lg:mt-0">
            {registrationSuccess ? (
              <div
                className="mb-6 border-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
                role="status"
              >
                {registrationSuccess}
              </div>
            ) : null}

            <p className="font-mono text-xs uppercase tracking-[0.24em] text-[var(--color-muted)]">
              Local routing checkpoint
            </p>

            <p className="mt-5 text-xl font-semibold leading-tight tracking-[-0.03em]">
              This page proves `/login` renders as a public route.
            </p>

            <p className="mt-4 leading-7 text-[var(--color-muted)]">
              Use the demo action to write `auth_token` into localStorage and redirect into the
              protected app route.
            </p>

            <div className="mt-8 flex flex-col gap-3">
              <button
                className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                type="button"
                onClick={handleDemoLogin}
              >
                Continue with demo token
                <span>↗</span>
              </button>

              <Link
                className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                to="/register"
              >
                Create account
                <span>→</span>
              </Link>
            </div>
          </div>
        </section>
      </section>
    </main>
  )
}
