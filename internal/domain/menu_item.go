package domain

import "time"

// MenuItem represents a node in the menu tree
// @Description Menu item with hierarchical relationships
type MenuItem struct {
	ID        int       `json:"id" db:"id" example:"1"`
	Name      string    `json:"name" db:"name" binding:"required,min=1,max=255" example:"Dashboard"`
	ParentID  *int      `json:"parent_id" db:"parent_id"`
	Position  int       `json:"position" db:"position" example:"0"`
	CreatedAt time.Time `json:"created_at" db:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" example:"2024-01-15T10:30:00Z"`
}

// CreateMenuRequest represents the request body for creating a menu item
// @Description Request payload for creating a new menu item
type CreateMenuRequest struct {
	Name     string `json:"name" binding:"required,min=1,max=255" example:"New Menu Item"`
	ParentID *int   `json:"parent_id"`
	Position *int   `json:"position" example:"0"`
}

// UpdateMenuRequest represents the request body for updating a menu item
// @Description Request payload for updating an existing menu item
type UpdateMenuRequest struct {
	Name string `json:"name" binding:"required,min=1,max=255" example:"Updated Menu Item"`
}

// ReorderRequest represents the request body for reordering a menu item
// @Description Request payload for reordering a menu item within its sibling group
type ReorderRequest struct {
	NewPosition int `json:"new_position" binding:"required,min=0" example:"2"`
}

// MoveRequest represents the request body for moving a menu item to a different parent
// @Description Request payload for moving a menu item to a different parent
type MoveRequest struct {
	NewParentID *int `json:"new_parent_id"`
	Position    *int `json:"position" example:"0"`
}

// ErrorResponse represents an error response
// @Description Standard error response format
type ErrorResponse struct {
	Code    string                 `json:"code" example:"VALIDATION_ERROR"`
	Message string                 `json:"message" example:"Invalid request parameters"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// DeleteResponse represents the response for a delete operation
// @Description Response payload for delete operations
type DeleteResponse struct {
	DeletedCount int    `json:"deleted_count" example:"5"`
	Message      string `json:"message" example:"Menu item and 4 descendants deleted successfully"`
}

// PositionUpdate represents a position update for batch operations
type PositionUpdate struct {
	ID       int
	Position int
}
