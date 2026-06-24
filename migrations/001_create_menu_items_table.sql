-- Migration: 001_create_menu_items_table
-- Description: Create the menu_items table with adjacency list schema for hierarchical menu structure
-- Author: Generated from spec design
-- Date: 2024
-- Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6

-- ============================================================================
-- Subtask 1.1: Create PostgreSQL database and menu_items table
-- ============================================================================

-- Create menu_items table with adjacency list model
CREATE TABLE menu_items (
    -- Primary key: auto-incrementing integer ID
    id SERIAL PRIMARY KEY,
    
    -- Menu item name: required, 1-255 characters
    name VARCHAR(255) NOT NULL,
    
    -- Parent reference: nullable for root-level items
    parent_id INTEGER NULL,
    
    -- Position within sibling group: non-negative integer, defaults to 0
    position INTEGER NOT NULL DEFAULT 0,
    
    -- Audit timestamps: track creation and last update
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key constraint with CASCADE DELETE
    -- When a parent is deleted, all its children are automatically deleted
    CONSTRAINT fk_parent FOREIGN KEY (parent_id) 
        REFERENCES menu_items(id) 
        ON DELETE CASCADE,
    
    -- CHECK constraint: name must be non-empty and within length limit
    CONSTRAINT chk_name_length CHECK (LENGTH(name) > 0 AND LENGTH(name) <= 255),
    
    -- CHECK constraint: position must be non-negative
    CONSTRAINT chk_position_non_negative CHECK (position >= 0)
);

-- ============================================================================
-- Indexes for query optimization
-- ============================================================================

-- Index on parent_id: optimizes queries that filter by parent
-- Used for: fetching children of a specific parent, hierarchical queries
CREATE INDEX idx_parent_id ON menu_items(parent_id);

-- Index on position: optimizes queries that order by position
-- Used for: sorting siblings, finding max position
CREATE INDEX idx_position ON menu_items(position);

-- Composite index on (parent_id, position): optimizes the common query pattern
-- Used for: fetching siblings in order (SELECT * FROM menu_items WHERE parent_id = ? ORDER BY position)
CREATE INDEX idx_parent_position ON menu_items(parent_id, position);

-- ============================================================================
-- Subtask 1.2: Implement database trigger for updated_at timestamp
-- ============================================================================

-- Function: update_updated_at_column()
-- Purpose: Automatically set updated_at to current timestamp on UPDATE operations
-- Language: PL/pgSQL (PostgreSQL's procedural language)
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    -- Set the updated_at column to the current timestamp
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger: update_menu_items_updated_at
-- Purpose: Call update_updated_at_column() before every UPDATE on menu_items
-- Timing: BEFORE UPDATE ensures updated_at is set before the row is written
CREATE TRIGGER update_menu_items_updated_at
    BEFORE UPDATE ON menu_items
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- Schema Design Rationale
-- ============================================================================
-- 
-- ADJACENCY LIST MODEL:
-- - Chosen over nested sets for frequent write operations (CRUD, reordering)
-- - Provides O(1) updates for parent changes and position adjustments
-- - Nested sets would require O(n) updates for insertions
-- 
-- SERIAL PRIMARY KEY:
-- - PostgreSQL auto-incrementing integer type for unique ID generation
-- - Efficient for indexing and foreign key relationships
-- 
-- FOREIGN KEY WITH CASCADE DELETE:
-- - Automatically removes descendant nodes when parent is deleted
-- - Ensures referential integrity without complex application logic
-- - Prevents orphaned records in the database
-- 
-- COMPOSITE INDEX (parent_id, position):
-- - Optimizes the common query: fetching siblings in order
-- - Covers both WHERE parent_id = ? and ORDER BY position
-- - Reduces index lookup overhead for hierarchical queries
-- 
-- TRIGGER FOR updated_at:
-- - PostgreSQL doesn't support MySQL's ON UPDATE CURRENT_TIMESTAMP
-- - Trigger function provides equivalent functionality
-- - Ensures updated_at is always current without application-level logic
-- 
-- TIMESTAMPS (created_at, updated_at):
-- - Enable audit trails for data changes
-- - Support potential optimistic locking for concurrent updates
-- - Useful for debugging and data analysis
-- ============================================================================

-- End of migration
