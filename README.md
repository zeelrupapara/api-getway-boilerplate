# GreenLync API Gateway

Cannabis compliance management system built as a scalable API gateway.

## By zeelrupapara@gmail.com, Lead Developer

This is a Golang API gateway boilerplate project for cannabis compliance and operations management.

## Project Overview

GreenLync API Gateway provides a complete foundation for cannabis businesses to manage:
- Compliance tracking and reporting
- Multi-jurisdiction operations
- User authentication and authorization
- Audit trails and documentation
- Real-time data management

## Quick Start

### Prerequisites
- Go 1.21.1+
- Docker & Docker Compose
- MySQL 8.0+
- Redis 6.0+

### Development Setup

1. **Start development dependencies:**
   ```bash
   make docker-stack-up
   ```

2. **Configure environment variables:**
   Copy and modify the environment template:
   ```bash
   cp staging.env .env
   # Edit .env with your settings
   ```

3. **Seed the database:**
   ```bash
   make seed -user admin -password your_secure_password
   ```

4. **Run the application:**
   ```bash
   make run
   ```

5. **Access the API:**
   - API Gateway: http://localhost:8888
   - Health Check: http://localhost:8888/health
   - Swagger Docs: http://localhost:8888/swagger/

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