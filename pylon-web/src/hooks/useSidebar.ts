import { useCallback, useEffect, useState } from 'react'

const SIDEBAR_STORAGE_KEY = 'sidebar-open'
const mobileSidebarQuery = '(max-width: 767px)'

export function useSidebar() {
  const [isOpen, setIsOpen] = useState(readInitialSidebarState)

  const open = useCallback(() => {
    writeSidebarState(true)
    setIsOpen(true)
  }, [])

  const close = useCallback(() => {
    writeSidebarState(false)
    setIsOpen(false)
  }, [])

  const toggle = useCallback(() => {
    setIsOpen((current) => {
      const nextValue = !current
      writeSidebarState(nextValue)
      return nextValue
    })
  }, [])

  const closeOnMobile = useCallback(() => {
    if (isMobileViewport()) {
      close()
    }
  }, [close])

  useEffect(() => {
    if (!isOpen) {
      return
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape' && isMobileViewport()) {
        close()
      }
    }

    document.addEventListener('keydown', handleKeyDown)

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [close, isOpen])

  return {
    isOpen,
    open,
    close,
    closeOnMobile,
    toggle,
  }
}

function readInitialSidebarState() {
  try {
    return window.localStorage.getItem(SIDEBAR_STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

function writeSidebarState(isOpen: boolean) {
  try {
    window.localStorage.setItem(SIDEBAR_STORAGE_KEY, String(isOpen))
  } catch {
    // Sidebar persistence is best-effort.
  }
}

function isMobileViewport() {
  return window.matchMedia(mobileSidebarQuery).matches
}
