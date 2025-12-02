#!/bin/bash
set -e # stop on any error

BLUE_SERVICE="portal-be-blue"
BLUE_CONTAINER="erp_portal_be_blue"
BLUE_PORT=3502
GREEN_SERVICE="portal-be-green"
GREEN_CONTAINER="erp_portal_be_green"
GREEN_PORT=3503
HEALTH_CHECK_PATH="/health" 
TIMEOUT=60
SLEEP_INTERVAL=5
MAX_RETRIES=$((TIMEOUT / SLEEP_INTERVAL))

ACTIVE_SERVICE=""
ACTIVE_CONTAINER=""
INACTIVE_SERVICE=""
INACTIVE_CONTAINER=""
INACTIVE_PORT=""

# Define Active and Inactive Services based on running containers
if docker ps --format "{{.Names}}" | grep -q "^${BLUE_CONTAINER}\$"; then
    ACTIVE_SERVICE=$BLUE_SERVICE
    ACTIVE_CONTAINER=$BLUE_CONTAINER
    INACTIVE_SERVICE=$GREEN_SERVICE
    INACTIVE_CONTAINER=$GREEN_CONTAINER
    INACTIVE_PORT=$GREEN_PORT
elif docker ps --format "{{.Names}}" | grep -q "^${GREEN_CONTAINER}\$"; then
    ACTIVE_SERVICE=$GREEN_SERVICE
    ACTIVE_CONTAINER=$GREEN_CONTAINER
    INACTIVE_SERVICE=$BLUE_SERVICE
    INACTIVE_CONTAINER=$BLUE_CONTAINER
    INACTIVE_PORT=$BLUE_PORT
else
    # First deployment, default to Blue
    echo "INFO: No active service found. Starting first deployment with Blue."
    ACTIVE_SERVICE=""
    ACTIVE_CONTAINER=""
    INACTIVE_SERVICE=$BLUE_SERVICE
    INACTIVE_CONTAINER=$BLUE_CONTAINER
    INACTIVE_PORT=$BLUE_PORT
fi

echo "========================================="
echo "Blue-Green Deployment Strategy"
echo "========================================="
echo "Active Service:   ${ACTIVE_SERVICE:-NONE}"
echo "Inactive Service: ${INACTIVE_SERVICE}"
echo "Image Tag:        ${IMAGE_TAG:-latest}"
echo "Registry Image:   ${CI_REGISTRY_IMAGE:-localhost:8001/portal-be}"
echo "========================================="

# Pull new image and start inactive service
echo -e "\nINFO: Pulling image for $INACTIVE_SERVICE"
docker compose pull $INACTIVE_SERVICE || {
    echo "Warning: Pull failed, will try to use existing image"
}

echo "INFO: Starting $INACTIVE_SERVICE (container: $INACTIVE_CONTAINER)"
docker compose up -d $INACTIVE_SERVICE

# Health check for inactive service
echo -e "\nINFO: Waiting for health check on $INACTIVE_SERVICE at port ${INACTIVE_PORT}"
HEALTH_CHECK_URL="http://localhost:${INACTIVE_PORT}${HEALTH_CHECK_PATH}"

for i in $(seq 1 $MAX_RETRIES); do
    response=$(curl -s -o /dev/null -w "%{http_code}" $HEALTH_CHECK_URL 2>/dev/null || echo "000")
    if [ "$response" -eq 200 ]; then
        echo -e "\n✅ SUCCESS: $INACTIVE_SERVICE is healthy (HTTP $response)"
        
        # Stop and remove old active service if exists
        if [ -n "$ACTIVE_SERVICE" ]; then
            echo -e "\nINFO: Stopping old active service ($ACTIVE_CONTAINER)..."
            docker compose stop $ACTIVE_SERVICE || true
            docker compose rm -f $ACTIVE_SERVICE || true
            echo "✅ Old service stopped and removed"
        fi
        
        echo -e "\n========================================="
        echo "✅ Blue-Green Deployment Successful!"
        echo "========================================="
        echo "New Active Service: $INACTIVE_SERVICE"
        echo "Port: $INACTIVE_PORT"
        echo "Container: $INACTIVE_CONTAINER"
        echo "========================================="
        exit 0
    fi
    echo "INFO: Health check failed (HTTP $response). Retrying in ${SLEEP_INTERVAL}s... (attempt $i/$MAX_RETRIES)"
    sleep $SLEEP_INTERVAL
done

# Rollback if inactive service is not healthy
echo -e "\n❌ ERROR: $INACTIVE_SERVICE IS NOT HEALTHY AFTER ${TIMEOUT}s"
echo "ERROR: Rolling back by stopping $INACTIVE_CONTAINER"
docker compose stop $INACTIVE_SERVICE || true
docker compose rm -f $INACTIVE_SERVICE || true

if [ -n "$ACTIVE_SERVICE" ]; then
    echo "INFO: Active service ($ACTIVE_CONTAINER) remains running"
fi

echo -e "\n========================================="
echo "❌ Deployment Failed - Rollback Complete"
echo "========================================="
exit 1