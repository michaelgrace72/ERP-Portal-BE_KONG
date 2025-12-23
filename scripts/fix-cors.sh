#!/bin/bash

# Fix Kong CORS Configuration
# This script removes the existing CORS plugin and creates a new one with proper configuration

KONG_ADMIN_URL="http://localhost:3602"
ALLOWED_ORIGIN="http://localhost:5052"

echo "=== Fixing Kong CORS Configuration ==="

# Get all plugins and find CORS plugin ID
echo -e "\nFetching existing plugins..."
CORS_PLUGIN_ID=$(curl -s "$KONG_ADMIN_URL/plugins" | jq -r '.data[] | select(.name == "cors") | .id')

# Find and delete existing CORS plugin
if [ ! -z "$CORS_PLUGIN_ID" ] && [ "$CORS_PLUGIN_ID" != "null" ]; then
    echo "Found existing CORS plugin with ID: $CORS_PLUGIN_ID"
    echo "Deleting existing CORS plugin..."
    curl -s -X DELETE "$KONG_ADMIN_URL/plugins/$CORS_PLUGIN_ID"
    echo "✓ Existing CORS plugin deleted"
    sleep 1
fi

# Create new CORS plugin with correct configuration
echo -e "\nCreating new CORS plugin..."

curl -s -X POST "$KONG_ADMIN_URL/plugins" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "cors",
    "config": {
        "origins": ["'"$ALLOWED_ORIGIN"'"],
        "methods": ["GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"],
        "headers": ["Accept", "Accept-Version", "Content-Length", "Content-MD5", "Content-Type", "Date", "X-Auth-Token", "X-Tenant-ID", "Authorization"],
        "exposed_headers": ["X-Auth-Token", "Content-Length", "Content-Type"],
        "credentials": true,
        "max_age": 3600,
        "preflight_continue": false
    }
}' | jq '.'

echo -e "\n✓ CORS configuration fixed successfully!"
echo -e "\nYou can now refresh your browser and try logging in again."
