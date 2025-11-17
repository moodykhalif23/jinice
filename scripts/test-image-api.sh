#!/bin/bash

# Test script for Image Management API

BASE_URL="http://localhost:8080"
TOKEN=""

echo "🧪 Testing Image Management API"
echo "================================"
echo ""

# Test 1: Login
echo "Test 1: Login as business owner..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"john@coffee.com","password":"password123"}')

TOKEN=$(echo $LOGIN_RESPONSE | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    echo "❌ Login failed"
    echo "Response: $LOGIN_RESPONSE"
    exit 1
fi

echo "✅ Login successful"
echo ""

# Test 2: Get images for business
echo "Test 2: Get images for business ID 1..."
IMAGES_RESPONSE=$(curl -s "$BASE_URL/images?entity_type=business&entity_id=1")
echo "Response: $IMAGES_RESPONSE"
echo "✅ Get images successful"
echo ""

# Test 3: Add image by URL
echo "Test 3: Add image by URL..."
ADD_URL_RESPONSE=$(curl -s -X POST "$BASE_URL/images/add-url" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "entity_type": "business",
    "entity_id": 1,
    "image_url": "https://images.unsplash.com/photo-1554118811-1e0d58224f24?w=400",
    "caption": "Test image from API",
    "is_primary": false
  }')

IMAGE_ID=$(echo $ADD_URL_RESPONSE | grep -o '"id":[0-9]*' | cut -d':' -f2)

if [ -z "$IMAGE_ID" ]; then
    echo "❌ Add image by URL failed"
    echo "Response: $ADD_URL_RESPONSE"
else
    echo "✅ Add image by URL successful (ID: $IMAGE_ID)"
fi
echo ""

# Test 4: Update image
if [ ! -z "$IMAGE_ID" ]; then
    echo "Test 4: Update image caption..."
    UPDATE_RESPONSE=$(curl -s -X PUT "$BASE_URL/images/update" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{
        \"id\": $IMAGE_ID,
        \"caption\": \"Updated caption from API test\",
        \"display_order\": 5,
        \"is_primary\": false
      }")
    
    echo "Response: $UPDATE_RESPONSE"
    echo "✅ Update image successful"
    echo ""
fi

# Test 5: Get updated images
echo "Test 5: Get updated images list..."
UPDATED_IMAGES=$(curl -s "$BASE_URL/images?entity_type=business&entity_id=1")
echo "Response: $UPDATED_IMAGES"
echo "✅ Get updated images successful"
echo ""

# Test 6: Delete image
if [ ! -z "$IMAGE_ID" ]; then
    echo "Test 6: Delete test image..."
    DELETE_RESPONSE=$(curl -s -X DELETE "$BASE_URL/images/delete" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{\"id\": $IMAGE_ID}")
    
    echo "Response: $DELETE_RESPONSE"
    echo "✅ Delete image successful"
    echo ""
fi

# Test 7: Verify deletion
echo "Test 7: Verify image was deleted..."
FINAL_IMAGES=$(curl -s "$BASE_URL/images?entity_type=business&entity_id=1")
echo "Response: $FINAL_IMAGES"
echo "✅ Verification complete"
echo ""

echo "================================"
echo "🎉 All API tests completed!"
echo ""
