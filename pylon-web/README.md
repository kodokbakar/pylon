# Pylon Web

![Runtime](https://img.shields.io/badge/runtime-Bun-black)
![React](https://img.shields.io/badge/React-19-20232a)
![Vite](https://img.shields.io/badge/Vite-8-646cff)
![TypeScript](https://img.shields.io/badge/TypeScript-6-3178c6)
![Tailwind CSS](https://img.shields.io/badge/Tailwind_CSS-4-38bdf8)
![Docker](https://img.shields.io/badge/Docker-ready-2496ed)

Frontend application for **Pylon**, a distributed realtime chat system built around room-based messaging, presence, WebSocket delivery, and Connect-RPC APIs.

Pylon Web provides the browser interface for:

- public landing, login, and registration flows
- authenticated room navigation
- room detail and member presence views
- realtime chat with optimistic sends
- WebSocket reconnect feedback
- token refresh and route protection
- production static serving through nginx

## Demo / Screenshot

Run the app locally, then open:

```text
http://localhost:3000
```

A screenshot or demo GIF can be added under `docs/` when a stable visual capture is available. This README intentionally avoids linking to a non-existent asset.

## Tech Stack

| Area                      | Technology                           |
| ------------------------- | ------------------------------------ |
| Runtime / package manager | Bun                                  |
| UI                        | React 19                             |
| Router                    | React Router 7                       |
| Build tool                | Vite 8                               |
| Language                  | TypeScript 6                         |
| Styling                   | Tailwind CSS 4, CSS-first setup      |
| Server state              | TanStack React Query 5               |
| RPC client                | Connect-RPC Web                      |
| Protobuf runtime          | `@bufbuild/protobuf`                 |
| Testing                   | Vitest, jsdom, React Testing Library |
| Formatting / linting      | Prettier, ESLint                     |
| Production runtime        | nginx on Alpine                      |
| Container build           | Multi-stage Docker build             |

## Requirements

Install these before running the frontend:

```text
Bun >= 1.3
Node.js >= 22
Docker + Docker Compose
```

The backend services are expected to run from the root Pylon monorepo.

## Repository Layout

```text
pylon/
  docker-compose.yml
  .env.example
  Dockerfile
  proto/
  services/
  cmd/
  migrations/

  pylon-web/
    Dockerfile
    nginx.conf
    package.json
    bun.lock
    vite.config.ts
    vitest.config.ts
    src/
```

## Setup

From the monorepo root:

```fish
cd pylon
cp .env.example .env
```

Install frontend dependencies:

```fish
cd pylon-web
bun install
```

Create a frontend env file if it does not exist:

```fish
cp .env.example .env
```

Expected frontend environment:

```env
VITE_API_URL=http://localhost:8080
```

## Environment Variables

### Frontend

| Variable       | Required | Default                 | Description                                                                                   |
| -------------- | -------- | ----------------------- | --------------------------------------------------------------------------------------------- |
| `VITE_API_URL` | No       | `http://localhost:8080` | Browser-facing API Gateway URL used by Connect-RPC, REST calls, and WebSocket URL generation. |

For local Docker Compose, keep this as `http://localhost:8080` because the browser runs on the host and cannot resolve Docker service names like `api-gateway`.

### Full-stack Compose

The root `.env.example` controls local compose ports and service credentials.

| Variable            | Default                      | Description                 |
| ------------------- | ---------------------------- | --------------------------- |
| `PYLON_WEB_PORT`    | `3000`                       | Host port for Pylon Web     |
| `PYLON_WEB_API_URL` | `http://localhost:8080`      | Build-time frontend API URL |
| `API_GATEWAY_PORT`  | `8080`                       | Host port for API Gateway   |
| `POSTGRES_PORT`     | `5433`                       | Host port for PostgreSQL    |
| `REDIS_PORT`        | `6380`                       | Host port for Redis         |
| `KAFKA_PORT`        | `9092`                       | Host port for Kafka         |
| `POSTGRES_USER`     | `pylon`                      | Local PostgreSQL user       |
| `POSTGRES_PASSWORD` | `pylon_dev`                  | Local PostgreSQL password   |
| `POSTGRES_DB`       | `pylon`                      | Local PostgreSQL database   |
| `JWT_SECRET`        | `pylon-dev-secret-change-me` | Local JWT signing secret    |

## Development

Start only the frontend dev server:

```fish
cd pylon-web
bun run dev
```

Default Vite dev URL:

```text
http://localhost:5173
```

Start the full stack with Docker Compose from the monorepo root:

```fish
cd pylon
docker compose up --build
```

Expected local URLs:

| Service     | URL                      |
| ----------- | ------------------------ |
| Pylon Web   | `http://localhost:3000`  |
| API Gateway | `http://localhost:8080`  |
| Kafka UI    | `http://localhost:8085`  |
| Jaeger UI   | `http://localhost:16686` |

## Available Scripts

Run from `pylon-web/`.

```fish
bun run dev
```

Start the Vite development server.

```fish
bun run build
```

Type-check and build production assets into `dist/`.

```fish
bun run preview
```

Preview the production build locally with Vite.

```fish
bun run lint
```

Run ESLint.

```fish
bun run format
```

Format files with Prettier.

```fish
bun run format:check
```

Check Prettier formatting.

```fish
bun run test
```

Run Vitest in watch mode.

```fish
bun run test:run
```

Run the full test suite once.

```fish
bun run proto
```

Regenerate frontend protobuf and Connect-RPC client code from the monorepo `proto/` directory.

## Verification

Before opening a PR or closing an issue, run:

```fish
cd pylon-web

bun run format
bun run lint
bun run test:run
bun run build
bun run format:check
```

For full-stack regression from the monorepo root:

```fish
cd pylon

go fmt ./...
go test ./...
make build

cd pylon-web
bun run format
bun run lint
bun run test:run
bun run build
bun run format:check
```

## Architecture

Pylon Web is a single-page React application. It communicates with the Pylon API Gateway using generated Connect-RPC clients, selected REST endpoints, and WebSocket connections.

```text
Browser
  |
  | HTTP / Connect-RPC / WebSocket
  v
API Gateway :8080
  |
  | internal service calls
  v
Room Service      :9003
Chat Service      :9001
Presence Service  :9002
Notification      :9004
  |
  v
PostgreSQL / Redis / Kafka
```

## Frontend Folder Structure

```text
src/
  api/
    auth.ts
    rooms.ts
    chat.ts
    presence.ts
    transport.ts
    fetch.ts
    config.ts
    authRefresh.ts
    gen/
      pylon/
        auth/v1/
        room/v1/
        presence/v1/

  assets/

  components/
    ErrorBoundary.tsx
    chat/
      MessageInput.tsx
      MessageItem.tsx
      MessageList.tsx
    layout/
      AppLayout.tsx
      ChatLayout.tsx
      PageHeader.tsx
      Sidebar.tsx
    presence/
      StatusIndicator.tsx
      TypingIndicator.tsx
    room/
      CreateRoomModal.tsx
      RoomItem.tsx
      RoomList.tsx
    ui/
      ConnectionStatus.tsx
      Skeleton.tsx

  context/
    authContext.ts
    AuthContext.tsx
    presenceContext.ts
    PresenceContext.tsx
    webSocketContext.ts
    WebSocketContext.tsx

  hooks/
    useAuth.ts
    useRooms.ts
    useRoom.ts
    useChatMessages.ts
    useRoomPresence.ts
    useWebSocket.ts
    useCreateRoom.ts
    useDocumentTitle.ts
    useLeaveRoom.ts
    usePresence.ts
    useRoomMembers.ts
    useSidebar.ts
    useStreamPresence.ts

  lib/
    queryClient.ts
    ws.ts

  pages/
    LandingPage.tsx
    Login.tsx
    Register.tsx
    HomePage.tsx
    RoomDetailPage.tsx
    ChatPage.tsx
    NotFoundPage.tsx

  routes/
    ProtectedRoute.tsx
    PublicRoute.tsx

  test/
    mocks/
    render.tsx
    setup.ts

  utils/
    authToken.ts
    backendError.ts
    format.ts
    object.ts
    tokenRefresh.ts
```

## Key Flows

### Authentication

1. Public users land on `/`.
2. `/login` and `/register` are public routes.
3. Successful login stores access token, refresh token, and user data in local storage.
4. Protected routes require an active token.
5. Access tokens are refreshed proactively before expiry.
6. Failed refresh clears the session and redirects to login.

### Room Navigation

1. Authenticated users enter the app shell.
2. The sidebar loads rooms through `RoomService.listRooms`.
3. Selecting a room navigates to the chat route.
4. Room detail pages load room metadata, members, and presence state.

### Chat

1. Chat history loads through the API Gateway REST message endpoint.
2. The WebSocket client joins the active room.
3. New messages are sent optimistically.
4. Realtime server messages are merged and deduplicated with optimistic messages.
5. Older messages can be paged with `Load older`.

### Presence

1. Room presence loads through the generated Presence Connect-RPC client.
2. Presence streams update online, offline, and typing states.
3. Typing indicators expire automatically after a short timeout.

### Realtime Resilience

The WebSocket client supports:

- manual connect/disconnect
- heartbeat ping/pong
- exponential reconnect backoff
- reconnect attempt metadata
- UI status banners for failures and retries

## Styling

Pylon Web uses a Swiss infrastructure-console visual direction:

- Archivo for primary UI typography
- IBM Plex Mono for metadata and operational labels
- paper/ink color system via CSS variables
- high-contrast borders
- hard shadows
- responsive layouts down to 320px
- visible focus styles
- reduced-motion support

Tailwind CSS v4 is wired through the official Vite plugin. There is intentionally no `tailwind.config.js` or `postcss.config.js`.

Relevant files:

```text
vite.config.ts
src/index.css
```

## Testing

The test stack uses:

- Vitest
- jsdom
- React Testing Library
- `@testing-library/jest-dom`
- `@testing-library/user-event`

Test helpers live in:

```text
src/test/setup.ts
src/test/render.tsx
src/test/mocks/
```

Run all frontend tests:

```fish
cd pylon-web
bun run test:run
```

## Docker Deployment

Pylon Web has a production multi-stage Dockerfile:

1. `oven/bun:1.3.14-alpine` builds the Vite app.
2. `nginx:alpine` serves the static `dist/` output.

Build the frontend image:

```fish
cd pylon-web

docker build \
  --build-arg VITE_API_URL=http://localhost:8080 \
  -t pylon-web .
```

Run the container:

```fish
docker run --rm -p 3000:8080 pylon-web
```

Open:

```text
http://localhost:3000
```

## nginx Runtime

The production nginx config provides:

- SPA routing fallback with `try_files $uri $uri/ /index.html`
- gzip compression
- long cache lifetime for hashed Vite assets
- no-cache policy for `index.html`
- security headers:
  - `X-Frame-Options: DENY`
  - `X-Content-Type-Options: nosniff`
  - `Referrer-Policy: strict-origin-when-cross-origin`

Relevant files:

```text
Dockerfile
nginx.conf
.dockerignore
```

## Full-stack Docker Compose

From the monorepo root:

```fish
cd pylon
docker compose up --build
```

This starts:

- `pylon-web`
- `api-gateway`
- `chat-service`
- `presence-service`
- `room-service`
- `notification-service`
- PostgreSQL
- Redis
- Kafka
- Kafka UI
- Jaeger
- OpenTelemetry Collector
- database migration job

Smoke test:

```fish
curl -I http://127.0.0.1:3000/
curl -I http://127.0.0.1:3000/rooms/smoke-test
curl -I http://127.0.0.1:8080/health
```

Expected:

```text
HTTP/1.1 200 OK
```

Stop the stack:

```fish
docker compose down
```

Remove local volumes when you need a clean database:

```fish
docker compose down -v
```

## Production Notes

Before production deployment:

- set a production API URL through `VITE_API_URL`
- set a strong backend `JWT_SECRET`
- configure strict backend CORS origins
- serve over HTTPS
- keep `index.html` no-cache
- keep hashed assets immutable
- do not use development secrets from `.env.example`

Example production build:

```fish
docker build \
  --build-arg VITE_API_URL=https://api.example.com \
  -t pylon-web:prod \
  pylon-web
```

## Troubleshooting

### Browser cannot reach `api-gateway`

Use `http://localhost:8080` for `VITE_API_URL` in local Docker Compose. The browser runs outside Docker, so Docker service names are not resolvable from the browser.

### Refreshing `/rooms/:id` returns 404

Use the nginx runtime config from `nginx.conf`. It falls back to `index.html` for client-side routes.

### Connect-RPC calls fail with CORS

Make sure the backend `CORS_ORIGINS` includes the frontend origin:

```text
http://localhost:3000
http://127.0.0.1:3000
http://localhost:5173
http://127.0.0.1:5173
```

### WebSocket stays disconnected

Check:

```fish
curl -I http://127.0.0.1:8080/health
docker compose logs --tail=80 api-gateway
```

Also verify that the user is authenticated and that the frontend build used the correct `VITE_API_URL`.

## Contributing

Recommended workflow:

1. Create a focused branch.
2. Keep changes small and reviewable.
3. Run format, lint, tests, and build before committing.
4. Update generated protobuf clients when proto definitions change.
5. Do not commit `.env`, local logs, `dist/`, or `node_modules`.

Frontend verification:

```fish
cd pylon-web

bun run format
bun run lint
bun run test:run
bun run build
bun run format:check
```

Suggested commit format:

```text
feat(pylon-web): add feature name
fix(pylon-web): fix bug name
test(pylon-web): cover behavior name
docs(pylon-web): update documentation
```
