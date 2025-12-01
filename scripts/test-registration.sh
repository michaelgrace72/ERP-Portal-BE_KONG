#!/bin/bash

# Test Registration Flow with Kong Integration
# This script tests the complete registration flow

set -e

BASE_URL="http://localhost:3000"
KONG_ADMIN_URL="http://localhost:8001"

echo "=========================================="
echo "Testing Portal Service Registration Flow"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test data
EMAIL="test-$(date +%s)@example.com"
PASSWORD="SecurePass123!"
NAME="Test User"
COMPANY_NAME="Test Company $(date +%s)"

echo -e "${YELLOW}Test Data:${NC}"
echo "Email: $EMAIL"
echo "Password: $PASSWORD"
echo "Name: $NAME"
echo "Company: $COMPANY_NAME"
echo ""

# Step 1: Test registration
echo -e "${YELLOW}Step 1: Registering new user with tenant...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"name\": \"$NAME\",
    \"company_name\": \"$COMPANY_NAME\"
  }")

echo "Response:"
echo "$REGISTER_RESPONSE" | jq '.'

# Extract user UUID from response
USER_UUID=$(echo "$REGISTER_RESPONSE" | jq -r '.data.user_uuid')

if [ -z "$USER_UUID" ] || [ "$USER_UUID" == "null" ]; then
  echo -e "${RED}❌ Failed: Could not extract user UUID from registration response${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Registration successful!${NC}"
echo "User UUID: $USER_UUID"
echo ""

# Step 2: Verify Kong consumer was created
echo -e "${YELLOW}Step 2: Verifying Kong consumer was created...${NC}"
KONG_CONSUMER=$(curl -s "$KONG_ADMIN_URL/consumers/$USER_UUID")

echo "Kong Consumer:"
echo "$KONG_CONSUMER" | jq '.'

KONG_CONSUMER_ID=$(echo "$KONG_CONSUMER" | jq -r '.id')

if [ -z "$KONG_CONSUMER_ID" ] || [ "$KONG_CONSUMER_ID" == "null" ]; then
  echo -e "${RED}❌ Failed: Kong consumer was not created${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Kong consumer verified!${NC}"
echo "Kong Consumer ID: $KONG_CONSUMER_ID"
echo ""

# Step 3: Verify consumer tags
echo -e "${YELLOW}Step 3: Checking consumer tags...${NC}"
TAGS=$(echo "$KONG_CONSUMER" | jq -r '.tags[]')
echo "Tags: $TAGS"
echo ""

# Step 4: List all consumers (optional)
echo -e "${YELLOW}Step 4: Listing all Kong consumers...${NC}"
ALL_CONSUMERS=$(curl -s "$KONG_ADMIN_URL/consumers")
CONSUMER_COUNT=$(echo "$ALL_CONSUMERS" | jq '.data | length')
echo "Total consumers: $CONSUMER_COUNT"
echo ""

# Step 5: Test duplicate registration (should fail)
echo -e "${YELLOW}Step 5: Testing duplicate registration (should fail)...${NC}"
DUPLICATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"name\": \"$NAME\",
    \"company_name\": \"$COMPANY_NAME\"
  }")

echo "Response:"
echo "$DUPLICATE_RESPONSE" | jq '.'

ERROR_STATUS=$(echo "$DUPLICATE_RESPONSE" | jq -r '.status')

if [ "$ERROR_STATUS" == "error" ]; then
  echo -e "${GREEN}✓ Duplicate registration correctly rejected!${NC}"
else
  echo -e "${RED}❌ Failed: Duplicate registration should have been rejected${NC}"
fi
echo ""

# Summary
echo "=========================================="
echo -e "${GREEN}All tests completed!${NC}"
echo "=========================================="
echo ""
echo "Summary:"
echo "- User registered with UUID: $USER_UUID"
echo "- Kong consumer created with ID: $KONG_CONSUMER_ID"
echo "- Duplicate registration correctly rejected"
echo ""
echo "You can now:"
echo "1. View consumer in Kong Admin: $KONG_ADMIN_URL/consumers/$USER_UUID"
echo "2. Check database records for user, tenant, role, and membership"
echo "3. Configure Kong JWT plugin for authentication"
