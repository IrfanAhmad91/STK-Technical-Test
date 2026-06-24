#!/bin/bash

# Database initialization script for Hierarchical Menu Tree System
# This script creates the database and runs migrations

set -e  # Exit on any error

# Configuration (can be overridden by environment variables)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-menu_system_db}"
DB_PASSWORD="${DB_PASSWORD:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Database Initialization Script"
echo "========================================="
echo ""

# Check if PostgreSQL is installed
if ! command -v psql &> /dev/null; then
    echo -e "${RED}Error: psql command not found. Please install PostgreSQL client.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ PostgreSQL client found${NC}"

# Check if we can connect to PostgreSQL server
echo "Testing connection to PostgreSQL server..."
if [ -n "$DB_PASSWORD" ]; then
    export PGPASSWORD="$DB_PASSWORD"
fi

if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c '\q' 2>/dev/null; then
    echo -e "${RED}Error: Cannot connect to PostgreSQL server at $DB_HOST:$DB_PORT${NC}"
    echo "Please check your connection settings and ensure PostgreSQL is running."
    exit 1
fi

echo -e "${GREEN}✓ Connected to PostgreSQL server${NC}"
echo ""

# Check if database already exists
echo "Checking if database '$DB_NAME' exists..."
DB_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'")

if [ "$DB_EXISTS" = "1" ]; then
    echo -e "${YELLOW}⚠ Database '$DB_NAME' already exists${NC}"
    read -p "Do you want to drop and recreate it? (yes/no): " CONFIRM
    if [ "$CONFIRM" = "yes" ]; then
        echo "Dropping existing database..."
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "DROP DATABASE $DB_NAME;"
        echo -e "${GREEN}✓ Database dropped${NC}"
    else
        echo "Skipping database creation. Will run migrations on existing database."
        echo ""
    fi
fi

# Create database if it doesn't exist
if [ "$DB_EXISTS" != "1" ] || [ "$CONFIRM" = "yes" ]; then
    echo "Creating database '$DB_NAME'..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;"
    echo -e "${GREEN}✓ Database '$DB_NAME' created${NC}"
    echo ""
fi

# Run migrations
echo "Running migrations..."
echo "----------------------------------------"

MIGRATION_FILE="migrations/001_create_menu_items_table.sql"

if [ ! -f "$MIGRATION_FILE" ]; then
    echo -e "${RED}Error: Migration file not found: $MIGRATION_FILE${NC}"
    exit 1
fi

echo "Applying migration: $MIGRATION_FILE"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_FILE"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Migration applied successfully${NC}"
else
    echo -e "${RED}✗ Migration failed${NC}"
    exit 1
fi

echo ""
echo "========================================="
echo "Verifying Schema"
echo "========================================="
echo ""

# Run verification script
VERIFY_SCRIPT="migrations/verify_schema.sql"
if [ -f "$VERIFY_SCRIPT" ]; then
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$VERIFY_SCRIPT"
else
    echo -e "${YELLOW}⚠ Verification script not found: $VERIFY_SCRIPT${NC}"
fi

echo ""
echo "========================================="
echo "Setup Complete!"
echo "========================================="
echo ""
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "User: $DB_USER"
echo ""
echo "You can now:"
echo "  1. Run tests: psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f migrations/test_schema.sql"
echo "  2. Connect: psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
echo "  3. Start developing your Go backend API"
echo ""
