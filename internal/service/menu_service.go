package service

import (
	"context"
	"fmt"

	"github.com/stk/menu-tree-api/internal/domain"
	"github.com/stk/menu-tree-api/internal/repository"
)

// MenuService defines the interface for menu business logic operations
type MenuService interface {
	// CRUD operations
	GetAll(ctx context.Context) ([]domain.MenuItem, error)
	GetByID(ctx context.Context, id int) (*domain.MenuItem, error)
	Create(ctx context.Context, req domain.CreateMenuRequest) (*domain.MenuItem, error)
	Update(ctx context.Context, id int, req domain.UpdateMenuRequest) (*domain.MenuItem, error)
	Delete(ctx context.Context, id int) (int, error)

	// Tree operations
	Reorder(ctx context.Context, id int, newPosition int) ([]domain.MenuItem, error)
	Move(ctx context.Context, id int, newParentID *int, position *int) ([]domain.MenuItem, error)
	GetDescendants(ctx context.Context, id int) ([]domain.MenuItem, error)
	ValidateMove(ctx context.Context, id int, newParentID *int) error
}

// MenuServiceImpl implements the MenuService interface
type MenuServiceImpl struct {
	repo repository.MenuRepository
}

// NewMenuService creates a new instance of MenuServiceImpl
func NewMenuService(repo repository.MenuRepository) MenuService {
	return &MenuServiceImpl{
		repo: repo,
	}
}

// GetAll retrieves all menu items from the repository
// Implements requirement 1.3: retrieve all menu items with hierarchical relationships
func (s *MenuServiceImpl) GetAll(ctx context.Context) ([]domain.MenuItem, error) {
	items, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve menu items: %w", err)
	}

	// Return empty slice instead of nil for consistent JSON response
	if items == nil {
		return []domain.MenuItem{}, nil
	}

	return items, nil
}

// GetByID retrieves a specific menu item by its ID
// Implements requirement 1.4: retrieve specific menu item with parent and children references
func (s *MenuServiceImpl) GetByID(ctx context.Context, id int) (*domain.MenuItem, error) {
	item, err := validateMenuItemExists(ctx, s.repo, id)
	if err != nil {
		return nil, err
	}

	return item, nil
}

// Create creates a new menu item with validation and position calculation
// Implements requirements 1.1, 1.2, 5.1, 5.2
func (s *MenuServiceImpl) Create(ctx context.Context, req domain.CreateMenuRequest) (*domain.MenuItem, error) {
	// Validate name length (requirement 5.1)
	if err := validateMenuName(req.Name); err != nil {
		return nil, err
	}

	// Validate parent existence if parent ID is provided (requirement 5.2)
	if err := validateParentExists(ctx, s.repo, req.ParentID); err != nil {
		return nil, err
	}

	// Calculate position: if not provided, use GetMaxPosition + 1 (requirement 1.1)
	position := 0
	if req.Position != nil {
		// Validate that position is non-negative (requirement 5.4)
		if err := validatePosition(*req.Position); err != nil {
			return nil, err
		}
		position = *req.Position
	} else {
		// Auto-calculate position as max + 1
		maxPos, err := s.repo.GetMaxPosition(ctx, req.ParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate position: %w", err)
		}
		position = maxPos + 1
	}

	// Create new menu item
	newItem := &domain.MenuItem{
		Name:     req.Name,
		ParentID: req.ParentID,
		Position: position,
	}

	// Call repository Create method (requirement 1.1)
	err := s.repo.Create(ctx, newItem)
	if err != nil {
		return nil, fmt.Errorf("failed to create menu item: %w", err)
	}

	// Return created MenuItem (requirement 1.1)
	return newItem, nil
}

// Update updates an existing menu item
// Implements requirements 1.5, 5.1, 5.3
func (s *MenuServiceImpl) Update(ctx context.Context, id int, req domain.UpdateMenuRequest) (*domain.MenuItem, error) {
	// Validate name length (requirement 5.1)
	if err := validateMenuName(req.Name); err != nil {
		return nil, err
	}

	// Verify menu item exists (requirement 5.3)
	existingItem, err := validateMenuItemExists(ctx, s.repo, id)
	if err != nil {
		return nil, err
	}

	// Update the name field
	existingItem.Name = req.Name

	// Call repository Update method (requirement 1.5)
	err = s.repo.Update(ctx, existingItem)
	if err != nil {
		return nil, fmt.Errorf("failed to update menu item: %w", err)
	}

	// Return updated MenuItem (requirement 1.5)
	return existingItem, nil
}

