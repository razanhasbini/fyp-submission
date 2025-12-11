# ==============================================================================
# Run Backend Directly (No Docker) - PowerShell Script
# ==============================================================================
# This script runs the Go backend directly on your machine without Docker
# Make sure you have Go installed: https://go.dev/dl/

Write-Host "🚀 Starting Interview AI Backend (Direct Mode - No Docker)" -ForegroundColor Green
Write-Host ""

# Check if Go is installed
$goVersion = go version 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ ERROR: Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Press any key to exit..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    exit 1
}

Write-Host "✅ Go found: $goVersion" -ForegroundColor Green
Write-Host ""

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "⚠️  WARNING: .env file not found!" -ForegroundColor Yellow
    Write-Host "Creating .env from env.example..." -ForegroundColor Yellow
    
    if (Test-Path "env.example") {
        Copy-Item "env.example" ".env"
        Write-Host "✅ Created .env file. Please edit it with your credentials!" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Press any key to open .env file in notepad..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        notepad .env
        Write-Host ""
        Write-Host "After saving .env, run this script again."
        Write-Host "Press any key to exit..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        exit 1
    } else {
        Write-Host "❌ ERROR: env.example not found!" -ForegroundColor Red
        Write-Host "Please create .env file manually with required variables." -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Press any key to exit..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        exit 1
    }
}

Write-Host "✅ .env file found" -ForegroundColor Green
Write-Host ""

# Install/update dependencies
Write-Host "📦 Installing/updating Go dependencies..." -ForegroundColor Cyan
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ ERROR: Failed to download dependencies" -ForegroundColor Red
    Write-Host "Press any key to exit..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    exit 1
}
Write-Host "✅ Dependencies ready" -ForegroundColor Green
Write-Host ""

# Display configuration
Write-Host "📋 Configuration:" -ForegroundColor Cyan
Write-Host "   Port: 8089" -ForegroundColor White
Write-Host "   Backend URL: http://localhost:8089" -ForegroundColor White
Write-Host "   Android App should connect to: http://192.168.18.5:8089" -ForegroundColor White
Write-Host ""

# Run the backend
Write-Host "🎯 Starting backend server..." -ForegroundColor Green
Write-Host "   Press Ctrl+C to stop the server" -ForegroundColor Yellow
Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
Write-Host ""

go run main.go cv_parser.go

# If we get here, the server stopped
Write-Host ""
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
Write-Host "Backend stopped." -ForegroundColor Yellow
Write-Host ""
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")


