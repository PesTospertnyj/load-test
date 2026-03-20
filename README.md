# Books API - TDD Implementation

A production-ready RESTful API for managing books, built using Test-Driven Development methodology with Go, Echo framework, PostgreSQL database, Redis caching, and Docker.

## Features

- **CRUD Operations**: Create, Read, Update, Delete books
- **PostgreSQL Database**: Production-grade relational database with migrations
- **Redis Caching**: Distributed caching with 5-minute TTL for GET requests
- **Type-Safe Queries**: sqlc-generated code for compile-time SQL validation
- **Database Migrations**: golang-migrate for version-controlled schema changes
- **Echo Framework**: Fast and minimalist web framework
- **Structured Logging**: Logrus-based logging with structured fields
- **Graceful Shutdown**: Proper signal handling and connection draining
- **Docker Support**: Full stack containerization with Docker Compose
- **Environment Configuration**: .env-based configuration management
- **No Authentication**: Simple API for testing purposes
- **Comprehensive Tests**: 100% test coverage for business logic

## Book Model

```json
{
  "id": 1,
  "title": "The Go Programming Language",
  "author": "Alan Donovan",
  "isbn": "978-0134190440"
}
```

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/books` | Create a new book |
| GET | `/books` | Get all books (cached) |
| GET | `/books/:id` | Get a book by ID (cached) |
| PUT | `/books/:id` | Update a book |
| DELETE | `/books/:id` | Delete a book |

## Prerequisites

### Option 1: Docker (Recommended)
- Docker and Docker Compose
- k6 (for load testing)

### Option 2: Local Development
- Go 1.25.5 or higher
- PostgreSQL 16+
- Redis 7+
- golang-migrate CLI
- sqlc CLI
- k6 (for load testing)

## Quick Start with Docker

```bash
# 1. Start all services (PostgreSQL, Redis, App)
make docker-up

# 2. Seed the database with 10,000 fake books
make docker-seed

# 3. Test the API
curl http://localhost:8080/books

# 4. View logs
make docker-logs

# 5. Stop services
make docker-down
```

The server will be available at `http://localhost:8080`

For detailed Docker setup, see [DOCKER_SETUP.md](DOCKER_SETUP.md)

## Local Development Setup

```bash
# 1. Clone or navigate to the project directory
cd /home/artur/Developer/load-test

# 2. Copy environment file
cp .env.example .env

# 3. Update .env with your local PostgreSQL and Redis settings
# POSTGRES_HOST=localhost
# REDIS_HOST=localhost

# 4. Install dependencies
go mod download

# 5. Install tools
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# 6. Generate sqlc code
make sqlc-generate

# 7. Run migrations
make migrate-up

# 8. Build the application
make build

# 9. Run the application
./books-api
```

The server will start on `http://localhost:8080`

## Seeding the Database

The project includes a seed script that generates 10,000 fake book records using the `go-faker` library:

```bash
# Using Make
make seed

# Or using the shell script
./seed.sh

# Or run directly
go run ./cmd/seed/main.go
```

The seed script will:
- Generate 10 books with realistic titles, author names, and ISBN numbers
- Use various genres (Programming, Science Fiction, Fantasy, Mystery, etc.)
- Create unique ISBN numbers in the format `978-X-XXXXX-XXX-X`
- Log each book creation with structured logging
- Display all books in the database after seeding

**Example output:**
```
INFO[2026-03-20T17:51:16+01:00] Starting database seeding...
INFO[2026-03-20T17:51:16+01:00] Generating 10 fake book records...
INFO[2026-03-20T17:51:16+01:00] Book created    author="Queen Liliane Ziemann" id=2 isbn=978-5-12848-922-7 title="Introduction to History"
...
INFO[2026-03-20T17:51:17+01:00] Database seeding completed successfully    total=11
```

## Running Tests

### Unit Tests

```bash
# Run all tests
go test ./... -v

# Run tests for specific package
go test ./internal/repository -v
go test ./internal/cache -v
go test ./internal/handlers -v

# Run tests with coverage
go test ./... -cover
```

### Load Tests with k6

First, ensure the API server is running, then execute the k6 tests:

#### Baseline Test (10 VUs, 30s)
```bash
k6 run tests/load/baseline_test.js
```

#### Full CRUD Load Test (ramping VUs)
```bash
k6 run tests/load/books_load_test.js
```

#### Stress Test (up to 200 VUs)
```bash
k6 run tests/load/stress_test.js
```

#### Spike Test (sudden traffic spike)
```bash
k6 run tests/load/spike_test.js
```

#### Cache Performance Test
```bash
k6 run tests/load/cache_performance_test.js
```

#### Custom Base URL
```bash
k6 run -e BASE_URL=http://your-server:8080 tests/load/books_load_test.js
```

## Example Usage

### Create a Book
```bash
curl -X POST http://localhost:8080/books \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Clean Code",
    "author": "Robert Martin",
    "isbn": "978-0132350884"
  }'
```

### Get All Books
```bash
curl http://localhost:8080/books
```

### Get Book by ID
```bash
curl http://localhost:8080/books/1
```

### Update a Book
```bash
curl -X PUT http://localhost:8080/books/1 \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Clean Code - Updated",
    "author": "Robert C. Martin",
    "isbn": "978-0132350884"
  }'
```

### Delete a Book
```bash
curl -X DELETE http://localhost:8080/books/1
```

## Project Structure

