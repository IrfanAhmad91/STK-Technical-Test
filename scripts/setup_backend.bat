@echo off
REM Setup script for Windows

echo ====================================
echo Hierarchical Menu Tree - Backend Setup
echo ====================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go is not installed or not in PATH
    echo Please install Go from https://go.dev/dl
    echo After installation, restart your terminal and run this script again
    pause
    exit /b 1
)

echo [1/4] Checking Go installation...
go version
echo.

echo [2/4] Installing Go dependencies...
go mod tidy
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to install dependencies
    pause
    exit /b 1
)
echo Dependencies installed successfully
echo.

echo [3/4] Installing Swag CLI for API documentation...
go install github.com/swaggo/swag/cmd/swag@latest
if %ERRORLEVEL% NEQ 0 (
    echo [WARNING] Failed to install swag CLI
    echo You can install it later with: go install github.com/swaggo/swag/cmd/swag@latest
)
echo.

echo [4/4] Checking database connection...
echo Make sure PostgreSQL is running and database 'menu_tree_db' is created
echo Run scripts\init_database.bat if you haven't set up the database yet
echo.

echo ====================================
echo Setup completed successfully!
echo ====================================
echo.
echo Next steps:
echo 1. Copy .env.example to .env and configure if needed
echo 2. Ensure database is set up (run scripts\init_database.bat)
echo 3. Run the application: go run cmd\api\main.go
echo.

pause
