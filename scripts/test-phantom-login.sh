#!/bin/bash

# Test script for phantom token authentication flow
BASE_URL="http://localhost:3000/api/v1"

echo "=== Testing Phantom Token Authentication Flow ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test 1: Login with phantom token
echo -e "${BLUE}Test 1: Login with phantom token${NC}"
echo "POST $BASE_URL/auth/phantom-login"
RESPONSE=$(curl -s -X POST "$BASE_URL/auth/phantom-login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "mikhagracia72+test@gmail.com",
    "password": "password"
  }')

echo "$RESPONSE" | jq '.'

# Extract access token if login successful
ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.data.access_token // empty')

if [ -n "$ACCESS_TOKEN" ]; then
  echo -e "${GREEN}✓ Login successful${NC}"
  echo "Access Token: $ACCESS_TOKEN"
  echo ""
  
  # Test 2: Get session context
  echo -e "${BLUE}Test 2: Get session context${NC}"
  echo "GET $BASE_URL/auth/session"
  curl -s -X GET "$BASE_URL/auth/session" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'
  echo ""
  
  # Test 3: Refresh session
  echo -e "${BLUE}Test 3: Refresh session${NC}"
  echo "POST $BASE_URL/auth/refresh"
  curl -s -X POST "$BASE_URL/auth/refresh" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'
  echo ""
  
  # Test 4: Logout
  echo -e "${BLUE}Test 4: Logout${NC}"
  echo "POST $BASE_URL/auth/logout"
  curl -s -X POST "$BASE_URL/auth/logout" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'
  echo ""
  
  # Test 5: Try to get session after logout (should fail)
  echo -e "${BLUE}Test 5: Try to get session after logout (should fail)${NC}"
  echo "GET $BASE_URL/auth/session"
  curl -s -X GET "$BASE_URL/auth/session" \
    -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'
  echo ""
else
  echo -e "${RED}✗ Login failed${NC}"
  echo "Response: $RESPONSE"
fi

echo ""
echo "=== Tests Complete ==="
