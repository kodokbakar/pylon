#!/usr/bin/env bash

set -euo pipefail

PROFILE="${MINIKUBE_PROFILE:-pylon}"
NAMESPACE="${NAMESPACE:-pylon}"

if command -v kubectl >/dev/null 2>&1; then
  kubectl delete namespace "$NAMESPACE" --ignore-not-found
fi

if command -v minikube >/dev/null 2>&1; then
  minikube -p "$PROFILE" stop
fi