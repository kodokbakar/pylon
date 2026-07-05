import { sanitizeErrorMessage } from '../utils/backendError'
import { Component, Fragment, type ErrorInfo, type PropsWithChildren } from 'react'

type ErrorBoundaryProps = PropsWithChildren<{
  eyebrow?: string
  title?: string
  homeHref?: string
  onNavigateHome?: (href: string) => void
}>

type ErrorBoundaryState = {
  error: Error | null
  resetKey: number
}

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = {
    error: null,
    resetKey: 0,
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Pylon UI boundary caught a render failure.', error, errorInfo)
  }

  private handleReset = () => {
    this.setState((state) => ({
      error: null,
      resetKey: state.resetKey + 1,
    }))
  }

  private handleGoHome = () => {
    const homeHref = this.props.homeHref ?? '/'

    this.handleReset()

    if (this.props.onNavigateHome) {
      this.props.onNavigateHome(homeHref)
      return
    }

    if (window.location.pathname !== homeHref) {
      window.location.assign(homeHref)
    }
  }

  render() {
    const { children, eyebrow = 'Interface fault', title = 'Panel crashed' } = this.props
    const { error, resetKey } = this.state

    if (!error) {
      return <Fragment key={resetKey}>{children}</Fragment>
    }

    const errorMessage = sanitizeErrorMessage(
      error,
      'This section stopped rendering. Please retry or return home.',
    )

    return (
      <main
        aria-labelledby="error-boundary-title"
        className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]"
        role="alert"
      >
        <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
          <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
            <p className="font-mono text-xs uppercase tracking-[0.34em] text-[var(--color-muted)]">
              Pylon Web / graceful recovery
            </p>
          </header>

          <section className="col-span-12 pt-12 lg:col-span-9 lg:pt-20">
            <p className="mb-5 inline-flex border border-[var(--color-accent)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em] text-[var(--color-accent)]">
              {eyebrow}
            </p>

            <h1
              className="max-w-5xl text-[clamp(4rem,15vw,11rem)] font-black uppercase leading-[0.82] tracking-[-0.09em]"
              id="error-boundary-title"
            >
              {title}
            </h1>

            <p className="mt-8 max-w-2xl border-t border-[var(--color-line)] pt-5 text-xl font-semibold leading-tight tracking-[-0.03em]">
              Pylon caught a rendering failure before it could collapse the entire interface.
            </p>

            <div className="mt-8 border-2 border-[var(--color-accent)] px-4 py-4 shadow-[8px_8px_0_var(--color-ink)]">
              <p className="font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-accent)]">
                Error message
              </p>
              <p className="mt-3 break-words text-base font-black leading-6 tracking-[-0.03em]">
                {errorMessage}
              </p>
            </div>

            <div className="mt-10 flex flex-col gap-3 sm:flex-row">
              <button
                className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                type="button"
                onClick={this.handleReset}
              >
                Try again
                <span className="ml-8">↻</span>
              </button>

              <button
                className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                type="button"
                onClick={this.handleGoHome}
              >
                Go home
                <span className="ml-8">→</span>
              </button>
            </div>
          </section>
        </section>
      </main>
    )
  }
}
