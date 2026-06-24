-- Verification Script: Database Schema Validation
-- Purpose: Verify that migration 001 was applied correctly
-- Usage: psql -U username -d database_name -f migrations/verify_schema.sql

\echo '\n========================================='
\echo 'Schema Verification for menu_items Table'
\echo '=========================================\n'

\echo '1. Checking table existence...'
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM information_schema.tables 
            WHERE table_name = 'menu_items'
        ) 
        THEN '✓ Table menu_items exists'
        ELSE '✗ Table menu_items NOT FOUND'
    END AS table_check;

\echo '\n2. Checking table structure...'
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'menu_items'
ORDER BY ordinal_position;

\echo '\n3. Checking constraints...'
SELECT
    con.conname AS constraint_name,
    con.contype AS constraint_type,
    CASE con.contype
        WHEN 'c' THEN 'CHECK'
        WHEN 'f' THEN 'FOREIGN KEY'
        WHEN 'p' THEN 'PRIMARY KEY'
        WHEN 'u' THEN 'UNIQUE'
    END AS constraint_description,
    pg_get_constraintdef(con.oid) AS constraint_definition
FROM pg_constraint con
JOIN pg_class rel ON rel.oid = con.conrelid
WHERE rel.relname = 'menu_items'
ORDER BY con.contype, con.conname;

\echo '\n4. Checking indexes...'
SELECT
    indexname AS index_name,
    indexdef AS index_definition
FROM pg_indexes
WHERE tablename = 'menu_items'
ORDER BY indexname;

\echo '\n5. Checking trigger...'
SELECT
    trigger_name,
    event_manipulation AS event,
    action_timing AS timing,
    action_statement AS action
FROM information_schema.triggers
WHERE event_object_table = 'menu_items';

\echo '\n6. Checking trigger function...'
SELECT
    routine_name AS function_name,
    routine_type AS type,
    data_type AS return_type
FROM information_schema.routines
WHERE routine_name = 'update_updated_at_column';

\echo '\n7. Summary - Expected vs Actual:'
SELECT 
    'Columns' AS component,
    6 AS expected_count,
    COUNT(*) AS actual_count,
    CASE WHEN COUNT(*) = 6 THEN '✓ PASS' ELSE '✗ FAIL' END AS status
FROM information_schema.columns
WHERE table_name = 'menu_items'
UNION ALL
SELECT 
    'Indexes',
    4,  -- 3 custom indexes + 1 primary key index
    COUNT(*),
    CASE WHEN COUNT(*) = 4 THEN '✓ PASS' ELSE '✗ FAIL' END
FROM pg_indexes
WHERE tablename = 'menu_items'
UNION ALL
SELECT 
    'Constraints',
    4,  -- 1 PK + 1 FK + 2 CHECK
    COUNT(*),
    CASE WHEN COUNT(*) = 4 THEN '✓ PASS' ELSE '✗ FAIL' END
FROM pg_constraint con
JOIN pg_class rel ON rel.oid = con.conrelid
WHERE rel.relname = 'menu_items'
UNION ALL
SELECT 
    'Triggers',
    1,
    COUNT(*),
    CASE WHEN COUNT(*) = 1 THEN '✓ PASS' ELSE '✗ FAIL' END
FROM information_schema.triggers
WHERE event_object_table = 'menu_items'
UNION ALL
SELECT 
    'Functions',
    1,
    COUNT(*),
    CASE WHEN COUNT(*) = 1 THEN '✓ PASS' ELSE '✗ FAIL' END
FROM information_schema.routines
WHERE routine_name = 'update_updated_at_column';

\echo '\n========================================='
\echo 'Verification Complete'
\echo '=========================================\n'

-- Optional: Display detailed table information
\echo 'Detailed table description:'
\d+ menu_items
