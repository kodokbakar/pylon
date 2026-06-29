import type { ServiceSignal } from '../types/service'

type SignalBadgeProps = {
  signal: ServiceSignal
}

export function SignalBadge({ signal }: SignalBadgeProps) {
  return (
    <article className="group grid grid-cols-[1fr_auto] items-end gap-4 py-5 transition-colors duration-200 hover:bg-[var(--color-grid)]">
      <div>
        <p className="font-mono text-[0.68rem] uppercase tracking-[0.24em] text-[var(--color-muted)]">
          {signal.status}
        </p>
        <h2 className="mt-1 text-2xl font-bold uppercase leading-none tracking-[-0.05em]">
          {signal.label}
        </h2>
      </div>

      <p className="font-mono text-sm font-semibold text-[var(--color-accent)]">{signal.value}</p>
    </article>
  )
}
