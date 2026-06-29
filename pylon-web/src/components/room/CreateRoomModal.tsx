import type { ChangeEvent, FormEvent } from 'react'
import { useCallback, useEffect, useId, useMemo, useRef, useState } from 'react'
import { Code, ConnectError } from '@connectrpc/connect'
import { useNavigate } from 'react-router-dom'

import { useCreateRoom } from '../../hooks/useCreateRoom'
import { cleanBackendMessage } from '../../utils/backendError'

type CreateRoomModalProps = {
  isOpen: boolean
  existingRoomNames: string[]
  onClose: () => void
}

const maxRoomNameLength = 50
const focusableSelector = [
  'button:not([disabled])',
  'textarea:not([disabled])',
  'input:not([disabled])',
  'select:not([disabled])',
  'a[href]',
  '[tabindex]:not([tabindex="-1"])',
].join(',')

export function CreateRoomModal({ isOpen, existingRoomNames, onClose }: CreateRoomModalProps) {
  const navigate = useNavigate()
  const titleId = useId()
  const descriptionId = useId()
  const dialogRef = useRef<HTMLDivElement>(null)
  const nameInputRef = useRef<HTMLInputElement>(null)
  const createRoom = useCreateRoom()

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [nameError, setNameError] = useState<string | null>(null)
  const [submitError, setSubmitError] = useState<string | null>(null)

  const existingNameSet = useMemo(
    () => new Set(existingRoomNames.map(normalizeRoomName).filter(Boolean)),
    [existingRoomNames],
  )

  const closeModal = useCallback(() => {
    setName('')
    setDescription('')
    setNameError(null)
    setSubmitError(null)
    createRoom.reset()
    onClose()
  }, [createRoom, onClose])

  useEffect(() => {
    if (!isOpen) {
      return
    }

    const previouslyFocused =
      document.activeElement instanceof HTMLElement ? document.activeElement : null
    const previousOverflow = document.body.style.overflow

    document.body.style.overflow = 'hidden'
    window.requestAnimationFrame(() => {
      nameInputRef.current?.focus()
    })

    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        event.preventDefault()
        closeModal()
        return
      }

      if (event.key === 'Tab') {
        trapFocus(event, dialogRef.current)
      }
    }

    document.addEventListener('keydown', handleKeyDown)

    return () => {
      document.removeEventListener('keydown', handleKeyDown)
      document.body.style.overflow = previousOverflow
      previouslyFocused?.focus()
    }
  }, [closeModal, isOpen])

  if (!isOpen) {
    return null
  }

  function handleNameChange(event: ChangeEvent<HTMLInputElement>) {
    setName(event.target.value)
    setNameError(null)
    setSubmitError(null)
  }

  function handleDescriptionChange(event: ChangeEvent<HTMLTextAreaElement>) {
    setDescription(event.target.value)
    setSubmitError(null)
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const trimmedName = name.trim()
    const validationError = validateRoomName(trimmedName, existingNameSet)
    if (validationError) {
      setNameError(validationError)
      return
    }

    setNameError(null)
    setSubmitError(null)

    try {
      const room = await createRoom.mutateAsync({
        name: trimmedName,
        description: description.trim() || undefined,
      })

      closeModal()
      navigate(`/rooms/${room.id}`)
    } catch (error) {
      setSubmitError(mapCreateRoomError(error))
    }
  }

  return (
    <div
      aria-labelledby={titleId}
      aria-modal="true"
      className="fixed inset-0 z-50 flex items-center justify-center bg-[color-mix(in_srgb,var(--color-ink)_72%,transparent)] px-4 py-6"
      role="dialog"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) {
          closeModal()
        }
      }}
    >
      <div
        className="w-full max-w-xl border-2 border-[var(--color-ink)] bg-[var(--color-paper)] shadow-[10px_10px_0_var(--color-ink)]"
        ref={dialogRef}
        tabIndex={-1}
      >
        <div className="flex items-start justify-between border-b-2 border-[var(--color-ink)] px-5 py-4">
          <div>
            <p className="font-mono text-[0.68rem] uppercase tracking-[0.24em] text-[var(--color-muted)]">
              Room factory
            </p>
            <h2
              className="mt-2 text-4xl font-black uppercase leading-none tracking-[-0.06em]"
              id={titleId}
            >
              Create Room
            </h2>
          </div>

          <button
            aria-label="Close create room modal"
            className="border-2 border-[var(--color-ink)] px-3 py-2 font-mono text-xs uppercase tracking-[0.16em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
            type="button"
            onClick={closeModal}
          >
            X
          </button>
        </div>

        <form className="space-y-5 px-5 py-5" noValidate onSubmit={handleSubmit}>
          <p className="text-sm leading-6 text-[var(--color-muted)]" id={descriptionId}>
            Create a group room and jump straight into the chat shell. Description is collected for
            the UI, but it is not persisted until the RoomService proto adds a description field.
          </p>

          {submitError ? (
            <div
              className="border-2 border-[var(--color-accent)] px-4 py-3 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
              role="alert"
            >
              {submitError}
            </div>
          ) : null}

          <div>
            <div className="flex items-center justify-between gap-4">
              <label
                className="font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-muted)]"
                htmlFor="create-room-name"
              >
                Room name
              </label>

              <span className="font-mono text-[0.65rem] uppercase tracking-[0.16em] text-[var(--color-muted)]">
                {name.trim().length}/{maxRoomNameLength}
              </span>
            </div>

            <input
              aria-describedby={nameError ? 'create-room-name-error' : descriptionId}
              aria-invalid={Boolean(nameError)}
              autoComplete="off"
              className="mt-3 w-full border-2 border-[var(--color-ink)] bg-transparent px-4 py-4 text-base font-semibold outline-none transition-colors placeholder:text-[var(--color-muted)] focus:bg-[var(--color-grid)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60"
              disabled={createRoom.isPending}
              id="create-room-name"
              maxLength={maxRoomNameLength + 10}
              name="name"
              placeholder="Engineering war room"
              ref={nameInputRef}
              value={name}
              onChange={handleNameChange}
            />

            {nameError ? (
              <p
                className="mt-2 font-mono text-xs uppercase tracking-[0.16em] text-[var(--color-accent)]"
                id="create-room-name-error"
              >
                {nameError}
              </p>
            ) : null}
          </div>

          <div>
            <label
              className="font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-muted)]"
              htmlFor="create-room-description"
            >
              Description
            </label>

            <textarea
              className="mt-3 min-h-28 w-full resize-none border-2 border-[var(--color-ink)] bg-transparent px-4 py-4 text-base font-semibold outline-none transition-colors placeholder:text-[var(--color-muted)] focus:bg-[var(--color-grid)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60"
              disabled={createRoom.isPending}
              id="create-room-description"
              name="description"
              placeholder="Optional note for this room"
              value={description}
              onChange={handleDescriptionChange}
            />
          </div>

          <div className="flex flex-col gap-3 sm:flex-row">
            <button
              className="inline-flex flex-1 items-center justify-between border-2 border-[var(--color-ink)] bg-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] text-[var(--color-paper)] transition-transform duration-200 hover:-translate-y-1 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)] disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
              disabled={createRoom.isPending}
              type="submit"
            >
              {createRoom.isPending ? 'Creating room' : 'Create room'}
              <span className="ml-8">{createRoom.isPending ? '…' : '↗'}</span>
            </button>

            <button
              className="inline-flex items-center justify-between border-2 border-[var(--color-ink)] px-5 py-4 font-mono text-xs uppercase tracking-[0.22em] transition-transform duration-200 hover:-translate-y-1 hover:bg-[var(--color-accent)] hover:text-[var(--color-paper)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-[var(--color-accent)]"
              type="button"
              onClick={closeModal}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

