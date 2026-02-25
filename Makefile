# Top-level development commands

.PHONY: help build up down logs test lint proto clean

help:
	@echo "Available commands:"
	@echo "  make up          - Start all services"
	@echo "  make down        - Stop all services"
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

logs:
	docker-compose logs -f

test:
	@echo "Running tests..."
	# Add test commands for all services

lint:
	@echo "Running linters..."
	# Add lint commands for all services

proto:
	./scripts/proto-gen.sh

migrate:
	./scripts/migrate.sh

seed:
	./scripts/seed.sh

health:
	./scripts/healthcheck.sh
