@echo off
echo ============================================
echo   CandyPro Frontend Deployment Packager
echo ============================================
echo.

cd /d "%~dp0frontend"

echo [1/3] Cleaning previous build...
if exist ".output" rmdir /s /q ".output"

echo [2/3] Building frontend (backend must be running on :8080)...
set BACKEND_URL=http://127.0.0.1:8080
set INTERNAL_API_BASE=http://127.0.0.1:8080/api/v1
call npx nuxi build
if %ERRORLEVEL% neq 0 (
    echo [FAIL] Build failed
    pause
    exit /b 1
)

echo [3/3] Packaging for deployment...
cd "%~dp0"
set PACKAGE_NAME=frontend-deploy-%date:~0,4%%date:~5,2%%date:~8,2%-%time:~0,2%%time:~3,2%%time:~6,2%.tar.gz
set PACKAGE_NAME=%PACKAGE_NAME: =0%
cd frontend\.output
tar -czf "..\..\%PACKAGE_NAME%" .
cd "..\.."

echo.
echo ============================================
echo   Package created: %PACKAGE_NAME%
echo ============================================
echo.
echo Upload to server and extract:
echo   scp %PACKAGE_NAME% user@117.50.33.136:/opt/candypro/
echo   ssh user@117.50.33.136
echo   cd /opt/candypro/frontend ^&^& tar -xzf ../%PACKAGE_NAME%
echo   node server/index.mjs
echo.
pause
