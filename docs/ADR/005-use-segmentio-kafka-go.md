# ADR-005: Use segmentio/kafka-go for Event Streaming

## Status
Accepted

## Context
Pylon needs asynchronous event delivery for workflows such as:

- Chat message created events
- Notification generation
- Future event-driven features

The event system must work from Go services, support producers and consumers, and integrate with context cancellation.

## Decision
We will use Kafka for event streaming and `github.com/segmentio/kafka-go` as the Go client.

The Chat Service publishes message events to Kafka. The Notification Service consumes those events and creates notifications.

## Consequences

### Positive
- Kafka provides durable asynchronous event delivery.
- Producers and consumers are implemented in Go without a JVM dependency.
- `kafka-go` integrates with Go contexts.
- Event-driven flow decouples chat persistence from notification delivery.
- Consumer groups allow notification processing to scale horizontally.

### Negative
- Kafka adds local and production infrastructure complexity.
- Message schemas must be maintained carefully.
- Consumer retries and poison message handling need explicit design.

### Risks
- Duplicated messages can happen and should be handled idempotently where needed.
- Consumer lag must be monitored.
- Event versioning will be needed as event payloads evolve.

## Alternatives Considered

### Synchronous RPC
- Simpler flow for small systems.
- Notification failures could slow down or break message sending.
- Tightly couples Chat Service to Notification Service.

### Redis Pub/Sub
- Lightweight and simple.
- Not durable by default.
- Less suitable for reliable event processing.

### NATS
- Lightweight and fast.
- Adds another infrastructure choice.
- Kafka is already aligned with the project’s event streaming direction.