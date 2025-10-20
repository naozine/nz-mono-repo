.PHONY: dev build migrate-up migrate-down migrate-status migrate-new clean

# Development
dev:
	go run ./cmd/whiscript

# Build
build:
	go build -o bin/whiscript ./cmd/whiscript

# Migration commands
migrate-up:
	@echo "Running migrations..."
	@DB_PATH=./whiscript.db go run ./cmd/whiscript migrate up

migrate-down:
	@echo "Rolling back migrations..."
	@goose -dir migrations sqlite3 ./whiscript.db down

migrate-status:
	@echo "Migration status..."
	@goose -dir migrations sqlite3 ./whiscript.db status

migrate-new:
	@if [ -z "$(name)" ]; then \
		echo "Error: name parameter is required. Usage: make migrate-new name=create_something"; \
		exit 1; \
	fi
	@goose -dir migrations create $(name) sql

# Clean
clean:
	rm -f bin/whiscript
	rm -f whiscript.db
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Help
help:
	@echo "Available commands:"
	@echo "  make dev              - Run the application in development mode"
	@echo "  make build            - Build the application binary"
	@echo "  make migrate-up       - Run database migrations"
	@echo "  make migrate-down     - Rollback last migration"
	@echo "  make migrate-status   - Show migration status"
	@echo "  make migrate-new name=<name> - Create a new migration file"
	@echo "  make clean            - Clean build artifacts and database"
	@echo "  make deps             - Download and tidy dependencies"
	@echo "  make test             - Run tests"
