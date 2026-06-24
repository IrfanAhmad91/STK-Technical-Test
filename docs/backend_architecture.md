# Backend Architecture Documentation

## Overview

The backend follows a layered architecture pattern with clear separation of concerns:

1. **Entry Point Layer** (`cmd/api`) - Application initialization and routing
2. **Configuration Layer** (`config`) - Database and environment configuration
3. **Handler Layer** (`internal/handler`) - HTTP request/response handling [Task 3]
4. **Service Layer** (`internal/service`) - Business logic and validation [Task 3]
5. **Repository Layer** (`internal/repository`) - Data access abstraction
6. **Domain Layer** (`internal/domain`) - Data models and DTOs
7. **Database Layer** (PostgreSQL) - Persistent storage

## Layer Interaction Flow

```
HTTP Request
    ↓
[Gin Router] → Middleware (CORS, Logging, Recovery)
    ↓
[Handler Layer] → Parse request, validate input
    ↓
[Service Layer] → Business rules, validation, orchestration
    ↓
[Repository Layer] → Data access, SQL queries, transactions
    ↓
[PostgreSQL Database] → CRUD, recursive CTEs, triggers
    ↓
[Repository Layer] → Map rows to domain models
    ↓
[Service Layer] → Additional processing, tree operations
    ↓
[Handler Layer] → Format response, set status codes
    ↓
HTTP Response (JSON)
```

## Component Details

### 1. Entry Point Layer (`cmd/api/main.go`)

**Responsibilities:**
- Initialize database connection
- Create repository instances
- Setup Gin router with middleware
- Configure routes
- Start HTTP server

**Key Code:**
```go
func main() {
    // Load config
    dbConfig := config.LoadDatabaseConfig()
    
    // Connect to database
    db, err := config.NewDatabaseConnection(dbConfig)
    
    // Initialize layers
    repo := repository.NewPostgreSQLRepository(db)
    service := service.NewMenuService(repo)
    handler := handler.NewMenuHandler(service)
    
    // Setup router
    router := gin.Default()
    router.Use(cors.Default())
    
    // Register routes
    v1 := router.Group("/api/v1")
    v1.GET("/menus", handler.GetAll)
    // ... more routes
    
    // Start server
    router.Run(":8080")
}
```

### 2. Configuration Layer (`config/database.go`)

**Responsibilities:**
- Load environment variables
- Create database connection pool
- Configure connection pool parameters
- Validate database connectivity

**Connection Pool Settings:**
- `MaxOpenConns: 25` - Prevents connection exhaustion
- `MaxIdleConns: 5` - Balances reuse and resource usage
- `ConnMaxLifetime: 5min` - Prevents stale connections
- `ConnMaxIdleTime: 2min` - Releases unused connections

