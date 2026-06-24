package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stk/menu-tree-api/internal/domain"
	"github.com/stk/menu-tree-api/internal/repository"
)

// MockMenuRepository is a mock implementation of MenuRepository for testing
type MockMenuRepository struct {
	// Functions to override behavior in tests
	FindAllFunc          func(ctx context.Context) ([]domain.MenuItem, error)
	FindByIDFunc         func(ctx context.Context, id int) (*domain.MenuItem, error)
	FindByParentIDFunc   func(ctx context.Context, parentID *int) ([]domain.MenuItem, error)
	CreateFunc           func(ctx context.Context, item *domain.MenuItem) error
	UpdateFunc           func(ctx context.Context, item *domain.MenuItem) error
	DeleteFunc           func(ctx context.Context, id int) (int, error)
	GetMaxPositionFunc   func(ctx context.Context, parentID *int) (int, error)
	UpdatePositionsFunc  func(ctx context.Context, updates []domain.PositionUpdate) error
	GetAncestorsFunc     func(ctx context.Context, id int) ([]domain.MenuItem, error)
	GetDescendantIDsFunc func(ctx context.Context, id int) ([]int, error)
	WithTransactionFunc  func(ctx context.Context, fn func(repo interface{}) error) error
}

func (m *MockMenuRepository) FindAll(ctx context.Context) ([]domain.MenuItem, error) {
	if m.FindAllFunc != nil {
		return m.FindAllFunc(ctx)
	}
	return []domain.MenuItem{}, nil
}

func (m *MockMenuRepository) FindByID(ctx context.Context, id int) (*domain.MenuItem, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *MockMenuRepository) FindByParentID(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
	if m.FindByParentIDFunc != nil {
		return m.FindByParentIDFunc(ctx, parentID)
	}
	return []domain.MenuItem{}, nil
}

func (m *MockMenuRepository) Create(ctx context.Context, item *domain.MenuItem) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, item)
	}
	return nil
}

func (m *MockMenuRepository) Update(ctx context.Context, item *domain.MenuItem) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, item)
	}
	return nil
}

func (m *MockMenuRepository) Delete(ctx context.Context, id int) (int, error) {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return 0, nil
}

func (m *MockMenuRepository) GetMaxPosition(ctx context.Context, parentID *int) (int, error) {
	if m.GetMaxPositionFunc != nil {
		return m.GetMaxPositionFunc(ctx, parentID)
	}
	return 0, nil
}

func (m *MockMenuRepository) UpdatePositions(ctx context.Context, updates []domain.PositionUpdate) error {
	if m.UpdatePositionsFunc != nil {
		return m.UpdatePositionsFunc(ctx, updates)
	}
	return nil
}

func (m *MockMenuRepository) GetAncestors(ctx context.Context, id int) ([]domain.MenuItem, error) {
	if m.GetAncestorsFunc != nil {
		return m.GetAncestorsFunc(ctx, id)
	}
	return []domain.MenuItem{}, nil
}

func (m *MockMenuRepository) GetDescendantIDs(ctx context.Context, id int) ([]int, error) {
	if m.GetDescendantIDsFunc != nil {
		return m.GetDescendantIDsFunc(ctx, id)
	}
	return []int{}, nil
}

func (m *MockMenuRepository) WithTransaction(ctx context.Context, fn func(repo repository.MenuRepository) error) error {
	if m.WithTransactionFunc != nil {
		// Convert the function signature for compatibility
		return m.WithTransactionFunc(ctx, func(r interface{}) error {
			return fn(r.(repository.MenuRepository))
		})
	}
	return fn(m)
}

