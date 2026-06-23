#!/usr/bin/env bash

set -euo pipefail

PROFILE="${MINIKUBE_PROFILE:-pylon}"
NAMESPACE="${NAMESPACE:-pylon}"
CPUS="${MINIKUBE_CPUS:-4}"
MEMORY="${MINIKUBE_MEMORY:-8192}"
TIMEOUT="${KUBECTL_TIMEOUT:-180s}"

require_command() {
  local name="$1"

  if ! command -v "$name" >/dev/null 2>&1; then
    echo "missing required command: $name" >&2
    exit 1
  fi
}

wait_for_pods() {
  local label="$1"
  local description="$2"

  echo "waiting for $description"
  kubectl wait \
    --for=condition=ready pod \
    -l "$label" \
    -n "$NAMESPACE" \
    --timeout="$TIMEOUT"
}

rollout() {
  local deployment="$1"

  echo "waiting for deployment/$deployment"
  kubectl rollout status "deployment/$deployment" \
    -n "$NAMESPACE" \
    --timeout="$TIMEOUT"
}

main() {
  require_command minikube
  require_command kubectl
  require_command docker
  require_command make

  if ! minikube -p "$PROFILE" status >/dev/null 2>&1; then
    minikube start -p "$PROFILE" --cpus="$CPUS" --memory="$MEMORY"
  fi

  minikube -p "$PROFILE" update-context
  kubectl config use-context "$PROFILE"

  minikube -p "$PROFILE" addons enable ingress
  minikube -p "$PROFILE" addons enable metrics-server

  eval "$(minikube -p "$PROFILE" docker-env --shell bash)"

  make docker-build

  kubectl apply -k deploy/minikube

  wait_for_pods "app=postgres" "PostgreSQL"
  wait_for_pods "app=redis" "Redis"
  wait_for_pods "app=zookeeper" "Zookeeper"
  wait_for_pods "app=kafka" "Kafka"

  kubectl create configmap pylon-migrations \
    --from-file=migrations \
    -n "$NAMESPACE" \
    --dry-run=client \
    -o yaml | kubectl apply -f -

  kubectl delete job pylon-migrate -n "$NAMESPACE" --ignore-not-found
  kubectl apply -f deploy/minikube/migrate-job.yaml
  kubectl wait \
    --for=condition=complete job/pylon-migrate \
    -n "$NAMESPACE" \
    --timeout="$TIMEOUT"

  kubectl logs job/pylon-migrate -n "$NAMESPACE"

  kubectl apply -k deploy/overlays/dev

  kubectl rollout restart deployment/api-gateway -n "$NAMESPACE"
  kubectl rollout restart deployment/chat-service -n "$NAMESPACE"
  kubectl rollout restart deployment/presence-service -n "$NAMESPACE"
  kubectl rollout restart deployment/room-service -n "$NAMESPACE"
  kubectl rollout restart deployment/notification-service -n "$NAMESPACE"
  
  rollout api-gateway
  rollout chat-service
  rollout presence-service
  rollout room-service
  rollout notification-service

  echo
  echo "Pylon deployed to Minikube."
  echo
  echo "API Gateway URL:"
  minikube -p "$PROFILE" service api-gateway -n "$NAMESPACE" --url

  echo
  echo "Health check:"
  local api_url
  api_url="$(minikube -p "$PROFILE" service api-gateway -n "$NAMESPACE" --url | head -n 1)"
  curl -sS "$api_url/health"
  echo
}

main "$@"