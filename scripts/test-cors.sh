#!/bin/bash
set -e

KONG_ADMIN="http://localhost:3602"

echo "Testing CORS Plugin Creation..."
echo "================================"

# Test 1: Check if Kong is accessible
echo "1. Checking Kong Admin API..."
curl -s "$KONG_ADMIN/status" | head -n 5

echo ""
echo "2. Listing existing plugins..."
curl -s "$KONG_ADMIN/plugins" | head -n 20

echo ""
echo "3. Attempting to create CORS plugin..."
RESPONSE=$(curl -s -X POST "$KONG_ADMIN/plugins" \
    --data "name=cors" \
    --data "config.origins=*" \
    --data "config.methods[]=GET" \
    --data "config.methods[]=HEAD" \
    --data "config.methods[]=PUT" \
    --data "config.methods[]=PATCH" \
    --data "config.methods[]=POST" \
    --data "config.methods[]=DELETE" \
    --data "config.methods[]=OPTIONS" \
    --data "config.headers[]=Accept" \
    --data "config.headers[]=Content-Type" \
    --data "config.headers[]=Authorization" \
    --data "config.headers[]=X-Refresh-Token" \
    --data "config.exposed_headers[]=X-Auth-Token" \
    --data "config.credentials=true" \
    --data "config.max_age=3600")

echo "$RESPONSE"

# Check if it was successful
if echo "$RESPONSE" | grep -q '"name":"cors"'; then
    echo ""
    echo "✅ CORS plugin created successfully!"
else
    echo ""
    echo "❌ CORS plugin creation failed. See response above."
fi
