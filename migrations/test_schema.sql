-- Test Script: Schema Functionality Testing
-- Purpose: Test all constraints, triggers, and cascade operations
-- Usage: psql -U username -d database_name -f migrations/test_schema.sql

\echo '\n========================================='
\echo 'Schema Functionality Tests'
\echo '=========================================\n'

-- Clean up any existing test data
DELETE FROM menu_items WHERE name LIKE 'Test%';

\echo '1. Testing basic INSERT operations...'

-- Insert root-level items
INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Test Home', NULL, 0);

INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Test About', NULL, 1);

INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Test Services', NULL, 2);

\echo '✓ Root-level items inserted'

-- Insert child items
INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Test Our Team', (SELECT id FROM menu_items WHERE name = 'Test About'), 0);

INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Test Our History', (SELECT id FROM menu_items WHERE name = 'Test About'), 1);

\echo '✓ Child items inserted'

-- Insert grandchild items
INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Test Management', (SELECT id FROM menu_items WHERE name = 'Test Our Team'), 0);

\echo '✓ Grandchild items inserted'

\echo '\n2. Testing CHECK constraints...'

-- Test name length constraint (should fail)
\echo 'Testing empty name (should fail)...'
INSERT INTO menu_items (name, parent_id, position) 
VALUES ('', NULL, 3);
-- Expected: ERROR: new row violates check constraint "chk_name_length"

-- Test position constraint (should fail)
\echo '\nTesting negative position (should fail)...'
INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Test Negative', NULL, -1);
-- Expected: ERROR: new row violates check constraint "chk_position_non_negative"

\echo '\n3. Testing foreign key constraint...'

-- Test invalid parent_id (should fail)
\echo 'Testing invalid parent_id (should fail)...'
INSERT INTO menu_items (name, parent_id, position) 
VALUES ('Test Invalid Parent', 99999, 0);
-- Expected: ERROR: insert or update on table "menu_items" violates foreign key constraint "fk_parent"

\echo '\n4. Testing updated_at trigger...'

-- Get initial timestamp
SELECT id, name, created_at, updated_at 
FROM menu_items 
WHERE name = 'Test Home';

-- Wait a moment (in PostgreSQL, we use pg_sleep)
SELECT pg_sleep(1);

-- Update the item
UPDATE menu_items 
SET name = 'Test Home Updated' 
WHERE name = 'Test Home';

-- Verify updated_at changed
\echo 'Verifying updated_at timestamp changed:'
SELECT 
    name, 
    created_at, 
    updated_at,
    CASE 
        WHEN updated_at > created_at THEN '✓ Trigger working - updated_at is newer'
        ELSE '✗ Trigger failed - timestamps are equal'
    END AS trigger_status
FROM menu_items 
WHERE name = 'Test Home Updated';

\echo '\n5. Testing CASCADE DELETE...'

\echo 'Current menu structure:'
SELECT 
    mi.id,
    mi.name,
    mi.parent_id,
    COALESCE(p.name, 'NULL (root)') AS parent_name,
    mi.position
FROM menu_items mi
LEFT JOIN menu_items p ON mi.parent_id = p.id
WHERE mi.name LIKE 'Test%'
ORDER BY 
    COALESCE(mi.parent_id, 0), 
    mi.position;

-- Count items before delete
\echo '\nCounting items before deletion:'
SELECT 
    'Total items' AS category,
    COUNT(*) AS count
FROM menu_items 
WHERE name LIKE 'Test%'
UNION ALL
SELECT 
    'Items under Test About',
    COUNT(*)
FROM menu_items 
WHERE parent_id = (SELECT id FROM menu_items WHERE name = 'Test About')
    OR id IN (
        SELECT id FROM menu_items 
        WHERE parent_id IN (
            SELECT id FROM menu_items 
            WHERE parent_id = (SELECT id FROM menu_items WHERE name = 'Test About')
        )
    );

-- Delete parent (should cascade to children and grandchildren)
DELETE FROM menu_items WHERE name = 'Test About';

\echo '\n✓ Deleted "Test About" item'

-- Verify cascade delete worked
\echo '\nCounting items after deletion:'
SELECT 
    'Total items' AS category,
    COUNT(*) AS count
FROM menu_items 
WHERE name LIKE 'Test%'
UNION ALL
SELECT 
    'Orphaned children (should be 0)',
    COUNT(*)
FROM menu_items 
WHERE name LIKE 'Test Our%' OR name = 'Test Management';

\echo '\nRemaining menu structure:'
SELECT 
    mi.id,
    mi.name,
    mi.parent_id,
    COALESCE(p.name, 'NULL (root)') AS parent_name,
    mi.position
FROM menu_items mi
LEFT JOIN menu_items p ON mi.parent_id = p.id
WHERE mi.name LIKE 'Test%'
ORDER BY 
    COALESCE(mi.parent_id, 0), 
    mi.position;

\echo '\n6. Testing index usage...'
\echo 'Checking query plan for parent_id lookup (should use idx_parent_id):'
EXPLAIN SELECT * FROM menu_items WHERE parent_id = 1;

\echo '\nChecking query plan for parent + position query (should use idx_parent_position):'
EXPLAIN SELECT * FROM menu_items WHERE parent_id = 1 ORDER BY position;

\echo '\n========================================='
\echo 'Test Summary'
\echo '=========================================\n'

\echo 'Expected Results:'
\echo '- ✓ Root-level items inserted successfully'
\echo '- ✓ Child and grandchild items inserted successfully'
\echo '- ✗ Empty name insertion rejected (CHECK constraint)'
\echo '- ✗ Negative position insertion rejected (CHECK constraint)'
\echo '- ✗ Invalid parent_id insertion rejected (FOREIGN KEY constraint)'
\echo '- ✓ updated_at trigger updated timestamp on UPDATE'
\echo '- ✓ CASCADE DELETE removed all descendants'
\echo '- ✓ Indexes used in query plans'

\echo '\n========================================='
\echo 'Cleaning up test data...'

-- Clean up remaining test data
DELETE FROM menu_items WHERE name LIKE 'Test%';

\echo '✓ Test data removed'
\echo '=========================================\n'
