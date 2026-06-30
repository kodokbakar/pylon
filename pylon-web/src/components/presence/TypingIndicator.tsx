type TypingIndicatorProps = {
  names: string[]
}

export function TypingIndicator({ names }: TypingIndicatorProps) {
  if (names.length === 0) {
    return null
  }

  return (
    <div
      className="border-x-2 border-[var(--color-ink)] bg-[var(--color-grid)] px-4 py-3"
      aria-live="polite"
    >
      <p className="inline-flex items-center gap-2 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-muted)]">
        <span>{formatTypingText(names)}</span>
        <span className="inline-flex items-end gap-1" aria-hidden="true">
          {[0, 1, 2].map((index) => (
            <span
              className="size-1.5 bg-[var(--color-presence-typing)] [animation:typing-dot-bounce_1.1s_infinite]"
              key={index}
              style={{ animationDelay: `${index * 140}ms` }}
            />
          ))}
        </span>
      </p>
    </div>
  )
}

function formatTypingText(names: string[]) {
  if (names.length === 1) {
    return `${names[0]} is typing`
  }

  if (names.length === 2) {
    return `${names[0]}, ${names[1]} are typing`
  }

  return `${names[0]}, ${names[1]}, +${names.length - 2} are typing`
}
