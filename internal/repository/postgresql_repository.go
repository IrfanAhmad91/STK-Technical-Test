package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/stk/menu-tree-api/internal/domain"
)

// PostgreSQLRepository implements MenuRepository using PostgreSQL
type PostgreSQLRepository struct {
	db *sql.DB
	tx *sql.Tx // Transaction handle, nil if not in transaction
}

// NewPostgreSQLRepository creates a new PostgreSQL repository
func NewPostgreSQLRepository(db *sql.DB) *PostgreSQLRepository {
	return &PostgreSQLRepository{db: db}
}

// getExecutor returns the appropriate executor (tx if in transaction, otherwise db)
func (r *PostgreSQLRepository) getExecutor() interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
} {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// FindAll retrieves all menu items from the database
func (r *PostgreSQLRepository) FindAll(ctx context.Context) ([]domain.MenuItem, error) {
	query := `
		SELECT id, name, parent_id, position, created_at, updated_at
		FROM menu_items
		ORDER BY parent_id NULLS FIRST, position
	`

	rows, err := r.getExecutor().QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu items: %w", err)
	}
	defer rows.Close()

	var items []domain.MenuItem
	for rows.Next() {
		var item domain.MenuItem
		err := rows.Scan(&item.ID, &item.Name, &item.ParentID, &item.Position, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating menu items: %w", err)
	}

	return items, nil
}

// FindByID retrieves a specific menu item by its ID
func (r *PostgreSQLRepository) FindByID(ctx context.Context, id int) (*domain.MenuItem, error) {
	query := `
		SELECT id, name, parent_id, position, created_at, updated_at
		FROM menu_items
		WHERE id = $1
	`

	var item domain.MenuItem
	err := r.getExecutor().QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.Name, &item.ParentID, &item.Position, &item.CreatedAt, &item.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("menu item with id %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query menu item: %w", err)
	}

	return &item, nil
}

// FindByParentID retrieves all menu items with the specified parent ID
func (r *PostgreSQLRepository) FindByParentID(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
	var query string
	var args []interface{}

	if parentID == nil {
		query = `
			SELECT id, name, parent_id, position, created_at, updated_at
			FROM menu_items
			WHERE parent_id IS NULL
			ORDER BY position
		`
	} else {
		query = `
			SELECT id, name, parent_id, position, created_at, updated_at
			FROM menu_items
			WHERE parent_id = $1
			ORDER BY position
		`
		args = append(args, *parentID)
	}

	rows, err := r.getExecutor().QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu items by parent: %w", err)
	}
	defer rows.Close()

	var items []domain.MenuItem
	for rows.Next() {
		var item domain.MenuItem
		err := rows.Scan(&item.ID, &item.Name, &item.ParentID, &item.Position, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating menu items: %w", err)
	}

	return items, nil
}

