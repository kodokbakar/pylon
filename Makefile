.PHONY: dev proto lint test build help

# Help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Development
dev: ## Start local development environment (PostgreSQL, Redis, Kafka)
	docker compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 5
	@echo "Services ready:"
	@echo "  PostgreSQL: localhost:5432"
	@echo "  Redis:      localhost:6379"
	@echo "  Kafka:      localhost:9092"
	@echo "  Kafka UI:   localhost:8080"

dev-down: ## Stop local development environment
	docker compose down

dev-logs: ## Show logs from development environment
	docker compose logs -f

# Proto
proto: ## Generate protobuf code
	cd proto && buf generate --path . --template buf.gen.yaml --output ..

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
docker-build: ## Build Docker images
	docker build -t pylon/api-gateway -f Dockerfile --target gateway .
	docker build -t pylon/chat-service -f Dockerfile --target chat .
	docker build -t pylon/presence-service -f Dockerfile --target presence .
	docker build -t pylon/room-service -f Dockerfile --target room .
	docker build -t pylon/notification-service -f Dockerfile --target notification .

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
