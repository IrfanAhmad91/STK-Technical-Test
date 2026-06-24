@echo off
REM Database initialization script for Hierarchical Menu Tree System (Windows)
REM This script creates the database and runs migrations

setlocal enabledelayedexpansion

REM Configuration (can be overridden by environment variables)
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=5432
if "%DB_USER%"=="" set DB_USER=postgres
if "%DB_NAME%"=="" set DB_NAME=menu_system_db

echo =========================================
echo Database Initialization Script (Windows)
echo =========================================
echo.

REM Check if psql is available
where psql >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] psql command not found. Please install PostgreSQL client.
    echo Add PostgreSQL bin directory to your PATH:
    echo Example: C:\Program Files\PostgreSQL\16\bin
    exit /b 1
)

echo [OK] PostgreSQL client found
echo.

REM Test connection
echo Testing connection to PostgreSQL server...
if not "%DB_PASSWORD%"=="" (
    set PGPASSWORD=%DB_PASSWORD%
)

psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d postgres -c "\q" >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Cannot connect to PostgreSQL server at %DB_HOST%:%DB_PORT%
    echo Please check your connection settings and ensure PostgreSQL is running.
    exit /b 1
)

echo [OK] Connected to PostgreSQL server
echo.

REM Check if database exists
echo Checking if database '%DB_NAME%' exists...
for /f "usebackq delims=" %%i in (`psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='%DB_NAME%'"`) do set DB_EXISTS=%%i

if "%DB_EXISTS%"=="1" (
    echo [WARNING] Database '%DB_NAME%' already exists
    set /p CONFIRM="Do you want to drop and recreate it? (yes/no): "
    if /i "!CONFIRM!"=="yes" (
        echo Dropping existing database...
        psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d postgres -c "DROP DATABASE %DB_NAME%;"
        echo [OK] Database dropped
        set DB_EXISTS=0
    ) else (
        echo Skipping database creation. Will run migrations on existing database.
    )
    echo.
)

REM Create database if needed
if not "%DB_EXISTS%"=="1" (
    echo Creating database '%DB_NAME%'...
    psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d postgres -c "CREATE DATABASE %DB_NAME%;"
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to create database
        exit /b 1
    )
    echo [OK] Database '%DB_NAME%' created
    echo.
)

REM Run migrations
echo Running migrations...
echo -----------------------------------------
echo.

set MIGRATION_FILE=migrations\001_create_menu_items_table.sql

if not exist "%MIGRATION_FILE%" (
    echo [ERROR] Migration file not found: %MIGRATION_FILE%
    exit /b 1
)

echo Applying migration: %MIGRATION_FILE%
psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f %MIGRATION_FILE%

if %errorlevel% neq 0 (
    echo [ERROR] Migration failed
    exit /b 1
)

echo [OK] Migration applied successfully
echo.

REM Verify schema
echo =========================================
echo Verifying Schema
echo =========================================
echo.

set VERIFY_SCRIPT=migrations\verify_schema.sql
if exist "%VERIFY_SCRIPT%" (
    psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f %VERIFY_SCRIPT%
) else (
    echo [WARNING] Verification script not found: %VERIFY_SCRIPT%
)

echo.
echo =========================================
echo Setup Complete!
echo =========================================
echo.
echo Database: %DB_NAME%
echo Host: %DB_HOST%:%DB_PORT%
echo User: %DB_USER%
echo.
echo You can now:
echo   1. Run tests: psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME% -f migrations\test_schema.sql
echo   2. Connect: psql -h %DB_HOST% -p %DB_PORT% -U %DB_USER% -d %DB_NAME%
echo   3. Start developing your Go backend API
echo.

endlocal
