import { useEffect, useMemo, useState } from 'react'

import type { ChatMessage } from '../../hooks/useChatMessages'

type MessageItemProps = {
  message: ChatMessage
  isOwnMessage: boolean
  isGroupedWithPrevious: boolean
}

export function MessageItem({ message, isOwnMessage, isGroupedWithPrevious }: MessageItemProps) {
  const now = useRelativeClock()
  const relativeTime = useMemo(
    () => formatRelativeTime(message.createdAt, now),
    [message.createdAt, now],
  )

  return (
    <article
      className={[
        'grid grid-cols-[2.75rem_1fr] gap-3 py-2',
        isGroupedWithPrevious ? 'pt-0' : 'pt-5',
        isOwnMessage ? 'text-[var(--color-ink)]' : '',
      ].join(' ')}
    >
      <div>
        {!isGroupedWithPrevious ? (
          <div className="flex size-11 items-center justify-center overflow-hidden border-2 border-[var(--color-ink)] bg-[var(--color-grid)] font-mono text-sm font-black uppercase">
            {message.senderAvatarUrl ? (
              <img alt="" className="size-full object-cover" src={message.senderAvatarUrl} />
            ) : (
              getInitial(message.senderName)
            )}
          </div>
        ) : null}
      </div>

      <div className="min-w-0">
        {!isGroupedWithPrevious ? (
          <div className="mb-2 flex flex-wrap items-baseline gap-x-3 gap-y-1">
            <p className="font-black uppercase tracking-[-0.03em]">
              {isOwnMessage ? 'You' : message.senderName}
            </p>

            <time
              className="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-[var(--color-muted)]"
              dateTime={message.createdAt}
            >
              {relativeTime}
            </time>

            {message.status !== 'sent' ? (
              <span className="font-mono text-[0.65rem] uppercase tracking-[0.14em] text-[var(--color-accent)]">
                {message.status}
              </span>
            ) : null}
          </div>
        ) : null}

        <div
          className={[
            'inline-block max-w-full border-2 px-4 py-3 text-sm font-semibold leading-6 shadow-[4px_4px_0_var(--color-ink)]',
            isOwnMessage
              ? 'border-[var(--color-ink)] bg-[var(--color-ink)] text-[var(--color-paper)]'
              : 'border-[var(--color-ink)] bg-[var(--color-paper)] text-[var(--color-ink)]',
            message.status === 'error'
              ? 'border-[var(--color-accent)] text-[var(--color-accent)]'
              : '',
          ].join(' ')}
        >
          <p className="whitespace-pre-wrap break-words">{message.content}</p>
        </div>
      </div>
    </article>
  )
}

function useRelativeClock() {
  const [now, setNow] = useState(() => Date.now())

  useEffect(() => {
    const timer = window.setInterval(() => {
      setNow(Date.now())
    }, 60_000)

    return () => {
      window.clearInterval(timer)
    }
  }, [])

  return now
}

function formatRelativeTime(value: string, now: number) {
  const time = Date.parse(value)
  if (!Number.isFinite(time)) {
    return 'unknown'
  }

  const diffMs = Math.max(now - time, 0)
  const minute = 60_000
  const hour = 60 * minute
  const day = 24 * hour

  if (diffMs < minute) {
    return 'now'
  }

  if (diffMs < hour) {
    return `${Math.floor(diffMs / minute)}m ago`
  }

  if (diffMs < day) {
    return `${Math.floor(diffMs / hour)}h ago`
  }

  if (isYesterday(new Date(time), new Date(now))) {
    return 'Yesterday'
  }

  return new Intl.DateTimeFormat('en', {
    month: 'short',
    day: '2-digit',
  }).format(new Date(time))
}

function isYesterday(date: Date, now: Date) {
  const yesterday = new Date(now)
  yesterday.setDate(now.getDate() - 1)

  return (
    date.getFullYear() === yesterday.getFullYear() &&
    date.getMonth() === yesterday.getMonth() &&
    date.getDate() === yesterday.getDate()
  )
}

function getInitial(name: string) {
  const trimmedName = name.trim()
  return trimmedName[0]?.toUpperCase() ?? '?'
}
