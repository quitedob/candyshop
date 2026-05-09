@echo off
echo Starting CandyPro OEM Backend...
cd /d "%~dp0backend"
go run cmd/api/main.go
pause
