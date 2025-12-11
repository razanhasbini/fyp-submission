@echo off
REM ==============================================================================
REM Run Backend Directly (No Docker) - Batch Script
REM ==============================================================================
REM This script runs the Go backend directly on your machine without Docker
REM Make sure you have Go installed: https://go.dev/dl/

echo.
echo 🚀 Starting Interview AI Backend (Direct Mode - No Docker)
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ ERROR: Go is not installed or not in PATH
    echo Please install Go from: https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

for /f "tokens=*" %%i in ('go version') do set GOVERSION=%%i
echo ✅ Go found: %GOVERSION%
echo.

REM Check if .env file exists
if not exist ".env" (
    echo ⚠️  WARNING: .env file not found!
    echo Creating .env from env.example...
    echo.
    
    if exist "env.example" (
        copy "env.example" ".env" >nul
        echo ✅ Created .env file. Please edit it with your credentials!
        echo.
        echo Press any key to open .env file in notepad...
        pause >nul
        notepad .env
        echo.
        echo After saving .env, run this script again.
        pause
        exit /b 1
    ) else (
        echo ❌ ERROR: env.example not found!
        echo Please create .env file manually with required variables.
        echo.
        pause
        exit /b 1
    )
)

echo ✅ .env file found
echo.

REM Install/update dependencies
echo 📦 Installing/updating Go dependencies...
go mod download
if %errorlevel% neq 0 (
    echo ❌ ERROR: Failed to download dependencies
    pause
    exit /b 1
)
echo ✅ Dependencies ready
echo.

REM Display configuration
echo 📋 Configuration:
echo    Port: 8089
echo    Backend URL: http://localhost:8089
echo    Android App should connect to: http://192.168.18.5:8089
echo.

REM Run the backend
echo 🎯 Starting backend server...
echo    Press Ctrl+C to stop the server
echo.
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo.

go run main.go cv_parser.go

REM If we get here, the server stopped
echo.
echo ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
echo Backend stopped.
echo.
pause


