# ADR-008: Use Buf for Protobuf Tooling

## Status
Accepted

## Context
Pylon uses Protobuf contracts for internal services.

The project needs:

- Consistent proto layout
- Linting for proto files
- Generated Go and Connect-Go code
- Repeatable generation commands
- Future breaking-change checks

## Decision
We will use Buf for Protobuf tooling.

Proto definitions live under `proto/`, and generated Go code is written to `gen/`.

## Consequences

### Positive
- Standardized proto linting.
- Repeatable code generation.
- Remote plugins reduce local plugin setup.
- Breaking-change checks can be added to CI workflows.
- Keeps service contracts explicit and reviewable.

### Negative
- Developers need Buf installed locally.
- Generated files must be kept in sync.
- Proto rules may require occasional lint exceptions.

### Risks
- Generated code drift can happen if developers forget to run generation.
- Breaking API changes must be managed carefully.
- Proto package structure must remain stable as services grow.

## Alternatives Considered

### protoc Directly
- Official compiler.
- More manual plugin setup.
- Harder to standardize across developers.

### Manual JSON Contracts
- Simple for HTTP-only APIs.
- Weak service-to-service contract enforcement.
- More runtime error risk.

### OpenAPI First
- Good for public REST APIs.
- Less suitable for internal streaming and RPC service definitions.