**Environment Variables:**
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=menu_tree_db
DB_SSLMODE=disable
```

### 3. Domain Layer (`internal/domain/menu_item.go`)

**Responsibilities:**
- Define data structures
- Provide validation rules
- Document API contracts (Swagger annotations)

**Models:**
```go
MenuItem              // Core entity
CreateMenuRequest     // POST payload
UpdateMenuRequest     // PUT payload
ReorderRequest        // Reorder operation
MoveRequest          // Move operation
ErrorResponse        // Error format
DeleteResponse       // Delete result
```

**Example:**
```go
type MenuItem struct {
    ID        int       `json:"id" db:"id"`
    Name      string    `json:"name" db:"name" binding:"required,min=1,max=255"`
    ParentID  *int      `json:"parent_id" db:"parent_id"`
    Position  int       `json:"position" db:"position"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
```

### 4. Repository Layer (`internal/repository/`)

**Interface:** `MenuRepository` - Defines data access contract

**Implementation:** `PostgreSQLRepository` - PostgreSQL-specific implementation

**Methods:**

**Basic CRUD:**
- `FindAll(ctx)` - SELECT all items
- `FindByID(ctx, id)` - SELECT by primary key
- `FindByParentID(ctx, parentID)` - SELECT children
- `Create(ctx, item)` - INSERT with RETURNING
- `Update(ctx, item)` - UPDATE with RETURNING
- `Delete(ctx, id)` - DELETE with CASCADE

**Position Management:**
- `GetMaxPosition(ctx, parentID)` - SELECT MAX(position)
- `UpdatePositions(ctx, updates)` - Batch UPDATE

**Hierarchical Queries:**
- `GetAncestors(ctx, id)` - WITH RECURSIVE (upward)
- `GetDescendantIDs(ctx, id)` - WITH RECURSIVE (downward)

**Transaction Support:**
- `WithTransaction(ctx, fn)` - BEGIN/COMMIT/ROLLBACK

**Key Features:**
- Context propagation for cancellation
- Error wrapping with fmt.Errorf
- Resource cleanup with defer
- NULL handling for root items
- Transaction isolation

### 5. Service Layer (`internal/service/`) [Task 3]

**Responsibilities:**
- Business rule enforcement
- Input validation
- Tree operation orchestration
- Circular reference prevention
- Position calculation
- Transaction coordination

**Planned Methods:**
```go
GetAll(ctx) []MenuItem
GetByID(ctx, id) *MenuItem
Create(ctx, req CreateMenuRequest) *MenuItem
Update(ctx, id, req UpdateMenuRequest) *MenuItem
Delete(ctx, id) int
Reorder(ctx, id, newPosition) []MenuItem
Move(ctx, id, newParentID, position) []MenuItem
GetDescendants(ctx, id) []MenuItem
ValidateMove(ctx, id, newParentID) error
```

**Validation Logic:**
- Name length (1-255 characters)
- Parent existence check
- Circular reference prevention
- Position bounds checking

### 6. Handler Layer (`internal/handler/`) [Task 3]

**Responsibilities:**
- Parse HTTP requests
- Validate request format
- Call service methods
- Format HTTP responses
- Set status codes
- Handle errors

**Planned Handlers:**
```go
GetAll(c *gin.Context)           // GET /menus
GetByID(c *gin.Context)          // GET /menus/:id
Create(c *gin.Context)           // POST /menus
Update(c *gin.Context)           // PUT /menus/:id
Delete(c *gin.Context)           // DELETE /menus/:id
Reorder(c *gin.Context)          // PUT /menus/:id/reorder
Move(c *gin.Context)             // PUT /menus/:id/move
```

**Error Mapping:**
- Validation error → 400 Bad Request
- Not found → 404 Not Found
- Circular reference → 409 Conflict
- Database error → 500 Internal Server Error

## Data Flow Examples

### Example 1: Create Menu Item

```
1. Client: POST /api/v1/menus
   Body: {"name": "Settings", "parent_id": 1}

2. Handler.Create()
   - Parse JSON to CreateMenuRequest
   - Validate request body
   
3. Service.Create()
   - Validate parent exists
   - Get max position for parent
   - Set position = maxPosition + 1
   
4. Repository.Create()
   - INSERT INTO menu_items
   - RETURNING id, created_at, updated_at
   
5. Response: 201 Created
   Body: {"id": 5, "name": "Settings", "parent_id": 1, "position": 3, ...}
```

### Example 2: Move Menu Item

```
1. Client: PUT /api/v1/menus/5/move
   Body: {"new_parent_id": 2, "position": 0}

2. Handler.Move()
   - Parse path parameter (id=5)
   - Parse JSON to MoveRequest
   
3. Service.Move()
   - Validate item exists
   - Validate new parent exists
   - Check for circular reference (ancestors query)
   - Begin transaction
     - Update positions in source parent (shift down)
     - Update positions in destination parent (shift right)
     - Update item's parent_id and position
   - Commit transaction
   
4. Repository operations (in transaction)
   - GetAncestors(2) → Check if 5 is an ancestor
   - FindByParentID(5) → Get siblings at source
   - UpdatePositions() → Adjust source positions
   - FindByParentID(2) → Get siblings at destination
   - UpdatePositions() → Adjust destination positions
   - Update(item) → Change parent_id and position
   
5. Response: 200 OK
   Body: [affected menu items with new positions]
```

### Example 3: Delete Menu Item

```
1. Client: DELETE /api/v1/menus/3

2. Handler.Delete()
   - Parse path parameter (id=3)
   
3. Service.Delete()
   - Validate item exists
   - Begin transaction
     - Count descendants (recursive CTE)
     - Delete item (CASCADE removes descendants)
     - Get siblings with position > deleted item's position
     - Update sibling positions (shift down)
   - Commit transaction
   
4. Repository operations (in transaction)
   - GetDescendantIDs(3) → Count items to delete
   - Delete(3) → DELETE FROM menu_items WHERE id=3
     (CASCADE automatically deletes descendants)
   - FindByParentID(parent) → Get remaining siblings
   - UpdatePositions() → Adjust positions
   
5. Response: 200 OK
   Body: {"deleted_count": 4, "message": "Menu item and 3 descendants deleted"}
```

## Database Query Examples

### Recursive CTE for Ancestors

```sql
WITH RECURSIVE ancestors AS (
    -- Base case
    SELECT id, parent_id FROM menu_items WHERE id = 5
    
    UNION ALL
    
    -- Recursive case
    SELECT mi.id, mi.parent_id
    FROM menu_items mi
    JOIN ancestors a ON mi.id = a.parent_id
)
SELECT id FROM ancestors WHERE id != 5;
```

**Use case:** Detect if moving item 5 under item 10 would create a cycle

### Recursive CTE for Descendants

```sql
WITH RECURSIVE descendants AS (
    -- Base case
    SELECT id, parent_id FROM menu_items WHERE id = 3
    
    UNION ALL
    
    -- Recursive case
    SELECT mi.id, mi.parent_id
    FROM menu_items mi
    JOIN descendants d ON mi.parent_id = d.id
)
SELECT COUNT(*) FROM descendants;
```

**Use case:** Count how many items will be deleted (including descendants)

### Batch Position Update

```sql
-- Prepared statement executed multiple times
UPDATE menu_items
SET position = $1, updated_at = CURRENT_TIMESTAMP
WHERE id = $2;
```

**Use case:** Efficiently update positions after reorder/move

## Transaction Patterns

### Pattern 1: Atomic Create with Position

```go
err := repo.WithTransaction(ctx, func(txRepo MenuRepository) error {
    // Get max position within transaction
    maxPos, err := txRepo.GetMaxPosition(ctx, parentID)
    if err != nil {
        return err
    }
    
    // Create with next position
    item.Position = maxPos + 1
    return txRepo.Create(ctx, item)
})
```

### Pattern 2: Atomic Reorder

```go
err := repo.WithTransaction(ctx, func(txRepo MenuRepository) error {
    // Get item
    item, err := txRepo.FindByID(ctx, id)
    if err != nil {
        return err
    }
    
    // Get siblings
    siblings, err := txRepo.FindByParentID(ctx, item.ParentID)
    if err != nil {
        return err
    }
    
    // Calculate position updates
    updates := calculatePositionUpdates(siblings, oldPos, newPos)
    
    // Apply updates
    return txRepo.UpdatePositions(ctx, updates)
})
```

### Pattern 3: Atomic Move

```go
err := repo.WithTransaction(ctx, func(txRepo MenuRepository) error {
    // Validate no circular reference
    ancestors, err := txRepo.GetAncestors(ctx, newParentID)
    if err != nil {
        return err
    }
    for _, ancestor := range ancestors {
        if ancestor.ID == itemID {
            return errors.New("circular reference")
        }
    }
    
    // Update source positions
    if err := updateSourcePositions(txRepo, ctx, item); err != nil {
        return err
    }
    
    // Update destination positions
    if err := updateDestinationPositions(txRepo, ctx, newParentID); err != nil {
        return err
    }
    
    // Move item
    item.ParentID = newParentID
    item.Position = newPosition
    return txRepo.Update(ctx, item)
})
```

## Error Handling Strategy

### Layer-Specific Error Handling

**Repository Layer:**
- Wrap database errors with context
- Return sql.ErrNoRows as domain errors
- Log query failures

**Service Layer:**
- Validate business rules
- Return domain-specific errors
- Handle repository errors

**Handler Layer:**
- Map errors to HTTP status codes
- Format error responses
- Log request failures

### Error Response Format

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Invalid request parameters",
  "details": {
    "field": "parent_id",
    "issue": "Parent menu item does not exist"
  }
}
```

### HTTP Status Code Mapping

| Error Type | Status Code | Example |
|------------|-------------|---------|
| Validation error | 400 | Invalid name length |
| Not found | 404 | Menu item doesn't exist |
| Circular reference | 409 | Moving item under its descendant |
| Database error | 500 | Connection failure |

## Performance Considerations

### Query Optimization

1. **Indexes Used:**
   - `idx_parent_id` - Single column for foreign key
   - `idx_position` - Single column for ordering
   - `idx_parent_position` - Composite for siblings query

2. **Query Patterns:**
   - Prepared statements for repeated queries
   - Batch updates for position changes
   - Single recursive query vs N+1 queries

3. **Connection Pooling:**
   - Reuse connections (avoid open/close overhead)
   - Limit concurrent connections (prevent exhaustion)
   - Close stale connections (free resources)

### Caching Opportunities (Future)

- Menu tree structure (invalidate on write)
- Max position per parent (invalidate on insert)
- Ancestor paths (invalidate on move)

## Testing Strategy (Task 3)

### Unit Tests

**Repository Tests:**
- Mock database with sqlmock
- Test each query independently
- Test transaction rollback

**Service Tests:**
- Mock repository interface
- Test business logic
- Test validation rules

**Handler Tests:**
- Mock service interface
- Test request parsing
- Test response formatting

### Integration Tests

- Real PostgreSQL database (test container)
- End-to-end API tests
- Transaction isolation tests
- Concurrent operation tests

## Security Considerations

### SQL Injection Prevention

- Parameterized queries (no string concatenation)
- Prepared statements
- Input validation

### Input Validation

- Name length limits (1-255 characters)
- Position non-negative
- Parent ID existence check

### Production Hardening

- SSL/TLS for database connection
- Rate limiting middleware
- Request size limits
- CORS configuration
- Authentication/authorization (future)

## Monitoring and Observability

### Logging

- Request logging (middleware)
- Error logging (all layers)
- Query logging (development)

### Metrics (Future)

- Request count by endpoint
- Response time percentiles
- Database connection pool usage
- Error rate by type

### Health Checks

- Database connectivity
- API responsiveness
- Dependency status

## Deployment

### Build Process

```bash
# Development build
go build -o bin/menu-api cmd/api/main.go

# Production build (optimized)
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/menu-api cmd/api/main.go
```

### Configuration

- Environment variables for config
- No hardcoded credentials
- SSL/TLS in production

### Database Migrations

- Apply migrations before deployment
- Version control migration files
- Rollback capability

---

**Status:** Task 2 Completed ✅
**Next:** Task 3 - Service Layer & Handlers
