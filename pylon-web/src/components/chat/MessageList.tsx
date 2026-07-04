import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import type { ChatMessage } from '../../hooks/useChatMessages'
import { Skeleton } from '../ui/Skeleton'
import { MessageItem } from './MessageItem'

type MessageListProps = {
  messages: ChatMessage[]
  currentUserId: string
  isLoading: boolean
  errorMessage: string | null
  hasMore: boolean
  isLoadingOlder: boolean
  onLoadOlder: () => Promise<void>
}

const groupingWindowMs = 5 * 60_000

export function MessageList({
  messages,
  currentUserId,
  isLoading,
  errorMessage,
  hasMore,
  isLoadingOlder,
  onLoadOlder,
}: MessageListProps) {
  const scrollContainerRef = useRef<HTMLDivElement>(null)
  const bottomRef = useRef<HTMLDivElement>(null)
  const shouldStickToBottomRef = useRef(true)
  const [isNearBottom, setIsNearBottom] = useState(true)

  const groupedMessages = useMemo(
    () =>
      messages.map((message, index) => ({
        message,
        isGroupedWithPrevious: isGroupedWithPrevious(message, messages[index - 1]),
      })),
    [messages],
  )

  const handleScroll = useCallback(() => {
    const element = scrollContainerRef.current
    if (!element) {
      return
    }

    const distanceFromBottom = element.scrollHeight - element.scrollTop - element.clientHeight
    const nextIsNearBottom = distanceFromBottom < 96
    shouldStickToBottomRef.current = nextIsNearBottom
    setIsNearBottom(nextIsNearBottom)
  }, [])

  useEffect(() => {
    if (shouldStickToBottomRef.current) {
      bottomRef.current?.scrollIntoView({ block: 'end' })
    }
  }, [messages.length])

  if (isLoading) {
    return <MessageListSkeleton />
  }

  if (errorMessage) {
    return (
      <div
        className="flex min-h-[20rem] items-center border-2 border-[var(--color-accent)] px-5 py-6 sm:min-h-[28rem]"
        role="alert"
      >
        <div>
          <p className="font-mono text-xs uppercase tracking-[0.2em] text-[var(--color-accent)]">
            Failed to load messages
          </p>
          <p className="mt-3 text-sm font-semibold leading-6">{errorMessage}</p>
        </div>
      </div>
    )
  }

  return (
    <section className="relative border-2 border-[var(--color-ink)]" aria-label="Chat messages">
      <div
        className="max-h-[58vh] min-h-[20rem] overflow-y-auto px-4 py-4 sm:max-h-[62vh] sm:min-h-[28rem]"
        ref={scrollContainerRef}
        onScroll={handleScroll}
      >
        {hasMore ? (
          <div className="mb-5 flex justify-center">
            <button
              className="border-2 border-[var(--color-ink)] px-4 py-3 font-mono text-xs uppercase tracking-[0.18em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
              disabled={isLoadingOlder}
              type="button"
              onClick={() => void onLoadOlder()}
            >
              {isLoadingOlder ? 'Loading older' : 'Load older'}
            </button>
          </div>
        ) : null}

        {messages.length === 0 ? (
          <div className="flex min-h-[18rem] items-center justify-center text-center sm:min-h-[24rem]">
            <div>
              <p className="font-mono text-xs uppercase tracking-[0.24em] text-[var(--color-muted)]">
                Empty room
              </p>
              <p className="mt-3 text-lg font-black uppercase tracking-[-0.04em]">
                Send the first message.
              </p>
            </div>
          </div>
        ) : null}

        {groupedMessages.map(({ message, isGroupedWithPrevious }) => (
          <MessageItem
            isGroupedWithPrevious={isGroupedWithPrevious}
            isOwnMessage={message.senderId === currentUserId}
            key={message.id}
            message={message}
          />
        ))}

        <div ref={bottomRef} />
      </div>

      {!isNearBottom ? (
        <button
          aria-label="Jump to latest message"
          className="absolute bottom-4 right-4 border-2 border-[var(--color-ink)] bg-[var(--color-paper)] px-4 py-3 font-mono text-xs uppercase tracking-[0.18em] shadow-[5px_5px_0_var(--color-ink)] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
          type="button"
          onClick={() => {
            shouldStickToBottomRef.current = true
            bottomRef.current?.scrollIntoView({ block: 'end', behavior: 'smooth' })
          }}
        >
          Jump to latest
        </button>
      ) : null}
    </section>
  )
}

function MessageListSkeleton() {
  return (
    <div
      aria-label="Loading messages"
      className="min-h-[20rem] border-2 border-[var(--color-ink)] px-4 py-5 sm:min-h-[28rem]"
    >
      {Array.from({ length: 8 }, (_, index) => (
        <div className="grid grid-cols-[2.75rem_1fr] gap-3 py-3" key={index}>
          <Skeleton className="size-11 border-2" />
          <div>
            <Skeleton className="h-3 w-32" />
            <Skeleton className="mt-3 h-12 w-4/5" />
          </div>
        </div>
      ))}
    </div>
  )
}

function isGroupedWithPrevious(message: ChatMessage, previousMessage: ChatMessage | undefined) {
  if (!previousMessage) {
    return false
  }

  if (message.senderId !== previousMessage.senderId) {
    return false
  }

  const diffMs = Date.parse(message.createdAt) - Date.parse(previousMessage.createdAt)
  return Number.isFinite(diffMs) && diffMs >= 0 && diffMs <= groupingWindowMs
}
