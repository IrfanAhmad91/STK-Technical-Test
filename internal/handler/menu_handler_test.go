package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stk/menu-tree-api/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMenuService is a mock implementation of MenuService
type MockMenuService struct {
	mock.Mock
}

func (m *MockMenuService) GetAll(ctx context.Context) ([]domain.MenuItem, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MenuItem), args.Error(1)
}

func (m *MockMenuService) GetByID(ctx context.Context, id int) (*domain.MenuItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuItem), args.Error(1)
}

func (m *MockMenuService) Create(ctx context.Context, req domain.CreateMenuRequest) (*domain.MenuItem, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuItem), args.Error(1)
}

func (m *MockMenuService) Update(ctx context.Context, id int, req domain.UpdateMenuRequest) (*domain.MenuItem, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MenuItem), args.Error(1)
}

func (m *MockMenuService) Delete(ctx context.Context, id int) (int, error) {
	args := m.Called(ctx, id)
	return args.Int(0), args.Error(1)
}

func (m *MockMenuService) Reorder(ctx context.Context, id int, newPosition int) ([]domain.MenuItem, error) {
	args := m.Called(ctx, id, newPosition)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MenuItem), args.Error(1)
}

func (m *MockMenuService) Move(ctx context.Context, id int, newParentID *int, position *int) ([]domain.MenuItem, error) {
	args := m.Called(ctx, id, newParentID, position)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MenuItem), args.Error(1)
}

func (m *MockMenuService) GetDescendants(ctx context.Context, id int) ([]domain.MenuItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MenuItem), args.Error(1)
}

func (m *MockMenuService) ValidateMove(ctx context.Context, id int, newParentID *int) error {
	args := m.Called(ctx, id, newParentID)
	return args.Error(0)
}

// Test setup helper
func setupTestRouter(handler *MenuHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// TestGetAll tests the GetAll handler
func TestGetAll(t *testing.T) {
	t.Run("Success - returns all menu items", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.GET("/menus", handler.GetAll)

		// Mock data
		expectedItems := []domain.MenuItem{
			{ID: 1, Name: "Menu 1", Position: 0},
			{ID: 2, Name: "Menu 2", Position: 1},
		}

		mockService.On("GetAll", mock.Anything).Return(expectedItems, nil)

		// Execute
		req, _ := http.NewRequest("GET", "/menus", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var result []domain.MenuItem
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "Menu 1", result[0].Name)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - database error", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.GET("/menus", handler.GetAll)

		mockService.On("GetAll", mock.Anything).Return(nil, errors.New("database error"))

		// Execute
		req, _ := http.NewRequest("GET", "/menus", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var result domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "DATABASE_ERROR", result.Code)

		mockService.AssertExpectations(t)
	})
}

// TestGetByID tests the GetByID handler
func TestGetByID(t *testing.T) {
	t.Run("Success - returns menu item by ID", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.GET("/menus/:id", handler.GetByID)

		expectedItem := &domain.MenuItem{ID: 1, Name: "Menu 1", Position: 0}
		mockService.On("GetByID", mock.Anything, 1).Return(expectedItem, nil)

		// Execute
		req, _ := http.NewRequest("GET", "/menus/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var result domain.MenuItem
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Menu 1", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - invalid ID parameter", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.GET("/menus/:id", handler.GetByID)

		// Execute
		req, _ := http.NewRequest("GET", "/menus/invalid", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var result domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "INVALID_ID", result.Code)
	})

	t.Run("Error - menu item not found", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.GET("/menus/:id", handler.GetByID)

		mockService.On("GetByID", mock.Anything, 999).Return(nil, errors.New("menu item with ID 999 does not exist"))

		// Execute
		req, _ := http.NewRequest("GET", "/menus/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)

		var result domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "NOT_FOUND", result.Code)

		mockService.AssertExpectations(t)
	})
}

// TestCreate tests the Create handler
func TestCreate(t *testing.T) {
	t.Run("Success - creates menu item", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.POST("/menus", handler.Create)

		createReq := domain.CreateMenuRequest{
			Name: "New Menu",
		}

		expectedItem := &domain.MenuItem{ID: 1, Name: "New Menu", Position: 0}
		mockService.On("Create", mock.Anything, createReq).Return(expectedItem, nil)

		// Execute
		body, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/menus", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)

		var result domain.MenuItem
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "New Menu", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - validation error (empty name)", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.POST("/menus", handler.Create)

		createReq := domain.CreateMenuRequest{
			Name: "", // Empty name should fail validation
		}

		// Execute
		body, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/menus", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)

		var result domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", result.Code)
	})
}

