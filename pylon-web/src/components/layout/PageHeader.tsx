import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

type PageHeaderProps = {
  eyebrow?: string
  title?: string
  backTo?: string
  backLabel?: string
  actions?: ReactNode
}

export function PageHeader({ eyebrow, title, backTo, backLabel, actions }: PageHeaderProps) {
  return (
    <header className="border-b border-[var(--color-line)] pb-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          {backTo ? (
            <Link
              className="inline-flex font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)] hover:text-[var(--color-accent)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
              to={backTo}
            >
              {backLabel ?? '← Back'}
            </Link>
          ) : null}

          {eyebrow ? (
            <p className="mt-4 font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)] first:mt-0">
              {eyebrow}
            </p>
          ) : null}

          {title ? (
            <h1 className="mt-2 truncate text-4xl font-black uppercase leading-none tracking-[-0.06em] sm:text-6xl">
              {title}
            </h1>
          ) : null}
        </div>

        {actions ? <div className="shrink-0">{actions}</div> : null}
      </div>
    </header>
  )
}
