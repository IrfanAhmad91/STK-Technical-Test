@echo off
REM Script to run repository tests for hierarchical menu tree system

echo ========================================
echo Hierarchical Menu Tree - Repository Tests
echo ========================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: Go is not installed or not in PATH
    echo Please install Go 1.25.0 or later from https://golang.org/dl/
    exit /b 1
)

echo Go version:
go version
echo.

REM Parse command line arguments
set TEST_TYPE=%1
if "%TEST_TYPE%"=="" set TEST_TYPE=unit

if "%TEST_TYPE%"=="unit" goto run_unit_tests
if "%TEST_TYPE%"=="integration" goto run_integration_tests
if "%TEST_TYPE%"=="all" goto run_all_tests
if "%TEST_TYPE%"=="coverage" goto run_coverage

echo Invalid test type: %TEST_TYPE%
echo Usage: run_tests.bat [unit^|integration^|all^|coverage]
echo.
echo   unit        - Run unit tests only (default)
echo   integration - Run integration tests (requires PostgreSQL)
echo   all         - Run both unit and integration tests
echo   coverage    - Run unit tests with coverage report
exit /b 1

:run_unit_tests
echo Running unit tests...
echo.
go test -v ./internal/repository
goto end

:run_integration_tests
echo Running integration tests...
echo NOTE: This requires a running PostgreSQL database
echo.
echo Database configuration (set these environment variables to override):
echo   DB_HOST=%DB_HOST% (default: localhost)
echo   DB_PORT=%DB_PORT% (default: 5432)
echo   DB_USER=%DB_USER% (default: postgres)
echo   DB_PASSWORD=****** (default: postgres)
echo   DB_NAME=%DB_NAME% (default: menu_tree_test)
echo.
set RUN_INTEGRATION_TESTS=true
go test -v -tags=integration ./internal/repository
goto end

:run_all_tests
echo Running all tests (unit + integration)...
echo.
echo [1/2] Running unit tests...
go test -v ./internal/repository
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Unit tests failed!
    exit /b 1
)
echo.
echo [2/2] Running integration tests...
set RUN_INTEGRATION_TESTS=true
go test -v -tags=integration ./internal/repository
goto end

:run_coverage
echo Running unit tests with coverage...
echo.
go test -v -coverprofile=coverage.out ./internal/repository
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Tests failed!
    exit /b 1
)
echo.
echo Coverage report generated: coverage.out
echo Opening coverage report in browser...
go tool cover -html=coverage.out
goto end

:end
echo.
echo ========================================
echo Tests completed
echo ========================================
