#!/bin/bash

# Kong Setup Script
# This script configures Kong Gateway for the Phantom Token pattern

set -e

KONG_ADMIN_URL="${KONG_ADMIN_URL:-http://localhost:8001}"
# Use Docker gateway IP for Linux (172.19.0.1), or host.docker.internal for Mac/Windows
PORTAL_SERVICE_URL="${PORTAL_SERVICE_URL:-http://172.19.0.1:3000}"

echo "================================"
echo "Kong Gateway Setup"
echo "================================"
echo ""
echo "Kong Admin URL: $KONG_ADMIN_URL"
echo "Portal Service URL: $PORTAL_SERVICE_URL"
echo ""

# Wait for Kong to be ready
echo "1. Waiting for Kong to be ready..."
until curl -s "$KONG_ADMIN_URL" > /dev/null; do
    echo "   Kong is not ready yet, waiting..."
    sleep 2
done
echo "   ✅ Kong is ready!"
echo ""

# Step 1: Register Portal Service
echo "2. Registering Portal Service..."
PORTAL_SERVICE=$(curl -s -X POST "$KONG_ADMIN_URL/services/" \
    --data "name=portal-service" \
    --data "url=$PORTAL_SERVICE_URL")

PORTAL_SERVICE_ID=$(echo "$PORTAL_SERVICE" | jq -r '.id')
echo "   ✅ Portal Service registered (ID: $PORTAL_SERVICE_ID)"
echo ""

# Step 2: Create route for auth endpoints (public, no authentication)
echo "3. Creating route for auth endpoints (public)..."
AUTH_ROUTE=$(curl -s -X POST "$KONG_ADMIN_URL/services/portal-service/routes" \
    --data "name=portal-auth-route" \
    --data "paths[]=/api/v1/auth" \
    --data "strip_path=false")

AUTH_ROUTE_ID=$(echo "$AUTH_ROUTE" | jq -r '.id')
echo "   ✅ Auth route created (ID: $AUTH_ROUTE_ID)"
echo "      Access: http://localhost:8000/api/v1/auth/*"
echo ""

# Step 3: Create route for user management (requires authentication)
echo "4. Creating route for user management (authenticated)..."
USERS_ROUTE=$(curl -s -X POST "$KONG_ADMIN_URL/services/portal-service/routes" \
    --data "name=portal-users-route" \
    --data "paths[]=/api/v1/users" \
    --data "paths[]=/api/v1/memberships" \
    --data "paths[]=/api/v1/tenants" \
    --data "strip_path=false")

USERS_ROUTE_ID=$(echo "$USERS_ROUTE" | jq -r '.id')
echo "   ✅ User management route created (ID: $USERS_ROUTE_ID)"
echo ""

# Step 5: Add pre-function plugin (introspection with header security)
echo "5. Adding pre-function plugin (phantom token validation with header security)..."

LUA_CODE='
local http = require "resty.http"
local cjson = require "cjson"

-- SECURITY: Strip any client-provided authentication headers to prevent spoofing
-- This must happen BEFORE introspection to ensure clients cannot fake their identity
kong.service.request.clear_header("X-Tenant-ID")
kong.service.request.clear_header("X-User-ID")
kong.service.request.clear_header("X-Role-ID")
kong.service.request.clear_header("X-Role-Name")
kong.service.request.clear_header("X-Permissions")
kong.service.request.clear_header("X-Authenticated")

-- Get auth header from client
local auth_header = kong.request.get_header("Authorization")
if not auth_header then
  return kong.response.exit(401, { message = "Missing Authorization header" })
end

-- Call introspection endpoint
local httpc = http.new()
httpc:set_timeout(5000)

local res, err = httpc:request_uri("http://172.19.0.1:3000/api/v1/auth/introspect", {
  method = "POST",
  headers = {
    ["Authorization"] = auth_header,
    ["Content-Type"] = "application/json",
  },
})

if not res then
  kong.log.err("Introspection failed: ", err)
  return kong.response.exit(503, { message = "Authentication service unavailable" })
end

if res.status ~= 200 then
  kong.log.warn("Introspection returned status: ", res.status)
  return kong.response.exit(401, { message = "Invalid or expired token" })
end

-- Parse response
local body = cjson.decode(res.body)
if not body.active then
  kong.log.info("Token is not active")
  return kong.response.exit(401, { message = "Token is not active" })
end

-- Remove Authorization header (prevent spoofing)
kong.service.request.clear_header("Authorization")

-- Inject headers from introspection response
local tenant_id = res.headers["X-Tenant-ID"] or tostring(body.tenant_id or "")
local user_id = res.headers["X-User-ID"] or tostring(body.user_id or "")
local permissions = res.headers["X-Permissions"] or ""
local role_name = res.headers["X-Role-Name"] or body.role_name or ""

kong.service.request.set_header("X-Tenant-ID", tenant_id)
kong.service.request.set_header("X-User-ID", user_id)
kong.service.request.set_header("X-Role-ID", res.headers["X-Role-ID"] or "")
kong.service.request.set_header("X-Role-Name", role_name)
kong.service.request.set_header("X-Permissions", permissions)
kong.service.request.set_header("X-Authenticated", "true")

kong.log.info("Authenticated user_id=", user_id, " tenant_id=", tenant_id, " role=", role_name)
'

curl -s -X POST "$KONG_ADMIN_URL/routes/$USERS_ROUTE_ID/plugins" \
    --data "name=pre-function" \
    --data-urlencode "config.access[1]=$LUA_CODE" > /dev/null

echo "   ✅ Introspection plugin added"
echo ""

echo "================================"
echo "Kong Configuration Complete!"
echo "================================"
echo ""
echo "Routes configured:"
echo "  - Public Auth:  http://localhost:8000/api/v1/auth/*"
echo "  - Authenticated: http://localhost:8000/api/v1/users/*"
echo "                   http://localhost:8000/api/v1/memberships/*"
echo "                   http://localhost:8000/api/v1/tenants/*"
echo ""
echo "Next steps:"
echo "  1. Login: curl -X POST http://localhost:8000/api/v1/auth/phantom-login \\"
echo "            -d '{\"email\":\"test@test4.com\",\"password\":\"12345678\"}'"
echo ""
echo "  2. Test:  curl -X GET http://localhost:8000/api/v1/users/me \\"
echo "            -H 'Authorization: Bearer <token>'"
echo ""
