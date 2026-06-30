import { Link, useParams } from 'react-router-dom'

import { MessageInput } from '../components/chat/MessageInput'
import { MessageList } from '../components/chat/MessageList'
import { TypingIndicator } from '../components/presence/TypingIndicator'
import { useChatMessages, type ChatMessage } from '../hooks/useChatMessages'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useRoom } from '../hooks/useRoom'
import { useRoomMembers, type RoomMemberListItem } from '../hooks/useRoomMembers'
import { useRoomPresence } from '../hooks/useRoomPresence'

export function ChatPage() {
  const { roomId } = useParams<{ roomId: string }>()
  const roomQuery = useRoom(roomId)
  const chat = useChatMessages(roomId)
  const membersQuery = useRoomMembers(roomId)
  const presence = useRoomPresence(roomId)
  const typingUserNames = getTypingUserNames(
    presence.typingUserIds,
    membersQuery.members,
    chat.messages,
  )

  const roomName = roomQuery.room?.name.trim() || 'Room chat'
  useDocumentTitle(`${roomName} / Pylon Chat`)

  return (
    <section className="pt-8 lg:pt-14">
      <header className="border-b border-[var(--color-line)] pb-4">
        <Link
          className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)] hover:text-[var(--color-accent)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
          to={`/rooms/${roomId ?? ''}`}
        >
          ← Back to room detail
        </Link>
      </header>

      <section className="pt-10 lg:pt-16">
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

            <TypingIndicator names={typingUserNames} />

            <MessageInput
              connectionState={chat.connectionState}
              disabled={!chat.canSend}
              sendError={chat.sendError}
              onSend={chat.sendMessage}
              onTyping={() => void presence.sendTyping()}
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

function getTypingUserNames(
  userIds: string[],
  members: RoomMemberListItem[],
  messages: ChatMessage[],
) {
  const namesByUserId = new Map<string, string>()

  for (const message of messages) {
    const name = message.senderName || message.senderUsername
    if (message.senderId && name) {
      namesByUserId.set(message.senderId, name)
    }
  }

  for (const member of members) {
    const name = member.name || member.username
    if (member.id && name) {
      namesByUserId.set(member.id, name)
    }
  }

  return userIds.map((userId) => namesByUserId.get(userId) ?? fallbackTypingName(userId))
}

function fallbackTypingName(userId: string) {
  const shortId = userId.trim().slice(0, 8)
  return shortId ? `User ${shortId}` : 'Someone'
}
