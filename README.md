# GreenLync API Gateway

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](Dockerfile)

A production-ready, event-driven API Gateway built with Go, Fiber, and comprehensive monitoring capabilities. Originally designed for cannabis compliance systems but adaptable for any domain.

## 🚀 Features

### Core Functionality
- **High-Performance HTTP Server** - Built with Fiber framework for optimal performance
- **Event-Driven Architecture** - NATS-based messaging for scalable microservices
- **OAuth2 Authentication** - Complete authentication and authorization system
- **Role-Based Access Control (RBAC)** - Casbin-powered authorization
- **Real-time WebSocket Support** - Bidirectional communication capabilities
- **RESTful API Design** - Well-structured API endpoints with Swagger documentation

### Monitoring & Observability
- **Prometheus Metrics** - Comprehensive application and system metrics
- **Jaeger Distributed Tracing** - Full request tracing across services
- **Grafana Dashboards** - Pre-built dashboards for monitoring
- **Centralized Logging** - Loki + Promtail for log aggregation
- **Health Checks** - Readiness and liveness probes
- **Alerting** - AlertManager integration with customizable alerts

### Infrastructure
- **Docker Support** - Multi-stage builds with security best practices
- **Kubernetes Ready** - Health checks and proper container configuration
- **Database Integration** - MySQL with connection pooling and migrations
- **Redis Caching** - High-performance caching layer
- **SMTP Integration** - Email notifications and templating

## 📋 Prerequisites

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Docker & Docker Compose** - [Download](https://docs.docker.com/get-docker/)
- **Make** - Build automation tool

### Optional (for development)
- **golangci-lint** - Code linting
- **swag** - Swagger documentation generation
- **Air** - Hot reload during development

## 🛠️ Quick Start

### 1. Clone the Repository
```bash
git clone <repository-url>
cd greenlync-api-gateway
```

### 2. Setup Environment
```bash
# Copy environment template
cp .env.example .env

# Edit .env with your configuration
nano .env
```

### 3. Install Development Tools
```bash
make setup
```

### 4. Start with Docker Compose (Recommended)
```bash
# Start all services including monitoring stack
make up

# Or start only dependencies
make stack-up
```

### 5. Verify Installation
```bash
# Check application health
make health

# View logs
make logs
```

## 📊 Monitoring Access

After starting the services, access the monitoring tools:

| Service | URL | Default Credentials |
|---------|-----|-------------------|
| **API Gateway** | http://localhost:8888 | - |
| **Swagger Docs** | http://localhost:8888/swagger | - |
| **Grafana** | http://localhost:3000 | admin/admin |
| **Prometheus** | http://localhost:9090 | - |
| **Jaeger UI** | http://localhost:16686 | - |
| **AlertManager** | http://localhost:9093 | - |

## Architecture

- **API Gateway Pattern**: Single entry point for all client requests
- **Fiber Framework**: High-performance HTTP server
- **OAuth2 + JWT**: Secure authentication with session management
- **Casbin RBAC**: Role-based access control with cannabis-specific permissions
- **Multi-Database**: MySQL (primary), Redis (cache), InfluxDB (metrics)
- **Cannabis Focused**: Built-in compliance and jurisdiction management

## Environment Variables

Essential configuration for development:

```bash
# Server Configuration
HTTP_HOST=:
HTTP_PORT=8888
BASE_URL=http://localhost:8888

# Database
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=greenlync
MYSQL_PASSWORD=your_password
MYSQL_DB=greenlync_db

# Cache & Messaging
REDIS_URL=redis:6379
NATS_HOST=nats://nats
NATS_PORT=4222

# OAuth Configuration
OAUTH_TOKEN_EXPIRES_IN=3600
OAUTH_LONG_TOKEN_EXPIRES_IN=2592000

# Monitoring
INFLUX_HOST=influx
INFLUX_PORT=8086
INFLUX_Token=your_influx_token
INFLUX_Org=GreenLync

# Email
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_FROM=noreply@greenlync.com
SMTP_LOGIN=your_email
SMTP_PASSWORD=your_app_password
```

## Development Commands

```bash
# Development
make run                    # Run application locally
make test                   # Run tests with coverage
make race                   # Run with race condition detection
make seed                   # Seed database with initial data

# Build & Deploy
make build-app             # Build Linux binary
make build-docker-image    # Build Docker image
make swagger              # Generate API documentation

# Docker
make docker-stack-up      # Start development dependencies
make docker-stack-down    # Stop development stack
```

## Project Structure

```
├── app/                   # Application bootstrap
├── cmd/                   # Main application entry point
├── config/                # Configuration management
├── data/                  # Database seeding scripts
├── handler/               # gRPC handlers (future microservices)
├── internal/
│   ├── middleware/        # HTTP middleware
│   └── server/           # HTTP server and routing
├── model/                 # Data models
├── pkg/                   # Reusable packages
│   ├── authz/            # Authorization (Casbin)
│   ├── cache/            # Redis caching
│   ├── db/               # Database utilities
│   ├── oauth2/           # OAuth2 implementation
│   └── ...
├── proto/                 # Protocol buffers (future)
└── public/               # Static assets
```

## Cannabis Compliance Features

- **Multi-Jurisdiction Support**: Manage compliance across different states/countries
- **Role-Based Access**: Cannabis-specific roles (admin, operator, compliance, viewer)
- **Audit Trails**: Complete tracking of all compliance-related activities
- **License Management**: Track and validate cannabis business licenses
- **Compliance Categories**: Built-in categorization for different compliance types
- **Reporting**: Generate compliance reports for regulatory authorities

## Authentication & Authorization

The system uses OAuth2 with JWT tokens and Casbin for RBAC:

1. **Authentication**: OAuth2 with JWT tokens stored in Redis sessions
2. **Authorization**: Casbin RBAC with cannabis-specific policies
3. **Permissions**: Granular permissions for users, inventory, compliance, and reports
4. **Multi-tenant**: Support for multiple business entities

## Testing

```bash
# Run all tests
make test

# Run with race detection
make race

# Run specific package tests
go test ./pkg/oauth2 -v
```

## Deployment

### Docker Deployment

1. **Build the application:**
   ```bash
   make build-app
   make build-docker-image
   ```

2. **Deploy with Docker Compose:**
   ```bash
   docker compose up -d
   ```

### Environment-Specific Builds

```bash
make push-dev      # Development build
make push-staging  # Staging build  
make push-prod     # Production build
```

## Database Seeding

The seeding script creates:
- Default admin user
- Cannabis-specific roles and permissions
- Compliance categories
- Jurisdiction data for major cannabis-legal states

```bash
# Seed with custom admin
make seed -user your_admin -password secure_password

# Interactive password prompt
make seed -user your_admin
```

## API Documentation

Generate and view API documentation:

```bash
make swagger
# Visit http://localhost:8888/swagger/
```

## Contributing

1. This is a boilerplate template - customize for your specific cannabis business needs
2. Follow the existing code structure and patterns
3. Add comprehensive tests for new features
4. Update documentation when adding new functionality

## License

This is a boilerplate template. Configure licensing as needed for your cannabis business.

## Support

For issues with this boilerplate:
- Create detailed issue reports
- Include environment details and error logs
- Provide steps to reproduce any problems

---

**Note**: This is a cannabis compliance boilerplate. Ensure you comply with all local and federal regulations in your jurisdiction before deploying in production.