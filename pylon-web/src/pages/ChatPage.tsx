import { Link, useParams } from 'react-router-dom'

import { MessageInput } from '../components/chat/MessageInput'
import { MessageList } from '../components/chat/MessageList'
import { RoomList } from '../components/room/RoomList'
import { useChatMessages } from '../hooks/useChatMessages'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useRoom } from '../hooks/useRoom'

export function ChatPage() {
  const { roomId } = useParams<{ roomId: string }>()
  const roomQuery = useRoom(roomId)
  const chat = useChatMessages(roomId)

  const roomName = roomQuery.room?.name.trim() || 'Room chat'
  useDocumentTitle(`${roomName} / Pylon Chat`)

  return (
    <main className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
        <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
          <Link
            className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)] hover:text-[var(--color-accent)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
            to={`/rooms/${roomId ?? ''}`}
          >
            ← Back to room detail
          </Link>
        </header>

        <div className="col-span-12 grid grid-cols-12 gap-x-4 pt-10 lg:pt-16">
          <aside className="col-span-12 lg:col-span-4">
            <RoomList />
          </aside>

          <section className="col-span-12 mt-12 lg:col-span-8 lg:mt-0 lg:pl-6">
            <div className="mb-5 flex flex-col justify-between gap-4 border-y-2 border-[var(--color-ink)] py-4 sm:flex-row sm:items-end">
              <div>
                <p className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)]">
                  Live room
                </p>

                <h1 className="mt-2 text-4xl font-black uppercase leading-none tracking-[-0.06em] sm:text-6xl">
                  {roomName}
                </h1>
              </div>

              <div className="grid grid-cols-2 border-2 border-[var(--color-ink)]">
                <RealtimeStat label="Realtime" value={chat.connectionState} />
                <RealtimeStat label="Messages" value={String(chat.messages.length)} />
              </div>
            </div>

            {!roomId ? (
              <div className="border-2 border-[var(--color-accent)] px-5 py-6" role="alert">
                <p className="font-mono text-xs uppercase tracking-[0.2em] text-[var(--color-accent)]">
                  Missing room id
                </p>
              </div>
            ) : (
              <>
                <MessageList
                  currentUserId={chat.currentUserId}
                  errorMessage={chat.errorMessage}
                  hasMore={chat.hasMore}
                  isLoading={chat.isLoading}
                  isLoadingOlder={chat.isLoadingOlder}
                  messages={chat.messages}
                  onLoadOlder={chat.loadOlder}
                />

                <MessageInput
                  connectionState={chat.connectionState}
                  disabled={!chat.canSend}
                  sendError={chat.sendError}
                  onSend={chat.sendMessage}
                />
              </>
            )}
          </section>
        </div>
      </section>
    </main>
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