function validateRoomName(name: string, existingNameSet: Set<string>) {
  if (name === '') {
    return 'Room name is required'
  }

  if (name.length > maxRoomNameLength) {
    return `Room name must be ${maxRoomNameLength} characters or fewer`
  }

  if (existingNameSet.has(normalizeRoomName(name))) {
    return 'A room with this name already exists'
  }

  return null
}

function normalizeRoomName(name: string) {
  return name.trim().toLowerCase()
}

function trapFocus(event: KeyboardEvent, dialog: HTMLDivElement | null) {
  if (!dialog) {
    return
  }

  const focusableElements = Array.from(dialog.querySelectorAll<HTMLElement>(focusableSelector))
    .filter((element) => !element.hasAttribute('disabled'))
    .filter((element) => element.getAttribute('aria-hidden') !== 'true')

  if (focusableElements.length === 0) {
    event.preventDefault()
    dialog.focus()
    return
  }

  const firstElement = focusableElements[0]
  const lastElement = focusableElements[focusableElements.length - 1]

  if (event.shiftKey && document.activeElement === firstElement) {
    event.preventDefault()
    lastElement.focus()
    return
  }

  if (!event.shiftKey && document.activeElement === lastElement) {
    event.preventDefault()
    firstElement.focus()
  }
}

function mapCreateRoomError(error: unknown) {
  if (error instanceof Error && !(error instanceof ConnectError)) {
    return error.message
  }

  const connectError = ConnectError.from(error)

  if (connectError.code === Code.InvalidArgument) {
    return (
      cleanBackendMessage(connectError.rawMessage || connectError.message) ||
      'Room input is invalid.'
    )
  }

  if (connectError.code === Code.AlreadyExists) {
    return 'A room with this name already exists.'
  }

  if (connectError.code === Code.Unauthenticated) {
    return 'Your session expired. Please log in again.'
  }

  if (connectError.code === Code.PermissionDenied) {
    return 'You do not have permission to create rooms.'
  }

  if (connectError.code === Code.Unavailable) {
    return 'Room service is unavailable. Please try again.'
  }

  return 'Failed to create room. Please try again.'
}
