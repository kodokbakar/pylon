#!/usr/bin/env bash

set -euo pipefail

COMPOSE_FILE="${COMPOSE_FILE:-docker-compose.test.yml}"
WAIT_TIMEOUT="${WAIT_TIMEOUT:-120}"

cleanup() {
  docker compose -f "$COMPOSE_FILE" down -v
}

trap cleanup EXIT

docker compose -f "$COMPOSE_FILE" up -d --wait --wait-timeout "$WAIT_TIMEOUT"

PYLON_TEST_DATABASE_URL="${PYLON_TEST_DATABASE_URL:-postgres://pylon:pylon_dev@localhost:15433/pylon?sslmode=disable}" \
PYLON_TEST_REDIS_URL="${PYLON_TEST_REDIS_URL:-redis://localhost:16380}" \
PYLON_TEST_KAFKA_BROKER="${PYLON_TEST_KAFKA_BROKER:-localhost:19092}" \
go test -tags=integration -v ./tests/...