# ADR-006: Use pgx for PostgreSQL Access

## Status
Accepted

## Context
Pylon stores durable data in PostgreSQL, including:

- Users
- Refresh tokens
- Rooms
- Room members
- Messages
- Notifications

The database layer needs connection pooling, context support, explicit SQL, and reliable error handling.

## Decision
We will use PostgreSQL as the primary durable database and `github.com/jackc/pgx/v5` for Go access.

The shared database package creates `pgxpool.Pool` instances, and repositories use explicit SQL queries.

## Consequences

### Positive
- `pgx` is a strong PostgreSQL-native driver.
- `pgxpool` provides built-in pooling.
- Explicit SQL keeps query behavior clear.
- Context-aware calls fit service shutdown and timeout patterns.
- PostgreSQL supports relational integrity for users, rooms, messages, and notifications.

### Negative
- Manual SQL requires more care than an ORM.
- Schema changes need migrations.
- Repository tests must cover important query behavior.

### Risks
- Poor indexing can hurt message and room queries.
- Long transactions can block high-traffic paths.
- Connection pool settings must be tuned for production.

## Alternatives Considered

### database/sql with lib/pq
- Standard library interface.
- `lib/pq` is older and less feature-rich than pgx.
- Less PostgreSQL-native behavior.

### GORM
- Faster CRUD scaffolding.
- More implicit behavior.
- Harder to optimize and reason about for performance-critical paths.

### SQLC
- Strong generated query typing.
- Adds code generation workflow.
- Could be considered later if repository SQL grows significantly.