// Delete deletes a menu item and all its descendants with cascade
// Implements requirements 1.6, 1.7, 2.5, 5.3
func (s *MenuServiceImpl) Delete(ctx context.Context, id int) (int, error) {
	// Verify menu item exists (requirement 5.3)
	existingItem, err := validateMenuItemExists(ctx, s.repo, id)
	if err != nil {
		return 0, err
	}

	// Store parent ID and position for sibling adjustment
	parentID := existingItem.ParentID
	deletedPosition := existingItem.Position

	// Delete the menu item (repository handles cascade and counting descendants)
	// Requirement 1.6: remove menu item and all child nodes
	// Requirement 2.5: delete all descendant child nodes via cascade
	deletedCount, err := s.repo.Delete(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("failed to delete menu item: %w", err)
	}

	// Adjust positions of remaining siblings (requirement 1.7)
	// Decrement positions > deleted position
	siblings, err := s.repo.FindByParentID(ctx, parentID)
	if err != nil {
		return deletedCount, fmt.Errorf("failed to retrieve siblings for position adjustment: %w", err)
	}

	// Build position updates for siblings that need adjustment
	var updates []domain.PositionUpdate
	for _, sibling := range siblings {
		if sibling.Position > deletedPosition {
			updates = append(updates, domain.PositionUpdate{
				ID:       sibling.ID,
				Position: sibling.Position - 1,
			})
		}
	}

	// Apply position updates if any siblings need adjustment
	if len(updates) > 0 {
		err = s.repo.UpdatePositions(ctx, updates)
		if err != nil {
			return deletedCount, fmt.Errorf("menu item deleted but failed to adjust sibling positions: %w", err)
		}
	}

	return deletedCount, nil
}

// Reorder changes the position of a menu item within its sibling group
// Implements requirements 3.1, 3.2, 3.3, 3.4, 3.5
func (s *MenuServiceImpl) Reorder(ctx context.Context, id int, newPosition int) ([]domain.MenuItem, error) {
	// Validate new position is non-negative (requirement 3.4)
	if err := validatePosition(newPosition); err != nil {
		return nil, err
	}

	// Get the menu item to reorder (requirement 3.1)
	item, err := validateMenuItemExists(ctx, s.repo, id)
	if err != nil {
		return nil, err
	}

	oldPosition := item.Position
	parentID := item.ParentID

	// If position unchanged, return early
	if oldPosition == newPosition {
		// Return the current sibling group
		siblings, err := s.repo.FindByParentID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve siblings: %w", err)
		}
		return siblings, nil
	}

	// Get all siblings to validate new position is within bounds (requirement 3.4)
	siblings, err := s.repo.FindByParentID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve siblings: %w", err)
	}

	// Validate new position is within valid range
	maxPosition := len(siblings) - 1
	if newPosition > maxPosition {
		return nil, fmt.Errorf("new position %d exceeds maximum valid position %d", newPosition, maxPosition)
	}

	// Calculate position updates (requirement 3.3)
	// If moving down: decrement positions in range [old+1, new]
	// If moving up: increment positions in range [new, old-1]
	var updates []domain.PositionUpdate

	if newPosition > oldPosition {
		// Moving down: shift items in range [oldPosition+1, newPosition] down by 1
		for _, sibling := range siblings {
			if sibling.Position > oldPosition && sibling.Position <= newPosition {
				updates = append(updates, domain.PositionUpdate{
					ID:       sibling.ID,
					Position: sibling.Position - 1,
				})
			}
		}
	} else {
		// Moving up: shift items in range [newPosition, oldPosition-1] up by 1
		for _, sibling := range siblings {
			if sibling.Position >= newPosition && sibling.Position < oldPosition {
				updates = append(updates, domain.PositionUpdate{
					ID:       sibling.ID,
					Position: sibling.Position + 1,
				})
			}
		}
	}

	// Add the moved item's new position
	updates = append(updates, domain.PositionUpdate{
		ID:       id,
		Position: newPosition,
	})

	// Execute position updates in a transaction (requirement 3.5)
	err = s.repo.UpdatePositions(ctx, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update positions: %w", err)
	}

	// Return affected menu items (requirement 3.5)
	affectedItems, err := s.repo.FindByParentID(ctx, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated siblings: %w", err)
	}

	return affectedItems, nil
}

// GetDescendants retrieves all descendants of a menu item
// Implements requirement 4.3: retrieve all descendant child nodes when moving a parent
func (s *MenuServiceImpl) GetDescendants(ctx context.Context, id int) ([]domain.MenuItem, error) {
	// Verify menu item exists
	_, err := validateMenuItemExists(ctx, s.repo, id)
	if err != nil {
		return nil, err
	}

	// Get descendant IDs using recursive CTE
	descendantIDs, err := s.repo.GetDescendantIDs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get descendant IDs: %w", err)
	}

	// If no descendants, return empty slice
	if len(descendantIDs) == 0 {
		return []domain.MenuItem{}, nil
	}

	// Retrieve full menu items for all descendants
	var descendants []domain.MenuItem
	for _, descendantID := range descendantIDs {
		descendant, err := s.repo.FindByID(ctx, descendantID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve descendant with ID %d: %w", descendantID, err)
		}
		if descendant != nil {
			descendants = append(descendants, *descendant)
		}
	}

	return descendants, nil
}

