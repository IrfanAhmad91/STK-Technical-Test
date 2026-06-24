package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stk/menu-tree-api/internal/domain"
	"github.com/stk/menu-tree-api/internal/service"
)

// MenuHandler handles HTTP requests for menu operations
type MenuHandler struct {
	service service.MenuService
}

// NewMenuHandler creates a new MenuHandler instance
func NewMenuHandler(service service.MenuService) *MenuHandler {
	return &MenuHandler{
		service: service,
	}
}

// GetAll godoc
// @Summary Get all menu items
// @Description Retrieves all menu items with their hierarchical relationships
// @Tags menus
// @Accept json
// @Produce json
// @Success 200 {array} domain.MenuItem
// @Failure 500 {object} domain.ErrorResponse
// @Router /menus [get]
func (h *MenuHandler) GetAll(c *gin.Context) {
	// Call service.GetAll (Requirement 1.3)
	items, err := h.service.GetAll(c.Request.Context())
	if err != nil {
		// Return 500 for database errors (Requirement 5.5, 6.2, 6.4)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Return JSON with 200 (Requirement 1.3, 5.3, 5.5, 6.2, 6.4)
	c.JSON(http.StatusOK, items)
}

// GetByID godoc
// @Summary Get menu item by ID
// @Description Retrieves a specific menu item with its parent and children references
// @Tags menus
// @Accept json
// @Produce json
// @Param id path int true "Menu Item ID"
// @Success 200 {object} domain.MenuItem
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /menus/{id} [get]
func (h *MenuHandler) GetByID(c *gin.Context) {
	// Parse ID from URL parameter (Requirement 1.4)
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_ID",
			Message: err.Error(),
		})
		return
	}

	// Call service.GetByID (Requirement 1.4)
	item, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		// Check if it's a not found error (Requirement 5.3, 5.5, 6.2, 6.4)
		if isNotFoundError(err) {
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			})
			return
		}
		// Return 500 for other errors (Requirement 5.5, 6.2, 6.4)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Return JSON with 200 (Requirement 1.4, 5.3, 5.5, 6.2, 6.4)
	c.JSON(http.StatusOK, item)
}

// Create godoc
// @Summary Create menu item
// @Description Creates a new menu item with optional parent and position
// @Tags menus
// @Accept json
// @Produce json
// @Param request body domain.CreateMenuRequest true "Menu item data"
// @Success 201 {object} domain.MenuItem
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /menus [post]
func (h *MenuHandler) Create(c *gin.Context) {
	// Parse CreateMenuRequest from JSON body (Requirement 1.1, 5.1, 5.2)
	var req domain.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Validate request using Gin binding (Requirement 5.1, 5.2, 5.5, 6.2, 6.4, 15.1)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Call service.Create (Requirement 1.1, 1.2, 5.1, 5.2)
	item, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		// Check if it's a validation error or database error
		if isValidationError(err) {
			// Return 400 for validation errors (Requirement 5.1, 5.2, 5.5, 6.2, 6.4, 15.1)
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			})
			return
		}
		// Return 500 for database errors (Requirement 5.5, 6.2, 6.4)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Return created MenuItem with 201 (Requirement 1.1, 1.2, 5.1, 5.2, 5.5, 6.2, 6.4, 15.1)
	c.JSON(http.StatusCreated, item)
}

// Update godoc
// @Summary Update menu item
// @Description Updates an existing menu item's name
// @Tags menus
// @Accept json
// @Produce json
// @Param id path int true "Menu Item ID"
// @Param request body domain.UpdateMenuRequest true "Updated menu item data"
// @Success 200 {object} domain.MenuItem
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /menus/{id} [put]
func (h *MenuHandler) Update(c *gin.Context) {
	// Parse ID from URL parameter (Requirement 1.5)
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_ID",
			Message: err.Error(),
		})
		return
	}

	// Parse UpdateMenuRequest from body (Requirement 1.5, 5.1)
	var req domain.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Return 400 for validation errors (Requirement 5.1, 5.5, 6.2, 6.4, 15.1, 15.2)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Call service.Update (Requirement 1.5, 5.1, 5.3)
	item, err := h.service.Update(c.Request.Context(), id, req)
	if err != nil {
		// Check error type and return appropriate status code
		if isNotFoundError(err) {
			// Return 404 for not found errors (Requirement 5.3, 5.5, 6.2, 6.4, 15.2)
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			})
			return
		}
		if isValidationError(err) {
			// Return 400 for validation errors (Requirement 5.1, 5.5, 6.2, 6.4, 15.1)
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			})
			return
		}
		// Return 500 for database errors (Requirement 5.5, 6.2, 6.4)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Return updated MenuItem with 200 (Requirement 1.5, 5.1, 5.3, 5.5, 6.2, 6.4, 15.1, 15.2)
	c.JSON(http.StatusOK, item)
}

