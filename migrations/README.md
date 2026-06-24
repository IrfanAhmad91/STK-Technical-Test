# Database Migrations

This directory contains SQL migration files for the Hierarchical Menu Tree System database schema.

## Migration Files

### 001_create_menu_items_table.sql

**Purpose**: Initial database schema setup for the menu_items table

**Subtasks Implemented**:
- **Subtask 1.1**: Create PostgreSQL database and menu_items table
  - Adjacency list schema with columns: id, name, parent_id, position, created_at, updated_at
  - Foreign key constraint with CASCADE DELETE for parent_id
  - CHECK constraints for name length (1-255 chars) and position (non-negative)
  - Indexes: idx_parent_id, idx_position, idx_parent_position

- **Subtask 1.2**: Implement database trigger for updated_at timestamp
  - PL/pgSQL function: `update_updated_at_column()`
  - Trigger: `update_menu_items_updated_at` (BEFORE UPDATE)

**Requirements Satisfied**: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6

### 001_create_menu_items_table_down.sql

**Purpose**: Rollback migration 001 (drop table, trigger, and function)

## Usage

### Prerequisites

1. PostgreSQL 12+ installed and running
2. Database created (e.g., `menu_system_db`)
3. Database user with appropriate permissions

### Running Migrations

#### Option 1: Using psql (PostgreSQL CLI)

```bash
# Apply migration
psql -U your_username -d menu_system_db -f migrations/001_create_menu_items_table.sql

# Rollback migration (if needed)
psql -U your_username -d menu_system_db -f migrations/001_create_menu_items_table_down.sql
```

#### Option 2: Using golang-migrate tool

```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Apply all migrations
migrate -path ./migrations -database "postgres://username:password@localhost:5432/menu_system_db?sslmode=disable" up

# Rollback last migration
migrate -path ./migrations -database "postgres://username:password@localhost:5432/menu_system_db?sslmode=disable" down 1
```

#### Option 3: Direct SQL execution in application

```go
// Example Go code using database/sql
func RunMigrations(db *sql.DB) error {
    migrationSQL, err := os.ReadFile("migrations/001_create_menu_items_table.sql")
    if err != nil {
        return fmt.Errorf("failed to read migration file: %w", err)
    }
    
    _, err = db.Exec(string(migrationSQL))
    if err != nil {
        return fmt.Errorf("failed to execute migration: %w", err)
    }
    
    return nil
}
```

### Verification

After running the migration, verify the schema:

```sql
-- Check table structure
\d menu_items

-- Check indexes
\di menu_items*

-- Check constraints
\d+ menu_items

-- Check trigger
\df update_updated_at_column

-- Verify trigger is attached
SELECT trigger_name, event_manipulation, event_object_table 
FROM information_schema.triggers 
WHERE event_object_table = 'menu_items';
```

Expected output:
- Table `menu_items` with 6 columns (id, name, parent_id, position, created_at, updated_at)
- 3 indexes: `idx_parent_id`, `idx_position`, `idx_parent_position`
- 2 CHECK constraints: `chk_name_length`, `chk_position_non_negative`
- 1 foreign key constraint: `fk_parent`
- 1 trigger: `update_menu_items_updated_at`
- 1 function: `update_updated_at_column()`

### Database Configuration

Create a `.env` file or configure your database connection:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=menu_system_db
DB_SSLMODE=disable
```

### Testing the Schema

```sql
-- Insert root-level menu items
INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Home', NULL, 0);

INSERT INTO menu_items (name, parent_id, position) 
VALUES ('About', NULL, 1);

-- Insert child menu items
INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Team', (SELECT id FROM menu_items WHERE name = 'About'), 0);

-- Verify cascade delete
DELETE FROM menu_items WHERE name = 'About';
-- This should also delete 'Team'

-- Verify updated_at trigger
UPDATE menu_items SET name = 'Homepage' WHERE name = 'Home';
SELECT name, created_at, updated_at FROM menu_items WHERE name = 'Homepage';
-- updated_at should be more recent than created_at
```

## Migration Strategy

1. **Development**: Apply migrations directly using psql or migration tool
2. **Testing**: Use automated migration scripts in CI/CD pipeline
3. **Production**: Use migration tools with version tracking (e.g., golang-migrate)

## Best Practices

- Always test migrations in development environment first
- Create rollback migrations for every forward migration
- Never modify existing migration files after they've been applied to production
- Document breaking changes and data migrations clearly
- Use transactions where possible to ensure atomic operations

## Schema Design

The schema uses an **adjacency list model** for hierarchical data:

**Advantages**:
- Simple and intuitive structure
- Efficient INSERT, UPDATE, DELETE operations (O(1))
- Easy to maintain referential integrity with foreign keys
- Excellent for frequent write operations

**Trade-offs**:
- Recursive queries needed for subtree operations (PostgreSQL WITH RECURSIVE)
- Depth calculation requires traversal

**Why not nested sets?**
- Nested sets require O(n) updates for every insertion
- More complex to maintain
- Better for read-heavy workloads with infrequent updates

The system requirements specify frequent CRUD operations and drag-and-drop reordering, making the adjacency list model the optimal choice.
