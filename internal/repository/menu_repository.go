package repository

import (
	"context"

	"github.com/stk/menu-tree-api/internal/domain"
)

// MenuRepository defines the interface for menu item data access operations
type MenuRepository interface {
	// Basic CRUD operations
	FindAll(ctx context.Context) ([]domain.MenuItem, error)
	FindByID(ctx context.Context, id int) (*domain.MenuItem, error)
	FindByParentID(ctx context.Context, parentID *int) ([]domain.MenuItem, error)
	Create(ctx context.Context, item *domain.MenuItem) error
	Update(ctx context.Context, item *domain.MenuItem) error
	Delete(ctx context.Context, id int) (int, error)

	// Position management operations
	GetMaxPosition(ctx context.Context, parentID *int) (int, error)
	UpdatePositions(ctx context.Context, updates []domain.PositionUpdate) error

	// Hierarchical query operations (using PostgreSQL recursive CTEs)
	GetAncestors(ctx context.Context, id int) ([]domain.MenuItem, error)
	GetDescendantIDs(ctx context.Context, id int) ([]int, error)

	// Transaction support
	WithTransaction(ctx context.Context, fn func(repo MenuRepository) error) error
}
