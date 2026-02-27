# Top-level development commands

.PHONY: help build up down up-infra logs test lint proto migrate seed health

help:
	@echo "Available commands:"
	@echo "  make up          - Start all services (services + infra)"
	@echo "  make down        - Stop all services"
	@echo "  make up-infra    - Start only infrastructure"
	@echo "  make logs        - View logs"
	@echo "  make test        - Run tests"
	@echo "  make lint        - Run linters"
	@echo "  make proto       - Generate protobuf code"
	@echo "  make migrate     - Run database migrations"
	@echo "  make seed        - Seed development data"
	@echo "  make health      - Check service health"

up:
	docker-compose up -d

down:
	docker-compose down

up-infra:
	docker-compose -f docker-compose.infra.yml up -d

logs:
	docker-compose logs -f

test:
	@echo "Running tests..."
	go test ./backend/...

lint:
	@echo "Running linters..."
	golangci-lint run

proto:
	./scripts/proto-gen.sh

migrate:
	./scripts/migrate.sh

seed:
	./scripts/seed.sh

health:
	./scripts/healthcheck.sh
