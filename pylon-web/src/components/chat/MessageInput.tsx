import type { FormEvent, KeyboardEvent } from 'react'
import { useRef, useState } from 'react'

type MessageInputProps = {
  disabled: boolean
  connectionState: string
  sendError: string | null
  onSend: (content: string) => boolean
  onTyping: () => void
}

const typingDebounceMs = 2_000
const maxMessageLength = 10_000

export function MessageInput({
  disabled,
  connectionState,
  sendError,
  onSend,
  onTyping,
}: MessageInputProps) {
  const [content, setContent] = useState('')
  const lastTypingAtRef = useRef(0)

  function handleSubmit(event?: FormEvent<HTMLFormElement>) {
    event?.preventDefault()

    const sent = onSend(content)
    if (sent) {
      setContent('')
    }
  }

  function notifyTyping() {
    if (disabled) {
      return
    }

    const now = Date.now()
    if (now - lastTypingAtRef.current < typingDebounceMs) {
      return
    }

    lastTypingAtRef.current = now
    onTyping()
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (shouldNotifyTyping(event)) {
      notifyTyping()
    }

    if (event.key !== 'Enter' || event.shiftKey) {
      return
    }

    event.preventDefault()
    handleSubmit()
  }

  const isContentEmpty = content.trim().length === 0
  const isTooLong = content.length > maxMessageLength
  const descriptionIds = ['chat-message-counter', isTooLong ? 'chat-message-length-error' : '']
    .filter(Boolean)
    .join(' ')

  return (
    <form
      aria-label="Send message"
      className="border-x-2 border-b-2 border-[var(--color-ink)]"
      onSubmit={handleSubmit}
    >
      {sendError ? (
        <div
          className="border-b-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
          role="alert"
        >
          {sendError}
        </div>
      ) : null}

      <div className="grid grid-cols-1 gap-0 lg:grid-cols-[1fr_auto]">
        <label className="sr-only" htmlFor="chat-message-input">
          Message
        </label>

        <textarea
          aria-describedby={descriptionIds}
          aria-invalid={isTooLong}
          className="min-h-28 resize-none bg-transparent px-4 py-4 text-base font-semibold leading-6 outline-none placeholder:text-[var(--color-muted)] focus:bg-[var(--color-grid)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-[-2px] focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60"
          disabled={disabled}
          id="chat-message-input"
          maxLength={maxMessageLength + 100}
          placeholder={disabled ? `Realtime is ${connectionState}` : 'Write a message…'}
          value={content}
          onChange={(event) => setContent(event.target.value)}
          onKeyDown={handleKeyDown}
        />

        <div className="flex flex-col gap-3 border-t-2 border-[var(--color-ink)] p-3 sm:flex-row sm:items-end sm:justify-between lg:block lg:border-l-2 lg:border-t-0">
          <p
            aria-live="polite"
            className="font-mono text-[0.65rem] uppercase tracking-[0.16em] text-[var(--color-muted)]"
            id="chat-message-counter"
          >
            Enter sends · Shift+Enter newline · {content.length}/{maxMessageLength}
          </p>

          <button
            className="mt-0 inline-flex min-w-36 items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0 lg:mt-3"
            disabled={disabled || isContentEmpty || isTooLong}
            type="submit"
          >
            Send
            <span className="ml-8">↗</span>
          </button>
        </div>
      </div>

      {isTooLong ? (
        <p
          className="border-t-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
          id="chat-message-length-error"
          role="alert"
        >
          Message is too long.
        </p>
      ) : null}
    </form>
  )
}

function shouldNotifyTyping(event: KeyboardEvent<HTMLTextAreaElement>) {
  if (event.ctrlKey || event.metaKey || event.altKey) {
    return false
  }

  return event.key.length === 1 || event.key === 'Backspace' || event.key === 'Delete'
}
