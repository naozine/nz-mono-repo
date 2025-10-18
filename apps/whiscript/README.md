# Whiscript

A minimal CRUD application for managing projects, built with Echo framework and Standard Go Project Layout.

## Features

- Full CRUD operations for projects (Create, Read, Update, Delete)
- Modern, responsive UI with htmx and Alpine.js
- SQLite database with automatic migrations
- No authentication required
- Single binary deployment with embedded templates and migrations

## Tech Stack

- **Backend**: Echo (Go web framework)
- **Database**: SQLite with goose migrations
- **Frontend**: htmx, Alpine.js, Tailwind CSS (all via CDN)
- **Architecture**: Standard Go Project Layout with separation of concerns (handler/service/repository/model)

## Project Structure

```
apps/whiscript/
├── cmd/
│   └── whiscript/
│       └── main.go           # Application entry point
├── internal/
│   ├── handler/              # HTTP handlers (presentation layer)
│   ├── service/              # Business logic layer
│   ├── repository/           # Data access layer
│   ├── model/                # Domain models and DTOs
│   ├── ui/
│   │   ├── embed.go          # Template embedder
│   │   └── templates/        # HTML templates
│   └── db/
│       ├── embed.go          # Migration embedder
│       └── migrations/       # Database migrations (goose)
├── go.mod
├── Makefile
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.23 or higher
- Make (optional, for convenience commands)

### Installation

1. Navigate to the whiscript directory:
```bash
cd apps/whiscript
```

2. Install dependencies:
```bash
go mod download
```

### Running the Application

#### Development Mode

```bash
make dev
```

Or without Make:
```bash
DB_PATH=./whiscript.db PORT=8080 go run .
```

The application will start on `http://localhost:8080`

#### Building a Binary

```bash
make build
```

This creates a binary at `bin/whiscript`. Run it with:
```bash
./bin/whiscript
```

### Available Make Commands

- `make help` - Show all available commands
- `make dev` - Run in development mode
- `make build` - Build production binary
- `make test` - Run tests
- `make clean` - Clean build artifacts and database
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback last migration
- `make migrate-status` - Show migration status

## Usage

### Managing Projects

1. **View Projects**: Navigate to `http://localhost:8080` to see all projects
2. **Create Project**: Click "Create New Project" and fill in the form
3. **Edit Project**: Click the edit icon on any project card
4. **Delete Project**: Click the delete icon (with confirmation)

All operations are performed asynchronously using htmx, providing a smooth user experience without page reloads.

## Configuration

Environment variables:

- `DB_PATH` - Database file path (default: `./whiscript.db`)
- `PORT` - Server port (default: `8080`)

Example:
```bash
DB_PATH=/data/whiscript.db PORT=3000 ./bin/whiscript
```

## Database Migrations

Migrations are automatically run on application startup. The migration files are embedded in the binary using `go:embed`.

Manual migration commands (requires goose CLI):
```bash
# Status
make migrate-status

# Up
make migrate-up

# Down
make migrate-down
```

## Development

### Adding New Features

1. Define models in `internal/model/`
2. Add database operations in `internal/repository/`
3. Implement business logic in `internal/service/`
4. Create HTTP handlers in `internal/handler/`
5. Update templates in `templates/`

### Running Tests

```bash
make test
```

## License

MIT
