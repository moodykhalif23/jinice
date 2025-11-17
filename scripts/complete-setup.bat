@echo off
REM Complete setup script for Business Directory with Image System

echo ========================================
echo Business Directory - Complete Setup
echo ========================================
echo.

REM Check if Docker is running
docker ps >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker is not running. Please start Docker Desktop.
    pause
    exit /b 1
)

echo Step 1: Stopping existing containers...
docker-compose down

echo.
echo Step 2: Building application...
docker-compose build --no-cache

echo.
echo Step 3: Starting services...
docker-compose up -d

echo.
echo Step 4: Waiting for MySQL to be ready...
timeout /t 15 /nobreak >nul

echo.
echo Step 5: Running migrations...

echo   - Adding image columns to businesses and events...
type migrations\add_images.sql | docker exec -i business-directory-mysql mysql -uapp -psecure_app_password_2025! business_directory

echo   - Creating comprehensive image system...
type migrations\comprehensive_images.sql | docker exec -i business-directory-mysql mysql -uapp -psecure_app_password_2025! business_directory

echo.
echo Step 6: Creating uploads directory...
if not exist "uploads" (
    mkdir uploads
    echo   Created uploads directory
) else (
    echo   Uploads directory already exists
)

echo.
echo Step 7: Restarting application...
docker-compose restart app

echo.
echo Step 8: Waiting for application to start...
timeout /t 5 /nobreak >nul

echo.
echo ========================================
echo Setup Complete!
echo ========================================
echo.
echo Your Business Directory is now running at:
echo   http://localhost:8080
echo.
echo Test pages:
echo   - Main site: http://localhost:8080
echo   - Image demo: http://localhost:8080/image-demo.html
echo   - Business detail: http://localhost:8080/business-detail.html?id=1
echo.
echo Test credentials:
echo   Email: john@coffee.com
echo   Password: password123
echo.
echo To view logs:
echo   docker logs business-directory-app
echo.
echo To stop:
echo   docker-compose down
echo.

pause
