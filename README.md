# Books API - High-Performance Load Testing

A production-ready RESTful API for managing books with pagination, Redis caching, and comprehensive load testing. Built with Go, Echo framework, PostgreSQL, Redis, and Docker.

## Features

- **Paginated CRUD Operations**: Create, Read, Update, Delete books with page-based pagination
- **Page-Based Pagination**: Default limit of 25 books per page, configurable up to 100
- **PostgreSQL Database**: Production-grade relational database with migrations
- **Redis Caching**: Distributed caching with 5-minute TTL for paginated results and individual books
- **OpenAPI 3.0 Schema**: Schema-first API design with auto-generated handlers via oapi-codegen
- **Swagger UI**: Interactive API documentation at `/swagger` and `/docs`
- **Type-Safe Queries**: sqlc-generated code for compile-time SQL validation
- **Database Migrations**: golang-migrate for version-controlled schema changes
- **Echo Framework**: Fast and minimalist web framework with request validation
- **Structured Logging**: Logrus-based logging with structured fields
- **Duplicate Detection**: 409 Conflict response for duplicate ISBNs
- **Graceful Shutdown**: Proper signal handling and connection draining
- **Docker Support**: Full stack containerization with Docker Compose
- **Environment Configuration**: .env-based configuration management
- **Load Testing**: k6 tests for baseline, CRUD, cache performance, spike, and stress testing
- **Zero Error Rate**: Tested under load with 100 concurrent users

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
| GET | `/books?page=1&limit=25` | Get paginated books (cached) |
| GET | `/books/:id` | Get a book by ID (cached) |
| PUT | `/books/:id` | Update a book |
| DELETE | `/books/:id` | Delete a book |

### Pagination Parameters

- **page** (optional, default: 1): Page number (1-indexed)
- **limit** (optional, default: 25): Books per page (1-100)

### Example Requests

```bash
# Get first page (25 books)
curl http://localhost:8080/books

# Get page 2 with 50 books per page
curl http://localhost:8080/books?page=2&limit=50

# Get page 5 with default limit
curl http://localhost:8080/books?page=5
```

### Example Response

```json
{
  "books": [
    {
      "id": 1,
      "title": "The Go Programming Language",
      "author": "Alan Donovan",
      "isbn": "978-0134190440"
    }
  ],
  "total": 10000,
  "page": 1,
  "limit": 25,
  "totalPages": 400
}
```

### Error Handling

- **400 Bad Request**: Invalid pagination parameters
- **409 Conflict**: Duplicate ISBN when creating a book
- **404 Not Found**: Book not found
- **500 Internal Server Error**: Server error

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

**Results**: 0% error rate, p(95)=13.33ms, p(99)=16.02ms

#### Full CRUD Load Test (ramping VUs, up to 100)
```bash
k6 run tests/load/books_load_test.js
```

**Results**: 0% error rate, 51,700 requests, p(95)=5.9ms, p(99)=8.24ms, 245 req/s

#### Stress Test (up to 200 VUs, 14 minutes)
```bash
k6 run tests/load/stress_test.js
```

#### Spike Test (sudden traffic spike to 200 VUs)
```bash
k6 run tests/load/spike_test.js
```

#### Cache Performance Test (pagination caching)
```bash
k6 run tests/load/cache_performance_test.js
```

**Tests**: Paginated list caching, individual book caching, cache hit performance

#### Custom Base URL
```bash
k6 run -e BASE_URL=http://your-server:8080 tests/load/books_load_test.js
```

### Load Test Results Summary

| Test | VUs | Duration | Requests | Error Rate | p(95) | p(99) |
|------|-----|----------|----------|-----------|-------|-------|
| Baseline | 10 | 30s | 300 | 0.00% | 13.33ms | 16.02ms |
| Full CRUD | 100 | 3m30s | 51,700 | 0.00% | 5.9ms | 8.24ms |
| Spike | 200 | 2m | - | <20% | <2000ms | - |
| Stress | 200 | 14m | - | <10% | <1000ms | - |

**Key Metrics**:
- ✅ Zero error rate under normal load (100 VUs)
- ✅ Sub-10ms response times with caching
- ✅ 245+ requests/second throughput
- ✅ Efficient pagination with Redis caching
- ✅ Duplicate ISBN handling (409 Conflict)

## OpenAPI & Swagger Documentation

The API uses OpenAPI 3.0 specification with auto-generated handlers via `oapi-codegen`.

### Access Swagger UI

Once the server is running, access the interactive API documentation:

- **Swagger UI**: http://localhost:8080/swagger
- **Alternative UI**: http://localhost:8080/docs
- **OpenAPI Spec**: http://localhost:8080/openapi.json

### Regenerate API Code from OpenAPI Spec

If you modify the OpenAPI specification (`api/openapi.yaml`), regenerate the server code:

```bash
make openapi-generate
```

