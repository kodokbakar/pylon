ARG GO_VERSION=1.26.4

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api-gateway ./cmd/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/chat-service ./cmd/chat-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/presence-service ./cmd/presence-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/room-service ./cmd/room-service
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/notification-service ./cmd/notification-service

FROM alpine:3.24 AS runtime

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S pylon \
    && adduser -S pylon -G pylon

USER pylon

FROM runtime AS gateway
COPY --from=builder /out/api-gateway /app/api-gateway
EXPOSE 8080
ENTRYPOINT ["/app/api-gateway"]

FROM runtime AS chat
COPY --from=builder /out/chat-service /app/chat-service
EXPOSE 9001
ENTRYPOINT ["/app/chat-service"]

FROM runtime AS presence
COPY --from=builder /out/presence-service /app/presence-service
EXPOSE 9002
ENTRYPOINT ["/app/presence-service"]

FROM runtime AS room
COPY --from=builder /out/room-service /app/room-service
EXPOSE 9003
ENTRYPOINT ["/app/room-service"]

FROM runtime AS notification
COPY --from=builder /out/notification-service /app/notification-service
EXPOSE 9004
ENTRYPOINT ["/app/notification-service"]