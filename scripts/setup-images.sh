#!/bin/bash

# Setup script for image management system

echo "🖼️  Setting up Image Management System..."

# Check if MySQL is running
if ! docker ps | grep -q business-directory-mysql; then
    echo "❌ MySQL container is not running. Please start it with: docker-compose up -d mysql"
    exit 1
fi

echo "✅ MySQL container is running"

# Wait for MySQL to be ready
echo "⏳ Waiting for MySQL to be ready..."
sleep 5

# Run the migration
echo "📦 Running image system migration..."
docker exec -i business-directory-mysql mysql -u${DB_USER:-app} -p${DB_PASSWORD:-secure_app_password_2025!} business_directory < migrations/comprehensive_images.sql

if [ $? -eq 0 ]; then
    echo "✅ Image system migration completed successfully!"
else
    echo "❌ Migration failed. Please check the error messages above."
    exit 1
fi

# Create uploads directory if it doesn't exist
if [ ! -d "./uploads" ]; then
    mkdir -p ./uploads
    echo "✅ Created uploads directory"
else
    echo "✅ Uploads directory already exists"
fi

# Set permissions
chmod 755 ./uploads
echo "✅ Set permissions on uploads directory"

echo ""
echo "🎉 Image Management System setup complete!"
echo ""
echo "Next steps:"
echo "1. Restart the application: docker-compose restart app"
echo "2. Visit http://localhost:8080/image-demo.html to test the system"
echo "3. Login with: john@coffee.com / password123"
echo ""
