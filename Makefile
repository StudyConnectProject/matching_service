.PHONY: help build run run-local stop clean test lint install-deps migrate up down logs

help:
	@echo "Matching Service - Available commands:"
	@echo "  make install-deps    - Install Go dependencies"
	@echo "  make build          - Build the Docker image"
	@echo "  make run            - Start services with docker-compose"
	@echo "  make run-local      - Run service locally (requires local PostgreSQL)"
	@echo "  make stop           - Stop docker-compose services"
	@echo "  make clean          - Remove built binaries and containers"
	@echo "  make logs           - Show docker-compose logs"
	@echo "  make test           - Run tests"
	@echo "  make lint           - Run linter"
	@echo "  make migrate        - Run database migrations"

install-deps:
	go mod download
	go mod tidy

build:
	docker build -f docker/Dockerfile -t studyconnect/matching-service:latest .

run: build
	docker-compose up -d

run-local:
	@if [ ! -f .env ]; then cp .env.example .env; fi
	go run cmd/server/main.go cmd/server/config.go

stop:
	docker-compose down

clean:
	docker-compose down -v
	rm -f server
	go clean

logs:
	docker-compose logs -f

test:
	go test -v ./...

lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

migrate-up:
	docker-compose exec postgres psql -U postgres -d matching_db -f /docker-entrypoint-initdb.d/001_init.sql

db-shell:
	docker-compose exec postgres psql -U postgres -d matching_db

ps:
	docker-compose ps

version:
	@echo "Matching Service v1.0.0"

# Development targets
dev: build
	@echo "Starting matching service in development mode..."
	docker-compose up

dev-logs:
	docker-compose logs -f matching_service

# Production-like build
prod-build:
	docker build -f docker/Dockerfile -t studyconnect/matching-service:latest --build-arg VERSION=1.0.0 .

# Format code
fmt:
	go fmt ./...

# Check for vulnerabilities
security:
	go list -json -m all | nancy sleuth

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	mockgen -source=internal/domain/repository.go -destination=internal/mocks/mock_repository.go -package=mocks
