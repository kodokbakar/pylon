import { fireEvent, render, screen } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { MessageList } from '../MessageList'
import type { ChatMessage } from '../../../hooks/useChatMessages'

describe('MessageList', () => {
  beforeEach(() => {
    window.HTMLElement.prototype.scrollIntoView = vi.fn()
  })

  it('renders one MessageItem per message', () => {
    const { container } = render(
      <MessageList
        currentUserId="user-1"
        errorMessage={null}
        hasMore={false}
        isLoading={false}
        isLoadingOlder={false}
        messages={[
          createMessage({
            id: 'message-1',
            senderId: 'user-1',
            senderName: 'Operator',
            content: 'Own message',
          }),
          createMessage({
            id: 'message-2',
            senderId: 'user-2',
            senderName: 'Teammate',
            content: 'Team message',
          }),
        ]}
        onLoadOlder={async () => undefined}
      />,
    )

    expect(container.querySelectorAll('article')).toHaveLength(2)
    expect(screen.getByText('Own message')).toBeInTheDocument()
    expect(screen.getByText('Team message')).toBeInTheDocument()
    expect(screen.getByText('You')).toBeInTheDocument()
    expect(screen.getByText('Teammate')).toBeInTheDocument()
  })

  it('renders loading skeleton state', () => {
    render(
      <MessageList
        currentUserId="user-1"
        errorMessage={null}
        hasMore={false}
        isLoading
        isLoadingOlder={false}
        messages={[]}
        onLoadOlder={async () => undefined}
      />,
    )

    expect(screen.getByLabelText('Loading messages')).toBeInTheDocument()
  })

  it('renders error state', () => {
    render(
      <MessageList
        currentUserId="user-1"
        errorMessage="Failed to load history."
        hasMore={false}
        isLoading={false}
        isLoadingOlder={false}
        messages={[]}
        onLoadOlder={async () => undefined}
      />,
    )

    expect(screen.getByRole('alert')).toHaveTextContent('Failed to load messages')
    expect(screen.getByRole('alert')).toHaveTextContent('Failed to load history.')
  })

  it('renders empty room state', () => {
    render(
      <MessageList
        currentUserId="user-1"
        errorMessage={null}
        hasMore={false}
        isLoading={false}
        isLoadingOlder={false}
        messages={[]}
        onLoadOlder={async () => undefined}
      />,
    )

    expect(screen.getByText('Empty room')).toBeInTheDocument()
    expect(screen.getByText('Send the first message.')).toBeInTheDocument()
  })

  it('calls onLoadOlder from the load older button', () => {
    const onLoadOlder = vi.fn().mockResolvedValue(undefined)

    render(
      <MessageList
        currentUserId="user-1"
        errorMessage={null}
        hasMore
        isLoading={false}
        isLoadingOlder={false}
        messages={[createMessage()]}
        onLoadOlder={onLoadOlder}
      />,
    )

    fireEvent.click(screen.getByRole('button', { name: 'Load older' }))

    expect(onLoadOlder).toHaveBeenCalledTimes(1)
  })

  it('disables the load older button while loading older messages', () => {
    render(
      <MessageList
        currentUserId="user-1"
        errorMessage={null}
        hasMore
        isLoading={false}
        isLoadingOlder
        messages={[createMessage()]}
        onLoadOlder={async () => undefined}
      />,
    )

    expect(screen.getByRole('button', { name: 'Loading older' })).toBeDisabled()
  })
})

function createMessage(overrides: Partial<ChatMessage> = {}): ChatMessage {
  return {
    id: 'message-1',
    roomId: 'room-1',
    senderId: 'user-1',
    senderName: 'Operator',
    senderUsername: 'operator',
    senderAvatarUrl: '',
    content: 'Message content',
    msgType: 'text',
    createdAt: '2026-07-01T00:00:00.000Z',
    status: 'sent',
    optimistic: false,
    ...overrides,
  }
}
