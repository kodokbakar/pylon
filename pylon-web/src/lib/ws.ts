export type WebSocketConnectionState =
  'connecting' | 'connected' | 'disconnected' | 'reconnecting' | 'error'

export type WebSocketMessage = {
  type: string
  timestamp?: number
  payload?: unknown
  [key: string]: unknown
}

export type WebSocketStateSnapshot = {
  state: WebSocketConnectionState
  reconnectAttempt: number
  error: string | null
}

export type WebSocketMessageHandler = (message: WebSocketMessage) => void

type WebSocketClientOptions = {
  url: string
  heartbeatIntervalMs?: number
  pongTimeoutMs?: number
  reconnectBaseDelayMs?: number
  reconnectMaxDelayMs?: number
  maxReconnectAttempts?: number
  onStateChange?: (snapshot: WebSocketStateSnapshot) => void
  onMessage?: WebSocketMessageHandler
  onError?: (error: Error) => void
}

const defaultHeartbeatIntervalMs = 30_000
const defaultPongTimeoutMs = 10_000
const defaultReconnectBaseDelayMs = 1_000
const defaultReconnectMaxDelayMs = 30_000
const defaultMaxReconnectAttempts = 10

export class PylonWebSocketClient {
  private socket: WebSocket | null = null
  private reconnectTimer: number | null = null
  private heartbeatTimer: number | null = null
  private pongTimeoutTimer: number | null = null
  private reconnectAttempt = 0
  private manualClose = false
  private readonly handlers = new Map<string, Set<WebSocketMessageHandler>>()

  private readonly options: WebSocketClientOptions

  constructor(options: WebSocketClientOptions) {
    this.options = options
  }

  connect() {
    this.manualClose = false
    this.openSocket(false)
  }

  disconnect() {
    this.manualClose = true
    this.clearReconnectTimer()
    this.clearHeartbeatTimers()

    const socket = this.socket
    this.socket = null

    if (socket && socket.readyState !== WebSocket.CLOSED) {
      socket.close(1000, 'client disconnected')
    }

    this.emitState('disconnected', null)
  }

  send(message: WebSocketMessage) {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      return false
    }

    const outgoingMessage = normalizeOutgoingMessage(message)
    this.socket.send(JSON.stringify(outgoingMessage))

