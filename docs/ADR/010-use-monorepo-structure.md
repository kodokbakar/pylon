# ADR-010: Use Monorepo Structure

## Status
Accepted

## Context
Pylon contains multiple services that share:

- Proto definitions
- Generated code
- Internal packages
- Deployment manifests
- CI workflows
- Observability setup
- Development conventions

Splitting every service into a separate repository would increase coordination overhead early in the project.

## Decision
We will use a monorepo structure.

The repository contains service entry points under `cmd/`, service implementations under `services/`, shared packages under `internal/`, proto contracts under `proto/`, generated code under `gen/`, and deployment files under `deploy/`.

## Consequences

### Positive
- Easier cross-service refactoring.
- Single CI pipeline for the full system.
- Shared proto and generated code stay in one place.
- Local development is simpler.
- Architecture is easier to understand from one repository.

### Negative
- Repository can grow large over time.
- CI can become slower if not optimized.
- Service ownership boundaries require discipline.

### Risks
- Shared internal packages can become a dumping ground.
- Unrelated changes can affect multiple services.
- Build and test workflows need to stay efficient.

## Alternatives Considered

### Multi-Repo
- Clear ownership boundaries.
- Independent versioning per service.
- Higher setup and coordination cost.
- More difficult cross-service changes.

### Hybrid Repo
- Group related services together and split mature ones later.
- More flexible long term.
- Adds ambiguity about where code should live.