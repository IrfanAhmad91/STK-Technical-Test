#!/bin/bash

# Script to run repository tests for hierarchical menu tree system

set -e

echo "========================================"
echo "Hierarchical Menu Tree - Repository Tests"
echo "========================================"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "ERROR: Go is not installed or not in PATH"
    echo "Please install Go 1.25.0 or later from https://golang.org/dl/"
    exit 1
fi

echo "Go version:"
go version
echo ""

# Parse command line arguments
TEST_TYPE=${1:-unit}

run_unit_tests() {
    echo "Running unit tests..."
    echo ""
    go test -v ./internal/repository
}

run_integration_tests() {
    echo "Running integration tests..."
    echo "NOTE: This requires a running PostgreSQL database"
    echo ""
    echo "Database configuration (set these environment variables to override):"
    echo "  DB_HOST=${DB_HOST:-localhost}"
    echo "  DB_PORT=${DB_PORT:-5432}"
    echo "  DB_USER=${DB_USER:-postgres}"
    echo "  DB_PASSWORD=****** (default: postgres)"
    echo "  DB_NAME=${DB_NAME:-menu_tree_test}"
    echo ""
    RUN_INTEGRATION_TESTS=true go test -v -tags=integration ./internal/repository
}

run_all_tests() {
    echo "Running all tests (unit + integration)..."
    echo ""
    echo "[1/2] Running unit tests..."
    go test -v ./internal/repository
    
    echo ""
    echo "[2/2] Running integration tests..."
    RUN_INTEGRATION_TESTS=true go test -v -tags=integration ./internal/repository
}

run_coverage() {
    echo "Running unit tests with coverage..."
    echo ""
    go test -v -coverprofile=coverage.out ./internal/repository
    
    echo ""
    echo "Coverage report generated: coverage.out"
    echo "Opening coverage report in browser..."
    go tool cover -html=coverage.out
}

case "$TEST_TYPE" in
    unit)
        run_unit_tests
        ;;
    integration)
        run_integration_tests
        ;;
    all)
        run_all_tests
        ;;
    coverage)
        run_coverage
        ;;
    *)
        echo "Invalid test type: $TEST_TYPE"
        echo "Usage: $0 [unit|integration|all|coverage]"
        echo ""
        echo "  unit        - Run unit tests only (default)"
        echo "  integration - Run integration tests (requires PostgreSQL)"
        echo "  all         - Run both unit and integration tests"
        echo "  coverage    - Run unit tests with coverage report"
        exit 1
        ;;
esac

echo ""
echo "========================================"
echo "Tests completed"
echo "========================================"
