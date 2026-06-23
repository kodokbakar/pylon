#!/usr/bin/env bash

set -euo pipefail

PROFILE="${MINIKUBE_PROFILE:-pylon}"
NAMESPACE="${NAMESPACE:-pylon}"

kubectl config use-context "$PROFILE"

echo "===== nodes ====="
kubectl get nodes -o wide

echo
echo "===== namespace resources ====="
kubectl get all,ingress,hpa,cm,secret -n "$NAMESPACE"

echo
echo "===== api gateway url ====="
minikube -p "$PROFILE" service api-gateway -n "$NAMESPACE" --url

echo
echo "===== recent pod events ====="
kubectl get events -n "$NAMESPACE" --sort-by=.lastTimestamp | tail -n 30