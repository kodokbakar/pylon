ARG GO_VERSION=1.26.4

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/bin/notification-service ./cmd/notification-service

FROM alpine:3.19

RUN apk --no-cache add ca-certificates \
    && addgroup -S pylon \
    && adduser -S pylon -G pylon

WORKDIR /app

COPY --from=builder --chown=pylon:pylon /app/bin/notification-service ./notification-service
COPY --from=builder --chown=pylon:pylon /app/migrations ./migrations

USER pylon

EXPOSE 9004

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:9004/health || exit 1

CMD ["./notification-service"]