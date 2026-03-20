# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o books-api .

# Build the seed tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/seed ./cmd/seed

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/books-api .

# Copy the seed binary
COPY --from=builder /app/bin/seed ./bin/seed

# Copy migration files
COPY --from=builder /app/db/migrations ./db/migrations

# Expose port
EXPOSE 8080

# Run the application
CMD ["./books-api"]