    return true
  }

  subscribe(type: string, handler: WebSocketMessageHandler) {
    const normalizedType = type.trim()
    if (!normalizedType) {
      return () => undefined
    }

    const existingHandlers = this.handlers.get(normalizedType) ?? new Set<WebSocketMessageHandler>()
    existingHandlers.add(handler)
    this.handlers.set(normalizedType, existingHandlers)

    return () => {
      const currentHandlers = this.handlers.get(normalizedType)
      if (!currentHandlers) {
        return
      }

      currentHandlers.delete(handler)

      if (currentHandlers.size === 0) {
        this.handlers.delete(normalizedType)
      }
    }
  }

  private openSocket(isReconnect: boolean) {
    this.clearReconnectTimer()

    if (
      this.socket &&
      (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)
    ) {
      return
    }

    this.emitState(isReconnect ? 'reconnecting' : 'connecting', null)

    const socket = new WebSocket(this.options.url)
    this.socket = socket

    socket.onopen = () => {
      if (this.socket !== socket) {
        return
      }

      this.reconnectAttempt = 0
      this.emitState('connected', null)
      this.startHeartbeat()
    }

    socket.onmessage = (event) => {
      if (this.socket !== socket) {
        return
      }

      this.handleMessage(event.data)
    }

    socket.onerror = () => {
      if (this.socket !== socket) {
        return
      }

      this.options.onError?.(new Error('WebSocket connection error'))
    }

    socket.onclose = () => {
      if (this.socket === socket) {
        this.socket = null
      }

      this.clearHeartbeatTimers()

      if (this.manualClose) {
        this.emitState('disconnected', null)
        return
      }

      this.scheduleReconnect()
    }
  }

  private handleMessage(data: unknown) {
    if (typeof data !== 'string') {
      this.options.onError?.(new Error('Unsupported WebSocket payload type'))
      return
    }

    const message = parseMessage(data)
    if (!message) {
      this.options.onError?.(new Error('Invalid WebSocket JSON message'))
      return
    }

    if (message.type === 'pong') {
      this.clearPongTimeout()
    }

    this.options.onMessage?.(message)

    const typeHandlers = this.handlers.get(message.type)
    typeHandlers?.forEach((handler) => handler(message))

    const wildcardHandlers = this.handlers.get('*')
    wildcardHandlers?.forEach((handler) => handler(message))
  }

  private startHeartbeat() {
    this.clearHeartbeatTimers()

    this.heartbeatTimer = window.setInterval(() => {
      if (!this.send({ type: 'ping' })) {
        return
      }

      this.clearPongTimeout()
      this.pongTimeoutTimer = window.setTimeout(() => {
        this.options.onError?.(new Error('WebSocket heartbeat timed out'))
        this.socket?.close(4000, 'heartbeat timeout')
      }, this.getPongTimeoutMs())
    }, this.getHeartbeatIntervalMs())
  }

  private scheduleReconnect() {
    if (this.reconnectAttempt >= this.getMaxReconnectAttempts()) {
      this.emitState('error', 'WebSocket reconnect attempts exhausted')
      return
    }

    this.reconnectAttempt += 1
    this.emitState('reconnecting', null)

    const delayMs = this.getReconnectDelayMs(this.reconnectAttempt)
    this.reconnectTimer = window.setTimeout(() => {
      this.openSocket(true)
    }, delayMs)
  }

  private getReconnectDelayMs(attempt: number) {
    const baseDelay = this.options.reconnectBaseDelayMs ?? defaultReconnectBaseDelayMs
    const maxDelay = this.options.reconnectMaxDelayMs ?? defaultReconnectMaxDelayMs

    return Math.min(baseDelay * 2 ** Math.max(attempt - 1, 0), maxDelay)
  }

  private getHeartbeatIntervalMs() {
    return this.options.heartbeatIntervalMs ?? defaultHeartbeatIntervalMs
  }

  private getPongTimeoutMs() {
    return this.options.pongTimeoutMs ?? defaultPongTimeoutMs
  }

  private getMaxReconnectAttempts() {
    return this.options.maxReconnectAttempts ?? defaultMaxReconnectAttempts
  }

  private clearReconnectTimer() {
    if (this.reconnectTimer === null) {
      return
    }

    window.clearTimeout(this.reconnectTimer)
    this.reconnectTimer = null
  }

  private clearHeartbeatTimers() {
    if (this.heartbeatTimer !== null) {
      window.clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }

    this.clearPongTimeout()
  }

  private clearPongTimeout() {
    if (this.pongTimeoutTimer === null) {
      return
    }

    window.clearTimeout(this.pongTimeoutTimer)
    this.pongTimeoutTimer = null
  }

  private emitState(state: WebSocketConnectionState, error: string | null) {
    this.options.onStateChange?.({
      state,
      reconnectAttempt: this.reconnectAttempt,
      error,
    })
  }
}

function parseMessage(data: string): WebSocketMessage | null {
  try {
    const parsed: unknown = JSON.parse(data)
    if (!isRecord(parsed) || typeof parsed.type !== 'string') {
      return null
    }

    return parsed as WebSocketMessage
  } catch {
    return null
  }
}

function normalizeOutgoingMessage(message: WebSocketMessage): WebSocketMessage {
  const outgoingMessage: WebSocketMessage = {
    ...message,
    type: normalizeOutgoingType(message.type),
    timestamp: typeof message.timestamp === 'number' ? message.timestamp : Date.now(),
  }

  if (typeof outgoingMessage.roomId === 'string' && typeof outgoingMessage.room_id !== 'string') {
    outgoingMessage.room_id = outgoingMessage.roomId
  }

  if (typeof outgoingMessage.msgType === 'string' && typeof outgoingMessage.msg_type !== 'string') {
    outgoingMessage.msg_type = outgoingMessage.msgType
  }

  return outgoingMessage
}

function normalizeOutgoingType(type: string) {
  switch (type) {
    case 'chat':
    case 'chat.message':
      return 'message'
    default:
      return type
  }
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}
