#!/bin/bash

# ============================================================================
# GreenLync API Gateway - Monitoring Stack Test Script
# ============================================================================
# This script tests the complete monitoring setup including:
# - Database connectivity and monitoring tables
# - Prometheus metrics collection
# - Jaeger tracing
# - Grafana dashboard access
# - AlertManager configuration
# - Health endpoint functionality

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_BASE_URL="http://localhost:8888"
PROMETHEUS_URL="http://localhost:9090"
GRAFANA_URL="http://localhost:3000"
JAEGER_URL="http://localhost:16686"
ALERTMANAGER_URL="http://localhost:9093"

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  GreenLync API Gateway - Monitoring Test${NC}"
echo -e "${BLUE}================================================${NC}"

# Function to check if a service is running
check_service() {
    local service_name=$1
    local url=$2
    local expected_code=${3:-200}
    
    echo -n "Testing ${service_name}... "
    
    if curl -f -s -o /dev/null -w "%{http_code}" "$url" | grep -q "$expected_code"; then
        echo -e "${GREEN}✓ PASSED${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

# Function to test API endpoint
test_api_endpoint() {
    local endpoint=$1
    local description=$2
    local expected_code=${3:-200}
    
    echo -n "Testing ${description}... "
    
    response_code=$(curl -s -o /dev/null -w "%{http_code}" "${API_BASE_URL}${endpoint}")
    
    if [ "$response_code" = "$expected_code" ]; then
        echo -e "${GREEN}✓ PASSED (${response_code})${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED (${response_code})${NC}"
        return 1
    fi
}

# Function to test database connectivity
test_database() {
    echo -n "Testing database connectivity... "
    
    # Try to connect to MySQL container
    if docker exec mysql mysqladmin -u root -ppassword ping >/dev/null 2>&1; then
        echo -e "${GREEN}✓ PASSED${NC}"
        
        echo -n "Testing monitoring tables... "
        # Check if monitoring tables exist
        tables=$(docker exec mysql mysql -u root -ppassword greenlync -e "SHOW TABLES LIKE '%health%'" 2>/dev/null | wc -l)
        if [ "$tables" -gt 0 ]; then
            echo -e "${GREEN}✓ PASSED${NC}"
        else
            echo -e "${YELLOW}⚠ WARNING - Monitoring tables not found${NC}"
        fi
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

# Function to test Prometheus metrics
test_prometheus_metrics() {
    echo -n "Testing Prometheus metrics collection... "
    
    # Check if Prometheus can scrape the API gateway
    response=$(curl -s "${PROMETHEUS_URL}/api/v1/query?query=up{job=\"greenlync-api-gateway\"}")
    
    if echo "$response" | grep -q '"value":\[.*,"1"\]'; then
        echo -e "${GREEN}✓ PASSED${NC}"
        
        echo -n "Testing custom metrics... "
        # Check for HTTP request metrics
        response=$(curl -s "${PROMETHEUS_URL}/api/v1/query?query=http_requests_total")
        if echo "$response" | grep -q '"result":\['; then
            echo -e "${GREEN}✓ PASSED${NC}"
        else
            echo -e "${YELLOW}⚠ WARNING - Custom metrics not found${NC}"
        fi
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

# Function to test Jaeger tracing
test_jaeger_tracing() {
    echo -n "Testing Jaeger tracing... "
    
    # Make a request to generate a trace
    curl -s "${API_BASE_URL}/api/v1/system/monitor/health" >/dev/null
    
    # Wait a moment for trace to be collected
    sleep 2
    
    # Check if Jaeger has traces
    response=$(curl -s "${JAEGER_URL}/api/services")
    
    if echo "$response" | grep -q "greenlync-api-gateway"; then
        echo -e "${GREEN}✓ PASSED${NC}"
    else
        echo -e "${YELLOW}⚠ WARNING - No traces found${NC}"
    fi
}

# Function to load test the API
load_test_api() {
    echo -e "${BLUE}Running load test to generate metrics...${NC}"
    
    for i in {1..20}; do
        curl -s "${API_BASE_URL}/api/v1/system/monitor/health" >/dev/null &
        curl -s "${API_BASE_URL}/api/v1/system/monitor/ready" >/dev/null &
        curl -s "${API_BASE_URL}/api/v1/system/monitor/live" >/dev/null &
    done
    
    wait  # Wait for all background jobs to complete
    echo -e "${GREEN}Load test completed${NC}"
}

# Function to test Grafana dashboards
test_grafana_dashboards() {
    echo -n "Testing Grafana dashboard access... "
    
    # Check if Grafana is accessible
    if curl -f -s "${GRAFANA_URL}/api/health" >/dev/null; then
        echo -e "${GREEN}✓ PASSED${NC}"
        
        echo -n "Testing dashboard availability... "
        # Check if dashboards are available (this requires authentication, so just check the login page)
        if curl -s "${GRAFANA_URL}" | grep -q "Grafana"; then
            echo -e "${GREEN}✓ PASSED${NC}"
        else
            echo -e "${YELLOW}⚠ WARNING - Dashboard content not accessible${NC}"
        fi
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

# Function to validate Prometheus configuration
validate_prometheus_config() {
    echo -n "Validating Prometheus configuration... "
    
    # Check if Prometheus config is valid
    response=$(curl -s "${PROMETHEUS_URL}/api/v1/status/config")
    
    if echo "$response" | grep -q "greenlync-api-gateway"; then
        echo -e "${GREEN}✓ PASSED${NC}"
    else
        echo -e "${YELLOW}⚠ WARNING - API Gateway not found in Prometheus config${NC}"
    fi
}

# Function to check AlertManager
test_alertmanager() {
    echo -n "Testing AlertManager... "
    
    response=$(curl -s "${ALERTMANAGER_URL}/api/v1/status")
    
    if echo "$response" | grep -q '"status":"success"'; then
        echo -e "${GREEN}✓ PASSED${NC}"
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

# Main test execution
main() {
    local failed_tests=0
    local total_tests=0
    
    echo -e "${YELLOW}Waiting for services to be ready...${NC}"
    sleep 5
    
    echo -e "\n${BLUE}=== Basic Service Health Checks ===${NC}"
    
    # Basic service checks
    check_service "API Gateway" "${API_BASE_URL}/api/v1/system/monitor/health" || ((failed_tests++))
    ((total_tests++))
    
    check_service "Prometheus" "${PROMETHEUS_URL}/-/healthy" || ((failed_tests++))
    ((total_tests++))
    
    check_service "Grafana" "${GRAFANA_URL}/api/health" || ((failed_tests++))
    ((total_tests++))
    
    check_service "Jaeger" "${JAEGER_URL}/" || ((failed_tests++))
    ((total_tests++))
    
    check_service "AlertManager" "${ALERTMANAGER_URL}/-/healthy" || ((failed_tests++))
    ((total_tests++))
    
    echo -e "\n${BLUE}=== API Endpoint Tests ===${NC}"
    
    # API endpoint tests
    test_api_endpoint "/metrics" "Prometheus metrics endpoint" || ((failed_tests++))
    ((total_tests++))
    
    test_api_endpoint "/api/v1/system/monitor/health" "Health check endpoint" || ((failed_tests++))
    ((total_tests++))
    
    test_api_endpoint "/api/v1/system/monitor/ready" "Readiness probe" || ((failed_tests++))
    ((total_tests++))
    
    test_api_endpoint "/api/v1/system/monitor/live" "Liveness probe" || ((failed_tests++))
    ((total_tests++))
    
    test_api_endpoint "/swagger/" "Swagger documentation" || ((failed_tests++))
    ((total_tests++))
    
    echo -e "\n${BLUE}=== Database Tests ===${NC}"
    
    # Database tests
    test_database || ((failed_tests++))
    ((total_tests++))
    
    echo -e "\n${BLUE}=== Monitoring Integration Tests ===${NC}"
    
    # Generate some test data
    load_test_api
    
    # Wait for metrics to be collected
    echo -e "${YELLOW}Waiting for metrics collection...${NC}"
    sleep 10
    
    # Monitoring tests
    test_prometheus_metrics || ((failed_tests++))
    ((total_tests++))
    
    test_jaeger_tracing || ((failed_tests++))
    ((total_tests++))
    
    test_grafana_dashboards || ((failed_tests++))
    ((total_tests++))
    
    validate_prometheus_config || ((failed_tests++))
    ((total_tests++))
    
    test_alertmanager || ((failed_tests++))
    ((total_tests++))
    
    # Final results
    echo -e "\n${BLUE}=== Test Results ===${NC}"
    
    if [ $failed_tests -eq 0 ]; then
        echo -e "${GREEN}✓ ALL TESTS PASSED (${total_tests}/${total_tests})${NC}"
        echo -e "\n${GREEN}🎉 Monitoring stack is fully operational!${NC}"
        echo -e "\n${BLUE}Access your monitoring tools:${NC}"
        echo -e "• API Gateway: ${API_BASE_URL}"
        echo -e "• Grafana: ${GRAFANA_URL} (admin/admin)"
        echo -e "• Prometheus: ${PROMETHEUS_URL}"
        echo -e "• Jaeger: ${JAEGER_URL}"
        echo -e "• AlertManager: ${ALERTMANAGER_URL}"
        return 0
    else
        echo -e "${RED}✗ ${failed_tests} TEST(S) FAILED (${failed_tests}/${total_tests})${NC}"
        echo -e "\n${YELLOW}Check the output above for details.${NC}"
        echo -e "${YELLOW}You may need to wait longer for services to start or check logs.${NC}"
        return 1
    fi
}

# Check if we're running in Docker context
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is required but not found. Please install Docker.${NC}"
    exit 1
fi

# Check if curl is available
if ! command -v curl &> /dev/null; then
    echo -e "${RED}curl is required but not found. Please install curl.${NC}"
    exit 1
fi

# Run the main test function
main "$@"