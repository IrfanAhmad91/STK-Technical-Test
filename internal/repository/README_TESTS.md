# Repository Tests

This directory contains unit tests and integration tests for the PostgreSQL repository implementation, with a focus on recursive CTE queries for hierarchical operations.

## Test Files

### `postgresql_repository_test.go`
Unit tests using sqlmock to test repository methods in isolation. These tests:
- Mock database interactions
- Verify SQL query structure and parameters
- Test error handling and edge cases
- Validate context cancellation support
- Test transaction behavior

**Key test cases:**
- `TestGetDescendantIDs`: Tests recursive CTE query for retrieving all descendant node IDs
- `TestGetAncestors`: Tests recursive CTE query for retrieving all ancestor nodes
- `TestRecursiveCTEWithTransaction`: Tests that recursive queries work within transactions
- `TestCircularReferenceDetection`: Tests using GetAncestors to prevent circular references
- `TestCascadeDeleteUsingDescendants`: Tests using GetDescendantIDs for cascade operations
- `TestContextCancellation`: Tests that queries respect context cancellation

### `postgresql_repository_integration_test.go`
Integration tests that run against a real PostgreSQL database. These tests:
- Validate actual database behavior
- Test complex hierarchical structures
- Verify CASCADE delete behavior
- Test deep hierarchies (10+ levels)

**Key test cases:**
- `TestGetDescendantIDsIntegration`: Tests descendant retrieval with real hierarchical data
- `TestGetAncestorsIntegration`: Tests ancestor retrieval with real hierarchical data
- `TestCircularReferenceDetectionIntegration`: Tests detection of circular references
- `TestCascadeDeleteIntegration`: Tests cascade delete using GetDescendantIDs
- `TestDeepHierarchyIntegration`: Tests with 10-level deep hierarchy
- `TestRecursiveCTEWithTransactionIntegration`: Tests recursive CTEs within transactions

## Running Tests

### Prerequisites

1. **Go installed**: Ensure Go 1.25.0 or later is installed
2. **PostgreSQL database**: For integration tests, a PostgreSQL instance must be running
3. **Test dependencies**: Install required packages:
   ```bash
   go get github.com/stretchr/testify
   go get github.com/DATA-DOG/go-sqlmock
   ```

### Unit Tests

Unit tests use mocked database connections and can run without a database:

```bash
# Run all unit tests in the repository package
go test -v ./internal/repository

# Run specific test
go test -v ./internal/repository -run TestGetDescendantIDs

# Run with coverage
go test -v -cover ./internal/repository
```

### Integration Tests

Integration tests require a running PostgreSQL database:

1. **Setup test database:**
   ```bash
   # Create test database
   createdb menu_tree_test
   
   # Run migrations
   psql -d menu_tree_test -f migrations/001_create_menu_items_table.sql
   ```

2. **Configure database connection:**
   Set environment variables (or use defaults):
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=postgres
   export DB_NAME=menu_tree_test
   ```

3. **Run integration tests:**
   ```bash
   # Run integration tests
   RUN_INTEGRATION_TESTS=true go test -v -tags=integration ./internal/repository
   
   # Run specific integration test
   RUN_INTEGRATION_TESTS=true go test -v -tags=integration ./internal/repository -run TestGetDescendantIDsIntegration
   ```

### Run All Tests

To run both unit and integration tests:

```bash
# Unit tests
go test -v ./internal/repository

# Integration tests
RUN_INTEGRATION_TESTS=true go test -v -tags=integration ./internal/repository
```

## Test Coverage

To generate a coverage report:

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./internal/repository

# View coverage in browser
go tool cover -html=coverage.out
```

## Recursive CTE Implementation

The tests validate two critical recursive CTE queries:

### 1. GetDescendantIDs
Retrieves all descendant node IDs for a given node using PostgreSQL's `WITH RECURSIVE` clause.

**Use cases:**
- Cascade delete operations (get all items that will be deleted)
- Subtree operations (get all items in a subtree)
- Validation (check if a node has children before certain operations)

**SQL Query:**
```sql
WITH RECURSIVE descendants AS (
    SELECT id, parent_id FROM menu_items WHERE id = $1
    UNION ALL
    SELECT mi.id, mi.parent_id FROM menu_items mi
    INNER JOIN descendants d ON mi.parent_id = d.id
)
SELECT id FROM descendants WHERE id != $1
```

### 2. GetAncestors
Retrieves all ancestor nodes for a given node using PostgreSQL's `WITH RECURSIVE` clause.

**Use cases:**
- Circular reference detection (check if new parent is a descendant)
- Breadcrumb generation (get path from root to node)
- Permission checking (check if any ancestor has specific permissions)

**SQL Query:**
```sql
WITH RECURSIVE ancestors AS (
    SELECT id, name, parent_id, position, created_at, updated_at
    FROM menu_items WHERE id = $1
    UNION ALL
    SELECT mi.id, mi.name, mi.parent_id, mi.position, mi.created_at, mi.updated_at
    FROM menu_items mi
    INNER JOIN ancestors a ON mi.id = a.parent_id
)
SELECT id, name, parent_id, position, created_at, updated_at
FROM ancestors WHERE id != $1
ORDER BY id
```

## Requirements Coverage

These tests validate the following requirements:

- **Requirement 2.3**: Hierarchical data structure with parent-child relationships
- **Requirement 2.5**: CASCADE delete of descendant nodes when parent is deleted
- **Requirement 2.6**: Prevention of circular references in parent-child relationships
- **Requirement 4.3**: Validation that prevents moving a parent node to become its own descendant

## Error Handling

The tests verify proper error handling for:
- Database connection errors
- Query execution errors
- Row scanning errors
- Context cancellation
- Transaction rollback scenarios
- Non-existent node IDs

## Performance Considerations

The recursive CTE queries are optimized for PostgreSQL:
- Single query retrieves entire subtree/ancestor chain (no N+1 queries)
- Uses indexes on parent_id for efficient traversal
- Works efficiently with deep hierarchies (tested up to 10+ levels)
- Supports transactions for atomic operations
