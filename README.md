# Hierarchical Menu Tree System

A full-stack application for managing unlimited-depth nested menu structures with drag-and-drop reorganization capabilities.

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **Database**: PostgreSQL
- **API Documentation**: OpenAPI 3.0 with Swaggo

### Frontend (Coming in Task 4)
- **Framework**: React with JavaScript
- **State Management**: Zustand
- **Styling**: Tailwind CSS
- **UI Components**: Custom components with drag-and-drop

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── config/
│   └── database.go              # Database configuration and connection pool
├── internal/
│   ├── domain/
│   │   └── menu_item.go         # Domain models and request/response structs
│   ├── repository/
│   │   ├── menu_repository.go   # Repository interface
│   │   └── postgresql_repository.go  # PostgreSQL implementation
│   ├── service/                 # Business logic layer (Task 3)
│   └── handler/                 # HTTP handlers (Task 3)
├── migrations/
│   └── 001_create_menu_items_table.sql  # Database schema
├── docs/                        # API documentation
├── scripts/                     # Utility scripts
└── go.mod                       # Go module dependencies
```

## Prerequisites

- Go 1.21 or higher
- PostgreSQL 12 or higher
- Git

## Setup Instructions

### 1. Install Go

Download and install Go from [go.dev/dl](https://go.dev/dl)

Verify installation:
```bash
go version
```

### 2. Database Setup

Follow instructions in `DATABASE_SETUP.md` to:
- Install PostgreSQL
- Create the database
- Run migrations

### 3. Install Dependencies

After restarting your terminal (to pick up Go in PATH), run:

```bash
go mod tidy
```

This will download all required dependencies:
- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/lib/pq` - PostgreSQL driver
- `github.com/swaggo/swag` - OpenAPI documentation generator
- `github.com/swaggo/gin-swagger` - Swagger UI integration

### 4. Configure Environment

Create a `.env` file in the project root (optional, defaults provided):

```env
# Database configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=menu_tree_db
DB_SSLMODE=disable

# Server configuration
PORT=8080
```

### 5. Run the Application

```bash
# Run directly
go run cmd/api/main.go

# Or build and run
go build -o bin/menu-api cmd/api/main.go
./bin/menu-api
```

The API will be available at `http://localhost:8080`

### 6. Generate API Documentation (Task 3)

Once handlers are implemented:

```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
swag init -g cmd/api/main.go -o docs

# Access Swagger UI at http://localhost:8080/swagger/index.html
```

## API Endpoints (Task 3)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |
| GET | `/api/v1/menus` | Get all menu items |
| GET | `/api/v1/menus/:id` | Get specific menu item |
| POST | `/api/v1/menus` | Create menu item |
| PUT | `/api/v1/menus/:id` | Update menu item |
| DELETE | `/api/v1/menus/:id` | Delete menu item |
| PUT | `/api/v1/menus/:id/reorder` | Reorder within level |
| PUT | `/api/v1/menus/:id/move` | Move to different parent |
| GET | `/swagger/*` | OpenAPI documentation |

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Code Formatting

```bash
# Format code
go fmt ./...

# Run linter (install golangci-lint first)
golangci-lint run
```

### Database Migrations

```bash
# Apply migration
psql -U postgres -d menu_tree_db -f migrations/001_create_menu_items_table.sql

# Rollback migration
psql -U postgres -d menu_tree_db -f migrations/001_create_menu_items_table_down.sql
```



MIT
