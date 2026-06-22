# ADR-001: Use Microservice Architecture

## Status
Accepted

## Context
Pylon is a distributed real-time chat system that needs:

- API Gateway for HTTP and WebSocket traffic
- Chat Service for message persistence and message events
- Presence Service for online, offline, and typing status
- Room Service for room and membership management
- Notification Service for asynchronous notification delivery
- Independent scaling for services with different workloads

A single process would be simpler, but chat, presence, room, and notification workloads have different scaling and storage needs.

## Decision
We will use a microservice architecture with five services:

1. API Gateway
2. Chat Service
3. Presence Service
4. Room Service
5. Notification Service

The API Gateway exposes public HTTP and WebSocket entry points. Internal services communicate through Connect-Go RPC over HTTP. Kafka is used for asynchronous message events.

## Consequences

### Positive
- Services can be developed and deployed independently.
- Each service can scale based on its own workload.
- Failures can be isolated more clearly.
- Storage choices can match service needs, such as Redis for presence and PostgreSQL for durable data.
- The architecture maps well to Kubernetes deployment units.

### Negative
- More operational complexity than a monolith.
- More network calls between components.
- Distributed tracing, metrics, and logs become necessary.
- Local development needs more infrastructure.

### Risks
- Service boundaries can become too granular if not managed carefully.
- Network failures must be handled explicitly.
- Data consistency across services is harder than in a single database transaction.

## Alternatives Considered

### Monolith
- Simpler to develop, test, and deploy at the beginning.
- No internal network latency.
- Harder to scale chat, presence, and notification workloads independently.
- Harder to keep ownership boundaries clean as the system grows.

### Modular Monolith
- Good middle ground for early-stage projects.
- Lower operational overhead than microservices.
- Still requires future extraction work when independent scaling becomes necessary.

### Serverless
- Managed scaling and lower infrastructure management.
- Cold starts are risky for real-time chat.
- WebSocket and event processing flows become vendor-dependent.