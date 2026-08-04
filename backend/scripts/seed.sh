#!/usr/bin/env bash

API_URL="http://localhost:8080/api/v1"
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZGVhNGQzNmYtZjdhOC00ODZiLTg0YmQtOGRhN2Q1MmQwMDE4IiwidXNlcl9yb2xlIjoiY3VzdG9tZXIiLCJ0eXBlIjoiYWNjZXNzX3Rva2VuIiwic3ViIjoiZGVhNGQzNmYtZjdhOC00ODZiLTg0YmQtOGRhN2Q1MmQwMDE4IiwiZXhwIjoxNzg1ODYwMjI0LCJpYXQiOjE3ODU4NTkzMjR9.w3dkQDR8S5_hlblKBjBzHBi91HERIcDKnWA4VSgutc0"

RAND_SUFFIX=$RANDOM

echo "1. Seeding Category..."
CAT_RESP=$(curl -s -X POST "$API_URL/admin/categories" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Apparel '$RAND_SUFFIX'"}')
CAT_ID=$(echo "$CAT_RESP" | jq -r '.data.id')
echo "Created Category ID: $CAT_ID"

echo "2. Seeding Brand..."
BRAND_RESP=$(curl -s -X POST "$API_URL/admin/brands" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Puma '$RAND_SUFFIX'", "link": "https://puma.com"}')
BRAND_ID=$(echo "$BRAND_RESP" | jq -r '.data.id')
echo "Created Brand ID: $BRAND_ID"

echo "3. Seeding Product..."
PROD_RESP=$(curl -s -X POST "$API_URL/admin/products" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Speedcat Sneaker '$RAND_SUFFIX'",
    "description": "Classic motor-sport lifestyle sneaker",
    "highlights": ["Classic", "Sleek", "Lightweight"],
    "category_id": "'"$CAT_ID"'",
    "brand_id": "'"$BRAND_ID"'"
  }')
PROD_PID=$(echo "$PROD_RESP" | jq -r '.data.id')
echo "Created Product Public ID: $PROD_PID"

echo "4. Fetching Internal Product UUID..."
ADMIN_PROD=$(curl -s -X GET "$API_URL/admin/products" -H "Authorization: Bearer $TOKEN")
INTERNAL_PROD_ID=$(echo "$ADMIN_PROD" | jq -r '.data[] | select(.public_id == "'"$PROD_PID"'") | .id')
echo "Internal Product UUID: $INTERNAL_PROD_ID"

echo "5. Seeding Variant..."
VAR_RESP=$(curl -s -X POST "$API_URL/admin/products/$INTERNAL_PROD_ID/variants" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Black / Size 42",
    "price": 12999,
    "crossed_out_price": 14999,
    "currency": "USD",
    "is_default": true,
    "attributes": {"color": "Black", "size": "42"}
  }')
VAR_ID=$(echo "$VAR_RESP" | jq -r '.data.id')
echo "Created Variant ID: $VAR_ID"

echo "6. Publishing Product..."
curl -s -X POST "$API_URL/admin/products/$INTERNAL_PROD_ID/publish" \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "Done seeding!"
