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
# Format: configure_service <service_name> <upstream_url> <route_path>

# Portal (Container)
configure_service "portal-service" "http://portal-service:3000" "/portal"

# Inventory (container)
configure_service "sie-erp_inventory-service" "http://sie-erp_inventory_be:3512" "/inventory"

# Manufacturing (container)
configure_service "sie-erp_manufacture-service" "http://sie-erp_manufacture_be:3514" "/manufacturing"

# General Ledger
configure_service "sie-erp_general_ledger-service" "http://sie-erp_general_ledger_be:3516" "/general-ledger"

# Cash Bank
configure_service "sie-erp_cash_bank-service" "http://sie-erp_cash_bank_be:3518" "/cash-bank"

# Purchase
configure_service "sie-erp_purchase-service" "http://sie-erp_purchase_be:3520" "/purchase"

# Sales
configure_service "sie-erp_sales-service" "http://sie-erp_sales_be:3522" "/sales"

#Fixed Asset
configure_service "sie-erp_fixed_asset-service" "http://sie-erp_fixed_asset_be:3524" "/fixed-asset"

#Account Receivable
configure_service "sie-erp_account_receivable-service" "http://sie-erp_account_receivable_be:3526" "/account-receivable"    

#Account Payable
configure_service "sie-erp_account_payable-service" "http://sie-erp_account_payable_be:3528" "/account-payable"

# Human Resource
configure_service "sie-erp_human_resource-service" "http://sie-erp_human_resource_be:3530" "/human-resource"

# Scheduling 
configure_service "sie-erp_scheduling-service" "http://sie-erp_scheduling_be:3532" "/scheduling"

# Taxation 
configure_service "sie-erp_taxation-service" "http://sie-erp_taxation_be:3534" "/taxation"

# General Settings (container)
configure_service "sie-erp_general_settings-service" "http://sie-erp_general_settings_be:3536" "/general-settings"

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