This will update `internal/api/generated.go` with new types and handler interfaces.

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
├── api/
│   ├── openapi.yaml                   # OpenAPI 3.0 specification
│   └── oapi-codegen.yaml              # oapi-codegen configuration
├── cmd/
│   └── seed/
│       └── main.go                    # Database seeding tool (PostgreSQL)
├── db/
│   ├── migrations/                    # golang-migrate SQL files
│   │   ├── 000001_create_books_table.up.sql
│   │   ├── 000001_create_books_table.down.sql
│   │   ├── 000002_increase_isbn_size.up.sql
│   │   └── 000002_increase_isbn_size.down.sql
│   └── queries/                       # sqlc query definitions
│       └── books.sql                  # Pagination queries (ListBooksPaginated, CountBooks)
├── internal/
│   ├── api/
│   │   ├── generated.go               # oapi-codegen generated types & interfaces
│   │   ├── server.go                  # ServerInterface implementation with pagination
│   │   └── swagger.go                 # Swagger UI handler
│   ├── config/
│   │   └── config.go                  # Environment configuration
│   ├── database/                      # sqlc generated code
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── books.sql.go               # Generated pagination queries
│   │   ├── querier.go
│   │   └── migrate.go                 # Migration runner
│   ├── models/
│   │   └── book.go                    # API models
│   ├── repository/
│   │   ├── interface.go               # Repository interface with pagination
│   │   ├── postgres_repository.go     # PostgreSQL implementation
│   │   └── repository_test.go         # Repository tests
│   ├── cache/
│   │   ├── interface.go               # Cache interface (paginated only)
│   │   ├── redis_cache.go             # Redis implementation with pagination caching
│   │   └── redis_cache_test.go        # Cache tests
│   └── handlers/                      # (Deprecated - use api/server.go)
├── pkg/
│   ├── postgres/
│   │   └── postgres.go                # PostgreSQL connection helper
│   └── redis/
│       └── redis.go                   # Redis connection helper
├── tests/
│   └── load/                          # k6 load tests
│       ├── baseline_test.js           # 10 VUs, 30s (pagination validation)
│       ├── books_load_test.js         # Full CRUD with pagination (up to 100 VUs)
│       ├── cache_performance_test.js  # Pagination cache performance
│       ├── spike_test.js              # Spike to 200 VUs
│       └── stress_test.js             # Stress test up to 200 VUs
├── .env                               # Environment variables (gitignored)
├── .env.example                       # Environment template
├── docker-compose.yml                 # Docker services (PostgreSQL, Redis, App)
├── Dockerfile                         # Application container
├── sqlc.yaml                          # sqlc configuration
├── main.go                            # Application entry point
├── Makefile                           # Build automation
├── go.mod & go.sum                    # Go dependencies
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

The application uses **Redis** for distributed caching with pagination support:

### Paginated List Caching
- **GET /books?page=X&limit=Y**: Cached per page/limit combination (key: `books:page:{page}:limit:{limit}`)
- **Cache Key Format**: Unique key for each pagination combination ensures efficient cache hits
- **Example**: `books:page:1:limit:25`, `books:page:2:limit:50`

### Individual Book Caching
- **GET /books/:id**: Cached in Redis for 5 minutes per book ID (key: `book:{id}`)
- **Fast Lookups**: Single book requests return from cache in <1ms

### Cache Invalidation
- **POST /books**: Invalidates all paginated caches (`books:page:*`)
- **PUT /books/:id**: Invalidates specific book cache and all paginated caches
- **DELETE /books/:id**: Invalidates specific book cache and all paginated caches
- **Consistency**: Ensures data consistency across all cache layers

### Cache Configuration
- **Cache TTL**: Configurable via `CACHE_TTL` environment variable (default: 5 minutes)
- **Redis Connection**: Configured via `REDIS_HOST` and `REDIS_PORT` environment variables

## Performance Thresholds

The k6 tests include the following performance thresholds:

- **Baseline Test**: 95th percentile < 300ms, 99th percentile < 500ms
- **Load Test**: 95th percentile < 500ms, 99th percentile < 1000ms
- **Stress Test**: 95th percentile < 1000ms, 99th percentile < 2000ms
- **Cache Test**: Cached requests 95th percentile < 100ms

## Development Approach

This project follows best practices for high-performance API development:

### Schema-First Design
1. **OpenAPI 3.0 Specification**: Define API contract first (`api/openapi.yaml`)
2. **Code Generation**: Auto-generate types and handler interfaces with `oapi-codegen`
3. **Type Safety**: Compile-time validation of API contract
4. **Documentation**: Swagger UI automatically generated from spec

### Pagination Implementation
1. **Database Layer**: sqlc-generated pagination queries (LIMIT/OFFSET)
2. **Repository Layer**: Page-based pagination with total count
3. **Cache Layer**: Per-page caching with Redis
4. **API Layer**: Query parameter validation and response formatting

### Load Testing
1. **Baseline Test**: Validate API functionality under light load
2. **CRUD Test**: Full create/read/update/delete cycle with ramping VUs
3. **Cache Performance**: Measure cache hit performance
4. **Spike Test**: Sudden traffic spike handling
5. **Stress Test**: Extended load to find breaking points

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
