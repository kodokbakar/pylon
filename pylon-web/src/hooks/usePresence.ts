import { useContext } from 'react'

import { PresenceContext } from '../context/presenceContext'

export function usePresence() {
  const context = useContext(PresenceContext)

  if (!context) {
    throw new Error('usePresence must be used within PresenceProvider')
  }

  return context
}
