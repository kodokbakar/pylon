import { ChatLayout } from '../components/layout/ChatLayout'
import { useChatMessages, type ChatMessage } from '../hooks/useChatMessages'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useRoom } from '../hooks/useRoom'
import { useRoomMembers, type RoomMemberListItem } from '../hooks/useRoomMembers'
import { useRoomPresence } from '../hooks/useRoomPresence'
import { useParams } from 'react-router-dom'

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
    <ChatLayout
      canSend={chat.canSend}
      connectionState={chat.connectionState}
      currentUserId={chat.currentUserId}
      errorMessage={chat.errorMessage}
      hasMore={chat.hasMore}
      isLoading={chat.isLoading}
      isLoadingOlder={chat.isLoadingOlder}
      messageCount={chat.messages.length}
      messages={chat.messages}
      roomId={roomId}
      roomName={roomName}
      sendError={chat.sendError}
      typingUserNames={typingUserNames}
      onLoadOlder={chat.loadOlder}
      onSend={chat.sendMessage}
      onTyping={() => void presence.sendTyping()}
    />
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
