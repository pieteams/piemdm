# PieMDM Makefile

.PHONY: help init mock test test-coverage build clean docker docker-push run dev deps lint format

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development setup
init: ## Initialize development environment
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golang/mock/mockgen@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go mod tidy

deps: ## Download dependencies
	go mod download
	go mod tidy

# Code generation
mock: ## Generate mock files
	# Repository mocks
	mockgen -source=internal/repository/entity.go -destination=test/mocks/repository/entity.go -package=mock_repository
	mockgen -source=internal/repository/table_field.go -destination=test/mocks/repository/table_field.go -package=mock_repository
	mockgen -source=internal/repository/table_approval_definition.go -destination=test/mocks/repository/table_approval_definition.go -package=mock_repository
	mockgen -source=internal/repository/approval_definition.go -destination=test/mocks/repository/approval_definition.go -package=mock_repository
	mockgen -source=internal/repository/approval_node.go -destination=test/mocks/repository/approval_node.go -package=mock_repository
	mockgen -source=internal/repository/approval.go -destination=test/mocks/repository/approval.go -package=mock_repository
	mockgen -source=internal/repository/approval_task.go -destination=test/mocks/repository/approval_task.go -package=mock_repository
	mockgen -source=internal/repository/application.go -destination=test/mocks/repository/application.go -package=mock_repository
	mockgen -source=internal/repository/user.go -destination=test/mocks/repository/user.go -package=mock_repository
	mockgen -source=internal/repository/user_role.go -destination=test/mocks/repository/user_role.go -package=mock_repository
	mockgen -source=internal/repository/notification_log.go -destination=test/mocks/repository/notification_log.go -package=mock_repository
	mockgen -source=internal/repository/notification_template.go -destination=test/mocks/repository/notification_template.go -package=mock_repository
	# Service mocks
	mockgen -source=internal/service/table_field.go -destination=test/mocks/service/table_field.go -package=mock_service
	mockgen -source=internal/service/global_id.go -destination=test/mocks/service/global_id.go -package=mock_service
	mockgen -source=internal/service/entity_log.go -destination=test/mocks/service/entity_log.go -package=mock_service
	mockgen -source=internal/service/autocode.go -destination=test/mocks/service/autocode.go -package=mock_service
	mockgen -source=internal/service/approval_workflow.go -destination=test/mocks/service/approval_workflow.go -package=mock_service
	mockgen -source=internal/service/approval.go -destination=test/mocks/service/approval.go -package=mock_service -aux_files=piemdm/internal/service=internal/service/approval_workflow.go
	mockgen -source=internal/service/approval_definition.go -destination=test/mocks/service/approval_definition.go -package=mock_service
	mockgen -source=internal/service/user.go -destination=test/mocks/service/user.go -package=mock_service
	mockgen -source=internal/service/role.go -destination=test/mocks/service/role.go -package=mock_service
	mockgen -source=internal/service/table_permission.go -destination=test/mocks/service/table_permission.go -package=mock_service
	mockgen -source=internal/service/application.go -destination=test/mocks/service/application.go -package=mock_service
	mockgen -destination=test/mocks/service/notification_service.go -package=mock_service piemdm/internal/service NotificationService


wire: ## Generate wire dependency injection
	cd cmd/server && wire

swagger: ## Generate swagger documentation
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal --exclude pkg/mod
	@echo "✅ Swagger documentation generated at docs/"

# Testing
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -coverpkg=./internal/handler,./internal/service,./internal/repository -coverprofile=./coverage.out ./test/server/...
	go tool cover -html=./coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	go test -race -v ./...

# Code quality
lint: ## Run linter
	golangci-lint run

format: ## Format code
	go fmt ./...
	goimports -w .

# Build
build: ## Build the application
	go build -ldflags="-s -w" -o ./bin/server ./cmd/server/

build-linux: ## Build for Linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server-linux ./cmd/server/

build-windows: ## Build for Windows
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ./bin/server.exe ./cmd/server/

# Development
run: ## Run the application
	go run ./cmd/server/

# dev: ## Run with hot reload (requires air)
# 	air

# Database
migrate-up: ## Run database migrations up
	go run ./cmd/migration/ up

migrate-down: ## Run database migrations down
	go run ./cmd/migration/ down

migrate-create: ## Create new migration (usage: make migrate-create name=migration_name)
	go run ./cmd/migration/ create $(name)

# Docker
docker: ## Build Docker image
	docker build -f backend/Dockerfile -t piemdm-api:latest ..

docker-dev: ## Build Docker image for development
	docker build -f backend/Dockerfile.dev -t piemdm-api:dev .

docker-push: ## Push Docker image
	docker push piemdm-api:latest

docker-run: ## Run Docker container
	docker run -p 8787:8787 --env-file .env piemdm-api:latest

# Docker Compose
compose-up: ## Start all services with docker-compose
	docker-compose up -d

compose-down: ## Stop all services
	docker-compose down

compose-logs: ## View logs
	docker-compose logs -f

compose-build: ## Build and start services
	docker-compose up --build -d

# Cleanup
clean: ## Clean build artifacts
	rm -rf ./bin/
	rm -f coverage.out coverage.html
	go clean -cache

clean-docker: ## Clean Docker images and containers
	docker system prune -f
	docker image prune -f

# Security
security-scan: ## Run security scan
	gosec ./...

# Performance
benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

# Release
release: clean test build ## Prepare release build
	@echo "Release build completed"

.DEFAULT_GOAL := help
