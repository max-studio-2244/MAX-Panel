@echo off
echo 🚀 MAX Panel Installer (Windows)
echo ================================

REM Check if Docker is installed
docker --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker is not installed. Please install Docker Desktop first.
    echo Visit: https://docs.docker.com/desktop/windows/
    pause
    exit /b 1
)

REM Check if Docker Compose is installed
docker-compose --version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker Compose is not installed. Please install Docker Desktop first.
    pause
    exit /b 1
)

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% equ 0 (
    echo ✅ Go is installed
    for /f "tokens=3" %%i in ('go version') do echo    Version: %%i
) else (
    echo ⚠️  Go is not installed (only needed for development)
)

echo.
echo Choose installation option:
echo 1) Dependencies  - Check ^& install system requirements
echo 2) Panel        - Install main panel with domain setup
echo 3) Wings        - Install Wings daemon with node token
echo 4) Cloudflare   - Setup Cloudflare Zero Trust tunnel
echo 5) Uninstall    - Remove all components
echo 6) Exit         - Quit installer

set /p choice="Enter your choice (1-6): "

if "%choice%"=="1" (
    echo 🐳 Installing with Docker Compose...
    
    REM Create data directory
    if not exist "data" mkdir data
    
    REM Build and start services
    docker-compose up -d --build
    
    echo ✅ Game Panel is starting!
    echo 🌐 Access it at: http://localhost:8080
    echo 📊 View logs: docker-compose logs -f
    echo 🛑 Stop: docker-compose down
    
) else if "%choice%"=="2" (
    echo 🔧 Setting up for local development...
    
    go version >nul 2>&1
    if %errorlevel% neq 0 (
        echo ❌ Go is required for local development
        pause
        exit /b 1
    )
    
    REM Download dependencies
    echo 📦 Downloading dependencies...
    go mod tidy
    
    REM Build the application
    echo 🔨 Building application...
    go build -o game-panel.exe .
    
    echo ✅ Build complete!
    echo 🚀 Run with: game-panel.exe
    echo 🌐 Access at: http://localhost:8080
    
) else if "%choice%"=="3" (
    echo 📦 Building binary...
    
    go version >nul 2>&1
    if %errorlevel% neq 0 (
        echo ❌ Go is required to build binary
        pause
        exit /b 1
    )
    
    REM Download dependencies
    go mod tidy
    
    REM Build for different platforms
    echo Building for Windows...
    set GOOS=windows
    set GOARCH=amd64
    go build -o game-panel-windows.exe .
    
    echo Building for Linux...
    set GOOS=linux
    set GOARCH=amd64
    go build -o game-panel-linux .
    
    echo Building for macOS...
    set GOOS=darwin
    set GOARCH=amd64
    go build -o game-panel-macos .
    
    echo ✅ Binaries built successfully!
    echo 📁 Files: game-panel-windows.exe, game-panel-linux, game-panel-macos
    
) else (
    echo ❌ Invalid choice
    pause
    exit /b 1
)

echo.
echo 🎉 Installation complete!
echo.
echo 📚 Quick Start Guide:
echo 1. Access the web interface
echo 2. Click 'New Server' to create a game server
echo 3. Select your game type (Minecraft, CS:GO, etc.)
echo 4. Configure resources and click 'Create'
echo 5. Start your server and enjoy!
echo.
echo 🔧 Troubleshooting:
echo - Make sure Docker Desktop is running
echo - Check Windows Firewall settings for the ports you use
echo - View logs for any errors
echo.
pause