# ADR-009: Use Kubernetes for Deployment

## Status
Accepted

## Context
Pylon has multiple independently deployable services and infrastructure components.

The deployment platform needs to support:

- Independent service deployments
- Horizontal scaling
- Service discovery
- ConfigMaps and Secrets
- Health and readiness checks
- Prometheus scraping
- Environment-specific overlays

## Decision
We will use Kubernetes manifests with Kustomize-style base and overlays.

Kubernetes resources live under `deploy/base` and `deploy/overlays`.

## Consequences

### Positive
- Services can scale independently.
- Service discovery is handled by Kubernetes Services.
- Config and secrets can be managed separately from images.
- Health and readiness probes improve operational safety.
- Kustomize overlays support dev and prod differences.

### Negative
- More complex than Docker Compose.
- Requires Kubernetes knowledge.
- Local debugging can be slower than direct process execution.

### Risks
- Incorrect resource limits can cause throttling or OOM kills.
- Misconfigured probes can restart healthy pods.
- Secret management must be handled carefully for production.

## Alternatives Considered

### Docker Compose
- Excellent for local development.
- Simple to understand.
- Not enough for production orchestration and scaling.

### Nomad
- Simpler operational model in some environments.
- Smaller ecosystem for Kubernetes-native tooling.
- Less aligned with common cloud-native deployment workflows.

### Plain VMs
- Simple at first.
- Manual deployment and service discovery burden.
- Harder to scale and recover automatically.