#!/bin/bash

# Quick health check test script
# Tests the health endpoints without requiring full stack

echo "🔍 Testing Health API Endpoints..."

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

API_BASE="http://localhost:8888"

# Test function
test_endpoint() {
    local endpoint=$1
    local description=$2
    
    echo -n "Testing ${description}... "
    
    response=$(curl -s -o /dev/null -w "%{http_code}" "${API_BASE}${endpoint}" 2>/dev/null)
    
    if [ "$response" = "200" ] || [ "$response" = "503" ]; then
        echo -e "${GREEN}✓ PASSED (${response})${NC}"
    else
        echo -e "${RED}✗ FAILED (${response})${NC}"
    fi
}

# Check if app is running
if ! curl -s "${API_BASE}/api/v1/system/monitor/health" >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠ Application not running. Start with: make run${NC}"
    echo "Expected endpoints when running:"
    echo "• Health Check: ${API_BASE}/api/v1/system/monitor/health"
    echo "• Readiness: ${API_BASE}/api/v1/system/monitor/ready"
    echo "• Liveness: ${API_BASE}/api/v1/system/monitor/live"
    echo "• Metrics: ${API_BASE}/metrics"
    exit 1
fi

echo -e "${GREEN}Application is running! Testing endpoints...${NC}\n"

# Test health endpoints
test_endpoint "/api/v1/system/monitor/health" "System Health"
test_endpoint "/api/v1/system/monitor/ready" "Readiness Probe"
test_endpoint "/api/v1/system/monitor/live" "Liveness Probe"
test_endpoint "/metrics" "Prometheus Metrics"

echo -e "\n${GREEN}Health API tests completed!${NC}"
echo -e "View detailed health status: ${YELLOW}curl ${API_BASE}/api/v1/system/monitor/health | jq${NC}"