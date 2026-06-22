# ADR-007: Use go-redis for Presence Cache

## Status
Accepted

## Context
Pylon needs fast, short-lived state for:

- Online status
- Last seen timestamps
- Typing indicators
- Room-level presence

This data changes frequently and does not need the same durability guarantees as messages or rooms.

## Decision
We will use Redis for presence and cache-like state, accessed through `github.com/redis/go-redis/v9`.

Presence state uses TTLs so stale online and typing states expire automatically.

## Consequences

### Positive
- Redis is fast for short-lived presence data.
- TTLs match online and typing semantics.
- Sorted sets and sets support room-level presence tracking.
- `go-redis` provides context-aware Go APIs.
- Keeps high-churn presence writes away from PostgreSQL.

### Negative
- Redis state is less durable than PostgreSQL.
- Presence behavior depends on TTL tuning.
- Multi-instance consistency needs careful key design.

### Risks
- Too-short TTLs can show users offline too aggressively.
- Too-long TTLs can leave stale online state.
- Redis outages affect presence and rate limiting behavior.

## Alternatives Considered

### PostgreSQL
- Durable and already used.
- Poor fit for high-frequency presence heartbeats.
- More write load on the primary database.

### In-Memory State
- Very fast and simple for one instance.
- Breaks in multi-instance deployments.
- State is lost on restart.

### etcd
- Strong consistency.
- More operational complexity.
- Not ideal for high-churn typing and presence state.