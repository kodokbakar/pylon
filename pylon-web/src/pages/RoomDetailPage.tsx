import { Code, ConnectError } from '@connectrpc/connect'
import { Link, useNavigate, useParams } from 'react-router-dom'

import { RoomList } from '../components/room/RoomList'
import { useLeaveRoom } from '../hooks/useLeaveRoom'
import { useRoom } from '../hooks/useRoom'
import { useRoomMembers } from '../hooks/useRoomMembers'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { StatusIndicator } from '../components/presence/StatusIndicator'
import { formatPresenceStatus, getPresenceStatus, useRoomPresence } from '../hooks/useRoomPresence'

export function RoomDetailPage() {
  const navigate = useNavigate()
  const { roomId } = useParams<{ roomId: string }>()
  const roomQuery = useRoom(roomId)
  const membersQuery = useRoomMembers(roomId)
  const presence = useRoomPresence(roomId)
  const leaveRoom = useLeaveRoom(roomId)

  const roomName = roomQuery.room?.name.trim() || 'Room detail'
  useDocumentTitle(`${roomName} / Pylon Room`)

  const isLoading = roomQuery.isLoading || membersQuery.isLoading
  const accessError = getAccessError(roomQuery.error ?? membersQuery.error)

  async function handleLeaveRoom() {
    if (!roomId || leaveRoom.isPending) {
      return
    }

    const confirmed = window.confirm(`Leave ${roomName}? You will lose access to this room.`)
    if (!confirmed) {
      return
    }

    try {
      await leaveRoom.mutateAsync()
      navigate('/', { replace: true })
    } catch {
      // The mutation error is rendered below.
    }
  }

  return (
    <main className="min-h-screen bg-[var(--color-paper)] text-[var(--color-ink)]">
      <section className="mx-auto grid min-h-screen w-full max-w-7xl grid-cols-12 gap-x-4 px-5 py-6 sm:px-8 lg:px-10">
        <header className="col-span-12 border-b border-[var(--color-line)] pb-4">
          <Link
            className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)] hover:text-[var(--color-accent)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
            to="/"
          >
            ← Back to rooms
          </Link>
        </header>

        <div className="col-span-12 grid grid-cols-12 gap-x-4 pt-10 lg:pt-16">
          <aside className="col-span-12 lg:col-span-4">
            <RoomList />
          </aside>

          <section className="col-span-12 mt-12 lg:col-span-8 lg:mt-0 lg:pl-6">
            {isLoading ? <RoomDetailSkeleton /> : null}

            {!isLoading && accessError ? <RoomAccessState state={accessError} /> : null}

            {!isLoading && !accessError && roomQuery.room ? (
              <>
                <div className="grid grid-cols-12 gap-4">
                  <div className="col-span-12 xl:col-span-8">
                    <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
                      Room detail
                    </p>

                    <h1 className="text-[clamp(3.8rem,13vw,9rem)] font-black uppercase leading-[0.86] tracking-[-0.08em]">
                      {roomQuery.room.name || 'Untitled room'}
                    </h1>

                    <div className="mt-8 border-t border-[var(--color-line)] pt-5">
                      <p className="font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-muted)]">
                        Description
                      </p>
                      <p className="mt-3 max-w-2xl text-lg font-semibold leading-7 tracking-[-0.03em]">
                        No description is exposed by the RoomService proto yet.
                      </p>
                    </div>
                  </div>

                  <div className="col-span-12 border-2 border-[var(--color-ink)] xl:col-span-4">
                    <RoomStat label="Members" value={String(membersQuery.members.length)} />
                    <RoomStat
                      label="Created"
                      value={formatDate(roomQuery.room.createdAt?.toDate())}
                    />
                    <RoomStat label="Room ID" value={roomQuery.room.id} />
                    <RoomStat label="Online" value={String(presence.onlineCount)} />
                  </div>
                </div>

                <div className="mt-8 flex flex-col gap-3 border-y-2 border-[var(--color-ink)] py-5 sm:flex-row">
                  <Link
                    className="inline-flex flex-1 items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
                    to={`/rooms/${roomQuery.room.id}/chat`}
                  >
                    Start chat
                    <span className="ml-8">→</span>
                  </Link>

                  <button
                    className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] opacity-60"
                    disabled
                    type="button"
                  >
                    Settings
                    <span className="ml-8">Soon</span>
                  </button>

                  <button
                    className="inline-flex items-center justify-between border-2 border-[var(--color-accent)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-accent)] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
                    disabled={leaveRoom.isPending}
                    type="button"
                    onClick={() => void handleLeaveRoom()}
                  >
                    {leaveRoom.isPending ? 'Leaving room' : 'Leave room'}
                    <span className="ml-8">{leaveRoom.isPending ? '…' : '!'}</span>
                  </button>
                </div>

                {leaveRoom.isError ? (
                  <div
                    className="mt-5 border-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
                    role="alert"
                  >
                    {leaveRoom.error instanceof Error
                      ? leaveRoom.error.message
                      : 'Failed to leave room.'}
                  </div>
                ) : null}

                <section className="mt-10" aria-labelledby="members-title">
                  <div className="border-y-2 border-[var(--color-ink)] py-4">
                    <p className="font-mono text-xs uppercase tracking-[0.28em] text-[var(--color-muted)]">
                      Member directory
                    </p>
                    <h2
                      className="mt-2 text-3xl font-black uppercase leading-none tracking-[-0.05em]"
                      id="members-title"
                    >
                      {membersQuery.members.length} members
                    </h2>
                  </div>

                  {membersQuery.members.length === 0 ? (
                    <p className="py-8 text-sm leading-6 text-[var(--color-muted)]">
                      No members returned for this room.
                    </p>
                  ) : (
                    <div className="divide-y divide-[var(--color-line)]">
                      {membersQuery.members.map((member) => {
                        const memberStatus = getPresenceStatus(
                          presence.presencesByUserId,
                          member.id,
                        )

                        return (
                          <article
                            className="grid grid-cols-[3rem_1fr_auto] gap-4 py-4"
                            key={member.id}
                          >
                            <div className="relative size-12">
                              <div className="flex size-full items-center justify-center border-2 border-[var(--color-ink)] font-mono text-sm font-black uppercase">
                                {member.avatarUrl ? (
                                  <img
                                    alt=""
                                    className="size-full object-cover"
                                    src={member.avatarUrl}
                                  />
                                ) : (
                                  member.initial
                                )}
                              </div>

                              <StatusIndicator
                                className="absolute -bottom-1 -right-1 border-2 border-[var(--color-paper)] bg-[var(--color-paper)]"
                                label={`${member.name} is ${formatPresenceStatus(memberStatus)}`}
                                size="md"
                                status={memberStatus}
                              />
                            </div>

                            <div className="min-w-0">
                              <p className="truncate text-base font-black uppercase tracking-[-0.03em]">
                                {member.name}
                              </p>
                              <p className="mt-1 truncate font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-muted)]">
                                {member.username || member.id}
                              </p>
                            </div>

                            <div className="text-right">
                              <p className="font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]">
                                {member.role}
                              </p>
                              <p className="mt-2 font-mono text-[0.65rem] uppercase tracking-[0.14em] text-[var(--color-muted)]">
                                {formatPresenceStatus(memberStatus)}
                              </p>
                              <p className="mt-2 font-mono text-[0.65rem] uppercase tracking-[0.14em] text-[var(--color-muted)]">
                                {formatDateString(member.joinedAt)}
                              </p>
                            </div>
                          </article>
                        )
                      })}
                    </div>
                  )}
                </section>
              </>
            ) : null}
          </section>
        </div>
      </section>
    </main>
  )
}

function RoomStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="border-b border-[var(--color-line)] px-4 py-5">
      <p className="font-mono text-[0.65rem] uppercase tracking-[0.18em] text-[var(--color-muted)]">
        {label}
      </p>
      <p className="mt-2 break-all text-lg font-black tracking-[-0.03em]">{value}</p>
    </div>
  )
}

function RoomDetailSkeleton() {
  return (
    <div aria-label="Loading room detail">
      <div className="h-6 w-36 animate-pulse bg-[var(--color-grid)]" />
      <div className="mt-6 h-24 w-3/4 animate-pulse bg-[var(--color-grid)]" />
      <div className="mt-8 h-20 w-full animate-pulse bg-[var(--color-grid)]" />
      <div className="mt-10 space-y-4">
        {Array.from({ length: 4 }, (_, index) => (
          <div className="h-16 animate-pulse bg-[var(--color-grid)]" key={index} />
        ))}
      </div>
    </div>
  )
}

function RoomAccessState({ state }: { state: { label: string; title: string; message: string } }) {
  return (
    <div role="alert">
      <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
        {state.label}
      </p>

      <h1 className="text-[clamp(3.8rem,13vw,9rem)] font-black uppercase leading-[0.86] tracking-[-0.08em]">
        {state.title}
      </h1>

      <p className="mt-8 max-w-2xl border-t border-[var(--color-line)] pt-5 text-xl font-semibold leading-tight tracking-[-0.03em]">
        {state.message}
      </p>

      <Link
        className="mt-10 inline-flex items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
        to="/"
      >
        Return to rooms
        <span className="ml-8">→</span>
      </Link>
    </div>
  )
}

function getAccessError(error: unknown) {
  if (!error) {
    return null
  }

  const connectError = ConnectError.from(error)

  if (connectError.code === Code.NotFound) {
    return {
      label: '404',
      title: 'Room not found',
      message: 'This room does not exist or has been removed.',
    }
  }

  if (connectError.code === Code.PermissionDenied) {
    return {
      label: 'Unauthorized',
      title: 'No access',
      message: 'You are not a member of this room.',
    }
  }

  if (connectError.code === Code.Unauthenticated) {
    return {
      label: 'Session',
      title: 'Login required',
      message: 'Your session expired. Please log in again.',
    }
  }

  return {
    label: 'Room error',
    title: 'Unavailable',
    message: connectError.rawMessage || connectError.message || 'Room detail could not be loaded.',
  }
}

function formatDate(date: Date | undefined) {
  if (!date || Number.isNaN(date.getTime())) {
    return '—'
  }

  return new Intl.DateTimeFormat('en', {
    month: 'short',
    day: '2-digit',
    year: 'numeric',
  }).format(date)
}

function formatDateString(value: string | null) {
  if (!value) {
    return '—'
  }

  const date = new Date(value)
  return formatDate(date)
}
