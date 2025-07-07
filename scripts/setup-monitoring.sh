#!/bin/bash

# ============================================================================
# GreenLync API Gateway - Complete Monitoring Setup Script
# ============================================================================
# This script sets up the complete monitoring infrastructure including:
# - Database initialization with monitoring tables
# - Prometheus configuration
# - Grafana dashboards
# - Jaeger tracing
# - AlertManager setup

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}  GreenLync API Gateway - Monitoring Setup${NC}"
echo -e "${BLUE}================================================${NC}"

# Function to print step headers
print_step() {
    echo -e "\n${BLUE}=== $1 ===${NC}"
}

# Function to print success message
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

# Function to print warning message
print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Function to print error message
print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Check prerequisites
check_prerequisites() {
    print_step "Checking Prerequisites"
    
    # Check Docker
    if command -v docker &> /dev/null; then
        print_success "Docker is installed"
    else
        print_error "Docker is required but not installed"
        exit 1
    fi
    
    # Check Docker Compose
    if command -v docker-compose &> /dev/null || docker compose version &> /dev/null; then
        print_success "Docker Compose is available"
    else
        print_error "Docker Compose is required"
        exit 1
    fi
    
    # Check if we're in the right directory
    if [ -f "docker-compose.yaml" ] && [ -f "Makefile" ]; then
        print_success "Running from correct directory"
    else
        print_error "Please run this script from the project root directory"
        exit 1
    fi
}

# Setup environment file
setup_environment() {
    print_step "Setting Up Environment"
    
    if [ ! -f ".env" ]; then
        if [ -f ".env.example" ]; then
            cp .env.example .env
            print_success "Created .env from .env.example"
            print_warning "Please review and update .env file with your specific configuration"
        else
            print_error ".env.example file not found"
            exit 1
        fi
    else
        print_success ".env file already exists"
    fi
}

# Create necessary directories
create_directories() {
    print_step "Creating Monitoring Directories"
    
    directories=(
        "logs"
        "reports"
        "monitoring/prometheus"
        "monitoring/grafana/dashboards"
        "monitoring/grafana/provisioning/dashboards"
        "monitoring/grafana/provisioning/datasources"
        "monitoring/loki"
        "monitoring/promtail"
        "monitoring/alertmanager"
        "monitoring/mysql"
        "monitoring/redis"
    )
    
    for dir in "${directories[@]}"; do
        if [ ! -d "$dir" ]; then
            mkdir -p "$dir"
            print_success "Created directory: $dir"
        fi
    done
}

# Stop any existing services
stop_existing_services() {
    print_step "Stopping Existing Services"
    
    if docker-compose ps | grep -q "Up"; then
        docker-compose down
        print_success "Stopped existing services"
    else
        print_success "No running services found"
    fi
}

# Pull required images
pull_images() {
    print_step "Pulling Docker Images"
    
    docker-compose pull
    print_success "All images pulled successfully"
}

# Start the monitoring stack
start_monitoring_stack() {
    print_step "Starting Monitoring Stack"
    
    # Start dependencies first
    echo -e "${YELLOW}Starting database and cache services...${NC}"
    docker-compose up -d mysql redis nats
    
    # Wait for database to be ready
    echo -e "${YELLOW}Waiting for database to be ready...${NC}"
    sleep 15
    
    # Check if database is ready
    max_attempts=30
    attempt=0
    while [ $attempt -lt $max_attempts ]; do
        if docker exec mysql mysqladmin -u root -ppassword ping >/dev/null 2>&1; then
            print_success "Database is ready"
            break
        fi
        attempt=$((attempt + 1))
        echo -e "${YELLOW}Waiting for database... (${attempt}/${max_attempts})${NC}"
        sleep 2
    done
    
    if [ $attempt -eq $max_attempts ]; then
        print_error "Database failed to start within expected time"
        exit 1
    fi
    
    # Start monitoring services
    echo -e "${YELLOW}Starting monitoring services...${NC}"
    docker-compose up -d prometheus grafana jaeger loki promtail alertmanager
    
    # Start exporters
    echo -e "${YELLOW}Starting metric exporters...${NC}"
    docker-compose up -d node-exporter redis-exporter mysql-exporter
    
    # Finally start the main application
    echo -e "${YELLOW}Starting API Gateway...${NC}"
    docker-compose up -d greenlync-api-gateway
    
    print_success "All services started"
}

# Setup monitoring database
setup_monitoring_database() {
    print_step "Setting Up Monitoring Database"
    
    echo -e "${YELLOW}Running monitoring database setup...${NC}"
    
    # Check if the comprehensive monitoring setup file exists
    if [ -f "monitoring/mysql/monitoring-setup.sql" ]; then
        # Execute the comprehensive monitoring setup
        if docker exec mysql mysql -u root -ppassword greenlync < monitoring/mysql/monitoring-setup.sql; then
            print_success "Monitoring database setup completed"
        else
            print_warning "Monitoring database setup had issues, but basic setup should work"
        fi
    else
        print_warning "monitoring-setup.sql not found, using basic setup only"
    fi
}