// ValidateMove validates whether a move operation is allowed
// Implements requirement 4.4: prevent moving a parent node to become a descendant of itself
func (s *MenuServiceImpl) ValidateMove(ctx context.Context, id int, newParentID *int) error {
	// Verify menu item exists (requirement 4.1)
	_, err := validateMenuItemExists(ctx, s.repo, id)
	if err != nil {
		return err
	}

	// If moving to root level (newParentID is nil), no validation needed
	if newParentID == nil {
		return nil
	}

	// Verify new parent exists (requirement 4.1)
	if err := validateParentExists(ctx, s.repo, newParentID); err != nil {
		return err
	}

	// Check for circular reference (requirement 4.4)
	if err := detectCircularReference(ctx, s.repo, id, newParentID); err != nil {
		return err
	}

	return nil
}

// Move moves a menu item to a different parent with position management
// Implements requirements 4.1, 4.2, 4.3, 4.4, 4.5
func (s *MenuServiceImpl) Move(ctx context.Context, id int, newParentID *int, position *int) ([]domain.MenuItem, error) {
	// Validate the move operation (requirement 4.1, 4.4)
	err := s.ValidateMove(ctx, id, newParentID)
	if err != nil {
		return nil, err
	}

	// Get the item being moved
	item, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve menu item: %w", err)
	}

	// Store source parent and position
	sourceParentID := item.ParentID
	sourcePosition := item.Position

	// Check if already in target location with same parent
	if (sourceParentID == nil && newParentID == nil) || 
	   (sourceParentID != nil && newParentID != nil && *sourceParentID == *newParentID) {
		// Already in the target parent, just reorder if position specified
		if position != nil && *position != sourcePosition {
			return s.Reorder(ctx, id, *position)
		}
		// No change needed, return current siblings
		siblings, err := s.repo.FindByParentID(ctx, sourceParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve siblings: %w", err)
		}
		return siblings, nil
	}

	// Calculate destination position
	destinationPosition := 0
	if position != nil {
		// Validate that position is non-negative
		if err := validatePosition(*position); err != nil {
			return nil, err
		}
		destinationPosition = *position
	} else {
		// Auto-calculate position as max + 1 at destination
		maxPos, err := s.repo.GetMaxPosition(ctx, newParentID)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate destination position: %w", err)
		}
		destinationPosition = maxPos + 1
	}

	// Execute move in transaction (requirement 4.2)
	var affectedItems []domain.MenuItem
	err = s.repo.WithTransaction(ctx, func(txRepo repository.MenuRepository) error {
		// Step 1: Remove from source - adjust positions of source siblings
		sourceSiblings, err := txRepo.FindByParentID(ctx, sourceParentID)
		if err != nil {
			return fmt.Errorf("failed to retrieve source siblings: %w", err)
		}

		var sourceUpdates []domain.PositionUpdate
		for _, sibling := range sourceSiblings {
			if sibling.ID != id && sibling.Position > sourcePosition {
				sourceUpdates = append(sourceUpdates, domain.PositionUpdate{
					ID:       sibling.ID,
					Position: sibling.Position - 1,
				})
			}
		}

		if len(sourceUpdates) > 0 {
			err = txRepo.UpdatePositions(ctx, sourceUpdates)
			if err != nil {
				return fmt.Errorf("failed to adjust source sibling positions: %w", err)
			}
		}

		// Step 2: Insert at destination - adjust positions of destination siblings
		destSiblings, err := txRepo.FindByParentID(ctx, newParentID)
		if err != nil {
			return fmt.Errorf("failed to retrieve destination siblings: %w", err)
		}

		// Validate destination position is within bounds
		maxDestPosition := len(destSiblings)
		if destinationPosition > maxDestPosition {
			destinationPosition = maxDestPosition
		}

		var destUpdates []domain.PositionUpdate
		for _, sibling := range destSiblings {
			if sibling.Position >= destinationPosition {
				destUpdates = append(destUpdates, domain.PositionUpdate{
					ID:       sibling.ID,
					Position: sibling.Position + 1,
				})
			}
		}

		if len(destUpdates) > 0 {
			err = txRepo.UpdatePositions(ctx, destUpdates)
			if err != nil {
				return fmt.Errorf("failed to adjust destination sibling positions: %w", err)
			}
		}

		// Step 3: Update the moved item's parent and position
		item.ParentID = newParentID
		item.Position = destinationPosition

		err = txRepo.Update(ctx, item)
		if err != nil {
			return fmt.Errorf("failed to update menu item: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("move operation failed: %w", err)
	}

	// Return affected menu items (requirement 4.5)
	// Get both source and destination siblings
	sourceSiblings, err := s.repo.FindByParentID(ctx, sourceParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated source siblings: %w", err)
	}

	destSiblings, err := s.repo.FindByParentID(ctx, newParentID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated destination siblings: %w", err)
	}

	// Combine affected items (avoiding duplicates if source == destination)
	affectedMap := make(map[int]domain.MenuItem)
	for _, sibling := range sourceSiblings {
		affectedMap[sibling.ID] = sibling
	}
	for _, sibling := range destSiblings {
		affectedMap[sibling.ID] = sibling
	}

	for _, item := range affectedMap {
		affectedItems = append(affectedItems, item)
	}

	return affectedItems, nil
}