// Create inserts a new menu item into the database
func (r *PostgreSQLRepository) Create(ctx context.Context, item *domain.MenuItem) error {
	query := `
		INSERT INTO menu_items (name, parent_id, position)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.getExecutor().QueryRowContext(ctx, query, item.Name, item.ParentID, item.Position).Scan(
		&item.ID, &item.CreatedAt, &item.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create menu item: %w", err)
	}

	return nil
}

// Update modifies an existing menu item in the database
func (r *PostgreSQLRepository) Update(ctx context.Context, item *domain.MenuItem) error {
	query := `
		UPDATE menu_items
		SET name = $1, parent_id = $2, position = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING updated_at
	`

	err := r.getExecutor().QueryRowContext(ctx, query, item.Name, item.ParentID, item.Position, item.ID).Scan(&item.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("menu item with id %d not found", item.ID)
	}
	if err != nil {
		return fmt.Errorf("failed to update menu item: %w", err)
	}

	return nil
}

// Delete removes a menu item and all its descendants (CASCADE)
func (r *PostgreSQLRepository) Delete(ctx context.Context, id int) (int, error) {
	// First, count the number of items that will be deleted (item + descendants)
	countQuery := `
		WITH RECURSIVE descendants AS (
			SELECT id FROM menu_items WHERE id = $1
			UNION ALL
			SELECT mi.id FROM menu_items mi
			INNER JOIN descendants d ON mi.parent_id = d.id
		)
		SELECT COUNT(*) FROM descendants
	`

	var count int
	err := r.getExecutor().QueryRowContext(ctx, countQuery, id).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count descendants: %w", err)
	}

	// Delete the menu item (CASCADE will delete descendants)
	deleteQuery := `DELETE FROM menu_items WHERE id = $1`
	result, err := r.getExecutor().ExecContext(ctx, deleteQuery, id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete menu item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return 0, fmt.Errorf("menu item with id %d not found", id)
	}

	return count, nil
}

// GetMaxPosition returns the maximum position value for items with the specified parent
func (r *PostgreSQLRepository) GetMaxPosition(ctx context.Context, parentID *int) (int, error) {
	var query string
	var args []interface{}

	if parentID == nil {
		query = `
			SELECT COALESCE(MAX(position), -1)
			FROM menu_items
			WHERE parent_id IS NULL
		`
	} else {
		query = `
			SELECT COALESCE(MAX(position), -1)
			FROM menu_items
			WHERE parent_id = $1
		`
		args = append(args, *parentID)
	}

	var maxPosition int
	err := r.getExecutor().QueryRowContext(ctx, query, args...).Scan(&maxPosition)
	if err != nil {
		return -1, fmt.Errorf("failed to get max position: %w", err)
	}

	return maxPosition, nil
}

// UpdatePositions performs batch position updates
func (r *PostgreSQLRepository) UpdatePositions(ctx context.Context, updates []domain.PositionUpdate) error {
	if len(updates) == 0 {
		return nil
	}

	// Use a prepared statement for efficiency
	stmt, err := r.getExecutor().(interface {
		PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	}).PrepareContext(ctx, `
		UPDATE menu_items
		SET position = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare update statement: %w", err)
	}
	defer stmt.Close()

	for _, update := range updates {
		_, err := stmt.ExecContext(ctx, update.Position, update.ID)
		if err != nil {
			return fmt.Errorf("failed to update position for item %d: %w", update.ID, err)
		}
	}

	return nil
}

// GetAncestors retrieves all ancestor menu items using recursive CTE
func (r *PostgreSQLRepository) GetAncestors(ctx context.Context, id int) ([]domain.MenuItem, error) {
	query := `
		WITH RECURSIVE ancestors AS (
			-- Base case: the node itself
			SELECT id, name, parent_id, position, created_at, updated_at
			FROM menu_items
			WHERE id = $1
			
			UNION ALL
			
			-- Recursive case: parent of ancestors
			SELECT mi.id, mi.name, mi.parent_id, mi.position, mi.created_at, mi.updated_at
			FROM menu_items mi
			INNER JOIN ancestors a ON mi.id = a.parent_id
		)
		SELECT id, name, parent_id, position, created_at, updated_at
		FROM ancestors
		WHERE id != $1
		ORDER BY id
	`

	rows, err := r.getExecutor().QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get ancestors: %w", err)
	}
	defer rows.Close()

	var items []domain.MenuItem
	for rows.Next() {
		var item domain.MenuItem
		err := rows.Scan(&item.ID, &item.Name, &item.ParentID, &item.Position, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ancestor: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating ancestors: %w", err)
	}

	return items, nil
}

// GetDescendantIDs retrieves IDs of all descendant menu items using recursive CTE
func (r *PostgreSQLRepository) GetDescendantIDs(ctx context.Context, id int) ([]int, error) {
	query := `
		WITH RECURSIVE descendants AS (
			-- Base case: the node itself
			SELECT id, parent_id
			FROM menu_items
			WHERE id = $1
			
			UNION ALL
			
			-- Recursive case: children of descendants
			SELECT mi.id, mi.parent_id
			FROM menu_items mi
			INNER JOIN descendants d ON mi.parent_id = d.id
		)
		SELECT id FROM descendants WHERE id != $1
	`

	rows, err := r.getExecutor().QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendants: %w", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan descendant id: %w", err)
		}
		ids = append(ids, id)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating descendants: %w", err)
	}

	return ids, nil
}

// WithTransaction executes a function within a database transaction
func (r *PostgreSQLRepository) WithTransaction(ctx context.Context, fn func(repo MenuRepository) error) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Create a new repository instance with the transaction
	txRepo := &PostgreSQLRepository{
		db: r.db,
		tx: tx,
	}

	// Execute the function
	err = fn(txRepo)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction (original error: %w): %v", err, rbErr)
		}
		return err
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
