#!/bin/bash

# Test Phantom Token Introspection Endpoint
# This script tests the Kong introspection flow

set -e

BASE_URL="${BASE_URL:-http://localhost:3000}"

echo "================================"
echo "Phantom Token Introspection Test"
echo "================================"
echo ""

# Step 1: Login
echo "1. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/phantom-login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@test4.com",
    "password": "12345678"
  }')

echo "Login Response:"
echo "$LOGIN_RESPONSE" | jq '.'

# Extract token
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
    echo "❌ Failed to get token"
    exit 1
fi

echo ""
echo "✅ Got reference token: ${TOKEN:0:20}..."
echo ""

# Step 2: Test Introspection
echo "2. Testing introspection endpoint..."
echo "   Calling: POST $BASE_URL/api/v1/auth/introspect"
echo ""

# Use -i to see headers
INTROSPECT_RESPONSE=$(curl -s -i -X POST "$BASE_URL/api/v1/auth/introspect" \
  -H "Authorization: Bearer $TOKEN")

# Extract headers
HEADERS=$(echo "$INTROSPECT_RESPONSE" | grep -E "^X-")

echo "Response Headers (Kong will inject these):"
echo "$HEADERS"
echo ""

# Extract body
BODY=$(echo "$INTROSPECT_RESPONSE" | tail -n 1)

echo "Response Body:"
echo "$BODY" | jq '.'
echo ""

# Verify active=true
IS_ACTIVE=$(echo "$BODY" | jq -r '.active')

if [ "$IS_ACTIVE" == "true" ]; then
    echo "✅ Token is active"
    
    # Show key info
    TENANT_ID=$(echo "$BODY" | jq -r '.tenant_id')
    USER_ID=$(echo "$BODY" | jq -r '.user_id')
    ROLE=$(echo "$BODY" | jq -r '.role_name')
    PERM_COUNT=$(echo "$BODY" | jq -r '.permissions | length')
    
    echo "   - Tenant ID: $TENANT_ID"
    echo "   - User ID: $USER_ID"
    echo "   - Role: $ROLE"
    echo "   - Permissions: $PERM_COUNT"
else
    echo "❌ Token is not active"
    exit 1
fi

echo ""
echo "================================"
echo "3. Testing with invalid token..."
echo ""

INVALID_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/introspect" \
  -H "Authorization: Bearer ref_invalid_token_12345678901234567890123456789012345678901234567890")

echo "$INVALID_RESPONSE" | jq '.'

IS_ACTIVE=$(echo "$INVALID_RESPONSE" | jq -r '.active')

if [ "$IS_ACTIVE" == "false" ]; then
    echo "✅ Invalid token correctly rejected"
else
    echo "❌ Invalid token was not rejected"
    exit 1
fi

echo ""
echo "================================"
echo "4. Simulating Kong flow..."
echo "   (Manually setting Kong headers)"
echo ""

# Extract headers from introspection
X_TENANT_ID=$(echo "$HEADERS" | grep "X-Tenant-ID" | cut -d: -f2 | tr -d '[:space:]')
X_USER_ID=$(echo "$HEADERS" | grep "X-User-ID" | cut -d: -f2 | tr -d '[:space:]')
X_PERMISSIONS=$(echo "$HEADERS" | grep "X-Permissions" | cut -d: -f2 | tr -d '[:space:]')

echo "   Headers that Kong would inject:"
echo "   - X-Tenant-ID: $X_TENANT_ID"
echo "   - X-User-ID: $X_USER_ID"
echo "   - X-Permissions: ${X_PERMISSIONS:0:50}..."
echo ""

echo "✅ All tests passed!"
echo ""
echo "Next Steps:"
echo "1. Set up Kong (see docs/KONG_SETUP.md)"
echo "2. Configure auth plugin to call /api/v1/auth/introspect"
echo "3. Update routes to use KongAuthMiddleware"
echo ""
