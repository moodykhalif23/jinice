@echo off
REM Setup script for image management system (Windows)

echo Setting up Image Management System...

REM Check if MySQL container is running
docker ps | findstr business-directory-mysql >nul
if errorlevel 1 (
    echo MySQL container is not running. Please start it with: docker-compose up -d mysql
    exit /b 1
)

echo MySQL container is running

REM Wait for MySQL to be ready
echo Waiting for MySQL to be ready...
timeout /t 5 /nobreak >nul

REM Run the migration
echo Running image system migration...
docker exec -i business-directory-mysql mysql -uapp -psecure_app_password_2025! business_directory < migrations/comprehensive_images.sql

if errorlevel 1 (
    echo Migration failed. Please check the error messages above.
    exit /b 1
)

echo Image system migration completed successfully!

REM Create uploads directory if it doesn't exist
if not exist "uploads" (
    mkdir uploads
    echo Created uploads directory
) else (
    echo Uploads directory already exists
)

echo.
echo Image Management System setup complete!
echo.
echo Next steps:
echo 1. Restart the application: docker-compose restart app
echo 2. Visit http://localhost:8080/image-demo.html to test the system
echo 3. Login with: john@coffee.com / password123
echo.

pause