// TestUpdate tests the Update handler
func TestUpdate(t *testing.T) {
	t.Run("Success - updates menu item", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.PUT("/menus/:id", handler.Update)

		updateReq := domain.UpdateMenuRequest{
			Name: "Updated Menu",
		}

		expectedItem := &domain.MenuItem{ID: 1, Name: "Updated Menu", Position: 0}
		mockService.On("Update", mock.Anything, 1, updateReq).Return(expectedItem, nil)

		// Execute
		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", "/menus/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var result domain.MenuItem
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Menu", result.Name)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - menu item not found", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.PUT("/menus/:id", handler.Update)

		updateReq := domain.UpdateMenuRequest{
			Name: "Updated Menu",
		}

		mockService.On("Update", mock.Anything, 999, updateReq).Return(nil, errors.New("menu item with ID 999 does not exist"))

		// Execute
		body, _ := json.Marshal(updateReq)
		req, _ := http.NewRequest("PUT", "/menus/999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)

		var result domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "NOT_FOUND", result.Code)

		mockService.AssertExpectations(t)
	})
}

// TestDelete tests the Delete handler
func TestDelete(t *testing.T) {
	t.Run("Success - deletes menu item", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.DELETE("/menus/:id", handler.Delete)

		mockService.On("Delete", mock.Anything, 1).Return(1, nil)

		// Execute
		req, _ := http.NewRequest("DELETE", "/menus/1", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var result domain.DeleteResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 1, result.DeletedCount)
		assert.Contains(t, result.Message, "deleted successfully")

		mockService.AssertExpectations(t)
	})

	t.Run("Error - menu item not found", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.DELETE("/menus/:id", handler.Delete)

		mockService.On("Delete", mock.Anything, 999).Return(0, errors.New("menu item with ID 999 does not exist"))

		// Execute
		req, _ := http.NewRequest("DELETE", "/menus/999", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusNotFound, w.Code)

		var result domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "NOT_FOUND", result.Code)

		mockService.AssertExpectations(t)
	})
}

// TestReorder tests the Reorder handler
func TestReorder(t *testing.T) {
	t.Run("Success - reorders menu item", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.PUT("/menus/:id/reorder", handler.Reorder)

		reorderReq := domain.ReorderRequest{
			NewPosition: 2,
		}

		expectedItems := []domain.MenuItem{
			{ID: 1, Position: 0},
			{ID: 2, Position: 1},
			{ID: 3, Position: 2},
		}

		mockService.On("Reorder", mock.Anything, 1, 2).Return(expectedItems, nil)

		// Execute
		body, _ := json.Marshal(reorderReq)
		req, _ := http.NewRequest("PUT", "/menus/1/reorder", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var result []domain.MenuItem
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 3)

		mockService.AssertExpectations(t)
	})
}

// TestMove tests the Move handler
func TestMove(t *testing.T) {
	t.Run("Success - moves menu item", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.PUT("/menus/:id/move", handler.Move)

		newParentID := 2
		moveReq := domain.MoveRequest{
			NewParentID: &newParentID,
		}

		expectedItems := []domain.MenuItem{
			{ID: 1, ParentID: &newParentID, Position: 0},
		}

		mockService.On("Move", mock.Anything, 1, &newParentID, (*int)(nil)).Return(expectedItems, nil)

		// Execute
		body, _ := json.Marshal(moveReq)
		req, _ := http.NewRequest("PUT", "/menus/1/move", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusOK, w.Code)

		var result []domain.MenuItem
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Len(t, result, 1)

		mockService.AssertExpectations(t)
	})

	t.Run("Error - circular reference detected", func(t *testing.T) {
		// Setup
		mockService := new(MockMenuService)
		handler := NewMenuHandler(mockService)
		router := setupTestRouter(handler)
		router.PUT("/menus/:id/move", handler.Move)

		newParentID := 2
		moveReq := domain.MoveRequest{
			NewParentID: &newParentID,
		}

		mockService.On("Move", mock.Anything, 1, &newParentID, (*int)(nil)).Return(nil, errors.New("cannot move menu item: circular reference detected"))

		// Execute
		body, _ := json.Marshal(moveReq)
		req, _ := http.NewRequest("PUT", "/menus/1/move", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusConflict, w.Code)

		var result domain.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "CIRCULAR_REFERENCE", result.Code)

		mockService.AssertExpectations(t)
	})
}
