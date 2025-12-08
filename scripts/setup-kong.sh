#!/bin/bash
set -e

# Configuration
KONG_ADMIN="${KONG_ADMIN_URL:-http://localhost:3602}"
# Portal service uses internal Docker network name
UPSTREAM_URL="${PORTAL_SERVICE_URL:-http://portal-service:3000}"

echo "Targeting Kong: $KONG_ADMIN"
echo "Upstream Service: $UPSTREAM_URL"

# 1. Wait for Kong
echo "Waiting for Kong..."
until curl -s -f "$KONG_ADMIN" > /dev/null; do
    sleep 2
done
echo "Kong is ready."

# 2. Configure Service (Idempotent PUT)
echo "Configuring Service..."
curl -s -X PUT "$KONG_ADMIN/services/portal-service" \
    --data "url=$UPSTREAM_URL" > /dev/null

# 3. Configure Routes (Idempotent PUT)

# Route A: Public Auth (No plugins)
echo "Configuring Route: Public Auth..."
curl -s -X PUT "$KONG_ADMIN/services/portal-service/routes/portal-auth-route" \
    --data "paths[]=/api/v1/auth" \
    --data "strip_path=false" > /dev/null

# Route B: Protected Users (Will have plugins)
echo "Configuring Route: Protected Users..."
# We retrieve the ID because we need it to manage the plugin cleanly
PROTECTED_ROUTE_ID=$(curl -s -X PUT "$KONG_ADMIN/services/portal-service/routes/portal-users-route" \
    --data "paths[]=/api/v1/users" \
    --data "paths[]=/api/v1/memberships" \
    --data "paths[]=/api/v1/tenants" \
    --data "strip_path=false" | jq -r .id)

# 4. Lua Logic for Phantom Token
LUA_CODE='
local http = require "resty.http"
local cjson = require "cjson"

local LOG_PREFIX = "[PhantomAuth] "
local INTROSPECT_URL = "'$UPSTREAM_URL'/api/v1/auth/introspect"

-- 1. Security: Sanitize incoming headers to prevent spoofing
local headers_to_clear = {"X-Tenant-ID", "X-User-ID", "X-Role-ID", "X-Role-Name", "X-Permissions", "X-Authenticated"}
for _, h in ipairs(headers_to_clear) do
    kong.service.request.clear_header(h)
end

-- 2. Check for Token
local auth_header = kong.request.get_header("Authorization")
if not auth_header then
    kong.log.warn(LOG_PREFIX, "Request denied: Missing Authorization header")
    return kong.response.exit(401, { message = "Missing Authorization header" })
end

-- 3. Perform Introspection
local httpc = http.new()
httpc:set_timeout(5000)

kong.log.debug(LOG_PREFIX, "Introspecting token with: ", INTROSPECT_URL)
local start_time = kong.request.get_start_time()

local res, err = httpc:request_uri(INTROSPECT_URL, {
    method = "POST",
    headers = { ["Authorization"] = auth_header, ["Content-Type"] = "application/json" },
})

-- Calculate latency
local duration = (kong.request.get_start_time() - start_time)
kong.log.debug(LOG_PREFIX, "Introspection took: ", duration, "ms")

-- 4. Handle Network Errors
if not res then
    kong.log.err(LOG_PREFIX, "Introspection Connection Failed: ", err)
    return kong.response.exit(503, { message = "Auth Service Unavailable" })
end

-- 5. Handle Logic Errors
if res.status ~= 200 then
    kong.log.warn(LOG_PREFIX, "Token rejected. Upstream status: ", res.status)
    return kong.response.exit(401, { message = "Invalid or expired token" })
end

-- 6. Parse Response
local status, body = pcall(cjson.decode, res.body)
if not status then
    kong.log.err(LOG_PREFIX, "Failed to parse JSON from auth service")
    return kong.response.exit(500, { message = "Auth Service Error" })
end

if not body.active then
    kong.log.info(LOG_PREFIX, "Token is valid syntax but marked inactive")
    return kong.response.exit(401, { message = "Token is not active" })
end

-- 7. Header Injection
kong.service.request.clear_header("Authorization") -- Hide token from upstream

local safe_headers = {
    ["X-Tenant-ID"]   = res.headers["X-Tenant-ID"] or tostring(body.tenant_id or ""),
    ["X-User-ID"]     = res.headers["X-User-ID"] or tostring(body.user_id or ""),
    ["X-Role-Name"]   = res.headers["X-Role-Name"] or body.role_name or "",
    ["X-Authenticated"] = "true"
}

for k, v in pairs(safe_headers) do
    kong.service.request.set_header(k, v)
end

kong.log.notice(LOG_PREFIX, "Access Granted | User: ", safe_headers["X-User-ID"], " | Role: ", safe_headers["X-Role-Name"])
'

# 5. Apply Plugin
# Strategy: Delete existing plugin on this route (to ensure code update) then create new
echo "Applying Introspection Plugin..."

# Find plugin ID if it exists
PLUGIN_ID=$(curl -s "$KONG_ADMIN/routes/$PROTECTED_ROUTE_ID/plugins" | jq -r '.data[] | select(.name == "pre-function") | .id')

if [ -n "$PLUGIN_ID" ] && [ "$PLUGIN_ID" != "null" ]; then
    curl -s -X DELETE "$KONG_ADMIN/routes/$PROTECTED_ROUTE_ID/plugins/$PLUGIN_ID" > /dev/null
fi

# Create new plugin config
curl -s -X POST "$KONG_ADMIN/routes/$PROTECTED_ROUTE_ID/plugins" \
    --data "name=pre-function" \
    --data-urlencode "config.access[1]=$LUA_CODE" > /dev/null

echo "Configuration Complete."
echo "Logs can be viewed via: docker logs <kong_container_name> | grep '\[PhantomAuth\]'"