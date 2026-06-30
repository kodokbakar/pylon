import { useEffect } from 'react'

import { usePresence } from './usePresence'

export function useStreamPresence(roomId: string | undefined) {
  const { startRoomStream } = usePresence()
  const normalizedRoomId = roomId?.trim() ?? ''

  useEffect(() => {
    if (!normalizedRoomId) {
      return
    }

    return startRoomStream(normalizedRoomId)
  }, [normalizedRoomId, startRoomStream])
}
