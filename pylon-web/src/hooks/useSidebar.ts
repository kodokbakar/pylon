import { useCallback, useEffect, useState } from 'react'

const SIDEBAR_STORAGE_KEY = 'sidebar-open'
const mobileSidebarQuery = '(max-width: 767px)'

export function useSidebar() {
  const [isMobile, setIsMobile] = useState(isMobileViewport)
  const [isOpen, setIsOpen] = useState(() => !isMobileViewport() || readStoredSidebarState())

  const open = useCallback(() => {
    writeSidebarState(true)
    setIsOpen(true)
  }, [])

  const close = useCallback(() => {
    if (isMobileViewport()) {
      writeSidebarState(false)
    }

    setIsOpen(false)
  }, [])

  const toggle = useCallback(() => {
    setIsOpen((current) => {
      const nextValue = !current

      if (isMobileViewport()) {
        writeSidebarState(nextValue)
      }

      return nextValue
    })
  }, [])

  const closeOnMobile = useCallback(() => {
    if (isMobileViewport()) {
      close()
    }
  }, [close])

  useEffect(() => {
    return subscribeToMobileViewport((nextIsMobile) => {
      setIsMobile(nextIsMobile)

      if (!nextIsMobile) {
        setIsOpen(true)
        return
      }

      setIsOpen(readStoredSidebarState())
    })
  }, [])

  useEffect(() => {
    if (!isOpen || !isMobile) {
      return
    }

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        close()
      }
    }

    document.addEventListener('keydown', handleKeyDown)

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [close, isMobile, isOpen])

  return {
    isOpen,
    isMobile,
    open,
    close,
    closeOnMobile,
    toggle,
  }
}

function readStoredSidebarState() {
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

function subscribeToMobileViewport(callback: (isMobile: boolean) => void) {
  const mediaQueryList = window.matchMedia(mobileSidebarQuery)

  callback(mediaQueryList.matches)

  function handleChange(event: MediaQueryListEvent) {
    callback(event.matches)
  }

  mediaQueryList.addEventListener('change', handleChange)

  return () => {
    mediaQueryList.removeEventListener('change', handleChange)
  }
}
