# Whiscript

A minimal CRUD application built with Echo framework and Standard Go Project Layout.

## Features

- **Framework**: Echo v4 for HTTP routing and middleware
- **Database**: SQLite with sqlx for simplified database operations
- **Migrations**: goose for database migrations (embedded with go:embed)
- **Frontend**: htmx + _hyperscript + Tailwind CSS (all via CDN)
- **Templates**: html/template with embedded templates
- **Architecture**: Clean separation of concerns (handler/service/repository/model)

## Project Structure

```
apps/whiscript/
├── cmd/whiscript/          # Application entry point
│   └── main.go
├── internal/
│   ├── handler/            # HTTP handlers (Echo)
│   ├── service/            # Business logic and validation
│   ├── repository/         # Database operations (sqlx)
│   └── model/              # Data models
├── migrations/             # Database migrations (goose)
├── web/
│   ├── templates/          # HTML templates
│   └── static/             # Static assets (if needed)
├── Makefile
└── README.md
```

## Getting Started

### Prerequisites

- Go 1.23 or higher
- goose (for migration management)

### Installation

```bash
# Install goose
go install github.com/pressly/goose/v3/cmd/goose@latest

# Download dependencies
make deps
```

### Running the Application

```bash
# Run in development mode
make dev

# Or build and run
make build
./bin/whiscript
```

The application will start on `http://localhost:8080`

## Makefile Commands

- `make dev` - Run the application in development mode
- `make build` - Build the application binary
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback last migration
- `make migrate-status` - Show migration status
- `make migrate-new name=<name>` - Create a new migration file
- `make clean` - Clean build artifacts and database
- `make deps` - Download and tidy dependencies
- `make test` - Run tests

## Technology Stack

### Backend
- **Echo v4**: High performance HTTP framework
- **sqlx**: Extensions to Go's database/sql package
- **goose**: Database migration tool
- **SQLite**: Lightweight database

### Frontend (CDN)
- **Tailwind CSS**: Utility-first CSS framework
- **htmx**: High power tools for HTML
- **_hyperscript**: Event handling and DOM manipulation
- **Idiomorph**: Morphing algorithm for htmx swaps

## Features

### Project Management
- List all projects with pagination
- Create new projects
- Edit existing projects
- Delete projects
- Search projects by name or description
- Sort projects by name or date

### HTMX Integration
- Partial template rendering for HTMX requests
- Morphing DOM updates using Idiomorph
- Non-HTMX fallback for full page rendering

## Environment Variables

- `DB_PATH`: Path to SQLite database (default: `./whiscript.db`)
- `PORT`: Server port (default: `8080`)

## License

MIT
