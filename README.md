<div align="center">

# Pylon

Distributed real-time chat system built with Go microservices, gRPC, Kafka, and Kubernetes.

[![CI](https://github.com/kodokbakar/pylon/actions/workflows/ci.yml/badge.svg)](https://github.com/kodokbakar/pylon/actions)
[![Go Version](https://img.shields.io/badge/Go-1.26.4-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

</div>

---

## Features

| Feature | Description |
|---------|-------------|
| Real-time Messaging | WebSocket-based instant message delivery |
| Microservice Architecture | Independent services communicating via gRPC |
| Event Streaming | Kafka for async event processing |
| User Presence | Online/offline/typing indicators |
| Chat Rooms | Direct, group, and channel support |
| Push Notifications | Kafka-driven notification pipeline |
| Observability | Prometheus metrics + Grafana dashboards |
| Container Orchestration | Kubernetes deployment with auto-scaling |

## Architecture

```
                    ┌─────────────────┐
                    │   API Gateway   │ ← WebSocket, REST
                    │   (connect-go)  │
                    └────────┬────────┘
                             │ gRPC
              ┌──────────────┼──────────────┐
              │              │              │
    ┌─────────▼──────┐ ┌────▼─────┐ ┌──────▼────────┐
    │  Chat Service  │ │  Room    │ │  Presence     │
    │  (WebSocket)   │ │  Service │ │  Service      │
    └────────┬───────┘ └──────────┘ └───────┬───────┘
             │                              │
             │ Kafka                        │ Redis
             │                              │
    ┌────────▼───────────────────┐  ┌───────▼───────┐
    │   Notification Service     │  │    Redis      │
    └────────────────────────────┘  └───────────────┘
             │
    ┌────────▼────────┐
    │   PostgreSQL    │
    └─────────────────┘
```

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | [Go 1.26.4](https://go.dev/) |
| gRPC | [connect-go](https://connectrpc.com/) |
| WebSocket | [coder/websocket](https://github.com/coder/websocket) |
| Message Broker | [segmentio/kafka-go](https://github.com/segmentio/kafka-go) |
| Database | [PostgreSQL 17](https://www.postgresql.org/) via [pgx](https://github.com/jackc/pgx) |
| Cache | [Redis 7](https://redis.io/) via [go-redis](https://github.com/redis/go-redis) |
| Protobuf | [buf](https://buf.build/) |
| Container | [Docker](https://www.docker.com/) |
| Orchestration | [Kubernetes](https://kubernetes.io/) |
| Monitoring | [Prometheus](https://prometheus.io/) + [Grafana](https://grafana.com/) |

## Getting Started

### Prerequisites

- Go 1.26.4+
- Docker & Docker Compose
- buf CLI (for protobuf generation)
- kubectl (for Kubernetes)

### Local Development

```bash
# Clone the repository
git clone https://github.com/kodokbakar/pylon.git
cd pylon

# Start infrastructure services
make dev

# Generate protobuf code
make proto

# Run services (in separate terminals)
go run ./cmd/api-gateway
go run ./cmd/chat-service
go run ./cmd/presence-service
go run ./cmd/room-service
go run ./cmd/notification-service
```

### Docker

```bash
# Build all service images
make docker-build

# Run with docker-compose
docker compose up -d
```

### Kubernetes

```bash
# Apply Kubernetes manifests
kubectl apply -f deploy/base/

# Check pod status
kubectl get pods -n pylon
```

## Project Structure

```
pylon/
├── proto/                          # Protobuf definitions
│   └── pylon/
│       ├── chat/v1/               # Chat service proto
│       ├── presence/v1/           # Presence service proto
│       ├── room/v1/               # Room service proto
│       ├── notification/v1/       # Notification service proto
│       └── gateway/v1/            # Gateway service proto
├── cmd/                            # Service entry points
│   ├── api-gateway/
│   ├── chat-service/
│   ├── presence-service/
│   ├── room-service/
│   └── notification-service/
├── internal/                       # Shared packages
│   ├── config/
│   ├── database/
│   ├── middleware/
│   └── response/
├── services/                       # Service implementations
├── deploy/                         # Kubernetes manifests
├── migrations/                     # Database migrations
├── docs/ADR/                       # Architecture Decision Records
├── docker-compose.yml
├── Makefile
└── README.md
```

## API

### gRPC Services

| Service | Port | Description |
|---------|------|-------------|
| API Gateway | 8080 | WebSocket + REST entry point |
| Chat Service | 9091 | Message handling |
| Presence Service | 9092 | User presence tracking |
| Room Service | 9093 | Room management |
| Notification Service | 9094 | Push notifications |

### Endpoints

#### Chat Service
- `SendMessage` - Send a message to a room
- `StreamMessages` - Stream real-time messages
- `GetMessages` - Get message history

#### Presence Service
- `SetOnline` - Mark user as online
- `SetOffline` - Mark user as offline
- `SetTyping` - Mark user as typing
- `GetPresence` - Get user presence status
- `StreamPresence` - Stream presence changes

#### Room Service
- `CreateRoom` - Create a chat room
- `GetRoom` - Get room details
- `ListRooms` - List user's rooms
- `JoinRoom` - Join a room
- `LeaveRoom` - Leave a room
- `GetRoomMembers` - Get room members

#### Notification Service
- `SendNotification` - Send a notification
- `GetNotifications` - Get user notifications
- `MarkAsRead` - Mark notification as read

## Monitoring

Access monitoring tools:

- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090
- **Kafka UI**: http://localhost:8085

Grafana dashboard provisioning files are available in `deploy/grafana`.

Infrastructure panels require the matching exporters to be installed and scraped by Prometheus:

- PostgreSQL panels require `postgres_exporter`.
- Redis panels require `redis_exporter`.
- Kubernetes pod and HPA panels require `kube-state-metrics`.
- Container CPU, memory, and network panels require cAdvisor or kubelet metrics.
- Kafka lag and broker panels require `kafka-exporter` and/or Kafka JMX exporter.

See `deploy/grafana/README.md` for dashboard details and exporter requirements.

## Development

### Available Commands

```bash
make help           # Show all available commands
make dev            # Start local infrastructure
make dev-down       # Stop local infrastructure
make proto          # Generate protobuf code
make lint           # Run linter
make test           # Run tests
make test-cover     # Run tests with coverage
make build          # Build all services
make migrate-up     # Run database migrations
make migrate-down   # Rollback migrations
```

### Load Testing

HTTP load tests are available under `tests/load`.

```bash
make test-load
```

The load test suite uses `clank-cli` for API Gateway HTTP endpoints and writes JSON results to `tests/load/results`.

See `tests/load/README.md` for setup, environment variables, target metrics, and result analysis.

### Creating Migrations

```bash
make migrate-create name=add_users
```

## License

MIT License - see [LICENSE](LICENSE) for details.
