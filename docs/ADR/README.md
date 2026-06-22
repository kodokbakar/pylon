# Architecture Decision Records

This directory contains Architecture Decision Records (ADRs) for Pylon.

An ADR explains one important technical decision, why it was chosen, what alternatives were considered, and what consequences we accept.

## ADR Index

| ADR | Title | Status |
|---|---|---|
| [ADR-001](001-use-microservice-architecture.md) | Use Microservice Architecture | Accepted |
| [ADR-002](002-use-go-as-primary-language.md) | Use Go as Primary Language | Accepted |
| [ADR-003](003-use-connect-go-for-rpc.md) | Use Connect-Go for RPC | Accepted |
| [ADR-004](004-use-coder-websocket.md) | Use coder/websocket for Real-Time Connections | Accepted |
| [ADR-005](005-use-segmentio-kafka-go.md) | Use segmentio/kafka-go for Event Streaming | Accepted |
| [ADR-006](006-use-pgx-for-postgresql.md) | Use pgx for PostgreSQL Access | Accepted |
| [ADR-007](007-use-go-redis-for-presence-cache.md) | Use go-redis for Redis Access | Accepted |
| [ADR-008](008-use-buf-for-protobuf.md) | Use Buf for Protobuf Tooling | Accepted |
| [ADR-009](009-use-kubernetes-for-deployment.md) | Use Kubernetes for Deployment | Accepted |
| [ADR-010](010-use-monorepo-structure.md) | Use Monorepo Structure | Accepted |
| [ADR-011](011-use-prometheus-grafana-and-opentelemetry.md) | Use Prometheus, Grafana, and OpenTelemetry for Observability | Accepted |

## Status Values

| Status | Meaning |
|---|---|
| Proposed | Under discussion |
| Accepted | Approved and implemented |
| Deprecated | No longer recommended |
| Superseded | Replaced by another ADR |

## Template

```md
# ADR-000: Title

## Status
Accepted

## Context
What problem are we solving?

## Decision
What decision did we make?

## Consequences

### Positive
- Good outcome

### Negative
- Trade-off

### Risks
- Risk to watch

## Alternatives Considered

### Alternative A
- Pros
- Cons