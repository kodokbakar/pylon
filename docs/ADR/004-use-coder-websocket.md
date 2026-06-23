# ADR-004: Use coder/websocket for Real-Time Connections

## Status
Accepted

## Context
Pylon needs real-time bidirectional communication for chat clients.

The API Gateway must support:

- WebSocket connection upgrades
- Authenticated connections
- Room join and leave events
- Message fan-out
- Typing events
- Graceful shutdown and connection cleanup

## Decision
We will use `github.com/coder/websocket` for WebSocket support in the API Gateway.

The API Gateway owns public WebSocket connections and forwards business operations to internal services.

## Consequences

### Positive
- Small and focused WebSocket package.
- Context-based reads and writes fit Go service patterns.
- Works cleanly with `net/http`.
- Keeps WebSocket handling inside the API Gateway.
- Avoids adding a large web framework just for WebSocket support.

### Negative
- WebSocket behavior must be tested carefully.
- Backpressure and send-buffer management are application responsibilities.
- Horizontal scaling requires external fan-out support in future iterations.

### Risks
- Long-lived connections can leak if shutdown and cleanup are not handled correctly.
- A single API Gateway instance only sees its own local connections.
- Future multi-instance fan-out may require Redis pub/sub, Kafka fan-out, or a dedicated gateway coordination layer.

## Alternatives Considered

### gorilla/websocket
- Popular and widely used.
- The old Gorilla project history creates maintenance concerns.
- API style is less context-first.

### nhooyr/websocket
- Strong design and widely known.
- `coder/websocket` is the maintained successor path and fits current needs.

### Server-Sent Events
- Simpler for server-to-client streams.
- Does not support bidirectional client-to-server messaging.
- Not enough for full chat interaction.