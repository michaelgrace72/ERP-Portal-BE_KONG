#!/bin/bash
set -e

# Configuration
KONG_ADMIN="${KONG_ADMIN_URL:-http://localhost:3602}"

echo "Targeting Kong Admin API: $KONG_ADMIN"

# 1. Wait for Kong
echo "Waiting for Kong..."
until curl -s -f "$KONG_ADMIN/status" > /dev/null; do
    sleep 2
done
echo "Kong is ready."

# 2. Define Helper Function (Service + Route)
configure_service() {
    local name=$1
    local upstream_url=$2
    local route_path=$3
    
    echo "Configuring $name..."
    
    # Upsert Service (PUT ensures creation or update without errors)
    curl -s -X PUT "$KONG_ADMIN/services/$name" \
        --data "url=$upstream_url" > /dev/null

    # Upsert Route
    curl -s -X PUT "$KONG_ADMIN/services/$name/routes/$name-route" \
        --data "paths[]=$route_path" \
        --data "strip_path=true" > /dev/null
}

# 3. Apply Configurations
# Format: configure_service <service_name> <upstrekm_url> <route_path>

# Portal (Container)
configure_service "portal-service" "http://portal-service:3000" "/portal"

# Inventory (External Host)
configure_service "inventory-service" "http://host.docker.internal:3512" "/inventory"

# Manufacturing (External Host)
configure_service "manufacturing-service" "http://host.docker.internal:3514" "/manufacturing"

# 4. Global CORS Plugin
echo "Configuring Global CORS..."
# We use POST for plugins. We grep to ignore "already exists" errors quietly.
curl -s -X POST "$KONG_ADMIN/plugins" \
    --data "name=cors" \
    --data "config.origins=*" \
    --data "config.methods=GET,POST,PUT,DELETE,PATCH,OPTIONS" \
    --data "config.headers=Accept,Content-Type,Authorization,X-Refresh-Token" \
    --data "config.exposed_headers=X-Auth-Token" \
    --data "config.credentials=true" \
    --data "config.max_age=3600" 2>/dev/null | grep -v "unique constraint" || true

echo "Configuration Complete."