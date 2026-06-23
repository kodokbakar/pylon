# Pylon Minikube Deployment

This directory contains Minikube-only infrastructure manifests and migration job support for Pylon.

## What This Adds

- PostgreSQL
- Redis
- Zookeeper
- Kafka
- Development Kubernetes Secret values
- Database migration Job

Application services are deployed from:

```text
deploy/overlays/dev
```

The dev overlay uses local Minikube Docker images:

```text
pylon/api-gateway:latest
pylon/chat-service:latest
pylon/presence-service:latest
pylon/room-service:latest
pylon/notification-service:latest
```

## Requirements

* minikube
* kubectl
* Docker
* make
* curl

## Deploy

```bash
make minikube-deploy
```

The deploy script will:

1. Start Minikube if it is not running.
2. Enable ingress and metrics-server addons.
3. Build service images inside Minikube Docker.
4. Deploy PostgreSQL, Redis, Zookeeper, and Kafka.
5. Create a ConfigMap from `migrations/`.
6. Run the database migration Job.
7. Deploy Pylon services with `deploy/overlays/dev`.
8. Print the API Gateway URL.
9. Call `/health`.

## Status

```bash
make minikube-status
```

## Clean

```bash
make minikube-clean
```

## Manual Health Check

```bash
API_URL="$(minikube -p pylon service api-gateway -n pylon --url | head -n 1)"
curl "$API_URL/health"
```

## Notes

This setup is for local Minikube only.

Do not use `deploy/minikube/secrets.yaml` in production.