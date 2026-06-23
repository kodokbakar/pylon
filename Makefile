.PHONY: dev proto lint test test-integration test-e2e test-load build docker-build docker-build-gateway docker-build-chat docker-build-presence docker-build-room docker-build-notification minikube-deploy minikube-status minikube-clean help

# Help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
dev: ## Start local development environment (PostgreSQL, Redis, Kafka)
	docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Services ready:"
	@echo "  PostgreSQL: localhost:5433"
	@echo "  Redis:      localhost:6380"
	@echo "  Kafka:      localhost:9092"
	@echo "  Kafka UI:   localhost:8085"

dev-down: ## Stop local development environment
	docker compose down

dev-logs: ## Show logs from development environment
	docker compose logs -f

# Proto
proto: ## Generate protobuf code
	cd proto && buf generate --template buf.gen.yaml

proto-lint: ## Lint protobuf definitions
	cd proto && buf lint

proto-breaking: ## Check for breaking changes
	cd proto && buf breaking --against '.git#branch=main'

# Code Quality
lint: ## Run golangci-lint
	golangci-lint run ./...

fmt: ## Format Go code
	go fmt ./...
	goimports -w .

vet: ## Run go vet
	go vet ./...

# Testing
test: ## Run all tests
	go test -v ./...

test-cover: ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests with test infrastructure
	scripts/test-integration.sh

test-e2e: ## Run E2E tests against PYLON_E2E_BASE_URL
	go test -tags=e2e -v ./tests/e2e/...

test-load: ## Run HTTP load tests with clank-cli
	tests/load/http/run.sh

# Build
build: ## Build all services
	go build -o bin/api-gateway ./cmd/api-gateway
	go build -o bin/chat-service ./cmd/chat-service
	go build -o bin/presence-service ./cmd/presence-service
	go build -o bin/room-service ./cmd/room-service
	go build -o bin/notification-service ./cmd/notification-service

build-gateway: ## Build API gateway
	go build -o bin/api-gateway ./cmd/api-gateway

build-chat: ## Build chat service
	go build -o bin/chat-service ./cmd/chat-service

build-presence: ## Build presence service
	go build -o bin/presence-service ./cmd/presence-service

build-room: ## Build room service
	go build -o bin/room-service ./cmd/room-service

build-notification: ## Build notification service
	go build -o bin/notification-service ./cmd/notification-service

# Docker
docker-build: docker-build-gateway docker-build-chat docker-build-presence docker-build-room docker-build-notification ## Build all Docker images

docker-build-gateway: ## Build API Gateway Docker image
	docker build -f cmd/api-gateway/Dockerfile -t pylon/api-gateway .

docker-build-chat: ## Build Chat Service Docker image
	docker build -f cmd/chat-service/Dockerfile -t pylon/chat-service .

docker-build-presence: ## Build Presence Service Docker image
	docker build -f cmd/presence-service/Dockerfile -t pylon/presence-service .

docker-build-room: ## Build Room Service Docker image
	docker build -f cmd/room-service/Dockerfile -t pylon/room-service .

docker-build-notification: ## Build Notification Service Docker image
	docker build -f cmd/notification-service/Dockerfile -t pylon/notification-service .

# Minikube
minikube-deploy: ## Build images and deploy Pylon to Minikube
	scripts/minikube-deploy.sh

minikube-status: ## Show Pylon Minikube resources
	scripts/minikube-status.sh

minikube-clean: ## Delete Pylon namespace and stop Minikube
	scripts/minikube-clean.sh

# Database
migrate-up: ## Run database migrations
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down: ## Rollback database migrations
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create new migration (usage: make migrate-create name=add_users)
	migrate create -ext sql -dir migrations -seq $(name)

# Dependencies
deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

# Clean
clean: ## Remove build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html
