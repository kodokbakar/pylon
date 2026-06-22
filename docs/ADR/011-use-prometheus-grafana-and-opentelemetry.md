# ADR-011: Use Prometheus, Grafana, and OpenTelemetry for Observability

## Status
Accepted

## Context
Pylon is a distributed system with HTTP, Connect-Go RPC, WebSocket, Kafka, PostgreSQL, Redis, and Kubernetes components.

The project needs observability for:

- HTTP request rates and latency
- RPC request rates and latency
- Kafka publish and consume behavior
- Business metrics such as active WebSocket connections and messages sent
- Dashboards for service and infrastructure health
- Distributed traces across API Gateway, internal services, and Kafka flows

## Decision
We will use:

- Prometheus for metrics collection
- Grafana for dashboards
- OpenTelemetry for distributed tracing
- Jaeger for trace visualization

Metrics are exposed through service `/metrics` endpoints. Tracing is initialized in each service and exported through an OpenTelemetry Collector to Jaeger.

## Consequences

### Positive
- Prometheus provides a standard metrics model.
- Grafana dashboards make service health easier to inspect.
- OpenTelemetry gives cross-service request visibility.
- Jaeger helps debug distributed request paths.
- Metrics and traces are implemented with common cloud-native tooling.

### Negative
- More infrastructure to deploy and maintain.
- Dashboard queries need to match actual metric labels.
- Tracing adds configuration and runtime overhead.

### Risks
- High-cardinality labels can damage Prometheus performance.
- Missing exporters can make infrastructure dashboards incomplete.
- Incorrect middleware ordering can silently drop metrics.
- Trace sampling must be tuned for production traffic.

## Alternatives Considered

### Logs Only
- Simple to start.
- Not enough for latency, rate, and distributed request analysis.
- Harder to debug multi-service flows.

### Vendor APM
- Faster setup and polished UI.
- Vendor lock-in.
- Cost can grow with traffic volume.

### Metrics Without Tracing
- Lower overhead.
- Good for aggregate health.
- Insufficient for debugging cross-service request paths.