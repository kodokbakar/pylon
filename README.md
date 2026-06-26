<div align="center">

# Pylon

Distributed real-time chat system built with Go microservices, Connect-Go, WebSocket, Kafka, PostgreSQL, Redis, Kubernetes, Prometheus, Grafana, and OpenTelemetry.

[![CI](https://github.com/kodokbakar/pylon/actions/workflows/ci.yml/badge.svg)](https://github.com/kodokbakar/pylon/actions)
[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?logo=go\&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## Features

| Feature                   | Description                                                                            |
| ------------------------- | -------------------------------------------------------------------------------------- |
| Real-time messaging       | Authenticated WebSocket gateway for room join, leave, message, typing, and ping events |
| Microservice architecture | API Gateway, Chat Service, Presence Service, Room Service, and Notification Service    |
| Typed RPC contracts       | Protobuf contracts generated with Buf and served through Connect-Go                    |
| Event streaming           | Kafka message events from Chat Service to Notification Service                         |
| Durable storage           | PostgreSQL for users, rooms, messages, refresh tokens, and notifications               |
| Fast ephemeral state      | Redis for presence, typing, rate limiting, and short-lived state                       |
| JWT authentication        | Register, login, refresh token, and protected HTTP/WebSocket routes                    |
| Rate limiting             | Redis-backed API Gateway rate limiter                                                  |
| Observability             | Prometheus metrics, Grafana dashboards, OpenTelemetry tracing, and Jaeger              |
| Kubernetes deployment     | Kustomize-style base manifests with probes, resources, HPA, and security contexts      |
| Load testing              | HTTP load-test suite using clank-cli                                                   |

---

## Architecture

```text
                                     ┌──────────────────────┐
                                     │      Clients         │
                                     │  HTTP + WebSocket    │
                                     └──────────┬───────────┘
                                                │
                                                ▼
                                  ┌─────────────────────────┐
                                  │      API Gateway        │
                                  │  REST, WebSocket, JWT   │
                                  │        :8080            │
                                  └───────┬─────────┬───────┘
                                          │         │
                         Connect RPC      │         │ WebSocket fan-out
                                          │         │
               ┌──────────────────────────┼─────────┼──────────────────────────┐
               │                          │         │                          │
               ▼                          ▼         ▼                          ▼
   ┌─────────────────────┐    ┌─────────────────────┐          ┌─────────────────────┐
   │    Chat Service     │    │    Room Service     │          │  Presence Service   │
   │       :9001         │    │       :9003         │          │       :9002         │
   │ PostgreSQL + Kafka  │    │    PostgreSQL       │          │        Redis        │
   └──────────┬──────────┘    └──────────┬──────────┘          └──────────┬──────────┘
              │                          │                                │
              │ message-events           │                                │
              ▼                          ▼                                ▼
        ┌──────────┐              ┌──────────────┐                 ┌──────────┐
        │  Kafka   │              │  PostgreSQL  │                 │  Redis   │
        └────┬─────┘              └──────────────┘                 └──────────┘
             │
             ▼
   ┌─────────────────────┐
   │ Notification Service│
   │       :9004         │
   │ PostgreSQL + Kafka  │
   └─────────────────────┘

Observability:
- `/metrics` endpoints -> Prometheus -> Grafana
- OpenTelemetry SDK -> OTel Collector -> Jaeger
```

---

## Services

| Service              |   Port | Purpose                                                           |
| -------------------- | -----: | ----------------------------------------------------------------- |
| API Gateway          | `8080` | Public HTTP API, WebSocket gateway, JWT auth, rate limiting       |
| Chat Service         | `9001` | Message persistence, message history, Kafka message events        |
| Presence Service     | `9002` | Online, offline, typing, and room presence state                  |
| Room Service         | `9003` | Room creation, listing, membership, join, and leave               |
| Notification Service | `9004` | Notification persistence and Kafka-driven notification processing |

---

## Tech Stack

| Area           | Technology                 | Purpose                                                    |
| -------------- | -------------------------- | ---------------------------------------------------------- |
| Language       | Go `1.26.4`                | Backend services                                           |
| RPC            | Connect-Go                 | Typed service-to-service communication                     |
| Protobuf       | Buf                        | Proto linting and Go code generation                       |
| WebSocket      | coder/websocket            | Real-time client connections                               |
| Message broker | Kafka + segmentio/kafka-go | Event streaming                                            |
| Database       | PostgreSQL 17 + pgx        | Durable relational data                                    |
| Cache/state    | Redis 7 + go-redis         | Presence, typing, and rate limiting                        |
| Auth           | JWT                        | API and WebSocket authentication                           |
| Metrics        | Prometheus client          | HTTP, RPC, Kafka, and business metrics                     |
| Dashboards     | Grafana                    | Service, infrastructure, Kafka, and business dashboards    |
| Tracing        | OpenTelemetry + Jaeger     | Distributed tracing                                        |
| Local infra    | Docker Compose             | PostgreSQL, Redis, Kafka, Kafka UI, Jaeger, OTel Collector |
| Deployment     | Kubernetes                 | Production-style service deployment                        |
| Load testing   | clank-cli                  | HTTP load testing for API Gateway endpoints                |

---

## Getting Started

### Prerequisites

Required:

* Go `1.26.4+`
* Docker
* Docker Compose
* Buf CLI
* golang-migrate
* golangci-lint

Optional:

* kubectl
* clank-cli
* jq
* curl

### Clone

```bash
git clone https://github.com/kodokbakar/pylon.git
cd pylon
```

### Configure Environment

Create a local `.env` file:

```bash
cp .env.example .env
```

When using the Docker Compose infrastructure from this repository, use the exposed host ports:

```env
DATABASE_URL=postgres://pylon:pylon_dev@localhost:5433/pylon?sslmode=disable
REDIS_URL=redis://localhost:6380
KAFKA_BROKERS=localhost:9092

CHAT_SERVICE_URL=http://localhost:9001
PRESENCE_SERVICE_URL=http://localhost:9002
ROOM_SERVICE_URL=http://localhost:9003
NOTIFICATION_SERVICE_URL=http://localhost:9004

JWT_SECRET=change-this-local-secret
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
```

### Start Local Infrastructure

```bash
make dev
```

This starts:

| Component             | Local Address            |
| --------------------- | ------------------------ |
| PostgreSQL            | `localhost:5433`         |
| Redis                 | `localhost:6380`         |
| Kafka                 | `localhost:9092`         |
| Kafka UI              | `http://localhost:8085`  |
| Jaeger                | `http://localhost:16686` |
| OTel Collector gRPC   | `localhost:4317`         |
| OTel Collector HTTP   | `localhost:4318`         |
| OTel Collector health | `http://localhost:13133` |

### Generate Protobuf Code

```bash
make proto
```

### Run Database Migrations & Seed Demo Data

```bash
export DATABASE_URL="postgres://pylon:pylon_dev@localhost:5433/pylon?sslmode=disable"
make migrate-up
make seed
```

### Run Services Locally

Run each service in a separate terminal:

```bash
go run ./cmd/chat-service
```

```bash
go run ./cmd/room-service
```

```bash
go run ./cmd/presence-service
```

```bash
go run ./cmd/notification-service
```

```bash
go run ./cmd/api-gateway
```

Then check the API Gateway health endpoint:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "data": {
    "service": "api-gateway",
    "status": "ok"
  }
}
```

---

## API Reference

All public HTTP routes are exposed through the API Gateway.

Base URL:

```text
http://localhost:8080
```

### Auth

| Method | Endpoint                | Auth | Description                               |
| ------ | ----------------------- | ---- | ----------------------------------------- |
| `POST` | `/api/v1/auth/register` | No   | Register a user                           |
| `POST` | `/api/v1/auth/login`    | No   | Login and receive access + refresh tokens |
| `POST` | `/api/v1/auth/refresh`  | No   | Refresh access token                      |

Register request:

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "password123"
}
```

Login request:

```json
{
  "email": "alice@example.com",
  "password": "password123"
}
```

Refresh request:

```json
{
  "refresh_token": "REFRESH_TOKEN"
}
```

Auth response shape:

```json
{
  "data": {
    "user": {
      "id": "user-id",
      "username": "alice",
      "email": "alice@example.com"
    },
    "token": "ACCESS_TOKEN",
    "refresh_token": "REFRESH_TOKEN",
    "expires_at": "2026-06-23T00:00:00Z",
    "refresh_expires_at": "2026-06-30T00:00:00Z"
  }
}
```

### Rooms

Protected routes require:

```http
Authorization: Bearer ACCESS_TOKEN
```

| Method | Endpoint                   | Auth | Description                           |
| ------ | -------------------------- | ---- | ------------------------------------- |
| `GET`  | `/api/v1/rooms`            | Yes  | List rooms for the authenticated user |
| `POST` | `/api/v1/rooms`            | Yes  | Create a room                         |
| `POST` | `/api/v1/rooms/{id}/join`  | Yes  | Join a room                           |
| `POST` | `/api/v1/rooms/{id}/leave` | Yes  | Leave a room                          |

Create room request:

```json
{
  "name": "general",
  "type": "channel",
  "member_ids": []
}
```

Supported room types:

| Type      | Description               |
| --------- | ------------------------- |
| `direct`  | 1-on-1 room               |
| `group`   | Group chat                |
| `channel` | Public channel-style room |

### Messages

| Method | Endpoint                      | Auth | Description                     |
| ------ | ----------------------------- | ---- | ------------------------------- |
| `GET`  | `/api/v1/rooms/{id}/messages` | Yes  | List message history for a room |

Query parameters:

| Parameter   | Default |   Max | Description                  |
| ----------- | ------: | ----: | ---------------------------- |
| `limit`     |    `50` | `100` | Number of messages to return |
| `before_id` |   empty |     - | Cursor for older messages    |

Example:

```bash
curl "http://localhost:8080/api/v1/rooms/ROOM_ID/messages?limit=50" \
  -H "Authorization: Bearer ACCESS_TOKEN"
```

Message sending is handled through WebSocket, not through an HTTP `POST /api/v1/rooms/{id}/messages` route.

---

## WebSocket Protocol

Endpoint:

```text
GET /ws
```

Authentication can be sent with either:

```http
Authorization: Bearer ACCESS_TOKEN
```

or query string:

```text
/ws?token=ACCESS_TOKEN
/ws?access_token=ACCESS_TOKEN
```

### Client to Server

All client messages are JSON text frames.

#### Join Room

```json
{
  "type": "join",
  "room_id": "ROOM_ID"
}
```

#### Leave Room

```json
{
  "type": "leave",
  "room_id": "ROOM_ID"
}
```

#### Send Message

```json
{
  "type": "message",
  "room_id": "ROOM_ID",
  "content": "hello world",
  "msg_type": "text"
}
```

Supported `msg_type` values:

| Type     | Description    |
| -------- | -------------- |
| `text`   | Text message   |
| `image`  | Image message  |
| `file`   | File message   |
| `system` | System message |

#### Typing

```json
{
  "type": "typing",
  "room_id": "ROOM_ID"
}
```

#### Ping

```json
{
  "type": "ping"
}
```

### Server to Client

#### Message

```json
{
  "type": "message",
  "message_id": "MESSAGE_ID",
  "room_id": "ROOM_ID",
  "sender": {
    "id": "USER_ID",
    "username": "alice",
    "display_name": "Alice",
    "avatar_url": ""
  },
  "content": "hello world",
  "msg_type": "text",
  "created_at": "2026-06-23T00:00:00Z"
}
```

#### User Joined

```json
{
  "type": "user_joined",
  "room_id": "ROOM_ID",
  "user": {
    "id": "USER_ID"
  }
}
```

#### User Left

```json
{
  "type": "user_left",
  "room_id": "ROOM_ID",
  "user_id": "USER_ID"
}
```

#### Typing

```json
{
  "type": "typing",
  "room_id": "ROOM_ID",
  "user_id": "USER_ID",
  "username": "alice"
}
```

#### Pong

```json
{
  "type": "pong"
}
```

#### Error

```json
{
  "type": "error",
  "code": "BAD_REQUEST",
  "message": "room_id is required"
}
```

---

## Internal RPC Services

Pylon uses Connect-Go generated from Protobuf definitions in `proto/`.

| Service              | Proto                                                    | Main RPCs                                                                                  |
| -------------------- | -------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| Chat Service         | `proto/pylon/chat/v1/chat_service.proto`                 | `SendMessage`, `StreamMessages`, `GetMessages`                                             |
| Presence Service     | `proto/pylon/presence/v1/presence_service.proto`         | `SetOnline`, `SetOffline`, `SetTyping`, `GetPresence`, `GetRoomPresence`, `StreamPresence` |
| Room Service         | `proto/pylon/room/v1/room_service.proto`                 | `CreateRoom`, `GetRoom`, `ListRooms`, `JoinRoom`, `LeaveRoom`, `GetRoomMembers`            |
| Notification Service | `proto/pylon/notification/v1/notification_service.proto` | `SendNotification`, `GetNotifications`, `MarkAsRead`                                       |
| Gateway Service      | `proto/pylon/gateway/v1/gateway_service.proto`           | `ValidateToken`, `RefreshToken`                                                            |

Generate code:

```bash
make proto
```

Lint proto files:

```bash
make proto-lint
```

Check breaking changes:

```bash
make proto-breaking
```

---

## Project Structure

```text
pylon/
├── cmd/                         # Service entry points and Dockerfiles
│   ├── api-gateway/
│   ├── chat-service/
│   ├── presence-service/
│   ├── room-service/
│   └── notification-service/
├── deploy/                      # Kubernetes, Grafana, Jaeger, and OTel configs
│   ├── base/
│   ├── grafana/
│   └── jaeger/
├── docs/
│   └── ADR/                     # Architecture Decision Records
├── gen/                         # Generated Go protobuf/connect code
├── internal/                    # Shared packages
│   ├── config/
│   ├── database/
│   ├── metrics/
│   ├── middleware/
│   ├── observability/
│   ├── response/
│   └── tracing/
├── migrations/                  # PostgreSQL migrations
├── proto/                       # Protobuf definitions and Buf config
├── services/                    # Service implementations
│   ├── api-gateway/
│   ├── chat-service/
│   ├── presence-service/
│   ├── room-service/
│   └── notification-service/
├── tests/
│   └── load/                    # clank-cli HTTP load tests
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## Development

### Commands

```bash
make help              # Show all available commands

make dev               # Start local infrastructure
make dev-down          # Stop local infrastructure
make dev-logs          # Follow local infrastructure logs

make proto             # Generate protobuf code
make proto-lint        # Lint protobuf definitions
make proto-breaking    # Check proto breaking changes against main

make fmt               # Format Go code
make vet               # Run go vet
make lint              # Run golangci-lint

make test              # Run all tests
make test-cover        # Generate coverage.out and coverage.html
make test-integration  # Run integration tests with integration build tag
make test-load         # Run HTTP load tests with clank-cli

make build             # Build all service binaries
make build-gateway     # Build API Gateway
make build-chat        # Build Chat Service
make build-presence    # Build Presence Service
make build-room        # Build Room Service
make build-notification # Build Notification Service

make docker-build      # Build all service Docker images

make migrate-up        # Run database migrations
make seed              # Seed demo users, rooms, messages, and notifications

make migrate-down      # Roll back one migration
make migrate-create name=add_users # Create a new migration

make deps              # Download and tidy Go dependencies
make clean             # Remove build and coverage artifacts
```

### Quality Check

Run this before opening a PR:

```bash
go fmt ./...
go test ./...
make build
golangci-lint run ./...
```

### Integration Tests

```bash
go test -tags=integration -v ./tests/...
```

---

## Load Testing

HTTP load tests live in `tests/load`.

Run:

```bash
make test-load
```

The script uses `clank-cli` against API Gateway HTTP endpoints and writes JSON results to:

```text
tests/load/results/
```

Covered HTTP scenarios:

| Scenario     | Endpoint                           | Method |
| ------------ | ---------------------------------- | ------ |
| List Rooms   | `/api/v1/rooms`                    | `GET`  |
| Get Messages | `/api/v1/rooms/{room_id}/messages` | `GET`  |
| Create Room  | `/api/v1/rooms`                    | `POST` |
| Join Room    | `/api/v1/rooms/{room_id}/join`     | `POST` |
| Leave Room   | `/api/v1/rooms/{room_id}/leave`    | `POST` |

WebSocket message sending is not tested with clank-cli because clank-cli does not support WebSocket load testing.

See:

* `tests/load/README.md`
* `tests/load/ws/README.md`
* `tests/load/grpc/README.md`

---

## Monitoring and Tracing

### Local URLs

| Tool                  | URL                      |
| --------------------- | ------------------------ |
| Kafka UI              | `http://localhost:8085`  |
| Jaeger                | `http://localhost:16686` |
| OTel Collector health | `http://localhost:13133` |

Prometheus and Grafana dashboards are provisioned under `deploy/grafana`, but a full local Prometheus/Grafana Compose stack is not currently included in `docker-compose.yml`.

### Metrics

Pylon services expose:

```text
GET /metrics
```

Application metrics include:

| Metric                                 | Source                               |
| -------------------------------------- | ------------------------------------ |
| `pylon_http_requests_total`            | API Gateway HTTP middleware          |
| `pylon_http_request_duration_seconds`  | API Gateway HTTP middleware          |
| `pylon_http_requests_in_flight`        | API Gateway HTTP middleware          |
| `pylon_grpc_requests_total`            | Service Connect/RPC middleware       |
| `pylon_grpc_request_duration_seconds`  | Service Connect/RPC middleware       |
| `pylon_kafka_messages_published_total` | Chat Service Kafka producer          |
| `pylon_kafka_messages_consumed_total`  | Notification Service Kafka consumer  |
| `pylon_kafka_publish_duration_seconds` | Chat Service Kafka producer          |
| `pylon_websocket_connections_active`   | API Gateway WebSocket handler        |
| `pylon_messages_sent_total`            | Chat Service message flow            |
| `pylon_rooms_created_total`            | Room Service room creation flow      |
| `pylon_users_online`                   | Presence Service online/offline flow |

### Grafana

Grafana provisioning files are available in:

```text
deploy/grafana/
```

Dashboards:

| Dashboard              | Purpose                                                                          |
| ---------------------- | -------------------------------------------------------------------------------- |
| Pylon Overview         | High-level HTTP, RPC, WebSocket, users, and messages overview                    |
| Pylon API Gateway      | Request rate, latency, in-flight requests, WebSocket activity, and rate limiting |
| Pylon Microservices    | RPC request rate, latency, errors, and pod resources                             |
| Pylon Kafka            | Kafka publish/consume metrics, latency, consumer lag, and broker metrics         |
| Pylon Infrastructure   | PostgreSQL, Redis, and Kubernetes infrastructure metrics                         |
| Pylon Business Metrics | Users online, rooms created, messages sent, and active WebSocket connections     |

See `deploy/grafana/README.md`.

### Tracing

OpenTelemetry tracing is initialized in every service and exported to the OTel Collector.

Local tracing flow:

```text
Pylon services -> OTel Collector :4317 -> Jaeger :16686
```

Main environment variables:

```env
OTEL_TRACING_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
OTEL_SERVICE_VERSION=1.0.0
OTEL_TRACES_SAMPLE_RATIO=1
```

---

## Deployment

### Docker Images

Build all service images:

```bash
make docker-build
```

Build one image:

```bash
make docker-build-gateway
make docker-build-chat
make docker-build-presence
make docker-build-room
make docker-build-notification
```

### Kubernetes

Render manifests:

```bash
kubectl kustomize deploy/base
```

Apply manifests:

```bash
kubectl apply -k deploy/base
```

Check resources:

```bash
kubectl get pods -n pylon
kubectl get svc -n pylon
kubectl get hpa -n pylon
```

The Kubernetes base includes:

* Namespace
* ConfigMap
* Example Secret placeholders
* Ingress
* Deployments and Services for all Pylon services
* HPA manifests
* Prometheus scrape config
* Jaeger
* OpenTelemetry Collector

Important production note:

`deploy/base/secrets.yaml` contains placeholder values only. Replace it with External Secrets, Sealed Secrets, or generated Kubernetes Secrets before production deployment.

### Minikube

Local Kubernetes deployment is available through:

```bash
make minikube-deploy
```

Useful commands:

```bash
make minikube-status
make minikube-clean
```

See `deploy/minikube/README.md` for the full Minikube workflow.

---

## Architecture Decision Records

Architecture decisions are documented in:

```text
docs/ADR/
```

Start here:

```text
docs/ADR/README.md
```

Current ADRs cover:

* Microservice architecture
* Go as primary language
* Connect-Go for RPC
* coder/websocket
* Kafka with segmentio/kafka-go
* PostgreSQL with pgx
* Redis with go-redis
* Buf for Protobuf
* Kubernetes deployment
* Monorepo structure
* Prometheus, Grafana, and OpenTelemetry

---

## Performance

Load-test targets are documented in `tests/load/README.md`.

Current targets:

| Metric               |         Target |         Alert |
| -------------------- | -------------: | ------------: |
| Request Duration P50 |      `< 100ms` |     `> 200ms` |
| Request Duration P95 |      `< 200ms` |     `> 500ms` |
| Request Duration P99 |      `< 500ms` |        `> 1s` |
| Error Rate           |       `< 0.1%` |        `> 1%` |
| Throughput           | `>= 500 req/s` | `< 200 req/s` |

Baseline result JSON files should only be committed after running tests against a known environment. Do not commit fake load-test results.

---

## Contributing

1. Create a feature branch from `dev`.
2. Keep changes small and focused.
3. Run verification before committing:

```bash
go fmt ./...
go test ./...
make build
golangci-lint run ./...
```

4. Update docs when behavior changes.
5. Add or update tests for service logic.
6. Open a PR with a clear summary and verification output.

Recommended commit style:

```text
feat: add room listing endpoint
fix: preserve metrics middleware in service handlers
docs: update load testing guide
test: add kafka propagation coverage
```

---

## License

MIT License. See [LICENSE](LICENSE) for details.
