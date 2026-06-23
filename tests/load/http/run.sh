#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
CLANK_BIN="${CLANK_BIN:-clank-cli}"
RESULTS_DIR="${RESULTS_DIR:-tests/load/results}"

LOAD_TEST_USERNAME="${LOAD_TEST_USERNAME:-loadtester}"
LOAD_TEST_EMAIL="${LOAD_TEST_EMAIL:-loadtester@example.com}"
LOAD_TEST_PASSWORD="${LOAD_TEST_PASSWORD:-password123}"

ROOM_NAME="${ROOM_NAME:-Load Test Room}"
ROOM_TYPE="${ROOM_TYPE:-channel}"

TOKEN="${TOKEN:-}"
ROOM_ID="${ROOM_ID:-}"

require_command() {
  local name="$1"

  if ! command -v "$name" >/dev/null 2>&1; then
    echo "missing required command: $name" >&2
    exit 1
  fi
}

request_json() {
  local method="$1"
  local path="$2"
  local body="$3"

  curl -sS \
    -X "$method" \
    "$BASE_URL$path" \
    -H "Content-Type: application/json" \
    --data "$body"
}

request_json_auth() {
  local method="$1"
  local path="$2"
  local body="$3"

  curl -sS \
    -X "$method" \
    "$BASE_URL$path" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    --data "$body"
}

login() {
  local response

  response="$(request_json "POST" "/api/v1/auth/login" "{\"email\":\"$LOAD_TEST_EMAIL\",\"password\":\"$LOAD_TEST_PASSWORD\"}")"

  TOKEN="$(printf '%s' "$response" | jq -r '.data.token // empty')"
  if [[ -z "$TOKEN" ]]; then
    echo "failed to login load test user" >&2
    printf '%s\n' "$response" >&2
    exit 1
  fi
}

register_or_login() {
  if [[ -n "$TOKEN" ]]; then
    echo "using TOKEN from environment"
    return
  fi

  local response
  response="$(request_json "POST" "/api/v1/auth/register" "{\"username\":\"$LOAD_TEST_USERNAME\",\"email\":\"$LOAD_TEST_EMAIL\",\"password\":\"$LOAD_TEST_PASSWORD\"}")"

  TOKEN="$(printf '%s' "$response" | jq -r '.data.token // empty')"
  if [[ -n "$TOKEN" ]]; then
    echo "registered load test user"
    return
  fi

  echo "load test user may already exist, trying login"
  login
}

create_room() {
  if [[ -n "$ROOM_ID" ]]; then
    echo "using ROOM_ID from environment: $ROOM_ID"
    return
  fi

  local suffix
  local response

  suffix="$(date +%s)"
  response="$(request_json_auth "POST" "/api/v1/rooms" "{\"name\":\"$ROOM_NAME $suffix\",\"type\":\"$ROOM_TYPE\"}")"

  ROOM_ID="$(printf '%s' "$response" | jq -r '.data.room.id // empty')"
  if [[ -z "$ROOM_ID" ]]; then
    echo "failed to create load test room" >&2
    printf '%s\n' "$response" >&2
    exit 1
  fi

  echo "created load test room: $ROOM_ID"
}

run_clank() {
  local name="$1"
  shift

  local output="$RESULTS_DIR/$name.json"

  echo "running $name -> $output"
  "$CLANK_BIN" "$@" -o json > "$output"
}

main() {
  require_command curl
  require_command jq
  require_command "$CLANK_BIN"

  mkdir -p "$RESULTS_DIR"

  register_or_login
  create_room

  run_clank "list_rooms" \
    "$BASE_URL/api/v1/rooms" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -c 20 \
    -d 60s \
    -r 100/s

  run_clank "get_messages" \
    "$BASE_URL/api/v1/rooms/$ROOM_ID/messages" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -c 20 \
    -d 60s \
    -r 100/s

  run_clank "create_room" \
    "$BASE_URL/api/v1/rooms" \
    -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    --body "{\"name\":\"Load Test Room $(date +%s)\",\"type\":\"channel\"}" \
    -c 5 \
    -d 30s \
    -r 10/s

  run_clank "join_room" \
    "$BASE_URL/api/v1/rooms/$ROOM_ID/join" \
    -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    --body "{}" \
    -c 10 \
    -d 30s \
    -r 20/s

  run_clank "leave_room" \
    "$BASE_URL/api/v1/rooms/$ROOM_ID/leave" \
    -X POST \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    --body "{}" \
    -c 10 \
    -d 30s \
    -r 20/s

  echo
  echo "load test results written to $RESULTS_DIR"
  echo "room id: $ROOM_ID"
}

main "$@"