```
load-test/
├── cmd/
│   └── seed/
│       └── main.go                    # Database seeding tool (PostgreSQL)
├── db/
│   ├── migrations/                    # golang-migrate SQL files
│   │   ├── 000001_create_books_table.up.sql
│   │   └── 000001_create_books_table.down.sql
│   └── queries/                       # sqlc query definitions
│       └── books.sql
├── internal/
│   ├── config/
│   │   └── config.go                  # Environment configuration
│   ├── database/                      # sqlc generated code
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── books.sql.go
│   │   ├── querier.go
│   │   └── migrate.go                 # Migration runner
│   ├── models/
│   │   └── book.go                    # API models
│   ├── repository/
│   │   ├── interface.go               # Repository interface
│   │   ├── postgres_repository.go     # PostgreSQL implementation
│   │   └── repository_test.go         # Repository tests
│   ├── cache/
│   │   ├── interface.go               # Cache interface
│   │   ├── redis_cache.go             # Redis implementation
│   │   └── redis_cache_test.go        # Cache tests
│   └── handlers/
│       ├── handlers.go                # HTTP handlers
│       └── handlers_test.go           # Handler tests
├── pkg/
│   ├── postgres/
│   │   └── postgres.go                # PostgreSQL connection helper
│   └── redis/
│       └── redis.go                   # Redis connection helper
├── tests/
│   └── load/                          # k6 load tests
│       ├── books_load_test.js
│       ├── baseline_test.js
│       ├── stress_test.js
│       ├── spike_test.js
│       └── cache_performance_test.js
├── .env                               # Environment variables (gitignored)
├── .env.example                       # Environment template
├── docker-compose.yml                 # Docker services
├── Dockerfile                         # Application container
├── sqlc.yaml                          # sqlc configuration
├── main.go                            # Application entry point
├── Makefile                           # Build automation
├── DOCKER_SETUP.md                    # Docker guide
└── README.md                          # This file
```

## API Documentation

The API is documented using **OpenAPI 3.0** specification with automatic code generation.

### OpenAPI Specification

The OpenAPI spec is available at:
```
http://localhost:8080/openapi.json
```

### Schema-First Approach

This project uses **oapi-codegen** for schema-first API development:

1. **API Contract**: Defined in `api/openapi.yaml` (OpenAPI 3.0)
2. **Code Generation**: Server types and handlers generated from spec
3. **Type Safety**: Compile-time validation of API contract
4. **Request Validation**: Automatic validation against OpenAPI spec

### Regenerating API Code

After modifying `api/openapi.yaml`:

```bash
# Regenerate server code
make openapi-generate

# Or manually
oapi-codegen -config api/oapi-codegen.yaml api/openapi.yaml
```

### Disabling OpenAPI Validation

Set `SWAGGER_ENABLED=false` in `.env` to disable request validation middleware.

## Caching Strategy

The application uses **Redis** for distributed caching:

- **GET /books**: Cached in Redis for 5 minutes (key: `books:all`)
- **GET /books/:id**: Cached in Redis for 5 minutes per book ID (key: `book:{id}`)
- **POST /PUT /DELETE**: Invalidates all related caches to ensure consistency
- **Cache TTL**: Configurable via `CACHE_TTL` environment variable (default: 5m)

## Performance Thresholds

The k6 tests include the following performance thresholds:

- **Baseline Test**: 95th percentile < 300ms, 99th percentile < 500ms
- **Load Test**: 95th percentile < 500ms, 99th percentile < 1000ms
- **Stress Test**: 95th percentile < 1000ms, 99th percentile < 2000ms
- **Cache Test**: Cached requests 95th percentile < 100ms

## TDD Approach

This project was built using Test-Driven Development:

1. **Repository Layer**: Tests written first, then implementation
2. **Cache Layer**: Tests written first, then implementation
3. **Handlers Layer**: Tests written first, then implementation
4. **Integration**: Main application wiring
5. **Load Testing**: k6 scripts for performance validation

## Logging

The application uses **Logrus** for structured logging with the following features:

- **Structured Fields**: All logs include contextual information (method, URI, status, errors)
- **Environment-based Formatting**:
  - Development: Human-readable text format with timestamps
  - Production: JSON format (set `ENVIRONMENT=production`)
- **Request Logging**: Every HTTP request is logged with method, URI, and status code
- **Error Tracking**: Failed requests include error details

### Log Levels

- `INFO`: Application lifecycle events (startup, shutdown, initialization)
- `ERROR`: Request failures and runtime errors
- `FATAL`: Critical errors that prevent application startup

### Example Logs

```
INFO[2026-03-20T17:39:56+01:00] Starting Books API application
INFO[2026-03-20T17:39:56+01:00] Database initialized successfully
INFO[2026-03-20T17:39:56+01:00] Cache initialized with 5 minute TTL
INFO[2026-03-20T17:39:56+01:00] Server starting                               port=8080
INFO[2026-03-20T17:40:01+01:00] Request completed                             method=POST status=201 uri=/books
INFO[2026-03-20T17:40:02+01:00] Request completed                             method=GET status=200 uri=/books
```

## Graceful Shutdown

The application implements graceful shutdown to ensure:

1. **Signal Handling**: Listens for `SIGINT` (Ctrl+C) and `SIGTERM` signals
2. **Connection Draining**: Waits up to 10 seconds for active requests to complete
3. **Resource Cleanup**: Properly closes database connections
4. **Logging**: Records shutdown events

### Shutdown Process

```bash
# Press Ctrl+C to trigger graceful shutdown
^C
INFO[2026-03-20T17:45:00+01:00] Shutting down server gracefully...
INFO[2026-03-20T17:45:00+01:00] Server stopped
INFO[2026-03-20T17:45:00+01:00] Server exited gracefully
```

## Database

The application uses SQLite with a file named `books.db` in the project root. The database is automatically initialized on startup with the following schema:

```sql
CREATE TABLE IF NOT EXISTS books (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    isbn TEXT NOT NULL UNIQUE
);
```

## License

This is a demonstration project for TDD methodology and load testing practices.
