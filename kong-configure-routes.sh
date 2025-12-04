#!/bin/bash

# Kong Routes Configuration Script
# Configures Kong to route:
#   /portal/* → Portal-BE (portal-service container)
#   /inventory/* → Inventory-BE (external service on port 3512)
#   /manufacturing/* → Manufacturing-BE (external service on port 3514)
#
# Current Ports:
#   - Kong Proxy (HTTP): 3600 (maps to internal 8000)
#   - Kong Admin API: 3602 (maps to internal 8001)
#   - Portal-BE: 3502 (exposed), portal-service:3000 (internal)
#   - Inventory-BE: 3512 (external via host.docker.internal)
#   - Manufacturing-BE: 3514 (external via host.docker.internal)

set -e

# Kong Admin API URL (internal Docker network)
KONG_ADMIN_URL="${KONG_ADMIN_URL:-http://localhost:3602}"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  Kong API Gateway - Routes Configuration${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Kong Admin API: $KONG_ADMIN_URL"
echo "Kong Proxy: http://localhost:3600"
echo ""

# Function to check if Kong is ready
check_kong_ready() {
    echo -e "${YELLOW}Checking if Kong is ready...${NC}"
    for i in {1..30}; do
        if curl -s "${KONG_ADMIN_URL}/status" > /dev/null 2>&1; then
            echo -e "${GREEN}✓ Kong is ready!${NC}"
            return 0
        fi
        echo "Waiting for Kong... (attempt $i/30)"
        sleep 2
    done
    echo -e "${RED}✗ Kong is not responding. Please start Kong first.${NC}"
    echo -e "${RED}  Make sure docker-compose is running: docker-compose ps${NC}"
    exit 1
}

# Function to delete if exists
delete_if_exists() {
    local resource_type=$1
    local resource_name=$2
    
    if curl -s "${KONG_ADMIN_URL}/${resource_type}/${resource_name}" 2>/dev/null | grep -q "id"; then
        echo -e "${YELLOW}  Deleting existing ${resource_type}: ${resource_name}${NC}"
        curl -s -X DELETE "${KONG_ADMIN_URL}/${resource_type}/${resource_name}" > /dev/null 2>&1 || true
        sleep 1
    fi
}

check_kong_ready

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  1️⃣  Portal Service${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Clean up existing portal configuration
delete_if_exists "routes" "portal-route"
delete_if_exists "services" "portal-service"

# Create Portal Service
# Using internal Docker network name and internal port
echo "Creating portal-service..."
curl -s -X POST "${KONG_ADMIN_URL}/services" \
  --data "name=portal-service" \
  --data "url=http://portal-service:3000" > /dev/null

echo -e "${GREEN}✓ Portal service created (http://portal-service:3000)${NC}"

# Create Portal Route - matches /portal/*
echo "Creating portal-route..."
curl -s -X POST "${KONG_ADMIN_URL}/services/portal-service/routes" \
  --data "name=portal-route" \
  --data "paths[]=/portal" \
  --data "strip_path=true" > /dev/null

echo -e "${GREEN}✓ Portal route created${NC}"
echo -e "  ${BLUE}http://localhost:3600/portal/api/v1/auth/register${NC}"
echo -e "  → ${YELLOW}http://portal-service:3000/api/v1/auth/register${NC}"

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  2️⃣  Inventory Service (External)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Clean up existing inventory configuration
delete_if_exists "routes" "inventory-route"
delete_if_exists "services" "inventory-service"

# Create Inventory Service
# Using host.docker.internal to reach external service on host
echo "Creating inventory-service..."
curl -s -X POST "${KONG_ADMIN_URL}/services" \
  --data "name=inventory-service" \
  --data "url=http://host.docker.internal:3512" > /dev/null

echo -e "${GREEN}✓ Inventory service created (http://host.docker.internal:3512)${NC}"

# Create Inventory Route - matches /inventory/*
echo "Creating inventory-route..."
curl -s -X POST "${KONG_ADMIN_URL}/services/inventory-service/routes" \
  --data "name=inventory-route" \
  --data "paths[]=/inventory" \
  --data "strip_path=true" > /dev/null

echo -e "${GREEN}✓ Inventory route created${NC}"
echo -e "  ${BLUE}http://localhost:3600/inventory/api/v1/warehouse${NC}"
echo -e "  → ${YELLOW}http://host.docker.internal:3512/api/v1/warehouse${NC}"

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  3️⃣  Manufacturing Service (External)${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Clean up existing manufacturing configuration
delete_if_exists "routes" "manufacturing-route"
delete_if_exists "services" "manufacturing-service"

# Create Manufacturing Service
echo "Creating manufacturing-service..."
curl -s -X POST "${KONG_ADMIN_URL}/services" \
  --data "name=manufacturing-service" \
  --data "url=http://host.docker.internal:3514" > /dev/null

echo -e "${GREEN}✓ Manufacturing service created (http://host.docker.internal:3514)${NC}"

# Create Manufacturing Route - matches /manufacturing/*
echo "Creating manufacturing-route..."
curl -s -X POST "${KONG_ADMIN_URL}/services/manufacturing-service/routes" \
  --data "name=manufacturing-route" \
  --data "paths[]=/manufacturing" \
  --data "strip_path=true" > /dev/null

echo -e "${GREEN}✓ Manufacturing route created${NC}"
echo -e "  ${BLUE}http://localhost:3600/manufacturing/api/v1/workorder${NC}"
echo -e "  → ${YELLOW}http://host.docker.internal:3514/api/v1/workorder${NC}"

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  4️⃣  Adding CORS Plugin${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

# Add CORS plugin globally (if not exists)
echo "Adding CORS plugin..."
CORS_RESULT=$(curl -s -X POST "${KONG_ADMIN_URL}/plugins" \
  --data "name=cors" \
  --data "config.origins=*" \
  --data "config.methods=GET" \
  --data "config.methods=POST" \
  --data "config.methods=PUT" \
  --data "config.methods=DELETE" \
  --data "config.methods=PATCH" \
  --data "config.methods=OPTIONS" \
  --data "config.headers=Accept" \
  --data "config.headers=Content-Type" \
  --data "config.headers=Authorization" \
  --data "config.headers=X-Refresh-Token" \
  --data "config.exposed_headers=X-Auth-Token" \
  --data "config.credentials=true" \
  --data "config.max_age=3600" 2>&1)

if echo "$CORS_RESULT" | grep -q "already exists"; then
    echo -e "${YELLOW}⚠ CORS plugin already exists (skipping)${NC}"
else
    echo -e "${GREEN}✓ CORS plugin added globally${NC}"
fi

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}  ✅ Configuration Complete!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}📍 Kong Gateway Routes:${NC}"
echo ""
echo "  Portal Service:"
echo -e "    ${GREEN}http://localhost:3600/portal/*${NC}"
echo -e "    Example: ${BLUE}http://localhost:3600/portal/api/v1/auth/register${NC}"
echo ""
echo "  Inventory Service:"
echo -e "    ${GREEN}http://localhost:3600/inventory/*${NC}"
echo -e "    Example: ${BLUE}http://localhost:3600/inventory/api/v1/warehouse${NC}"
echo ""
echo "  Manufacturing Service:"
echo -e "    ${GREEN}http://localhost:3600/manufacturing/*${NC}"
echo -e "    Example: ${BLUE}http://localhost:3600/manufacturing/api/v1/workorder${NC}"
echo ""
echo -e "${BLUE}🔗 Useful Links:${NC}"
echo -e "  Kong Proxy:     ${GREEN}http://localhost:3600${NC}"
echo -e "  Kong Admin API: ${GREEN}http://localhost:3602${NC}"
echo -e "  Portal Direct:  ${GREEN}http://localhost:3502${NC}"
echo ""
echo -e "${YELLOW}📝 Testing:${NC}"
echo "  # Test portal registration (via Kong)"
echo "  curl -X POST http://localhost:3600/portal/api/v1/auth/register \\"
echo "    -H 'Content-Type: application/json' \\"
echo "    -d '{\"email\":\"test@example.com\",\"password\":\"Test123!\",\"name\":\"Test User\",\"company_name\":\"Test Company\"}'"
echo ""
echo "  # Check Kong services"
echo "  curl http://localhost:3602/services | jq"
echo ""
echo "  # Check Kong routes"
echo "  curl http://localhost:3602/routes | jq"
echo ""
echo -e "${YELLOW}📝 Next Steps:${NC}"
echo "  1. Update frontend .env files to use Kong URLs (http://localhost:3600/portal)"
echo "  2. Ensure Inventory and Manufacturing services are running on ports 3512 and 3514"
echo "  3. Test the endpoints using the commands above"
echo ""
