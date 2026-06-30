import { MessageInput } from '../chat/MessageInput'
import { MessageList } from '../chat/MessageList'
import { TypingIndicator } from '../presence/TypingIndicator'
import type { ChatMessage } from '../../hooks/useChatMessages'
import { PageHeader } from './PageHeader'

type ChatLayoutProps = {
  roomId: string | undefined
  roomName: string
  connectionState: string
  messageCount: number
  messages: ChatMessage[]
  currentUserId: string
  isLoading: boolean
  errorMessage: string | null
  hasMore: boolean
  isLoadingOlder: boolean
  sendError: string | null
  canSend: boolean
  typingUserNames: string[]
  onLoadOlder: () => Promise<void>
  onSend: (content: string) => boolean
  onTyping: () => void
}

export function ChatLayout({
  roomId,
  roomName,
  connectionState,
  messageCount,
  messages,
  currentUserId,
  isLoading,
  errorMessage,
  hasMore,
  isLoadingOlder,
  sendError,
  canSend,
  typingUserNames,
  onLoadOlder,
  onSend,
  onTyping,
}: ChatLayoutProps) {
  return (
    <section className="pt-8 lg:pt-14">
      <PageHeader
        backLabel="← Back to room detail"
        backTo={`/rooms/${roomId ?? ''}`}
        eyebrow="Live room"
        title={roomName}
        actions={
          <div className="grid grid-cols-2 border-2 border-[var(--color-ink)]">
            <RealtimeStat label="Realtime" value={connectionState} />
            <RealtimeStat label="Messages" value={String(messageCount)} />
          </div>
        }
      />

      <section className="pt-10 lg:pt-16">
        {!roomId ? (
          <div className="border-2 border-[var(--color-accent)] px-5 py-6" role="alert">
            <p className="font-mono text-xs uppercase tracking-[0.2em] text-[var(--color-accent)]">
              Missing room id
            </p>
          </div>
        ) : (
          <>
            <MessageList
              currentUserId={currentUserId}
              errorMessage={errorMessage}
              hasMore={hasMore}
              isLoading={isLoading}
              isLoadingOlder={isLoadingOlder}
              messages={messages}
              onLoadOlder={onLoadOlder}
            />

            <TypingIndicator names={typingUserNames} />

            <MessageInput
              connectionState={connectionState}
              disabled={!canSend}
              sendError={sendError}
              onSend={onSend}
              onTyping={onTyping}
            />
          </>
        )}
      </section>
    </section>
  )
}

function RealtimeStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="min-w-32 border-r border-[var(--color-line)] px-4 py-3">
      <p className="font-mono text-[0.65rem] uppercase tracking-[0.18em] text-[var(--color-muted)]">
        {label}
      </p>
      <p className="mt-1 break-all text-sm font-black uppercase tracking-[-0.03em]">{value}</p>
    </div>
  )
}
