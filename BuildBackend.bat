@echo off
echo Building backend for Linux (amd64)...
cd /d "%~dp0backend"
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0
go build -o ../candypro-server ./cmd/api/...
if %ERRORLEVEL% equ 0 (
    echo [OK] Build successful: candypro-server
) else (
    echo [FAIL] Build failed
    pause
    exit /b 1
)
