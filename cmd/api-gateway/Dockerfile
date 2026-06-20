ARG GO_VERSION=1.26.4

FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/bin/api-gateway ./cmd/api-gateway

FROM alpine:3.19

RUN apk --no-cache add ca-certificates \
    && addgroup -S pylon \
    && adduser -S pylon -G pylon

WORKDIR /app

COPY --from=builder --chown=pylon:pylon /app/bin/api-gateway ./api-gateway
COPY --from=builder --chown=pylon:pylon /app/migrations ./migrations

USER pylon

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./api-gateway"]