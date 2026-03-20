.PHONY: help build test run clean load-test seed docker-up docker-down docker-logs docker-seed migrate-up migrate-down migrate-create sqlc-generate openapi-generate

help:
	@echo "Available targets:"
	@echo "  build             - Build the application"
	@echo "  test              - Run all unit tests"
	@echo "  run               - Run the application locally"
	@echo "  seed-local        - Seed local PostgreSQL database"
	@echo "  clean             - Clean build artifacts"
	@echo "  clean-docker      - Remove Docker containers and volumes (PostgreSQL, Redis)"
	@echo "  clean-all         - Clean everything (artifacts + Docker volumes)"
	@echo "  load-test         - Run k6 load tests (requires running server)"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-up         - Start all services (PostgreSQL, Redis, App)"
	@echo "  docker-down       - Stop all services"
	@echo "  docker-logs       - View logs from all services"
	@echo "  docker-seed       - Seed database in Docker (requires built image)"
	@echo "  docker-seed-build - Rebuild app and seed database"
	@echo ""
	@echo "Database commands:"
	@echo "  migrate-up        - Run database migrations"
	@echo "  migrate-down      - Rollback database migrations"
	@echo "  migrate-create    - Create new migration (NAME=migration_name)"
	@echo "  sqlc-generate     - Generate sqlc code from queries"
	@echo "  openapi-generate  - Generate OpenAPI server code from spec"

build:
	@echo "Building application..."
	go build -o books-api

test:
	@echo "Running unit tests..."
	go test ./... -v

run:
	@echo "Starting Books API server on :8080..."
	go run main.go

seed:
	@echo "Seeding database with fake books (local PostgreSQL)..."
	@echo "Note: Make sure PostgreSQL is running locally or use 'make docker-seed-build'"
	@go build -o bin/seed ./cmd/seed
	@./bin/seed

seed-local:
	@echo "Seeding local PostgreSQL (requires local PostgreSQL on localhost:5432)..."
	@POSTGRES_HOST=localhost go build -o bin/seed ./cmd/seed
	@POSTGRES_HOST=localhost ./bin/seed

clean:
	@echo "Cleaning up build artifacts..."
	rm -f books-api books.db test_*.db bin/seed

clean-docker:
	@echo "Cleaning up Docker containers and volumes..."
	docker compose down -v
	@echo "PostgreSQL and Redis data volumes removed"

clean-all: clean clean-docker
	@echo "All artifacts and Docker volumes cleaned"

load-test:
	@echo "Running k6 load tests..."
	@echo "Make sure the server is running on :8080"
	@echo ""
	@echo "Running baseline test..."
	k6 run tests/load/baseline_test.js
	@echo ""
	@echo "Running full CRUD load test..."
	k6 run tests/load/books_load_test.js

docker-up:
	@echo "Starting Docker services..."
	docker compose up -d
	@echo "Services started. Waiting for health checks..."
	@sleep 5
	@echo "Services ready!"

docker-down:
	@echo "Stopping Docker services..."
	docker compose down

docker-logs:
	@echo "Viewing logs..."
	docker compose logs -f

docker-seed:
	@echo "Seeding database in Docker..."
	@echo "Note: If this fails, run 'make docker-seed-build' first"
	docker compose exec app ./bin/seed

docker-seed-build:
	@echo "Rebuilding app with seed binary and seeding database..."
	docker compose up -d --build app
	@echo "Waiting for services to be ready..."
	@sleep 3
	docker compose exec app ./bin/seed

migrate-up:
	@echo "Running migrations..."
	migrate -path db/migrations -database "postgres://booksapi:booksapi_password@localhost:5432/booksapi?sslmode=disable" up

migrate-down:
	@echo "Rolling back migrations..."
	migrate -path db/migrations -database "postgres://booksapi:booksapi_password@localhost:5432/booksapi?sslmode=disable" down

migrate-create:
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir db/migrations -seq $(NAME)

sqlc-generate:
	@echo "Generating sqlc code..."
	sqlc generate

openapi-generate:
	@echo "Generating OpenAPI server code..."
	oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml
