# Pylon Load Testing

Load testing setup for Pylon HTTP endpoints using `clank-cli`.

## Scope

This load test suite covers HTTP endpoints exposed by the API Gateway:

| Scenario | Endpoint | Method | Tool |
|---|---:|---:|---|
| List Rooms | `/api/v1/rooms` | GET | clank-cli |
| Get Messages | `/api/v1/rooms/{room_id}/messages` | GET | clank-cli |
| Create Room | `/api/v1/rooms` | POST | clank-cli |
| Join Room | `/api/v1/rooms/{room_id}/join` | POST | clank-cli |
| Leave Room | `/api/v1/rooms/{room_id}/leave` | POST | clank-cli |

The API Gateway currently sends chat messages through WebSocket (`GET /ws`). There is no HTTP `POST /api/v1/rooms/{room_id}/messages` route, so message sending is not included in the clank-cli HTTP suite.

## Requirements

- Running Pylon services
- `clank-cli` v0.2.0 or newer
- `curl`
- `jq`

Install `clank-cli` from source:

```bash
git clone https://github.com/kodokbakar/clank-cli
cd clank-cli
git checkout v0.2.0
cargo build --release
```

Then either copy the binary into your PATH or run the suite with `CLANK_BIN`:

```bash
CLANK_BIN=/path/to/clank-cli/target/release/clank-cli tests/load/http/run.sh
```

## Run

Start infrastructure:

```bash
make dev
```

Run Pylon services in separate terminals:

```bash
go run ./cmd/chat-service
go run ./cmd/room-service
go run ./cmd/presence-service
go run ./cmd/notification-service
go run ./cmd/api-gateway
```

Run HTTP load tests:

```bash
tests/load/http/run.sh
```

Optional environment variables:

| Variable             | Default                  | Description                                                          |
| -------------------- | ------------------------ | -------------------------------------------------------------------- |
| `BASE_URL`           | `http://localhost:8080`  | API Gateway URL                                                      |
| `CLANK_BIN`          | `clank-cli`              | clank-cli binary path                                                |
| `RESULTS_DIR`        | `tests/load/results`     | Output directory                                                     |
| `TOKEN`              | empty                    | Existing bearer token. If empty, script registers/logins a test user |
| `ROOM_ID`            | empty                    | Existing room id. If empty, script creates a room                    |
| `LOAD_TEST_USERNAME` | `loadtester`             | Test username                                                        |
| `LOAD_TEST_EMAIL`    | `loadtester@example.com` | Test email                                                           |
| `LOAD_TEST_PASSWORD` | `password123`            | Test password                                                        |
| `ROOM_NAME`          | `Load Test Room`         | Seed room name                                                       |
| `ROOM_TYPE`          | `channel`                | Seed room type                                                       |

Example:

```bash
BASE_URL=http://localhost:8080 \
CLANK_BIN=clank-cli \
tests/load/http/run.sh
```

## Results

Generated files:

```text
tests/load/results/list_rooms.json
tests/load/results/get_messages.json
tests/load/results/create_room.json
tests/load/results/join_room.json
tests/load/results/leave_room.json
```

Read selected metrics with `jq`:

```bash
jq '.latency.p50_ms' tests/load/results/list_rooms.json
jq '.latency.p95_ms' tests/load/results/list_rooms.json
jq '.latency.p99_ms' tests/load/results/list_rooms.json
jq '.error_rate' tests/load/results/list_rooms.json
```

## Targets

| Metric               |         Target |         Alert |
| -------------------- | -------------: | ------------: |
| Request Duration P50 |      `< 100ms` |     `> 200ms` |
| Request Duration P95 |      `< 200ms` |     `> 500ms` |
| Request Duration P99 |      `< 500ms` |        `> 1s` |
| Error Rate           |       `< 0.1%` |        `> 1%` |
| Throughput           | `>= 500 req/s` | `< 200 req/s` |

## Notes

* `tests/load/http/clank.yaml` is a shared reference config.
* `tests/load/http/run.sh` uses CLI flags because token and room id are dynamic.
* WebSocket load testing should use a WebSocket-capable tool such as k6 or websocat.
* gRPC load testing should use a gRPC-capable tool such as ghz.
* Do not commit fake result JSON files. Commit real baseline results only after running tests against a known environment.