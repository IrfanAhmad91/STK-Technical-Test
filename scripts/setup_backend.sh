#!/bin/bash

# Setup script for Linux/macOS

echo "===================================="
echo "Hierarchical Menu Tree - Backend Setup"
echo "===================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "[ERROR] Go is not installed or not in PATH"
    echo "Please install Go from https://go.dev/dl"
    echo "After installation, restart your terminal and run this script again"
    exit 1
fi

echo "[1/4] Checking Go installation..."
go version
echo ""

echo "[2/4] Installing Go dependencies..."
if ! go mod tidy; then
    echo "[ERROR] Failed to install dependencies"
    exit 1
fi
echo "Dependencies installed successfully"
echo ""

echo "[3/4] Installing Swag CLI for API documentation..."
if ! go install github.com/swaggo/swag/cmd/swag@latest; then
    echo "[WARNING] Failed to install swag CLI"
    echo "You can install it later with: go install github.com/swaggo/swag/cmd/swag@latest"
fi
echo ""

echo "[4/4] Checking database connection..."
echo "Make sure PostgreSQL is running and database 'menu_tree_db' is created"
echo "Run scripts/init_database.sh if you haven't set up the database yet"
echo ""

echo "===================================="
echo "Setup completed successfully!"
echo "===================================="
echo ""
echo "Next steps:"
echo "1. Copy .env.example to .env and configure if needed"
echo "2. Ensure database is set up (run scripts/init_database.sh)"
echo "3. Run the application: go run cmd/api/main.go"
echo ""