// Delete godoc
// @Summary Delete menu item
// @Description Deletes a menu item and all its descendants
// @Tags menus
// @Accept json
// @Produce json
// @Param id path int true "Menu Item ID"
// @Success 200 {object} domain.DeleteResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /menus/{id} [delete]
func (h *MenuHandler) Delete(c *gin.Context) {
	// Parse ID from URL parameter (Requirement 1.6)
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_ID",
			Message: err.Error(),
		})
		return
	}

	// Call service.Delete (Requirement 1.6, 1.7)
	deletedCount, err := h.service.Delete(c.Request.Context(), id)
	if err != nil {
		// Check error type and return appropriate status code
		if isNotFoundError(err) {
			// Return 404 for not found errors (Requirement 5.3, 5.5, 6.2, 6.4, 15.2, 15.3)
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			})
			return
		}
		// Return 500 for database errors (Requirement 5.5, 6.2, 6.4)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Build response message
	var message string
	if deletedCount == 1 {
		message = "Menu item deleted successfully"
	} else {
		message = fmt.Sprintf("Menu item and %d descendants deleted successfully", deletedCount-1)
	}

	// Return DeleteResponse with 200 (Requirement 1.6, 1.7, 5.3, 5.5, 6.2, 6.4, 15.2, 15.3)
	c.JSON(http.StatusOK, domain.DeleteResponse{
		DeletedCount: deletedCount,
		Message:      message,
	})
}

// Reorder godoc
// @Summary Reorder menu item
// @Description Changes the position of a menu item within its sibling group
// @Tags menus
// @Accept json
// @Produce json
// @Param id path int true "Menu Item ID"
// @Param request body domain.ReorderRequest true "New position data"
// @Success 200 {array} domain.MenuItem
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /menus/{id}/reorder [put]
func (h *MenuHandler) Reorder(c *gin.Context) {
	// Parse ID from URL parameter (Requirement 3.1)
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_ID",
			Message: err.Error(),
		})
		return
	}

	// Parse ReorderRequest from body (Requirement 3.1, 5.4)
	var req domain.ReorderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Return 400 for validation errors (Requirement 5.4, 5.5, 6.2, 6.4, 15.1, 15.2)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Call service.Reorder (Requirement 3.1, 3.2, 3.3, 3.4, 3.5)
	affectedItems, err := h.service.Reorder(c.Request.Context(), id, req.NewPosition)
	if err != nil {
		// Check error type and return appropriate status code
		if isNotFoundError(err) {
			// Return 404 for not found errors (Requirement 5.3, 5.5, 6.2, 6.4, 15.2)
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			})
			return
		}
		if isValidationError(err) {
			// Return 400 for validation errors (Requirement 3.4, 5.4, 5.5, 6.2, 6.4, 15.1)
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			})
			return
		}
		// Return 500 for database errors (Requirement 5.5, 6.2, 6.4)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Return affected items with 200 (Requirement 3.1, 3.2, 3.3, 3.4, 3.5, 5.4, 6.2, 6.4, 15.1, 15.2)
	c.JSON(http.StatusOK, affectedItems)
}

// Move godoc
// @Summary Move menu item
// @Description Moves a menu item to a different parent with optional position
// @Tags menus
// @Accept json
// @Produce json
// @Param id path int true "Menu Item ID"
// @Param request body domain.MoveRequest true "New parent and position data"
// @Success 200 {array} domain.MenuItem
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 409 {object} domain.ErrorResponse "Circular reference detected"
// @Failure 500 {object} domain.ErrorResponse
// @Router /menus/{id}/move [put]
func (h *MenuHandler) Move(c *gin.Context) {
	// Parse ID from URL parameter (Requirement 4.1)
	id, err := parseIDParam(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "INVALID_ID",
			Message: err.Error(),
		})
		return
	}

	// Parse MoveRequest from body (Requirement 4.1, 5.2)
	var req domain.MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Return 400 for validation errors (Requirement 5.2, 5.5, 6.2, 6.4, 15.1, 15.2)
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Call service.Move (Requirement 4.1, 4.2, 4.3, 4.4, 4.5)
	affectedItems, err := h.service.Move(c.Request.Context(), id, req.NewParentID, req.Position)
	if err != nil {
		// Check error type and return appropriate status code
		if isNotFoundError(err) {
			// Return 404 for not found errors (Requirement 5.3, 5.5, 6.2, 6.4, 15.2)
			c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: err.Error(),
			})
			return
		}
		if isCircularReferenceError(err) {
			// Return 409 for circular reference errors (Requirement 4.4, 5.5, 6.2, 6.4, 15.4)
			c.JSON(http.StatusConflict, domain.ErrorResponse{
				Code:    "CIRCULAR_REFERENCE",
				Message: err.Error(),
			})
			return
		}
		if isValidationError(err) {
			// Return 400 for validation errors (Requirement 5.2, 5.5, 6.2, 6.4, 15.1)
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			})
			return
		}
		// Return 500 for database errors (Requirement 5.5, 6.2, 6.4)
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Code:    "DATABASE_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Return affected items with 200 (Requirement 4.1, 4.2, 4.3, 4.4, 4.5, 5.2, 6.2, 6.4, 15.1, 15.2, 15.4)
	c.JSON(http.StatusOK, affectedItems)
}

// Helper functions

// parseIDParam extracts and validates the ID parameter from URL
func parseIDParam(c *gin.Context) (int, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errors.New("invalid ID parameter: must be a positive integer")
	}
	if id <= 0 {
		return 0, errors.New("invalid ID parameter: must be a positive integer")
	}
	return id, nil
}

// isNotFoundError checks if error indicates resource not found
func isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "does not exist") || contains(msg, "not found")
}

// isValidationError checks if error is a validation error
func isValidationError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "cannot be empty") ||
		contains(msg, "cannot exceed") ||
		contains(msg, "must be non-negative") ||
		contains(msg, "exceeds maximum") ||
		contains(msg, "invalid")
}

// isCircularReferenceError checks if error indicates circular reference
func isCircularReferenceError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "circular reference")
}

// contains checks if string contains substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
