# ADR-003: Use Connect-Go for RPC

## Status
Accepted

## Context
Pylon services need typed internal communication between:

- API Gateway and Chat Service
- API Gateway and Room Service
- Notification Service and Room Service
- Other future service-to-service calls

The RPC layer should support Protobuf contracts, generated clients and servers, and HTTP-friendly deployment.

## Decision
We will use Connect-Go for RPC between services.

Service contracts are defined in `proto/`, generated into `gen/`, and served through Connect-Go HTTP handlers.

## Consequences

### Positive
- Strongly typed request and response contracts.
- Works well with Protobuf and Buf.
- Uses normal HTTP handlers, which makes middleware integration simple.
- Easier to expose through API Gateway patterns than raw gRPC-only servers.
- Compatible with Connect protocol and gRPC-style service definitions.

### Negative
- Smaller ecosystem than traditional grpc-go.
- Some external load-testing and debugging tools focus on classic gRPC.
- New contributors may need to learn Connect-specific conventions.

### Risks
- Incorrect HTTP middleware ordering can break metrics or tracing.
- Generated code must stay in sync with proto definitions.
- RPC boundaries must not replace clear domain boundaries.

## Alternatives Considered

### grpc-go
- Very mature and widely adopted.
- More examples and tooling.
- Requires more gRPC-specific server setup.
- Browser support usually needs additional proxying or gRPC-Web setup.

### REST Between Services
- Simple and familiar.
- Less strict contract enforcement.
- More manual client and response handling.

### GraphQL
- Flexible query layer.
- Not necessary for internal service-to-service calls.
- Adds complexity that does not fit the current backend architecture.