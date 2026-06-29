import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { useRooms } from '../../hooks/useRooms'
import { CreateRoomModal } from './CreateRoomModal'
import { RoomItem } from './RoomItem'

export function RoomList() {
  const navigate = useNavigate()
  const { id: activeRoomId } = useParams<{ id: string }>()
  const { rooms, isLoading, isError, error, refetch, missingUser } = useRooms()
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)

  return (
    <section
      aria-labelledby="room-list-title"
      className="border-y-2 border-[var(--color-ink)] bg-[var(--color-paper)]"
    >
      <div className="flex items-start justify-between gap-3 border-b border-[var(--color-line)] px-3 py-4">
        <div>
          <p className="font-mono text-[0.68rem] uppercase tracking-[0.24em] text-[var(--color-muted)]">
            Sidebar
          </p>
          <h2
            className="mt-1 text-2xl font-black uppercase leading-none tracking-[-0.05em]"
            id="room-list-title"
          >
            Rooms
          </h2>
        </div>

        <div className="flex flex-col items-end gap-2">
          <span className="border border-[var(--color-ink)] px-2 py-1 font-mono text-[0.65rem] uppercase tracking-[0.16em]">
            {rooms.length}
          </span>

          <button
            className="border-2 border-[var(--color-ink)] px-3 py-2 font-mono text-[0.65rem] uppercase tracking-[0.16em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
            type="button"
            onClick={() => setIsCreateModalOpen(true)}
          >
            Create room
          </button>
        </div>
      </div>

      {isLoading ? <RoomListSkeleton /> : null}

      {!isLoading && missingUser ? (
        <RoomListMessage message="Your auth session is missing user data. Log out and log in again." />
      ) : null}

      {!isLoading && isError ? (
        <div className="px-3 py-6" role="alert">
          <p className="font-mono text-xs uppercase tracking-[0.18em] text-[var(--color-accent)]">
            Failed to load rooms
          </p>

          <p className="mt-3 text-sm leading-6 text-[var(--color-muted)]">
            {error instanceof Error ? error.message : 'Room service is unavailable.'}
          </p>

          <button
            className="mt-5 border-2 border-[var(--color-ink)] px-4 py-3 font-mono text-xs uppercase tracking-[0.18em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
            type="button"
            onClick={() => void refetch()}
          >
            Retry
          </button>
        </div>
      ) : null}

      {!isLoading && !missingUser && !isError && rooms.length === 0 ? (
        <RoomListMessage message="No rooms yet. Create one to start chatting." />
      ) : null}

      {!isLoading && !missingUser && !isError && rooms.length > 0 ? (
        <nav aria-label="Rooms">
          {rooms.map((room) => (
            <RoomItem
              isActive={room.id === activeRoomId}
              key={room.id}
              room={room}
              onClick={() => navigate(`/rooms/${room.id}`)}
            />
          ))}
        </nav>
      ) : null}

      <CreateRoomModal
        existingRoomNames={rooms.map((room) => room.name)}
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
      />
    </section>
  )
}

function RoomListSkeleton() {
  return (
    <div aria-label="Loading rooms" className="divide-y divide-[var(--color-line)]">
      {Array.from({ length: 5 }, (_, index) => (
        <div className="grid grid-cols-[3rem_1fr_auto] gap-3 px-3 py-4" key={index}>
          <div className="size-12 animate-pulse border-2 border-[var(--color-line)] bg-[var(--color-grid)]" />
          <div className="space-y-3 py-1">
            <div className="h-4 w-3/4 animate-pulse bg-[var(--color-grid)]" />
            <div className="h-3 w-1/2 animate-pulse bg-[var(--color-grid)]" />
          </div>
          <div className="h-3 w-10 animate-pulse bg-[var(--color-grid)]" />
        </div>
      ))}
    </div>
  )
}

function RoomListMessage({ message }: { message: string }) {
  return (
    <div className="px-3 py-8">
      <p className="max-w-xs text-sm leading-6 text-[var(--color-muted)]">{message}</p>
    </div>
  )
}