// Test GetAll method
func TestGetAll_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	expectedItems := []domain.MenuItem{
		{ID: 1, Name: "Dashboard", ParentID: nil, Position: 0, CreatedAt: now, UpdatedAt: now},
		{ID: 2, Name: "Settings", ParentID: nil, Position: 1, CreatedAt: now, UpdatedAt: now},
	}

	mockRepo := &MockMenuRepository{
		FindAllFunc: func(ctx context.Context) ([]domain.MenuItem, error) {
			return expectedItems, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	items, err := service.GetAll(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(items) != len(expectedItems) {
		t.Errorf("Expected %d items, got %d", len(expectedItems), len(items))
	}

	for i, item := range items {
		if item.ID != expectedItems[i].ID {
			t.Errorf("Expected item ID %d at index %d, got %d", expectedItems[i].ID, i, item.ID)
		}
		if item.Name != expectedItems[i].Name {
			t.Errorf("Expected item name %s at index %d, got %s", expectedItems[i].Name, i, item.Name)
		}
	}
}

func TestGetAll_EmptyResult(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{
		FindAllFunc: func(ctx context.Context) ([]domain.MenuItem, error) {
			return nil, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	items, err := service.GetAll(ctx)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if items == nil {
		t.Error("Expected empty slice, got nil")
	}

	if len(items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(items))
	}
}

func TestGetAll_RepositoryError(t *testing.T) {
	// Arrange
	expectedError := errors.New("database connection failed")
	mockRepo := &MockMenuRepository{
		FindAllFunc: func(ctx context.Context) ([]domain.MenuItem, error) {
			return nil, expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	items, err := service.GetAll(ctx)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if items != nil {
		t.Errorf("Expected nil items on error, got %v", items)
	}
}

// Test GetByID method
func TestGetByID_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	expectedItem := &domain.MenuItem{
		ID:        1,
		Name:      "Dashboard",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 1 {
				return expectedItem, nil
			}
			return nil, errors.New("not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	item, err := service.GetByID(ctx, 1)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if item == nil {
		t.Fatal("Expected item, got nil")
	}

	if item.ID != expectedItem.ID {
		t.Errorf("Expected ID %d, got %d", expectedItem.ID, item.ID)
	}

	if item.Name != expectedItem.Name {
		t.Errorf("Expected name %s, got %s", expectedItem.Name, item.Name)
	}
}

func TestGetByID_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Test cases
	invalidIDs := []int{0, -1, -100}

	for _, id := range invalidIDs {
		// Act
		item, err := service.GetByID(ctx, id)

		// Assert
		if err == nil {
			t.Errorf("Expected error for invalid ID %d, got nil", id)
		}

		if item != nil {
			t.Errorf("Expected nil item for invalid ID %d, got %v", id, item)
		}
	}
}

func TestGetByID_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			return nil, errors.New("menu item with id 999 not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	item, err := service.GetByID(ctx, 999)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent ID, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item for non-existent ID, got %v", item)
	}
}

func TestGetByID_RepositoryError(t *testing.T) {
	// Arrange
	expectedError := errors.New("database connection failed")
	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			return nil, expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	item, err := service.GetByID(ctx, 1)

	// Assert
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}
}

// Test Create method
func TestCreate_Success_WithAutoPosition(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{
		GetMaxPositionFunc: func(ctx context.Context, parentID *int) (int, error) {
			return 2, nil // Existing items have positions 0, 1, 2
		},
		CreateFunc: func(ctx context.Context, item *domain.MenuItem) error {
			// Simulate database auto-generating ID
			item.ID = 123
			item.CreatedAt = time.Now()
			item.UpdatedAt = time.Now()
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.CreateMenuRequest{
		Name:     "New Menu Item",
		ParentID: nil,
		Position: nil, // Auto-calculate position
	}

	// Act
	item, err := service.Create(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if item == nil {
		t.Fatal("Expected item, got nil")
	}

	if item.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, item.Name)
	}

	if item.Position != 3 {
		t.Errorf("Expected position 3 (max 2 + 1), got %d", item.Position)
	}

	if item.ID != 123 {
		t.Errorf("Expected ID 123, got %d", item.ID)
	}
}

func TestCreate_Success_WithSpecifiedPosition(t *testing.T) {
	// Arrange
	parentID := 5
	position := 10

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == parentID {
				return &domain.MenuItem{ID: parentID, Name: "Parent"}, nil
			}
			return nil, errors.New("not found")
		},
		CreateFunc: func(ctx context.Context, item *domain.MenuItem) error {
			item.ID = 456
			item.CreatedAt = time.Now()
			item.UpdatedAt = time.Now()
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.CreateMenuRequest{
		Name:     "Child Item",
		ParentID: &parentID,
		Position: &position,
	}

	// Act
	item, err := service.Create(ctx, req)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if item == nil {
		t.Fatal("Expected item, got nil")
	}

	if item.Position != position {
		t.Errorf("Expected position %d, got %d", position, item.Position)
	}

	if *item.ParentID != parentID {
		t.Errorf("Expected parent ID %d, got %d", parentID, *item.ParentID)
	}
}

func TestCreate_EmptyName(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.CreateMenuRequest{
		Name:     "",
		ParentID: nil,
		Position: nil,
	}

	// Act
	item, err := service.Create(ctx, req)

	// Assert
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}

	if err.Error() != "menu item name cannot be empty" {
		t.Errorf("Expected 'menu item name cannot be empty', got: %s", err.Error())
	}
}

func TestCreate_NameTooLong(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Create a name longer than 255 characters
	longName := string(make([]byte, 256))
	for i := range longName {
		longName = longName[:i] + "a" + longName[i+1:]
	}

	req := domain.CreateMenuRequest{
		Name:     longName,
		ParentID: nil,
		Position: nil,
	}

	// Act
	item, err := service.Create(ctx, req)

	// Assert
	if err == nil {
		t.Error("Expected error for name too long, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}

	if err.Error() != "menu item name cannot exceed 255 characters" {
		t.Errorf("Expected 'menu item name cannot exceed 255 characters', got: %s", err.Error())
	}
}

func TestCreate_ParentNotFound(t *testing.T) {
	// Arrange
	parentID := 999

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			return nil, errors.New("menu item with id 999 not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.CreateMenuRequest{
		Name:     "Orphan Item",
		ParentID: &parentID,
		Position: nil,
	}

	// Act
	item, err := service.Create(ctx, req)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent parent, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}
}

func TestCreate_NegativePosition(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	position := -5
	req := domain.CreateMenuRequest{
		Name:     "Invalid Position Item",
		ParentID: nil,
		Position: &position,
	}

	// Act
	item, err := service.Create(ctx, req)

	// Assert
	if err == nil {
		t.Error("Expected error for negative position, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}
}

func TestCreate_RepositoryCreateError(t *testing.T) {
	// Arrange
	expectedError := errors.New("database insert failed")

	mockRepo := &MockMenuRepository{
		GetMaxPositionFunc: func(ctx context.Context, parentID *int) (int, error) {
			return 0, nil
		},
		CreateFunc: func(ctx context.Context, item *domain.MenuItem) error {
			return expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.CreateMenuRequest{
		Name:     "Test Item",
		ParentID: nil,
		Position: nil,
	}

	// Act
	item, err := service.Create(ctx, req)

	// Assert
	if err == nil {
		t.Error("Expected error from repository, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}
}

func TestCreate_GetMaxPositionError(t *testing.T) {
	// Arrange
	expectedError := errors.New("failed to query max position")

	mockRepo := &MockMenuRepository{
		GetMaxPositionFunc: func(ctx context.Context, parentID *int) (int, error) {
			return 0, expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.CreateMenuRequest{
		Name:     "Test Item",
		ParentID: nil,
		Position: nil,
	}

	// Act
	item, err := service.Create(ctx, req)

	// Assert
	if err == nil {
		t.Error("Expected error from GetMaxPosition, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}
}

// Test Update method
func TestUpdate_Success(t *testing.T) {
	// Arrange
	now := time.Now()
	existingItem := &domain.MenuItem{
		ID:        1,
		Name:      "Old Name",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 1 {
				return existingItem, nil
			}
			return nil, errors.New("not found")
		},
		UpdateFunc: func(ctx context.Context, item *domain.MenuItem) error {
			// Simulate database update
			item.UpdatedAt = time.Now()
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.UpdateMenuRequest{
		Name: "New Name",
	}

	// Act
	item, err := service.Update(ctx, 1, req)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if item == nil {
		t.Fatal("Expected item, got nil")
	}

	if item.Name != req.Name {
		t.Errorf("Expected name '%s', got '%s'", req.Name, item.Name)
	}

	if item.ID != 1 {
		t.Errorf("Expected ID 1, got %d", item.ID)
	}
}

func TestUpdate_EmptyName(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.UpdateMenuRequest{
		Name: "",
	}

	// Act
	item, err := service.Update(ctx, 1, req)

	// Assert
	if err == nil {
		t.Error("Expected error for empty name, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}

	if err.Error() != "menu item name cannot be empty" {
		t.Errorf("Expected 'menu item name cannot be empty', got: %s", err.Error())
	}
}

func TestUpdate_NameTooLong(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Create a name longer than 255 characters
	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}

	req := domain.UpdateMenuRequest{
		Name: longName,
	}

	// Act
	item, err := service.Update(ctx, 1, req)

	// Assert
	if err == nil {
		t.Error("Expected error for name too long, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}

	if err.Error() != "menu item name cannot exceed 255 characters" {
		t.Errorf("Expected 'menu item name cannot exceed 255 characters', got: %s", err.Error())
	}
}

func TestUpdate_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.UpdateMenuRequest{
		Name: "Valid Name",
	}

	// Test cases
	invalidIDs := []int{0, -1, -100}

	for _, id := range invalidIDs {
		// Act
		item, err := service.Update(ctx, id, req)

		// Assert
		if err == nil {
			t.Errorf("Expected error for invalid ID %d, got nil", id)
		}

		if item != nil {
			t.Errorf("Expected nil item for invalid ID %d, got %v", id, item)
		}
	}
}

func TestUpdate_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			return nil, errors.New("menu item with id 999 not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.UpdateMenuRequest{
		Name: "Valid Name",
	}

	// Act
	item, err := service.Update(ctx, 999, req)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent ID, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item for non-existent ID, got %v", item)
	}
}

func TestUpdate_RepositoryUpdateError(t *testing.T) {
	// Arrange
	now := time.Now()
	existingItem := &domain.MenuItem{
		ID:        1,
		Name:      "Old Name",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	expectedError := errors.New("database update failed")

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 1 {
				return existingItem, nil
			}
			return nil, errors.New("not found")
		},
		UpdateFunc: func(ctx context.Context, item *domain.MenuItem) error {
			return expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	req := domain.UpdateMenuRequest{
		Name: "New Name",
	}

	// Act
	item, err := service.Update(ctx, 1, req)

	// Assert
	if err == nil {
		t.Error("Expected error from repository, got nil")
	}

	if item != nil {
		t.Errorf("Expected nil item on error, got %v", item)
	}
}

func TestUpdate_ValidNameLengths(t *testing.T) {
	// Arrange
	now := time.Now()
	existingItem := &domain.MenuItem{
		ID:        1,
		Name:      "Old Name",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 1 {
				// Return a copy to avoid mutation issues
				itemCopy := *existingItem
				return &itemCopy, nil
			}
			return nil, errors.New("not found")
		},
		UpdateFunc: func(ctx context.Context, item *domain.MenuItem) error {
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Test cases: boundary conditions
	testCases := []struct {
		name        string
		nameLength  int
		shouldPass  bool
	}{
		{"Single character", 1, true},
		{"255 characters (max)", 255, true},
		{"128 characters (middle)", 128, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create name of specific length
			testName := ""
			for i := 0; i < tc.nameLength; i++ {
				testName += "a"
			}

			req := domain.UpdateMenuRequest{
				Name: testName,
			}

			// Act
			item, err := service.Update(ctx, 1, req)

			// Assert
			if tc.shouldPass {
				if err != nil {
					t.Errorf("Expected no error for %s, got: %v", tc.name, err)
				}
				if item == nil {
					t.Errorf("Expected item for %s, got nil", tc.name)
				} else if item.Name != testName {
					t.Errorf("Expected name length %d, got %d", tc.nameLength, len(item.Name))
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for %s, got nil", tc.name)
				}
			}
		})
	}
}

// Test Delete method
func TestDelete_Success_NoChildren(t *testing.T) {
	// Arrange
	now := time.Now()
	parentID := 1
	existingItem := &domain.MenuItem{
		ID:        5,
		Name:      "Item to Delete",
		ParentID:  &parentID,
		Position:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Siblings at positions 0, 1 (deleted), 2, 3
	siblings := []domain.MenuItem{
		{ID: 4, Name: "Sibling 0", ParentID: &parentID, Position: 0},
		{ID: 6, Name: "Sibling 2", ParentID: &parentID, Position: 2},
		{ID: 7, Name: "Sibling 3", ParentID: &parentID, Position: 3},
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 5 {
				return existingItem, nil
			}
			return nil, errors.New("not found")
		},
		DeleteFunc: func(ctx context.Context, id int) (int, error) {
			if id == 5 {
				return 1, nil // Only the item itself deleted (no children)
			}
			return 0, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			if parentID != nil && *parentID == 1 {
				return siblings, nil
			}
			return []domain.MenuItem{}, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			// Verify correct position adjustments
			if len(updates) != 2 {
				t.Errorf("Expected 2 position updates, got %d", len(updates))
			}
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	count, err := service.Delete(ctx, 5)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected deleted count 1, got %d", count)
	}
}

func TestDelete_Success_WithChildren(t *testing.T) {
	// Arrange
	now := time.Now()
	existingItem := &domain.MenuItem{
		ID:        10,
		Name:      "Parent to Delete",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 10 {
				return existingItem, nil
			}
			return nil, errors.New("not found")
		},
		DeleteFunc: func(ctx context.Context, id int) (int, error) {
			if id == 10 {
				return 5, nil // Item + 4 descendants deleted (cascade)
			}
			return 0, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			// No siblings at root level after deletion
			return []domain.MenuItem{}, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			// No position updates needed (no siblings)
			if len(updates) != 0 {
				t.Errorf("Expected 0 position updates, got %d", len(updates))
			}
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	count, err := service.Delete(ctx, 10)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected deleted count 5 (parent + 4 descendants), got %d", count)
	}
}

func TestDelete_Success_AdjustsSiblingPositions(t *testing.T) {
	// Arrange
	now := time.Now()
	existingItem := &domain.MenuItem{
		ID:        20,
		Name:      "Middle Item",
		ParentID:  nil,
		Position:  2,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Root level siblings: positions 0, 1, 2 (deleted), 3, 4
	siblings := []domain.MenuItem{
		{ID: 18, Name: "Item 0", ParentID: nil, Position: 0},
		{ID: 19, Name: "Item 1", ParentID: nil, Position: 1},
		{ID: 21, Name: "Item 3", ParentID: nil, Position: 3},
		{ID: 22, Name: "Item 4", ParentID: nil, Position: 4},
	}

	var capturedUpdates []domain.PositionUpdate

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 20 {
				return existingItem, nil
			}
			return nil, errors.New("not found")
		},
		DeleteFunc: func(ctx context.Context, id int) (int, error) {
			if id == 20 {
				return 1, nil
			}
			return 0, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return siblings, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			capturedUpdates = updates
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	count, err := service.Delete(ctx, 20)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected deleted count 1, got %d", count)
	}

	// Verify position updates
	if len(capturedUpdates) != 2 {
		t.Fatalf("Expected 2 position updates (items with pos > 2), got %d", len(capturedUpdates))
	}

	// Verify that items at position 3 and 4 are decremented to 2 and 3
	expectedUpdates := map[int]int{
		21: 2, // Item at position 3 → 2
		22: 3, // Item at position 4 → 3
	}

	for _, update := range capturedUpdates {
		expectedPos, exists := expectedUpdates[update.ID]
		if !exists {
			t.Errorf("Unexpected position update for ID %d", update.ID)
		}
		if update.Position != expectedPos {
			t.Errorf("Expected ID %d to have position %d, got %d", update.ID, expectedPos, update.Position)
		}
	}
}

func TestDelete_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Test cases
	invalidIDs := []int{0, -1, -100}

	for _, id := range invalidIDs {
		// Act
		count, err := service.Delete(ctx, id)

		// Assert
		if err == nil {
			t.Errorf("Expected error for invalid ID %d, got nil", id)
		}

		if count != 0 {
			t.Errorf("Expected count 0 for invalid ID %d, got %d", id, count)
		}
	}
}

func TestDelete_NotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			return nil, errors.New("menu item with id 999 not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	count, err := service.Delete(ctx, 999)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent ID, got nil")
	}

	if count != 0 {
		t.Errorf("Expected count 0 for non-existent ID, got %d", count)
	}
}

func TestDelete_RepositoryDeleteError(t *testing.T) {
	// Arrange
	now := time.Now()
	existingItem := &domain.MenuItem{
		ID:        30,
		Name:      "Item to Delete",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	expectedError := errors.New("database delete failed")

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 30 {
				return existingItem, nil
			}
			return nil, errors.New("not found")
		},
		DeleteFunc: func(ctx context.Context, id int) (int, error) {
			return 0, expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	count, err := service.Delete(ctx, 30)

	// Assert
	if err == nil {
		t.Error("Expected error from repository, got nil")
	}

	if count != 0 {
		t.Errorf("Expected count 0 on error, got %d", count)
	}
}

func TestDelete_FindByParentIDError(t *testing.T) {
	// Arrange
	now := time.Now()
	existingItem := &domain.MenuItem{
		ID:        40,
		Name:      "Item to Delete",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	expectedError := errors.New("failed to fetch siblings")

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 40 {
				return existingItem, nil
			}
			return nil, errors.New("not found")
		},
		DeleteFunc: func(ctx context.Context, id int) (int, error) {
			if id == 40 {
				return 1, nil
			}
			return 0, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return nil, expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	count, err := service.Delete(ctx, 40)

	// Assert
	if err == nil {
		t.Error("Expected error from FindByParentID, got nil")
	}

	// Note: count is still returned as the delete succeeded
	if count != 1 {
		t.Errorf("Expected count 1 (delete succeeded), got %d", count)
	}
}

func TestDelete_UpdatePositionsError(t *testing.T) {
	// Arrange
	now := time.Now()
	parentID := 1
	existingItem := &domain.MenuItem{
		ID:        50,
		Name:      "Item to Delete",
		ParentID:  &parentID,
		Position:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 51, Name: "Sibling 2", ParentID: &parentID, Position: 2},
	}

	expectedError := errors.New("failed to update positions")

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 50 {
				return existingItem, nil
			}
			return nil, errors.New("not found")
		},
		DeleteFunc: func(ctx context.Context, id int) (int, error) {
			if id == 50 {
				return 1, nil
			}
			return 0, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			if parentID != nil && *parentID == 1 {
				return siblings, nil
			}
			return []domain.MenuItem{}, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			return expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	count, err := service.Delete(ctx, 50)

	// Assert
	if err == nil {
		t.Error("Expected error from UpdatePositions, got nil")
	}

	// Note: count is still returned as the delete succeeded
	if count != 1 {
		t.Errorf("Expected count 1 (delete succeeded), got %d", count)
	}
}

func TestDelete_NoSiblings(t *testing.T) {
	// Arrange
	now := time.Now()
	parentID := 5
	existingItem := &domain.MenuItem{
		ID:        60,
		Name:      "Only Child",
		ParentID:  &parentID,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 60 {
				return existingItem, nil
			}
			return nil, errors.New("not found")
		},
		DeleteFunc: func(ctx context.Context, id int) (int, error) {
			if id == 60 {
				return 1, nil
			}
			return 0, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			// No siblings remaining
			return []domain.MenuItem{}, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			// Should not be called if no updates
			t.Error("UpdatePositions should not be called when there are no siblings to update")
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	count, err := service.Delete(ctx, 60)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected deleted count 1, got %d", count)
	}
}


// Test Reorder method
func TestReorder_Success_MoveDown(t *testing.T) {
	// Arrange: Siblings at positions [0, 1, 2, 3, 4], move position 1 to position 3
	now := time.Now()
	itemToMove := &domain.MenuItem{
		ID:        10,
		Name:      "Item at Pos 1",
		ParentID:  nil,
		Position:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 9, Name: "Item at Pos 0", ParentID: nil, Position: 0},
		{ID: 10, Name: "Item at Pos 1", ParentID: nil, Position: 1},
		{ID: 11, Name: "Item at Pos 2", ParentID: nil, Position: 2},
		{ID: 12, Name: "Item at Pos 3", ParentID: nil, Position: 3},
		{ID: 13, Name: "Item at Pos 4", ParentID: nil, Position: 4},
	}

	var capturedUpdates []domain.PositionUpdate

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 10 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return siblings, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			capturedUpdates = updates
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act: Move from position 1 to position 3
	affectedItems, err := service.Reorder(ctx, 10, 3)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	// Verify position updates: items at positions 2, 3 should decrement, item 10 moves to 3
	if len(capturedUpdates) != 3 {
		t.Fatalf("Expected 3 position updates, got %d", len(capturedUpdates))
	}

	// Expected updates:
	// ID 11 (pos 2) -> pos 1
	// ID 12 (pos 3) -> pos 2
	// ID 10 (pos 1) -> pos 3
	expectedUpdates := map[int]int{
		11: 1, // Position 2 -> 1
		12: 2, // Position 3 -> 2
		10: 3, // Position 1 -> 3 (moved item)
	}

	for _, update := range capturedUpdates {
		expectedPos, exists := expectedUpdates[update.ID]
		if !exists {
			t.Errorf("Unexpected position update for ID %d", update.ID)
		}
		if update.Position != expectedPos {
			t.Errorf("Expected ID %d to have position %d, got %d", update.ID, expectedPos, update.Position)
		}
	}
}

func TestReorder_Success_MoveUp(t *testing.T) {
	// Arrange: Siblings at positions [0, 1, 2, 3, 4], move position 3 to position 1
	now := time.Now()
	itemToMove := &domain.MenuItem{
		ID:        12,
		Name:      "Item at Pos 3",
		ParentID:  nil,
		Position:  3,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 9, Name: "Item at Pos 0", ParentID: nil, Position: 0},
		{ID: 10, Name: "Item at Pos 1", ParentID: nil, Position: 1},
		{ID: 11, Name: "Item at Pos 2", ParentID: nil, Position: 2},
		{ID: 12, Name: "Item at Pos 3", ParentID: nil, Position: 3},
		{ID: 13, Name: "Item at Pos 4", ParentID: nil, Position: 4},
	}

	var capturedUpdates []domain.PositionUpdate

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 12 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return siblings, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			capturedUpdates = updates
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act: Move from position 3 to position 1
	affectedItems, err := service.Reorder(ctx, 12, 1)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	// Verify position updates: items at positions 1, 2 should increment, item 12 moves to 1
	if len(capturedUpdates) != 3 {
		t.Fatalf("Expected 3 position updates, got %d", len(capturedUpdates))
	}

	// Expected updates:
	// ID 10 (pos 1) -> pos 2
	// ID 11 (pos 2) -> pos 3
	// ID 12 (pos 3) -> pos 1 (moved item)
	expectedUpdates := map[int]int{
		10: 2, // Position 1 -> 2
		11: 3, // Position 2 -> 3
		12: 1, // Position 3 -> 1 (moved item)
	}

	for _, update := range capturedUpdates {
		expectedPos, exists := expectedUpdates[update.ID]
		if !exists {
			t.Errorf("Unexpected position update for ID %d", update.ID)
		}
		if update.Position != expectedPos {
			t.Errorf("Expected ID %d to have position %d, got %d", update.ID, expectedPos, update.Position)
		}
	}
}

func TestReorder_Success_SamePosition(t *testing.T) {
	// Arrange: Item already at target position
	now := time.Now()
	itemToMove := &domain.MenuItem{
		ID:        10,
		Name:      "Item at Pos 2",
		ParentID:  nil,
		Position:  2,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 9, Name: "Item at Pos 0", ParentID: nil, Position: 0},
		{ID: 10, Name: "Item at Pos 1", ParentID: nil, Position: 1},
		{ID: 11, Name: "Item at Pos 2", ParentID: nil, Position: 2},
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 10 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return siblings, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			t.Error("UpdatePositions should not be called when position is unchanged")
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act: Move to same position (2 to 2)
	affectedItems, err := service.Reorder(ctx, 10, 2)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	if len(affectedItems) != 3 {
		t.Errorf("Expected 3 siblings returned, got %d", len(affectedItems))
	}
}

func TestReorder_Success_WithParent(t *testing.T) {
	// Arrange: Test reordering within a non-root sibling group
	now := time.Now()
	parentID := 5

	itemToMove := &domain.MenuItem{
		ID:        20,
		Name:      "Child at Pos 0",
		ParentID:  &parentID,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 20, Name: "Child at Pos 0", ParentID: &parentID, Position: 0},
		{ID: 21, Name: "Child at Pos 1", ParentID: &parentID, Position: 1},
		{ID: 22, Name: "Child at Pos 2", ParentID: &parentID, Position: 2},
	}

	var capturedUpdates []domain.PositionUpdate

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 20 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			if parentID != nil && *parentID == 5 {
				return siblings, nil
			}
			return []domain.MenuItem{}, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			capturedUpdates = updates
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act: Move from position 0 to position 2
	affectedItems, err := service.Reorder(ctx, 20, 2)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	// Verify correct position updates
	if len(capturedUpdates) != 3 {
		t.Fatalf("Expected 3 position updates, got %d", len(capturedUpdates))
	}
}

func TestReorder_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Test cases
	invalidIDs := []int{0, -1, -100}

	for _, id := range invalidIDs {
		// Act
		items, err := service.Reorder(ctx, id, 0)

		// Assert
		if err == nil {
			t.Errorf("Expected error for invalid ID %d, got nil", id)
		}

		if items != nil {
			t.Errorf("Expected nil items for invalid ID %d, got %v", id, items)
		}
	}
}

func TestReorder_NegativePosition(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	items, err := service.Reorder(ctx, 1, -5)

	// Assert
	if err == nil {
		t.Error("Expected error for negative position, got nil")
	}

	if items != nil {
		t.Errorf("Expected nil items on error, got %v", items)
	}

	if err.Error() != "position must be non-negative, got -5" {
		t.Errorf("Expected 'position must be non-negative' error, got: %s", err.Error())
	}
}

func TestReorder_ItemNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			return nil, errors.New("menu item with id 999 not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	items, err := service.Reorder(ctx, 999, 2)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent item, got nil")
	}

	if items != nil {
		t.Errorf("Expected nil items on error, got %v", items)
	}
}

func TestReorder_PositionOutOfBounds(t *testing.T) {
	// Arrange
	now := time.Now()
	itemToMove := &domain.MenuItem{
		ID:        10,
		Name:      "Item",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Only 3 siblings (positions 0, 1, 2), so max valid position is 2
	siblings := []domain.MenuItem{
		{ID: 10, Name: "Item 0", ParentID: nil, Position: 0},
		{ID: 11, Name: "Item 1", ParentID: nil, Position: 1},
		{ID: 12, Name: "Item 2", ParentID: nil, Position: 2},
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 10 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return siblings, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act: Try to move to position 5 (out of bounds)
	items, err := service.Reorder(ctx, 10, 5)

	// Assert
	if err == nil {
		t.Error("Expected error for position out of bounds, got nil")
	}

	if items != nil {
		t.Errorf("Expected nil items on error, got %v", items)
	}

	expectedError := "new position 5 exceeds maximum valid position 2"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got: %s", expectedError, err.Error())
	}
}

func TestReorder_FindByParentIDError(t *testing.T) {
	// Arrange
	now := time.Now()
	itemToMove := &domain.MenuItem{
		ID:        10,
		Name:      "Item",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	expectedError := errors.New("database query failed")

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 10 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return nil, expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	items, err := service.Reorder(ctx, 10, 2)

	// Assert
	if err == nil {
		t.Error("Expected error from FindByParentID, got nil")
	}

	if items != nil {
		t.Errorf("Expected nil items on error, got %v", items)
	}
}

func TestReorder_UpdatePositionsError(t *testing.T) {
	// Arrange
	now := time.Now()
	itemToMove := &domain.MenuItem{
		ID:        10,
		Name:      "Item",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 10, Name: "Item 0", ParentID: nil, Position: 0},
		{ID: 11, Name: "Item 1", ParentID: nil, Position: 1},
		{ID: 12, Name: "Item 2", ParentID: nil, Position: 2},
	}

	expectedError := errors.New("database update failed")

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 10 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return siblings, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			return expectedError
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	items, err := service.Reorder(ctx, 10, 2)

	// Assert
	if err == nil {
		t.Error("Expected error from UpdatePositions, got nil")
	}

	if items != nil {
		t.Errorf("Expected nil items on error, got %v", items)
	}
}

func TestReorder_EdgeCase_TwoItems(t *testing.T) {
	// Arrange: Only 2 items, swap them
	now := time.Now()
	itemToMove := &domain.MenuItem{
		ID:        10,
		Name:      "Item 0",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 10, Name: "Item 0", ParentID: nil, Position: 0},
		{ID: 11, Name: "Item 1", ParentID: nil, Position: 1},
	}

	var capturedUpdates []domain.PositionUpdate

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 10 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return siblings, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			capturedUpdates = updates
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act: Move from position 0 to position 1
	affectedItems, err := service.Reorder(ctx, 10, 1)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	// Verify position updates: item 11 moves down, item 10 moves to 1
	if len(capturedUpdates) != 2 {
		t.Fatalf("Expected 2 position updates, got %d", len(capturedUpdates))
	}

	expectedUpdates := map[int]int{
		11: 0, // Position 1 -> 0
		10: 1, // Position 0 -> 1 (moved item)
	}

	for _, update := range capturedUpdates {
		expectedPos, exists := expectedUpdates[update.ID]
		if !exists {
			t.Errorf("Unexpected position update for ID %d", update.ID)
		}
		if update.Position != expectedPos {
			t.Errorf("Expected ID %d to have position %d, got %d", update.ID, expectedPos, update.Position)
		}
	}
}

func TestReorder_EdgeCase_MoveToFirstPosition(t *testing.T) {
	// Arrange: Move last item to first position
	now := time.Now()
	itemToMove := &domain.MenuItem{
		ID:        13,
		Name:      "Item at last position",
		ParentID:  nil,
		Position:  3,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 10, Name: "Item 0", ParentID: nil, Position: 0},
		{ID: 11, Name: "Item 1", ParentID: nil, Position: 1},
		{ID: 12, Name: "Item 2", ParentID: nil, Position: 2},
		{ID: 13, Name: "Item 3", ParentID: nil, Position: 3},
	}

	var capturedUpdates []domain.PositionUpdate

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 13 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return siblings, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			capturedUpdates = updates
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act: Move from position 3 to position 0
	affectedItems, err := service.Reorder(ctx, 13, 0)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	// Verify all items shift up
	if len(capturedUpdates) != 4 {
		t.Fatalf("Expected 4 position updates, got %d", len(capturedUpdates))
	}

	expectedUpdates := map[int]int{
		10: 1, // Position 0 -> 1
		11: 2, // Position 1 -> 2
		12: 3, // Position 2 -> 3
		13: 0, // Position 3 -> 0 (moved item)
	}

	for _, update := range capturedUpdates {
		expectedPos, exists := expectedUpdates[update.ID]
		if !exists {
			t.Errorf("Unexpected position update for ID %d", update.ID)
		}
		if update.Position != expectedPos {
			t.Errorf("Expected ID %d to have position %d, got %d", update.ID, expectedPos, update.Position)
		}
	}
}

func TestReorder_EdgeCase_MoveToLastPosition(t *testing.T) {
	// Arrange: Move first item to last position
	now := time.Now()
	itemToMove := &domain.MenuItem{
		ID:        10,
		Name:      "Item at first position",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 10, Name: "Item 0", ParentID: nil, Position: 0},
		{ID: 11, Name: "Item 1", ParentID: nil, Position: 1},
		{ID: 12, Name: "Item 2", ParentID: nil, Position: 2},
		{ID: 13, Name: "Item 3", ParentID: nil, Position: 3},
	}

	var capturedUpdates []domain.PositionUpdate

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 10 {
				return itemToMove, nil
			}
			return nil, errors.New("not found")
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			return siblings, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			capturedUpdates = updates
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act: Move from position 0 to position 3
	affectedItems, err := service.Reorder(ctx, 10, 3)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	// Verify all items shift down
	if len(capturedUpdates) != 4 {
		t.Fatalf("Expected 4 position updates, got %d", len(capturedUpdates))
	}

	expectedUpdates := map[int]int{
		11: 0, // Position 1 -> 0
		12: 1, // Position 2 -> 1
		13: 2, // Position 3 -> 2
		10: 3, // Position 0 -> 3 (moved item)
	}

	for _, update := range capturedUpdates {
		expectedPos, exists := expectedUpdates[update.ID]
		if !exists {
			t.Errorf("Unexpected position update for ID %d", update.ID)
		}
		if update.Position != expectedPos {
			t.Errorf("Expected ID %d to have position %d, got %d", update.ID, expectedPos, update.Position)
		}
	}
}

// Test GetDescendants method
func TestGetDescendants_Success_NoDescendants(t *testing.T) {
	// Arrange
	now := time.Now()
	leafItem := &domain.MenuItem{
		ID:        100,
		Name:      "Leaf Item",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 100 {
				return leafItem, nil
			}
			return nil, errors.New("not found")
		},
		GetDescendantIDsFunc: func(ctx context.Context, id int) ([]int, error) {
			return []int{}, nil // No descendants
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	descendants, err := service.GetDescendants(ctx, 100)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if descendants == nil {
		t.Error("Expected empty slice, got nil")
	}

	if len(descendants) != 0 {
		t.Errorf("Expected 0 descendants, got %d", len(descendants))
	}
}

func TestGetDescendants_Success_WithDescendants(t *testing.T) {
	// Arrange
	now := time.Now()
	parentItem := &domain.MenuItem{
		ID:        200,
		Name:      "Parent Item",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	descendantItems := map[int]*domain.MenuItem{
		201: {ID: 201, Name: "Child 1", ParentID: intPtr(200), Position: 0, CreatedAt: now, UpdatedAt: now},
		202: {ID: 202, Name: "Child 2", ParentID: intPtr(200), Position: 1, CreatedAt: now, UpdatedAt: now},
		203: {ID: 203, Name: "Grandchild 1", ParentID: intPtr(201), Position: 0, CreatedAt: now, UpdatedAt: now},
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 200 {
				return parentItem, nil
			}
			if item, ok := descendantItems[id]; ok {
				return item, nil
			}
			return nil, errors.New("not found")
		},
		GetDescendantIDsFunc: func(ctx context.Context, id int) ([]int, error) {
			if id == 200 {
				return []int{201, 202, 203}, nil
			}
			return []int{}, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	descendants, err := service.GetDescendants(ctx, 200)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(descendants) != 3 {
		t.Errorf("Expected 3 descendants, got %d", len(descendants))
	}

	// Verify all descendants are present
	descendantIDs := make(map[int]bool)
	for _, d := range descendants {
		descendantIDs[d.ID] = true
	}

	for expectedID := range descendantItems {
		if !descendantIDs[expectedID] {
			t.Errorf("Expected descendant ID %d not found", expectedID)
		}
	}
}

func TestGetDescendants_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Test cases
	invalidIDs := []int{0, -1, -100}

	for _, id := range invalidIDs {
		// Act
		descendants, err := service.GetDescendants(ctx, id)

		// Assert
		if err == nil {
			t.Errorf("Expected error for invalid ID %d, got nil", id)
		}

		if descendants != nil {
			t.Errorf("Expected nil descendants for invalid ID %d, got %v", id, descendants)
		}
	}
}

func TestGetDescendants_ItemNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			return nil, errors.New("menu item with id 999 not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	descendants, err := service.GetDescendants(ctx, 999)

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent ID, got nil")
	}

	if descendants != nil {
		t.Errorf("Expected nil descendants for non-existent ID, got %v", descendants)
	}
}

// Test ValidateMove method
func TestValidateMove_Success_ToRoot(t *testing.T) {
	// Arrange
	now := time.Now()
	item := &domain.MenuItem{
		ID:        300,
		Name:      "Item to Move",
		ParentID:  intPtr(1),
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 300 {
				return item, nil
			}
			return nil, errors.New("not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	err := service.ValidateMove(ctx, 300, nil)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for moving to root, got: %v", err)
	}
}

func TestValidateMove_Success_ToValidParent(t *testing.T) {
	// Arrange
	now := time.Now()
	item := &domain.MenuItem{
		ID:        400,
		Name:      "Item to Move",
		ParentID:  intPtr(1),
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	newParent := &domain.MenuItem{
		ID:        2,
		Name:      "New Parent",
		ParentID:  nil,
		Position:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 400 {
				return item, nil
			}
			if id == 2 {
				return newParent, nil
			}
			return nil, errors.New("not found")
		},
		GetAncestorsFunc: func(ctx context.Context, id int) ([]domain.MenuItem, error) {
			if id == 2 {
				// New parent has no ancestors (root level)
				return []domain.MenuItem{}, nil
			}
			return []domain.MenuItem{}, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	err := service.ValidateMove(ctx, 400, intPtr(2))

	// Assert
	if err != nil {
		t.Errorf("Expected no error for valid move, got: %v", err)
	}
}

func TestValidateMove_Error_ItemNotFound(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			return nil, errors.New("menu item with id 500 not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	err := service.ValidateMove(ctx, 500, intPtr(2))

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent item, got nil")
	}
}

func TestValidateMove_Error_ParentNotFound(t *testing.T) {
	// Arrange
	now := time.Now()
	item := &domain.MenuItem{
		ID:        600,
		Name:      "Item to Move",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 600 {
				return item, nil
			}
			return nil, errors.New("parent menu item with id 999 not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	err := service.ValidateMove(ctx, 600, intPtr(999))

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent parent, got nil")
	}
}

func TestValidateMove_Error_MoveToItself(t *testing.T) {
	// Arrange
	now := time.Now()
	item := &domain.MenuItem{
		ID:        700,
		Name:      "Item to Move",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 700 {
				return item, nil
			}
			return nil, errors.New("not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	err := service.ValidateMove(ctx, 700, intPtr(700))

	// Assert
	if err == nil {
		t.Error("Expected error for moving item to itself, got nil")
	}

	if err != nil && err.Error() != "cannot move menu item to itself" {
		t.Errorf("Expected 'cannot move menu item to itself', got: %s", err.Error())
	}
}

func TestValidateMove_Error_CircularReference(t *testing.T) {
	// Arrange
	now := time.Now()
	// Item 800 is parent of item 801
	parentItem := &domain.MenuItem{
		ID:        800,
		Name:      "Parent Item",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	childItem := &domain.MenuItem{
		ID:        801,
		Name:      "Child Item",
		ParentID:  intPtr(800),
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 800 {
				return parentItem, nil
			}
			if id == 801 {
				return childItem, nil
			}
			return nil, errors.New("not found")
		},
		GetAncestorsFunc: func(ctx context.Context, id int) ([]domain.MenuItem, error) {
			if id == 801 {
				// Child's ancestors include parent 800
				return []domain.MenuItem{*parentItem}, nil
			}
			return []domain.MenuItem{}, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act - Try to move parent (800) to be child of its own child (801)
	err := service.ValidateMove(ctx, 800, intPtr(801))

	// Assert
	if err == nil {
		t.Error("Expected error for circular reference, got nil")
	}

	if err != nil && err.Error() != "cannot move menu item: circular reference detected (new parent would become descendant of moved item)" {
		t.Errorf("Expected circular reference error, got: %s", err.Error())
	}
}

func TestValidateMove_InvalidID(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{}
	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Test cases
	invalidIDs := []int{0, -1, -100}

	for _, id := range invalidIDs {
		// Act
		err := service.ValidateMove(ctx, id, intPtr(1))

		// Assert
		if err == nil {
			t.Errorf("Expected error for invalid ID %d, got nil", id)
		}
	}
}

// Test Move method
func TestMove_Success_ToNewParent(t *testing.T) {
	// Arrange
	now := time.Now()
	parentID1 := 1
	parentID2 := 2

	// Item to move (currently under parent 1 at position 1)
	itemToMove := &domain.MenuItem{
		ID:        900,
		Name:      "Item to Move",
		ParentID:  &parentID1,
		Position:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Source siblings
	sourceSiblings := []domain.MenuItem{
		{ID: 901, Name: "Source Sibling 0", ParentID: &parentID1, Position: 0},
		{ID: 900, Name: "Item to Move", ParentID: &parentID1, Position: 1},
		{ID: 902, Name: "Source Sibling 2", ParentID: &parentID1, Position: 2},
	}

	// Destination siblings (before move)
	destSiblings := []domain.MenuItem{
		{ID: 903, Name: "Dest Sibling 0", ParentID: &parentID2, Position: 0},
		{ID: 904, Name: "Dest Sibling 1", ParentID: &parentID2, Position: 1},
	}

	// After move, destination siblings
	destSiblingsAfter := []domain.MenuItem{
		{ID: 903, Name: "Dest Sibling 0", ParentID: &parentID2, Position: 0},
		{ID: 904, Name: "Dest Sibling 1", ParentID: &parentID2, Position: 1},
		{ID: 900, Name: "Item to Move", ParentID: &parentID2, Position: 2},
	}

	sourceSiblingsAfter := []domain.MenuItem{
		{ID: 901, Name: "Source Sibling 0", ParentID: &parentID1, Position: 0},
		{ID: 902, Name: "Source Sibling 2", ParentID: &parentID1, Position: 1},
	}

	newParent := &domain.MenuItem{
		ID:       2,
		Name:     "New Parent",
		ParentID: nil,
		Position: 0,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 900 {
				itemCopy := *itemToMove
				return &itemCopy, nil
			}
			if id == 2 {
				return newParent, nil
			}
			return nil, errors.New("not found")
		},
		GetAncestorsFunc: func(ctx context.Context, id int) ([]domain.MenuItem, error) {
			return []domain.MenuItem{}, nil
		},
		GetMaxPositionFunc: func(ctx context.Context, parentID *int) (int, error) {
			if parentID != nil && *parentID == 2 {
				return 1, nil // Max position in destination is 1
			}
			return 0, nil
		},
		WithTransactionFunc: func(ctx context.Context, fn func(repo interface{}) error) error {
			// Create a transaction mock repo
			txRepo := &MockMenuRepository{
				FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
					if parentID != nil && *parentID == 1 {
						return sourceSiblings, nil
					}
					if parentID != nil && *parentID == 2 {
						return destSiblings, nil
					}
					return []domain.MenuItem{}, nil
				},
				UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
					return nil
				},
				UpdateFunc: func(ctx context.Context, item *domain.MenuItem) error {
					return nil
				},
			}
			return fn(txRepo)
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			// After transaction
			if parentID != nil && *parentID == 1 {
				return sourceSiblingsAfter, nil
			}
			if parentID != nil && *parentID == 2 {
				return destSiblingsAfter, nil
			}
			return []domain.MenuItem{}, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	affectedItems, err := service.Move(ctx, 900, intPtr(2), nil)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	// Should return items from both source and destination groups
	if len(affectedItems) < 3 {
		t.Errorf("Expected at least 3 affected items, got %d", len(affectedItems))
	}
}

func TestMove_Success_ToRoot(t *testing.T) {
	// Arrange
	now := time.Now()
	parentID1 := 1

	itemToMove := &domain.MenuItem{
		ID:        1000,
		Name:      "Item to Move to Root",
		ParentID:  &parentID1,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	sourceSiblings := []domain.MenuItem{
		{ID: 1000, Name: "Item to Move to Root", ParentID: &parentID1, Position: 0},
	}

	rootItems := []domain.MenuItem{
		{ID: 1001, Name: "Root Item 1", ParentID: nil, Position: 0},
		{ID: 1002, Name: "Root Item 2", ParentID: nil, Position: 1},
	}

	sourceSiblingsAfter := []domain.MenuItem{} // Empty after item moved out

	rootItemsAfter := []domain.MenuItem{
		{ID: 1001, Name: "Root Item 1", ParentID: nil, Position: 0},
		{ID: 1002, Name: "Root Item 2", ParentID: nil, Position: 1},
		{ID: 1000, Name: "Item to Move to Root", ParentID: nil, Position: 2},
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 1000 {
				itemCopy := *itemToMove
				return &itemCopy, nil
			}
			return nil, errors.New("not found")
		},
		GetMaxPositionFunc: func(ctx context.Context, parentID *int) (int, error) {
			if parentID == nil {
				return 1, nil // Max position at root is 1
			}
			return 0, nil
		},
		WithTransactionFunc: func(ctx context.Context, fn func(repo interface{}) error) error {
			txRepo := &MockMenuRepository{
				FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
					if parentID != nil && *parentID == 1 {
						return sourceSiblings, nil
					}
					if parentID == nil {
						return rootItems, nil
					}
					return []domain.MenuItem{}, nil
				},
				UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
					return nil
				},
				UpdateFunc: func(ctx context.Context, item *domain.MenuItem) error {
					return nil
				},
			}
			return fn(txRepo)
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			// After transaction
			if parentID != nil && *parentID == 1 {
				return sourceSiblingsAfter, nil
			}
			if parentID == nil {
				return rootItemsAfter, nil
			}
			return []domain.MenuItem{}, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	affectedItems, err := service.Move(ctx, 1000, nil, nil)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	// Should return items from both source and destination (root)
	if len(affectedItems) < 1 {
		t.Errorf("Expected at least 1 affected item, got %d", len(affectedItems))
	}
}

func TestMove_Success_SameParentDifferentPosition(t *testing.T) {
	// Arrange
	now := time.Now()
	parentID := 1

	itemToMove := &domain.MenuItem{
		ID:        1100,
		Name:      "Item to Reorder",
		ParentID:  &parentID,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	parentItem := &domain.MenuItem{
		ID:        1,
		Name:      "Parent Item",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 1100, Name: "Item to Reorder", ParentID: &parentID, Position: 0},
		{ID: 1101, Name: "Sibling 1", ParentID: &parentID, Position: 1},
		{ID: 1102, Name: "Sibling 2", ParentID: &parentID, Position: 2},
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 1100 {
				itemCopy := *itemToMove
				return &itemCopy, nil
			}
			if id == 1 {
				return parentItem, nil
			}
			return nil, errors.New("not found")
		},
		GetAncestorsFunc: func(ctx context.Context, id int) ([]domain.MenuItem, error) {
			return []domain.MenuItem{}, nil
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			if parentID != nil && *parentID == 1 {
				return siblings, nil
			}
			return []domain.MenuItem{}, nil
		},
		UpdatePositionsFunc: func(ctx context.Context, updates []domain.PositionUpdate) error {
			return nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act - Move to position 2 within same parent (triggers Reorder)
	affectedItems, err := service.Move(ctx, 1100, &parentID, intPtr(2))

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}
}

func TestMove_Error_ValidationFails(t *testing.T) {
	// Arrange
	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			return nil, errors.New("menu item with id 1200 not found")
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	affectedItems, err := service.Move(ctx, 1200, intPtr(2), nil)

	// Assert
	if err == nil {
		t.Error("Expected error from validation, got nil")
	}

	if affectedItems != nil {
		t.Errorf("Expected nil affected items on error, got %v", affectedItems)
	}
}

func TestMove_Error_NegativePosition(t *testing.T) {
	// Arrange
	now := time.Now()
	parentID := 1

	itemToMove := &domain.MenuItem{
		ID:        1300,
		Name:      "Item to Move",
		ParentID:  &parentID,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	newParent := &domain.MenuItem{
		ID:       2,
		Name:     "New Parent",
		ParentID: nil,
		Position: 0,
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 1300 {
				return itemToMove, nil
			}
			if id == 2 {
				return newParent, nil
			}
			return nil, errors.New("not found")
		},
		GetAncestorsFunc: func(ctx context.Context, id int) ([]domain.MenuItem, error) {
			return []domain.MenuItem{}, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act
	affectedItems, err := service.Move(ctx, 1300, intPtr(2), intPtr(-1))

	// Assert
	if err == nil {
		t.Error("Expected error for negative position, got nil")
	}

	if affectedItems != nil {
		t.Errorf("Expected nil affected items on error, got %v", affectedItems)
	}
}

func TestMove_NoChange_SameParentSamePosition(t *testing.T) {
	// Arrange
	now := time.Now()
	parentID := 1

	itemToMove := &domain.MenuItem{
		ID:        1400,
		Name:      "Item",
		ParentID:  &parentID,
		Position:  1,
		CreatedAt: now,
		UpdatedAt: now,
	}

	parentItem := &domain.MenuItem{
		ID:        1,
		Name:      "Parent Item",
		ParentID:  nil,
		Position:  0,
		CreatedAt: now,
		UpdatedAt: now,
	}

	siblings := []domain.MenuItem{
		{ID: 1401, Name: "Sibling 0", ParentID: &parentID, Position: 0},
		{ID: 1400, Name: "Item", ParentID: &parentID, Position: 1},
		{ID: 1402, Name: "Sibling 2", ParentID: &parentID, Position: 2},
	}

	mockRepo := &MockMenuRepository{
		FindByIDFunc: func(ctx context.Context, id int) (*domain.MenuItem, error) {
			if id == 1400 {
				itemCopy := *itemToMove
				return &itemCopy, nil
			}
			if id == 1 {
				return parentItem, nil
			}
			return nil, errors.New("not found")
		},
		GetAncestorsFunc: func(ctx context.Context, id int) ([]domain.MenuItem, error) {
			return []domain.MenuItem{}, nil
		},
		FindByParentIDFunc: func(ctx context.Context, parentID *int) ([]domain.MenuItem, error) {
			if parentID != nil && *parentID == 1 {
				return siblings, nil
			}
			return []domain.MenuItem{}, nil
		},
	}

	service := NewMenuService(mockRepo)
	ctx := context.Background()

	// Act - Move to same parent and same position (no-op)
	affectedItems, err := service.Move(ctx, 1400, &parentID, intPtr(1))

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if affectedItems == nil {
		t.Fatal("Expected affected items, got nil")
	}

	// Should return current siblings
	if len(affectedItems) != 3 {
		t.Errorf("Expected 3 siblings, got %d", len(affectedItems))
	}
}

// Helper function to create int pointers
func intPtr(i int) *int {
	return &i
}
