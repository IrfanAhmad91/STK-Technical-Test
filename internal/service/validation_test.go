package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stk/menu-tree-api/internal/domain"
	"github.com/stk/menu-tree-api/internal/repository"
)

// Mock repository for testing validation functions
type mockValidationRepo struct {
	items     map[int]*domain.MenuItem
	ancestors map[int][]domain.MenuItem
	findError error
}

func (m *mockValidationRepo) FindByID(ctx context.Context, id int) (*domain.MenuItem, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	return m.items[id], nil
}

func (m *mockValidationRepo) GetAncestors(ctx context.Context, id int) ([]domain.MenuItem, error) {
	if m.findError != nil {
		return nil, m.findError
	}
	return m.ancestors[id], nil
}

// Implement remaining required interface methods (not used in these tests)
func (m *mockValidationRepo) FindAll(ctx context.Context) ([]domain.MenuItem, error) {
	return nil, nil
}
func (m *mockValidationRepo) FindByParentID(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
	return nil, nil
}
func (m *mockValidationRepo) Create(ctx context.Context, item *domain.MenuItem) error {
	return nil
}
func (m *mockValidationRepo) Update(ctx context.Context, item *domain.MenuItem) error {
	return nil
}
func (m *mockValidationRepo) Delete(ctx context.Context, id int) (int, error) {
	return 0, nil
}
func (m *mockValidationRepo) GetMaxPosition(ctx context.Context, parentID *int) (int, error) {
	return 0, nil
}
func (m *mockValidationRepo) UpdatePositions(ctx context.Context, updates []domain.PositionUpdate) error {
	return nil
}
func (m *mockValidationRepo) GetDescendantIDs(ctx context.Context, id int) ([]int, error) {
	return nil, nil
}
func (m *mockValidationRepo) WithTransaction(ctx context.Context, fn func(repository.MenuRepository) error) error {
	return nil
}

// TestValidateMenuName tests the validateMenuName function
func TestValidateMenuName(t *testing.T) {
	tests := []struct {
		name        string
		menuName    string
		expectError bool
	}{
		{
			name:        "Valid name",
			menuName:    "Menu Item",
			expectError: false,
		},
		{
			name:        "Empty name",
			menuName:    "",
			expectError: true,
		},
		{
			name:        "Name at max length (255)",
			menuName:    string(make([]byte, 255)),
			expectError: false,
		},
		{
			name:        "Name exceeds max length (256)",
			menuName:    string(make([]byte, 256)),
			expectError: true,
		},
		{
			name:        "Single character name",
			menuName:    "A",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMenuName(tt.menuName)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

// TestValidateParentExists tests the validateParentExists function
func TestValidateParentExists(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		parentID    *int
		repo        *mockValidationRepo
		expectError bool
	}{
		{
			name:        "Nil parent ID (root level)",
			parentID:    nil,
			repo:        &mockValidationRepo{},
			expectError: false,
		},
		{
			name:     "Parent exists",
			parentID: intPtr(1),
			repo: &mockValidationRepo{
				items: map[int]*domain.MenuItem{
					1: {ID: 1, Name: "Parent"},
				},
			},
			expectError: false,
		},
		{
			name:     "Parent does not exist",
			parentID: intPtr(999),
			repo: &mockValidationRepo{
				items: map[int]*domain.MenuItem{},
			},
			expectError: true,
		},
		{
			name:     "Database error",
			parentID: intPtr(1),
			repo: &mockValidationRepo{
				findError: errors.New("database error"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateParentExists(ctx, tt.repo, tt.parentID)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

// TestValidatePosition tests the validatePosition function
func TestValidatePosition(t *testing.T) {
	tests := []struct {
		name        string
		position    int
		expectError bool
	}{
		{
			name:        "Valid position (0)",
			position:    0,
			expectError: false,
		},
		{
			name:        "Valid position (positive)",
			position:    5,
			expectError: false,
		},
		{
			name:        "Invalid position (negative)",
			position:    -1,
			expectError: true,
		},
		{
			name:        "Invalid position (large negative)",
			position:    -100,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePosition(tt.position)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

// TestDetectCircularReference tests the detectCircularReference function
func TestDetectCircularReference(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		itemID      int
		newParentID *int
		repo        *mockValidationRepo
		expectError bool
	}{
		{
			name:        "Nil parent ID (moving to root)",
			itemID:      1,
			newParentID: nil,
			repo:        &mockValidationRepo{},
			expectError: false,
		},
		{
			name:        "Moving to itself",
			itemID:      1,
			newParentID: intPtr(1),
			repo:        &mockValidationRepo{},
			expectError: true,
		},
		{
			name:        "No circular reference",
			itemID:      3,
			newParentID: intPtr(1),
			repo: &mockValidationRepo{
				ancestors: map[int][]domain.MenuItem{
					1: {}, // Parent 1 has no ancestors
				},
			},
			expectError: false,
		},
		{
			name:        "Circular reference detected (item is ancestor of new parent)",
			itemID:      1,
			newParentID: intPtr(3),
			repo: &mockValidationRepo{
				ancestors: map[int][]domain.MenuItem{
					3: {
						{ID: 2, Name: "Item 2"},
						{ID: 1, Name: "Item 1"}, // Item 1 is ancestor of item 3
					},
				},
			},
			expectError: true,
		},
		{
			name:        "Database error",
			itemID:      1,
			newParentID: intPtr(2),
			repo: &mockValidationRepo{
				findError: errors.New("database error"),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := detectCircularReference(ctx, tt.repo, tt.itemID, tt.newParentID)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
		})
	}
}

// TestValidateMenuItemExists tests the validateMenuItemExists function
func TestValidateMenuItemExists(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		itemID      int
		repo        *mockValidationRepo
		expectError bool
		expectItem  bool
	}{
		{
			name:        "Invalid ID (zero)",
			itemID:      0,
			repo:        &mockValidationRepo{},
			expectError: true,
			expectItem:  false,
		},
		{
			name:        "Invalid ID (negative)",
			itemID:      -1,
			repo:        &mockValidationRepo{},
			expectError: true,
			expectItem:  false,
		},
		{
			name:   "Item exists",
			itemID: 1,
			repo: &mockValidationRepo{
				items: map[int]*domain.MenuItem{
					1: {ID: 1, Name: "Item 1"},
				},
			},
			expectError: false,
			expectItem:  true,
		},
		{
			name:   "Item does not exist",
			itemID: 999,
			repo: &mockValidationRepo{
				items: map[int]*domain.MenuItem{},
			},
			expectError: true,
			expectItem:  false,
		},
		{
			name:   "Database error",
			itemID: 1,
			repo: &mockValidationRepo{
				findError: errors.New("database error"),
			},
			expectError: true,
			expectItem:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, err := validateMenuItemExists(ctx, tt.repo, tt.itemID)
			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("expected no error but got: %v", err)
			}
			if tt.expectItem && item == nil {
				t.Errorf("expected item but got nil")
			}
			if !tt.expectItem && item != nil {
				t.Errorf("expected no item but got: %v", item)
			}
		})
	}
}
