package service

import (
	"context"
	"fmt"

	"github.com/stk/menu-tree-api/internal/domain"
	"github.com/stk/menu-tree-api/internal/repository"
)

// validateMenuName validates that menu name is between 1-255 characters
// Implements requirement 5.1: validate menu item name length
func validateMenuName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("menu item name cannot be empty")
	}
	if len(name) > 255 {
		return fmt.Errorf("menu item name cannot exceed 255 characters")
	}
	return nil
}

// validateParentExists checks if parent ID exists in database
// Implements requirement 5.2: validate parent ID reference exists
func validateParentExists(ctx context.Context, repo repository.MenuRepository, parentID *int) error {
	// If parent ID is nil (root level item), no validation needed
	if parentID == nil {
		return nil
	}

	// Check if parent exists
	parent, err := repo.FindByID(ctx, *parentID)
	if err != nil {
		return fmt.Errorf("failed to verify parent existence: %w", err)
	}
	if parent == nil {
		return fmt.Errorf("parent menu item with ID %d does not exist", *parentID)
	}

	return nil
}

// validatePosition validates that position is a non-negative integer
// Implements requirement 5.4: validate position values are non-negative integers
func validatePosition(position int) error {
	if position < 0 {
		return fmt.Errorf("position must be non-negative, got %d", position)
	}
	return nil
}

// detectCircularReference checks if moving item to new parent would create circular reference
// Uses GetAncestors to check if the item being moved is in the ancestor chain of new parent
// Implements requirement 2.6: prevent circular references in parent-child relationships
func detectCircularReference(ctx context.Context, repo repository.MenuRepository, itemID int, newParentID *int) error {
	// If moving to root level (newParentID is nil), no circular reference possible
	if newParentID == nil {
		return nil
	}

	// Cannot move item to itself
	if itemID == *newParentID {
		return fmt.Errorf("cannot move menu item to itself")
	}

	// Get all ancestors of the new parent
	ancestors, err := repo.GetAncestors(ctx, *newParentID)
	if err != nil {
		return fmt.Errorf("failed to check circular reference: %w", err)
	}

	// Check if the item being moved is in the ancestor chain
	for _, ancestor := range ancestors {
		if ancestor.ID == itemID {
			return fmt.Errorf("cannot move menu item: circular reference detected (new parent would become descendant of moved item)")
		}
	}

	return nil
}

// validateMenuItemExists checks if a menu item exists in database
// Implements requirement 5.3: validate menu item ID exists
func validateMenuItemExists(ctx context.Context, repo repository.MenuRepository, id int) (*domain.MenuItem, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid menu item ID: %d", id)
	}

	item, err := repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find menu item with ID %d: %w", id, err)
	}
	if item == nil {
		return nil, fmt.Errorf("menu item with ID %d does not exist", id)
	}

	return item, nil
}
