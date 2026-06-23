# ADR-002: Use Go as Primary Language

## Status
Accepted

## Context
Pylon needs a language that works well for:

- Network services
- Concurrent connections
- WebSocket handling
- RPC services
- Kafka consumers and producers
- Small container images
- Simple deployment as static binaries

The project also benefits from fast builds, strong typing, and a mature standard library.

## Decision
We will use Go as the primary language for all Pylon backend services.

## Consequences

### Positive
- Goroutines make concurrent service and WebSocket workloads straightforward.
- Static binaries are easy to package in Docker images.
- Strong standard library support for HTTP and context cancellation.
- Good ecosystem for PostgreSQL, Redis, Kafka, Prometheus, and OpenTelemetry.
- Fast compilation improves developer feedback loops.

### Negative
- Error handling is explicit and verbose.
- Some abstractions require more boilerplate than dynamic languages.
- Runtime-level dependency injection is not built in.

### Risks
- Poor context handling can cause leaks in long-running services.
- Inconsistent error wrapping can make debugging harder.
- Developers must be disciplined with tests and package boundaries.

## Alternatives Considered

### Node.js
- Strong real-time ecosystem.
- Good WebSocket support.
- Single-threaded runtime can become harder to manage for CPU-heavy or highly concurrent workloads.

### Rust
- Excellent performance and safety.
- Higher learning curve.
- Slower development for this project scope.

### Java/Kotlin
- Mature microservice ecosystem.
- More runtime and deployment overhead than Go for this project.