import { Link, useParams } from 'react-router-dom'

import { RoomList } from '../components/room/RoomList'
import { useDocumentTitle } from '../hooks/useDocumentTitle'
import { useWebSocketContext } from '../contexts/webSocketContext'

export function RoomPage() {
  const { roomId } = useParams<{ roomId: string }>()
  const webSocket = useWebSocketContext()

  useDocumentTitle(`Pylon Chat ${roomId ?? ''}`.trim())

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
            <p className="mb-5 inline-flex border border-[var(--color-ink)] px-3 py-1 font-mono text-xs uppercase tracking-[0.26em]">
              Chat shell
            </p>

            <h1 className="text-[clamp(3.8rem,13vw,9rem)] font-black uppercase leading-[0.86] tracking-[-0.08em]">
              Room {roomId}
            </h1>

            <div className="mt-8 grid max-w-3xl grid-cols-2 border-2 border-[var(--color-ink)]">
              <RealtimeStat label="Realtime" value={webSocket.state} />
              <RealtimeStat
                label="Reconnect"
                value={`${webSocket.reconnectAttempt}/${webSocket.maxReconnectAttempts}`}
              />
            </div>

            {webSocket.error ? (
              <div
                className="mt-5 border-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
                role="alert"
              >
                {webSocket.error}
              </div>
            ) : null}

            <div className="mt-8 flex flex-col gap-3 sm:flex-row">
              <button
                className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
                disabled={!webSocket.isConnected || !roomId}
                type="button"
                onClick={() => {
                  if (!roomId) {
                    return
                  }

                  webSocket.send({
                    type: 'join',
                    room_id: roomId,
                  })
                }}
              >
                Join realtime room
                <span className="ml-8">↗</span>
              </button>

              <button
                className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
                disabled={!webSocket.isConnected || !roomId}
                type="button"
                onClick={() => {
                  if (!roomId) {
                    return
                  }

                  webSocket.send({
                    type: 'typing',
                    room_id: roomId,
                  })
                }}
              >
                Send typing probe
                <span className="ml-8">⌁</span>
              </button>
            </div>

            <p className="mt-8 max-w-2xl border-t border-[var(--color-line)] pt-5 text-xl font-semibold leading-tight tracking-[-0.03em]">
              Chat streaming will attach here in the next issue. The WebSocket foundation is active
              and exposes connection state, heartbeat, reconnect, and send routing.
            </p>
          </section>
        </div>
      </section>
    </main>
  )
}

function RealtimeStat({ label, value }: { label: string; value: string }) {
  return (
    <div className="border-b border-r border-[var(--color-line)] px-4 py-5">
      <p className="font-mono text-[0.65rem] uppercase tracking-[0.18em] text-[var(--color-muted)]">
        {label}
      </p>
      <p className="mt-2 break-all text-lg font-black uppercase tracking-[-0.03em]">{value}</p>
    </div>
  )
}