# Wait for services to be ready
wait_for_services() {
    print_step "Waiting for Services to be Ready"
    
    services=(
        "http://localhost:8888/api/v1/system/monitor/health:API Gateway"
        "http://localhost:9090/-/healthy:Prometheus"
        "http://localhost:3000/api/health:Grafana"
        "http://localhost:16686/:Jaeger"
        "http://localhost:9093/-/healthy:AlertManager"
    )
    
    echo -e "${YELLOW}Waiting for services to be ready (this may take a few minutes)...${NC}"
    sleep 30
    
    for service in "${services[@]}"; do
        url="${service%%:*}"
        name="${service##*:}"
        
        max_attempts=30
        attempt=0
        while [ $attempt -lt $max_attempts ]; do
            if curl -f -s "$url" >/dev/null 2>&1; then
                print_success "$name is ready"
                break
            fi
            attempt=$((attempt + 1))
            if [ $((attempt % 5)) -eq 0 ]; then
                echo -e "${YELLOW}Still waiting for $name... (${attempt}/${max_attempts})${NC}"
            fi
            sleep 2
        done
        
        if [ $attempt -eq $max_attempts ]; then
            print_warning "$name may not be ready yet, but setup will continue"
        fi
    done
}

# Generate initial metrics
generate_initial_metrics() {
    print_step "Generating Initial Metrics"
    
    echo -e "${YELLOW}Making requests to generate initial metrics...${NC}"
    
    # Make various API calls to generate metrics
    for i in {1..10}; do
        curl -s http://localhost:8888/api/v1/system/monitor/health >/dev/null 2>&1 || true
        curl -s http://localhost:8888/api/v1/system/monitor/ready >/dev/null 2>&1 || true
        curl -s http://localhost:8888/api/v1/system/monitor/live >/dev/null 2>&1 || true
        curl -s http://localhost:8888/metrics >/dev/null 2>&1 || true
    done
    
    print_success "Initial metrics generated"
}

# Display final information
display_final_info() {
    print_step "Setup Complete!"
    
    echo -e "\n${GREEN}🎉 GreenLync API Gateway monitoring stack is now running!${NC}\n"
    
    echo -e "${BLUE}Access your services:${NC}"
    echo -e "• ${YELLOW}API Gateway:${NC}     http://localhost:8888"
    echo -e "• ${YELLOW}Swagger Docs:${NC}    http://localhost:8888/swagger/"
    echo -e "• ${YELLOW}Health Check:${NC}    http://localhost:8888/api/v1/system/monitor/health"
    echo -e "• ${YELLOW}Metrics:${NC}         http://localhost:8888/metrics"
    echo ""
    echo -e "${BLUE}Monitoring Tools:${NC}"
    echo -e "• ${YELLOW}Grafana:${NC}         http://localhost:3000 (admin/admin)"
    echo -e "• ${YELLOW}Prometheus:${NC}      http://localhost:9090"
    echo -e "• ${YELLOW}Jaeger:${NC}          http://localhost:16686"
    echo -e "• ${YELLOW}AlertManager:${NC}    http://localhost:9093"
    echo ""
    echo -e "${BLUE}Useful Commands:${NC}"
    echo -e "• ${YELLOW}View logs:${NC}       make logs"
    echo -e "• ${YELLOW}Test monitoring:${NC} ./scripts/test-monitoring.sh"
    echo -e "• ${YELLOW}Stop services:${NC}   make down"
    echo -e "• ${YELLOW}Restart:${NC}         make up"
    echo ""
    echo -e "${GREEN}Next steps:${NC}"
    echo -e "1. Visit Grafana and explore the pre-built dashboards"
    echo -e "2. Check Prometheus for metrics collection"
    echo -e "3. Generate some API traffic to see tracing in Jaeger"
    echo -e "4. Run the monitoring test: ${YELLOW}./scripts/test-monitoring.sh${NC}"
    echo ""
}

# Cleanup function for errors
cleanup_on_error() {
    print_error "Setup failed. Cleaning up..."
    docker-compose down >/dev/null 2>&1 || true
    exit 1
}

# Set up error handling
trap cleanup_on_error ERR

# Main execution
main() {
    echo -e "${GREEN}Starting GreenLync API Gateway monitoring setup...${NC}\n"
    
    check_prerequisites
    setup_environment
    create_directories
    stop_existing_services
    pull_images
    start_monitoring_stack
    setup_monitoring_database
    wait_for_services
    generate_initial_metrics
    display_final_info
    
    echo -e "\n${GREEN}Setup completed successfully! 🚀${NC}"
}

# Run main function
main "$@"