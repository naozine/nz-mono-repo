.PHONY: help dev build test clean migrate-up migrate-down migrate-status

# Default target
help:
	@echo "Available targets:"
	@echo "  make dev           - Run the application in development mode"
	@echo "  make build         - Build the application binary"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts and database"
	@echo "  make migrate-up    - Run database migrations"
	@echo "  make migrate-down  - Rollback last migration"
	@echo "  make migrate-status - Show migration status"

# Development mode
dev:
	@echo "Starting development server..."
	@DB_PATH=./whiscript.db PORT=8080 go run cmd/whiscript/main.go

# Build binary
build:
	@echo "Building whiscript..."
	@go build -o bin/whiscript cmd/whiscript/main.go
	@echo "Binary created at bin/whiscript"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean artifacts
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f whiscript.db
	@echo "Cleanup complete"

# Migration targets (for manual migration management)
migrate-up:
	@echo "Running migrations..."
	@DB_PATH=./whiscript.db go run cmd/whiscript/main.go

migrate-down:
	@echo "Rolling back last migration..."
	@goose -dir internal/db/migrations sqlite3 ./whiscript.db down

migrate-status:
	@echo "Migration status:"
	@goose -dir internal/db/migrations sqlite3 ./whiscript.db status
