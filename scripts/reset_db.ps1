# Reset and Seed Database Script (PowerShell)
# This script connects to the MySQL container and runs the migration

Write-Host "🔄 Resetting and seeding database..." -ForegroundColor Cyan
Write-Host ""

# Run the SQL script in the MySQL container
Get-Content migrations/reset_and_seed.sql | docker exec -i business-directory-mysql mysql -uroot -pvery_secure_root_password_2025! business_directory

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Database reset and seeded successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📝 Test Credentials:" -ForegroundColor Yellow
    Write-Host "   Email: alice@coffeehaven.com"
    Write-Host "   Password: password123"
    Write-Host ""
    Write-Host "   Email: bob@techsolutions.com"
    Write-Host "   Password: password123"
    Write-Host ""
    Write-Host "   (All users have the same password: password123)" -ForegroundColor Gray
} else {
    Write-Host ""
    Write-Host "❌ Error resetting database. Please check the error messages above." -ForegroundColor Red
    exit 1
}
