-- Migration Rollback: 001_create_menu_items_table
-- Description: Rollback the menu_items table creation
-- Author: Generated from spec design
-- Date: 2024

-- ============================================================================
-- Rollback Subtask 1.2: Drop trigger and function
-- ============================================================================

-- Drop the trigger first (depends on the function)
DROP TRIGGER IF EXISTS update_menu_items_updated_at ON menu_items;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- ============================================================================
-- Rollback Subtask 1.1: Drop table and indexes
-- ============================================================================

-- Drop the menu_items table (this will also drop all indexes and constraints)
DROP TABLE IF EXISTS menu_items CASCADE;

-- End of rollback migration
